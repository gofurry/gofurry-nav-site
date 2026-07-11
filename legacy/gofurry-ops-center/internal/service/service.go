package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/config"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/model"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/repository"
)

type Store interface {
	Ingest(ctx context.Context, payload model.AgentPayload) (repository.ServiceFailureCounts, error)
	UpsertAlert(ctx context.Context, input repository.AlertInput) error
	ResolveAlert(ctx context.Context, key string) error
	ListNodes(ctx context.Context) ([]model.Node, error)
	GetNode(ctx context.Context, nodeID string) (model.Node, error)
	MarkNodeStatus(ctx context.Context, nodeID, status string) error
	ListServiceStatuses(ctx context.Context) ([]model.ServiceStatus, error)
	ListAlerts(ctx context.Context, activeOnly bool) ([]model.AlertState, error)
	UpsertPeerSummary(ctx context.Context, item model.PeerSummary) error
	LatestPeerSummary(ctx context.Context) (*model.PeerSummary, error)
	CreateSyncRun(ctx context.Context, req model.SyncEventRequest) (model.SyncRun, error)
	LatestSyncRun(ctx context.Context) (*model.SyncRun, error)
	ListSyncRuns(ctx context.Context, limit int) ([]model.SyncRun, error)
	CreateDeployEvent(ctx context.Context, req model.DeployEventRequest) (model.DeployEvent, error)
	ListDeployEvents(ctx context.Context, limit int) ([]model.DeployEvent, error)
	OverviewMetrics(ctx context.Context, since time.Time, bucket time.Duration) (model.OverviewMetrics, error)
	NodeMetrics(ctx context.Context, nodeID string, since time.Time, bucket time.Duration) (model.NodeMetrics, error)
}

type Service struct {
	cfg   config.Config
	store Store
	now   func() time.Time
}

func New(cfg config.Config, store Store) *Service {
	return &Service{cfg: cfg, store: store, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) Ingest(ctx context.Context, payload model.AgentPayload) error {
	if strings.TrimSpace(payload.NodeID) == "" {
		return errors.New("node_id is required")
	}
	if strings.TrimSpace(payload.Region) == "" {
		payload.Region = s.cfg.Region
	}
	if payload.Timestamp.IsZero() {
		payload.Timestamp = s.now()
	}
	if err := validateAndNormalizePayload(&payload); err != nil {
		return err
	}
	failureCounts, err := s.store.Ingest(ctx, payload)
	if err != nil {
		return err
	}
	if s.cfg.Alert.Enabled {
		return s.evaluatePayloadAlerts(ctx, payload, failureCounts)
	}
	return nil
}

func (s *Service) Overview(ctx context.Context) (model.Overview, error) {
	if s.cfg.Alert.Enabled {
		_ = s.EvaluateNodeOffline(ctx)
	}
	nodes, err := s.store.ListNodes(ctx)
	if err != nil {
		return model.Overview{}, err
	}
	services, err := s.store.ListServiceStatuses(ctx)
	if err != nil {
		return model.Overview{}, err
	}
	alerts, err := s.store.ListAlerts(ctx, true)
	if err != nil {
		return model.Overview{}, err
	}
	syncRun, err := s.store.LatestSyncRun(ctx)
	if err != nil {
		return model.Overview{}, err
	}
	peer, err := s.store.LatestPeerSummary(ctx)
	if err != nil {
		return model.Overview{}, err
	}
	overview := model.Overview{
		CenterID: cfgString(s.cfg.CenterID),
		Region:   s.cfg.Region,
		Status:   "ok",
		Services: services,
		Alerts:   alerts,
		LastSync: syncRun,
		Peer:     peer,
	}
	overview.NodesTotal = len(nodes)
	for _, node := range nodes {
		if node.LastSeenAt != nil && (overview.LastHeartbeatAt == nil || node.LastSeenAt.After(*overview.LastHeartbeatAt)) {
			t := *node.LastSeenAt
			overview.LastHeartbeatAt = &t
		}
		if node.Status != "ok" {
			overview.NodesDown++
		}
	}
	for _, alert := range alerts {
		switch alert.Level {
		case "critical":
			overview.CriticalAlerts++
		case "warning":
			overview.WarningAlerts++
		}
	}
	if overview.NodesDown > 0 || overview.CriticalAlerts > 0 {
		overview.Status = "critical"
	} else if overview.WarningAlerts > 0 {
		overview.Status = "warning"
	}
	return overview, nil
}

func (s *Service) PeerSummary(ctx context.Context) (model.PeerSummary, error) {
	overview, err := s.Overview(ctx)
	if err != nil {
		return model.PeerSummary{}, err
	}
	status := overview.Status
	if status == "" {
		status = "ok"
	}
	lastSyncStatus := ""
	if overview.LastSync != nil {
		lastSyncStatus = overview.LastSync.Status
	}
	return model.PeerSummary{
		Region:          s.cfg.Region,
		CenterID:        s.cfg.CenterID,
		Status:          status,
		LastHeartbeatAt: overview.LastHeartbeatAt,
		NodesTotal:      overview.NodesTotal,
		NodesDown:       overview.NodesDown,
		CriticalAlerts:  overview.CriticalAlerts,
		WarningAlerts:   overview.WarningAlerts,
		LastSyncStatus:  lastSyncStatus,
		UpdatedAt:       s.now(),
	}, nil
}

type metricsWindow struct {
	Label  string
	Since  time.Time
	Bucket time.Duration
}

func resolveMetricsWindow(value string, now time.Time) metricsWindow {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "6h":
		return metricsWindow{Label: "6h", Since: now.Add(-6 * time.Hour), Bucket: 5 * time.Minute}
	case "24h":
		return metricsWindow{Label: "24h", Since: now.Add(-24 * time.Hour), Bucket: 15 * time.Minute}
	default:
		return metricsWindow{Label: "1h", Since: now.Add(-time.Hour), Bucket: time.Minute}
	}
}

