package steamclient

import (
	"context"
	"fmt"
	"sync"
	"time"

	steam "github.com/gofurry/steam-go"
	"golang.org/x/time/rate"
)

// Bucket identifies one collector-side Steam traffic bucket.
type Bucket string

const (
	BucketOfficialAPI Bucket = "steam_api"
	BucketStore       Bucket = "steam_store"
)

// Adapter owns the collector v2 Steam SDK client and conservative cooldown state.
type Adapter struct {
	client *steam.Client
	cfg    ResolvedConfig

	mu       sync.RWMutex
	cooldown map[Bucket]time.Time
	observer Observer
	now      func() time.Time
}

// Observer receives sanitized SDK request metadata after adapter cooldown handling.
type Observer interface {
	ObserveSteamRequest(Event)
}

// ObserverFunc adapts one function into an Observer.
type ObserverFunc func(Event)

// ObserveSteamRequest implements Observer.
func (f ObserverFunc) ObserveSteamRequest(event Event) {
	f(event)
}

// Event is one collector-side sanitized Steam request event.
type Event struct {
	Bucket        Bucket
	TrafficClass  steam.TrafficClass
	Method        string
	Host          string
	Path          string
	StatusCode    int
	ErrorKind     string
	Attempts      int
	CacheHit      bool
	BlockDetected bool
	Duration      time.Duration
	CooldownUntil time.Time
}

