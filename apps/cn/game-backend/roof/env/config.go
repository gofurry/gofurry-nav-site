package env

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var configuration = new(serverConfig)

const defaultOnlinePeakCacheDays = 30

func init() {
	if isGoTestBinary() {
		applyTestConfigDefaults(configuration)
	}
}

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
	Thread     ThreadConfig     `yaml:"thread"`
	Middleware MiddlewareConfig `yaml:"middleware"`
	Waf        WafConfig        `yaml:"waf"`
	Email      EmailConfig      `yaml:"email"`
	Admin      AdminConfig      `yaml:"admin"`
	Prize      PrizeConfig      `yaml:"prize"`
	Game       GameConfig       `yaml:"game"`
}

type GameConfig struct {
	OnlinePeakCacheDays int `yaml:"online_peak_cache_days"`
}

type AdminConfig struct {
	Token  string `yaml:"token"`
	Header string `yaml:"header"`
}

type PrizeConfig struct {
	ActivationFrontendURL string `yaml:"activation_frontend_url"`
}

// EmailConfig 邮箱服务配置
type EmailConfig struct {
	EmailHost     string `yaml:"email_host"`
	EmailPort     int    `yaml:"email_port"`
	EmailUser     string `yaml:"email_user"`
	EmailPassword string `yaml:"email_password"`
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
	RedisAddr     string `yaml:"redis_addr"`
	RedisPassword string `yaml:"redis_password"`
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

	for _, file := range configCandidates(projectName, fileName) {
		if tryLoadConfig(file, conf) {
			hit = true
			break
		}
	}

	if !hit && isGoTestBinary() {
		if applyTestConfigDefaults(conf) {
			return
		}
	}

	if hit == false {
		fmt.Println("can not find any " + fileName + " file")
		panic("can not find any " + fileName + " file")
	}
}

func configCandidates(projectName string, fileName string) []string {
	files := []string{"/etc/" + projectName + "/" + fileName}
	files = append(files, localConfigCandidates(fileName)...)
	return files
}

func localConfigCandidates(fileName string) []string {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error loading pwd dir:", err.Error())
		return nil
	}

	var files []string
	for dir := pwd; ; dir = filepath.Dir(dir) {
		files = append(files, filepath.Join(dir, "conf", fileName))
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}

	return files
}

func isGoTestBinary() bool {
	name := filepath.Base(os.Args[0])
	return strings.HasSuffix(name, ".test") || strings.HasSuffix(name, ".test.exe")
}

func applyTestConfigDefaults(conf interface{}) bool {
	cfg, ok := conf.(*serverConfig)
	if !ok {
		return false
	}

	*cfg = serverConfig{
		ClusterId: 1,
		Server: ServerConfig{
			Mode:        "test",
			MemoryLimit: 1,
			GCPercent:   1000,
			Network:     "tcp",
		},
		Thread: ThreadConfig{
			EventPublishThread: 1,
		},
		Game: GameConfig{
			OnlinePeakCacheDays: defaultOnlinePeakCacheDays,
		},
	}

	return true
}

func (cfg *serverConfig) OnlinePeakCacheDays() int {
	if cfg == nil || cfg.Game.OnlinePeakCacheDays <= 0 {
		return defaultOnlinePeakCacheDays
	}
	return cfg.Game.OnlinePeakCacheDays
}

func tryLoadConfig(file string, conf interface{}) bool {
	if !FileExists(file) {
		return false
	}

	err := loadYaml(file, conf)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	return true
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

func GetServerConfig() *serverConfig {
	return configuration
}
