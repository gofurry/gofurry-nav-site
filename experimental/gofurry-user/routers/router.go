package routers

/*
 * @Desc: 路由层
 * @author: 福狼
 * @version: v1.0.0
 */

import (
	"sync"

	"github.com/gofurry/gofurry-user/common"
	"github.com/gofurry/gofurry-user/middleware"
	"github.com/gofurry/gofurry-user/roof/env"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

var Router *router

type router struct{}

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
		// 使用 ipv6 不用 NAT 速度更快, 降低被扫地址的可能性
		Network:      fiber.NetworkTCP4, // tcp tcp4 tcp6 三种模式
		AppName:      common.COMMON_PROJECT_NAME,
		ServerHeader: "gofurry-User",
		Prefork:      false, // 多核cpu处理高并发 业务量小需关闭
		// 在生产环境禁用错误堆栈跟踪
		EnablePrintRoutes: env.GetServerConfig().Server.Mode == "debug",
		// 配置默认404处理
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// 检查是否是404错误
			if err == fiber.ErrNotFound {
				return common.NewResponse(c).Error("链接不存在")
			}
			// 检查是否是405错误
			if err == fiber.ErrMethodNotAllowed {
				return common.NewResponse(c).Error("方法不存在")
			}
			// 其他错误
			return common.NewResponse(c).Error(err.Error())
		},
		EnableTrustedProxyCheck: true, // 信任 Nginx 反向代理
	})

	cfg := swagger.Config{
		BasePath: env.GetServerConfig().Middleware.Swagger.BasePath,
		FilePath: env.GetServerConfig().Middleware.Swagger.FilePath,
		Path:     env.GetServerConfig().Middleware.Swagger.Path,
		Title:    env.GetServerConfig().Middleware.Swagger.Title,
	}
	// 中间件
	if env.GetServerConfig().Middleware.Swagger.IsOn == "on" {
		app.Use(swagger.New(cfg)) // 访问路径类似 https://[::1]:9999/swagger
	}
	// 调试模式下启用pprof
	if env.GetServerConfig().Server.Mode == "debug" {
		app.Use(pprof.New())
	}
	// 跨域
	app.Use(cors.New())
	//app.Use(cors.New(cors.Config{
	//	AllowOrigins:     env.GetServerConfig().Middleware.Cors.AllowOrigins,
	//	AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
	//	AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
	//	AllowCredentials: true,
	//}))
	// 恢复
	app.Use(recover.New())
	if env.GetServerConfig().Waf.WafSwitch == "on" {
		app.Use(middleware.CorazaMiddleware()) // CorazaWAF
	}

	// 路由分组
	userApi(app.Group("/api/user"))
	utilApi(app.Group("/api/util"))
	oauthApi(app.Group("/oauth"))

	app.Get("/api/swagger/doc.json", func(c *fiber.Ctx) error {
		return c.SendFile("./docs/swagger.json")
	})

	return app
}
