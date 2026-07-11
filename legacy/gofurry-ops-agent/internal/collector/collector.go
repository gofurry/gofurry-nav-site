package collector

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/config"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/model"
)

type Collector struct {
	cfg     config.Config
	version string
}

func New(cfg config.Config, version string) *Collector {
	return &Collector{cfg: cfg, version: version}
}

func (c *Collector) Collect(ctx context.Context) model.Payload {
	payload := model.Payload{
		NodeID:       c.cfg.Node.ID,
		NodeName:     c.cfg.Node.Name,
		Region:       c.cfg.Node.Region,
		Role:         c.cfg.Node.Role,
		Timestamp:    time.Now().UTC(),
		AgentVersion: c.version,
	}

	if c.cfg.System.Enabled {
		system, disks, networks, err := collectSystem(ctx, c.cfg.System)
		if err != nil {
			payload.Errors = append(payload.Errors, "system: "+err.Error())
			slog.Warn("system collection failed", "error", err)
		}
		payload.System = system
		payload.Disks = disks
		payload.Networks = networks
	}
	if c.cfg.Docker.Enabled {
		results, err := collectDocker(ctx, c.cfg.Docker)
		if err != nil {
			payload.Errors = append(payload.Errors, "docker: "+err.Error())
			slog.Warn("docker collection failed", "error", err)
		}
		payload.Docker = results
	}
	for _, item := range c.cfg.HTTPChecks {
		payload.HTTPChecks = append(payload.HTTPChecks, collectHTTP(ctx, item))
	}
	for _, item := range c.cfg.Postgres {
		if item.Enabled {
			payload.Postgres = append(payload.Postgres, collectPostgres(ctx, item))
		}
	}
	for _, item := range c.cfg.Redis {
		if item.Enabled {
			payload.Redis = append(payload.Redis, collectRedis(ctx, item))
		}
	}
	for _, item := range c.cfg.CertChecks {
		payload.Certs = append(payload.Certs, collectCert(ctx, item))
	}
	return payload
}

func statusFromError(err error) string {
	if err == nil {
		return "ok"
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return "timeout"
	}
	return "down"
}