func (s *Service) OverviewMetrics(ctx context.Context, rangeValue string) (model.OverviewMetrics, error) {
	now := s.now()
	window := resolveMetricsWindow(rangeValue, now)
	result, err := s.store.OverviewMetrics(ctx, window.Since, window.Bucket)
	if err != nil {
		return model.OverviewMetrics{}, err
	}
	result.CenterID = cfgString(s.cfg.CenterID)
	result.Region = s.cfg.Region
	result.Range = window.Label
	result.GeneratedAt = now
	if result.LastSampleAt != nil {
		result.SampleFreshnessSeconds = int64(now.Sub(*result.LastSampleAt).Seconds())
		if result.SampleFreshnessSeconds < 0 {
			result.SampleFreshnessSeconds = 0
		}
	}
	return result, nil
}

func (s *Service) NodeMetrics(ctx context.Context, nodeID, rangeValue string) (model.NodeMetrics, error) {
	node, err := s.store.GetNode(ctx, nodeID)
	if err != nil {
		return model.NodeMetrics{}, err
	}
	now := s.now()
	window := resolveMetricsWindow(rangeValue, now)
	result, err := s.store.NodeMetrics(ctx, nodeID, window.Since, window.Bucket)
	if err != nil {
		return model.NodeMetrics{}, err
	}
	result.Node = node
	result.Range = window.Label
	result.GeneratedAt = now
	if result.LastSampleAt != nil {
		result.SampleFreshnessSeconds = int64(now.Sub(*result.LastSampleAt).Seconds())
		if result.SampleFreshnessSeconds < 0 {
			result.SampleFreshnessSeconds = 0
		}
	}
	return result, nil
}

func (s *Service) RecordPeerSummary(ctx context.Context, item model.PeerSummary) error {
	if strings.TrimSpace(item.Region) == "" || strings.TrimSpace(item.CenterID) == "" {
		return errors.New("peer region and center_id are required")
	}
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = s.now()
	}
	return s.store.UpsertPeerSummary(ctx, item)
}

