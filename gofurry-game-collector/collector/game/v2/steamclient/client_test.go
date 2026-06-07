package steamclient

import (
	"context"
	"errors"
	"testing"
	"time"

	steam "github.com/gofurry/steam-go"
)

func TestResolveConfigAppliesSafeDefaults(t *testing.T) {
	t.Parallel()

	cfg, err := ResolveConfig(Config{})
	if err != nil {
		t.Fatalf("ResolveConfig returned error: %v", err)
	}

	if cfg.APIInterval != 60*time.Second {
		t.Fatalf("unexpected api interval: %s", cfg.APIInterval)
	}
	if cfg.StoreInterval != 60*time.Second {
		t.Fatalf("unexpected store interval: %s", cfg.StoreInterval)
	}
	if cfg.Burst != 3 {
		t.Fatalf("unexpected burst: %d", cfg.Burst)
	}
	if cfg.MaxWorkers != 3 {
		t.Fatalf("unexpected max workers: %d", cfg.MaxWorkers)
	}
	if cfg.RequestTimeout != 10*time.Second {
		t.Fatalf("unexpected request timeout: %s", cfg.RequestTimeout)
	}
	if cfg.RetryMaxAttempts != 2 {
		t.Fatalf("unexpected retry max attempts: %d", cfg.RetryMaxAttempts)
	}
	if cfg.RetryBaseDelay != 5*time.Second {
		t.Fatalf("unexpected retry base delay: %s", cfg.RetryBaseDelay)
	}
	if cfg.CooldownDuration != 300*time.Second {
		t.Fatalf("unexpected cooldown duration: %s", cfg.CooldownDuration)
	}
}

func TestResolveConfigSplitsProxyURLs(t *testing.T) {
	t.Parallel()

	cfg, err := ResolveConfig(Config{Proxy: "http://127.0.0.1:7897, socks5://127.0.0.1:7898"})
	if err != nil {
		t.Fatalf("ResolveConfig returned error: %v", err)
	}
	if len(cfg.ProxyURLs) != 2 {
		t.Fatalf("unexpected proxy count: %d", len(cfg.ProxyURLs))
	}
	if cfg.ProxyURLs[0] != "http://127.0.0.1:7897" || cfg.ProxyURLs[1] != "socks5://127.0.0.1:7898" {
		t.Fatalf("unexpected proxies: %#v", cfg.ProxyURLs)
	}
}

func TestResolveConfigRejectsNegativeValues(t *testing.T) {
	t.Parallel()

	cases := []Config{
		{APIIntervalSeconds: -1},
		{StoreIntervalSeconds: -1},
		{Burst: -1},
		{MaxWorkers: -1},
		{RequestTimeoutSeconds: -1},
		{Retry: RetryConfig{MaxAttempts: -1}},
		{Retry: RetryConfig{BaseDelaySeconds: -1}},
		{Retry: RetryConfig{CooldownOn429Seconds: -1}},
	}
	for _, cfg := range cases {
		if _, err := ResolveConfig(cfg); err == nil {
			t.Fatalf("expected validation error for %#v", cfg)
		}
	}
}

func TestNewRejectsInvalidProxy(t *testing.T) {
	t.Parallel()

	if _, err := New(Config{Proxy: "not-a-proxy"}); err == nil {
		t.Fatal("expected invalid proxy error")
	}
}

func TestObserverSetsCooldownForStoreFailures(t *testing.T) {
	t.Parallel()

	adapter, err := New(Config{
		Retry: RetryConfig{CooldownOn429Seconds: 7},
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	defer adapter.Close()

	now := time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC)
	adapter.now = func() time.Time { return now }

	var observed Event
	adapter.SetObserver(ObserverFunc(func(event Event) {
		observed = event
	}))

	adapter.observeRequest(steam.RequestEvent{
		TrafficClass: steam.TrafficClassPublicStorePage,
		Method:       "GET",
		Host:         "store.steampowered.com",
		Path:         "/events/ajaxgetadjacentpartnerevents",
		StatusCode:   429,
		Attempts:     2,
		Duration:     10 * time.Millisecond,
	})

	wantUntil := now.Add(7 * time.Second)
	if got := adapter.CooldownUntil(BucketStore); !got.Equal(wantUntil) {
		t.Fatalf("unexpected cooldown deadline: got %s want %s", got, wantUntil)
	}
	if observed.Bucket != BucketStore {
		t.Fatalf("unexpected observed bucket: %s", observed.Bucket)
	}
	if !observed.CooldownUntil.Equal(wantUntil) {
		t.Fatalf("unexpected observed cooldown: %s", observed.CooldownUntil)
	}
}

func TestObserverIgnoresNonSteamBuckets(t *testing.T) {
	t.Parallel()

	adapter, err := New(Config{})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	defer adapter.Close()

	adapter.observeRequest(steam.RequestEvent{
		TrafficClass: steam.TrafficClassCommunityWeb,
		StatusCode:   429,
	})

	if got := adapter.CooldownUntil(BucketStore); !got.IsZero() {
		t.Fatalf("expected no store cooldown, got %s", got)
	}
	if got := adapter.CooldownUntil(BucketOfficialAPI); !got.IsZero() {
		t.Fatalf("expected no api cooldown, got %s", got)
	}
}

func TestRunWaitsForCooldownAndHonorsContext(t *testing.T) {
	t.Parallel()

	adapter, err := New(Config{Retry: RetryConfig{CooldownOn429Seconds: 60}})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	defer adapter.Close()

	adapter.setCooldown(BucketOfficialAPI)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	err = adapter.Run(ctx, BucketOfficialAPI, func(context.Context, *steam.Client) error {
		return errors.New("should not run")
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context deadline exceeded, got %v", err)
	}
}
