package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/config"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/httpapi"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/model"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/repository"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/service"
)

type App struct {
	version string
	cfg     config.Config
}

func New(version string, cfg config.Config) *App {
	cfg.Version = version
	return &App{version: version, cfg: cfg}
}

func (a *App) Run(ctx context.Context) error {
	startCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	pool, err := repository.Connect(startCtx, a.cfg.Storage.DSN)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	repo := repository.New(pool)
	defer repo.Close()
	if a.cfg.Storage.AutoMigrate {
		if err := repo.Migrate(startCtx); err != nil {
			return fmt.Errorf("migrate database: %w", err)
		}
	}

	svc := service.New(a.cfg, repo)
	app := httpapi.New(a.cfg, svc)
	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Listen(a.cfg.Server.Addr(), fiber.ListenConfig{EnablePrintRoutes: a.cfg.LogLevel == "debug"})
	}()
	peerDone := a.startPeerPoller(ctx, svc)
	cleanupDone := a.startCleanupLoop(ctx, repo)

	select {
	case <-ctx.Done():
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		err := app.ShutdownWithContext(shutdownCtx)
		<-peerDone
		<-cleanupDone
		return err
	case err := <-errCh:
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *App) startPeerPoller(ctx context.Context, svc *service.Service) <-chan struct{} {
	done := make(chan struct{})
	if !a.cfg.Peer.Enabled || a.cfg.Peer.RemoteSummaryURL == "" {
		close(done)
		return done
	}
	go func() {
		defer close(done)
		ticker := time.NewTicker(a.cfg.Peer.Interval.Duration)
		defer ticker.Stop()
		a.pollPeer(ctx, svc)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				a.pollPeer(ctx, svc)
			}
		}
	}()
	return done
}

func (a *App) pollPeer(ctx context.Context, svc *service.Service) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.cfg.Peer.RemoteSummaryURL, nil)
	if err != nil {
		slog.Warn("build peer summary request failed", "error", err)
		return
	}
	if a.cfg.Peer.Token != "" {
		req.Header.Set("Authorization", "Bearer "+a.cfg.Peer.Token)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		slog.Warn("peer summary request failed", "error", err)
		_ = svc.RecordPeerSummary(context.Background(), model.PeerSummary{
			Region:    "peer",
			CenterID:  a.cfg.Peer.RemoteSummaryURL,
			Status:    "center_unreachable",
			UpdatedAt: time.Now().UTC(),
		})
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.Warn("peer summary returned non-success", "status", resp.Status)
		return
	}
	var result struct {
		Data model.PeerSummary `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		slog.Warn("decode peer summary failed", "error", err)
		return
	}
	if result.Data.Region == "" {
		return
	}
	if err := svc.RecordPeerSummary(context.Background(), result.Data); err != nil {
		slog.Warn("record peer summary failed", "error", err)
	}
}

func (a *App) startCleanupLoop(ctx context.Context, repo *repository.Repository) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				cleanupCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
				err := repo.CleanupRawSamples(cleanupCtx, time.Now().AddDate(0, 0, -a.cfg.Retention.RawSamplesDays))
				cancel()
				if err != nil && !errors.Is(err, context.Canceled) {
					slog.Warn("cleanup raw samples failed", "error", err)
				}
			}
		}
	}()
	return done
}
