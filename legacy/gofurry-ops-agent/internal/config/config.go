package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const defaultConfigPath = "/etc/gofurry-ops-agent/config.yaml"

type Duration struct {
	time.Duration
}

type Config struct {
	Node       NodeConfig        `yaml:"node"`
	Center     CenterConfig      `yaml:"center"`
	Collect    CollectConfig     `yaml:"collect"`
	Runtime    RuntimeConfig     `yaml:"runtime"`
	LogLevel   string            `yaml:"log_level"`
	System     SystemConfig      `yaml:"system"`
	Docker     DockerConfig      `yaml:"docker"`
	HTTPChecks []HTTPCheckConfig `yaml:"http_checks"`
	Postgres   []PostgresConfig  `yaml:"postgres"`
	Redis      []RedisConfig     `yaml:"redis"`
	CertChecks []CertCheckConfig `yaml:"cert_checks"`
	Spool      SpoolConfig       `yaml:"spool"`
}

type NodeConfig struct {
	ID     string `yaml:"id"`
	Name   string `yaml:"name"`
	Region string `yaml:"region"`
	Role   string `yaml:"role"`
}

type CenterConfig struct {
	Endpoint string   `yaml:"endpoint"`
	Token    string   `yaml:"token"`
	Timeout  Duration `yaml:"timeout"`
	Retries  int      `yaml:"retries"`
}

type CollectConfig struct {
	Interval Duration `yaml:"interval"`
}

type RuntimeConfig struct {
	MemoryLimitMB int `yaml:"memory_limit_mb"`
	GCPercent     int `yaml:"gc_percent"`
}

type SystemConfig struct {
	Enabled    bool     `yaml:"enabled"`
	DiskMounts []string `yaml:"disk_mounts"`
}

type DockerConfig struct {
	Enabled    bool     `yaml:"enabled"`
	Socket     string   `yaml:"socket"`
	Containers []string `yaml:"containers"`
	Timeout    Duration `yaml:"timeout"`
}

type HTTPCheckConfig struct {
	Name         string   `yaml:"name"`
	URL          string   `yaml:"url"`
	Method       string   `yaml:"method"`
	Timeout      Duration `yaml:"timeout"`
	ExpectStatus int      `yaml:"expect_status"`
	ExpectBody   string   `yaml:"expect_body"`
}

type PostgresConfig struct {
	Name    string   `yaml:"name"`
	Enabled bool     `yaml:"enabled"`
	DSN     string   `yaml:"dsn"`
	Timeout Duration `yaml:"timeout"`
}

type RedisConfig struct {
	Name     string   `yaml:"name"`
	Enabled  bool     `yaml:"enabled"`
	Addr     string   `yaml:"addr"`
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
	DB       int      `yaml:"db"`
	Timeout  Duration `yaml:"timeout"`
}

type CertCheckConfig struct {
	Name     string   `yaml:"name"`
	Host     string   `yaml:"host"`
	Port     string   `yaml:"port"`
	Server   string   `yaml:"server_name"`
	Timeout  Duration `yaml:"timeout"`
	WarnDays int      `yaml:"warn_days"`
}

type SpoolConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Dir      string `yaml:"dir"`
	MaxFiles int    `yaml:"max_files"`
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

