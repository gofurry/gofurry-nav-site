package bootstrap

import (
	env "github.com/gofurry/awesome-fiber-template/v3/medium/config"
	cache "github.com/gofurry/awesome-fiber-template/v3/medium/internal/infra/cache"
	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/infra/db"
)

func Live() bool {
	return true
}

func Started() bool {
	return started.Load()
}

func Ready() bool {
	if !Started() {
		return false
	}

	cfg := env.GetServerConfig()
	if cfg.DataBase.Enabled && !db.Databases.ReadyAll() {
		return false
	}
	if cfg.Redis.Enabled && !cache.RedisReady() {
		return false
	}

	return true
}
