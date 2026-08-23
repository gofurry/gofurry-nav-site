package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofurry/gofurry-uptime/internal/config"
	"github.com/gofurry/gofurry-uptime/internal/logging"
	systemdinstaller "github.com/gofurry/gofurry-uptime/internal/systemd"
	uptimeservice "github.com/gofurry/gofurry-uptime/internal/uptime"
	"github.com/spf13/cobra"
)

const (
	serviceName        = "gf-uptime"
	serviceDescription = "GoFurry uptime status service"
	version            = "1.0.0"
)

type cliOptions struct {
	configFile string
	config     *config.Config
}

func executeCLI() {
	if err := newRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	options := new(cliOptions)
	root := &cobra.Command{
		Use:           "gf-uptime",
		Short:         "GoFurry uptime status service",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	root.PersistentFlags().StringVar(&options.configFile, "config", "", "path to the YAML configuration file")
	root.AddCommand(newServeCommand(options), newInstallCommand(options), newUninstallCommand(), newVersionCommand())
	return root
}

func loadConfig(options *cliOptions) error {
	cfg, err := config.Load(options.configFile)
	if err != nil {
		return err
	}
	options.config = cfg
	return nil
}

func newServeCommand(options *cliOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Run the uptime service in the foreground",
		Args:  cobra.NoArgs,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return loadConfig(options)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			logger, err := logging.New(options.config.Log)
			if err != nil {
				return err
			}
			service, err := uptimeservice.NewService(options.config, logger)
			if err != nil {
				_ = logger.Sync()
				return err
			}
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			return service.Serve(ctx)
		},
	}
}

func newInstallCommand(options *cliOptions) *cobra.Command {
	var force bool
	command := &cobra.Command{
		Use:   "install",
		Short: "Install and enable the systemd unit without starting it",
		Args:  cobra.NoArgs,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return loadConfig(options)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			unitPath, err := systemdinstaller.Install(cmd.Context(), systemdinstaller.InstallRequest{
				ServiceName: serviceName,
				Description: serviceDescription,
				ConfigFile:  options.configFile,
				Force:       force,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Installed: %s\nEnabled:   yes\nStarted:   no\n\nStart manually with:\n  sudo systemctl start %s\n", unitPath, serviceName)
			return nil
		},
	}
	command.Flags().BoolVar(&force, "force", false, "replace an existing systemd unit after validation")
	return command
}

func newUninstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Stop, disable, and remove the systemd unit",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := systemdinstaller.Uninstall(cmd.Context(), serviceName); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Uninstalled: %s.service\nApplication files and data were not removed.\n", serviceName)
			return nil
		},
	}
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "gf-uptime V%s\n", version)
		},
	}
}
