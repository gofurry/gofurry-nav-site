package dao

import "github.com/gofurry/gofurry-game-backend/common/abstract"

var newGameDao = new(gameDao)

func init() {
	newGameDao.Init()
}

type gameDao struct{ abstract.Dao }

func GetGameDao() *gameDao { return newGameDao }
