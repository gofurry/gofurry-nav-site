package env

import (
	"errors"
	"fmt"
	"os"

	"github.com/gofurry/gofurry-oauth-login/common"
	"gopkg.in/yaml.v2"
)

func init() {
	InitServerConfig(common.COMMON_PROJECT_NAME)
}

var configuration = new(serverConfig)

type serverConfig struct {
	ClusterId int          `yaml:"cluster_id"`
	Server    ServerConfig `yaml:"server"`
	Github    GithubConfig `yaml:"github"`
	Proxy     ProxyConfig  `yaml:"proxy"`
	Etcd      EtcdConfig   `yaml:"etcd"`
	Log       LogConfig    `yaml:"log"`
	Key       KeyConfig    `yaml:"key"`
}

type KeyConfig struct {
	TlsCrt string `yaml:"tls_crt"`
	TlsKey string `yaml:"tls_key"`
}

type LogConfig struct {
	LogRotationCount int    `yaml:"log_rotation_count"`
	LogPath          string `yaml:"log_path"`
	LogLevel         string `yaml:"log_level"`
	LogChokeLength   int    `yaml:"log_choke_length"`
}

type EtcdConfig struct {
	EtcdHost string `yaml:"etcd_host"`
	EtcdPort string `yaml:"etcd_port"`
	EtcdKey  string `yaml:"etcd_key"`
}

type ProxyConfig struct {
	Url string `yaml:"url"`
}

type ServerConfig struct {
	AppName     string `yaml:"app_name"`
	MemoryLimit int    `yaml:"memory_limit"`
	IPAddress   string `yaml:"ip_address"`
	Port        string `yaml:"port"`
}

type GithubConfig struct {
	ClientId     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectUrl  string `yaml:"redirect_url"`
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
