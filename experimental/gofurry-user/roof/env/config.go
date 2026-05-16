package env

import (
	"errors"
	"fmt"
	"os"

	"github.com/gofurry/gofurry-user/common"
	"gopkg.in/yaml.v2"
)

func init() {
	InitServerConfig(common.COMMON_PROJECT_NAME)
}

var configuration = new(serverConfig)

type serverConfig struct {
	ClusterId  int              `yaml:"cluster_id"`
	Server     ServerConfig     `yaml:"server"`
	Key        KeyConfig        `yaml:"key"`
	DataBase   DataBaseConfig   `yaml:"database"`
	Log        LogConfig        `yaml:"log"`
	Redis      RedisConfig      `yaml:"redis"`
	Email      EmailConfig      `yaml:"email"`
	Thread     ThreadConfig     `yaml:"thread"`
	Github     GithubConfig     `yaml:"github"`
	Gitee      GiteeConfig      `yaml:"gitee"`
	Middleware MiddlewareConfig `yaml:"middleware"`
	Waf        WafConfig        `yaml:"waf"`
	Proxy      ProxyConfig      `yaml:"proxy"`
	Resource   ResourceConfig   `yaml:"resource"`
	Etcd       EtcdConfig       `yaml:"etcd"`
	Auth       AuthConfig       `yaml:"auth"`
}

type AuthConfig struct {
	AuthSalt  string `yaml:"auth_salt"`
	JwtSecret string `yaml:"jwt_secret"`
}

type EtcdConfig struct {
	EtcdHost string `yaml:"etcd_host"`
	EtcdPort string `yaml:"etcd_port"`
	EtcdKey  string `yaml:"etcd_key"`
}

type ResourceConfig struct {
	ImagePath       string `yaml:"bg_image_path"`
	ResizeImagePath string `yaml:"bg_resize_image_path"`
	ImageExts       string `yaml:"image_exts"`
	Geolite2Path    string `yaml:"geolite2_path"`
}

type ProxyConfig struct {
	Url string `yaml:"url"`
}

type WafConfig struct {
	ConfPath  string `yaml:"conf_path"`
	WafSwitch string `yaml:"waf_switch"`
}

type MiddlewareConfig struct {
	Swagger SwaggerConfig `yaml:"swagger"`
	Cors    CorsConfig    `yaml:"cors"`
}

type CorsConfig struct {
	AllowOrigins string `yaml:"allow_origins"`
}

type SwaggerConfig struct {
	IsOn     string `yaml:"is_on"`
	FilePath string `yaml:"file_path"`
	BasePath string `yaml:"base_path"`
	Path     string `yaml:"path"`
	Title    string `yaml:"title"`
}

type GithubConfig struct {
	ClientId     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectUrl  string `yaml:"redirect_url"`
}

type GiteeConfig struct {
	ClientId     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectUrl  string `yaml:"redirect_url"`
}

type ThreadConfig struct {
	WsStatusReceiveThread int `yaml:"ws_status_receive_thread"`
	WsSubscribeSendCache  int `yaml:"ws_subscribe_send_cache"`
	SteamAppInfoThread    int `yaml:"steam_app_info_thread"`
	EventPublishThread    int `yaml:"event_publish_thread"`
}

type EmailConfig struct {
	EmailHost     string `yaml:"email_host"`
	EmailPort     int    `yaml:"email_port"`
	EmailUser     string `yaml:"email_user"`
	EmailPassword string `yaml:"email_password"`
}

type RedisConfig struct {
	RedisAddr     string `yaml:"redis_addr"`
	RedisPassword string `yaml:"redis_password"`
}

type LogConfig struct {
	LogRotationCount int    `yaml:"log_rotation_count"`
	LogPath          string `yaml:"log_path"`
	LogLevel         string `yaml:"log_level"`
	LogChokeLength   int    `yaml:"log_choke_length"`
}

type DataBaseConfig struct {
	DBName     string `yaml:"db_name"`
	DBUsername string `yaml:"db_username"`
	DBPassword string `yaml:"db_password"`
	DBHost     string `yaml:"db_host"`
	DBPort     string `yaml:"db_port"`
}

type ServerConfig struct {
	AppName     string `yaml:"app_name"`
	AppVersion  string `yaml:"app_version"`
	Mode        string `yaml:"models"`
	IPAddress   string `yaml:"ip_address"`
	Port        string `yaml:"port"`
	MemoryLimit int    `yaml:"memory_limit"`
}

type KeyConfig struct {
	LoginPrivate string `yaml:"login_private"`
	LoginPublic  string `yaml:"login_public"`
	TlsKey       string `yaml:"tls_key"`
	TlsPem       string `yaml:"tls_pem"`
	GrpcTls      string `yaml:"grpc_tls"`
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
