package controller

import (
	"context"

	"github.com/gofiber/fiber/v3"
	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	"github.com/gofurry/gofurry-game-backend/common"
)

func (api *GameV2API) GetCollectStatus(c fiber.Ctx) error {
	data, err := api.newReadModelService().GetCollectStatus(context.Background())
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *GameV2API) ListCollectRuns(c fiber.Ctx) error {
	data, err := api.newReadModelService().ListCollectRuns(context.Background(), v2models.GameV2CollectRunQuery{
		TaskType: c.Query("task_type", ""),
		Status:   c.Query("status", ""),
		Limit:    parseInt(c.Query("limit", "20")),
		Offset:   parseInt(c.Query("offset", "0")),
	})
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *GameV2API) GetCollectRun(c fiber.Ctx) error {
	data, err := api.newReadModelService().GetCollectRun(context.Background(), c.Params("run_id"))
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *GameV2API) ListCollectTaskResults(c fiber.Ctx) error {
	data, err := api.newReadModelService().ListCollectTaskResults(context.Background(), v2models.GameV2CollectTaskResultQuery{
		RunID:    c.Query("run_id", ""),
		TaskType: c.Query("task_type", ""),
		Status:   c.Query("status", ""),
		GameID:   parseInt64(c.Query("game_id", "0")),
		AppID:    parseInt64(c.Query("appid", "0")),
		Limit:    parseInt(c.Query("limit", "50")),
		Offset:   parseInt(c.Query("offset", "0")),
	})
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *GameV2API) GetGameCollectStatus(c fiber.Ctx) error {
	id := parseInt64(c.Params("id"))
	appid := parseInt64(c.Query("appid", "0"))
	data, err := api.newReadModelService().GetGameCollectStatus(context.Background(), id, appid)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}
