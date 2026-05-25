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
	Scheduler   SchedulerConfig   `yaml:"scheduler"`
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
	LightProbe    LightProbeConfig   `yaml:"light_probe"`
	EdgeHints     EdgeHintsConfig    `yaml:"edge_hints"`
}

type CollectorProtocols struct {
	Ping bool `yaml:"ping"`
	HTTP bool `yaml:"http"`
	DNS  bool `yaml:"dns"`
}

type EdgeHintsConfig struct {
	Enabled *bool `yaml:"enabled"`
}

type LightProbeConfig struct {
	RDAP        LightProbeRDAPConfig        `yaml:"rdap"`
	Robots      LightProbeRobotsConfig      `yaml:"robots"`
	SecurityTXT LightProbeSecurityTXTConfig `yaml:"security_txt"`
	PageAssets  LightProbePageAssetsConfig  `yaml:"page_assets"`
	PortCheck   LightProbePortCheckConfig   `yaml:"port_check"`
	WAFCanary   LightProbeWAFCanaryConfig   `yaml:"waf_canary"`
}

type LightProbeRDAPConfig struct {
	Enabled        bool `yaml:"enabled"`
	IntervalHours  int  `yaml:"interval_hours"`
	TimeoutSeconds int  `yaml:"timeout_seconds"`
}

type LightProbeRobotsConfig struct {
	Enabled          bool `yaml:"enabled"`
	IntervalHours    int  `yaml:"interval_hours"`
	TimeoutSeconds   int  `yaml:"timeout_seconds"`
	MaxResponseBytes int  `yaml:"max_response_bytes"`
	MaxSitemapLinks  int  `yaml:"max_sitemap_links"`
}

type LightProbeSecurityTXTConfig struct {
	Enabled          bool `yaml:"enabled"`
	IntervalHours    int  `yaml:"interval_hours"`
	TimeoutSeconds   int  `yaml:"timeout_seconds"`
	MaxResponseBytes int  `yaml:"max_response_bytes"`
}

type LightProbePageAssetsConfig struct {
	Enabled           bool     `yaml:"enabled"`
	IntervalHours     int      `yaml:"interval_hours"`
	TimeoutSeconds    int      `yaml:"timeout_seconds"`
	MaxIconBytes      int      `yaml:"max_icon_bytes"`
	MaxManifestBytes  int      `yaml:"max_manifest_bytes"`
	AllowedAssetHosts []string `yaml:"allowed_asset_hosts"`
}

type LightProbePortCheckConfig struct {
	Enabled           bool  `yaml:"enabled"`
	IntervalHours     int   `yaml:"interval_hours"`
	TimeoutSeconds    int   `yaml:"timeout_seconds"`
	Concurrency       int   `yaml:"concurrency"`
	MaxPortsPerTarget int   `yaml:"max_ports_per_target"`
	Ports             []int `yaml:"ports"`
}

type LightProbeWAFCanaryConfig struct {
	Enabled          bool   `yaml:"enabled"`
	IntervalHours    int    `yaml:"interval_hours"`
	TimeoutSeconds   int    `yaml:"timeout_seconds"`
	CanaryPath       string `yaml:"canary_path"`
	UserAgent        string `yaml:"user_agent"`
	RunOnStart       bool   `yaml:"run_on_start"`
	MaxTargetsPerRun int    `yaml:"max_targets_per_run"`
}

