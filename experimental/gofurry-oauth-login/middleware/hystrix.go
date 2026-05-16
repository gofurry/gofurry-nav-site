package middleware

import (
	"fmt"
	"strings"

	"github.com/gofurry/gofurry-oauth-login/common/log"
	"github.com/afex/hystrix-go/hystrix"
)

// HystrixConfig 熔断器配置
type HystrixConfig struct {
	Timeout                int // 单次请求超时时间（毫秒）
	MaxConcurrentRequests  int // 最大并发数
	ErrorPercentThreshold  int // 触发熔断的失败率阈值（0-100）
	SleepWindow            int // 熔断后休眠时间（毫秒）
	RequestVolumeThreshold int // 触发熔断的最小请求数（避免少量失败误触发）
}

// 默认配置
var defaultHystrixConfig = HystrixConfig{
	Timeout:                10000, // 10秒超时
	MaxConcurrentRequests:  20,    // 最大20并发
	ErrorPercentThreshold:  50,    // 失败率50%触发熔断
	SleepWindow:            5000,  // 熔断后休眠5秒
	RequestVolumeThreshold: 10,    // 最少10次请求才触发熔断
}

// InitHystrix 初始化熔断器
// serviceName
// methodName
// cfg: 自定义配置
func InitHystrix(serviceName, methodName string, cfg ...HystrixConfig) {
	commandKey := GetHystrixCommandKey(serviceName, methodName)
	var finalCfg HystrixConfig

	if len(cfg) > 0 {
		finalCfg = cfg[0]
	} else {
		finalCfg = defaultHystrixConfig
	}

	hystrixCfg := hystrix.CommandConfig{
		Timeout:                finalCfg.Timeout,
		MaxConcurrentRequests:  finalCfg.MaxConcurrentRequests,
		ErrorPercentThreshold:  finalCfg.ErrorPercentThreshold,
		SleepWindow:            finalCfg.SleepWindow,
		RequestVolumeThreshold: finalCfg.RequestVolumeThreshold,
	}

	// 配置熔断器
	hystrix.ConfigureCommand(commandKey, hystrixCfg)
	log.Info("熔断器初始化完成，commandKey: ", commandKey, ", 配置: ", finalCfg)
}

// HystrixDo 执行熔断器逻辑
// commandKey: 熔断器唯一标识
// run: 正常业务逻辑
// fallback: 熔断降级逻辑
// errorFilter: 可选，自定义失败判断（返回true表示计入失败率）
func HystrixDo(
	commandKey string,
	run func() error,
	fallback func(error) error,
	errorFilter ...func(error) bool,
) error {
	// 捕获异常
	defer func() {
		if r := recover(); r != nil {
			log.Error("熔断器执行Panic，commandKey: ", commandKey, ", 异常: ", r)
		}
	}()

	// 校验commandKey
	if commandKey == "" {
		log.Error("HystrixDo: commandKey不能为空")
		return fallback(fmt.Errorf("熔断器配置错误"))
	}

	// 默认错误过滤器
	defaultFilter := func(err error) bool {
		if err == nil {
			return false
		}
		errMsg := err.Error()
		// 匹配网络/超时相关错误
		return strings.Contains(errMsg, "timeout") ||
			strings.Contains(errMsg, "connection") ||
			strings.Contains(errMsg, "dial") ||
			strings.Contains(errMsg, "refused") ||
			strings.Contains(errMsg, "unreachable")
	}

	// 优先使用用户自定义过滤器
	filter := defaultFilter
	if len(errorFilter) > 0 {
		filter = errorFilter[0]
	}

	// 根据过滤器判断是否计入失败
	wrappedRun := func() error {
		err := run()
		if err != nil && !filter(err) {
			return nil
		}
		return err
	}

	// fallback
	wrappedFallback := func(err error) error {
		log.Error("熔断器触发，commandKey: ", commandKey, ", 错误原因: ", err)
		return fallback(err)
	}

	return hystrix.Do(commandKey, wrappedRun, wrappedFallback)
}

// GetHystrixCommandKey 生成熔断器唯一标识
func GetHystrixCommandKey(serviceName, methodName string) string {
	return serviceName + "_" + methodName
}

// GetCircuitState 获取熔断器当前状态
func GetCircuitState(commandKey string) string {
	if commandKey == "" {
		log.Error("GetCircuitState: commandKey不能为空")
		return "unknown"
	}

	circuit, exists, err := hystrix.GetCircuit(commandKey)
	if err != nil {
		log.Error("获取熔断器电路失败，err: ", err)
		return "error"
	}
	if !exists {
		log.Error("熔断器电路不存在，commandKey: ", commandKey)
		return "unknown"
	}

	if circuit.IsOpen() {
		return "open"
	}
	return "closed"
}
