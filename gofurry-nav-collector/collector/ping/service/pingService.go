package service

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bytedance/sonic"
	"github.com/go-ping/ping"
	"github.com/gofurry/gofurry-nav-collector/collector/observation"
	"github.com/gofurry/gofurry-nav-collector/collector/ping/dao"
	models2 "github.com/gofurry/gofurry-nav-collector/collector/ping/models"
	"github.com/gofurry/gofurry-nav-collector/common"
	"github.com/gofurry/gofurry-nav-collector/common/log"
	cm "github.com/gofurry/gofurry-nav-collector/common/models"
	cs "github.com/gofurry/gofurry-nav-collector/common/service"
	cu "github.com/gofurry/gofurry-nav-collector/common/util"
	"github.com/gofurry/gofurry-nav-collector/roof/env"
	"github.com/sourcegraph/conc/pool"
)

var pingRunning atomic.Bool

// ============== Ping模块 - 初始化部分 ==============

// 初始化
func InitPingOnStart() {
	defer func() {
		if err := recover(); err != nil {
			log.ErrorFields(map[string]interface{}{
				"event":    "init_recovered",
				"protocol": "ping",
			}, err)
		}
	}()
	log.InfoFields(map[string]interface{}{
		"event":           "module_init_start",
		"interval":        time.Duration(env.GetServerConfig().Collector.Ping.PingInterval) * time.Second,
		"protocol":        "ping",
		"retention_every": time.Hour * 24,
		"workers":         env.GetServerConfig().Collector.Ping.PingThread,
	}, "Ping 采集模块初始化开始")

	//初始化后执行一次 Ping
	go Ping()
	go Delete()
	// 定时任务执行 Ping
	cs.AddCronJob(time.Duration(env.GetServerConfig().Collector.Ping.PingInterval)*time.Second, Ping)
	cs.AddCronJob(24*time.Hour, Delete)

	log.InfoFields(map[string]interface{}{
		"event":    "module_init_complete",
		"protocol": "ping",
	}, "Ping 采集模块初始化完成")
}

// 每天清理一次日志表
func Delete() {
	defer func() {
		if err := recover(); err != nil {
			log.ErrorFields(map[string]interface{}{
				"event":    "retention_recovered",
				"protocol": "ping",
			}, err)
		}
	}()

	start := time.Now()
	keepCount := env.GetServerConfig().Collector.Ping.LogCount
	log.InfoFields(map[string]interface{}{
		"event":      "retention_start",
		"keep_count": keepCount,
		"protocol":   "ping",
	}, "Ping 历史日志保留清理开始")

	// 每个域名仅保留 5000 条 ping 记录
	count, deleteErr := dao.GetPingDao().DeleteByNum(keepCount)
	if deleteErr != nil {
		log.ErrorFields(map[string]interface{}{
			"deleted":    count,
			"duration":   time.Since(start),
			"event":      "retention_failed",
			"keep_count": keepCount,
			"protocol":   "ping",
		}, "Ping 历史日志保留清理失败: "+deleteErr.GetMsg())
	} else {
		log.InfoFields(map[string]interface{}{
			"deleted":    count,
			"duration":   time.Since(start),
			"event":      "retention_complete",
			"keep_count": keepCount,
			"protocol":   "ping",
		}, "Ping 历史日志保留清理完成")
	}
	if env.GetServerConfig().Collector.V2.ProtocolEnabled(observation.ProtocolPing) {
		v2Count, v2DeleteErr := observation.DeleteByProtocolLimit(observation.ProtocolPing, keepCount)
		if v2DeleteErr != nil {
			log.ErrorFields(map[string]interface{}{
				"deleted":    v2Count,
				"duration":   time.Since(start),
				"event":      "v2_retention_failed",
				"keep_count": keepCount,
				"protocol":   "ping",
			}, "Ping v2 observation 保留清理失败: "+v2DeleteErr.GetMsg())
		} else if v2Count > 0 {
			log.InfoFields(map[string]interface{}{
				"deleted":    v2Count,
				"duration":   time.Since(start),
				"event":      "v2_retention_complete",
				"keep_count": keepCount,
				"protocol":   "ping",
			}, "Ping v2 observation 保留清理完成")
		}
	}
}

