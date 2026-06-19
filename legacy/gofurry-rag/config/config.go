package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/gofurry/gofurry-rag/pkg/common"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var (
	configuration *Config
	configErr     error
	configOnce    sync.Once
	configOptions = configLoaderOptions{
		projectName: common.COMMON_PROJECT_NAME,
		fileName:    "server.yaml",
	}
	configOptionsMu sync.Mutex
)

const (
	defaultConsolePasscode = "change-me"
	defaultJWTSecret       = "change-this-jwt-secret"
)

type configLoaderOptions struct {
	projectName string
	fileName    string
	configFile  string
}

type Config struct {
	ClusterID  int              `mapstructure:"cluster_id" yaml:"cluster_id"`
	Server     ServerConfig     `mapstructure:"server" yaml:"server"`
	Database   DatabaseConfig   `mapstructure:"database" yaml:"database"`
	Log        LogConfig        `mapstructure:"log" yaml:"log"`
	Middleware MiddlewareConfig `mapstructure:"middleware" yaml:"middleware"`
	Waf        WafConfig        `mapstructure:"waf" yaml:"waf"`
	Auth       AuthConfig       `mapstructure:"auth" yaml:"auth"`
	RAG        RAGConfig        `mapstructure:"rag" yaml:"rag"`

	AppName                       string  `mapstructure:"-" yaml:"-"`
	AppEnv                        string  `mapstructure:"-" yaml:"-"`
	AppAddr                       string  `mapstructure:"-" yaml:"-"`
	AdminToken                    string  `mapstructure:"-" yaml:"-"`
	ConsolePasscode               string  `mapstructure:"-" yaml:"-"`
	JWTSecret                     string  `mapstructure:"-" yaml:"-"`
	AuthCookieName                string  `mapstructure:"-" yaml:"-"`
	SessionTTLHours               int     `mapstructure:"-" yaml:"-"`
	DatabaseDSN                   string  `mapstructure:"-" yaml:"-"`
	OllamaBaseURL                 string  `mapstructure:"-" yaml:"-"`
	EmbedModel                    string  `mapstructure:"-" yaml:"-"`
	EmbedDim                      int     `mapstructure:"-" yaml:"-"`
	ChunkSize                     int     `mapstructure:"-" yaml:"-"`
	ChunkOverlap                  int     `mapstructure:"-" yaml:"-"`
	TopK                          int     `mapstructure:"-" yaml:"-"`
	QueryTimeoutSeconds           int     `mapstructure:"-" yaml:"-"`
	EmbedTimeoutSeconds           int     `mapstructure:"-" yaml:"-"`
	IngestTimeoutSeconds          int     `mapstructure:"-" yaml:"-"`
	MaxQueryQuestionRunes         int     `mapstructure:"-" yaml:"-"`
	MaxQueryTopK                  int     `mapstructure:"-" yaml:"-"`
	PublicQueryRateLimitRequests  int     `mapstructure:"-" yaml:"-"`
	PublicQueryRateLimitWindowSec int     `mapstructure:"-" yaml:"-"`
	PublicQueryMaxQuestionRunes   int     `mapstructure:"-" yaml:"-"`
	PublicQueryMaxTopK            int     `mapstructure:"-" yaml:"-"`
	PublicQueryContextMaxTurns    int     `mapstructure:"-" yaml:"-"`
	PublicQueryContextMaxRunes    int     `mapstructure:"-" yaml:"-"`
	OllamaMaxConcurrency          int     `mapstructure:"-" yaml:"-"`
	OllamaQueryQueueSize          int     `mapstructure:"-" yaml:"-"`
	OllamaIngestQueueSize         int     `mapstructure:"-" yaml:"-"`
	OllamaQueueWaitTimeoutSeconds int     `mapstructure:"-" yaml:"-"`
	TencentBaseURL                string  `mapstructure:"-" yaml:"-"`
	TencentModel                  string  `mapstructure:"-" yaml:"-"`
	TencentAPIKey                 string  `mapstructure:"-" yaml:"-"`
	TencentTimeoutSeconds         int     `mapstructure:"-" yaml:"-"`
	TencentTemperature            float64 `mapstructure:"-" yaml:"-"`
	TencentTopP                   float64 `mapstructure:"-" yaml:"-"`
	TencentMaxTokens              int     `mapstructure:"-" yaml:"-"`
	TencentReasoningEffort        string  `mapstructure:"-" yaml:"-"`
	IngestWorkers                 int     `mapstructure:"-" yaml:"-"`
	EmbedBatchSize                int     `mapstructure:"-" yaml:"-"`
	SyncEnabled                   bool    `mapstructure:"-" yaml:"-"`
	SyncIntervalMinutes           int     `mapstructure:"-" yaml:"-"`
	SyncTimeoutSeconds            int     `mapstructure:"-" yaml:"-"`
	SyncNavBaseURL                string  `mapstructure:"-" yaml:"-"`
	SyncGameBaseURL               string  `mapstructure:"-" yaml:"-"`
}

type ServerConfig struct {
	AppID               string   `mapstructure:"app_id" yaml:"app_id"`
	AppName             string   `mapstructure:"app_name" yaml:"app_name"`
	AppVersion          string   `mapstructure:"app_version" yaml:"app_version"`
	Mode                string   `mapstructure:"mode" yaml:"mode"`
	IPAddress           string   `mapstructure:"ip_address" yaml:"ip_address"`
	Port                string   `mapstructure:"port" yaml:"port"`
	MemoryLimit         int      `mapstructure:"memory_limit" yaml:"memory_limit"`
	GCPercent           int      `mapstructure:"gc_percent" yaml:"gc_percent"`
	Network             string   `mapstructure:"network" yaml:"network"`
	EnablePrefork       bool     `mapstructure:"enable_prefork" yaml:"enable_prefork"`
	IsFullStack         bool     `mapstructure:"is_full_stack" yaml:"is_full_stack"`
	TrustProxy          bool     `mapstructure:"trust_proxy" yaml:"trust_proxy"`
	ProxyHeader         string   `mapstructure:"proxy_header" yaml:"proxy_header"`
	TrustedProxies      []string `mapstructure:"trusted_proxies" yaml:"trusted_proxies"`
	TrustProxyLoopback  bool     `mapstructure:"trust_proxy_loopback" yaml:"trust_proxy_loopback"`
	TrustProxyPrivate   bool     `mapstructure:"trust_proxy_private" yaml:"trust_proxy_private"`
	TrustProxyLinkLocal bool     `mapstructure:"trust_proxy_link_local" yaml:"trust_proxy_link_local"`
}

