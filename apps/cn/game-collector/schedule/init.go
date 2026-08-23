package schedule

import (
	game "github.com/gofurry/gofurry-game-collector/collector/game/controller"
	"github.com/gofurry/gofurry-game-collector/common/log"
)

func InitSchedule() (initialized bool) {
	defer func() {
		if err := recover(); err != nil {
			initialized = false
			log.Error(err)
		}
	}()

	// 初始化 Game 采集模块
	game.GameApi.InitGameCollection()
	return true
}
