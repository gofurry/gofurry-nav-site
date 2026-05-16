package service

/*
 * @Desc: redis服务
 * @author: 福狼
 * @version: v1.0.0
 */

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/common/log"
	"github.com/gofurry/gofurry-game-backend/roof/env"
	"github.com/redis/go-redis/v9"
)

var client *redis.Client
var ctx = context.Background()

func GetRedisService() *redis.Client { return client }

func InitRedisOnStart() {
	connect()
}

func connect() {
	connCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	client = redis.NewClient(&redis.Options{
		Addr:      env.GetServerConfig().Redis.RedisAddr,
		Password:  env.GetServerConfig().Redis.RedisPassword,
		DB:        0,
		OnConnect: OnConnectFunc,
	})
	_, err := client.Ping(connCtx).Result()
	if err != nil {
		panic("Failed to connect to Redis:" + err.Error())
	}
	log.Debug("connected to redis ok.")

}

func OnConnectFunc(ctx context.Context, cn *redis.Conn) error {
	log.Debug("new redis connect...")
	return nil
}

func Del(keys ...string) common.GFError {
	err := client.Del(ctx, keys...).Err()
	if err != nil {
		log.Error("删除缓存失败..." + err.Error())
		return common.NewServiceError("删除缓存失败.")
	}
	return nil
}

func SetNX(key string, value any, expiration time.Duration) bool {
	bool, err := client.SetNX(ctx, key, value, expiration).Result()
	if err != nil {
		log.Error("设置缓存失败..." + err.Error())
		return false
	}
	return bool
}

func Set(key string, value any) common.GFError {
	err := SetExpire(key, value, 0)
	return err
}

func SetExpire(key string, value any, expiration time.Duration) common.GFError {
	err := client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		log.Error("设置缓存失败..." + err.Error())
		return common.NewServiceError("设置缓存失败.")
	}
	return nil
}

func Get(key string) *redis.Cmd {
	val := client.Do(ctx, "get", key)
	return val
}

func GetString(key string) (data string, gfsError common.GFError) {
	val, err := client.Get(ctx, key).Result()

	switch {
	case errors.Is(err, redis.Nil):
		return "", nil
	case err != nil:
		log.Error("获取缓存失败..." + err.Error())
		return "", common.NewServiceError("获取缓存失败.")
	}
	return strings.TrimSpace(val), nil
}

func HSetMap(key string, kvMap map[string]string) common.GFError {
	err := client.HSet(ctx, key, kvMap).Err()
	if err != nil {
		log.Error("设置缓存失败..." + err.Error())
		return common.NewServiceError("设置缓存失败.")
	}
	return nil
}

func HSet(key string, fieldName string, fieldVal string) common.GFError {
	err := client.HSet(ctx, key, fieldName, fieldVal).Err()
	if err != nil {
		log.Error("设置缓存失败..." + err.Error())
		return common.NewServiceError("设置缓存失败.")
	}
	return nil
}

func HGet(key string, fieldName string) (data string, gfsError common.GFError) {
	res, err := client.HGet(ctx, key, fieldName).Result()
	switch {
	case errors.Is(err, redis.Nil):
		return "", common.NewServiceError(key + "缓存不存在.")
	case err != nil:
		log.Error("获取缓存失败..." + err.Error())
		return "", common.NewServiceError("获取缓存失败.")
	}
	return res, nil
}

func HMGet(key string, fields ...string) (data []interface{}, gfsError common.GFError) {
	res, err := client.HMGet(ctx, key, fields...).Result()
	switch {
	case errors.Is(err, redis.Nil):
		return nil, common.NewServiceError(key + "缓存不存在.")
	case err != nil:
		log.Error("获取缓存失败..." + err.Error())
		return nil, common.NewServiceError("获取缓存失败.")
	}
	return res, nil
}

func HGetAll(key string) (data map[string]string, gfsError common.GFError) {
	res, err := client.HGetAll(ctx, key).Result()
	switch {
	case errors.Is(err, redis.Nil):
		return nil, common.NewServiceError(key + "缓存不存在.")
	case err != nil:
		log.Error("获取缓存失败..." + err.Error())
		return nil, common.NewServiceError("获取缓存失败.")
	}
	return res, nil
}

func HDel(key string, fields ...string) (res int64, gfsError common.GFError) {
	intVal, err := client.HDel(ctx, key, fields...).Result()
	switch {
	case errors.Is(err, redis.Nil):
		return 0, common.NewServiceError(key + "缓存不存在.")
	case err != nil:
		log.Error("获取缓存失败..." + err.Error())
		return intVal, nil
	}
	return intVal, nil
}

func Incr(key string) {
	client.Incr(ctx, key)
}

// redis 前缀统计
func CountByPrefix(prefix string) (res int64, gfsError common.GFError) {
	var cursor uint64 = 0
	var count int
	pattern := prefix + "*" // 匹配指定前缀的键

	for {
		// SCAN 命令，返回匹配的键和新的游标
		keys, newCursor, err := client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return 0, common.NewServiceError("redis统计失败.")
		}

		count += len(keys) // 累加匹配的键数
		cursor = newCursor // 更新游标

		if cursor == 0 {
			break // 游标为 0 时，扫描完成
		}
	}

	return int64(count), nil
}

// redis 前缀删除
func DelByPrefix(prefix string) common.GFError {
	var cursor uint64 = 0
	pattern := prefix + "*" // 匹配指定前缀的键

	for {
		// SCAN 命令，返回匹配的键和新的游标
		keys, newCursor, err := client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			log.Error(fmt.Sprintf("redis scan err:%v", err))
			return common.NewServiceError(err.Error())
		}
		if len(keys) != 0 {
			err := Del(keys...)
			if err != nil {
				log.Error(fmt.Sprintf("redis del err:%v", err))
				return err
			}
		}

		cursor = newCursor // 更新游标
		if cursor == 0 {
			break // 游标为 0 时，扫描完成
		}
	}
	return nil
}

// redis 前缀查询
func FindByPrefix(prefix string) ([]string, common.GFError) {
	var cursor uint64 = 0
	var resList []string
	pattern := prefix + "*" // 匹配指定前缀的键

	for {
		// SCAN 命令，返回匹配的键和新的游标
		keys, newCursor, err := client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			log.Error(fmt.Sprintf("redis scan err:%v", err))
			return nil, common.NewServiceError(err.Error())
		}
		if len(keys) != 0 {
			for idx := range keys {
				resList = append(resList, keys[idx])
			}
		}

		cursor = newCursor // 更新游标
		if cursor == 0 {
			break // 游标为 0 时，扫描完成
		}
	}
	return resList, nil
}
