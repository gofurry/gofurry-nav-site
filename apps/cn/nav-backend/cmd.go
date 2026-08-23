package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofurry/gofurry-nav-backend/common"
	"github.com/gofurry/gofurry-nav-backend/roof/env"
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
		Use:           "gf-nav",
		Short:         "GoFurry navigation API",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	root.PersistentFlags().StringVar(&options.configFile, "config", "", "path to the YAML configuration file")
	root.AddCommand(newServeCommand(options), newVersionCommand())
	return root
}

func newServeCommand(options *cliOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Run the API in the foreground",
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

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("%s V1.0.0\n", common.COMMON_PROJECT_NAME)
		},
	}
}
