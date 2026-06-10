package controller

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	"github.com/gofurry/gofurry-game-backend/common"
)

func (api *gameV2Api) GetSyncGameList(c fiber.Ctx) error {
	data, err := newReadModelService().ListSyncGames(context.Background(), v2models.GameV2SyncListQuery{
		Lang:         c.Query("lang", "zh"),
		Region:       c.Query("region", "CN"),
		Limit:        parseInt(c.Query("limit", "5000")),
		Offset:       parseInt(c.Query("offset", "0")),
		UpdatedSince: parseUpdatedSince(c.Query("updated_since", "")),
	})
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *gameV2Api) GetSyncGameInfo(c fiber.Ctx) error {
	id := parseInt64(c.Query("id", "0"))
	appid := parseInt64(c.Query("appid", "0"))
	if id <= 0 && appid <= 0 {
		return common.NewResponse(c).Error("id 或 appid 不能为空")
	}
	data, err := newReadModelService().GetSyncGameDetail(context.Background(), v2models.GameV2DetailRequest{
		GameID: id,
		AppID:  appid,
		Lang:   c.Query("lang", "zh"),
		Region: c.Query("region", "CN"),
	})
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *gameV2Api) GetSyncGameNews(c fiber.Ctx) error {
	data, err := newReadModelService().ListSyncGameNews(context.Background(), v2models.GameV2SyncNewsQuery{
		Lang:         c.Query("lang", "zh"),
		Limit:        parseInt(c.Query("limit", "5000")),
		Offset:       parseInt(c.Query("offset", "0")),
		UpdatedSince: parseUpdatedSince(c.Query("updated_since", "")),
	})
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *gameV2Api) GetSyncCreators(c fiber.Ctx) error {
	data, err := newReadModelService().ListSyncCreators(context.Background(), c.Query("lang", "zh"))
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func parseUpdatedSince(value string) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed
	}
	if parsed, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
		return parsed
	}
	if parsed, err := time.Parse("2006-01-02", value); err == nil {
		return parsed
	}
	return time.Time{}
}
