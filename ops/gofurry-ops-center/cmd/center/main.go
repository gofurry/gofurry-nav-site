package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"

	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/app"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/config"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	var configPath string
	root := &cobra.Command{
		Use:           "gofurry-ops-center",
		Short:         "GoFurry lightweight ops center",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringVar(&configPath, "config", "", "path to center config file")
	root.AddCommand(
		newServeCmd(&configPath),
		newCheckConfigCmd(&configPath),
		newVersionCmd(),
	)
	return root
}

func newServeCmd(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the center server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(*configPath)
			if err != nil {
				return err
			}
			initLogger(cfg.LogLevel)
			debug.SetGCPercent(cfg.Runtime.GCPercent)
			if cfg.Runtime.MemoryLimitMB > 0 {
				debug.SetMemoryLimit(int64(cfg.Runtime.MemoryLimitMB) << 20)
			}
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			return app.New(version, cfg).Run(ctx)
		},
	}
}

func newCheckConfigCmd(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "check-config",
		Short: "Validate center config",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(*configPath)
			if err != nil {
				return err
			}
			fmt.Printf("config ok: center=%s region=%s addr=%s\n", cfg.CenterID, cfg.Region, cfg.Server.Addr())
			return nil
		},
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	}
}

func initLogger(level string) {
	slogLevel := slog.LevelInfo
	if strings.EqualFold(strings.TrimSpace(level), "debug") {
		slogLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slogLevel})))
}
