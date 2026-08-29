package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofurry/gofurry-nav-collector/collector/facts"
	"github.com/gofurry/gofurry-nav-collector/collector/metrics"
	"github.com/gofurry/gofurry-nav-collector/common"
	"github.com/gofurry/gofurry-nav-collector/internal/infra/postgres"
	"github.com/gofurry/gofurry-nav-collector/roof/env"
	"github.com/spf13/cobra"
)

type cliOptions struct {
	configFile string
}

func executeCLI() {
	root := newRootCommand()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	options := new(cliOptions)
	root := &cobra.Command{
		Use:           "gf-nav-collector",
		Short:         "GoFurry navigation collector",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	root.PersistentFlags().StringVar(&options.configFile, "config", "", "path to the YAML configuration file")
	root.AddCommand(
		newCollectorServeCommand(options),
		newInstallCommand(options),
		newUninstallCommand(),
		newCollectorVersionCommand(),
		newFactsCommand(options),
		newMetricsCommand(options),
	)
	return root
}

func newFactsCommand(options *cliOptions) *cobra.Command {
	command := &cobra.Command{Use: "facts", Short: "Inspect and project historical Nav facts", Args: cobra.NoArgs}
	command.PersistentPreRunE = func(_ *cobra.Command, _ []string) error { return env.LoadServerConfig(options.configFile) }
	command.AddCommand(newFactsStatusCommand(), newFactsBackfillCommand(), newFactsRebuildCommand())
	return command
}

func newFactsStatusCommand() *cobra.Command {
	return &cobra.Command{Use: "status", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		engine, closePool, err := openFactsEngine(cmd.Context())
		if err != nil {
			return err
		}
		defer closePool()
		status, err := engine.Status(cmd.Context())
		if err != nil {
			return err
		}
		return json.NewEncoder(cmd.OutOrStdout()).Encode(status)
	}}
}

func newFactsBackfillCommand() *cobra.Command {
	var pipeline, fromText, toText string
	var maxDays int
	var dryRun bool
	command := &cobra.Command{Use: "backfill", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		from, err := optionalFactDate(fromText, "--from")
		if err != nil {
			return err
		}
		to, err := optionalFactDate(toText, "--to")
		if err != nil {
			return err
		}
		if maxDays < 0 {
			return errors.New("--max-days must not be negative")
		}
		engine, closePool, err := openFactsEngine(cmd.Context())
		if err != nil {
			return err
		}
		defer closePool()
		summary, err := engine.BackfillWithOptions(cmd.Context(), facts.BackfillOptions{
			Pipeline: pipeline, From: from, Through: to, MaxDays: maxDays, DryRun: dryRun,
		})
		if err != nil {
			return err
		}
		return json.NewEncoder(cmd.OutOrStdout()).Encode(summary)
	}}
	command.Flags().StringVar(&pipeline, "pipeline", "", "pipeline: target, site, or canonical pipeline key (default: all)")
	command.Flags().StringVar(&fromText, "from", "", "optional first UTC date (YYYY-MM-DD)")
	command.Flags().StringVar(&toText, "to", "", "optional last UTC date, inclusive (YYYY-MM-DD)")
	command.Flags().IntVar(&maxDays, "max-days", 0, "maximum total days to inspect or process (0 is unlimited)")
	command.Flags().BoolVar(&dryRun, "dry-run", false, "report finalizable days without writing facts or checkpoints")
	return command
}

func newFactsRebuildCommand() *cobra.Command {
	var pipeline, fromText, throughText, toText string
	var dryRun bool
	command := &cobra.Command{Use: "rebuild", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		if throughText == "" {
			throughText = toText
		} else if toText != "" && toText != throughText {
			return errors.New("--to and --through must identify the same date when both are set")
		}
		from, err := time.Parse(time.DateOnly, fromText)
		if err != nil {
			return fmt.Errorf("parse --from: %w", err)
		}
		through, err := time.Parse(time.DateOnly, throughText)
		if err != nil {
			return fmt.Errorf("parse --through: %w", err)
		}
		engine, closePool, err := openFactsEngine(cmd.Context())
		if err != nil {
			return err
		}
		defer closePool()
		summary, err := engine.Rebuild(cmd.Context(), pipeline, from, through, dryRun)
		if err != nil {
			return err
		}
		return json.NewEncoder(cmd.OutOrStdout()).Encode(summary)
	}}
	command.Flags().StringVar(&pipeline, "pipeline", "", "pipeline key: nav.target_facts or nav.site_facts")
	command.Flags().StringVar(&fromText, "from", "", "first UTC date (YYYY-MM-DD)")
	command.Flags().StringVar(&throughText, "through", "", "last UTC date, inclusive (YYYY-MM-DD)")
	command.Flags().StringVar(&toText, "to", "", "alias for --through")
	command.Flags().BoolVar(&dryRun, "dry-run", false, "validate arguments and report without writing")
	_ = command.MarkFlagRequired("pipeline")
	_ = command.MarkFlagRequired("from")
	return command
}

