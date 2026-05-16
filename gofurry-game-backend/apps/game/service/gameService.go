package service

import (
	"github.com/gofurry/gofurry-game-backend/apps/game/dao"
	"github.com/gofurry/gofurry-game-backend/apps/game/models"
	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/common/log"
	cs "github.com/gofurry/gofurry-game-backend/common/service"
	"github.com/gofurry/gofurry-game-backend/common/util"
	"github.com/bytedance/sonic"
)

type gameService struct{}

var gameSingleton = new(gameService)

func GetGameService() *gameService { return gameSingleton }

// 查询 weight 前 num 条游戏记录
func (s gameService) GetGameList(num string, lang string) (gameVo []models.GameRespVo, err common.GFError) {
	intNum, parseErr := util.String2Int(num)
	if parseErr != nil {
		return gameVo, common.NewServiceError("入参转换错误")
	}
	gameList, err := dao.GetGameDao().GetGameList(intNum)
	if err != nil {
		return
	}

	for _, v := range gameList {
		newGameVo := models.GameRespVo{
			ID:          util.Int642String(v.ID),
			CreateTime:  v.CreateTime,
			UpdateTime:  v.UpdateTime,
			ReleaseDate: v.ReleaseDate,
			Appid:       util.Int642String(v.Appid),
			Header:      v.Header,
		}
		switch lang {
		case "zh":
			newGameVo.Name = v.Name
			newGameVo.Info = v.Info
		case "en":
			newGameVo.Name = v.NameEn
			newGameVo.Info = v.InfoEn
		default:
			newGameVo.Name = v.Name
			newGameVo.Info = v.Info
		}
		jsonErr := sonic.Unmarshal([]byte(v.Developers), &newGameVo.Developers)
		if jsonErr != nil {
			log.Error(v.Name, " ([]byte(*v.Developers), &newGameVo.Developers) err: ", jsonErr)
		}
		jsonErr = sonic.Unmarshal([]byte(v.Publishers), &newGameVo.Publishers)
		if jsonErr != nil {
			log.Error(v.Name, " ([]byte(*v.Publishers), &newGameVo.Publishers) err: ", jsonErr)
		}
		if v.Resources != nil {
			jsonErr = sonic.Unmarshal([]byte(*v.Resources), &newGameVo.Resources)
			if jsonErr != nil {
				log.Error(v.Name, " ([]byte(*v.Resources), &newGameVo.Resources) err: ", jsonErr)
			}
		}
		if v.Groups != nil {
			jsonErr = sonic.Unmarshal([]byte(*v.Groups), &newGameVo.Groups)
			if jsonErr != nil {
				log.Error(v.Name, " ([]byte(*v.Groups), &newGameVo.Groups) err: ", jsonErr)
			}
		}
		if v.Links != nil {
			jsonErr = sonic.Unmarshal([]byte(*v.Links), &newGameVo.Links)
			if jsonErr != nil {
				log.Error(v.Name, " ([]byte(*v.Links), &newGameVo.Links) err: ", jsonErr)
			}
		}
		gameVo = append(gameVo, newGameVo)
	}
	return
}

func (s gameService) GetGameMainList() (res models.GameMainInfoVo, err common.GFError) {
	jsonStr, err := cs.GetString("game-info:latest")
	if err != nil {
		return res, err
	}
	jsonErr := sonic.Unmarshal([]byte(jsonStr), &res.Latest)
	if jsonErr != nil {
		return res, common.NewServiceError(err.GetMsg())
	}

	jsonStr, err = cs.GetString("game-info:recent")
	if err != nil {
		return res, err
	}
	jsonErr = sonic.Unmarshal([]byte(jsonStr), &res.Recent)
	if jsonErr != nil {
		return res, common.NewServiceError(err.GetMsg())
	}

	jsonStr, err = cs.GetString("game-info:hot")
	if err != nil {
		return res, err
	}
	jsonErr = sonic.Unmarshal([]byte(jsonStr), &res.Hot)
	if jsonErr != nil {
		return res, common.NewServiceError(err.GetMsg())
	}

	jsonStr, err = cs.GetString("game-info:free")
	if err != nil {
		return res, err
	}
	jsonErr = sonic.Unmarshal([]byte(jsonStr), &res.Free)
	if jsonErr != nil {
		return res, common.NewServiceError(err.GetMsg())
	}
	return
}

