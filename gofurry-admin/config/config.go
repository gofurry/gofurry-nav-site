package env

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/gofurry/awesome-fiber-template/v3/medium/pkg/common"
	"github.com/spf13/viper"
)

var (
	configuration *serverConfig
	configErr     error
	configOnce    sync.Once
	configOptions = configLoaderOptions{
		projectName: common.COMMON_PROJECT_NAME,
		fileName:    "server.yaml",
	}
	configOptionsMu sync.Mutex
)

type configLoaderOptions struct {
	projectName string
	fileName    string
	configFile  string
}

type serverConfig struct {
	ClusterId         int                     `mapstructure:"cluster_id" yaml:"cluster_id"`
	Server            ServerConfig            `mapstructure:"server" yaml:"server"`
	DataBase          DataBaseConfig          `mapstructure:"database" yaml:"database"`
	BusinessDatabases BusinessDatabasesConfig `mapstructure:"business_databases" yaml:"business_databases"`
	Log               LogConfig               `mapstructure:"log" yaml:"log"`
	Redis             RedisConfig             `mapstructure:"redis" yaml:"redis"`
	Middleware        MiddlewareConfig        `mapstructure:"middleware" yaml:"middleware"`
	Waf               WafConfig               `mapstructure:"waf" yaml:"waf"`
	Auth              AuthConfig              `mapstructure:"auth" yaml:"auth"`
	ExternalServices  ExternalServicesConfig  `mapstructure:"external_services" yaml:"external_services"`
}

type ServerConfigHolder = serverConfig

type ExternalServicesConfig struct {
	GameBackend BackendServiceConfig `mapstructure:"game_backend" yaml:"game_backend"`
	NavBackend  BackendServiceConfig `mapstructure:"nav_backend" yaml:"nav_backend"`
}

type BackendServiceConfig struct {
	BaseURL          string `mapstructure:"base_url" yaml:"base_url"`
	AdminToken       string `mapstructure:"admin_token" yaml:"admin_token"`
	AdminTokenHeader string `mapstructure:"admin_token_header" yaml:"admin_token_header"`
	TimeoutSeconds   int    `mapstructure:"timeout_seconds" yaml:"timeout_seconds"`
}

type BusinessDatabasesConfig struct {
	Nav  DataBaseConfig `mapstructure:"nav" yaml:"nav"`
	Game DataBaseConfig `mapstructure:"game" yaml:"game"`
}

type AuthConfig struct {
	JWTSecret          string `mapstructure:"jwt_secret" yaml:"jwt_secret"`
	CookieName         string `mapstructure:"cookie_name" yaml:"cookie_name"`
	CookieDomain       string `mapstructure:"cookie_domain" yaml:"cookie_domain"`
	CookieSecure       bool   `mapstructure:"cookie_secure" yaml:"cookie_secure"`
	CookieMaxAgeSecs   int    `mapstructure:"cookie_max_age_seconds" yaml:"cookie_max_age_seconds"`
	PBKDF2Iterations   int    `mapstructure:"pbkdf2_iterations" yaml:"pbkdf2_iterations"`
	SessionTTLHours    int    `mapstructure:"session_ttl_hours" yaml:"session_ttl_hours"`
	SameSite           string `mapstructure:"cookie_same_site" yaml:"cookie_same_site"`
	CSRFHeaderRequired bool   `mapstructure:"csrf_header_required" yaml:"csrf_header_required"`
}

type WafConfig struct {
	Enabled  bool     `mapstructure:"enabled" yaml:"enabled"`
	ConfPath []string `mapstructure:"conf_path" yaml:"conf_path"`
}

