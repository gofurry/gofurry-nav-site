package main

import (
	"fmt"

	systemdinstaller "github.com/gofurry/gofurry-nav-backend/internal/systemd"
	"github.com/gofurry/gofurry-nav-backend/roof/env"
	"github.com/spf13/cobra"
)

const (
	systemdServiceName = "gf-nav"
	systemdDescription = "GoFurry navigation API"
)

func newInstallCommand(options *cliOptions) *cobra.Command {
	var force bool
	command := &cobra.Command{
		Use:   "install",
		Short: "Install and enable the systemd unit without starting it",
		Args:  cobra.NoArgs,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return env.LoadServerConfig(options.configFile)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			unitPath, err := systemdinstaller.Install(cmd.Context(), systemdinstaller.InstallRequest{
				ServiceName: systemdServiceName,
				Description: systemdDescription,
				ConfigFile:  options.configFile,
				Force:       force,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Installed: %s\nEnabled:   yes\nStarted:   no\n\nStart manually with:\n  sudo systemctl start %s\n", unitPath, systemdServiceName)
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
			if err := systemdinstaller.Uninstall(cmd.Context(), systemdServiceName); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Uninstalled: %s.service\nApplication files and data were not removed.\n", systemdServiceName)
			return nil
		},
	}
}
