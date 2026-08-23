package task

import (
	"context"

	v2dao "github.com/gofurry/gofurry-game-backend/apps/game/v2/dao"
	v2service "github.com/gofurry/gofurry-game-backend/apps/game/v2/service"
	"github.com/gofurry/gofurry-game-backend/common/log"
)

var gameHomeCacheLangs = []string{"zh", "en"}

func RefreshGameHomeCache(readDAO *v2dao.ReadModelDAO) {
	log.Info("[RefreshGameHomeCache] start")

	svc := v2service.NewReadModelServiceWithReader(readDAO)
	ctx := context.Background()

	for _, lang := range gameHomeCacheLangs {
		if _, err := svc.RefreshHomeCache(ctx, lang, "CN"); err != nil {
			log.Error("[RefreshGameHomeCache] refresh failed", "lang", lang, "region", "CN", "error", err)
			continue
		}
		log.Info("[RefreshGameHomeCache] refresh success", "lang", lang, "region", "CN")
	}

	log.Info("[RefreshGameHomeCache] done")
}
