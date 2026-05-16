package service

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/gofurry/gofurry-nav-backend/apps/system/stat/models"
	"github.com/gofurry/gofurry-nav-backend/common"
	"github.com/gofurry/gofurry-nav-backend/common/log"
	cs "github.com/gofurry/gofurry-nav-backend/common/service"
	"github.com/gofurry/gofurry-nav-backend/common/util"
	"github.com/gofurry/gofurry-nav-backend/roof/env"
	"github.com/redis/go-redis/v9"
)

type metricsService struct{}

var metricsSingleton = new(metricsService)

func GetMetricsService() *metricsService { return metricsSingleton }

func (s metricsService) GetPromMetrics() (vo models.PromMetricsVo, err common.GFError) {
	vo.Node, _ = cs.HGetAll("prom:node:current")
	vo.Nav, _ = cs.HGetAll("prom:service:gf_nav:current")
	vo.Game, _ = cs.HGetAll("prom:service:gf_game:current")

	vo.NavPath = map[string]map[string]string{
		"avg_response_1h": getPathMetric("gf_nav", "avg_response_1h"),
	}

	vo.GamePath = map[string]map[string]string{
		"avg_response_1h": getPathMetric("gf_game", "avg_response_1h"),
	}

	return
}

func getPathMetric(svc, metric string) map[string]string {
	key := fmt.Sprintf("prom:service:%s:path:%s", svc, metric)
	data, _ := cs.HGetAll(key)
	return data
}

func (s metricsService) GetPromMetricsHistory() (vo models.PromMetricsHistoryVo, err common.GFError) {
	metricKeys := map[string]string{
		"cpu":     "prom:node:history:cpu_usage",
		"connect": "prom:node:history:tcp_connections",
		"mem":     "prom:node:history:mem_usage",
	}

	// CPU 时序数据
	vo.CPU.TwentyMinutes = getMetricsByTimeRange(
		metricKeys["cpu"],
		20*time.Minute, // 前20分钟
		20,             // 20个原始数据
	)
	vo.CPU.OneHour = getMetricsByTimeRange(
		metricKeys["cpu"],
		1*time.Hour, // 前1小时
		20,          // 60个抽样为20个
	)
	vo.CPU.TwentyHours = getMetricsByTimeRange(
		metricKeys["cpu"],
		20*time.Hour, // 前20小时
		20,           // 1200个抽样为20个
	)

	// Connect 时序数据
	vo.Connect.TwentyMinutes = getMetricsByTimeRange(
		metricKeys["connect"],
		20*time.Minute,
		20,
	)
	vo.Connect.OneHour = getMetricsByTimeRange(
		metricKeys["connect"],
		1*time.Hour,
		20,
	)
	vo.Connect.TwentyHours = getMetricsByTimeRange(
		metricKeys["connect"],
		20*time.Hour,
		20,
	)

	// Memory 时序数据
	vo.Memory.TwentyMinutes = getMetricsByTimeRange(
		metricKeys["mem"],
		20*time.Minute,
		20,
	)
	vo.Memory.OneHour = getMetricsByTimeRange(
		metricKeys["mem"],
		1*time.Hour,
		20,
	)
	vo.Memory.TwentyHours = getMetricsByTimeRange(
		metricKeys["mem"],
		20*time.Hour,
		20,
	)

	return vo, nil
}

func (s metricsService) GetImageUrl() string {
	rand.Seed(time.Now().UnixNano())
	return "https://qcdn.go-furry.com/nav/stat-bg/" + util.Int2String(rand.Intn(env.GetServerConfig().Resource.StatImageNum)+1) + ".jpg"
}

func getRedisZSetData(key string, startTS, endTS int64) []models.MetricsModel {
	zRangeBy := &redis.ZRangeBy{
		Min:   strconv.FormatInt(startTS, 10),
		Max:   strconv.FormatInt(endTS, 10),
		Count: 0,
	}

	rawData, err := cs.GetRedisService().ZRevRangeByScoreWithScores(context.Background(), key, zRangeBy).Result()
	if err != nil {
		log.Error("读取 Redis ZSet 失败, key=%s, err=%v", key, err)
		return nil
	}

	// 转换为 MetricsModel 并反转
	var metrics []models.MetricsModel
	for i := len(rawData) - 1; i >= 0; i-- {
		item := rawData[i]
		// 解析时间戳
		ts := int64(item.Score)
		// 解析指标值
		val, parseErr := strconv.ParseFloat(item.Member.(string), 64)
		if parseErr != nil {
			log.Warn("解析指标值失败, key=%s, member=%s, err=%v", key, item.Member, parseErr)
			continue
		}
		metrics = append(metrics, models.MetricsModel{
			Time:  ts,
			Usage: val,
		})
	}

	return metrics
}

func sampleMetrics(rawData []models.MetricsModel, targetCount int) []models.MetricsModel {
	// 原始数据量 ≤ 目标数量
	if len(rawData) <= targetCount || targetCount <= 0 {
		return rawData
	}

	// 计算抽样步长
	step := len(rawData) / targetCount
	if len(rawData)%targetCount != 0 {
		step += 1
	}

	// 等步长抽样
	var sampled []models.MetricsModel
	for i := 0; i < len(rawData); i += step {
		sampled = append(sampled, rawData[i])
		// 避免超过目标数量
		if len(sampled) >= targetCount {
			break
		}
	}

	// 若抽样结果不足 targetCount 则补充最后一条数据
	if len(sampled) < targetCount && len(rawData) > 0 {
		sampled = append(sampled, rawData[len(rawData)-1])
	}

	// 避免超出
	if len(sampled) > targetCount {
		sampled = sampled[:targetCount]
	}

	return sampled
}

func getMetricsByTimeRange(key string, duration time.Duration, targetCount int) []models.MetricsModel {
	// 计算时间范围
	endTS := time.Now().Unix()
	startTS := time.Now().Add(-duration).Unix()

	// 读取 Redis ZSet 数据
	rawData := getRedisZSetData(key, startTS, endTS)

	// 抽样处理
	return sampleMetrics(rawData, targetCount)
}