func optionalFactDate(value, flagName string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", flagName, err)
	}
	return &parsed, nil
}

func openFactsEngine(ctx context.Context) (*facts.Engine, func(), error) {
	dbConfig := env.GetServerConfig().DataBase
	openCtx, cancel := context.WithTimeout(ctx, durationOrDefault(dbConfig.ConnectTimeoutSeconds+dbConfig.PingTimeoutSeconds, 8*time.Second))
	defer cancel()
	pool, err := postgres.Open(openCtx, postgres.Config{
		ConnectionString: dbConfig.ConnectionString(), MaxConns: dbConfig.MaxConns, MinConns: dbConfig.MinConns,
		MaxConnLifetime: seconds(dbConfig.MaxConnLifetimeSeconds), MaxConnLifetimeJitter: seconds(dbConfig.MaxConnLifetimeJitterSeconds),
		MaxConnIdleTime: seconds(dbConfig.MaxConnIdleTimeSeconds), HealthCheckPeriod: seconds(dbConfig.HealthCheckPeriodSeconds),
		ConnectTimeout: seconds(dbConfig.ConnectTimeoutSeconds), PingTimeout: seconds(dbConfig.PingTimeoutSeconds),
	}, "gofurry-nav-collector-facts")
	if err != nil {
		return nil, func() {}, fmt.Errorf("open PostgreSQL pool: %w", err)
	}
	cfg := env.GetServerConfig().Facts
	return facts.New(pool, facts.Options{ReconcileInterval: cfg.ReconcileInterval(), FinalizationGrace: cfg.FinalizationGrace(),
		RetentionEnabled: cfg.RetentionEnabled, ObservationKeep: cfg.KeepCount(), RetentionBatch: cfg.BatchSize()}), pool.Close, nil
}

func newMetricsCommand(options *cliOptions) *cobra.Command {
	command := &cobra.Command{Use: "metrics", Short: "Inspect and project versioned Nav metrics", Args: cobra.NoArgs}
	command.PersistentPreRunE = func(_ *cobra.Command, _ []string) error { return env.LoadServerConfig(options.configFile) }
	command.AddCommand(newMetricsStatusCommand(), newMetricsBackfillCommand(), newMetricsRebuildCommand())
	return command
}

func newMetricsStatusCommand() *cobra.Command {
	return &cobra.Command{Use: "status", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		engine, closePool, err := openMetricsEngine(cmd.Context())
		if err != nil {
			return err
		}
		defer closePool()
		status, err := engine.Status(cmd.Context())
		if err != nil {
			return err
		}
		return json.NewEncoder(cmd.OutOrStdout()).Encode(status)
	}}
}

func newMetricsBackfillCommand() *cobra.Command {
	var metric, fromText, throughText, toText string
	var version int32
	var maxDays int
	var dryRun bool
	command := &cobra.Command{Use: "backfill", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		throughText, err := metricThroughFlag(throughText, toText)
		if err != nil {
			return err
		}
		from, err := optionalFactDate(fromText, "--from")
		if err != nil {
			return err
		}
		through, err := optionalFactDate(throughText, "--through")
		if err != nil {
			return err
		}
		engine, closePool, err := openMetricsEngine(cmd.Context())
		if err != nil {
			return err
		}
		defer closePool()
		summary, err := engine.Backfill(cmd.Context(), metrics.BackfillOptions{
			Metric: metric, Version: version, From: from, Through: through, MaxDays: maxDays, DryRun: dryRun,
		})
		if err != nil {
			return err
		}
		return json.NewEncoder(cmd.OutOrStdout()).Encode(summary)
	}}
	addMetricSelectionFlags(command, &metric, &version, &fromText, &throughText, &toText, &maxDays, &dryRun)
	return command
}

