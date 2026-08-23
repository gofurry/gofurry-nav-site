package env

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var configuration = new(serverConfig)

func LoadServerConfig(configFile string) error {
	cfg, err := readServerConfig(configFile)
	if err != nil {
		return err
	}
	configuration = cfg
	return nil
}

func ValidateServerConfigFile(configFile string) error {
	_, err := readServerConfig(configFile)
	return err
}

func readServerConfig(configFile string) (*serverConfig, error) {
	configFile = strings.TrimSpace(configFile)
	if configFile == "" {
		return nil, errors.New("--config is required")
	}
	v := viper.New()
	v.SetConfigFile(configFile)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config %q: %w", configFile, err)
	}
	data, err := yaml.Marshal(v.AllSettings())
	if err != nil {
		return nil, fmt.Errorf("normalize config %q: %w", configFile, err)
	}
	cfg := new(serverConfig)
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(cfg); err != nil {
		return nil, fmt.Errorf("decode config %q: %w", configFile, err)
	}
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validate config %q: %w", configFile, err)
	}
	return cfg, nil
}

func (cfg *serverConfig) validate() error {
	if cfg.ClusterId < 0 || cfg.ClusterId > 1023 {
		return errors.New("cluster_id must be between 0 and 1023")
	}
	if cfg.Server.Port == "" || cfg.Server.Network == "" {
		return errors.New("server.port and server.network are required")
	}
	if cfg.Server.MemoryLimit <= 0 {
		return errors.New("server.memory_limit must be positive")
	}
	if cfg.DataBase.DBName == "" || cfg.DataBase.DBUsername == "" || cfg.DataBase.DBHost == "" || cfg.DataBase.DBPort == "" {
		return errors.New("database name, username, host, and port are required")
	}
	if cfg.Redis.RedisAddr == "" {
		return errors.New("redis.redis_addr is required")
	}
	return nil
}

type serverConfig struct {
	ClusterId  int              `yaml:"cluster_id"`
	Server     ServerConfig     `yaml:"server"`
	DataBase   DataBaseConfig   `yaml:"database"`
	Log        LogConfig        `yaml:"log"`
	Redis      RedisConfig      `yaml:"redis"`
	NavV2      NavV2Config      `yaml:"nav_v2"`
	Security   SecurityConfig   `yaml:"security"`
	Thread     ThreadConfig     `yaml:"thread"`
	Middleware MiddlewareConfig `yaml:"middleware"`
	Waf        WafConfig        `yaml:"waf"`
	Proxy      ProxyConfig      `yaml:"proxy"`
	Resource   ResourceConfig   `yaml:"resource"`
	Admin      AdminConfig      `yaml:"admin"`
}

type AdminConfig struct {
	Token  string `yaml:"token"`
	Header string `yaml:"header"`
}

type ResourceConfig struct {
	NavImageNum        int `yaml:"nav_image_num"`
	NavResizedImageNum int `yaml:"nav_resized_image_num"`
}

type ProxyConfig struct {
	Url string `yaml:"url"`
}

type WafConfig struct {
	ConfPath        string   `yaml:"conf_path"`
	DirectivesFiles []string `yaml:"directives_files"`
	WafSwitch       bool     `yaml:"waf_switch"`
}

func (cfg WafConfig) ResolveDirectivesFiles() []string {
	if len(cfg.DirectivesFiles) > 0 {
		return append([]string(nil), cfg.DirectivesFiles...)
	}

	baseFile := cfg.ConfPath
	if baseFile == "" {
		baseFile = "./conf/coraza.conf"
	}

	return []string{baseFile}
}

type MiddlewareConfig struct {
	Cors    CorsConfig    `yaml:"cors"`
	Limiter LimiterConfig `yaml:"limiter"`
}

// LimiterConfig 限流器配置
type LimiterConfig struct {
	IsOn        bool `yaml:"is_on"`
	MaxRequests int  `yaml:"max_requests"`
	Expiration  int  `yaml:"expiration"`
}

type CorsConfig struct {
	AllowOrigins string `yaml:"allow_origins"`
}

type ThreadConfig struct {
	EventPublishThread int `yaml:"event_publish_thread"`
}

