package env

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofurry/gofurry-nav-backend/common"
	"gopkg.in/yaml.v2"
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
	NavV2      NavV2Config      `yaml:"nav_v2"`
	Thread     ThreadConfig     `yaml:"thread"`
	Middleware MiddlewareConfig `yaml:"middleware"`
	Waf        WafConfig        `yaml:"waf"`
	Proxy      ProxyConfig      `yaml:"proxy"`
	Resource   ResourceConfig   `yaml:"resource"`
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
	Swagger SwaggerConfig `yaml:"swagger"`
	Cors    CorsConfig    `yaml:"cors"`
	Limiter LimiterConfig `yaml:"limiter"`
}

// LimiterConfig 限流器配置
type LimiterConfig struct {
	IsOn        bool          `yaml:"is_on"`
	MaxRequests int           `yaml:"max_requests"`
	Expiration  time.Duration `yaml:"expiration"`
}

type CorsConfig struct {
	AllowOrigins string `yaml:"allow_origins"`
}

type SwaggerConfig struct {
	IsOn     bool   `yaml:"is_on"`
	FilePath string `yaml:"file_path"`
	BasePath string `yaml:"base_path"`
	Path     string `yaml:"path"`
	Title    string `yaml:"title"`
}

type ThreadConfig struct {
	EventPublishThread int `yaml:"event_publish_thread"`
}

type RedisConfig struct {
	RedisAddr     string `yaml:"redis_addr"`
	RedisPassword string `yaml:"redis_password"`
}

type NavV2Config struct {
	Enabled                  *bool `yaml:"enabled"`
	SummaryEnabled           bool  `yaml:"summary_enabled"`
	DetailEnabled            *bool `yaml:"detail_enabled"`
	ReadModelEnabled         *bool `yaml:"read_model_enabled"`
	SummaryStaleAfterSeconds int   `yaml:"summary_stale_after_seconds"`
	RawPayloadPreviewBytes   int   `yaml:"raw_payload_preview_bytes"`
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
	DBName                 string `yaml:"db_name"`
	DBUsername             string `yaml:"db_username"`
	DBPassword             string `yaml:"db_password"`
	DBHost                 string `yaml:"db_host"`
	DBPort                 string `yaml:"db_port"`
	MaxOpenConns           int    `yaml:"max_open_conns"`
	MaxIdleConns           int    `yaml:"max_idle_conns"`
	ConnMaxLifetimeSeconds int    `yaml:"conn_max_lifetime_seconds"`
	ConnMaxIdleTimeSeconds int    `yaml:"conn_max_idle_time_seconds"`
}

type ServerConfig struct {
	Mode          string `yaml:"mode"`
	IPAddress     string `yaml:"ip_address"`
	Port          string `yaml:"port"`
	MemoryLimit   int    `yaml:"memory_limit"`
	GCPercent     int    `yaml:"gc_percent"`
	Network       string `yaml:"network"`
	EnablePrefork bool   `yaml:"enable_prefork"`
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
			fmt.Println(err.Error())
		} else {
			hit = true
		}
	}

	//默认启动本地路径下conf.env
	if !hit {
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Println("Error loading pwd dir:", err.Error())
		} else {
			filePath := pwd + "/conf/" + fileName
			if FileExists(filePath) {
				err = loadYaml(filePath, conf)
				if err != nil {
					fmt.Println("Error loading "+fileName+" file:", err.Error())
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
	fmt.Println("check filepath:" + path)
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func loadYaml(path string, conf interface{}) (err error) {
	fmt.Println("load config:" + path)
	if FileExists(path) {
		fileBytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return yaml.Unmarshal(fileBytes, conf)
	}
	return errors.New("未找到配置文件" + path)
}

func isRunningGoTest() bool {
	name := filepath.Base(os.Args[0])
	return strings.HasSuffix(name, ".test") || strings.HasSuffix(name, ".test.exe")
}

func GetServerConfig() *serverConfig {
	return configuration
}
