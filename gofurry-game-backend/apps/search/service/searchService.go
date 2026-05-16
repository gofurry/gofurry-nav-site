package service

import (
	"github.com/gofurry/gofurry-game-backend/apps/search/dao"
	"github.com/gofurry/gofurry-game-backend/apps/search/models"
	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/common/log"
	cm "github.com/gofurry/gofurry-game-backend/common/models"
	"github.com/gofurry/gofurry-game-backend/common/util"
)

type searchService struct{}

var searchSingleton = new(searchService)

func GetSearchService() *searchService { return searchSingleton }

func (s searchService) SimpleSearchQuery(req models.SearchRequest) (res []models.SearchGameVo, err common.GFError) {
	games, err := dao.GetSearchDao().GetGameListByText(req.Txt, 8)
	if err != nil {
		return nil, common.NewServiceError(err.GetMsg())
	}
	for _, game := range games {
		newRecord := models.SearchGameVo{
			ID:    util.Int642String(game.ID),
			Cover: game.Header,
		}
		switch req.Lang {
		case "zh":
			newRecord.Name = game.Name
			newRecord.Info = game.Info
		case "en":
			newRecord.Name = game.NameEn
			newRecord.Info = game.InfoEn
		default:
			newRecord.Name = game.Name
			newRecord.Info = game.Info
		}
		res = append(res, newRecord)
	}
	return
}

func (s searchService) SearchPageQuery(req models.SearchPageQueryRequest) (res cm.PageResponse, err common.GFError) {
	res, dbErr := dao.GetSearchDao().Paginate(req)
	if dbErr != nil {
		log.Error("SearchPageQuery Error:", dbErr.GetMsg())
		return res, common.NewServiceError("分页查询失败.")
	}
	return
}
