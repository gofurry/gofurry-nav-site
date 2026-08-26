package controller

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofurry/gofurry-game-collector/collector/game/models"
	"github.com/gofurry/gofurry-game-collector/collector/game/service"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/report"
	"github.com/gofurry/gofurry-game-collector/common/log"
	cs "github.com/gofurry/gofurry-game-collector/common/service"
)

const (
	pendingGameCollectSetKey        = "game:v2:collect:pending"
	pendingGameCollectLockKeyPrefix = "game:v2:collect:inflight:"
	pendingGameCollectScanCount     = int64(5)
	pendingGameCollectLockTTL       = 30 * time.Minute
	gameHomeCacheZHKey              = "game:v2:home:zh:CN"
	gameHomeCacheENKey              = "game:v2:home:en:CN"
)

// SchedulePendingGameCollection returns true when lower-priority scheduled work
// must yield because a collection is active, pending work exists, or the queue
// cannot be inspected safely.
func SchedulePendingGameCollection() bool {
	if collectFlag.Load() {
		return true
	}

	client := cs.GetRedisService()
	if client == nil {
		return true
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	members, _, err := client.SScan(ctx, pendingGameCollectSetKey, 0, "", pendingGameCollectScanCount).Result()
	if err != nil {
		log.Error("扫描单游戏采集队列失败: ", err)
		return true
	}
	if len(members) == 0 {
		return false
	}
	if !collectFlag.CompareAndSwap(false, true) {
		return true
	}
	defer collectFlag.Store(false)

	for _, member := range members {
		game, ok := parsePendingGameCollectMember(member)
		if !ok {
			removePendingGameCollectMember(member)
			log.Warn("移除非法单游戏采集队列成员: ", member)
			continue
		}
		processPendingGameCollect(member, game)
	}
	return true
}

func parsePendingGameCollectMember(member string) (models.GameID, bool) {
	gameIDRaw, appIDRaw, ok := strings.Cut(strings.TrimSpace(member), ":")
	if !ok {
		return models.GameID{}, false
	}
	gameID, gameErr := strconv.ParseInt(strings.TrimSpace(gameIDRaw), 10, 64)
	appID, appErr := strconv.ParseInt(strings.TrimSpace(appIDRaw), 10, 64)
	if gameErr != nil || appErr != nil || gameID <= 0 || appID <= 0 {
		return models.GameID{}, false
	}
	return models.GameID{ID: gameID, Appid: appID}, true
}

func processPendingGameCollect(member string, game models.GameID) {
	lockKey := fmt.Sprintf("%s%d", pendingGameCollectLockKeyPrefix, game.ID)
	if !acquirePendingGameCollectLock(lockKey, member) {
		return
	}

	completed := false
	defer func() {
		if r := recover(); r != nil {
			log.Error("单游戏采集任务异常，保留队列等待重试, game_id=", game.ID, " appid=", game.Appid, " err=", r)
		}
		if completed {
			releasePendingGameCollectLock(lockKey)
			removePendingGameCollectMember(member)
		}
	}()

	log.Info("开始执行单游戏采集任务, game_id=", game.ID, " appid=", game.Appid)
	summary, err := service.GetGameService().CollectSingleGame(game)
	if err != nil {
		log.Error("单游戏采集任务执行失败, game_id=", game.ID, " appid=", game.Appid, " run_id=", summary.ID, " err=", err)
		return
	}

	completed = pendingGameDetailsCollected(summary)
	if !completed {
		log.Error("单游戏详情采集未完成，保留队列等待重试, game_id=", game.ID, " appid=", game.Appid, " run_id=", summary.ID, " status=", summary.Status)
		return
	}

	invalidateGameHomeCache()
	log.Info("单游戏采集任务执行完成, game_id=", game.ID, " appid=", game.Appid, " run_id=", summary.ID, " status=", summary.Status)
}

func pendingGameDetailsCollected(summary report.RunSummary) bool {
	for _, result := range summary.Results {
		if result.Task != domain.TaskDetails {
			continue
		}
		return result.Status == domain.StatusSuccess || result.Status == domain.StatusPartial
	}
	return false
}

func invalidateGameHomeCache() {
	if err := cs.Del(gameHomeCacheZHKey, gameHomeCacheENKey); err != nil {
		log.Error("失效游戏首页缓存失败: ", err)
	}
}

func acquirePendingGameCollectLock(lockKey string, member string) bool {
	client := cs.GetRedisService()
	if client == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ok, err := client.SetNX(ctx, lockKey, member, pendingGameCollectLockTTL).Result()
	if err != nil {
		log.Error("设置单游戏采集锁失败: ", err)
		return false
	}
	return ok
}

func releasePendingGameCollectLock(lockKey string) {
	if err := cs.Del(lockKey); err != nil {
		log.Error("释放单游戏采集锁失败: ", err)
	}
}

func removePendingGameCollectMember(member string) {
	client := cs.GetRedisService()
	if client == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.SRem(ctx, pendingGameCollectSetKey, member).Err(); err != nil {
		log.Error("移除单游戏采集队列成员失败: ", err)
	}
}
