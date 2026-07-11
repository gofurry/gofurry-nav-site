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
	"time"

	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/config"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/runtime"
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
		Use:           "gofurry-ops-agent",
		Short:         "GoFurry lightweight ops probe",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringVar(&configPath, "config", "", "path to agent config file")

	root.AddCommand(
		newRunCmd(&configPath),
		newOnceCmd(&configPath),
		newCheckConfigCmd(&configPath),
		newVersionCmd(),
	)
	return root
}

func newRunCmd(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Run the agent collection loop",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(*configPath)
			if err != nil {
				return err
			}
			initLogger(cfg.LogLevel)
			if cfg.Runtime.GCPercent > 0 {
				debug.SetGCPercent(cfg.Runtime.GCPercent)
			}
			if cfg.Runtime.MemoryLimitMB > 0 {
				debug.SetMemoryLimit(int64(cfg.Runtime.MemoryLimitMB) << 20)
			}

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			return runtime.New(version, cfg).Run(ctx)
		},
	}
}

func newOnceCmd(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "once",
		Short: "Collect and report one sample",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(*configPath)
			if err != nil {
				return err
			}
			initLogger(cfg.LogLevel)
			ctx, cancel := context.WithTimeout(context.Background(), cfg.Center.Timeout.Duration+30*time.Second)
			defer cancel()
			return runtime.New(version, cfg).Once(ctx)
		},
	}
}

func newCheckConfigCmd(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "check-config",
		Short: "Validate the agent config",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(*configPath)
			if err != nil {
				return err
			}
			fmt.Printf("config ok: node=%s region=%s endpoint=%s\n", cfg.Node.ID, cfg.Node.Region, cfg.Center.Endpoint)
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

func loadConfig(path string) (config.Config, error) {
	cfg, err := config.Load(path)
	if err != nil {
		return config.Config{}, err
	}
	return cfg, nil
}

func initLogger(level string) {
	slogLevel := slog.LevelInfo
	if strings.EqualFold(strings.TrimSpace(level), "debug") {
		slogLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slogLevel})))
}