// New builds a collector v2 Steam adapter.
func New(cfg Config) (*Adapter, error) {
	resolved, err := ResolveConfig(cfg)
	if err != nil {
		return nil, err
	}

	adapter := &Adapter{
		cfg:      resolved,
		cooldown: make(map[Bucket]time.Time),
		now:      time.Now,
	}

	proxySelector, err := buildProxySelector(resolved)
	if err != nil {
		return nil, err
	}

	retryBackoff := steam.DefaultRetryBackoffConfig()
	retryBackoff.BaseDelay = resolved.RetryBaseDelay
	retryBackoff.MaxDelay = resolved.RetryBaseDelay * 4
	retryBackoff.RespectRetryAfter = true

	client, err := steam.NewClient(
		steam.WithTimeout(resolved.RequestTimeout),
		steam.WithProxySelector(proxySelector),
		steam.WithRequestObserver(steam.RequestObserverFunc(adapter.observeRequest)),
		steam.WithTrafficPolicy(steam.TrafficClassOfficialAPI, steam.TrafficPolicy{
			RateLimiter: ratePolicy(resolved.APIInterval, resolved.Burst),
			Retry:       retryPolicy(resolved.RetryMaxAttempts, retryBackoff),
		}),
		steam.WithTrafficPolicy(steam.TrafficClassPublicStorePage, steam.TrafficPolicy{
			RateLimiter: ratePolicy(resolved.StoreInterval, resolved.Burst),
			Retry:       retryPolicy(resolved.RetryMaxAttempts, retryBackoff),
			BlockPolicy: &steam.TrafficBlockPolicy{HTMLSniffBytes: 512},
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("create steam client: %w", err)
	}
	adapter.client = client
	return adapter, nil
}

// Client returns the underlying steam-go client for v2 tasks.
func (a *Adapter) Client() *steam.Client {
	if a == nil {
		return nil
	}
	return a.client
}

// Config returns the resolved adapter config.
func (a *Adapter) Config() ResolvedConfig {
	if a == nil {
		return ResolvedConfig{}
	}
	return a.cfg
}

// SetObserver installs one optional sanitized observer for collector logs or metrics.
func (a *Adapter) SetObserver(observer Observer) {
	if a == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.observer = observer
}

// WaitCooldown waits until the selected bucket leaves adapter-level cooldown.
func (a *Adapter) WaitCooldown(ctx context.Context, bucket Bucket) error {
	if a == nil {
		return fmt.Errorf("steam adapter is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	for {
		wait := a.cooldownRemaining(bucket)
		if wait <= 0 {
			return nil
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

// Run waits for adapter cooldown, then executes fn with the underlying SDK client.
func (a *Adapter) Run(ctx context.Context, bucket Bucket, fn func(context.Context, *steam.Client) error) error {
	if fn == nil {
		return fmt.Errorf("steam adapter run function is required")
	}
	if err := a.WaitCooldown(ctx, bucket); err != nil {
		return err
	}
	return fn(ctx, a.client)
}

// CooldownUntil returns the current cooldown deadline for one bucket.
func (a *Adapter) CooldownUntil(bucket Bucket) time.Time {
	if a == nil {
		return time.Time{}
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.cooldown[bucket]
}

// Close releases idle connections held by the SDK client.
func (a *Adapter) Close() {
	if a == nil || a.client == nil {
		return
	}
	a.client.Close()
}

func (a *Adapter) cooldownRemaining(bucket Bucket) time.Duration {
	a.mu.RLock()
	until := a.cooldown[bucket]
	a.mu.RUnlock()

	now := a.now()
	if until.IsZero() || !until.After(now) {
		return 0
	}
	return until.Sub(now)
}

func (a *Adapter) observeRequest(event steam.RequestEvent) {
	bucket := bucketFromTrafficClass(event.TrafficClass)
	collectorEvent := Event{
		Bucket:        bucket,
		TrafficClass:  event.TrafficClass,
		Method:        event.Method,
		Host:          event.Host,
		Path:          event.Path,
		StatusCode:    event.StatusCode,
		ErrorKind:     event.ErrorKind,
		Attempts:      event.Attempts,
		CacheHit:      event.CacheHit,
		BlockDetected: event.BlockDetected,
		Duration:      event.Duration,
	}

	if a.shouldCooldown(event) {
		collectorEvent.CooldownUntil = a.setCooldown(bucket)
	}

	a.mu.RLock()
	observer := a.observer
	a.mu.RUnlock()
	if observer != nil {
		observer.ObserveSteamRequest(collectorEvent)
	}
}

func (a *Adapter) shouldCooldown(event steam.RequestEvent) bool {
	if a == nil || a.cfg.CooldownDuration <= 0 {
		return false
	}
	if bucketFromTrafficClass(event.TrafficClass) == "" {
		return false
	}
	if event.BlockDetected {
		return true
	}
	if event.ErrorKind != "" {
		return true
	}
	switch {
	case event.StatusCode == 429:
		return true
	case event.StatusCode == 403:
		return true
	case event.StatusCode >= 500:
		return true
	default:
		return false
	}
}

func (a *Adapter) setCooldown(bucket Bucket) time.Time {
	until := a.now().Add(a.cfg.CooldownDuration)
	a.mu.Lock()
	defer a.mu.Unlock()
	if current := a.cooldown[bucket]; current.After(until) {
		return current
	}
	a.cooldown[bucket] = until
	return until
}

func buildProxySelector(cfg ResolvedConfig) (steam.ProxySelector, error) {
	if len(cfg.ProxyURLs) == 0 {
		return nil, nil
	}
	return steam.NewHealthCheckedRoundRobinProxySelector(
		steam.ProxyHealthConfig{
			FailureThreshold: 2,
			Cooldown:         cfg.CooldownDuration,
		},
		cfg.ProxyURLs...,
	)
}

func ratePolicy(interval time.Duration, burst int) *steam.TrafficRateLimiterPolicy {
	if interval <= 0 || burst <= 0 {
		return nil
	}
	return &steam.TrafficRateLimiterPolicy{
		Limit: rate.Every(interval),
		Burst: burst,
	}
}

func retryPolicy(maxAttempts int, backoff steam.RetryBackoffConfig) *steam.TrafficRetryPolicy {
	if maxAttempts <= 1 {
		return &steam.TrafficRetryPolicy{Retry: 0, Backoff: backoff}
	}
	return &steam.TrafficRetryPolicy{Retry: maxAttempts - 1, Backoff: backoff}
}

func bucketFromTrafficClass(class steam.TrafficClass) Bucket {
	switch class {
	case steam.TrafficClassOfficialAPI:
		return BucketOfficialAPI
	case steam.TrafficClassPublicStorePage:
		return BucketStore
	default:
		return ""
	}
}
