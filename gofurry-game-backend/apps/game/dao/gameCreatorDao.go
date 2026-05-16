package dao

import (
	"github.com/gofurry/gofurry-game-backend/apps/game/models"
	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/common/abstract"
)

var newGameCreatorDao = new(gameCreatorDao)

func init() {
	newGameCreatorDao.Init()
}

type gameCreatorDao struct{ abstract.Dao }

func GetGameCreatorDao() *gameCreatorDao { return newGameCreatorDao }

func (dao gameCreatorDao) GetGameCreator(lang string) (res []models.TempCreator, err common.GFError) {
	var selectFields string
	switch lang {
	case "en":
		selectFields = `
			id, 
			COALESCE(name_en, name) as name, 
			COALESCE(info_en, info) as info, 
			main_url as url, 
			cover as avatar, 
			links, 
			contact, 
			type, 
			create_time, 
			update_time
		`
	case "zh":
		selectFields = `
			id, 
			name, 
			info, 
			main_url as url, 
			cover as avatar, 
			links, 
			contact, 
			type, 
			create_time, 
			update_time
		`
	default:
		return res, common.NewDaoError("unsupported language: " + lang)
	}

	db := dao.Gm.Table(models.TableNameGfgGameCreator).Select(selectFields)
	db.Where("deleted IS NOT TRUE")
	if dbErr := db.Find(&res).Error; dbErr != nil {
		return res, common.NewDaoError(dbErr.Error())
	}
	return res, nil
}
