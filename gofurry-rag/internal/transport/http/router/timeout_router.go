package router

import (
	"time"

	env "github.com/gofurry/gofurry-rag/config"
	"github.com/gofurry/gofurry-rag/internal/api"
	"github.com/gofiber/fiber/v3"
	fibertimeout "github.com/gofiber/fiber/v3/middleware/timeout"
)

type timeoutRouter struct {
	fiber.Router
	config env.TimeoutConfig
}

func wrapTimeoutRouter(router fiber.Router, config env.TimeoutConfig) fiber.Router {
	if !config.Enabled || config.DurationSeconds <= 0 {
		return router
	}
	return &timeoutRouter{Router: router, config: config}
}

func (router *timeoutRouter) Group(prefix string, handlers ...any) fiber.Router {
	return &timeoutRouter{Router: router.Router.Group(prefix, handlers...), config: router.config}
}

func (router *timeoutRouter) Get(path string, handler any, handlers ...any) fiber.Router {
	handler, handlers = router.wrapHandlers(handler, handlers...)
	router.Router.Get(path, handler, handlers...)
	return router
}

func (router *timeoutRouter) Head(path string, handler any, handlers ...any) fiber.Router {
	handler, handlers = router.wrapHandlers(handler, handlers...)
	router.Router.Head(path, handler, handlers...)
	return router
}

func (router *timeoutRouter) Post(path string, handler any, handlers ...any) fiber.Router {
	handler, handlers = router.wrapHandlers(handler, handlers...)
	router.Router.Post(path, handler, handlers...)
	return router
}

func (router *timeoutRouter) Put(path string, handler any, handlers ...any) fiber.Router {
	handler, handlers = router.wrapHandlers(handler, handlers...)
	router.Router.Put(path, handler, handlers...)
	return router
}

func (router *timeoutRouter) Delete(path string, handler any, handlers ...any) fiber.Router {
	handler, handlers = router.wrapHandlers(handler, handlers...)
	router.Router.Delete(path, handler, handlers...)
	return router
}

func (router *timeoutRouter) Options(path string, handler any, handlers ...any) fiber.Router {
	handler, handlers = router.wrapHandlers(handler, handlers...)
	router.Router.Options(path, handler, handlers...)
	return router
}

func (router *timeoutRouter) Patch(path string, handler any, handlers ...any) fiber.Router {
	handler, handlers = router.wrapHandlers(handler, handlers...)
	router.Router.Patch(path, handler, handlers...)
	return router
}

func (router *timeoutRouter) Add(methods []string, path string, handler any, handlers ...any) fiber.Router {
	handler, handlers = router.wrapHandlers(handler, handlers...)
	router.Router.Add(methods, path, handler, handlers...)
	return router
}

func (router *timeoutRouter) All(path string, handler any, handlers ...any) fiber.Router {
	handler, handlers = router.wrapHandlers(handler, handlers...)
	router.Router.All(path, handler, handlers...)
	return router
}

func (router *timeoutRouter) wrapHandlers(handler any, handlers ...any) (any, []any) {
	if len(handlers) == 0 {
		return router.wrapHandler(handler), nil
	}
	return router.wrapHandler(handler), append([]any(nil), handlers...)
}

func (router *timeoutRouter) wrapHandler(handler any) any {
	switch value := handler.(type) {
	case fiber.Handler:
		return fibertimeout.New(value, router.middlewareConfig())
	default:
		return handler
	}
}

func (router *timeoutRouter) middlewareConfig() fibertimeout.Config {
	return fibertimeout.Config{
		Timeout: time.Duration(router.config.DurationSeconds) * time.Second,
		Next: func(c fiber.Ctx) bool {
			return pathExcluded(c.Path(), router.config.ExcludePaths)
		},
		OnTimeout: func(c fiber.Ctx) error {
			return api.ErrorWithCode(c, fiber.StatusRequestTimeout, "request timeout")
		},
	}
}