// 添加数据库全部采集域名到 redis
func addAllIpToPing() (map[string]int64, common.GFError) {
	// 查记录
	domainRecords, err := dao.GetPingDao().GetList()
	if err != nil {
		log.Error(fmt.Sprintf("查询 Ping 目标失败: %v", err.GetMsg()))
		return nil, common.NewServiceError(fmt.Sprintf("查询 Ping 目标失败: %v", err))
	}

	// 添加 ping 的站点
	pingList, siteIDByDomain := buildPingTargets(domainRecords)

	// 存入 redis
	pingJsonList, jsonErr := sonic.Marshal(pingList)
	if jsonErr != nil {
		log.Error(fmt.Sprintf("json转换失败: %v", jsonErr))
		return siteIDByDomain, nil
	}

	err = cs.Del(env.GetServerConfig().Collector.Ping.PingKey)
	if err != nil {
		log.Error("删除ping结果失败: ", err)
		return siteIDByDomain, err
	}

	cs.SetNX(env.GetServerConfig().Collector.Ping.PingKey, pingJsonList, 24*time.Hour)

	return siteIDByDomain, nil
}

func buildPingTargets(domainRecords []models2.GfnCollectorDomain) ([]string, map[string]int64) {
	pingList := []string{}
	siteIDByDomain := map[string]int64{}
	for _, v := range domainRecords {
		domain := collectorDomainTarget(v)
		if domain == "" || v.SiteID <= 0 {
			continue
		}
		pingList = append(pingList, domain)
		siteIDByDomain[domain] = v.SiteID
	}
	return pingList, siteIDByDomain
}

func collectorDomainTarget(record models2.GfnCollectorDomain) string {
	if record.Prefix == nil {
		return record.Name
	}
	return *record.Prefix + record.Name
}

// ============== Ping解析 - 执行部分 ==============

