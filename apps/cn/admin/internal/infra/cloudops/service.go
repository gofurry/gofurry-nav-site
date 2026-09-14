// Package cloudops exposes narrowly scoped CDN operations; it owns no database.
package cloudops

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"sync"

	env "github.com/gofurry/gofurry-admin/config"
	"github.com/gofurry/gofurry-admin/internal/infra/assets"
)

type PurgeRequest struct {
	Type    string   `json:"type"`
	Targets []string `json:"targets"`
}
type PurgeResult struct {
	JobID     string    `json:"job_id"`
	RequestID string    `json:"request_id,omitempty"`
	Status    string    `json:"status"`
	Failures  []Failure `json:"failures,omitempty"`
}
type Failure struct {
	Reason  string   `json:"reason"`
	Targets []string `json:"targets"`
}
type Task struct {
	JobID     string `json:"job_id"`
	Target    string `json:"target"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Failure   string `json:"failure,omitempty"`
}
type Tasks struct {
	Total uint64 `json:"total"`
	List  []Task `json:"list"`
}

type Service struct {
	Storage       *assets.Service
	storageConfig env.AssetStorageConfig
	config        env.CloudOpsConfig
	edge          *edgeOne
	flare         *cloudflare
}

func New(storage *assets.Service, storageConfig env.AssetStorageConfig, cfg env.CloudOpsConfig) (*Service, error) {
	s := &Service{Storage: storage, storageConfig: storageConfig, config: cfg}
	var err error
	if edgeConfigured(cfg.EdgeOne) {
		s.edge, err = newEdgeOne(cfg.EdgeOne)
		if err != nil {
			return nil, err
		}
	}
	if flareConfigured(cfg.Cloudflare) {
		s.flare = newCloudflare(cfg.Cloudflare)
	}
	return s, nil
}

func edgeConfigured(c env.EdgeOneConfig) bool {
	return c.Enabled && c.ZoneID != "" && c.ZoneID != "*" && c.SecretID != "" && c.SecretKey != ""
}
func flareConfigured(c env.CloudflareConfig) bool {
	return c.Enabled && c.ZoneID != "" && c.CacheToken != ""
}

type StoreStatus struct {
	Provider      string `json:"provider"`
	Bucket        string `json:"bucket"`
	Region        string `json:"region"`
	PublicBaseURL string `json:"public_base_url"`
	Configured    bool   `json:"configured"`
	Reachable     bool   `json:"reachable"`
	ProbeExists   bool   `json:"probe_exists"`
	Error         string `json:"error,omitempty"`
}
type CDNStatus struct {
	Configured bool   `json:"configured"`
	MainHost   string `json:"main_host,omitempty"`
	AssetHost  string `json:"asset_host"`
}
type Overview struct {
	Primary    StoreStatus `json:"primary"`
	Mirror     StoreStatus `json:"mirror"`
	EdgeOne    CDNStatus   `json:"edgeone"`
	Cloudflare CDNStatus   `json:"cloudflare"`
}

func storeStatus(ctx context.Context, store assets.ObjectStore, cfg env.ObjectStoreConfig) StoreStatus {
	status := StoreStatus{Provider: cfg.Provider, Bucket: cfg.Bucket, Region: cfg.Region, PublicBaseURL: cfg.PublicBaseURL, Configured: store != nil}
	if store == nil {
		return status
	}
	info, err := store.Head(ctx, assets.ProbeKey)
	if err != nil {
		status.Error = err.Error()
		return status
	}
	status.Reachable, status.ProbeExists = true, info.Exists
	return status
}

func (s *Service) Overview(ctx context.Context) Overview {
	result := Overview{EdgeOne: CDNStatus{Configured: s.edge != nil, MainHost: s.config.EdgeOne.MainHost, AssetHost: s.config.EdgeOne.AssetHost}, Cloudflare: CDNStatus{Configured: s.flare != nil, AssetHost: s.config.Cloudflare.AssetHost}}
	var primary, mirror assets.ObjectStore
	if s.Storage != nil {
		primary, mirror = s.Storage.Primary, s.Storage.Mirror
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); result.Primary = storeStatus(ctx, primary, s.storageConfig.Primary) }()
	go func() { defer wg.Done(); result.Mirror = storeStatus(ctx, mirror, s.storageConfig.Mirror) }()
	wg.Wait()
	return result
}

type ObjectState struct {
	State string       `json:"state"`
	Info  *assets.Info `json:"info,omitempty"`
	Error string       `json:"error,omitempty"`
}
type Inspection struct {
	Key        string      `json:"object_key"`
	PrimaryURL string      `json:"primary_url"`
	MirrorURL  string      `json:"mirror_url"`
	Primary    ObjectState `json:"primary"`
	Mirror     ObjectState `json:"mirror"`
	Comparison string      `json:"comparison"`
}

func inspect(ctx context.Context, store assets.ObjectStore, key string) ObjectState {
	if store == nil {
		return ObjectState{State: "not_configured"}
	}
	info, err := store.Head(ctx, key)
	if err != nil {
		return ObjectState{State: "error", Error: err.Error()}
	}
	if !info.Exists {
		return ObjectState{State: "missing"}
	}
	return ObjectState{State: "ready", Info: &info}
}
func (s *Service) Inspect(ctx context.Context, key string) (Inspection, error) {
	if !assets.ValidKey(key) {
		return Inspection{}, errors.New("invalid managed object key")
	}
	result := Inspection{Key: key, PrimaryURL: assets.URL(s.storageConfig.Primary.PublicBaseURL, key), MirrorURL: assets.URL(s.storageConfig.Mirror.PublicBaseURL, key), Comparison: "unknown"}
	var primary, mirror assets.ObjectStore
	if s.Storage != nil {
		primary, mirror = s.Storage.Primary, s.Storage.Mirror
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); result.Primary = inspect(ctx, primary, key) }()
	go func() { defer wg.Done(); result.Mirror = inspect(ctx, mirror, key) }()
	wg.Wait()
	if a, b := result.Primary.Info, result.Mirror.Info; a != nil && b != nil {
		result.Comparison = "different"
		if a.SHA256 != "" && a.SHA256 == b.SHA256 && a.Size == b.Size && a.ContentType == b.ContentType && a.Kind == b.Kind && a.CacheControl == b.CacheControl {
			result.Comparison = "matching"
		}
	}
	return result, nil
}

// Purges are confined to explicit hosts, never a caller-supplied zone or suffix.
func ValidatePurge(req PurgeRequest, hosts ...string) (PurgeRequest, error) {
	if req.Type != "url" && req.Type != "prefix" && req.Type != "host" {
		return PurgeRequest{}, errors.New("type must be url, prefix or host")
	}
	if len(req.Targets) == 0 || len(req.Targets) > 20 {
		return PurgeRequest{}, errors.New("provide 1 to 20 targets")
	}
	allowed := map[string]bool{}
	for _, host := range hosts {
		if host != "" {
			allowed[strings.ToLower(host)] = true
		}
	}
	result := PurgeRequest{Type: req.Type, Targets: []string{}}
	seen := map[string]bool{}
	for _, raw := range req.Targets {
		target := strings.TrimSpace(raw)
		if len(target) > 2048 || strings.ContainsAny(target, "\r\n\\*") {
			return PurgeRequest{}, errors.New("invalid purge target")
		}
		if req.Type == "host" {
			target = strings.ToLower(target)
			if !allowed[target] {
				return PurgeRequest{}, errors.New("host is not configured for this CDN")
			}
		} else {
			u, err := url.Parse(target)
			if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || u.Port() != "" || u.Fragment != "" || !allowed[strings.ToLower(u.Host)] {
				return PurgeRequest{}, errors.New("URL must use a configured CDN host without credentials, port or fragment")
			}
			if req.Type == "prefix" && u.RawQuery != "" {
				return PurgeRequest{}, errors.New("prefix must not include a query")
			}
			u.Host = strings.ToLower(u.Host)
			if u.Path == "" {
				u.Path = "/"
			}
			target = u.String()
		}
		if !seen[target] {
			result.Targets = append(result.Targets, target)
			seen[target] = true
		}
	}
	return result, nil
}

func (s *Service) Validate(provider string, req PurgeRequest) (PurgeRequest, error) {
	switch provider {
	case "edgeone":
		return ValidatePurge(req, s.config.EdgeOne.MainHost, s.config.EdgeOne.AssetHost)
	case "cloudflare":
		return ValidatePurge(req, s.config.Cloudflare.AssetHost)
	}
	return PurgeRequest{}, errors.New("unknown CDN")
}
func (s *Service) Purge(ctx context.Context, provider string, req PurgeRequest) (PurgeResult, error) {
	req, err := s.Validate(provider, req)
	if err != nil {
		return PurgeResult{}, err
	}
	if provider == "edgeone" && s.edge != nil {
		return s.edge.purge(ctx, req)
	}
	if provider == "cloudflare" && s.flare != nil {
		return s.flare.purge(ctx, req)
	}
	return PurgeResult{}, errors.New("CDN is not configured")
}
func (s *Service) PurgeAll(ctx context.Context) (PurgeResult, error) {
	if s.edge == nil {
		return PurgeResult{}, errors.New("EdgeOne is not configured")
	}
	return s.edge.purge(ctx, PurgeRequest{Type: "all"})
}
func (s *Service) Tasks(ctx context.Context, jobID string, offset int64) (Tasks, error) {
	if s.edge == nil {
		return Tasks{}, errors.New("EdgeOne is not configured")
	}
	return s.edge.tasks(ctx, jobID, offset)
}
