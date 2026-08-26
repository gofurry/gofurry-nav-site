package controller

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/gofurry/gofurry-game-collector/collector/game/service"
	"github.com/gofurry/gofurry-game-collector/common/log"
	cs "github.com/gofurry/gofurry-game-collector/common/service"
	"github.com/gofurry/gofurry-game-collector/roof/env"
)

type gameApi struct{}

var GameApi *gameApi

func init() {
	GameApi = &gameApi{}
}

var collectFlag atomic.Bool

// 初始化 Game 采集模块
func (api *gameApi) InitGameCollection() {
	defer func() {
		if err := recover(); err != nil {
			log.Error("receive InitGameCollection recover: ", err)
		}
	}()
	fmt.Println("Game 模块初始化开始...")

	// 初始化限流器
	service.InitLimiter()

	// Pending single-game work has startup priority. Only collect all players on
	// startup when no pending request exists and configuration allows it.
	go runStartupCollection(
		SchedulePendingGameCollection,
		env.GetServerConfig().Collector.Game.PlayersOnStartupEnabled(),
		service.GetGameService().CollectCurrentPlayers,
	)

	// 指令感知(1分钟)
	cs.AddCronJob(1*time.Minute, ScheduleByOneMin)

	// 定时任务执行 Ping
	cs.AddCronJob(time.Duration(env.GetServerConfig().Collector.Game.GamePlayerInterval)*time.Hour, service.GetGameService().CollectCurrentPlayers)

	fmt.Println("Game 模块初始化结束...")
}

// 1分钟任务表
func ScheduleByOneMin() {
	// A running full or pending collection owns the shared collector slot. When
	// idle, pending single-game work always runs before daily/manual full work.
	if SchedulePendingGameCollection() {
		return
	}

	// 转换为北京时间
	now := time.Now().In(time.FixedZone("CST", 8*3600))
	year, month, day := now.Date()

	// 构建今日采集状态键
	todayKey := now.Format("20060102")
	statusKey := "game-collector:collect-" + todayKey
	cmdKey := "game-collector:cmd:collect"

	// 获取今日采集状态
	todayCollected, err := cs.GetString(statusKey)
	if err != nil {
		log.Error("获取今日采集状态失败: ", err)
		return // 状态获取失败时直接返回
	}

	// 获取手动采集指令
	manualCmd, err := cs.GetString(cmdKey)
	if err != nil {
		log.Error("获取手动采集指令失败: ", err)
		manualCmd = "" // 异常时置空
	}

	// 检查当前是否正在采集
	isCollecting := collectFlag.Load()

	// 构建北京时间今日凌晨3点的时间戳
	threeAM := time.Date(year, month, day, 3, 0, 0, 0, now.Location())
	// 未在采集 && (今日未采集且已过凌晨3点 || 收到手动采集指令)
	needCollect := !isCollecting &&
		((todayCollected != "1" && now.After(threeAM)) || manualCmd == "1")

	if !needCollect {
		return // 不满足采集条件
	}

	// Close the window in which Admin may enqueue after the first scan but
	// before the daily collection acquires the shared slot.
	if SchedulePendingGameCollection() {
		return
	}

	// 标记开始采集
	if !collectFlag.CompareAndSwap(false, true) {
		return
	}
	// 定义结束时的清理逻辑
	defer func() {
		// 无论成功失败都重置采集状态
		collectFlag.Store(false)

		// 如果是手动指令触发执行后清理指令
		if manualCmd == "1" {
			if delErr := cs.Del(cmdKey); delErr != nil {
				log.Error("清理手动采集指令失败: ", delErr)
			}
		}

		cs.SetExpire(statusKey, "1", 7*24*time.Hour)

		// 捕获 panic
		if r := recover(); r != nil {
			log.Error("采集任务执行异常: ", r)
		}
	}()

	// 执行采集任务
	log.Info("开始执行游戏数据采集任务")
	service.GetGameService().Collect()
}

func runStartupCollection(processPending func() bool, collectPlayers bool, collectPlayersNow func()) {
	if processPending != nil && processPending() {
		return
	}
	if collectPlayers && collectPlayersNow != nil {
		collectPlayersNow()
	}
}
