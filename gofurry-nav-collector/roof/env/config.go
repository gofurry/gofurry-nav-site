package env

import (
	"errors"
	"fmt"
	"os"

	"github.com/gofurry/gofurry-nav-collector/common"
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
	Collector CollectorConfig `yaml:"collector"`
}

type CollectorConfig struct {
	Proxy   string        `yaml:"proxy"`
	Ping    PingConfig    `yaml:"ping"`
	Request RequestConfig `yaml:"request"`
	Dns     DnsConfig     `yaml:"dns"`
}

type DnsConfig struct {
	DnsThread    int    `yaml:"dns_thread"`
	QueryThread  int    `yaml:"query_thread"`
	DnsInterval  int    `yaml:"dns_interval"`
	Resolver     string `yaml:"resolver"`
	Geolite2Path string `yaml:"geolite2_path"`
	LogCount     string `yaml:"log_count"`
}

type RequestConfig struct {
	RequestThread   int    `yaml:"request_thread"`
	RequestInterval int    `yaml:"request_interval"`
	LogCount        string `yaml:"log_count"`
}

type PingConfig struct {
	PingThread   int    `yaml:"ping_thread"`
	PingInterval int    `yaml:"ping_interval"`
	PingKey      string `yaml:"ping_key"`
	ResultKey    string `yaml:"result_key"`
	LogCount     string `yaml:"log_count"`
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
