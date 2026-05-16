package middleware

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofurry/gofurry-nav-backend/common/log"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gofurry/gofurry-nav-backend/metrics"
)

/*
 * @Desc: Prometheus 全局中间件
 * @author: 福狼
 * @version: v1.0.1
 */

var (
	MetricsHandler    fiber.Handler
	once, routesOnce  sync.Once
	registeredRoutes  map[string]struct{}
	skipPaths         map[string]bool
	ignoreStatusCodes map[int]bool
)

// FiberPromConf Prometheus 全局中间件配置
type FiberPromConf struct {
	SkipPaths         []string
	IgnoreStatusCodes []int
}

// InitPrometheus 初始化 Prometheus 全局中间件
func InitPrometheus(cfg ...FiberPromConf) {
	log.Debug("[InitPrometheus init] try to init prometheus middleware...")
	once.Do(func() {
		// 初始化默认配置
		conf := FiberPromConf{}
		if len(cfg) > 0 {
			conf = cfg[0]
		}

		// 默认跳过/metrics
		defaultSkipPaths := []string{"/metrics"}
		if conf.SkipPaths != nil {
			conf.SkipPaths = append(defaultSkipPaths, conf.SkipPaths...)
		} else {
			conf.SkipPaths = defaultSkipPaths
		}

		// 设置忽略的 status code 和 path
		if conf.SkipPaths != nil {
			setSkipPaths(conf.SkipPaths)
		}
		if conf.IgnoreStatusCodes != nil {
			setIgnoreStatusCodes(conf.IgnoreStatusCodes)
		}

		// 注册实例
		registry := prometheus.DefaultRegisterer
		registry.MustRegister(metrics.HttpRequestsTotal)
		registry.MustRegister(metrics.HttpRequestDuration)
		registry.MustRegister(metrics.HttpActiveRequests)
		MetricsHandler = adaptor.HTTPHandler(promhttp.Handler())
	})
	log.Debug("[InitPrometheus init] init prometheus middleware ok.")
}

// PrometheusMiddleware Prometheus 全局中间件
func PrometheusMiddleware(c fiber.Ctx) error {
	// method
	method := c.Method()

	// QPS
	metrics.HttpActiveRequests.Inc()
	defer metrics.HttpActiveRequests.Dec()

	// 开始计时
	start := time.Now()

	// 栈内后续操作
	err := c.Next()

	// 记录已注册的路由
	routesOnce.Do(func() {
		registeredRoutes = make(map[string]struct{})
		for _, r := range c.App().GetRoutes(true) {
			p := r.Path
			if p != "" && p != "/" {
				p = normalizePath(p)
			}
			registeredRoutes[r.Method+" "+p] = struct{}{}
		}
	})
	// 获取路由
	routePath := c.Route().Path
	if routePath == "/" {
		routePath = c.Path()
	}
	if routePath != "" && routePath != "/" {
		routePath = normalizePath(routePath)
	}
	// 跳过未注册的路由
	if _, ok := registeredRoutes[method+" "+routePath]; !ok {
		log.Warn("[Try to req unregistered route] 尝试请求未注册的路由: ", method+" "+routePath)
		return err
	}
	// 跳过忽略的路由
	if skipPaths[routePath] {
		return nil
	}

	// 获取状态码
	status := fiber.StatusInternalServerError
	if err != nil {
		if e, ok := err.(*fiber.Error); ok {
			status = e.Code
		}
	} else {
		status = c.Response().StatusCode()
	}
	// 跳过忽略的状态码
	if ignoreStatusCodes[status] {
		return err
	}

	// 更新指标
	writeNormalMetrics(method, routePath, strconv.Itoa(status), start)

	return err
}

// writeNormalMetrics 更新指标
func writeNormalMetrics(method string, path string, status string, start time.Time) {
	metrics.HttpRequestsTotal.WithLabelValues(method, path, status).Inc()
	metrics.HttpRequestDuration.WithLabelValues(method, path).Observe(time.Since(start).Seconds())
}

// =============== internal util ===============

// normalizePath 标准化路径
func normalizePath(routePath string) string {
	normalized := strings.TrimRight(routePath, "/")
	if normalized == "" {
		return "/"
	}
	return normalized
}

// setSkipPaths 设置跳过路径
func setSkipPaths(paths []string) {
	if skipPaths == nil {
		skipPaths = make(map[string]bool)
	}
	for _, path := range paths {
		skipPaths[path] = true
	}
}

// setIgnoreStatusCodes 设置跳过状态码
func setIgnoreStatusCodes(codes []int) {
	if ignoreStatusCodes == nil {
		ignoreStatusCodes = make(map[int]bool)
	}
	for _, code := range codes {
		ignoreStatusCodes[code] = true
	}
}
