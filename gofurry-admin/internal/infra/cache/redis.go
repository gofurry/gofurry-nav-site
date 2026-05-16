package cache

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	env "github.com/gofurry/awesome-fiber-template/v3/medium/config"
	log "github.com/gofurry/awesome-fiber-template/v3/medium/internal/infra/logging"
	"github.com/gofurry/awesome-fiber-template/v3/medium/pkg/common"
	"github.com/redis/go-redis/v9"
)

var client *redis.Client
var backgroundContext = context.Background()

func GetRedisService() *redis.Client { return client }

func RedisReady() bool {
	if client == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return client.Ping(ctx).Err() == nil
}

func InitRedisOnStart() error {
	return connect()
}

func Close() error {
	if client == nil {
		return nil
	}

	err := client.Close()
	client = nil
	return err
}

func connect() error {
	cfg := env.GetServerConfig().Redis
	connCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client = redis.NewClient(&redis.Options{
		Addr:      cfg.RedisAddr,
		Username:  cfg.RedisUsername,
		Password:  cfg.RedisPassword,
		DB:        cfg.RedisDB,
		PoolSize:  cfg.RedisPoolSize,
		OnConnect: onConnect,
	})

	if _, err := client.Ping(connCtx).Result(); err != nil {
		client = nil
		return fmt.Errorf("connect to redis failed: %w", err)
	}

	log.Debug("redis connected")
	return nil
}

func onConnect(ctx context.Context, cn *redis.Conn) error {
	log.Debug("redis connection opened")
	return nil
}

func Del(keys ...string) common.Error {
	if client == nil {
		return common.NewServiceError("redis service is not ready")
	}
	if err := client.Del(backgroundContext, keys...).Err(); err != nil {
		log.Errorf("redis DEL failed: %v", err)
		return common.NewServiceError("delete redis keys failed")
	}
	return nil
}

func SetNX(key string, value any, expiration time.Duration) bool {
	if client == nil {
		return false
	}

	ok, err := client.SetNX(backgroundContext, key, value, expiration).Result()
	if err != nil {
		log.Errorf("redis SETNX failed: %v", err)
		return false
	}
	return ok
}

func Set(key string, value any) common.Error {
	return SetExpire(key, value, 0)
}

func SetExpire(key string, value any, expiration time.Duration) common.Error {
	if client == nil {
		return common.NewServiceError("redis service is not ready")
	}
	if err := client.Set(backgroundContext, key, value, expiration).Err(); err != nil {
		log.Errorf("redis SET failed: %v", err)
		return common.NewServiceError("set redis key failed")
	}
	return nil
}

func Get(key string) *redis.Cmd {
	if client == nil {
		cmd := redis.NewCmd(backgroundContext)
		cmd.SetErr(errors.New("redis service is not ready"))
		return cmd
	}
	return client.Do(backgroundContext, "GET", key)
}

func GetString(key string) (string, common.Error) {
	if client == nil {
		return "", common.NewServiceError("redis service is not ready")
	}

	val, err := client.Get(backgroundContext, key).Result()
	switch {
	case errors.Is(err, redis.Nil):
		return "", nil
	case err != nil:
		log.Errorf("redis GET failed: %v", err)
		return "", common.NewServiceError("get redis key failed")
	default:
		return strings.TrimSpace(val), nil
	}
}

func HSetMap(key string, kvMap map[string]string) common.Error {
	if client == nil {
		return common.NewServiceError("redis service is not ready")
	}
	if err := client.HSet(backgroundContext, key, kvMap).Err(); err != nil {
		log.Errorf("redis HSET map failed: %v", err)
		return common.NewServiceError("set redis hash failed")
	}
	return nil
}

func HSet(key string, fieldName string, fieldVal string) common.Error {
	if client == nil {
		return common.NewServiceError("redis service is not ready")
	}
	if err := client.HSet(backgroundContext, key, fieldName, fieldVal).Err(); err != nil {
		log.Errorf("redis HSET failed: %v", err)
		return common.NewServiceError("set redis hash field failed")
	}
	return nil
}

