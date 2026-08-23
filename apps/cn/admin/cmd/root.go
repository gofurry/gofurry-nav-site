package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	env "github.com/gofurry/gofurry-admin/config"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	"github.com/gofurry/gofurry-admin/internal/bootstrap"
	applog "github.com/gofurry/gofurry-admin/internal/infra/logging"
	"github.com/gofurry/gofurry-admin/pkg/common"
	"github.com/spf13/cobra"
)

type cliOptions struct {
	configFile string
}

func Execute() {
	rootCmd := newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		_ = rootCmd.Help()
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	options := new(cliOptions)
	rootCmd := &cobra.Command{
		Use:           common.COMMON_PROJECT_NAME,
		Short:         "gofurry admin service",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	rootCmd.PersistentFlags().StringVar(&options.configFile, "config", "", "path to server config file")

	rootCmd.AddCommand(
		newServeCmd(options),
		newInstallCmd(options),
		newUninstallCmd(options),
		newResetPasswordCmd(options),
		newVersionCmd(),
	)
	return rootCmd
}

func initConfig(configFile string) error {
	if strings.TrimSpace(configFile) == "" {
		return fmt.Errorf("--config is required")
	}
	env.ConfigureServerConfig(common.COMMON_PROJECT_NAME, "server.yaml", configFile)
	return env.InitServerConfig(common.COMMON_PROJECT_NAME)
}

func newServeCmd(options *cliOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the web service",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig(options.configFile)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			return runService(ctx)
		},
	}
}

func newInstallCmd(options *cliOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install service to systemd",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig(options.configFile)
		},
		Run: func(cmd *cobra.Command, args []string) {
			svc, err := newService(options.configFile)
			if err != nil {
				applog.ErrorKV("service install failed", "error", err)
				return
			}

			if err = svc.Install(); err != nil {
				applog.ErrorKV("service install failed", "error", err)
				return
			}
			applog.InfoKV("service installed")
		},
	}
}

func newUninstallCmd(options *cliOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall service from systemd",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig(options.configFile)
		},
		Run: func(cmd *cobra.Command, args []string) {
			svc, err := newService(options.configFile)
			if err != nil {
				applog.ErrorKV("service uninstall failed", "error", err)
				return
			}

			if err = svc.Uninstall(); err != nil {
				applog.ErrorKV("service uninstall failed", "error", err)
				return
			}
			applog.InfoKV("service uninstalled")
		},
	}
}

func newResetPasswordCmd(options *cliOptions) *cobra.Command {
	var password string

	cmd := &cobra.Command{
		Use:   "reset-password",
		Short: "Reset the single admin password",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := initConfig(options.configFile); err != nil {
				return err
			}
			if strings.TrimSpace(password) == "" {
				return fmt.Errorf("--password is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			runtime, err := bootstrap.Start()
			if err != nil {
				return err
			}
			defer func() { _ = runtime.Shutdown() }()

			if err := runtime.AuthService.ResetPassword(password, audit.SystemMeta("cli/reset-password")); err != nil {
				return err
			}
			applog.InfoKV("admin password reset successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&password, "password", "", "new admin password")
	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show service version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("gofurry-admin v1.0.0")
		},
	}
}
