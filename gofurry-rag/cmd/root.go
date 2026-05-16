package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	env "github.com/gofurry/gofurry-rag/config"
	"github.com/gofurry/gofurry-rag/pkg/common"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Execute() {
	rootCmd := newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		_ = rootCmd.Help()
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           common.COMMON_PROJECT_NAME,
		Short:         "gofurry RAG service",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	rootCmd.PersistentFlags().String("config", "", "path to server config file")
	_ = viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))

	rootCmd.AddCommand(
		newServeCmd(),
		newInstallCmd(),
		newUninstallCmd(),
		newResetPasswordCmd(),
		newVersionCmd(),
	)
	return rootCmd
}

func initConfig() error {
	env.ConfigureServerConfig(common.COMMON_PROJECT_NAME, "server.yaml", viper.GetString("config"))
	return env.InitServerConfig(common.COMMON_PROJECT_NAME)
}

func newServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the web service",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runService()
		},
	}
}

func newInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install service to systemd",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig()
		},
		Run: func(cmd *cobra.Command, args []string) {
			svc, err := newService()
			if err != nil {
				slog.Error("service install failed", "error", err)
				return
			}
			if err = svc.Install(); err != nil {
				slog.Error("service install failed", "error", err)
				return
			}
			slog.Info("service installed")
		},
	}
}

func newUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall service from systemd",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig()
		},
		Run: func(cmd *cobra.Command, args []string) {
			svc, err := newService()
			if err != nil {
				slog.Error("service uninstall failed", "error", err)
				return
			}
			if err = svc.Uninstall(); err != nil {
				slog.Error("service uninstall failed", "error", err)
				return
			}
			slog.Info("service uninstalled")
		},
	}
}

func newResetPasswordCmd() *cobra.Command {
	var password string

	cmd := &cobra.Command{
		Use:   "reset-password",
		Short: "Reset the single console passcode",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(password) == "" {
				return fmt.Errorf("--password is required")
			}
			return initConfig()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := env.ResolveConfigFile(common.COMMON_PROJECT_NAME, "server.yaml", viper.GetString("config"))
			if err != nil {
				return err
			}
			if err := env.UpdateConsolePasscode(configPath, password); err != nil {
				return err
			}
			slog.Info("console passcode reset successfully", "config", configPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&password, "password", "", "new console passcode")
	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show service version",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig()
		},
		Run: func(cmd *cobra.Command, args []string) {
			cfg := env.GetServerConfig()
			_, appName := appIdentity()
			slog.Info(appName + " " + cfg.Server.AppVersion)
		},
	}
}