func HGet(key string, fieldName string) (string, common.Error) {
	if client == nil {
		return "", common.NewServiceError("redis service is not ready")
	}

	res, err := client.HGet(backgroundContext, key, fieldName).Result()
	switch {
	case errors.Is(err, redis.Nil):
		return "", common.NewServiceError(fmt.Sprintf("redis hash field not found: %s.%s", key, fieldName))
	case err != nil:
		log.Errorf("redis HGET failed: %v", err)
		return "", common.NewServiceError("get redis hash field failed")
	default:
		return res, nil
	}
}

func HMGet(key string, fields ...string) ([]interface{}, common.Error) {
	if client == nil {
		return nil, common.NewServiceError("redis service is not ready")
	}

	res, err := client.HMGet(backgroundContext, key, fields...).Result()
	switch {
	case errors.Is(err, redis.Nil):
		return nil, common.NewServiceError("redis hash not found")
	case err != nil:
		log.Errorf("redis HMGET failed: %v", err)
		return nil, common.NewServiceError("get redis hash fields failed")
	default:
		return res, nil
	}
}

func HGetAll(key string) (map[string]string, common.Error) {
	if client == nil {
		return nil, common.NewServiceError("redis service is not ready")
	}

	res, err := client.HGetAll(backgroundContext, key).Result()
	switch {
	case errors.Is(err, redis.Nil):
		return nil, common.NewServiceError("redis hash not found")
	case err != nil:
		log.Errorf("redis HGETALL failed: %v", err)
		return nil, common.NewServiceError("get redis hash failed")
	default:
		return res, nil
	}
}

func HDel(key string, fields ...string) (int64, common.Error) {
	if client == nil {
		return 0, common.NewServiceError("redis service is not ready")
	}

	res, err := client.HDel(backgroundContext, key, fields...).Result()
	switch {
	case errors.Is(err, redis.Nil):
		return 0, common.NewServiceError("redis hash not found")
	case err != nil:
		log.Errorf("redis HDEL failed: %v", err)
		return 0, common.NewServiceError("delete redis hash fields failed")
	default:
		return res, nil
	}
}

func Incr(key string) {
	if client == nil {
		return
	}
	client.Incr(backgroundContext, key)
}

func CountByPrefix(prefix string) (int64, common.Error) {
	if client == nil {
		return 0, common.NewServiceError("redis service is not ready")
	}

	var (
		cursor uint64
		count  int
	)
	pattern := prefix + "*"

	for {
		keys, newCursor, err := client.Scan(backgroundContext, cursor, pattern, 100).Result()
		if err != nil {
			return 0, common.NewServiceError("scan redis keys failed")
		}

		count += len(keys)
		cursor = newCursor
		if cursor == 0 {
			break
		}
	}

	return int64(count), nil
}

func DelByPrefix(prefix string) common.Error {
	if client == nil {
		return common.NewServiceError("redis service is not ready")
	}

	var cursor uint64
	pattern := prefix + "*"

	for {
		keys, newCursor, err := client.Scan(backgroundContext, cursor, pattern, 100).Result()
		if err != nil {
			log.Errorf("redis scan failed: %v", err)
			return common.NewServiceError("scan redis keys failed")
		}

		if len(keys) > 0 {
			if err := Del(keys...); err != nil {
				return err
			}
		}

		cursor = newCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}

func FindByPrefix(prefix string) ([]string, common.Error) {
	if client == nil {
		return nil, common.NewServiceError("redis service is not ready")
	}

	var (
		cursor  uint64
		results []string
	)
	pattern := prefix + "*"

	for {
		keys, newCursor, err := client.Scan(backgroundContext, cursor, pattern, 100).Result()
		if err != nil {
			log.Errorf("redis scan failed: %v", err)
			return nil, common.NewServiceError("scan redis keys failed")
		}

		if len(keys) > 0 {
			results = append(results, keys...)
		}

		cursor = newCursor
		if cursor == 0 {
			break
		}
	}

	return results, nil
}

func PipelineExec(fn func(pipe redis.Pipeliner)) common.Error {
	if client == nil {
		return common.NewServiceError("redis service is not ready")
	}

	pipe := client.Pipeline()
	fn(pipe)

	if _, err := pipe.Exec(backgroundContext); err != nil {
		log.Errorf("redis pipeline execution failed: %v", err)
		return common.NewServiceError("execute redis pipeline failed")
	}
	return nil
}
