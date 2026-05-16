package task

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/gofurry/gofurry-nav-backend/common/log"
	cs "github.com/gofurry/gofurry-nav-backend/common/service"
	"github.com/gofurry/gofurry-nav-backend/roof/env"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

/* ================= Prometheus Client ================= */
var (
	promClient v1.API
	once       sync.Once
)

// initPromClient 初始化 Prometheus Client 单例
func initPromClient() v1.API {
	once.Do(func() {
		promURL := env.GetServerConfig().Prometheus.Url
		client, err := api.NewClient(api.Config{Address: promURL})
		if err != nil {
			log.Error("[initPromClient] create prom client err:", err)
			return // 初始化失败时，promClient 保持 nil, 后续查询会跳过
		}
		promClient = v1.NewAPI(client)
		log.Info("[initPromClient] prometheus client init success, url: %s", promURL)
	})
	return promClient
}

// getPromClient 获取全局 Prometheus Client 单例
func getPromClient() v1.API {
	if promClient == nil {
		return initPromClient()
	}
	return promClient
}

/* ================= Entry ================= */

func UpdateMetricsCache() {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	log.Debug("[UpdateMetricsCache] start metrics cache update")

	prom := getPromClient()
	if prom == nil {
		log.Error("[UpdateMetricsCache] prometheus client is nil, skip metrics update")
		return
	}

	updateNodeMetrics(ctx, prom)
	updateServiceMetrics(ctx, prom)
	updateServicePerPathMetrics(ctx, prom)

	log.Debugf("[UpdateMetricsCache] metrics cache update finished, cost: %v", time.Since(start))
}

/* ================= Node Metrics ================= */

func updateNodeMetrics(ctx context.Context, prom v1.API) {
	metrics := map[string]struct {
		Query     string
		IsHistory bool
	}{
		"cpu_usage":       {`100 - avg(irate(node_cpu_seconds_total{job="servers",mode="idle"}[5m])) * 100`, true},
		"mem_usage":       {`sum(node_memory_MemTotal_bytes{job="servers"}) - sum(node_memory_MemAvailable_bytes{job="servers"})`, true},
		"disk_usage":      {`100 * (sum(node_filesystem_size_bytes{job="servers",fstype!~"tmpfs|overlay"}) - sum(node_filesystem_avail_bytes{job="servers",fstype!~"tmpfs|overlay"})) / sum(node_filesystem_size_bytes{job="servers",fstype!~"tmpfs|overlay"})`, false},
		"net_rx_1d":       {`sum(increase(node_network_receive_bytes_total{job="servers", device!~"lo|docker0"}[1d]))`, false},
		"net_tx_1d":       {`sum(increase(node_network_transmit_bytes_total{job="servers", device!~"lo|docker0"}[1d]))`, false},
		"tcp_connections": {`sum(node_netstat_Tcp_CurrEstab{job="servers"}) or vector(0)`, true},
		"uptime":          {`avg(time() - node_boot_time_seconds{job="servers"})`, false},
	}

	nodeCurrentFields := make(map[string]string)
	for key, cfg := range metrics {
		val, ok := queryPromAgg(ctx, prom, cfg.Query)
		if !ok {
			continue
		}
		// 收集到 map
		nodeCurrentFields[key] = fmt.Sprintf("%.4f", val)

		if cfg.IsHistory {
			cacheHistory(ctx, "prom:node:history:"+key, val, 7)
		}
	}

	// 批量 HSet 写入 Redis
	if len(nodeCurrentFields) > 0 {
		err := cs.GetRedisService().HSet(ctx, "prom:node:current", nodeCurrentFields).Err()
		if err != nil {
			log.Error("[cacheNodeCurrentBatch] err: %v", err)
		}
	}
}

/* ================= Service Metrics ================= */

