package service

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bytedance/sonic"
	"github.com/go-ping/ping"
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
	}, "Ping collector module initialization started")

	//初始化后执行一次 Ping
	go Ping()
	go Delete()
	// 定时任务执行 Ping
	cs.AddCronJob(time.Duration(env.GetServerConfig().Collector.Ping.PingInterval)*time.Second, Ping)
	cs.AddCronJob(24*time.Hour, Delete)

	log.InfoFields(map[string]interface{}{
		"event":    "module_init_complete",
		"protocol": "ping",
	}, "Ping collector module initialization completed")
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
	}, "Ping retention cleanup started")

	// 每个域名仅保留 5000 条 ping 记录
	count, deleteErr := dao.GetPingDao().DeleteByNum(keepCount)
	if deleteErr != nil {
		log.ErrorFields(map[string]interface{}{
			"deleted":    count,
			"duration":   time.Since(start),
			"event":      "retention_failed",
			"keep_count": keepCount,
			"protocol":   "ping",
		}, "Ping retention cleanup failed: "+deleteErr.GetMsg())
	} else {
		log.InfoFields(map[string]interface{}{
			"deleted":    count,
			"duration":   time.Since(start),
			"event":      "retention_complete",
			"keep_count": keepCount,
			"protocol":   "ping",
		}, "Ping retention cleanup completed")
	}
}

// 添加数据库全部 IP 到 redis
func addAllIpToPing() common.GFError {
	// 查记录
	domainRecords, err := dao.GetPingDao().GetList()
	if err != nil {
		log.Error(fmt.Sprintf("查询IP失败: %v", err.GetMsg()))
		return common.NewServiceError(fmt.Sprintf("查询IP失败: %v", err))
	}

	// 添加 ping 的站点
	var pingList = []string{}
	for _, v := range domainRecords {
		newDomains := models2.Domains{}
		if jsonErr := sonic.Unmarshal([]byte(v.Domain), &newDomains); jsonErr != nil {
			log.Error(fmt.Sprintf("json转换失败: %v", jsonErr))
			return nil
		}
		for _, domain := range newDomains.Domain {
			pingList = append(pingList, domain)
		}
	}

	// 存入 redis
	pingJsonList, jsonErr := sonic.Marshal(pingList)
	if jsonErr != nil {
		log.Error(fmt.Sprintf("json转换失败: %v", jsonErr))
		return nil
	}

	err = cs.Del(env.GetServerConfig().Collector.Ping.PingKey)
	if err != nil {
		log.Error("删除ping结果失败: ", err)
		return err
	}

	cs.SetNX(env.GetServerConfig().Collector.Ping.PingKey, pingJsonList, 24*time.Hour)

	return nil
}

// ============== Ping解析 - 执行部分 ==============

// 检测是否在线
func Ping() {
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("receive Ping recover: %v", err))
		}
	}()
	if !pingRunning.CompareAndSwap(false, true) {
		log.WarnFields(map[string]interface{}{
			"event":    "run_skipped",
			"protocol": "ping",
			"reason":   "previous_run_running",
			"status":   "skipped",
		}, "Ping collection skipped because the previous run is still running")
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
	}, "Ping collection run started")

	// 查询数据库所有 IP 存 redis 每次采集都请求记录 热更新
	err := addAllIpToPing()
	if err != nil {
		log.ErrorFields(map[string]interface{}{
			"duration": time.Since(start),
			"event":    "run_failed",
			"protocol": "ping",
			"stage":    "load_targets_to_redis",
		}, "Ping collection run failed: "+err.GetMsg())
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
		}, "Ping target list load failed: "+err.GetMsg())
		return
	}
	// 判空
	if domains == "" || len(domains) < 1 {
		log.InfoFields(map[string]interface{}{
			"duration": time.Since(start),
			"event":    "run_complete",
			"protocol": "ping",
			"reason":   "empty_target_list",
			"targets":  0,
		}, "Ping collection completed with no targets")
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
		}, "Ping previous result load failed: "+err.GetMsg())
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
		}, "Ping target list JSON decode failed: "+jsonErr.Error())
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
	}, "Ping probes started")
	pingThread := pool.New().WithMaxGoroutines(env.GetServerConfig().Collector.Ping.PingThread)
	var pingRWLock sync.Mutex
	// 遍历 IP 列表, 每个 IP 开一个线程执行 Ping
	for _, v := range pingList {
		pingThread.Go(getPingResult(v, nowData, &pingRWLock))
	}
	// 等待所有 Ping 执行完毕
	pingThread.Wait()
	log.InfoFields(map[string]interface{}{
		"duration": time.Since(start),
		"event":    "probe_complete",
		"protocol": "ping",
		"targets":  len(pingList),
	}, "Ping probes completed")

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
			}, "Ping stale Redis results cleanup failed: "+delErr.GetMsg())
		}
		if count > 0 {
			log.InfoFields(map[string]interface{}{
				"deleted":   count,
				"event":     "redis_cleanup_complete",
				"protocol":  "ping",
				"redis_key": resultKey,
			}, "Ping stale Redis results cleanup completed")
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
		}, "Ping result save failed: "+err.GetMsg())
		return
	}
	log.InfoFields(map[string]interface{}{
		"duration":  time.Since(start),
		"event":     "run_complete",
		"protocol":  "ping",
		"redis_key": resultKey,
		"targets":   len(nowData),
	}, "Ping collection run completed")
}

// ============== Ping解析 - 采集和解析部分 ==============

// 执行 ping 采集
func performPing(ip string) models2.PingModel {
	pinger, err := ping.NewPinger(ip)
	// 初始化结果字段
	var pingModel models2.PingModel
	pingModel.PingTime = cm.LocalTime(time.Now())
	if err != nil {
		return pingModel
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
		return pingModel
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
func getPingResult(ip string, data map[string]string, pingRWLock *sync.Mutex) func() {
	return func() {
		defer func() {
			if err := recover(); err != nil {
				log.Error(fmt.Sprintf("receive PingThread recover: %v", err))
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

		// 开启读写锁
		pingRWLock.Lock()
		defer pingRWLock.Unlock()
		// 更新字典
		data[ip] = string(jsonResult)

		// 存数据库
		dao.GetPingDao().Add(pindSaveRecord)
	}
}
