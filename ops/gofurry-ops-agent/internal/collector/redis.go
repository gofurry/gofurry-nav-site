package collector

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/config"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/model"
	"github.com/redis/go-redis/v9"
)

func collectRedis(ctx context.Context, cfg config.RedisConfig) model.ServiceCheck {
	result := model.ServiceCheck{Name: cfg.Name}
	checkCtx, cancel := context.WithTimeout(ctx, cfg.Timeout.Duration)
	defer cancel()

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	defer client.Close()

	start := time.Now()
	if err := client.Ping(checkCtx).Err(); err != nil {
		result.LatencyMS = time.Since(start).Milliseconds()
		result.Status = statusFromError(err)
		result.ErrorMessage = err.Error()
		return result
	}
	result.LatencyMS = time.Since(start).Milliseconds()
	result.Status = "ok"
	if info, err := client.Info(checkCtx, "memory", "clients", "keyspace").Result(); err == nil {
		result.MemoryUsed = parseRedisInfoInt(info, "used_memory")
		result.Connections = parseRedisInfoInt(info, "connected_clients")
		result.KeyCount = parseRedisKeyCount(info)
	}
	return result
}

func parseRedisInfoInt(info, key string) int64 {
	for _, line := range strings.Split(info, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, key+":") {
			value := strings.TrimPrefix(line, key+":")
			parsed, _ := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
			return parsed
		}
	}
	return 0
}

func parseRedisKeyCount(info string) int64 {
	var total int64
	for _, line := range strings.Split(info, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "db") {
			continue
		}
		idx := strings.Index(line, "keys=")
		if idx < 0 {
			continue
		}
		rest := line[idx+len("keys="):]
		if comma := strings.IndexByte(rest, ','); comma >= 0 {
			rest = rest[:comma]
		}
		value, _ := strconv.ParseInt(rest, 10, 64)
		total += value
	}
	return total
}
