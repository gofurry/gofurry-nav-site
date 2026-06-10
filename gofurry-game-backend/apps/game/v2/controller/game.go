package controller

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v3"
	v2dao "github.com/gofurry/gofurry-game-backend/apps/game/v2/dao"
	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	v2service "github.com/gofurry/gofurry-game-backend/apps/game/v2/service"
	"github.com/gofurry/gofurry-game-backend/common"
)

type gameV2Api struct{}

var GameV2Api *gameV2Api

func init() {
	GameV2Api = &gameV2Api{}
}

func (api *gameV2Api) GetGameList(c fiber.Ctx) error {
	data, err := newReadModelService().GetGameList(context.Background(), v2models.GameV2ListQuery{
		Lang:   c.Query("lang", "zh"),
		Region: c.Query("region", "CN"),
		Limit:  parseInt(c.Query("limit", "20")),
		Offset: parseInt(c.Query("offset", "0")),
		Sort:   c.Query("sort", "weight"),
	})
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *gameV2Api) GetGameInfo(c fiber.Ctx) error {
	id := parseInt64(c.Query("id", "0"))
	appid := parseInt64(c.Query("appid", "0"))
	if id <= 0 && appid <= 0 {
		return common.NewResponse(c).Error("id 或 appid 不能为空")
	}
	data, err := newReadModelService().GetGameDetail(context.Background(), v2models.GameV2DetailRequest{
		GameID:    id,
		AppID:     appid,
		Lang:      c.Query("lang", "zh"),
		Region:    c.Query("region", "CN"),
		NewsLimit: parseInt(c.Query("news_limit", "5")),
	})
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *gameV2Api) SearchSimple(c fiber.Ctx) error {
	req := v2models.GameV2SearchRequest{}
	if err := c.Bind().Body(&req); err != nil {
		return common.NewResponse(c).Error("解析请求体失败")
	}
	data, err := newReadModelService().SimpleSearch(context.Background(), req)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *gameV2Api) SearchPage(c fiber.Ctx) error {
	req := v2models.GameV2SearchPageQueryRequest{}
	if err := c.Bind().Body(&req); err != nil {
		return common.NewResponse(c).Error("解析请求体失败")
	}
	data, err := newReadModelService().SearchPage(context.Background(), req)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *gameV2Api) GetTags(c fiber.Ctx) error {
	data, err := newReadModelService().ListTags(context.Background(), c.Query("lang", "zh"))
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *gameV2Api) GetGameNews(c fiber.Ctx) error {
	id := parseInt64(c.Query("id", "0"))
	appid := parseInt64(c.Query("appid", "0"))
	if id <= 0 && appid <= 0 {
		return common.NewResponse(c).Error("id 或 appid 不能为空")
	}
	data, err := newReadModelService().GetGameNews(context.Background(), v2models.GameV2NewsQuery{
		GameID: id,
		AppID:  appid,
		Lang:   c.Query("lang", "zh"),
		Limit:  parseInt(c.Query("limit", "20")),
		Offset: parseInt(c.Query("offset", "0")),
	})
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *gameV2Api) GetLatestGameNews(c fiber.Ctx) error {
	data, err := newReadModelService().GetLatestGameNews(context.Background(), v2models.GameV2NewsQuery{
		Lang:   c.Query("lang", "zh"),
		Limit:  parseInt(c.Query("limit", "20")),
		Offset: parseInt(c.Query("offset", "0")),
	})
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *gameV2Api) GetPanelMain(c fiber.Ctx) error {
	data, err := newReadModelService().GetPanelMain(context.Background(), v2models.GameV2PanelQuery{
		Lang:      c.Query("lang", "zh"),
		Region:    c.Query("region", "CN"),
		Limit:     parseInt(c.Query("limit", "8")),
		NewsLimit: parseInt(c.Query("news_limit", "8")),
	})
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func newReadModelService() *v2service.ReadModelService {
	return v2service.NewReadModelServiceWithReader(v2dao.NewReadModelDAO())
}

func parseInt(value string) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

func parseInt64(value string) int64 {
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}