type DatabaseConfig struct {
	Enabled     bool              `mapstructure:"enabled" yaml:"enabled"`
	AutoMigrate bool              `mapstructure:"auto_migrate" yaml:"auto_migrate"`
	DBType      string            `mapstructure:"db_type" yaml:"db_type"`
	Postgres    SQLDatabaseConfig `mapstructure:"postgres" yaml:"postgres"`
	DSN         string            `mapstructure:"dsn" yaml:"dsn"`
	DBName      string            `mapstructure:"db_name" yaml:"db_name"`
	DBHost      string            `mapstructure:"db_host" yaml:"db_host"`
	DBPort      string            `mapstructure:"db_port" yaml:"db_port"`
	DBUser      string            `mapstructure:"db_username" yaml:"db_username"`
	DBPass      string            `mapstructure:"db_password" yaml:"db_password"`
}

type SQLDatabaseConfig struct {
	DSN    string `mapstructure:"dsn" yaml:"dsn"`
	DBName string `mapstructure:"db_name" yaml:"db_name"`
	DBHost string `mapstructure:"db_host" yaml:"db_host"`
	DBPort string `mapstructure:"db_port" yaml:"db_port"`
	DBUser string `mapstructure:"db_username" yaml:"db_username"`
	DBPass string `mapstructure:"db_password" yaml:"db_password"`
}

type LogConfig struct {
	LogLevel      string `mapstructure:"log_level" yaml:"log_level"`
	LogMode       string `mapstructure:"log_mode" yaml:"log_mode"`
	LogPath       string `mapstructure:"log_path" yaml:"log_path"`
	LogMaxSize    int    `mapstructure:"log_max_size" yaml:"log_max_size"`
	LogMaxBackups int    `mapstructure:"log_max_backups" yaml:"log_max_backups"`
	LogMaxAge     int    `mapstructure:"log_max_age" yaml:"log_max_age"`
}

type MiddlewareConfig struct {
	Cors            CorsConfig            `mapstructure:"cors" yaml:"cors"`
	RequestID       RequestIDConfig       `mapstructure:"request_id" yaml:"request_id"`
	AccessLog       AccessLogConfig       `mapstructure:"access_log" yaml:"access_log"`
	Timeout         TimeoutConfig         `mapstructure:"timeout" yaml:"timeout"`
	Health          HealthConfig          `mapstructure:"health" yaml:"health"`
	SecurityHeaders SecurityHeadersConfig `mapstructure:"security_headers" yaml:"security_headers"`
	Compression     CompressionConfig     `mapstructure:"compression" yaml:"compression"`
	Limiter         LimiterConfig         `mapstructure:"limiter" yaml:"limiter"`
	ETag            ETagConfig            `mapstructure:"etag" yaml:"etag"`
}

type CorsConfig struct {
	AllowOrigins []string `mapstructure:"allow_origins" yaml:"allow_origins"`
}

type RequestIDConfig struct {
	Enabled bool   `mapstructure:"enabled" yaml:"enabled"`
	Header  string `mapstructure:"header" yaml:"header"`
}

type AccessLogConfig struct {
	Enabled    bool   `mapstructure:"enabled" yaml:"enabled"`
	Format     string `mapstructure:"format" yaml:"format"`
	TimeFormat string `mapstructure:"time_format" yaml:"time_format"`
	TimeZone   string `mapstructure:"time_zone" yaml:"time_zone"`
}

type TimeoutConfig struct {
	Enabled         bool     `mapstructure:"enabled" yaml:"enabled"`
	DurationSeconds int      `mapstructure:"duration_seconds" yaml:"duration_seconds"`
	ExcludePaths    []string `mapstructure:"exclude_paths" yaml:"exclude_paths"`
}

type HealthConfig struct {
	Enabled       bool `mapstructure:"enabled" yaml:"enabled"`
	IncludeLegacy bool `mapstructure:"include_legacy" yaml:"include_legacy"`
}

type SecurityHeadersConfig struct {
	Enabled               bool   `mapstructure:"enabled" yaml:"enabled"`
	ContentSecurityPolicy string `mapstructure:"content_security_policy" yaml:"content_security_policy"`
	PermissionPolicy      string `mapstructure:"permission_policy" yaml:"permission_policy"`
	HSTSMaxAge            int    `mapstructure:"hsts_max_age" yaml:"hsts_max_age"`
	HSTSExcludeSubdomains bool   `mapstructure:"hsts_exclude_subdomains" yaml:"hsts_exclude_subdomains"`
	HSTSPreloadEnabled    bool   `mapstructure:"hsts_preload_enabled" yaml:"hsts_preload_enabled"`
	CSPReportOnly         bool   `mapstructure:"csp_report_only" yaml:"csp_report_only"`
}

type CompressionConfig struct {
	Enabled bool   `mapstructure:"enabled" yaml:"enabled"`
	Level   string `mapstructure:"level" yaml:"level"`
}

