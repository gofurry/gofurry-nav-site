package bootstrap

import (
	"context"
	"time"

	env "github.com/gofurry/gofurry-admin/config"
	cache "github.com/gofurry/gofurry-admin/internal/infra/cache"
)

func (runtime *Runtime) Live() bool { return true }

func (runtime *Runtime) Started() bool { return runtime != nil && runtime.started.Load() }

func (runtime *Runtime) Ready() bool {
	if !runtime.Started() {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if !runtime.Pools.Ready(ctx) {
		return false
	}
	return !env.GetServerConfig().Redis.Enabled || cache.RedisReady()
}
