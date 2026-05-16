package middleware

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofurry/gofurry-oauth-login/common/log"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RateLimitConfig 限流器配置
type RateLimitConfig struct {
	QPS   int           // 每秒允许的请求数
	Burst int           // 突发最大请求数（超过QPS时的缓冲）
	Delay time.Duration // 限流等待延迟（>0时启用等待模式，超时则拒绝）
}

// 默认配置
var defaultRateLimitConfig = RateLimitConfig{
	QPS:   10,
	Burst: 20,
	Delay: 0, // 默认不等待，直接拒绝
}

// 方法级限流器 + 全局限流器 + 统计变量
var (
	methodLimiters = make(map[string]*rate.Limiter) // 方法级限流器（key: 方法全名）
	globalLimiter  *rate.Limiter                    // 全局限流器（兜底用）
	mu             sync.RWMutex                     // 并发安全锁
	rateLimitCount uint64                           // 累计限流次数
)

// InitRateLimiter 初始化全局限流器
func InitRateLimiter(cfg ...RateLimitConfig) {
	var finalCfg RateLimitConfig
	if len(cfg) > 0 {
		finalCfg = cfg[0]
	} else {
		finalCfg = defaultRateLimitConfig
	}

	// 配置合法性校验
	finalCfg = validateRateLimitConfig(finalCfg)

	// 初始化全局限流器
	globalLimiter = rate.NewLimiter(rate.Limit(finalCfg.QPS), finalCfg.Burst)
	log.Info("全局限流器初始化完成，QPS: ", finalCfg.QPS, ", 突发数: ", finalCfg.Burst, ", 等待延迟: ", finalCfg.Delay)
}

// InitMethodRateLimiter 初始化方法级限流器（优先级高于全局）
// methodName: gRPC方法全名（如 "/githuboauth.GithubOAuthService/GetAccessToken"）
// cfg: 该方法专属配置
func InitMethodRateLimiter(methodName string, cfg RateLimitConfig) {
	if methodName == "" {
		log.Error("InitMethodRateLimiter: 方法名不能为空")
		return
	}

	// 配置合法性校验
	finalCfg := validateRateLimitConfig(cfg)

	// 初始化方法级限流器
	limiter := rate.NewLimiter(rate.Limit(finalCfg.QPS), finalCfg.Burst)

	// 并发安全存储
	mu.Lock()
	methodLimiters[methodName] = limiter
	mu.Unlock()

	log.Info("方法级限流器初始化完成，method: ", methodName, ", QPS: ", finalCfg.QPS, ", 突发数: ", finalCfg.Burst, ", 等待延迟: ", finalCfg.Delay)
}

// RateLimitInterceptor gRPC限流器拦截器
func RateLimitInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 捕获异常
	defer func() {
		if r := recover(); r != nil {
			log.Error("限流器拦截器Panic，方法: ", info.FullMethod, ", 异常: ", r)
		}
	}()

	// 优先获取方法级限流器
	var limiter *rate.Limiter
	var cfg RateLimitConfig // 记录当前使用的配置
	mu.RLock()
	methodLimiter, exists := methodLimiters[info.FullMethod]
	mu.RUnlock()

	if exists {
		limiter = methodLimiter
		// 方法级配置默认继承全局限流器的Delay
		mu.RLock()
		if globalLimiter != nil {
			cfg.Delay = defaultRateLimitConfig.Delay
		}
		mu.RUnlock()
	} else {
		// 无方法级配置，使用全局限流器
		if globalLimiter == nil {
			InitRateLimiter() // 自动初始化
		}
		limiter = globalLimiter
		cfg.Delay = defaultRateLimitConfig.Delay
		cfg.QPS = int(globalLimiter.Limit())
		cfg.Burst = globalLimiter.Burst()
	}

	// 执行限流判断
	allow := true
	if cfg.Delay > 0 {
		// 超时则拒绝
		ctxWithTimeout, cancel := context.WithTimeout(ctx, cfg.Delay)
		defer cancel()
		if err := limiter.Wait(ctxWithTimeout); err != nil {
			allow = false
		}
	} else {
		// 直接判断是否允许
		allow = limiter.Allow()
	}

	// 限流触发处理
	if !allow {
		atomic.AddUint64(&rateLimitCount, 1)
		log.Warn(
			"gRPC请求限流触发，方法: %s, 累计限流次数: %d, 当前QPS限制: %d, 突发数: %d",
			info.FullMethod,
			atomic.LoadUint64(&rateLimitCount),
			int(limiter.Limit()),
			limiter.Burst(),
		)
		return nil, status.Error(
			codes.ResourceExhausted,
			"当前请求过多，请稍后重试（错误码：429）",
		)
	}

	// 限流通过，执行业务逻辑
	return handler(ctx, req)
}

// GetRateLimitCount 获取累计限流次数
func GetRateLimitCount() uint64 {
	return atomic.LoadUint64(&rateLimitCount)
}

// 私有工具：校验限流器配置合法性
func validateRateLimitConfig(cfg RateLimitConfig) RateLimitConfig {
	if cfg.QPS <= 0 {
		cfg.QPS = 10
		log.Warn("限流器QPS配置无效，使用默认值10")
	}
	if cfg.Burst <= 0 {
		cfg.Burst = cfg.QPS * 2
		log.Warn("限流器Burst配置无效，自动设为QPS的2倍：", cfg.Burst)
	}
	if cfg.Delay < 0 {
		cfg.Delay = 0
		log.Warn("限流器Delay配置不能为负，使用默认值0（不等待）")
	}
	return cfg
}
