package config

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Log     LogConfig     `yaml:"log"`
	Storage StorageConfig `yaml:"storage"`
	Uptime  UptimeConfig  `yaml:"uptime"`
}

type ServerConfig struct {
	Mode       string `yaml:"mode"`
	IPAddress  string `yaml:"ip_address"`
	Port       string `yaml:"port"`
	Network    string `yaml:"network"`
	AppVersion string `yaml:"app_version"`
}

type LogConfig struct {
	LogLevel      string `yaml:"log_level"`
	LogMode       string `yaml:"log_mode"`
	LogPath       string `yaml:"log_path"`
	LogMaxSize    int    `yaml:"log_max_size"`
	LogMaxBackups int    `yaml:"log_max_backups"`
	LogMaxAge     int    `yaml:"log_max_age"`
}

type StorageConfig struct {
	Path string `yaml:"path"`
}

type UptimeConfig struct {
	SampleIntervalSeconds int              `yaml:"sample_interval_seconds"`
	RetentionDays         int              `yaml:"retention_days"`
	DaysToShow            int              `yaml:"days_to_show"`
	Timezone              string           `yaml:"timezone"`
	UI                    UIConfig         `yaml:"ui"`
	Endpoints             []EndpointConfig `yaml:"endpoints"`
}

type UIConfig struct {
	Path            string  `yaml:"path"`
	Title           string  `yaml:"title"`
	Description     string  `yaml:"description"`
	Footer          string  `yaml:"footer"`
	GreenThreshold  float64 `yaml:"green_threshold"`
	YellowThreshold float64 `yaml:"yellow_threshold"`
}

type EndpointConfig struct {
	ID                  string            `yaml:"id"`
	Name                string            `yaml:"name"`
	Description         string            `yaml:"description"`
	URL                 string            `yaml:"url"`
	Method              string            `yaml:"method"`
	Headers             map[string]string `yaml:"headers"`
	ExpectedStatusCodes []int             `yaml:"expected_status_codes"`
	IntervalSeconds     int               `yaml:"interval_seconds"`
	TimeoutSeconds      int               `yaml:"timeout_seconds"`
}

func Load(path string) (*Config, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, errors.New("--config is required")
	}
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}
	data, err := yaml.Marshal(v.AllSettings())
	if err != nil {
		return nil, fmt.Errorf("normalize config %q: %w", path, err)
	}
	cfg := new(Config)
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err = decoder.Decode(cfg); err != nil {
		return nil, fmt.Errorf("decode config %q: %w", path, err)
	}
	if err = cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config %q: %w", path, err)
	}
	return cfg, nil
}

func (cfg *Config) Validate() error {
	if cfg == nil {
		return errors.New("config is required")
	}
	if strings.TrimSpace(cfg.Server.IPAddress) == "" || strings.TrimSpace(cfg.Server.Port) == "" {
		return errors.New("server.ip_address and server.port are required")
	}
	if network := strings.TrimSpace(cfg.Server.Network); network != "tcp" && network != "tcp4" && network != "tcp6" {
		return errors.New("server.network must be tcp, tcp4, or tcp6")
	}
	if strings.TrimSpace(cfg.Storage.Path) == "" {
		return errors.New("storage.path is required")
	}
	if cfg.Uptime.SampleIntervalSeconds < 1 {
		return errors.New("uptime.sample_interval_seconds must be positive")
	}
	if cfg.Uptime.RetentionDays < 1 {
		return errors.New("uptime.retention_days must be positive")
	}
	if cfg.Uptime.DaysToShow < 1 || cfg.Uptime.DaysToShow > cfg.Uptime.RetentionDays {
		return errors.New("uptime.days_to_show must be positive and not exceed retention_days")
	}
	if _, err := time.LoadLocation(strings.TrimSpace(cfg.Uptime.Timezone)); err != nil {
		return fmt.Errorf("uptime.timezone: %w", err)
	}
	if strings.TrimSpace(cfg.Uptime.UI.Path) == "" {
		return errors.New("uptime.ui.path is required")
	}
	if cfg.Uptime.UI.GreenThreshold <= 0 || cfg.Uptime.UI.GreenThreshold > 1 ||
		cfg.Uptime.UI.YellowThreshold <= 0 || cfg.Uptime.UI.YellowThreshold > cfg.Uptime.UI.GreenThreshold {
		return errors.New("uptime UI thresholds must satisfy 0 < yellow <= green <= 1")
	}
	if len(cfg.Uptime.Endpoints) == 0 {
		return errors.New("uptime.endpoints must contain at least one endpoint")
	}
	seen := make(map[string]struct{}, len(cfg.Uptime.Endpoints))
	for index := range cfg.Uptime.Endpoints {
		endpoint := &cfg.Uptime.Endpoints[index]
		endpoint.ID = strings.TrimSpace(endpoint.ID)
		endpoint.Name = strings.TrimSpace(endpoint.Name)
		endpoint.Description = strings.TrimSpace(endpoint.Description)
		endpoint.URL = strings.TrimSpace(endpoint.URL)
		endpoint.Method = strings.ToUpper(strings.TrimSpace(endpoint.Method))
		if endpoint.ID == "" || endpoint.URL == "" {
			return fmt.Errorf("uptime endpoint %d requires id and url", index)
		}
		if _, exists := seen[endpoint.ID]; exists {
			return fmt.Errorf("uptime endpoint id %q is duplicated", endpoint.ID)
		}
		seen[endpoint.ID] = struct{}{}
		parsed, err := url.Parse(endpoint.URL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return fmt.Errorf("uptime endpoint %q has an invalid HTTP URL", endpoint.ID)
		}
		if endpoint.Method == "" {
			endpoint.Method = http.MethodGet
		}
		if endpoint.Method != http.MethodGet && endpoint.Method != http.MethodHead {
			return fmt.Errorf("uptime endpoint %q method must be GET or HEAD", endpoint.ID)
		}
		if endpoint.IntervalSeconds < 1 || endpoint.TimeoutSeconds < 1 {
			return fmt.Errorf("uptime endpoint %q interval and timeout must be positive", endpoint.ID)
		}
		for _, status := range endpoint.ExpectedStatusCodes {
			if status < 100 || status > 599 {
				return fmt.Errorf("uptime endpoint %q expected status must be between 100 and 599", endpoint.ID)
			}
		}
	}
	return nil
}
