package steamclient

import (
	"fmt"
	"strings"
	"time"
)

const (
	defaultIntervalSeconds       = 60
	defaultBurst                 = 3
	defaultMaxWorkers            = 3
	defaultRequestTimeoutSeconds = 10
	defaultRetryMaxAttempts      = 2
	defaultRetryBaseDelaySeconds = 5
	defaultCooldownSeconds       = 300
)

// Config describes the Steam client adapter settings used by collector v2.
type Config struct {
	Proxy string

	APIRequestsPer5Min   int
	StoreRequestsPer5Min int

	Burst                 int
	MaxWorkers            int
	RequestTimeoutSeconds int

	Retry RetryConfig
}

// RetryConfig controls conservative retry and cooldown behavior.
type RetryConfig struct {
	MaxAttempts          int
	BaseDelaySeconds     int
	CooldownOn429Seconds int
}

// ResolvedConfig is Config after defaults and validation have been applied.
type ResolvedConfig struct {
	ProxyURLs []string

	APIInterval    time.Duration
	StoreInterval  time.Duration
	Burst          int
	MaxWorkers     int
	RequestTimeout time.Duration

	RetryMaxAttempts int
	RetryBaseDelay   time.Duration
	CooldownDuration time.Duration
}

// ResolveConfig applies safe defaults and validates one adapter config.
func ResolveConfig(cfg Config) (ResolvedConfig, error) {
	if cfg.APIRequestsPer5Min < 0 {
		return ResolvedConfig{}, fmt.Errorf("api requests per 5 minutes must not be negative")
	}
	if cfg.StoreRequestsPer5Min < 0 {
		return ResolvedConfig{}, fmt.Errorf("store requests per 5 minutes must not be negative")
	}
	if cfg.Burst < 0 {
		return ResolvedConfig{}, fmt.Errorf("burst must not be negative")
	}
	if cfg.MaxWorkers < 0 {
		return ResolvedConfig{}, fmt.Errorf("max workers must not be negative")
	}
	if cfg.RequestTimeoutSeconds < 0 {
		return ResolvedConfig{}, fmt.Errorf("request timeout seconds must not be negative")
	}
	if cfg.Retry.MaxAttempts < 0 {
		return ResolvedConfig{}, fmt.Errorf("retry max attempts must not be negative")
	}
	if cfg.Retry.BaseDelaySeconds < 0 {
		return ResolvedConfig{}, fmt.Errorf("retry base delay seconds must not be negative")
	}
	if cfg.Retry.CooldownOn429Seconds < 0 {
		return ResolvedConfig{}, fmt.Errorf("retry cooldown seconds must not be negative")
	}
	resolvedBurst := intOrDefault(cfg.Burst, defaultBurst)
	if cfg.APIRequestsPer5Min > 0 && cfg.APIRequestsPer5Min <= resolvedBurst {
		return ResolvedConfig{}, fmt.Errorf("api requests per 5 minutes must be greater than burst")
	}
	if cfg.StoreRequestsPer5Min > 0 && cfg.StoreRequestsPer5Min <= resolvedBurst {
		return ResolvedConfig{}, fmt.Errorf("store requests per 5 minutes must be greater than burst")
	}

	resolved := ResolvedConfig{
		ProxyURLs:        splitProxyURLs(cfg.Proxy),
		APIInterval:      intervalFromWindowOrDefault(cfg.APIRequestsPer5Min, defaultIntervalSeconds, resolvedBurst),
		StoreInterval:    intervalFromWindowOrDefault(cfg.StoreRequestsPer5Min, defaultIntervalSeconds, resolvedBurst),
		Burst:            resolvedBurst,
		MaxWorkers:       intOrDefault(cfg.MaxWorkers, defaultMaxWorkers),
		RequestTimeout:   secondsOrDefault(cfg.RequestTimeoutSeconds, defaultRequestTimeoutSeconds),
		RetryMaxAttempts: intOrDefault(cfg.Retry.MaxAttempts, defaultRetryMaxAttempts),
		RetryBaseDelay:   secondsOrDefault(cfg.Retry.BaseDelaySeconds, defaultRetryBaseDelaySeconds),
		CooldownDuration: secondsOrDefault(cfg.Retry.CooldownOn429Seconds, defaultCooldownSeconds),
	}
	return resolved, nil
}

func secondsOrDefault(value int, fallback int) time.Duration {
	if value == 0 {
		value = fallback
	}
	return time.Duration(value) * time.Second
}

func intervalFromWindowOrDefault(requestsPer5Min int, fallbackSeconds int, burst int) time.Duration {
	if requestsPer5Min > 0 {
		return (5 * time.Minute) / time.Duration(requestsPer5Min-burst)
	}
	return secondsOrDefault(0, fallbackSeconds)
}

func intOrDefault(value int, fallback int) int {
	if value == 0 {
		return fallback
	}
	return value
}

func splitProxyURLs(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r' || r == '\t'
	})
	proxies := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			proxies = append(proxies, part)
		}
	}
	return proxies
}
