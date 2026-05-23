package env

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofurry/gofurry-nav-collector/common"
	"gopkg.in/yaml.v2"
)

func init() {
	if isRunningGoTest() && os.Getenv("GF_NAV_COLLECTOR_LOAD_CONFIG_IN_TEST") != "1" {
		return
	}
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
	Proxy       string            `yaml:"proxy"`
	Ping        PingConfig        `yaml:"ping"`
	Request     RequestConfig     `yaml:"request"`
	Dns         DnsConfig         `yaml:"dns"`
	ProbeBudget ProbeBudgetConfig `yaml:"probe_budget"`
	V2          CollectorV2Config `yaml:"v2"`
}

type ProbeBudgetConfig struct {
	RedisTimeoutSeconds        int `yaml:"redis_timeout_seconds"`
	HTTPTimeoutSeconds         int `yaml:"http_timeout_seconds"`
	HTTPRedirects              int `yaml:"http_redirects"`
	TLSHandshakeTimeoutSeconds int `yaml:"tls_handshake_timeout_seconds"`
	DNSTimeoutSeconds          int `yaml:"dns_timeout_seconds"`
	PTRTimeoutSeconds          int `yaml:"ptr_timeout_seconds"`
	MaxDNSRecordsPerQuery      int `yaml:"max_dns_records_per_query"`
	MaxResponseBytes           int `yaml:"max_response_bytes"`
}

type CollectorV2Config struct {
	Enabled       bool               `yaml:"enabled"`
	ObservationDB bool               `yaml:"observation_db"`
	LatestRedis   bool               `yaml:"latest_redis"`
	CompareLog    bool               `yaml:"compare_log"`
	Protocols     CollectorProtocols `yaml:"protocols"`
}

type CollectorProtocols struct {
	Ping bool `yaml:"ping"`
	HTTP bool `yaml:"http"`
	DNS  bool `yaml:"dns"`
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
	Mode        string `yaml:"mode"`
	MemoryLimit int    `yaml:"memory_limit"`
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
			filePath := findLocalConfig(pwd, fileName)
			if filePath != "" && FileExists(filePath) {
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

func findLocalConfig(startDir string, fileName string) string {
	dir := startDir
	for {
		filePath := filepath.Join(dir, "conf", fileName)
		if FileExists(filePath) {
			return filePath
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func isRunningGoTest() bool {
	name := filepath.Base(os.Args[0])
	return strings.HasSuffix(name, ".test") || strings.HasSuffix(name, ".test.exe")
}

func RunningInGoTest() bool {
	return isRunningGoTest()
}

func getOrDefault(key string, def string) string {
	value := os.Getenv(key)
	if value == "" {
		return def
	}
	return value
}

func FileExists(path string) bool {
	traceConfigPath("check filepath:" + path)
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func loadYaml(path string, conf interface{}) (err error) {
	traceConfigPath("load config:" + path)
	if FileExists(path) {
		fileBytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return yaml.Unmarshal(fileBytes, conf)
	}
	return errors.New("未找到配置文件" + path)
}

func traceConfigPath(message string) {
	if os.Getenv("GF_NAV_COLLECTOR_CONFIG_TRACE") == "1" {
		fmt.Println(message)
	}
}

func GetServerConfig() *serverConfig {
	return configuration
}

func (cfg ProbeBudgetConfig) RedisTimeout() time.Duration {
	return secondsOrDefault(cfg.RedisTimeoutSeconds, 2)
}

func (cfg ProbeBudgetConfig) HTTPTimeout() time.Duration {
	return secondsOrDefault(cfg.HTTPTimeoutSeconds, 25)
}

func (cfg ProbeBudgetConfig) MaxHTTPRedirects() int {
	return intOrDefault(cfg.HTTPRedirects, 5)
}

func (cfg ProbeBudgetConfig) TLSHandshakeTimeout() time.Duration {
	return secondsOrDefault(cfg.TLSHandshakeTimeoutSeconds, 5)
}

func (cfg ProbeBudgetConfig) DNSTimeout() time.Duration {
	return secondsOrDefault(cfg.DNSTimeoutSeconds, 5)
}

func (cfg ProbeBudgetConfig) PTRTimeout() time.Duration {
	return secondsOrDefault(cfg.PTRTimeoutSeconds, 2)
}

func (cfg ProbeBudgetConfig) MaxDNSRecords() int {
	return intOrDefault(cfg.MaxDNSRecordsPerQuery, 64)
}

func (cfg ProbeBudgetConfig) MaxHTTPResponseBytes() int64 {
	return int64(intOrDefault(cfg.MaxResponseBytes, 1024*1024))
}

func secondsOrDefault(value int, def int) time.Duration {
	return time.Duration(intOrDefault(value, def)) * time.Second
}

func intOrDefault(value int, def int) int {
	if value > 0 {
		return value
	}
	return def
}

func (cfg CollectorV2Config) ProtocolEnabled(protocol string) bool {
	if !cfg.Enabled {
		return false
	}
	switch protocol {
	case "ping":
		return cfg.Protocols.Ping
	case "http":
		return cfg.Protocols.HTTP
	case "dns":
		return cfg.Protocols.DNS
	default:
		return false
	}
}

func (cfg CollectorV2Config) ObservationEnabled(protocol string) bool {
	return cfg.ObservationDB && cfg.ProtocolEnabled(protocol)
}

func (cfg CollectorV2Config) LatestRedisEnabled(protocol string) bool {
	return cfg.LatestRedis && cfg.ProtocolEnabled(protocol)
}