func newMetricsRebuildCommand() *cobra.Command {
	var metric, fromText, throughText, toText string
	var version int32
	var maxDays int
	var dryRun bool
	command := &cobra.Command{Use: "rebuild", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		throughText, err := metricThroughFlag(throughText, toText)
		if err != nil {
			return err
		}
		from, err := time.Parse(time.DateOnly, fromText)
		if err != nil {
			return fmt.Errorf("parse --from: %w", err)
		}
		through, err := time.Parse(time.DateOnly, throughText)
		if err != nil {
			return fmt.Errorf("parse --through: %w", err)
		}
		engine, closePool, err := openMetricsEngine(cmd.Context())
		if err != nil {
			return err
		}
		defer closePool()
		summary, err := engine.Rebuild(cmd.Context(), metric, version, from, through, maxDays, dryRun)
		if err != nil {
			return err
		}
		return json.NewEncoder(cmd.OutOrStdout()).Encode(summary)
	}}
	addMetricSelectionFlags(command, &metric, &version, &fromText, &throughText, &toText, &maxDays, &dryRun)
	_ = command.MarkFlagRequired("metric")
	_ = command.MarkFlagRequired("version")
	_ = command.MarkFlagRequired("from")
	return command
}

func addMetricSelectionFlags(command *cobra.Command, metric *string, version *int32, from, through, to *string, maxDays *int, dryRun *bool) {
	command.Flags().StringVar(metric, "metric", "", "metric key (default: all active metrics)")
	command.Flags().Int32Var(version, "version", 0, "metric version (requires --metric; default: active version)")
	command.Flags().StringVar(from, "from", "", "optional first UTC date (YYYY-MM-DD)")
	command.Flags().StringVar(through, "through", "", "optional last UTC date, inclusive (YYYY-MM-DD)")
	command.Flags().StringVar(to, "to", "", "alias for --through")
	command.Flags().IntVar(maxDays, "max-days", 0, "maximum total days to inspect or process (0 is unlimited)")
	command.Flags().BoolVar(dryRun, "dry-run", false, "validate and report without writing metrics or checkpoints")
}

func metricThroughFlag(through, to string) (string, error) {
	if through == "" {
		return to, nil
	}
	if to != "" && to != through {
		return "", errors.New("--to and --through must identify the same date when both are set")
	}
	return through, nil
}

func openMetricsEngine(ctx context.Context) (*metrics.Engine, func(), error) {
	dbConfig := env.GetServerConfig().DataBase
	openCtx, cancel := context.WithTimeout(ctx, durationOrDefault(dbConfig.ConnectTimeoutSeconds+dbConfig.PingTimeoutSeconds, 8*time.Second))
	defer cancel()
	pool, err := postgres.Open(openCtx, postgres.Config{
		ConnectionString: dbConfig.ConnectionString(), MaxConns: dbConfig.MaxConns, MinConns: dbConfig.MinConns,
		MaxConnLifetime: seconds(dbConfig.MaxConnLifetimeSeconds), MaxConnLifetimeJitter: seconds(dbConfig.MaxConnLifetimeJitterSeconds),
		MaxConnIdleTime: seconds(dbConfig.MaxConnIdleTimeSeconds), HealthCheckPeriod: seconds(dbConfig.HealthCheckPeriodSeconds),
		ConnectTimeout: seconds(dbConfig.ConnectTimeoutSeconds), PingTimeout: seconds(dbConfig.PingTimeoutSeconds),
	}, "gofurry-nav-collector-metrics")
	if err != nil {
		return nil, func() {}, fmt.Errorf("open PostgreSQL pool: %w", err)
	}
	return metrics.New(pool, metrics.Options{ReconcileInterval: env.GetServerConfig().Metrics.ReconcileInterval()}), pool.Close, nil
}

func newCollectorServeCommand(options *cliOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Run scheduled collection in the foreground",
		Args:  cobra.NoArgs,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return env.LoadServerConfig(options.configFile)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			return new(goFurry).Serve(ctx)
		},
	}
}

func newCollectorVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("%s v3.0.0-alpha.4\n", common.COMMON_PROJECT_NAME)
		},
	}
}