type MiddlewareConfig struct {
	Swagger         SwaggerConfig         `mapstructure:"swagger" yaml:"swagger"`
	Cors            CorsConfig            `mapstructure:"cors" yaml:"cors"`
	RequestID       RequestIDConfig       `mapstructure:"request_id" yaml:"request_id"`
	AccessLog       AccessLogConfig       `mapstructure:"access_log" yaml:"access_log"`
	Timeout         TimeoutConfig         `mapstructure:"timeout" yaml:"timeout"`
	Health          HealthConfig          `mapstructure:"health" yaml:"health"`
	SecurityHeaders SecurityHeadersConfig `mapstructure:"security_headers" yaml:"security_headers"`
	Compression     CompressionConfig     `mapstructure:"compression" yaml:"compression"`
	Limiter         LimiterConfig         `mapstructure:"limiter" yaml:"limiter"`
	CSRF            CSRFConfig            `mapstructure:"csrf" yaml:"csrf"`
	ETag            ETagConfig            `mapstructure:"etag" yaml:"etag"`
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

type CSRFConfig struct {
	Enabled            bool     `mapstructure:"enabled" yaml:"enabled"`
	TokenPath          string   `mapstructure:"token_path" yaml:"token_path"`
	CookieName         string   `mapstructure:"cookie_name" yaml:"cookie_name"`
	CookieSameSite     string   `mapstructure:"cookie_same_site" yaml:"cookie_same_site"`
	CookieSecure       bool     `mapstructure:"cookie_secure" yaml:"cookie_secure"`
	CookieHTTPOnly     bool     `mapstructure:"cookie_http_only" yaml:"cookie_http_only"`
	CookieSessionOnly  bool     `mapstructure:"cookie_session_only" yaml:"cookie_session_only"`
	IdleTimeoutSeconds int      `mapstructure:"idle_timeout_seconds" yaml:"idle_timeout_seconds"`
	SingleUseToken     bool     `mapstructure:"single_use_token" yaml:"single_use_token"`
	TrustedOrigins     []string `mapstructure:"trusted_origins" yaml:"trusted_origins"`
	ExcludePaths       []string `mapstructure:"exclude_paths" yaml:"exclude_paths"`
}

type ETagConfig struct {
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`
	Weak    bool `mapstructure:"weak" yaml:"weak"`
}

type CorsConfig struct {
	AllowOrigins []string `mapstructure:"allow_origins" yaml:"allow_origins"`
}

type SwaggerConfig struct {
	Enabled  bool   `mapstructure:"enabled" yaml:"enabled"`
	FilePath string `mapstructure:"file_path" yaml:"file_path"`
	BasePath string `mapstructure:"base_path" yaml:"base_path"`
	Path     string `mapstructure:"path" yaml:"path"`
	Title    string `mapstructure:"title" yaml:"title"`
}

type RedisConfig struct {
	Enabled       bool   `mapstructure:"enabled" yaml:"enabled"`
	RedisUsername string `mapstructure:"redis_username" yaml:"redis_username"`
	RedisAddr     string `mapstructure:"redis_addr" yaml:"redis_addr"`
	RedisPassword string `mapstructure:"redis_password" yaml:"redis_password"`
	RedisDB       int    `mapstructure:"redis_db" yaml:"redis_db"`
	RedisPoolSize int    `mapstructure:"redis_pool_size" yaml:"redis_pool_size"`
}

type LogConfig struct {
	LogLevel      string `mapstructure:"log_level" yaml:"log_level"`
	LogMode       string `mapstructure:"log_mode" yaml:"log_mode"`
	LogPath       string `mapstructure:"log_path" yaml:"log_path"`
	LogMaxSize    int    `mapstructure:"log_max_size" yaml:"log_max_size"`
	LogMaxBackups int    `mapstructure:"log_max_backups" yaml:"log_max_backups"`
	LogMaxAge     int    `mapstructure:"log_max_age" yaml:"log_max_age"`
}

type DataBaseConfig struct {
	Enabled     bool                 `mapstructure:"enabled" yaml:"enabled"`
	AutoMigrate bool                 `mapstructure:"auto_migrate" yaml:"auto_migrate"`
	DBType      string               `mapstructure:"db_type" yaml:"db_type"`
	SQLite      SQLiteDataBaseConfig `mapstructure:"sqlite" yaml:"sqlite"`
	Postgres    SQLDataBaseConfig    `mapstructure:"postgres" yaml:"postgres"`
	MySQL       SQLDataBaseConfig    `mapstructure:"mysql" yaml:"mysql"`
	DSN         string               `mapstructure:"dsn" yaml:"dsn"`
	DBName      string               `mapstructure:"db_name" yaml:"db_name"`
	DBHost      string               `mapstructure:"db_host" yaml:"db_host"`
	DBPort      string               `mapstructure:"db_port" yaml:"db_port"`
	DBUser      string               `mapstructure:"db_username" yaml:"db_username"`
	DBPass      string               `mapstructure:"db_password" yaml:"db_password"`
	SQLPath     string               `mapstructure:"sqlite_path" yaml:"sqlite_path"`
}

type SQLDataBaseConfig struct {
	DSN    string `mapstructure:"dsn" yaml:"dsn"`
	DBName string `mapstructure:"db_name" yaml:"db_name"`
	DBHost string `mapstructure:"db_host" yaml:"db_host"`
	DBPort string `mapstructure:"db_port" yaml:"db_port"`
	DBUser string `mapstructure:"db_username" yaml:"db_username"`
	DBPass string `mapstructure:"db_password" yaml:"db_password"`
}

type SQLiteDataBaseConfig struct {
	DSN  string `mapstructure:"dsn" yaml:"dsn"`
	Path string `mapstructure:"path" yaml:"path"`
}

type ServerConfig struct {
	AppID         string `mapstructure:"app_id" yaml:"app_id"`
	AppName       string `mapstructure:"app_name" yaml:"app_name"`
	AppVersion    string `mapstructure:"app_version" yaml:"app_version"`
	Mode          string `mapstructure:"mode" yaml:"mode"`
	IPAddress     string `mapstructure:"ip_address" yaml:"ip_address"`
	Port          string `mapstructure:"port" yaml:"port"`
	MemoryLimit   int    `mapstructure:"memory_limit" yaml:"memory_limit"`
	GCPercent     int    `mapstructure:"gc_percent" yaml:"gc_percent"`
	Network       string `mapstructure:"network" yaml:"network"`
	EnablePrefork bool   `mapstructure:"enable_prefork" yaml:"enable_prefork"`
	IsFullStack   bool   `mapstructure:"is_full_stack" yaml:"is_full_stack"`
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

func (cfg *serverConfig) normalize() {
	if cfg.ClusterId == 0 {
		cfg.ClusterId = 1
	}

	if cfg.Server.AppID == "" {
		cfg.Server.AppID = common.COMMON_PROJECT_NAME
	}
	if cfg.Server.AppName == "" {
		cfg.Server.AppName = "gofurry Admin"
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
		cfg.Server.Port = "10099"
	}
	if cfg.Server.Network == "" {
		cfg.Server.Network = "tcp"
	}
	cfg.Server.IsFullStack = true

	cfg.DataBase.normalize("gfa")
	cfg.DataBase.AutoMigrate = true
	cfg.BusinessDatabases.Nav.normalize("gfn")
	cfg.BusinessDatabases.Nav.DBType = "postgres"
	cfg.BusinessDatabases.Nav.Enabled = true
	cfg.BusinessDatabases.Nav.AutoMigrate = false
	cfg.BusinessDatabases.Game.normalize("gfg")
	cfg.BusinessDatabases.Game.DBType = "postgres"
	cfg.BusinessDatabases.Game.Enabled = true
	cfg.BusinessDatabases.Game.AutoMigrate = false

	if cfg.Middleware.Swagger.Title == "" {
		cfg.Middleware.Swagger.Title = cfg.Server.AppName
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
	if cfg.Middleware.Timeout.DurationSeconds <= 0 {
		cfg.Middleware.Timeout.DurationSeconds = 30
	}
	cfg.Middleware.Timeout.ExcludePaths = normalizeStringList(cfg.Middleware.Timeout.ExcludePaths)
	if cfg.Middleware.Compression.Level == "" {
		cfg.Middleware.Compression.Level = "default"
	}
	if cfg.Middleware.Limiter.Strategy == "" {
		cfg.Middleware.Limiter.Strategy = "fixed"
	}
	if cfg.Middleware.Limiter.KeySource == "" {
		cfg.Middleware.Limiter.KeySource = "ip"
	}
	cfg.Middleware.Limiter.ExcludePaths = normalizeStringList(cfg.Middleware.Limiter.ExcludePaths)
	if cfg.Middleware.CSRF.TokenPath == "" {
		cfg.Middleware.CSRF.TokenPath = "/csrf/token"
	}
	if cfg.Middleware.CSRF.CookieName == "" {
		cfg.Middleware.CSRF.CookieName = "csrf_"
	}
	if cfg.Middleware.CSRF.CookieSameSite == "" {
		cfg.Middleware.CSRF.CookieSameSite = "Lax"
	}
	if cfg.Middleware.CSRF.IdleTimeoutSeconds <= 0 {
		cfg.Middleware.CSRF.IdleTimeoutSeconds = 7200
	}
	cfg.Middleware.CSRF.Enabled = true
	cfg.Middleware.CSRF.CookieHTTPOnly = false
	cfg.Middleware.CSRF.ExcludePaths = normalizeStringList(append(cfg.Middleware.CSRF.ExcludePaths,
		"/api/v1/auth/bootstrap",
		"/api/v1/auth/login",
		"/api/v1/auth/logout",
	))

	if cfg.Auth.JWTSecret == "" {
		cfg.Auth.JWTSecret = "gofurry-admin-dev-secret"
	}
	if cfg.Auth.CookieName == "" {
		cfg.Auth.CookieName = "gofurry_admin_token"
	}
	if cfg.Auth.CookieMaxAgeSecs <= 0 {
		cfg.Auth.CookieMaxAgeSecs = 86400 * 30
	}
	if cfg.Auth.PBKDF2Iterations <= 0 {
		cfg.Auth.PBKDF2Iterations = 300000
	}
	if cfg.Auth.SessionTTLHours <= 0 {
		cfg.Auth.SessionTTLHours = 24 * 30
	}
	if cfg.Auth.SameSite == "" {
		cfg.Auth.SameSite = "Lax"
	}

	if cfg.ExternalServices.GameBackend.BaseURL == "" {
		cfg.ExternalServices.GameBackend.BaseURL = "http://127.0.0.1:9998"
	}
	if cfg.ExternalServices.GameBackend.AdminTokenHeader == "" {
		cfg.ExternalServices.GameBackend.AdminTokenHeader = "X-GoFurry-Admin-Token"
	}
	if cfg.ExternalServices.GameBackend.TimeoutSeconds <= 0 {
		cfg.ExternalServices.GameBackend.TimeoutSeconds = 10
	}
	if cfg.ExternalServices.NavBackend.BaseURL == "" {
		cfg.ExternalServices.NavBackend.BaseURL = "http://127.0.0.1:9999"
	}
	if cfg.ExternalServices.NavBackend.AdminTokenHeader == "" {
		cfg.ExternalServices.NavBackend.AdminTokenHeader = "X-GoFurry-Admin-Token"
	}
	if cfg.ExternalServices.NavBackend.TimeoutSeconds <= 0 {
		cfg.ExternalServices.NavBackend.TimeoutSeconds = 10
	}
}

func (cfg *serverConfig) validate() error {
	var errs []error

	switch cfg.Server.Mode {
	case "debug", "release", "prod":
	default:
		errs = append(errs, fmt.Errorf("server.mode must be one of debug, release, prod"))
	}

	if port, err := strconv.Atoi(cfg.Server.Port); err != nil || port <= 0 || port > 65535 {
		errs = append(errs, fmt.Errorf("server.port must be a valid port"))
	}
	if cfg.Auth.PBKDF2Iterations < 300000 {
		errs = append(errs, fmt.Errorf("auth.pbkdf2_iterations must be >= 300000"))
	}
	if strings.TrimSpace(cfg.Auth.JWTSecret) == "" {
		errs = append(errs, fmt.Errorf("auth.jwt_secret is required"))
	}
	if cfg.DataBase.Enabled {
		if err := cfg.DataBase.validate("database"); err != nil {
			errs = append(errs, err)
		}
	}
	if cfg.BusinessDatabases.Nav.Enabled {
		if err := cfg.BusinessDatabases.Nav.validate("business_databases.nav"); err != nil {
			errs = append(errs, err)
		}
	}
	if cfg.BusinessDatabases.Game.Enabled {
		if err := cfg.BusinessDatabases.Game.validate("business_databases.game"); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (cfg *DataBaseConfig) normalize(defaultDBName string) {
	cfg.DBType = strings.ToLower(strings.TrimSpace(cfg.DBType))
	if cfg.DBType == "" {
		cfg.DBType = "postgres"
	}

	cfg.applyLegacyConfig()

	if cfg.SQLite.Path == "" {
		cfg.SQLite.Path = "./data/app.db"
	}

	normalizeSQLDefaults(&cfg.Postgres, SQLDataBaseConfig{
		DBHost: "127.0.0.1",
		DBPort: "5432",
		DBName: defaultDBName,
		DBUser: "postgres",
		DBPass: "123456",
	})
	normalizeSQLDefaults(&cfg.MySQL, SQLDataBaseConfig{
		DBHost: "127.0.0.1",
		DBPort: "3306",
		DBName: defaultDBName,
		DBUser: "root",
		DBPass: "123456",
	})
}

func (cfg *DataBaseConfig) validate(prefix string) error {
	switch cfg.DBType {
	case "postgres", "postgresql", "mysql", "sqlite":
	default:
		return fmt.Errorf("%s.db_type %q is not supported", prefix, cfg.DBType)
	}
	switch cfg.DBType {
	case "sqlite":
		if strings.TrimSpace(cfg.SQLite.DSN) == "" && strings.TrimSpace(cfg.SQLite.Path) == "" {
			return fmt.Errorf("%s.sqlite.path or %s.sqlite.dsn is required", prefix, prefix)
		}
	case "mysql":
		if strings.TrimSpace(cfg.MySQL.DSN) == "" && strings.TrimSpace(cfg.MySQL.DBName) == "" {
			return fmt.Errorf("%s.mysql db config is required", prefix)
		}
	default:
		if strings.TrimSpace(cfg.Postgres.DSN) == "" && strings.TrimSpace(cfg.Postgres.DBName) == "" {
			return fmt.Errorf("%s.postgres db config is required", prefix)
		}
	}
	return nil
}

func (cfg *DataBaseConfig) applyLegacyConfig() {
	switch cfg.DBType {
	case "sqlite":
		if cfg.SQLite.DSN == "" {
			cfg.SQLite.DSN = strings.TrimSpace(cfg.DSN)
		}
		if cfg.SQLite.Path == "" {
			cfg.SQLite.Path = strings.TrimSpace(cfg.SQLPath)
		}
		if cfg.SQLite.Path == "" {
			cfg.SQLite.Path = strings.TrimSpace(cfg.DBName)
		}
	case "mysql":
		applyLegacySQLConfig(&cfg.MySQL, cfg)
	default:
		applyLegacySQLConfig(&cfg.Postgres, cfg)
	}
}

func applyLegacySQLConfig(target *SQLDataBaseConfig, legacy *DataBaseConfig) {
	if target.DSN == "" {
		target.DSN = strings.TrimSpace(legacy.DSN)
	}
	if target.DBName == "" {
		target.DBName = strings.TrimSpace(legacy.DBName)
	}
	if target.DBHost == "" {
		target.DBHost = strings.TrimSpace(legacy.DBHost)
	}
	if target.DBPort == "" {
		target.DBPort = strings.TrimSpace(legacy.DBPort)
	}
	if target.DBUser == "" {
		target.DBUser = strings.TrimSpace(legacy.DBUser)
	}
	if target.DBPass == "" {
		target.DBPass = strings.TrimSpace(legacy.DBPass)
	}
}

func normalizeSQLDefaults(target *SQLDataBaseConfig, defaults SQLDataBaseConfig) {
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

func normalizeStringList(items []string) []string {
	if len(items) == 0 {
		return nil
	}

	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func InitConfig(projectName, fileName, configFile string, conf interface{}) error {
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

func ensureServerConfig() {
	configOnce.Do(func() {
		opts := currentConfigOptions()
		cfg := new(serverConfig)
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

func applyDefaults(v *viper.Viper) {
	v.SetDefault("cluster_id", 1)
	v.SetDefault("server.app_id", common.COMMON_PROJECT_NAME)
	v.SetDefault("server.app_name", "gofurry Admin")
	v.SetDefault("server.app_version", "v1.0.0")
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.ip_address", "127.0.0.1")
	v.SetDefault("server.port", "10099")
	v.SetDefault("server.memory_limit", 1)
	v.SetDefault("server.gc_percent", 1000)
	v.SetDefault("server.network", "tcp")
	v.SetDefault("server.enable_prefork", false)
	v.SetDefault("server.is_full_stack", true)

	v.SetDefault("database.enabled", true)
	v.SetDefault("database.auto_migrate", true)
	v.SetDefault("database.db_type", "postgres")
	v.SetDefault("database.postgres.db_name", "gfa")
	v.SetDefault("database.postgres.db_host", "127.0.0.1")
	v.SetDefault("database.postgres.db_port", "5432")
	v.SetDefault("database.postgres.db_username", "postgres")
	v.SetDefault("database.postgres.db_password", "123456")

	v.SetDefault("business_databases.nav.enabled", true)
	v.SetDefault("business_databases.nav.db_type", "postgres")
	v.SetDefault("business_databases.nav.postgres.db_name", "gfn")
	v.SetDefault("business_databases.nav.postgres.db_host", "127.0.0.1")
	v.SetDefault("business_databases.nav.postgres.db_port", "5432")
	v.SetDefault("business_databases.nav.postgres.db_username", "postgres")
	v.SetDefault("business_databases.nav.postgres.db_password", "123456")

	v.SetDefault("business_databases.game.enabled", true)
	v.SetDefault("business_databases.game.db_type", "postgres")
	v.SetDefault("business_databases.game.postgres.db_name", "gfg")
	v.SetDefault("business_databases.game.postgres.db_host", "127.0.0.1")
	v.SetDefault("business_databases.game.postgres.db_port", "5432")
	v.SetDefault("business_databases.game.postgres.db_username", "postgres")
	v.SetDefault("business_databases.game.postgres.db_password", "123456")

	v.SetDefault("log.log_level", "debug")
	v.SetDefault("log.log_mode", "dev")
	v.SetDefault("log.log_path", "./logs/gofurry-admin.log")
	v.SetDefault("log.log_max_size", 100)
	v.SetDefault("log.log_max_backups", 30)
	v.SetDefault("log.log_max_age", 7)

	v.SetDefault("redis.enabled", false)
	v.SetDefault("redis.redis_username", "")
	v.SetDefault("redis.redis_addr", "127.0.0.1:6379")
	v.SetDefault("redis.redis_password", "")
	v.SetDefault("redis.redis_db", 0)
	v.SetDefault("redis.redis_pool_size", 10)

	v.SetDefault("middleware.request_id.enabled", true)
	v.SetDefault("middleware.request_id.header", "X-Request-ID")
	v.SetDefault("middleware.access_log.enabled", true)
	v.SetDefault("middleware.access_log.format", "${time} | ${status} | ${latency} | ${method} | ${path} | rid=${respHeader:X-Request-ID}")
	v.SetDefault("middleware.access_log.time_format", common.TIME_FORMAT_LOG)
	v.SetDefault("middleware.access_log.time_zone", "Local")
	v.SetDefault("middleware.timeout.enabled", true)
	v.SetDefault("middleware.timeout.duration_seconds", 30)
	v.SetDefault("middleware.timeout.exclude_paths", []string{"/livez", "/readyz", "/startupz", "/healthz"})
	v.SetDefault("middleware.health.enabled", true)
	v.SetDefault("middleware.health.include_legacy", true)
	v.SetDefault("middleware.security_headers.enabled", true)
	v.SetDefault("middleware.security_headers.hsts_max_age", 0)
	v.SetDefault("middleware.security_headers.hsts_exclude_subdomains", false)
	v.SetDefault("middleware.security_headers.hsts_preload_enabled", false)
	v.SetDefault("middleware.security_headers.csp_report_only", false)
	v.SetDefault("middleware.compression.enabled", true)
	v.SetDefault("middleware.compression.level", "default")
	v.SetDefault("middleware.limiter.enabled", true)
	v.SetDefault("middleware.limiter.max_requests", 3000)
	v.SetDefault("middleware.limiter.expiration", 60)
	v.SetDefault("middleware.limiter.strategy", "fixed")
	v.SetDefault("middleware.limiter.key_source", "ip")
	v.SetDefault("middleware.limiter.skip_failed_requests", false)
	v.SetDefault("middleware.limiter.skip_successful_requests", false)
	v.SetDefault("middleware.limiter.disable_headers", false)
	v.SetDefault("middleware.limiter.exclude_paths", []string{"/livez", "/readyz", "/startupz", "/healthz"})
	v.SetDefault("middleware.csrf.enabled", true)
	v.SetDefault("middleware.csrf.token_path", "/csrf/token")
	v.SetDefault("middleware.csrf.cookie_name", "csrf_")
	v.SetDefault("middleware.csrf.cookie_same_site", "Lax")
	v.SetDefault("middleware.csrf.cookie_secure", false)
	v.SetDefault("middleware.csrf.cookie_http_only", false)
	v.SetDefault("middleware.csrf.cookie_session_only", false)
	v.SetDefault("middleware.csrf.idle_timeout_seconds", 7200)
	v.SetDefault("middleware.csrf.single_use_token", false)
	v.SetDefault("middleware.csrf.exclude_paths", []string{
		"/api/v1/auth/bootstrap",
		"/api/v1/auth/login",
		"/api/v1/auth/logout",
	})
	v.SetDefault("middleware.etag.enabled", true)
	v.SetDefault("middleware.etag.weak", false)

	v.SetDefault("waf.enabled", false)
	v.SetDefault("waf.conf_path", []string{"./config/coraza.conf"})

	v.SetDefault("auth.jwt_secret", "gofurry-admin-dev-secret")
	v.SetDefault("auth.cookie_name", "gofurry_admin_token")
	v.SetDefault("auth.cookie_domain", "")
	v.SetDefault("auth.cookie_secure", false)
	v.SetDefault("auth.cookie_max_age_seconds", 2592000)
	v.SetDefault("auth.pbkdf2_iterations", 300000)
	v.SetDefault("auth.session_ttl_hours", 720)
	v.SetDefault("auth.cookie_same_site", "Lax")
	v.SetDefault("auth.csrf_header_required", true)

	v.SetDefault("external_services.game_backend.base_url", "http://127.0.0.1:9998")
	v.SetDefault("external_services.game_backend.admin_token", "")
	v.SetDefault("external_services.game_backend.admin_token_header", "X-GoFurry-Admin-Token")
	v.SetDefault("external_services.game_backend.timeout_seconds", 10)
	v.SetDefault("external_services.nav_backend.base_url", "http://127.0.0.1:9999")
	v.SetDefault("external_services.nav_backend.admin_token", "")
	v.SetDefault("external_services.nav_backend.admin_token_header", "X-GoFurry-Admin-Token")
	v.SetDefault("external_services.nav_backend.timeout_seconds", 10)
}

func GetServerConfig() *serverConfig {
	ensureServerConfig()
	if configuration != nil {
		return configuration
	}
	cfg := new(serverConfig)
	cfg.normalize()
	return cfg
}

func currentConfigOptions() configLoaderOptions {
	configOptionsMu.Lock()
	defer configOptionsMu.Unlock()
	return configOptions
}
