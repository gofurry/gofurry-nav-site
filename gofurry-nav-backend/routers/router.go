package routers

/*
 * @Desc: 路由层
 * @author: 福狼
 * @version: v1.0.0
 */

import (
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofiber/contrib/v3/swagger"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/pprof"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofurry/gofurry-nav-backend/common"
	"github.com/gofurry/gofurry-nav-backend/common/util"
	"github.com/gofurry/gofurry-nav-backend/middleware"
	"github.com/gofurry/gofurry-nav-backend/roof/env"
	"github.com/gofurry/monitor"
)

var Router *router

type router struct{}

var (
	navMonitorOnce    sync.Once
	navMonitorHandler http.Handler
)

func NewRouter() *router {
	return &router{}
}

func init() {
	Router = NewRouter()
}

var once = sync.Once{}

func (router *router) Init() *fiber.App {
	once.Do(func() {
	})

	app := fiber.New(fiber.Config{
		AppName:      common.COMMON_PROJECT_NAME,
		ServerHeader: "gofurry-Nav",
		ErrorHandler: customErrorHandler,
		TrustProxy:   false,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		JSONEncoder:  sonic.Marshal,
		JSONDecoder:  sonic.Unmarshal,
	})

	// 注册全局中间件
	registerMiddlewares(app)

	// 路由分组
	registerRoutes(app)

	app.Get("/api/swagger/doc.json", func(c fiber.Ctx) error {
		return c.SendFile("./docs/swagger.json")
	})
	return app
}

func registerRoutes(app *fiber.App) {
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// v1 正式路由：/api/v1/nav/...
	navApi(v1.Group("/nav"))

	if env.GetServerConfig().NavV2.AnyRouteEnabled() {
		v2 := api.Group("/v2")
		navV2Api(v2.Group("/nav"), env.GetServerConfig().NavV2)
	}
}

// registerMiddlewares 注册中间件
func registerMiddlewares(app *fiber.App) {
	cfg := env.GetServerConfig()
	registerMonitor(app)

	// 恢复 panic
	app.Use(recover.New(recover.Config{
		EnableStackTrace: cfg.Server.Mode == "debug", // 仅调试模式打印堆栈
	}))

	// 跨域中间件
	app.Use(cors.New(cors.Config{
		AllowOrigins:     splitAndTrimCSV(cfg.Middleware.Cors.AllowOrigins),
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Content-Length"},
		MaxAge:           86400, // 预检请求缓存 24 小时
	}))

	// 请求限流
	if cfg.Middleware.Limiter.IsOn {
		app.Use(limiter.New(limiter.Config{
			Max:        cfg.Middleware.Limiter.MaxRequests,              // 单位时间最大请求数
			Expiration: cfg.Middleware.Limiter.Expiration * time.Second, // 时间窗口
			KeyGenerator: func(c fiber.Ctx) string {
				return util.GetClientIP(c) // 按可信客户端 IP 限流
			},
			LimitReached: func(c fiber.Ctx) error {
				return common.NewResponse(c).ErrorWithCode("请求过于频繁, 请稍后再试", fiber.StatusTooManyRequests)
			},
		}))
	}

	// WAF 中间件
	if cfg.Waf.WafSwitch {
		app.Use(middleware.CorazaMiddleware())
	}

	// 调试模式专属
	if cfg.Server.Mode == "debug" {
		// pprof 性能分析
		app.Use(pprof.New())

		// Swagger 文档
		if cfg.Middleware.Swagger.IsOn {
			// 校验 Swagger 文件是否存在
			if _, err := os.Stat(cfg.Middleware.Swagger.FilePath); os.IsNotExist(err) {
				panic("Swagger 文件不存在: " + cfg.Middleware.Swagger.FilePath)
			}
			swaggerCfg := swagger.Config{
				BasePath: cfg.Middleware.Swagger.BasePath,
				FilePath: cfg.Middleware.Swagger.FilePath,
				Path:     cfg.Middleware.Swagger.Path,
				Title:    cfg.Middleware.Swagger.Title,
			}
			app.Use(swagger.New(swaggerCfg))
		}
	}

}

func registerMonitor(app *fiber.App) {
	handler := getNavMonitorHandler()
	app.All("/monitor", adaptor.HTTPHandler(handler))
}

func getNavMonitorHandler() http.Handler {
	navMonitorOnce.Do(func() {
		navMonitorHandler = monitor.NewMonitor(http.NotFoundHandler(), monitor.Config{
			Path:            "/monitor",
			Title:           "GoFurry Nav Monitor",
			Description:     "GoFurry navigation backend single-service monitor.",
			DefaultLanguage: "zh-CN",
			DefaultTheme:    "dark",
			Refresh:         5 * time.Second,
		})
	})
	return navMonitorHandler
}

// customErrorHandler 自定义错误处理
func customErrorHandler(c fiber.Ctx, err error) error {
	// 获取错误状态码
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// 标准化错误响应
	response := common.NewResponse(c)
	switch code {
	case fiber.StatusNotFound:
		return response.ErrorWithCode("链接不存在", code)
	case fiber.StatusMethodNotAllowed:
		return response.ErrorWithCode("方法不存在", code)
	case fiber.StatusRequestTimeout:
		return response.ErrorWithCode("请求超时", code)
	default:
		// 生产环境隐藏具体错误信息
		if env.GetServerConfig().Server.Mode != "debug" {
			return response.ErrorWithCode("服务器内部错误", code)
		}
		return response.ErrorWithCode(err.Error(), code)
	}
}

func splitAndTrimCSV(value string) []string {
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
