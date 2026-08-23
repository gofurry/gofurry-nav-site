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
	systemdinstaller "github.com/gofurry/gofurry-admin/internal/systemd"
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
		newUninstallCmd(),
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
	var force bool
	command := &cobra.Command{
		Use:   "install",
		Short: "Install and enable the systemd unit without starting it",
		Args:  cobra.NoArgs,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return initConfig(options.configFile)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			unitPath, err := systemdinstaller.Install(cmd.Context(), systemdinstaller.InstallRequest{
				ServiceName: common.COMMON_PROJECT_NAME,
				Description: "GoFurry admin service",
				ConfigFile:  options.configFile,
				Force:       force,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Installed: %s\nEnabled:   yes\nStarted:   no\n\nStart manually with:\n  sudo systemctl start %s\n", unitPath, common.COMMON_PROJECT_NAME)
			return nil
		},
	}
	command.Flags().BoolVar(&force, "force", false, "replace an existing systemd unit after validation")
	return command
}

func newUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Stop, disable, and remove the systemd unit",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := systemdinstaller.Uninstall(cmd.Context(), common.COMMON_PROJECT_NAME); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Uninstalled: %s.service\nApplication files and data were not removed.\n", common.COMMON_PROJECT_NAME)
			return nil
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
