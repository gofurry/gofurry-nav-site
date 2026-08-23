package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	gameService "github.com/gofurry/gofurry-game-collector/collector/game/service"
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
		newCollectorOnceCommand(options, "collect", []string{"full"}, gameService.GetGameService().Collect),
		newCollectorOnceCommand(options, "players", nil, gameService.GetGameService().CollectCurrentPlayers),
		newCollectorOnceCommand(options, "all", nil, func() {
			gameService.GetGameService().CollectCurrentPlayers()
			gameService.GetGameService().Collect()
		}),
		newCollectorVersionCommand(),
	)
	return root
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
