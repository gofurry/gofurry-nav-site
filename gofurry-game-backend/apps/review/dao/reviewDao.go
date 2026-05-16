package dao

import (
	"errors"

	gm "github.com/gofurry/gofurry-game-backend/apps/game/models"
	"github.com/gofurry/gofurry-game-backend/apps/review/models"
	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/common/abstract"
	"gorm.io/gorm"
)

var newReviewDao = new(reviewDao)

func init() {
	newReviewDao.Init()
}

type reviewDao struct{ abstract.Dao }

func GetReviewDao() *reviewDao { return newReviewDao }

func (dao reviewDao) GetHotGame(num int) (res []models.AvgScoreResult, err common.GFError) {
	var results []models.AvgScoreResult
	commentTable := models.TableNameGfgGameComment
	gameTable := gm.TableNameGfgGame

	db := dao.Gm.Table(commentTable).
		Joins("LEFT JOIN "+gameTable+" ON "+commentTable+".game_id = "+gameTable+".id").
		Select(
			commentTable+".game_id",
			"AVG("+commentTable+".score) AS avg_score",
			"COUNT(*) AS comment_count",
			gameTable+".name",
			gameTable+".name_en",
			gameTable+".info",
			gameTable+".info_en",
			gameTable+".header",
		).
		Group(commentTable + ".game_id, " + gameTable + ".name, " + gameTable + ".name_en, " +
			gameTable + ".info, " + gameTable + ".info_en, " + gameTable + ".header").
		Order("avg_score DESC").
		Limit(num)

	if dbErr := db.Find(&results).Error; dbErr != nil {
		return res, common.NewDaoError(dbErr.Error())
	}

	return results, nil
}

func (dao reviewDao) GetScoreById(id int64) (res models.AvgScoreResult, err common.GFError) {
	commentTable := models.TableNameGfgGameComment
	gameTable := gm.TableNameGfgGame

	db := dao.Gm.Table(commentTable).
		Joins("LEFT JOIN "+gameTable+" ON "+commentTable+".game_id = "+gameTable+".id").
		Select(
			commentTable+".game_id",
			"AVG("+commentTable+".score) AS avg_score",
			"COUNT(*) AS comment_count",
			gameTable+".name",
			gameTable+".name_en",
			gameTable+".info",
			gameTable+".info_en",
			gameTable+".header",
		).
		Where(commentTable+".game_id = ?", id).
		Group(commentTable + ".game_id, " + gameTable + ".name, " + gameTable + ".name_en, " + gameTable +
			".info, " + gameTable + ".info_en, " + gameTable + ".header")

	if dbErr := db.Take(&res).Error; dbErr != nil {
		return res, common.NewDaoError(dbErr.Error())
	}

	return res, nil
}

func (dao reviewDao) GetReviewByIPAndName(id string, ip string, name string) (res models.GfgGameComment, err common.GFError) {
	db := dao.Gm.Table(models.TableNameGfgGameComment).Where("ip = ? AND game_id = ? AND name = ?", ip, id, name)
	db.Take(&res)

	if dbErr := db.Take(&res).Error; dbErr != nil {
		if errors.Is(dbErr, gorm.ErrRecordNotFound) {
			return res, common.NewDaoError(common.RETURN_RECORD_NOT_FOUND)
		}
		return res, common.NewDaoError(dbErr.Error())
	}
	return res, nil
}

func (dao reviewDao) GetListByLimit(num int, lang string) (res []models.AnonymousReviewResponse, err common.GFError) {
	selectFields := `
		gfg_game_comment.region, 
		gfg_game_comment.score, 
		gfg_game_comment.content, 
		gfg_game_comment.ip, 
		gfg_game_comment.create_time as time,
		gfg_game.header as game_cover
	`

	// 根据语言选择游戏名称字段
	if lang == "en" {
		selectFields += ", gfg_game.name_en as game_name"
	} else {
		selectFields += ", gfg_game.name as game_name"
	}

	db := dao.Gm.Table(models.TableNameGfgGameComment).
		Select(selectFields).
		Joins("LEFT JOIN gfg_game ON gfg_game_comment.game_id = gfg_game.id").
		Order("gfg_game_comment.create_time DESC").
		Limit(num).
		Find(&res)

	// 检查数据库错误
	if dbErr := db.Error; dbErr != nil {
		return res, common.NewDaoError(dbErr.Error())
	}
	return res, nil
}