func (s *Service) CreateSyncRun(ctx context.Context, req model.SyncEventRequest) (model.SyncRun, error) {
	if req.Region == "" {
		req.Region = s.cfg.Region
	}
	if strings.TrimSpace(req.SyncName) == "" {
		return model.SyncRun{}, errors.New("sync_name is required")
	}
	if strings.TrimSpace(req.Status) == "" {
		req.Status = "success"
	}
	run, err := s.store.CreateSyncRun(ctx, req)
	if err != nil {
		return model.SyncRun{}, err
	}
	if s.cfg.Alert.Enabled {
		if run.Status == "success" {
			_ = s.store.ResolveAlert(ctx, alertKey("sync", run.Region, run.SyncName))
		} else {
			_ = s.store.UpsertAlert(ctx, repository.AlertInput{
				Key:     alertKey("sync", run.Region, run.SyncName),
				Region:  run.Region,
				Level:   "critical",
				Type:    "sync_failed",
				Title:   "Sync failed: " + run.SyncName,
				Message: firstNonEmpty(run.ErrorMessage, "latest sync status is "+run.Status),
			})
		}
	}
	return run, nil
}

func (s *Service) CreateDeployEvent(ctx context.Context, req model.DeployEventRequest) (model.DeployEvent, error) {
	if req.Region == "" {
		req.Region = s.cfg.Region
	}
	if strings.TrimSpace(req.ServiceName) == "" {
		return model.DeployEvent{}, errors.New("service_name is required")
	}
	if strings.TrimSpace(req.Status) == "" {
		req.Status = "success"
	}
	return s.store.CreateDeployEvent(ctx, req)
}

