package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

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
		newCollectorVersionCommand(),
	)
	return root
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
			fmt.Println("gf-game-collector V1.0.0")
		},
	}
}
