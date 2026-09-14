package env

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
	Enabled   bool   `mapstructure:"enabled" yaml:"enabled"`
	ZoneID    string `mapstructure:"zone_id" yaml:"zone_id"`
	SecretID  string `mapstructure:"secret_id" yaml:"secret_id" json:"-"`
	SecretKey string `mapstructure:"secret_key" yaml:"secret_key" json:"-"`
	MainHost  string `mapstructure:"main_host" yaml:"main_host"`
	AssetHost string `mapstructure:"asset_host" yaml:"asset_host"`
}
type CloudflareConfig struct {
	Enabled    bool   `mapstructure:"enabled" yaml:"enabled"`
	ZoneID     string `mapstructure:"zone_id" yaml:"zone_id"`
	CacheToken string `mapstructure:"cache_token" yaml:"cache_token" json:"-"`
	AssetHost  string `mapstructure:"asset_host" yaml:"asset_host"`
}