// 检测是否在线
func Ping() {
	defer func() {
		if err := recover(); err != nil {
			log.ErrorFields(map[string]interface{}{
				"event":    "run_recovered",
				"protocol": "ping",
			}, fmt.Sprintf("Ping 采集运行触发 panic，已恢复: %v", err))
		}
	}()
	if !pingRunning.CompareAndSwap(false, true) {
		log.WarnFields(map[string]interface{}{
			"event":    "run_skipped",
			"protocol": "ping",
			"reason":   "上一轮采集仍在运行",
			"status":   "skipped",
		}, "Ping 采集已跳过：上一轮仍在运行")
		return
	}
	defer pingRunning.Store(false)

	start := time.Now()
	log.InfoFields(map[string]interface{}{
		"event":      "run_start",
		"ping_key":   env.GetServerConfig().Collector.Ping.PingKey,
		"protocol":   "ping",
		"result_key": env.GetServerConfig().Collector.Ping.ResultKey,
		"workers":    env.GetServerConfig().Collector.Ping.PingThread,
	}, "Ping 采集运行开始")

	// 查询数据库所有 IP 存 redis 每次采集都请求记录 热更新
	siteIDByDomain, err := addAllIpToPing()
	if err != nil {
		log.ErrorFields(map[string]interface{}{
			"duration": time.Since(start),
			"event":    "run_failed",
			"protocol": "ping",
			"stage":    "load_targets_to_redis",
		}, "Ping 采集运行失败: "+err.GetMsg())
		return
	}

	// redis 中获取 ping 的站点列表
	var pingKey = env.GetServerConfig().Collector.Ping.PingKey
	domains, err := cs.GetString(pingKey)
	if err != nil {
		log.ErrorFields(map[string]interface{}{
			"duration":  time.Since(start),
			"event":     "run_failed",
			"protocol":  "ping",
			"redis_key": pingKey,
			"stage":     "load_targets_from_redis",
		}, "Ping 目标列表读取失败: "+err.GetMsg())
		return
	}
	// 判空
	if domains == "" || len(domains) < 1 {
		log.InfoFields(map[string]interface{}{
			"duration": time.Since(start),
			"event":    "run_complete",
			"protocol": "ping",
			"reason":   "目标列表为空",
			"targets":  0,
		}, "Ping 采集完成：没有需要探测的目标")
		return
	}

	// redis 中获取旧记录
	var resultKey = env.GetServerConfig().Collector.Ping.ResultKey
	data, err := cs.HGetAll(resultKey)
	if err != nil {
		log.ErrorFields(map[string]interface{}{
			"duration":  time.Since(start),
			"event":     "run_failed",
			"protocol":  "ping",
			"redis_key": resultKey,
			"stage":     "load_previous_results",
		}, "Ping 历史结果读取失败: "+err.GetMsg())
		return
	}
	// 判空
	if data == nil || len(data) < 1 {
		data = map[string]string{}
	}

	var pingList = []string{}
	if jsonErr := sonic.Unmarshal([]byte(domains), &pingList); jsonErr != nil {
		log.ErrorFields(map[string]interface{}{
			"duration": time.Since(start),
			"event":    "run_failed",
			"protocol": "ping",
			"stage":    "decode_target_list",
		}, "Ping 目标列表 JSON 解析失败: "+jsonErr.Error())
		return
	}

	// 复制旧记录中在站点列表中的部分到新纪录
	var nowData = map[string]string{}
	for _, domain := range pingList {
		nowData[domain] = data[domain]
	}

	log.InfoFields(map[string]interface{}{
		"event":            "probe_start",
		"previous_results": len(data),
		"protocol":         "ping",
		"targets":          len(pingList),
		"workers":          env.GetServerConfig().Collector.Ping.PingThread,
	}, "Ping 探测开始")
	pingThread := pool.New().WithMaxGoroutines(env.GetServerConfig().Collector.Ping.PingThread)
	var pingRWLock sync.Mutex
	// 遍历 IP 列表, 每个 IP 开一个线程执行 Ping
	for _, v := range pingList {
		target := models2.PingTarget{SiteID: siteIDByDomain[v], Domain: v}
		pingThread.Go(getPingResult(target, nowData, &pingRWLock))
	}
	// 等待所有 Ping 执行完毕
	pingThread.Wait()
	log.InfoFields(map[string]interface{}{
		"duration": time.Since(start),
		"event":    "probe_complete",
		"protocol": "ping",
		"targets":  len(pingList),
	}, "Ping 探测完成")

	// 删除旧记录中不在站点列表中的部分
	var deleteList = []string{}
	for k, _ := range data {
		if nowData[k] == "" {
			deleteList = append(deleteList, k)
		}
	}
	if len(deleteList) != 0 {
		count, delErr := cs.HDel(resultKey, deleteList...)
		if delErr != nil {
			log.ErrorFields(map[string]interface{}{
				"event":     "redis_cleanup_failed",
				"protocol":  "ping",
				"redis_key": resultKey,
				"targets":   len(deleteList),
			}, "Ping 过期 Redis 结果清理失败: "+delErr.GetMsg())
		}
		if count > 0 {
			log.InfoFields(map[string]interface{}{
				"deleted":   count,
				"event":     "redis_cleanup_complete",
				"protocol":  "ping",
				"redis_key": resultKey,
			}, "Ping 过期 Redis 结果清理完成")
		}
	}
	// Ping 结果储存回 redis
	err = cs.HSetMap(resultKey, nowData)
	if err != nil {
		log.ErrorFields(map[string]interface{}{
			"duration":  time.Since(start),
			"event":     "run_failed",
			"protocol":  "ping",
			"redis_key": resultKey,
			"stage":     "save_results",
			"targets":   len(nowData),
		}, "Ping 结果保存失败: "+err.GetMsg())
		return
	}
	log.InfoFields(map[string]interface{}{
		"duration":  time.Since(start),
		"event":     "run_complete",
		"protocol":  "ping",
		"redis_key": resultKey,
		"targets":   len(nowData),
	}, "Ping 采集运行完成")
}

// ============== Ping解析 - 采集和解析部分 ==============

// 执行 ping 采集
func performPing(ip string) (pingModel models2.PingModel) {
	start := time.Now()
	pingModel.PingTime = cm.LocalTime(time.Now())
	defer func() {
		pingModel.ProbeDurationMS = time.Since(start).Milliseconds()
	}()

	pinger, err := ping.NewPinger(ip)
	if err != nil {
		pingModel.ErrorCode = "ping_init_failed"
		pingModel.ErrorMessage = err.Error()
		return
	}
	defer pinger.Stop()
	pingModel.AvgLossRate = 100
	pingModel.AvgDelayTime = 100000000
	// 初始化 pinger
	pinger.Count = 5
	pinger.Size = 64
	pinger.Interval = time.Second
	pinger.Timeout = time.Second * 5
	pinger.SetPrivileged(true)
	// 运行 Pinger
	err = pinger.Run()
	if err != nil {
		pingModel.ErrorCode = "ping_run_failed"
		pingModel.ErrorMessage = err.Error()
		return
	}
	// 转换数据
	stats := pinger.Statistics()
	pingModel.AvgLossRate = stats.PacketLoss
	pingModel.AvgDelayTime = stats.AvgRtt.Milliseconds()
	if pingModel.AvgDelayTime == 0 && pingModel.AvgLossRate != 100 {
		pingModel.AvgDelayTime = 1
	}
	return pingModel
}