func (s gameService) GetPanelMainList() (res models.GameMainPanelVo, err common.GFError) {
	// 在线人数 1
	jsonStr, err := cs.GetString("game-panel:top-player-count-1st")
	if err != nil {
		return res, err
	}
	if jsonErr := sonic.Unmarshal([]byte(jsonStr), &res.TopCount.One); jsonErr != nil {
		return res, common.NewServiceError(err.GetMsg())
	}
	// 在线人数 2
	jsonStr, err = cs.GetString("game-panel:top-player-count-2st")
	if err != nil {
		return res, err
	}
	if jsonErr := sonic.Unmarshal([]byte(jsonStr), &res.TopCount.Two); jsonErr != nil {
		return res, common.NewServiceError(err.GetMsg())
	}
	// 在线人数 3
	jsonStr, err = cs.GetString("game-panel:top-player-count-3st")
	if err != nil {
		return res, err
	}
	if jsonErr := sonic.Unmarshal([]byte(jsonStr), &res.TopCount.Three); jsonErr != nil {
		return res, common.NewServiceError(err.GetMsg())
	}
	// 在线人数 4
	jsonStr, err = cs.GetString("game-panel:top-player-count-4st")
	if err != nil {
		return res, err
	}
	if jsonErr := sonic.Unmarshal([]byte(jsonStr), &res.TopCount.Four); jsonErr != nil {
		return res, common.NewServiceError(err.GetMsg())
	}

	// 最高价格
	jsonStr, err = cs.GetString("game-panel:top-price")
	if err != nil {
		return res, err
	}
	if jsonErr := sonic.Unmarshal([]byte(jsonStr), &res.TopPriceVo); jsonErr != nil {
		return res, common.NewServiceError(err.GetMsg())
	}
	// 最高折扣
	jsonStr, err = cs.GetString("game-panel:top-discount")
	if err != nil {
		return res, err
	}
	if jsonErr := sonic.Unmarshal([]byte(jsonStr), &res.TopDiscountVo); jsonErr != nil {
		return res, common.NewServiceError(err.GetMsg())
	}

	// 最低价格 1
	jsonStr, err = cs.GetString("game-panel:bottom-price-1st")
	if err != nil {
		return res, err
	}
	if jsonErr := sonic.Unmarshal([]byte(jsonStr), &res.BottomPrice.One); jsonErr != nil {
		return res, common.NewServiceError(err.GetMsg())
	}
	// 最低价格 2
	jsonStr, err = cs.GetString("game-panel:bottom-price-2st")
	if err != nil {
		return res, err
	}
	if jsonErr := sonic.Unmarshal([]byte(jsonStr), &res.BottomPrice.Two); jsonErr != nil {
		return res, common.NewServiceError(err.GetMsg())
	}
	// 最低价格 3
	jsonStr, err = cs.GetString("game-panel:bottom-price-3st")
	if err != nil {
		return res, err
	}
	if jsonErr := sonic.Unmarshal([]byte(jsonStr), &res.BottomPrice.Three); jsonErr != nil {
		return res, common.NewServiceError(err.GetMsg())
	}
	// 最低价格 4
	jsonStr, err = cs.GetString("game-panel:bottom-price-4st")
	if err != nil {
		return res, err
	}
	if jsonErr := sonic.Unmarshal([]byte(jsonStr), &res.BottomPrice.Four); jsonErr != nil {
		return res, common.NewServiceError(err.GetMsg())
	}

	return
}

func (s gameService) GetUpdateNews() (res models.UpdateNewsVo, err common.GFError) {
	jsonStr, err := cs.GetString("game-news:latest")
	if err != nil {
		return res, err
	}
	jsonErr := sonic.Unmarshal([]byte(jsonStr), &res)
	if jsonErr != nil {
		return res, common.NewServiceError(err.GetMsg())
	}

	return
}

func (s gameService) GetTagList(lang string) (res []models.TagModelVo, err common.GFError) {
	return dao.GetGameDao().GetTagList(lang)
}

