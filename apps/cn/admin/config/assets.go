package env

import (
	"fmt"
	"strings"
	"time"
	_ "time/tzdata"
)

type AssetStorageConfig struct {
	Primary ObjectStoreConfig `mapstructure:"primary" yaml:"primary"`
	Mirror  ObjectStoreConfig `mapstructure:"mirror" yaml:"mirror"`
}

type ObjectStoreConfig struct {
	Provider        string `mapstructure:"provider" yaml:"provider"`
	Bucket          string `mapstructure:"bucket" yaml:"bucket"`
	Region          string `mapstructure:"region" yaml:"region"`
	Endpoint        string `mapstructure:"endpoint" yaml:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id" yaml:"access_key_id" json:"-"`
	SecretAccessKey string `mapstructure:"secret_access_key" yaml:"secret_access_key" json:"-"`
	PublicBaseURL   string `mapstructure:"public_base_url" yaml:"public_base_url"`
}

func (cfg ObjectStoreConfig) Configured() bool {
	return cfg.Bucket != "" && cfg.Region != "" && cfg.AccessKeyID != "" && cfg.SecretAccessKey != ""
}

type CloudOpsConfig struct {
	EdgeOne    EdgeOneConfig    `mapstructure:"edgeone" yaml:"edgeone"`
	Cloudflare CloudflareConfig `mapstructure:"cloudflare" yaml:"cloudflare"`
}
type EdgeOneConfig struct {
	Enabled        bool                        `mapstructure:"enabled" yaml:"enabled"`
	ZoneID         string                      `mapstructure:"zone_id" yaml:"zone_id"`
	SecretID       string                      `mapstructure:"secret_id" yaml:"secret_id" json:"-"`
	SecretKey      string                      `mapstructure:"secret_key" yaml:"secret_key" json:"-"`
	MainHost       string                      `mapstructure:"main_host" yaml:"main_host"`
	AssetHost      string                      `mapstructure:"asset_host" yaml:"asset_host"`
	ScheduledPurge EdgeOneScheduledPurgeConfig `mapstructure:"scheduled_purge" yaml:"scheduled_purge"`
}

type EdgeOneScheduledPurgeConfig struct {
	Enabled  bool   `mapstructure:"enabled" yaml:"enabled"`
	Time     string `mapstructure:"time" yaml:"time"`
	Timezone string `mapstructure:"timezone" yaml:"timezone"`
}

// ScheduledPurgeSchedule validates enabled schedules and produces an internal
// daily cron expression. Disabled schedules require no provider credentials.
func (cfg EdgeOneConfig) ScheduledPurgeSchedule() (string, *time.Location, error) {
	if !cfg.ScheduledPurge.Enabled {
		return "", nil, nil
	}
	fail := func(message string) (string, *time.Location, error) {
		return "", nil, fmt.Errorf("external_services.cloud_ops.edgeone.scheduled_purge: %s", message)
	}
	clock := cfg.ScheduledPurge.Time
	parsed, err := time.Parse("15:04", clock)
	if err != nil || len(clock) != 5 || parsed.Format("15:04") != clock {
		return fail("time must be HH:mm (00:00–23:59)")
	}
	zone := cfg.ScheduledPurge.Timezone
	location, err := time.LoadLocation(zone)
	if err != nil || zone == "" || zone == "Local" {
		return fail("timezone must be a valid explicit IANA timezone")
	}
	if !cfg.Enabled || strings.TrimSpace(cfg.ZoneID) == "" || strings.TrimSpace(cfg.ZoneID) == "*" || strings.TrimSpace(cfg.SecretID) == "" || strings.TrimSpace(cfg.SecretKey) == "" {
		return fail("EdgeOne must be enabled with zone_id, secret_id and secret_key configured")
	}
	host := strings.TrimSpace(cfg.MainHost)
	if host == "" || strings.ContainsAny(host, "/\\:*?#@ \t\r\n") {
		return fail("main_host must be a non-empty hostname without scheme, path, port or wildcard")
	}
	return fmt.Sprintf("%d %d * * *", parsed.Minute(), parsed.Hour()), location, nil
}

type CloudflareConfig struct {
	Enabled    bool   `mapstructure:"enabled" yaml:"enabled"`
	ZoneID     string `mapstructure:"zone_id" yaml:"zone_id"`
	CacheToken string `mapstructure:"cache_token" yaml:"cache_token" json:"-"`
	AssetHost  string `mapstructure:"asset_host" yaml:"asset_host"`
}
