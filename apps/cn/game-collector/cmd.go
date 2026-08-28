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

	"github.com/gofurry/gofurry-game-collector/collector/facts"
	gameService "github.com/gofurry/gofurry-game-collector/collector/game/service"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/backfill"
	"github.com/gofurry/gofurry-game-collector/internal/infra/postgres"
	"github.com/gofurry/gofurry-game-collector/roof/env"
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
		Use:           "gf-game-collector",
		Short:         "GoFurry game collector",
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
		newCollectorOnceCommand(options, "collect", []string{"full"}, func() { gameService.GetGameService().Collect() }),
		newCollectorOnceCommand(options, "players", nil, func() { gameService.GetGameService().CollectCurrentPlayers() }),
		newCollectorOnceCommand(options, "all", nil, func() {
			gameService.GetGameService().CollectCurrentPlayers()
			gameService.GetGameService().Collect()
		}),
		newFirstAvailableBackfillCommand(options),
		newFactsCommand(options),
		newCollectorVersionCommand(),
	)
	return root
}

func newFactsCommand(options *cliOptions) *cobra.Command {
	command := &cobra.Command{Use: "facts", Short: "Inspect and project historical Game facts", Args: cobra.NoArgs}
	command.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return env.LoadServerConfig(options.configFile)
	}
	command.AddCommand(newFactsStatusCommand(), newFactsBackfillCommand(), newFactsRebuildCommand())
	return command
}

func newFactsStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use: "status", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
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
		},
	}
}

func newFactsBackfillCommand() *cobra.Command {
	var pipeline, fromText, toText string
	var maxDays int
	var dryRun bool
	command := &cobra.Command{
		Use: "backfill", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
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
		},
	}
	command.Flags().StringVar(&pipeline, "pipeline", "", "pipeline: player, state, or canonical pipeline key (default: all)")
	command.Flags().StringVar(&fromText, "from", "", "optional first UTC date (YYYY-MM-DD)")
	command.Flags().StringVar(&toText, "to", "", "optional last UTC date, inclusive (YYYY-MM-DD)")
	command.Flags().IntVar(&maxDays, "max-days", 0, "maximum total days to inspect or process (0 is unlimited)")
	command.Flags().BoolVar(&dryRun, "dry-run", false, "report finalizable days without writing facts or checkpoints")
	return command
}

func newFactsRebuildCommand() *cobra.Command {
	var pipeline, fromText, throughText, toText string
	var dryRun bool
	command := &cobra.Command{
		Use: "rebuild", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
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
		},
	}
	command.Flags().StringVar(&pipeline, "pipeline", "", "pipeline key: game.player_facts or game.state_facts")
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
		ConnectionString: dbConfig.ConnectionString(), MaxConns: dbConfig.MaxConns,
		MinConns: dbConfig.MinConns, MaxConnLifetime: seconds(dbConfig.MaxConnLifetimeSeconds),
		MaxConnLifetimeJitter: seconds(dbConfig.MaxConnLifetimeJitterSeconds),
		MaxConnIdleTime:       seconds(dbConfig.MaxConnIdleTimeSeconds),
		HealthCheckPeriod:     seconds(dbConfig.HealthCheckPeriodSeconds),
		ConnectTimeout:        seconds(dbConfig.ConnectTimeoutSeconds), PingTimeout: seconds(dbConfig.PingTimeoutSeconds),
	}, "gofurry-game-collector-facts")
	if err != nil {
		return nil, func() {}, fmt.Errorf("open PostgreSQL pool: %w", err)
	}
	cfg := env.GetServerConfig()
	engine := facts.New(pool, facts.Options{
		ReconcileInterval: cfg.Facts.ReconcileInterval(), FinalizationGrace: cfg.Facts.FinalizationGrace(),
		RetentionEnabled: cfg.Facts.RetentionEnabled,
		PlayerRawAge:     time.Duration(cfg.Collector.V2.Retention.PlayerCountsDays) * 24 * time.Hour,
		RetentionBatch:   cfg.Facts.BatchSize(),
	})
	return engine, pool.Close, nil
}

func newFirstAvailableBackfillCommand(options *cliOptions) *cobra.Command {
	var dryRun bool
	command := &cobra.Command{
		Use:   "backfill-first-available",
		Short: "Backfill canonical first-available facts from trusted legacy values",
		Args:  cobra.NoArgs,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return env.LoadServerConfig(options.configFile)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			summary, err := runFirstAvailableBackfill(cmd.Context(), dryRun)
			if err != nil {
				return err
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(summary)
		},
	}
	command.Flags().BoolVar(&dryRun, "dry-run", false, "scan and report without writing first-available facts")
	return command
}

func runFirstAvailableBackfill(ctx context.Context, dryRun bool) (backfill.Summary, error) {
	dbConfig := env.GetServerConfig().DataBase
	openCtx, cancel := context.WithTimeout(ctx, durationOrDefault(dbConfig.ConnectTimeoutSeconds+dbConfig.PingTimeoutSeconds, 8*time.Second))
	defer cancel()
	pool, err := postgres.Open(openCtx, postgres.Config{
		ConnectionString:      dbConfig.ConnectionString(),
		MaxConns:              dbConfig.MaxConns,
		MinConns:              dbConfig.MinConns,
		MaxConnLifetime:       seconds(dbConfig.MaxConnLifetimeSeconds),
		MaxConnLifetimeJitter: seconds(dbConfig.MaxConnLifetimeJitterSeconds),
		MaxConnIdleTime:       seconds(dbConfig.MaxConnIdleTimeSeconds),
		HealthCheckPeriod:     seconds(dbConfig.HealthCheckPeriodSeconds),
		ConnectTimeout:        seconds(dbConfig.ConnectTimeoutSeconds),
		PingTimeout:           seconds(dbConfig.PingTimeoutSeconds),
	}, "gofurry-game-collector-backfill")
	if err != nil {
		return backfill.Summary{}, fmt.Errorf("open PostgreSQL pool: %w", err)
	}
	defer pool.Close()
	return backfill.New(pool).Run(ctx, dryRun)
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

func newCollectorOnceCommand(options *cliOptions, use string, aliases []string, run func()) *cobra.Command {
	return &cobra.Command{
		Use:     use,
		Aliases: aliases,
		Short:   "Run a one-shot collection task",
		Args:    cobra.NoArgs,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return env.LoadServerConfig(options.configFile)
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return new(goFurry).RunCollectorOnce(run)
		},
	}
}

func newCollectorVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println("gf-game-collector v3.0.0-alpha.3")
		},
	}
}