type LimiterConfig struct {
	Enabled                bool     `mapstructure:"enabled" yaml:"enabled"`
	MaxRequests            int      `mapstructure:"max_requests" yaml:"max_requests"`
	Expiration             int      `mapstructure:"expiration" yaml:"expiration"`
	Strategy               string   `mapstructure:"strategy" yaml:"strategy"`
	KeySource              string   `mapstructure:"key_source" yaml:"key_source"`
	KeyHeader              string   `mapstructure:"key_header" yaml:"key_header"`
	SkipFailedRequests     bool     `mapstructure:"skip_failed_requests" yaml:"skip_failed_requests"`
	SkipSuccessfulRequests bool     `mapstructure:"skip_successful_requests" yaml:"skip_successful_requests"`
	DisableHeaders         bool     `mapstructure:"disable_headers" yaml:"disable_headers"`
	ExcludePaths           []string `mapstructure:"exclude_paths" yaml:"exclude_paths"`
}

type ETagConfig struct {
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`
	Weak    bool `mapstructure:"weak" yaml:"weak"`
}

type WafConfig struct {
	Enabled  bool     `mapstructure:"enabled" yaml:"enabled"`
	ConfPath []string `mapstructure:"conf_path" yaml:"conf_path"`
}

type AuthConfig struct {
	ConsolePasscode   string `mapstructure:"console_passcode" yaml:"console_passcode"`
	JWTSecret         string `mapstructure:"jwt_secret" yaml:"jwt_secret"`
	CookieName        string `mapstructure:"cookie_name" yaml:"cookie_name"`
	CookieDomain      string `mapstructure:"cookie_domain" yaml:"cookie_domain"`
	CookieSecure      bool   `mapstructure:"cookie_secure" yaml:"cookie_secure"`
	CookieMaxAgeSecs  int    `mapstructure:"cookie_max_age_seconds" yaml:"cookie_max_age_seconds"`
	SessionTTLHours   int    `mapstructure:"session_ttl_hours" yaml:"session_ttl_hours"`
	SameSite          string `mapstructure:"cookie_same_site" yaml:"cookie_same_site"`
	CSRFHeaderRequire bool   `mapstructure:"csrf_header_required" yaml:"csrf_header_required"`
}

type RAGConfig struct {
	OllamaBaseURL                 string  `mapstructure:"ollama_base_url" yaml:"ollama_base_url"`
	EmbedModel                    string  `mapstructure:"embed_model" yaml:"embed_model"`
	EmbedDim                      int     `mapstructure:"embed_dim" yaml:"embed_dim"`
	ChunkSize                     int     `mapstructure:"chunk_size" yaml:"chunk_size"`
	ChunkOverlap                  int     `mapstructure:"chunk_overlap" yaml:"chunk_overlap"`
	TopK                          int     `mapstructure:"top_k" yaml:"top_k"`
	QueryTimeoutSeconds           int     `mapstructure:"query_timeout_seconds" yaml:"query_timeout_seconds"`
	EmbedTimeoutSeconds           int     `mapstructure:"embed_timeout_seconds" yaml:"embed_timeout_seconds"`
	IngestTimeoutSeconds          int     `mapstructure:"ingest_timeout_seconds" yaml:"ingest_timeout_seconds"`
	MaxQueryQuestionRunes         int     `mapstructure:"max_query_question_runes" yaml:"max_query_question_runes"`
	MaxQueryTopK                  int     `mapstructure:"max_query_top_k" yaml:"max_query_top_k"`
	PublicQueryRateLimitRequests  int     `mapstructure:"public_query_rate_limit_requests" yaml:"public_query_rate_limit_requests"`
	PublicQueryRateLimitWindowSec int     `mapstructure:"public_query_rate_limit_window_seconds" yaml:"public_query_rate_limit_window_seconds"`
	PublicQueryMaxQuestionRunes   int     `mapstructure:"public_query_max_question_runes" yaml:"public_query_max_question_runes"`
	PublicQueryMaxTopK            int     `mapstructure:"public_query_max_top_k" yaml:"public_query_max_top_k"`
	PublicQueryContextMaxTurns    int     `mapstructure:"public_query_context_max_turns" yaml:"public_query_context_max_turns"`
	PublicQueryContextMaxRunes    int     `mapstructure:"public_query_context_max_runes" yaml:"public_query_context_max_runes"`
	OllamaMaxConcurrency          int     `mapstructure:"ollama_max_concurrency" yaml:"ollama_max_concurrency"`
	OllamaQueryQueueSize          int     `mapstructure:"ollama_query_queue_size" yaml:"ollama_query_queue_size"`
	OllamaIngestQueueSize         int     `mapstructure:"ollama_ingest_queue_size" yaml:"ollama_ingest_queue_size"`
	OllamaQueueWaitTimeoutSeconds int     `mapstructure:"ollama_queue_wait_timeout_seconds" yaml:"ollama_queue_wait_timeout_seconds"`
	TencentBaseURL                string  `mapstructure:"tencent_base_url" yaml:"tencent_base_url"`
	TencentModel                  string  `mapstructure:"tencent_model" yaml:"tencent_model"`
	TencentAPIKey                 string  `mapstructure:"tencent_api_key" yaml:"tencent_api_key"`
	TencentTimeoutSeconds         int     `mapstructure:"tencent_timeout_seconds" yaml:"tencent_timeout_seconds"`
	TencentTemperature            float64 `mapstructure:"tencent_temperature" yaml:"tencent_temperature"`
	TencentTopP                   float64 `mapstructure:"tencent_top_p" yaml:"tencent_top_p"`
	TencentMaxTokens              int     `mapstructure:"tencent_max_tokens" yaml:"tencent_max_tokens"`
	TencentReasoningEffort        string  `mapstructure:"tencent_reasoning_effort" yaml:"tencent_reasoning_effort"`
	IngestWorkers                 int     `mapstructure:"ingest_workers" yaml:"ingest_workers"`
	EmbedBatchSize                int     `mapstructure:"embed_batch_size" yaml:"embed_batch_size"`
	SyncEnabled                   bool    `mapstructure:"sync_enabled" yaml:"sync_enabled"`
	SyncIntervalMinutes           int     `mapstructure:"sync_interval_minutes" yaml:"sync_interval_minutes"`
	SyncTimeoutSeconds            int     `mapstructure:"sync_timeout_seconds" yaml:"sync_timeout_seconds"`
	SyncNavBaseURL                string  `mapstructure:"sync_nav_base_url" yaml:"sync_nav_base_url"`
	SyncGameBaseURL               string  `mapstructure:"sync_game_base_url" yaml:"sync_game_base_url"`
}

func ConfigureServerConfig(projectName, fileName, configFile string) {
	configOptionsMu.Lock()
	defer configOptionsMu.Unlock()
	if configuration != nil {
		return
	}
	if projectName = strings.TrimSpace(projectName); projectName != "" {
		configOptions.projectName = projectName
	}
	if fileName = strings.TrimSpace(fileName); fileName != "" {
		configOptions.fileName = fileName
	}
	configOptions.configFile = strings.TrimSpace(configFile)
}

func InitServerConfig(projectName string) error {
	opts := currentConfigOptions()
	ConfigureServerConfig(projectName, opts.fileName, opts.configFile)
	ensureServerConfig()
	return configErr
}

func MustInitServerConfig(projectName, configFile string) error {
	ConfigureServerConfig(projectName, "server.yaml", configFile)
	ensureServerConfig()
	return configErr
}

func InitConfig(projectName, fileName, configFile string, conf any) error {
	v := viper.New()
	configFile = strings.TrimSpace(configFile)
	if configFile != "" {
		v.SetConfigFile(configFile)
		if ext := strings.TrimPrefix(filepath.Ext(configFile), "."); ext != "" {
			v.SetConfigType(ext)
		}
	} else {
		configName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		configType := strings.TrimPrefix(filepath.Ext(fileName), ".")
		if configName == "" {
			configName = fileName
		}
		if configType == "" {
			configType = "yaml"
		}
		v.SetConfigName(configName)
		v.SetConfigType(configType)
		v.AddConfigPath(filepath.Join("/etc", projectName))
		if pwd, err := os.Getwd(); err == nil {
			v.AddConfigPath(filepath.Join(pwd, "config"))
		}
	}

	applyDefaults(v)
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("can not find any %s file: %w", fileName, err)
	}
	if err := v.Unmarshal(conf); err != nil {
		return fmt.Errorf("unmarshal config failed: %w", err)
	}
	return nil
}

func GetServerConfig() *Config {
	ensureServerConfig()
	if configuration != nil {
		return configuration
	}
	cfg := new(Config)
	cfg.normalize()
	return cfg
}

func ResolveConfigFile(projectName, fileName, configFile string) (string, error) {
	configFile = strings.TrimSpace(configFile)
	if configFile != "" {
		return filepath.Abs(configFile)
	}
	if pwd, err := os.Getwd(); err == nil {
		candidate := filepath.Join(pwd, "config", fileName)
		if _, statErr := os.Stat(candidate); statErr == nil {
			return candidate, nil
		}
	}
	candidate := filepath.Join("/etc", projectName, fileName)
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}
	return "", fmt.Errorf("can not find any %s file", fileName)
}

func UpdateConsolePasscode(configPath, password string) error {
	password = strings.TrimSpace(password)
	if password == "" {
		return fmt.Errorf("password is required")
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	var root yaml.Node
	if err := yaml.Unmarshal(content, &root); err != nil {
		return err
	}
	if len(root.Content) == 0 {
		return fmt.Errorf("invalid yaml document")
	}
	doc := root.Content[0]
	authNode := mappingChild(doc, "auth")
	if authNode == nil {
		authNode = &yaml.Node{Kind: yaml.MappingNode}
		doc.Content = append(doc.Content, scalarNode("auth"), authNode)
	}
	setMappingScalar(authNode, "console_passcode", password)
	encoded, err := yaml.Marshal(&root)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, encoded, 0o600)
}

func ResetForTest() {
	configOptionsMu.Lock()
	defer configOptionsMu.Unlock()
	configuration = nil
	configErr = nil
	configOnce = sync.Once{}
	configOptions = configLoaderOptions{projectName: common.COMMON_PROJECT_NAME, fileName: "server.yaml"}
}

func ensureServerConfig() {
	configOnce.Do(func() {
		opts := currentConfigOptions()
		cfg := new(Config)
		if err := InitConfig(opts.projectName, opts.fileName, opts.configFile, cfg); err != nil {
			configErr = err
			return
		}
		cfg.normalize()
		if err := cfg.validate(); err != nil {
			configErr = err
			return
		}
		configuration = cfg
	})
}

func currentConfigOptions() configLoaderOptions {
	configOptionsMu.Lock()
	defer configOptionsMu.Unlock()
	return configOptions
}

func (cfg *Config) normalize() {
	if cfg.ClusterID == 0 {
		cfg.ClusterID = 1
	}
	if cfg.Server.AppID == "" {
		cfg.Server.AppID = common.COMMON_PROJECT_NAME
	}
	if cfg.Server.AppName == "" {
		cfg.Server.AppName = "gofurry RAG"
	}
	if cfg.Server.AppVersion == "" {
		cfg.Server.AppVersion = "v1.0.0"
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}
	if cfg.Server.IPAddress == "" {
		cfg.Server.IPAddress = "127.0.0.1"
	}
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.Server.MemoryLimit <= 0 {
		cfg.Server.MemoryLimit = 1
	}
	if cfg.Server.GCPercent <= 0 {
		cfg.Server.GCPercent = 1000
	}
	if cfg.Server.Network == "" {
		cfg.Server.Network = "tcp"
	}
	if cfg.Server.ProxyHeader == "" {
		cfg.Server.ProxyHeader = "X-Forwarded-For"
	}
	cfg.Server.IsFullStack = true

	cfg.Database.normalize()
	if cfg.Log.LogLevel == "" {
		cfg.Log.LogLevel = "debug"
	}
	if cfg.Log.LogMode == "" {
		cfg.Log.LogMode = "dev"
	}
	if cfg.Log.LogPath == "" {
		cfg.Log.LogPath = "./logs/gofurry-rag.log"
	}
	if cfg.Middleware.RequestID.Header == "" {
		cfg.Middleware.RequestID.Header = "X-Request-ID"
	}
	if cfg.Middleware.AccessLog.Format == "" {
		cfg.Middleware.AccessLog.Format = "${time} | ${status} | ${latency} | ${method} | ${path} | rid=${respHeader:X-Request-ID}"
	}
	if cfg.Middleware.AccessLog.TimeFormat == "" {
		cfg.Middleware.AccessLog.TimeFormat = common.TIME_FORMAT_LOG
	}
	if cfg.Middleware.AccessLog.TimeZone == "" {
		cfg.Middleware.AccessLog.TimeZone = "Local"
	}
	if len(cfg.Middleware.Cors.AllowOrigins) == 0 {
		cfg.Middleware.Cors.AllowOrigins = []string{"http://127.0.0.1:5173", "http://127.0.0.1:8080"}
	}
	if cfg.Middleware.Timeout.DurationSeconds <= 0 {
		cfg.Middleware.Timeout.DurationSeconds = 30
	}
	if cfg.Middleware.Compression.Level == "" {
		cfg.Middleware.Compression.Level = "default"
	}
	if cfg.Middleware.Limiter.Strategy == "" {
		cfg.Middleware.Limiter.Strategy = "fixed"
	}
	if cfg.Middleware.Limiter.KeySource == "" {
		cfg.Middleware.Limiter.KeySource = "ip"
	}
	if cfg.Auth.ConsolePasscode == "" {
		cfg.Auth.ConsolePasscode = defaultConsolePasscode
	}
	if cfg.Auth.JWTSecret == "" {
		cfg.Auth.JWTSecret = defaultJWTSecret
	}
	if cfg.Auth.CookieName == "" {
		cfg.Auth.CookieName = "gofurry_rag_session"
	}
	if cfg.Auth.CookieMaxAgeSecs <= 0 {
		cfg.Auth.CookieMaxAgeSecs = 12 * 3600
	}
	if cfg.Auth.SessionTTLHours <= 0 {
		cfg.Auth.SessionTTLHours = 12
	}
	if cfg.Auth.SameSite == "" {
		cfg.Auth.SameSite = "Lax"
	}
	cfg.RAG.normalize()
	cfg.fillCompatibilityFields()
}

func (cfg *Config) validate() error {
	var errs []error
	switch cfg.Server.Mode {
	case "debug", "release", "prod":
	default:
		errs = append(errs, fmt.Errorf("server.mode must be one of debug, release, prod"))
	}
	if port, err := strconv.Atoi(cfg.Server.Port); err != nil || port <= 0 || port > 65535 {
		errs = append(errs, fmt.Errorf("server.port must be a valid port"))
	}
	if cfg.Database.Enabled {
		if err := cfg.Database.validate("database"); err != nil {
			errs = append(errs, err)
		}
	}
	if strings.TrimSpace(cfg.Auth.ConsolePasscode) == "" {
		errs = append(errs, fmt.Errorf("auth.console_passcode is required"))
	}
	if strings.TrimSpace(cfg.Auth.JWTSecret) == "" {
		errs = append(errs, fmt.Errorf("auth.jwt_secret is required"))
	}
	if cfg.Server.Mode != "debug" {
		if strings.TrimSpace(cfg.Auth.ConsolePasscode) == defaultConsolePasscode {
			errs = append(errs, fmt.Errorf("auth.console_passcode must be changed outside debug mode"))
		}
		if strings.TrimSpace(cfg.Auth.JWTSecret) == defaultJWTSecret {
			errs = append(errs, fmt.Errorf("auth.jwt_secret must be changed outside debug mode"))
		}
		if !cfg.Auth.CookieSecure {
			errs = append(errs, fmt.Errorf("auth.cookie_secure must be true outside debug mode"))
		}
	}
	if cfg.RAG.EmbedDim <= 0 {
		errs = append(errs, fmt.Errorf("rag.embed_dim must be positive"))
	} else if cfg.RAG.EmbedDim != 1024 {
		errs = append(errs, fmt.Errorf("rag.embed_dim must be 1024 with the current rag_chunks.embedding schema"))
	}
	if cfg.Server.TrustProxy && len(cfg.Server.TrustedProxies) == 0 && !cfg.Server.TrustProxyLoopback && !cfg.Server.TrustProxyPrivate && !cfg.Server.TrustProxyLinkLocal {
		errs = append(errs, fmt.Errorf("server.trusted_proxies or a trust_proxy_* range is required when server.trust_proxy is true"))
	}
	if cfg.RAG.SyncEnabled && strings.TrimSpace(cfg.RAG.SyncNavBaseURL) == "" {
		errs = append(errs, fmt.Errorf("rag.sync_nav_base_url is required when rag.sync_enabled is true"))
	}
	if cfg.RAG.SyncEnabled && strings.TrimSpace(cfg.RAG.SyncGameBaseURL) == "" {
		errs = append(errs, fmt.Errorf("rag.sync_game_base_url is required when rag.sync_enabled is true"))
	}
	return errors.Join(errs...)
}

func (cfg *Config) fillCompatibilityFields() {
	cfg.AppName = cfg.Server.AppName
	cfg.AppEnv = cfg.Server.Mode
	cfg.AppAddr = cfg.Server.IPAddress + ":" + cfg.Server.Port
	cfg.AdminToken = cfg.Auth.ConsolePasscode
	cfg.ConsolePasscode = cfg.Auth.ConsolePasscode
	cfg.JWTSecret = cfg.Auth.JWTSecret
	cfg.AuthCookieName = cfg.Auth.CookieName
	cfg.SessionTTLHours = cfg.Auth.SessionTTLHours
	cfg.DatabaseDSN = cfg.Database.BuildDSN()
	cfg.OllamaBaseURL = cfg.RAG.OllamaBaseURL
	cfg.EmbedModel = cfg.RAG.EmbedModel
	cfg.EmbedDim = cfg.RAG.EmbedDim
	cfg.ChunkSize = cfg.RAG.ChunkSize
	cfg.ChunkOverlap = cfg.RAG.ChunkOverlap
	cfg.TopK = cfg.RAG.TopK
	cfg.QueryTimeoutSeconds = cfg.RAG.QueryTimeoutSeconds
	cfg.EmbedTimeoutSeconds = cfg.RAG.EmbedTimeoutSeconds
	cfg.IngestTimeoutSeconds = cfg.RAG.IngestTimeoutSeconds
	cfg.MaxQueryQuestionRunes = cfg.RAG.MaxQueryQuestionRunes
	cfg.MaxQueryTopK = cfg.RAG.MaxQueryTopK
	cfg.PublicQueryRateLimitRequests = cfg.RAG.PublicQueryRateLimitRequests
	cfg.PublicQueryRateLimitWindowSec = cfg.RAG.PublicQueryRateLimitWindowSec
	cfg.PublicQueryMaxQuestionRunes = cfg.RAG.PublicQueryMaxQuestionRunes
	cfg.PublicQueryMaxTopK = cfg.RAG.PublicQueryMaxTopK
	cfg.PublicQueryContextMaxTurns = cfg.RAG.PublicQueryContextMaxTurns
	cfg.PublicQueryContextMaxRunes = cfg.RAG.PublicQueryContextMaxRunes
	cfg.OllamaMaxConcurrency = cfg.RAG.OllamaMaxConcurrency
	cfg.OllamaQueryQueueSize = cfg.RAG.OllamaQueryQueueSize
	cfg.OllamaIngestQueueSize = cfg.RAG.OllamaIngestQueueSize
	cfg.OllamaQueueWaitTimeoutSeconds = cfg.RAG.OllamaQueueWaitTimeoutSeconds
	cfg.TencentBaseURL = cfg.RAG.TencentBaseURL
	cfg.TencentModel = cfg.RAG.TencentModel
	cfg.TencentAPIKey = cfg.RAG.TencentAPIKey
	cfg.TencentTimeoutSeconds = cfg.RAG.TencentTimeoutSeconds
	cfg.TencentTemperature = cfg.RAG.TencentTemperature
	cfg.TencentTopP = cfg.RAG.TencentTopP
	cfg.TencentMaxTokens = cfg.RAG.TencentMaxTokens
	cfg.TencentReasoningEffort = cfg.RAG.TencentReasoningEffort
	cfg.IngestWorkers = cfg.RAG.IngestWorkers
	cfg.EmbedBatchSize = cfg.RAG.EmbedBatchSize
	cfg.SyncEnabled = cfg.RAG.SyncEnabled
	cfg.SyncIntervalMinutes = cfg.RAG.SyncIntervalMinutes
	cfg.SyncTimeoutSeconds = cfg.RAG.SyncTimeoutSeconds
	cfg.SyncNavBaseURL = cfg.RAG.SyncNavBaseURL
	cfg.SyncGameBaseURL = cfg.RAG.SyncGameBaseURL
}

func (cfg *DatabaseConfig) normalize() {
	cfg.DBType = strings.ToLower(strings.TrimSpace(cfg.DBType))
	if cfg.DBType == "" {
		cfg.DBType = "postgres"
	}
	cfg.applyLegacyConfig()
	normalizeSQLDefaults(&cfg.Postgres, SQLDatabaseConfig{
		DBName: "gfr",
		DBHost: "192.168.153.121",
		DBPort: "5432",
		DBUser: "postgres",
		DBPass: "your_password",
	})
}

func (cfg *DatabaseConfig) validate(prefix string) error {
	if cfg.DBType != "postgres" && cfg.DBType != "postgresql" {
		return fmt.Errorf("%s.db_type %q is not supported", prefix, cfg.DBType)
	}
	if strings.TrimSpace(cfg.Postgres.DSN) == "" && strings.TrimSpace(cfg.Postgres.DBName) == "" {
		return fmt.Errorf("%s.postgres db config is required", prefix)
	}
	return nil
}

func (cfg *DatabaseConfig) BuildDSN() string {
	if dsn := strings.TrimSpace(cfg.Postgres.DSN); dsn != "" {
		return dsn
	}
	user := url.QueryEscape(strings.TrimSpace(cfg.Postgres.DBUser))
	pass := url.QueryEscape(strings.TrimSpace(cfg.Postgres.DBPass))
	host := strings.TrimSpace(cfg.Postgres.DBHost)
	port := strings.TrimSpace(cfg.Postgres.DBPort)
	dbName := strings.TrimSpace(cfg.Postgres.DBName)
	if port != "" {
		host += ":" + port
	}
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, pass, host, url.PathEscape(dbName))
}

func (cfg *DatabaseConfig) applyLegacyConfig() {
	if cfg.Postgres.DSN == "" {
		cfg.Postgres.DSN = strings.TrimSpace(cfg.DSN)
	}
	if cfg.Postgres.DBName == "" {
		cfg.Postgres.DBName = strings.TrimSpace(cfg.DBName)
	}
	if cfg.Postgres.DBHost == "" {
		cfg.Postgres.DBHost = strings.TrimSpace(cfg.DBHost)
	}
	if cfg.Postgres.DBPort == "" {
		cfg.Postgres.DBPort = strings.TrimSpace(cfg.DBPort)
	}
	if cfg.Postgres.DBUser == "" {
		cfg.Postgres.DBUser = strings.TrimSpace(cfg.DBUser)
	}
	if cfg.Postgres.DBPass == "" {
		cfg.Postgres.DBPass = strings.TrimSpace(cfg.DBPass)
	}
}

func (cfg *RAGConfig) normalize() {
	if cfg.OllamaBaseURL == "" {
		cfg.OllamaBaseURL = "http://148.70.18.111:43434"
	}
	if cfg.EmbedModel == "" {
		cfg.EmbedModel = "qwen3-embedding:0.6b"
	}
	if cfg.EmbedDim <= 0 {
		cfg.EmbedDim = 1024
	}
	if cfg.ChunkSize <= 0 {
		cfg.ChunkSize = 700
	}
	if cfg.ChunkOverlap <= 0 {
		cfg.ChunkOverlap = 120
	}
	if cfg.TopK <= 0 {
		cfg.TopK = 6
	}
	if cfg.QueryTimeoutSeconds <= 0 {
		cfg.QueryTimeoutSeconds = 60
	}
	if cfg.EmbedTimeoutSeconds <= 0 {
		cfg.EmbedTimeoutSeconds = 60
	}
	if cfg.IngestTimeoutSeconds <= 0 {
		cfg.IngestTimeoutSeconds = 300
	}
	if cfg.MaxQueryQuestionRunes <= 0 {
		cfg.MaxQueryQuestionRunes = 4000
	}
	if cfg.MaxQueryTopK <= 0 {
		cfg.MaxQueryTopK = 12
	}
	if cfg.PublicQueryRateLimitRequests <= 0 {
		cfg.PublicQueryRateLimitRequests = 10
	}
	if cfg.PublicQueryRateLimitWindowSec <= 0 {
		cfg.PublicQueryRateLimitWindowSec = 60
	}
	if cfg.PublicQueryMaxQuestionRunes <= 0 {
		cfg.PublicQueryMaxQuestionRunes = 800
	}
	if cfg.PublicQueryMaxTopK <= 0 {
		cfg.PublicQueryMaxTopK = 6
	}
	if cfg.PublicQueryContextMaxTurns <= 0 {
		cfg.PublicQueryContextMaxTurns = 3
	}
	if cfg.PublicQueryContextMaxRunes <= 0 {
		cfg.PublicQueryContextMaxRunes = 8000
	}
	if cfg.OllamaMaxConcurrency <= 0 {
		cfg.OllamaMaxConcurrency = 4
	}
	if cfg.OllamaQueryQueueSize <= 0 {
		cfg.OllamaQueryQueueSize = 8
	}
	if cfg.OllamaIngestQueueSize <= 0 {
		cfg.OllamaIngestQueueSize = 32
	}
	if cfg.OllamaQueueWaitTimeoutSeconds <= 0 {
		cfg.OllamaQueueWaitTimeoutSeconds = 30
	}
	if cfg.TencentBaseURL == "" {
		cfg.TencentBaseURL = "https://tokenhub.tencentmaas.com/v1"
	}
	if cfg.TencentModel == "" {
		cfg.TencentModel = "deepseek-v4-flash"
	}
	if cfg.TencentTimeoutSeconds <= 0 {
		cfg.TencentTimeoutSeconds = 60
	}
	if cfg.TencentTemperature == 0 {
		cfg.TencentTemperature = 0.2
	}
	if cfg.TencentTopP == 0 {
		cfg.TencentTopP = 0.8
	}
	if cfg.TencentMaxTokens <= 0 {
		cfg.TencentMaxTokens = 1024
	}
	if cfg.TencentReasoningEffort == "" {
		cfg.TencentReasoningEffort = "low"
	}
	if cfg.IngestWorkers <= 0 {
		cfg.IngestWorkers = 1
	}
	if cfg.EmbedBatchSize <= 0 {
		cfg.EmbedBatchSize = 8
	}
	if cfg.SyncIntervalMinutes <= 0 {
		cfg.SyncIntervalMinutes = 12 * 60
	}
	if cfg.SyncTimeoutSeconds <= 0 {
		cfg.SyncTimeoutSeconds = 30
	}
	cfg.SyncNavBaseURL = strings.TrimRight(strings.TrimSpace(cfg.SyncNavBaseURL), "/")
	cfg.SyncGameBaseURL = strings.TrimRight(strings.TrimSpace(cfg.SyncGameBaseURL), "/")
}

func normalizeSQLDefaults(target *SQLDatabaseConfig, defaults SQLDatabaseConfig) {
	if target.DBHost == "" {
		target.DBHost = defaults.DBHost
	}
	if target.DBPort == "" {
		target.DBPort = defaults.DBPort
	}
	if target.DBName == "" {
		target.DBName = defaults.DBName
	}
	if target.DBUser == "" {
		target.DBUser = defaults.DBUser
	}
	if target.DBPass == "" {
		target.DBPass = defaults.DBPass
	}
}

func applyDefaults(v *viper.Viper) {
	v.SetDefault("cluster_id", 1)
	v.SetDefault("server.app_id", common.COMMON_PROJECT_NAME)
	v.SetDefault("server.app_name", "gofurry RAG")
	v.SetDefault("server.app_version", "v1.0.0")
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.ip_address", "127.0.0.1")
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.memory_limit", 1)
	v.SetDefault("server.gc_percent", 1000)
	v.SetDefault("server.network", "tcp")
	v.SetDefault("server.enable_prefork", false)
	v.SetDefault("server.is_full_stack", true)
	v.SetDefault("server.trust_proxy", false)
	v.SetDefault("server.proxy_header", "X-Forwarded-For")
	v.SetDefault("server.trusted_proxies", []string{})
	v.SetDefault("server.trust_proxy_loopback", false)
	v.SetDefault("server.trust_proxy_private", false)
	v.SetDefault("server.trust_proxy_link_local", false)
	v.SetDefault("database.enabled", true)
	v.SetDefault("database.auto_migrate", true)
	v.SetDefault("database.db_type", "postgres")
	v.SetDefault("database.postgres.db_name", "gfr")
	v.SetDefault("database.postgres.db_host", "192.168.153.121")
	v.SetDefault("database.postgres.db_port", "5432")
	v.SetDefault("database.postgres.db_username", "postgres")
	v.SetDefault("database.postgres.db_password", "your_password")
	v.SetDefault("log.log_level", "debug")
	v.SetDefault("log.log_mode", "dev")
	v.SetDefault("log.log_path", "./logs/gofurry-rag.log")
	v.SetDefault("middleware.request_id.enabled", true)
	v.SetDefault("middleware.request_id.header", "X-Request-ID")
	v.SetDefault("middleware.access_log.enabled", true)
	v.SetDefault("middleware.access_log.format", "${time} | ${status} | ${latency} | ${method} | ${path} | rid=${respHeader:X-Request-ID}")
	v.SetDefault("middleware.access_log.time_format", common.TIME_FORMAT_LOG)
	v.SetDefault("middleware.access_log.time_zone", "Local")
	v.SetDefault("middleware.timeout.enabled", true)
	v.SetDefault("middleware.timeout.duration_seconds", 30)
	v.SetDefault("middleware.timeout.exclude_paths", []string{"/livez", "/readyz", "/startupz", "/healthz", "/api/v1/chat/stream"})
	v.SetDefault("middleware.health.enabled", true)
	v.SetDefault("middleware.health.include_legacy", true)
	v.SetDefault("middleware.security_headers.enabled", true)
	v.SetDefault("middleware.compression.enabled", true)
	v.SetDefault("middleware.compression.level", "default")
	v.SetDefault("middleware.cors.allow_origins", []string{"http://127.0.0.1:5173", "http://127.0.0.1:8080"})
	v.SetDefault("middleware.limiter.enabled", true)
	v.SetDefault("middleware.limiter.max_requests", 3000)
	v.SetDefault("middleware.limiter.expiration", 60)
	v.SetDefault("middleware.limiter.strategy", "fixed")
	v.SetDefault("middleware.limiter.key_source", "ip")
	v.SetDefault("middleware.limiter.exclude_paths", []string{"/livez", "/readyz", "/startupz", "/healthz", "/api/v1/chat/stream"})
	v.SetDefault("middleware.etag.enabled", true)
	v.SetDefault("waf.enabled", false)
	v.SetDefault("waf.conf_path", []string{"./config/coraza.conf"})
	v.SetDefault("auth.console_passcode", defaultConsolePasscode)
	v.SetDefault("auth.jwt_secret", defaultJWTSecret)
	v.SetDefault("auth.cookie_name", "gofurry_rag_session")
	v.SetDefault("auth.cookie_domain", "")
	v.SetDefault("auth.cookie_secure", false)
	v.SetDefault("auth.cookie_max_age_seconds", 43200)
	v.SetDefault("auth.session_ttl_hours", 12)
	v.SetDefault("auth.cookie_same_site", "Lax")
	v.SetDefault("auth.csrf_header_required", false)
	v.SetDefault("rag.ollama_base_url", "http://148.70.18.111:43434")
	v.SetDefault("rag.embed_model", "qwen3-embedding:0.6b")
	v.SetDefault("rag.embed_dim", 1024)
	v.SetDefault("rag.chunk_size", 700)
	v.SetDefault("rag.chunk_overlap", 120)
	v.SetDefault("rag.top_k", 6)
	v.SetDefault("rag.query_timeout_seconds", 60)
	v.SetDefault("rag.embed_timeout_seconds", 60)
	v.SetDefault("rag.ingest_timeout_seconds", 300)
	v.SetDefault("rag.max_query_question_runes", 4000)
	v.SetDefault("rag.max_query_top_k", 12)
	v.SetDefault("rag.public_query_rate_limit_requests", 10)
	v.SetDefault("rag.public_query_rate_limit_window_seconds", 60)
	v.SetDefault("rag.public_query_max_question_runes", 800)
	v.SetDefault("rag.public_query_max_top_k", 6)
	v.SetDefault("rag.public_query_context_max_turns", 3)
	v.SetDefault("rag.public_query_context_max_runes", 8000)
	v.SetDefault("rag.ollama_max_concurrency", 4)
	v.SetDefault("rag.ollama_query_queue_size", 8)
	v.SetDefault("rag.ollama_ingest_queue_size", 32)
	v.SetDefault("rag.ollama_queue_wait_timeout_seconds", 30)
	v.SetDefault("rag.tencent_base_url", "https://tokenhub.tencentmaas.com/v1")
	v.SetDefault("rag.tencent_model", "deepseek-v4-flash")
	v.SetDefault("rag.tencent_api_key", "")
	v.SetDefault("rag.tencent_timeout_seconds", 60)
	v.SetDefault("rag.tencent_temperature", 0.2)
	v.SetDefault("rag.tencent_top_p", 0.8)
	v.SetDefault("rag.tencent_max_tokens", 1024)
	v.SetDefault("rag.tencent_reasoning_effort", "low")
	v.SetDefault("rag.ingest_workers", 1)
	v.SetDefault("rag.embed_batch_size", 8)
	v.SetDefault("rag.sync_enabled", false)
	v.SetDefault("rag.sync_interval_minutes", 720)
	v.SetDefault("rag.sync_timeout_seconds", 30)
	v.SetDefault("rag.sync_nav_base_url", "")
	v.SetDefault("rag.sync_game_base_url", "")
}

func mappingChild(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func setMappingScalar(node *yaml.Node, key, value string) {
	if node.Kind != yaml.MappingNode {
		node.Kind = yaml.MappingNode
		node.Content = nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			node.Content[i+1] = scalarNode(value)
			return
		}
	}
	node.Content = append(node.Content, scalarNode(key), scalarNode(value))
}

func scalarNode(value string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value}
}
