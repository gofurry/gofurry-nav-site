package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const defaultConfigPath = "/etc/gofurry-ops-center/config.yaml"

type Duration struct {
	time.Duration
}

type Config struct {
	CenterID  string          `yaml:"center_id"`
	Region    string          `yaml:"region"`
	Version   string          `yaml:"-"`
	Server    ServerConfig    `yaml:"server"`
	Runtime   RuntimeConfig   `yaml:"runtime"`
	Storage   StorageConfig   `yaml:"storage"`
	Security  SecurityConfig  `yaml:"security"`
	Alert     AlertConfig     `yaml:"alert"`
	Peer      PeerConfig      `yaml:"peer"`
	Retention RetentionConfig `yaml:"retention"`
	LogLevel  string          `yaml:"log_level"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type RuntimeConfig struct {
	MemoryLimitMB int `yaml:"memory_limit_mb"`
	GCPercent     int `yaml:"gc_percent"`
}

type StorageConfig struct {
	DSN         string `yaml:"dsn"`
	AutoMigrate bool   `yaml:"auto_migrate"`
}

type SecurityConfig struct {
	DashboardPasscode string       `yaml:"dashboard_passcode"`
	SessionSecret     string       `yaml:"session_secret"`
	CookieName        string       `yaml:"cookie_name"`
	CookieSecure      bool         `yaml:"cookie_secure"`
	AgentTokens       []AgentToken `yaml:"agent_tokens"`
	PeerToken         string       `yaml:"peer_token"`
	EventToken        string       `yaml:"event_token"`
	SignatureWindow   Duration     `yaml:"signature_window"`
	SessionTTL        Duration     `yaml:"session_ttl"`
}

type AgentToken struct {
	NodeID string `yaml:"node_id"`
	Token  string `yaml:"token"`
}

type AlertConfig struct {
	Enabled              bool     `yaml:"enabled"`
	NodeDownAfter        Duration `yaml:"node_down_after"`
	HTTPFailureThreshold int      `yaml:"http_failure_threshold"`
	DiskUsageWarn        float64  `yaml:"disk_usage_warn"`
	MemoryUsageWarn      float64  `yaml:"memory_usage_warn"`
	CertWarnDays         int      `yaml:"cert_warn_days"`
	SyncStaleAfter       Duration `yaml:"sync_stale_after"`
	PeerUnreachableAfter Duration `yaml:"peer_unreachable_after"`
}

type PeerConfig struct {
	Enabled          bool     `yaml:"enabled"`
	RemoteSummaryURL string   `yaml:"remote_summary_url"`
	Token            string   `yaml:"token"`
	Interval         Duration `yaml:"interval"`
}

type RetentionConfig struct {
	RawSamplesDays int `yaml:"raw_samples_days"`
}

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	if value == nil || value.Value == "" {
		return nil
	}
	parsed, err := time.ParseDuration(value.Value)
	if err != nil {
		return fmt.Errorf("parse duration %q: %w", value.Value, err)
	}
	d.Duration = parsed
	return nil
}

func (s ServerConfig) Addr() string {
	host := strings.TrimSpace(s.Host)
	if host == "" {
		host = "127.0.0.1"
	}
	return fmt.Sprintf("%s:%d", host, s.Port)
}

func Load(path string) (Config, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		path = defaultConfigPath
	}
	if !filepath.IsAbs(path) {
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal([]byte(os.ExpandEnv(string(content))), &cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config %s: %w", path, err)
	}
	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (cfg *Config) normalize() {
	cfg.CenterID = strings.TrimSpace(cfg.CenterID)
	cfg.Region = strings.TrimSpace(cfg.Region)
	if cfg.CenterID == "" {
		cfg.CenterID = "ops-center-local"
	}
	if cfg.Region == "" {
		cfg.Region = "local"
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = "127.0.0.1"
	}
	if cfg.Server.Port <= 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Runtime.GCPercent == 0 {
		cfg.Runtime.GCPercent = 100
	}
	if cfg.Security.CookieName == "" {
		cfg.Security.CookieName = "gofurry_ops_session"
	}
	if cfg.Security.SignatureWindow.Duration <= 0 {
		cfg.Security.SignatureWindow.Duration = 5 * time.Minute
	}
	if cfg.Security.SessionTTL.Duration <= 0 {
		cfg.Security.SessionTTL.Duration = 12 * time.Hour
	}
	if cfg.Alert.NodeDownAfter.Duration <= 0 {
		cfg.Alert.NodeDownAfter.Duration = 3 * time.Minute
	}
	if cfg.Alert.HTTPFailureThreshold <= 0 {
		cfg.Alert.HTTPFailureThreshold = 3
	}
	if cfg.Alert.DiskUsageWarn <= 0 {
		cfg.Alert.DiskUsageWarn = 85
	}
	if cfg.Alert.MemoryUsageWarn <= 0 {
		cfg.Alert.MemoryUsageWarn = 90
	}
	if cfg.Alert.CertWarnDays <= 0 {
		cfg.Alert.CertWarnDays = 14
	}
	if cfg.Alert.SyncStaleAfter.Duration <= 0 {
		cfg.Alert.SyncStaleAfter.Duration = 24 * time.Hour
	}
	if cfg.Alert.PeerUnreachableAfter.Duration <= 0 {
		cfg.Alert.PeerUnreachableAfter.Duration = 5 * time.Minute
	}
	if cfg.Peer.Interval.Duration <= 0 {
		cfg.Peer.Interval.Duration = time.Minute
	}
	if cfg.Retention.RawSamplesDays <= 0 {
		cfg.Retention.RawSamplesDays = 7
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
}

func (cfg Config) Validate() error {
	var errs []error
	if strings.TrimSpace(cfg.Storage.DSN) == "" {
		errs = append(errs, errors.New("storage.dsn is required"))
	}
	if strings.TrimSpace(cfg.Security.DashboardPasscode) == "" {
		errs = append(errs, errors.New("security.dashboard_passcode is required"))
	}
	if strings.TrimSpace(cfg.Security.SessionSecret) == "" {
		errs = append(errs, errors.New("security.session_secret is required"))
	}
	if len(cfg.Security.AgentTokens) == 0 {
		errs = append(errs, errors.New("security.agent_tokens is required"))
	}
	for _, item := range cfg.Security.AgentTokens {
		if strings.TrimSpace(item.NodeID) == "" || strings.TrimSpace(item.Token) == "" {
			errs = append(errs, errors.New("security.agent_tokens entries require node_id and token"))
		}
	}
	return errors.Join(errs...)
}

func (cfg Config) AgentTokenMap() map[string]string {
	result := make(map[string]string, len(cfg.Security.AgentTokens))
	for _, item := range cfg.Security.AgentTokens {
		result[strings.TrimSpace(item.NodeID)] = strings.TrimSpace(item.Token)
	}
	return result
}