func (s gameService) GetGameInfo(id string, lang string) (res models.GameBaseInfoVo, err common.GFError) {
	intId, parseErr := util.String2Int64(id)
	if parseErr != nil {
		return res, common.NewServiceError("Game ID 转换有误")
	}
	game, err := dao.GetGameDao().GetGame(intId)
	if err != nil {
		return res, err
	}
	// 名字 & 简介
	switch lang {
	case "en":
		res.Name = game.NameEn
		res.Info = game.InfoEn
	default:
		res.Name = game.Name
		res.Info = game.Info
	}
	// 创建时间
	res.CreateTime = game.CreateTime
	// 资源 & 社群 & 相关链接
	if game.Resources != nil {
		parseErr = sonic.Unmarshal([]byte(*game.Resources), &res.Resources)
		if parseErr != nil {
			return res, common.NewServiceError(parseErr.Error())
		}
	}
	if game.Groups != nil {
		parseErr = sonic.Unmarshal([]byte(*game.Groups), &res.Groups)
		if parseErr != nil {
			return res, common.NewServiceError(parseErr.Error())
		}
	}
	if game.Links != nil {
		parseErr = sonic.Unmarshal([]byte(*game.Links), &res.Links)
		if parseErr != nil {
			return res, common.NewServiceError(parseErr.Error())
		}
	}
	// 发售时间 & appid & 封面图
	res.ReleaseDate = game.ReleaseDate
	res.Appid = game.Appid
	res.Cover = game.Header
	// 发行商 & 开发商
	var devs, pubs []string
	parseErr = sonic.Unmarshal([]byte(game.Developers), &devs)
	if parseErr != nil {
		return res, common.NewServiceError(parseErr.Error())
	}
	parseErr = sonic.Unmarshal([]byte(game.Publishers), &pubs)
	if parseErr != nil {
		return res, common.NewServiceError(parseErr.Error())
	}
	res.Developers = devs
	res.Publishers = pubs

	record, err := dao.GetGameDao().GetGameRecord(intId, lang)
	if err != nil {
		return res, err
	}
	var priceList []models.PriceModel
	parseErr = sonic.Unmarshal([]byte(record.PriceList), &priceList)
	if parseErr != nil {
		return res, common.NewServiceError(parseErr.Error())
	}
	res.PriceList = priceList

	news, err := dao.GetGameDao().GetGameNews(intId, lang)
	if err != nil {
		return res, err
	}
	for _, v := range news {
		res.News = append(res.News, models.NewsVo{
			Headline: v.Headline,
			Content:  v.Content,
			PostTime: v.PostTime,
			Author:   v.Author,
			URL:      v.URL,
		})
	}

	tags, err := dao.GetGameDao().GetGameTags(intId, lang)
	if err != nil {
		return res, err
	}
	res.Tags = tags

	key := "game:"
	switch lang {
	case "en":
		key += "en"
	default:
		key += "zh"
	}
	data, err := cs.GetString(key + "-info" + id)
	if err != nil {
		return res, err
	}
	var gameRecord models.GameSaveModel
	jsonErr := sonic.Unmarshal([]byte(data), &gameRecord)
	if jsonErr != nil {
		return res, common.NewServiceError(jsonErr.Error())
	}
	res.SupportedLanguages = gameRecord.SupportedLanguages
	res.RequiredAge = gameRecord.RequiredAge
	res.Website = gameRecord.Website
	res.DetailedDescription = gameRecord.DetailedDescription
	res.AboutTheGame = gameRecord.AboutTheGame
	res.PcRequirements = gameRecord.PcRequirements
	res.Support = gameRecord.Support
	res.Movies = gameRecord.Movies
	res.Screenshots = gameRecord.Screenshots
	res.Platform = gameRecord.Platforms
	// 采集时间
	res.UpdateTime = gameRecord.CollectDate

	// 在线人数
	data, err = cs.GetString("game:online" + id)
	if err != nil {
		return res, err
	}
	var gameOnlineModel models.GameOnlineModel
	jsonErr = sonic.Unmarshal([]byte(data), &gameOnlineModel)
	if jsonErr != nil {
		return res, common.NewServiceError(jsonErr.Error())
	}
	res.OnlineCount = gameOnlineModel.Count
	res.CountCollectTime = gameOnlineModel.CreateTime

	return
}

func (s gameService) GetGameRemark(id string) (res models.GameRemarkVo, err common.GFError) {
	intId, parseErr := util.String2Int64(id)
	if parseErr != nil {
		return res, common.NewServiceError("Game ID 转换有误")
	}
	res, err = dao.GetGameDao().GetGameComment(intId)
	if err != nil {
		return res, err
	}

	// 脱敏 IP
	for i := range res.Remarks {
		res.Remarks[i].IP = util.DesensitizeIP(res.Remarks[i].IP)
	}
	return
}

func (s gameService) GetGameCreator(lang string) (res []models.CreatorVo, err common.GFError) {
	record, err := cs.GetString("game-creator:list")
	if err != nil {
		return res, err
	}
	var gameCreatorModel models.UpdateCreatorVo
	jsonErr := sonic.Unmarshal([]byte(record), &gameCreatorModel)
	if jsonErr != nil {
		return res, common.NewServiceError(jsonErr.Error())
	}
	switch lang {
	case "en":
		res = gameCreatorModel.CreatorEn
	case "zh":
		res = gameCreatorModel.CreatorZh
	}

	return
}

func (s gameService) GetMoreUpdateNews(lang string) (res []models.UpdateNewsModels, err common.GFError) {
	jsonStr, err := cs.GetString("game-news:latest-more")
	if err != nil {
		return res, err
	}
	var record models.UpdateNewsVo
	jsonErr := sonic.Unmarshal([]byte(jsonStr), &record)
	if jsonErr != nil {
		return res, common.NewServiceError(err.GetMsg())
	}

	switch lang {
	case "en":
		res = record.NewsEn
	case "zh":
		res = record.NewsZh
	}

	return
}
