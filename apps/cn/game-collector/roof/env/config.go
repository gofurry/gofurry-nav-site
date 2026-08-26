package env

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofurry/gofurry-game-collector/internal/health"
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
	if cfg.Server.MemoryLimit <= 0 {
		return errors.New("server.memory_limit must be positive")
	}
	if cfg.DataBase.DBName == "" || cfg.DataBase.DBUsername == "" || cfg.DataBase.DBHost == "" || cfg.DataBase.DBPort == "" {
		return errors.New("data_base name, username, host, and port are required")
	}
	if cfg.Redis.RedisAddr == "" {
		return errors.New("redis.redis_addr is required")
	}
	if err := health.ValidateConfig(cfg.Health); err != nil {
		return err
	}
	return nil
}

type serverConfig struct {
	ClusterId int             `yaml:"cluster_id"`
	Server    ServerConfig    `yaml:"server"`
	DataBase  DataBaseConfig  `yaml:"data_base"`
	Redis     RedisConfig     `yaml:"redis"`
	Log       LogConfig       `yaml:"log"`
	Health    health.Config   `yaml:"health"`
	Collector CollectorConfig `yaml:"collector"`
}

type LogConfig struct {
	LogLevel         string `yaml:"log_level"`
	LogMode          string `yaml:"log_mode"`
	LogPath          string `yaml:"log_path"`
	LogMaxSize       int    `yaml:"log_max_size"`
	LogMaxBackups    int    `yaml:"log_max_backups"`
	LogMaxAge        int    `yaml:"log_max_age"`
	LogCompress      bool   `yaml:"log_compress"`
	LogRotationCount int    `yaml:"log_rotation_count"`
	LogChokeLength   int    `yaml:"log_choke_length"`
}

type CollectorConfig struct {
	Proxy string            `yaml:"proxy"`
	Game  GameConfig        `yaml:"game"`
	V2    CollectorV2Config `yaml:"v2"`
}

type GameConfig struct {
	GamePlayerInterval      int   `yaml:"game_player_interval"`
	CollectPlayersOnStartup *bool `yaml:"collect_players_on_startup"`
}

// PlayersOnStartupEnabled preserves the historical startup collection unless
// an operator explicitly disables it in typed configuration.
func (cfg GameConfig) PlayersOnStartupEnabled() bool {
	return cfg.CollectPlayersOnStartup == nil || *cfg.CollectPlayersOnStartup
}

type CollectorV2Config struct {
	Steam     CollectorV2SteamConfig     `yaml:"steam"`
	Retention CollectorV2RetentionConfig `yaml:"retention"`
}

type CollectorV2SteamConfig struct {
	APIRequestsPer5Minutes   int                    `yaml:"api_requests_per_5_minutes"`
	StoreRequestsPer5Minutes int                    `yaml:"store_requests_per_5_minutes"`
	Burst                    int                    `yaml:"burst"`
	MaxWorkers               int                    `yaml:"max_workers"`
	RequestTimeoutSeconds    int                    `yaml:"request_timeout_seconds"`
	Retry                    CollectorV2RetryConfig `yaml:"retry"`
}

type CollectorV2RetryConfig struct {
	MaxAttempts          int `yaml:"max_attempts"`
	BaseDelaySeconds     int `yaml:"base_delay_seconds"`
	CooldownOn429Seconds int `yaml:"cooldown_on_429_seconds"`
}

type CollectorV2RetentionConfig struct {
	PlayerCountsDays       int `yaml:"player_counts_days"`
	CollectRunsDays        int `yaml:"collect_runs_days"`
	CollectTaskResultsDays int `yaml:"collect_task_results_days"`
}

type ServerConfig struct {
	AppName     string `yaml:"app_name"`
	AppVersion  string `yaml:"app_version"`
	Mode        string `yaml:"models"`
	MemoryLimit int    `yaml:"memory_limit"`
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

type RedisConfig struct {
	RedisAddr     string `yaml:"redis_addr"`
	RedisPassword string `yaml:"redis_password"`
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

func getOrDefault(key string, def string) string {
	value := os.Getenv(key)
	if value == "" {
		return def
	}
	return value
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

func isRunningGoTest() bool {
	name := filepath.Base(os.Args[0])
	return strings.HasSuffix(name, ".test") || strings.HasSuffix(name, ".test.exe")
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