func Load(path string) (Config, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		path = defaultConfigPath
	}
	if !filepath.IsAbs(path) {
		abs, err := filepath.Abs(path)
		if err == nil {
			path = abs
		}
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}
	expanded := os.ExpandEnv(string(content))
	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config %s: %w", path, err)
	}
	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (cfg *Config) normalize() {
	cfg.Node.ID = strings.TrimSpace(cfg.Node.ID)
	cfg.Node.Name = strings.TrimSpace(cfg.Node.Name)
	cfg.Node.Region = strings.TrimSpace(cfg.Node.Region)
	cfg.Node.Role = strings.TrimSpace(cfg.Node.Role)
	if cfg.Node.Region == "" {
		cfg.Node.Region = "local"
	}
	if cfg.Node.Role == "" {
		cfg.Node.Role = "server"
	}
	if cfg.Node.Name == "" {
		cfg.Node.Name = cfg.Node.ID
	}
	if cfg.Center.Timeout.Duration <= 0 {
		cfg.Center.Timeout.Duration = 5 * time.Second
	}
	if cfg.Center.Retries <= 0 {
		cfg.Center.Retries = 3
	}
	if cfg.Collect.Interval.Duration <= 0 {
		cfg.Collect.Interval.Duration = 30 * time.Second
	}
	if cfg.Runtime.GCPercent == 0 {
		cfg.Runtime.GCPercent = 100
	}
	if len(cfg.System.DiskMounts) == 0 {
		cfg.System.DiskMounts = []string{"/"}
	}
	if cfg.Docker.Socket == "" {
		cfg.Docker.Socket = "/var/run/docker.sock"
	}
	if cfg.Docker.Timeout.Duration <= 0 {
		cfg.Docker.Timeout.Duration = 3 * time.Second
	}
	for i := range cfg.HTTPChecks {
		cfg.HTTPChecks[i].Name = strings.TrimSpace(cfg.HTTPChecks[i].Name)
		cfg.HTTPChecks[i].URL = strings.TrimSpace(cfg.HTTPChecks[i].URL)
		if cfg.HTTPChecks[i].Method == "" {
			cfg.HTTPChecks[i].Method = "GET"
		}
		if cfg.HTTPChecks[i].Timeout.Duration <= 0 {
			cfg.HTTPChecks[i].Timeout.Duration = 5 * time.Second
		}
		if cfg.HTTPChecks[i].ExpectStatus == 0 {
			cfg.HTTPChecks[i].ExpectStatus = 200
		}
	}
	for i := range cfg.Postgres {
		if cfg.Postgres[i].Name == "" {
			cfg.Postgres[i].Name = "postgres"
		}
		if cfg.Postgres[i].Timeout.Duration <= 0 {
			cfg.Postgres[i].Timeout.Duration = 3 * time.Second
		}
	}
	for i := range cfg.Redis {
		if cfg.Redis[i].Name == "" {
			cfg.Redis[i].Name = "redis"
		}
		if cfg.Redis[i].Timeout.Duration <= 0 {
			cfg.Redis[i].Timeout.Duration = 3 * time.Second
		}
	}
	for i := range cfg.CertChecks {
		if cfg.CertChecks[i].Name == "" {
			cfg.CertChecks[i].Name = cfg.CertChecks[i].Host
		}
		if cfg.CertChecks[i].Port == "" {
			cfg.CertChecks[i].Port = "443"
		}
		if cfg.CertChecks[i].Server == "" {
			cfg.CertChecks[i].Server = cfg.CertChecks[i].Host
		}
		if cfg.CertChecks[i].Timeout.Duration <= 0 {
			cfg.CertChecks[i].Timeout.Duration = 5 * time.Second
		}
		if cfg.CertChecks[i].WarnDays <= 0 {
			cfg.CertChecks[i].WarnDays = 14
		}
	}
	if cfg.Spool.Dir == "" {
		cfg.Spool.Dir = "/var/lib/gofurry-ops-agent/spool"
	}
	if cfg.Spool.MaxFiles <= 0 {
		cfg.Spool.MaxFiles = 1000
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
}

func (cfg Config) Validate() error {
	var errs []error
	if cfg.Node.ID == "" {
		errs = append(errs, errors.New("node.id is required"))
	}
	if cfg.Center.Endpoint == "" {
		errs = append(errs, errors.New("center.endpoint is required"))
	} else if _, err := url.ParseRequestURI(cfg.Center.Endpoint); err != nil {
		errs = append(errs, fmt.Errorf("center.endpoint is invalid: %w", err))
	}
	if strings.TrimSpace(cfg.Center.Token) == "" {
		errs = append(errs, errors.New("center.token is required"))
	}
	for _, item := range cfg.HTTPChecks {
		if item.Name == "" {
			errs = append(errs, errors.New("http_checks[].name is required"))
		}
		if item.URL == "" {
			errs = append(errs, fmt.Errorf("http check %q url is required", item.Name))
		}
	}
	for _, item := range cfg.Postgres {
		if item.Enabled && strings.TrimSpace(item.DSN) == "" {
			errs = append(errs, fmt.Errorf("postgres %q dsn is required when enabled", item.Name))
		}
	}
	for _, item := range cfg.Redis {
		if item.Enabled && strings.TrimSpace(item.Addr) == "" {
			errs = append(errs, fmt.Errorf("redis %q addr is required when enabled", item.Name))
		}
	}
	for _, item := range cfg.CertChecks {
		if strings.TrimSpace(item.Host) == "" {
			errs = append(errs, errors.New("cert_checks[].host is required"))
		}
	}
	return errors.Join(errs...)
}
