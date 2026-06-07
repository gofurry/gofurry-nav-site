package env

import (
	"errors"
	"fmt"
	"os"

	"github.com/gofurry/gofurry-game-collector/common"
	"gopkg.in/yaml.v2"
)

func init() {
	InitServerConfig(common.COMMON_PROJECT_NAME)
}

var configuration = new(serverConfig)

type serverConfig struct {
	ClusterId int             `yaml:"cluster_id"`
	Server    ServerConfig    `yaml:"server"`
	DataBase  DataBaseConfig  `yaml:"data_base"`
	Redis     RedisConfig     `yaml:"redis"`
	Mongodb   MongodbConfig   `yaml:"mongodb"`
	Collector CollectorConfig `yaml:"collector"`
}

type MongodbConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	AuthDB   string `yaml:"auth_db"`
	DBName   string `yaml:"db_name"`
}

type CollectorConfig struct {
	Proxy   string            `yaml:"proxy"`
	Limiter LimiterConfig     `yaml:"limiter"`
	Game    GameConfig        `yaml:"game"`
	V2      CollectorV2Config `yaml:"v2"`
}

type LimiterConfig struct {
	SteamApi   int `yaml:"steam_api"`
	SteamStore int `yaml:"steam_store"`
}

type GameConfig struct {
	GameThread         int `yaml:"game_thread"`
	GamePlayerInterval int `yaml:"game_player_interval"`
}

type CollectorV2Config struct {
	Enabled bool                   `yaml:"enabled"`
	DryRun  bool                   `yaml:"dry_run"`
	Steam   CollectorV2SteamConfig `yaml:"steam"`
	Tasks   CollectorV2TaskConfig  `yaml:"tasks"`
}

type CollectorV2SteamConfig struct {
	APIIntervalSeconds    int                    `yaml:"api_interval_seconds"`
	StoreIntervalSeconds  int                    `yaml:"store_interval_seconds"`
	Burst                 int                    `yaml:"burst"`
	MaxWorkers            int                    `yaml:"max_workers"`
	RequestTimeoutSeconds int                    `yaml:"request_timeout_seconds"`
	Retry                 CollectorV2RetryConfig `yaml:"retry"`
}

type CollectorV2RetryConfig struct {
	MaxAttempts          int `yaml:"max_attempts"`
	BaseDelaySeconds     int `yaml:"base_delay_seconds"`
	CooldownOn429Seconds int `yaml:"cooldown_on_429_seconds"`
}

type CollectorV2TaskConfig struct {
	NewsEnabled    bool `yaml:"news_enabled"`
	PlayersEnabled bool `yaml:"players_enabled"`
	DetailsEnabled bool `yaml:"details_enabled"`
}

type ServerConfig struct {
	AppName     string `yaml:"app_name"`
	AppVersion  string `yaml:"app_version"`
	Mode        string `yaml:"models"`
	MemoryLimit int    `yaml:"memory_limit"`
}

type DataBaseConfig struct {
	DBName     string `yaml:"db_name"`
	DBUsername string `yaml:"db_username"`
	DBPassword string `yaml:"db_password"`
	DBHost     string `yaml:"db_host"`
	DBPort     string `yaml:"db_port"`
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
