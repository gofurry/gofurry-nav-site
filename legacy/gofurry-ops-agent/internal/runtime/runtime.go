package runtime

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/collector"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/config"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/reporter"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/spool"
)

type Runtime struct {
	cfg       config.Config
	collector *collector.Collector
	reporter  *reporter.Reporter
	spool     *spool.Store
}

func New(version string, cfg config.Config) *Runtime {
	rt := &Runtime{
		cfg:       cfg,
		collector: collector.New(cfg, version),
		reporter:  reporter.New(cfg.Center, cfg.Node.ID),
	}
	if cfg.Spool.Enabled {
		rt.spool = spool.New(cfg.Spool.Dir, cfg.Spool.MaxFiles)
	}
	return rt
}

func (r *Runtime) Run(ctx context.Context) error {
	ticker := time.NewTicker(r.cfg.Collect.Interval.Duration)
	defer ticker.Stop()

	if err := r.Once(ctx); err != nil {
		slog.Warn("initial collection failed", "error", err)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := r.Once(ctx); err != nil {
				slog.Warn("collection failed", "error", err)
			}
		}
	}
}

func (r *Runtime) Once(ctx context.Context) error {
	if r.spool != nil {
		if err := r.spool.Replay(ctx, r.reporter.SendRaw); err != nil {
			slog.Warn("spool replay failed", "error", err)
		}
	}
	payload := r.collector.Collect(ctx)
	body, err := r.reporter.Send(ctx, payload)
	if err == nil {
		return nil
	}
	if r.spool != nil && len(body) > 0 {
		if appendErr := r.spool.Append(body); appendErr != nil {
			return appendErr
		}
	}
	return err
}