type RedisConfig struct {
	RedisAddr      string `yaml:"redis_addr"`
	RedisPassword  string `yaml:"redis_password"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
}

func (cfg RedisConfig) Timeout() time.Duration {
	if cfg.TimeoutSeconds <= 0 {
		return 2 * time.Second
	}
	return time.Duration(cfg.TimeoutSeconds) * time.Second
}

type NavV2Config struct {
	Enabled                  *bool `yaml:"enabled"`
	SummaryEnabled           bool  `yaml:"summary_enabled"`
	DetailEnabled            *bool `yaml:"detail_enabled"`
	ReadModelEnabled         *bool `yaml:"read_model_enabled"`
	SummaryStaleAfterSeconds int   `yaml:"summary_stale_after_seconds"`
	RawPayloadPreviewBytes   int   `yaml:"raw_payload_preview_bytes"`
	FullPayloadEnabled       bool  `yaml:"full_payload_enabled"`
	PayloadResponseMaxBytes  int   `yaml:"payload_response_max_bytes"`
}

func (cfg NavV2Config) SummaryStaleAfter() time.Duration {
	if cfg.SummaryStaleAfterSeconds <= 0 {
		return 24 * time.Hour
	}
	return time.Duration(cfg.SummaryStaleAfterSeconds) * time.Second
}

func (cfg NavV2Config) AnyRouteEnabled() bool {
	return cfg.SummaryRoutesEnabled() || cfg.DetailRoutesEnabled() || cfg.ReadModelRoutesEnabled()
}

func (cfg NavV2Config) SummaryRoutesEnabled() bool {
	return cfg.v2RoutesEnabled() && cfg.SummaryEnabled
}

func (cfg NavV2Config) DetailRoutesEnabled() bool {
	return cfg.v2RoutesEnabled() && boolConfigValue(cfg.DetailEnabled, cfg.SummaryEnabled)
}

func (cfg NavV2Config) ReadModelRoutesEnabled() bool {
	return cfg.v2RoutesEnabled() && boolConfigValue(cfg.ReadModelEnabled, cfg.DetailRoutesEnabled())
}

func (cfg NavV2Config) RawPayloadPreviewBytesOrDefault() int {
	if cfg.RawPayloadPreviewBytes <= 0 {
		return 64 * 1024
	}
	return cfg.RawPayloadPreviewBytes
}

func (cfg NavV2Config) PayloadResponseMaxBytesOrDefault() int {
	if cfg.PayloadResponseMaxBytes <= 0 {
		return 2 * 1024 * 1024
	}
	return cfg.PayloadResponseMaxBytes
}

func (cfg NavV2Config) v2RoutesEnabled() bool {
	return boolConfigValue(cfg.Enabled, cfg.SummaryEnabled)
}

func boolConfigValue(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

type LogConfig struct {
	LogLevel      string `yaml:"log_level"`
	LogMode       string `yaml:"log_mode"`
	LogPath       string `yaml:"log_path"`
	LogMaxSize    int    `yaml:"log_max_size"`
	LogMaxBackups int    `yaml:"log_max_backups"`
	LogMaxAge     int    `yaml:"log_max_age"`
}

type DataBaseConfig struct {
	DBName                       string `yaml:"db_name"`
	DBUsername                   string `yaml:"db_username"`
	DBPassword                   string `yaml:"db_password"`
	DBHost                       string `yaml:"db_host"`
	DBPort                       string `yaml:"db_port"`
	MaxConns                     int32  `yaml:"max_conns"`
	MinConns                     int32  `yaml:"min_conns"`
	MaxConnLifetimeSeconds       int    `yaml:"max_conn_lifetime_seconds"`
	MaxConnLifetimeJitterSeconds int    `yaml:"max_conn_lifetime_jitter_seconds"`
	MaxConnIdleTimeSeconds       int    `yaml:"max_conn_idle_time_seconds"`
	HealthCheckPeriodSeconds     int    `yaml:"health_check_period_seconds"`
	ConnectTimeoutSeconds        int    `yaml:"connect_timeout_seconds"`
	PingTimeoutSeconds           int    `yaml:"ping_timeout_seconds"`
}

func (cfg DataBaseConfig) ConnectionString() string {
	u := &url.URL{Scheme: "postgres", Host: cfg.DBHost + ":" + cfg.DBPort, Path: cfg.DBName}
	u.User = url.UserPassword(cfg.DBUsername, cfg.DBPassword)
	query := u.Query()
	query.Set("sslmode", "prefer")
	u.RawQuery = query.Encode()
	return u.String()
}

type ServerConfig struct {
	Mode              string `yaml:"mode"`
	IPAddress         string `yaml:"ip_address"`
	Port              string `yaml:"port"`
	MemoryLimit       int    `yaml:"memory_limit"`
	GCPercent         int    `yaml:"gc_percent"`
	Network           string `yaml:"network"`
	EnablePrefork     bool   `yaml:"enable_prefork"`
	TrustedProxyCIDRs string `yaml:"trusted_proxy_cidrs"`
}

type SecurityConfig struct {
	JWTSecret string `yaml:"jwt_secret"`
}

func InitServerConfig(projectName string) {
	InitConfig(projectName, "server.yaml", configuration)
}

func InitConfig(projectName string, fileName string, conf interface{}) {
	hit := false

	file := "/etc/" + projectName + "/" + fileName
	if FileExists(file) {
		err := loadYaml(file, conf)
		if err != nil {
			traceConfig(err.Error())
		} else {
			hit = true
		}
	}

	//默认启动本地路径下conf.env
	if !hit {
		pwd, err := os.Getwd()
		if err != nil {
			traceConfig("Error loading pwd dir: " + err.Error())
		} else {
			filePath := pwd + "/conf/" + fileName
			if FileExists(filePath) {
				err = loadYaml(filePath, conf)
				if err != nil {
					traceConfig("Error loading " + fileName + " file: " + err.Error())
				} else {
					hit = true
				}
			}
		}
	}

	if hit == false {
		fmt.Println("can not find any " + fileName + " file")
		panic("can not find any " + fileName + " file")
	}
}

func FileExists(path string) bool {
	traceConfig("check filepath:" + path)
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func loadYaml(path string, conf interface{}) (err error) {
	traceConfig("load config:" + path)
	if FileExists(path) {
		fileBytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return yaml.Unmarshal(fileBytes, conf)
	}
	return errors.New("未找到配置文件" + path)
}

func traceConfig(message string) {
	if os.Getenv("GF_NAV_BACKEND_CONFIG_TRACE") == "1" {
		fmt.Println(message)
	}
}

func isRunningGoTest() bool {
	name := filepath.Base(os.Args[0])
	return strings.HasSuffix(name, ".test") || strings.HasSuffix(name, ".test.exe")
}

func GetServerConfig() *serverConfig {
	return configuration
}