func updateServiceMetrics(ctx context.Context, prom v1.API) {
	services := map[string]map[string]string{
		"gf_nav": {
			"http_requests_1d": `sum(increase(gf_nav_http_requests_total[1d]))`,
			"http_requests_7d": `sum(increase(gf_nav_http_requests_total[7d]))`,
			"avg_response_1h":  `sum(rate(gf_nav_http_request_duration_seconds_sum[1h])) / sum(rate(gf_nav_http_request_duration_seconds_count[1h])) or vector(0)`,
			"p99_response_1h":  `histogram_quantile(0.99, sum by(le) (rate(gf_nav_http_request_duration_seconds_bucket[1h])))`,
			"p95_response_1h":  `histogram_quantile(0.95, sum by(le) (rate(gf_nav_http_request_duration_seconds_bucket[1h])))`,
			"fail_rate_1h":     `sum(rate(gf_nav_http_requests_total{status!~"2.."}[1h])) / sum(rate(gf_nav_http_requests_total[1h])) or vector(0)`,
		},
		"gf_game": {
			"http_requests_1d": `sum(increase(gf_game_http_requests_total[1d]))`,
			"http_requests_7d": `sum(increase(gf_game_http_requests_total[7d]))`,
			"avg_response_1h":  `sum(rate(gf_game_http_request_duration_seconds_sum[1h])) / sum(rate(gf_game_http_request_duration_seconds_count[1h])) or vector(0)`,
			"p99_response_1h":  `histogram_quantile(0.99, sum by(le) (rate(gf_game_http_request_duration_seconds_bucket[1h])))`,
			"p95_response_1h":  `histogram_quantile(0.95, sum by(le) (rate(gf_game_http_request_duration_seconds_bucket[1h])))`,
			"fail_rate_1h":     `sum(rate(gf_game_http_requests_total{status!~"2.."}[1h])) / sum(rate(gf_game_http_requests_total[1h])) or vector(0)`,
		},
	}

	for svc, m := range services {
		// 批量收集当前服务的所有指标字段
		svcCurrentFields := make(map[string]string)
		for key, query := range m {
			val, ok := queryPromAgg(ctx, prom, query)
			if !ok {
				continue
			}
			svcCurrentFields[key] = fmt.Sprintf("%.4f", val)
		}

		// 批量 HSet 写入 Redis
		if len(svcCurrentFields) > 0 {
			redisKey := "prom:service:" + svc + ":current"
			err := cs.GetRedisService().HSet(ctx, redisKey, svcCurrentFields).Err()
			if err != nil {
				log.Error("[cacheServiceCurrentBatch] svc=%s err=%v", svc, err)
			}
		}
	}
}

/* ================= Per Path Metrics ================= */

func updateServicePerPathMetrics(ctx context.Context, prom v1.API) {
	queries := map[string]string{
		"http_requests_1d": `sum by(path) (increase(%s_http_requests_total[1d]))`,
		"http_requests_7d": `sum by(path) (increase(%s_http_requests_total[7d]))`,
		"avg_response_1h":  `sum by(path) (rate(%s_http_request_duration_seconds_sum[1h])) / sum by(path) (rate(%s_http_request_duration_seconds_count[1h]))`,
	}

	for _, svc := range []string{"gf_nav", "gf_game"} {
		for metric, q := range queries {
			query := fmt.Sprintf(q, svc, svc)
			result, ok := queryPromVector(ctx, prom, query)
			if !ok {
				continue
			}
			for path, val := range result {
				cacheServicePathMetric(ctx, svc, metric, path, val)
			}
		}
	}
}

/* ================= Prom Query ================= */

func queryPromAgg(ctx context.Context, api v1.API, query string) (float64, bool) {
	result, _, err := api.Query(ctx, query, time.Now())
	if err != nil {
		return 0, false
	}

	switch v := result.(type) {
	case model.Vector:
		var sum float64
		for _, s := range v {
			sum += float64(s.Value)
		}
		return sum, !math.IsNaN(sum)
	case *model.Scalar:
		return float64(v.Value), true
	}
	return 0, false
}

func queryPromVector(ctx context.Context, api v1.API, query string) (map[string]float64, bool) {
	result, _, err := api.Query(ctx, query, time.Now())
	if err != nil {
		return nil, false
	}

	vec, ok := result.(model.Vector)
	if !ok {
		return nil, false
	}

	out := make(map[string]float64)
	for _, s := range vec {
		out[string(s.Metric["path"])] = float64(s.Value)
	}
	return out, true
}

/* ================= internal util ================= */

func cacheServicePathMetric(ctx context.Context, svc, metric, path string, val float64) {
	key := fmt.Sprintf("prom:service:%s:path:%s", svc, metric)
	cs.GetRedisService().HSet(ctx, key, path, fmt.Sprintf("%.4f", val))
	cs.GetRedisService().Expire(ctx, key, 10*time.Minute)
}

func cacheHistory(ctx context.Context, key string, val float64, days int) {
	ts := time.Now().Unix()
	value := fmt.Sprintf("%.4f", val)

	// ZADD key score member
	if _, err := cs.GetRedisService().
		Do(ctx, "ZADD", key, ts, value).Result(); err != nil {

		log.Error("[cacheHistory RedisZADD] key=%s err=%v", key, err)
		return
	}

	// 只保留最近 N 天
	expireTS := time.Now().Add(-time.Duration(days*24) * time.Hour).Unix()
	cs.GetRedisService().Do(ctx, "ZREMRANGEBYSCORE", key, 0, expireTS)
	cs.GetRedisService().Do(ctx, "EXPIRE", key, days*24*3600)
}