type SchedulerConfig struct {
	CollectorID      string `yaml:"collector_id"`
	LeaseEnabled     bool   `yaml:"lease_enabled"`
	LeaseTTLSeconds  int    `yaml:"lease_ttl_seconds"`
	RunStateRedis    *bool  `yaml:"run_state_redis"`
	RunStateTTLHours int    `yaml:"run_state_ttl_hours"`
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
	case "rdap":
		return cfg.LightProbe.RDAP.Enabled
	case "robots":
		return cfg.LightProbe.Robots.Enabled
	case "security_txt":
		return cfg.LightProbe.SecurityTXT.Enabled
	case "page_assets":
		return cfg.LightProbe.PageAssets.Enabled
	case "port_check":
		return cfg.LightProbe.PortCheck.Enabled
	case "waf_canary":
		return cfg.LightProbe.WAFCanary.Enabled
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

func (cfg EdgeHintsConfig) EnabledOrDefault() bool {
	if cfg.Enabled == nil {
		return true
	}
	return *cfg.Enabled
}

func (cfg SchedulerConfig) RunStateEnabled() bool {
	if cfg.RunStateRedis == nil {
		return true
	}
	return *cfg.RunStateRedis
}

func (cfg SchedulerConfig) RunStateTTL() time.Duration {
	hours := cfg.RunStateTTLHours
	if hours <= 0 {
		hours = 168
	}
	return time.Duration(hours) * time.Hour
}

func (cfg LightProbeRDAPConfig) Interval() time.Duration {
	return hoursOrDefault(cfg.IntervalHours, 168)
}

func (cfg LightProbeRDAPConfig) Timeout() time.Duration {
	return secondsOrDefault(cfg.TimeoutSeconds, 10)
}

func (cfg LightProbeRobotsConfig) Interval() time.Duration {
	return hoursOrDefault(cfg.IntervalHours, 168)
}

func (cfg LightProbeRobotsConfig) Timeout() time.Duration {
	return secondsOrDefault(cfg.TimeoutSeconds, 10)
}

func (cfg LightProbeRobotsConfig) MaxResponseSize() int64 {
	return int64(intOrDefault(cfg.MaxResponseBytes, 64*1024))
}

func (cfg LightProbeRobotsConfig) MaxSitemaps() int {
	return intOrDefault(cfg.MaxSitemapLinks, 20)
}

func (cfg LightProbeSecurityTXTConfig) Interval() time.Duration {
	return hoursOrDefault(cfg.IntervalHours, 168)
}

func (cfg LightProbeSecurityTXTConfig) Timeout() time.Duration {
	return secondsOrDefault(cfg.TimeoutSeconds, 10)
}

func (cfg LightProbeSecurityTXTConfig) MaxResponseSize() int64 {
	return int64(intOrDefault(cfg.MaxResponseBytes, 64*1024))
}

func (cfg LightProbePageAssetsConfig) Interval() time.Duration {
	return hoursOrDefault(cfg.IntervalHours, 168)
}

func (cfg LightProbePageAssetsConfig) Timeout() time.Duration {
	return secondsOrDefault(cfg.TimeoutSeconds, 10)
}

func (cfg LightProbePageAssetsConfig) MaxIconSize() int64 {
	return int64(intOrDefault(cfg.MaxIconBytes, 256*1024))
}

func (cfg LightProbePageAssetsConfig) MaxManifestSize() int64 {
	return int64(intOrDefault(cfg.MaxManifestBytes, 64*1024))
}

func (cfg LightProbePortCheckConfig) Interval() time.Duration {
	return hoursOrDefault(cfg.IntervalHours, 168)
}

func (cfg LightProbePortCheckConfig) Timeout() time.Duration {
	return secondsOrDefault(cfg.TimeoutSeconds, 2)
}

func (cfg LightProbePortCheckConfig) WorkerCount() int {
	return intOrDefault(cfg.Concurrency, 8)
}

func (cfg LightProbePortCheckConfig) MaxPorts() int {
	return intOrDefault(cfg.MaxPortsPerTarget, 24)
}

func (cfg LightProbeWAFCanaryConfig) Interval() time.Duration {
	return hoursOrDefault(cfg.IntervalHours, 720)
}

func (cfg LightProbeWAFCanaryConfig) Timeout() time.Duration {
	return secondsOrDefault(cfg.TimeoutSeconds, 10)
}

func (cfg LightProbeWAFCanaryConfig) Path() string {
	path := strings.TrimSpace(cfg.CanaryPath)
	if path == "" {
		return "/.well-known/gofurry-waf-canary"
	}
	return path
}

func (cfg LightProbeWAFCanaryConfig) UserAgentOrDefault() string {
	userAgent := strings.TrimSpace(cfg.UserAgent)
	if userAgent == "" {
		return "GoFurry-Nav-Collector-WAF-Canary/1.0"
	}
	if !strings.Contains(userAgent, "WAF-Canary") {
		return userAgent + " GoFurry-Nav-Collector-WAF-Canary/1.0"
	}
	return userAgent
}

func (cfg LightProbeWAFCanaryConfig) MaxTargets() int {
	if cfg.MaxTargetsPerRun < 0 {
		return 0
	}
	return cfg.MaxTargetsPerRun
}

func hoursOrDefault(value int, def int) time.Duration {
	return time.Duration(intOrDefault(value, def)) * time.Hour
}