func (s *Service) EvaluateNodeOffline(ctx context.Context) error {
	nodes, err := s.store.ListNodes(ctx)
	if err != nil {
		return err
	}
	now := s.now()
	for _, node := range nodes {
		key := alertKey("node_down", node.Region, node.NodeID)
		if node.LastSeenAt == nil || now.Sub(*node.LastSeenAt) > s.cfg.Alert.NodeDownAfter.Duration {
			_ = s.store.MarkNodeStatus(ctx, node.NodeID, "down")
			if err := s.store.UpsertAlert(ctx, repository.AlertInput{
				Key:     key,
				Region:  node.Region,
				NodeID:  node.NodeID,
				Level:   "critical",
				Type:    "node_down",
				Title:   "Node heartbeat missing: " + node.NodeID,
				Message: fmt.Sprintf("last heartbeat is older than %s", s.cfg.Alert.NodeDownAfter.Duration),
			}); err != nil {
				return err
			}
		} else {
			_ = s.store.MarkNodeStatus(ctx, node.NodeID, "ok")
			if err := s.store.ResolveAlert(ctx, key); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) Nodes(ctx context.Context) ([]model.Node, error) {
	_ = s.EvaluateNodeOffline(ctx)
	return s.store.ListNodes(ctx)
}

func (s *Service) Node(ctx context.Context, nodeID string) (model.Node, error) {
	return s.store.GetNode(ctx, nodeID)
}

func (s *Service) Services(ctx context.Context) ([]model.ServiceStatus, error) {
	return s.store.ListServiceStatuses(ctx)
}

func (s *Service) Alerts(ctx context.Context, activeOnly bool) ([]model.AlertState, error) {
	return s.store.ListAlerts(ctx, activeOnly)
}

func (s *Service) SyncRuns(ctx context.Context, limit int) ([]model.SyncRun, error) {
	return s.store.ListSyncRuns(ctx, limit)
}

func (s *Service) DeployEvents(ctx context.Context, limit int) ([]model.DeployEvent, error) {
	return s.store.ListDeployEvents(ctx, limit)
}

func (s *Service) PeerStatus(ctx context.Context) (*model.PeerSummary, error) {
	return s.store.LatestPeerSummary(ctx)
}

func (s *Service) evaluatePayloadAlerts(ctx context.Context, payload model.AgentPayload, failureCounts repository.ServiceFailureCounts) error {
	if payload.System != nil {
		key := alertKey("memory", payload.Region, payload.NodeID)
		if payload.System.MemoryUsage >= s.cfg.Alert.MemoryUsageWarn {
			if err := s.store.UpsertAlert(ctx, repository.AlertInput{
				Key:     key,
				Region:  payload.Region,
				NodeID:  payload.NodeID,
				Level:   "warning",
				Type:    "memory_high",
				Title:   "Memory usage high: " + payload.NodeID,
				Message: fmt.Sprintf("memory usage %.1f%% >= %.1f%%", payload.System.MemoryUsage, s.cfg.Alert.MemoryUsageWarn),
			}); err != nil {
				return err
			}
		} else if err := s.store.ResolveAlert(ctx, key); err != nil {
			return err
		}
	}
	for _, disk := range payload.Disks {
		key := alertKey("disk", payload.NodeID, disk.Mount)
		if disk.Usage >= s.cfg.Alert.DiskUsageWarn {
			if err := s.store.UpsertAlert(ctx, repository.AlertInput{
				Key:     key,
				Region:  payload.Region,
				NodeID:  payload.NodeID,
				Level:   "critical",
				Type:    "disk_high",
				Title:   "Disk usage high: " + payload.NodeID + " " + disk.Mount,
				Message: fmt.Sprintf("disk usage %.1f%% >= %.1f%%", disk.Usage, s.cfg.Alert.DiskUsageWarn),
			}); err != nil {
				return err
			}
		} else if err := s.store.ResolveAlert(ctx, key); err != nil {
			return err
		}
	}
	for _, item := range payload.HTTPChecks {
		key := alertKey("http", payload.NodeID, item.Name)
		if item.Status != "ok" {
			failureCount := failureCounts[repository.ServiceStatusKey(payload.NodeID, "http", item.Name)]
			if failureCount < s.cfg.Alert.HTTPFailureThreshold {
				if err := s.store.ResolveAlert(ctx, key); err != nil {
					return err
				}
				continue
			}
			if err := s.store.UpsertAlert(ctx, repository.AlertInput{
				Key:     key,
				Region:  payload.Region,
				NodeID:  payload.NodeID,
				Level:   "critical",
				Type:    "http_failed",
				Title:   "HTTP check failed: " + item.Name,
				Message: firstNonEmpty(item.ErrorMessage, item.URL),
			}); err != nil {
				return err
			}
		} else if err := s.store.ResolveAlert(ctx, key); err != nil {
			return err
		}
	}
	for _, item := range payload.Postgres {
		if err := s.serviceAlert(ctx, payload.Region, payload.NodeID, "postgres", item.Name, item.Status, item.ErrorMessage); err != nil {
			return err
		}
	}
	for _, item := range payload.Redis {
		if err := s.serviceAlert(ctx, payload.Region, payload.NodeID, "redis", item.Name, item.Status, item.ErrorMessage); err != nil {
			return err
		}
	}
	for _, item := range payload.Certs {
		key := alertKey("cert", payload.NodeID, item.Name)
		if item.Status != "ok" || item.DaysRemaining < s.cfg.Alert.CertWarnDays {
			level := "warning"
			if item.DaysRemaining < 0 || item.Status == "down" {
				level = "critical"
			}
			if err := s.store.UpsertAlert(ctx, repository.AlertInput{
				Key:     key,
				Region:  payload.Region,
				NodeID:  payload.NodeID,
				Level:   level,
				Type:    "cert_expiring",
				Title:   "Certificate needs attention: " + item.Name,
				Message: firstNonEmpty(item.ErrorMessage, fmt.Sprintf("%d days remaining", item.DaysRemaining)),
			}); err != nil {
				return err
			}
		} else if err := s.store.ResolveAlert(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) serviceAlert(ctx context.Context, region, nodeID, typ, name, status, message string) error {
	key := alertKey(typ, nodeID, name)
	if status == "ok" {
		return s.store.ResolveAlert(ctx, key)
	}
	return s.store.UpsertAlert(ctx, repository.AlertInput{
		Key:     key,
		Region:  region,
		NodeID:  nodeID,
		Level:   "critical",
		Type:    typ + "_failed",
		Title:   strings.ToUpper(typ) + " check failed: " + name,
		Message: message,
	})
}

func alertKey(parts ...string) string {
	cleaned := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(strings.ReplaceAll(part, "/", "_"))
		if part != "" {
			cleaned = append(cleaned, part)
		}
	}
	return strings.Join(cleaned, ":")
}

func cfgString(value string) string {
	return strings.TrimSpace(value)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