// 解析 ping 采集结果
func getPingResult(target models2.PingTarget, data map[string]string, pingRWLock *sync.Mutex) func() {
	return func() {
		ip := target.Domain
		defer func() {
			if err := recover(); err != nil {
				log.ErrorFields(map[string]interface{}{
					"event":    "probe_recovered",
					"protocol": "ping",
					"target":   ip,
				}, fmt.Sprintf("Ping 单目标探测触发 panic，已恢复: %v", err))
			}
		}()

		// 执行 Ping 获取结果
		result := performPing(ip)

		pingRecord := &models2.PingSaveModel{}
		pingRecord.Time = result.PingTime
		pingRecord.Delay = cu.Int642String(result.AvgDelayTime) + "ms"
		pingRecord.Loss = cu.Float642String(result.AvgLossRate)
		if result.AvgLossRate < 99 && result.AvgDelayTime > 0 {
			pingRecord.Status = "up"
		} else {
			pingRecord.Status = "down"
		}
		// 序列化为 json
		jsonResult, _ := sonic.Marshal(pingRecord)

		// 存数据库
		pindSaveRecord := &models2.GfnCollectorLogPing{
			ID:         cu.GenerateId(),
			Name:       ip,
			Delay:      cu.Int642String(result.AvgDelayTime) + "ms",
			Loss:       cu.Float642String(result.AvgLossRate),
			CreateTime: result.PingTime,
		}
		if result.AvgLossRate < 99 && result.AvgDelayTime > 0 {
			pindSaveRecord.Status = "up"
		} else {
			pindSaveRecord.Status = "down"
		}

		// 只在更新共享结果 map 时持有锁，外部存储写入不占用全局锁。
		pingRWLock.Lock()
		data[ip] = string(jsonResult)
		pingRWLock.Unlock()

		// 存数据库
		if daoErr := dao.GetPingDao().Add(pindSaveRecord); daoErr != nil {
			log.ErrorFields(map[string]interface{}{
				"event":    "db_write_failed",
				"protocol": "ping",
				"site_id":  target.SiteID,
				"status":   pindSaveRecord.Status,
				"target":   ip,
			}, "Ping 探测结果写入数据库失败: "+daoErr.GetMsg())
		}
		payload, errorCode := buildPingObservationPayload(result, pingRecord, pindSaveRecord.Status)
		saveErr := observation.SaveIfEnabled(observation.Input{
			SiteID:       target.SiteID,
			Target:       ip,
			Protocol:     observation.ProtocolPing,
			Status:       observationStatusFromPing(pindSaveRecord.Status),
			ObservedAt:   time.Time(result.PingTime),
			DurationMS:   result.ProbeDurationMS,
			ErrorCode:    errorCode,
			ErrorMessage: result.ErrorMessage,
			Payload:      payload,
		})
		if saveErr != nil {
			log.WarnFields(map[string]interface{}{
				"event":    "v2_observation_write_failed",
				"protocol": "ping",
				"site_id":  target.SiteID,
				"target":   ip,
			}, "Ping v2 observation 旁路写入失败: "+saveErr.GetMsg())
		}
	}
}

func buildPingObservationPayload(result models2.PingModel, pingRecord *models2.PingSaveModel, status string) (map[string]any, string) {
	errorCode := result.ErrorCode
	icmpStatus := "unreachable"
	var avgRTTMS any
	if status == "up" {
		icmpStatus = "reachable"
		avgRTTMS = result.AvgDelayTime
	} else if errorCode == "" {
		errorCode = "ping_unreachable"
	}

	return map[string]any{
		"icmp_status":   icmpStatus,
		"avg_rtt_ms":    avgRTTMS,
		"loss_rate":     result.AvgLossRate,
		"duration_ms":   result.ProbeDurationMS,
		"error_code":    errorCode,
		"delay_ms":      result.AvgDelayTime,
		"legacy_delay":  pingRecord.Delay,
		"legacy_loss":   pingRecord.Loss,
		"legacy_status": pingRecord.Status,
	}, errorCode
}

func observationStatusFromPing(status string) string {
	if status == "up" {
		return observation.StatusSuccess
	}
	return observation.StatusFailure
}

func errorCodeFromStatus(status string, code string) string {
	if status == "up" || status == observation.StatusSuccess {
		return ""
	}
	return code
}
