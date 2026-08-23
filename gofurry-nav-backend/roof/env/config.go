package env

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofurry/gofurry-nav-backend/common"
	"gopkg.in/yaml.v3"
)

func init() {
	if isRunningGoTest() && os.Getenv("GF_NAV_BACKEND_LOAD_CONFIG_IN_TEST") != "1" {
		return
	}
	InitServerConfig(common.COMMON_PROJECT_NAME)
}

var configuration = new(serverConfig)

type serverConfig struct {
	ClusterId  int              `yaml:"cluster_id"`
	Server     ServerConfig     `yaml:"server"`
	DataBase   DataBaseConfig   `yaml:"database"`
	Log        LogConfig        `yaml:"log"`
	Redis      RedisConfig      `yaml:"redis"`
	Uptime     UptimeConfig     `yaml:"uptime"`
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
	CRSRoot         string   `yaml:"crs_root"`
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

	files := []string{baseFile}
	if cfg.CRSRoot != "" {
		files = append(files,
			filepath.Join(cfg.CRSRoot, "crs-setup.conf.example"),
			filepath.Join(cfg.CRSRoot, "rules", "*.conf"),
		)
	}

	return files
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

type UptimeConfig struct {
	Enabled               bool                   `yaml:"enabled"`
	ServiceID             string                 `yaml:"service_id"`
	ServiceName           string                 `yaml:"service_name"`
	ServiceDescription    string                 `yaml:"service_description"`
	SampleIntervalSeconds int                    `yaml:"sample_interval_seconds"`
	RetentionDays         int                    `yaml:"retention_days"`
	DaysToShow            int                    `yaml:"days_to_show"`
	Timezone              string                 `yaml:"timezone"`
	StorageKeyPrefix      string                 `yaml:"storage_key_prefix"`
	UI                    UptimeUIConfig         `yaml:"ui"`
	Endpoints             []UptimeEndpointConfig `yaml:"endpoints"`
}

type UptimeUIConfig struct {
	Path            string  `yaml:"path"`
	Title           string  `yaml:"title"`
	Description     string  `yaml:"description"`
	Footer          string  `yaml:"footer"`
	GreenThreshold  float64 `yaml:"green_threshold"`
	YellowThreshold float64 `yaml:"yellow_threshold"`
}

type UptimeEndpointConfig struct {
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
