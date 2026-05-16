package env

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofurry/gofurry-game-backend/common"
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
	Mongodb    MongodbConfig    `yaml:"mongodb"`
	Thread     ThreadConfig     `yaml:"thread"`
	Middleware MiddlewareConfig `yaml:"middleware"`
	Waf        WafConfig        `yaml:"waf"`
	Proxy      ProxyConfig      `yaml:"proxy"`
	Resource   ResourceConfig   `yaml:"resource"`
	Auth       AuthConfig       `yaml:"auth"`
	Prometheus PrometheusConfig `yaml:"prometheus"`
	Email      EmailConfig      `yaml:"email"`
}

// EmailConfig 邮箱服务配置
type EmailConfig struct {
	EmailHost     string `yaml:"email_host"`
	EmailPort     int    `yaml:"email_port"`
	EmailUser     string `yaml:"email_user"`
	EmailPassword string `yaml:"email_password"`
}

type PrometheusConfig struct {
	Url      string `yaml:"url"`
	AuthName string `yaml:"auth_name"`
	AuthPass string `yaml:"auth_pass"`
}

type MongodbConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	AuthDB   string `yaml:"auth_db"`
	DBName   string `yaml:"db_name"`
}

type AuthConfig struct {
	AuthSalt  string `yaml:"auth_salt"`
	JwtSecret string `yaml:"jwt_secret"`
}

type ResourceConfig struct {
	Geolite2Path string `yaml:"geolite2_path"`
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
	WsStatusReceiveThread int `yaml:"ws_status_receive_thread"`
	WsSubscribeSendCache  int `yaml:"ws_subscribe_send_cache"`
	SteamAppInfoThread    int `yaml:"steam_app_info_thread"`
	EventPublishThread    int `yaml:"event_publish_thread"`
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
	DBName     string `yaml:"db_name"`
	DBUsername string `yaml:"db_username"`
	DBPassword string `yaml:"db_password"`
	DBHost     string `yaml:"db_host"`
	DBPort     string `yaml:"db_port"`
}

type ServerConfig struct {
	AppName       string `yaml:"app_name"`
	AppVersion    string `yaml:"app_version"`
	Mode          string `yaml:"mode"`
	IPAddress     string `yaml:"ip_address"`
	Port          string `yaml:"port"`
	MemoryLimit   int    `yaml:"memory_limit"`
	GCPercent     int    `yaml:"gc_percent"`
	Network       string `yaml:"network"`
	EnablePrefork bool   `yaml:"enable_prefork"`
}

type KeyConfig struct {
	LoginPrivate string `yaml:"login_private"`
	LoginPublic  string `yaml:"login_public"`
	TlsKey       string `yaml:"tls_key"`
	TlsPem       string `yaml:"tls_pem"`
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
