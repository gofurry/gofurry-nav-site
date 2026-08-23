package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/gofurry/gofurry-admin/internal/app/gameadmin/models"
	"github.com/gofurry/gofurry-admin/internal/infra/cache"
	log "github.com/gofurry/gofurry-admin/internal/infra/logging"
)

const gameCollectPendingSetKey = "game:v2:collect:pending"

func enqueueCreatedGameCollect(game models.Game) {
	if game.ID <= 0 || game.Appid <= 0 {
		return
	}

	client := cache.GetRedisService()
	if client == nil {
		log.Warn("skip game v2 single collect enqueue: redis service is not ready")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	member := fmt.Sprintf("%d:%d", game.ID, game.Appid)
	if err := client.SAdd(ctx, gameCollectPendingSetKey, member).Err(); err != nil {
		log.Errorf("enqueue game v2 single collect failed, game_id=%d appid=%d err=%v", game.ID, game.Appid, err)
		return
	}

	log.Infof("enqueue game v2 single collect, game_id=%d appid=%d", game.ID, game.Appid)
}
