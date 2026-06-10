package controller

import (
	"context"

	"github.com/gofiber/fiber/v3"
	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	"github.com/gofurry/gofurry-game-backend/common"
)

func (api *gameV2Api) GetCollectStatus(c fiber.Ctx) error {
	data, err := newReadModelService().GetCollectStatus(context.Background())
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *gameV2Api) ListCollectRuns(c fiber.Ctx) error {
	data, err := newReadModelService().ListCollectRuns(context.Background(), v2models.GameV2CollectRunQuery{
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

func (api *gameV2Api) GetCollectRun(c fiber.Ctx) error {
	data, err := newReadModelService().GetCollectRun(context.Background(), c.Params("run_id"))
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *gameV2Api) ListCollectTaskResults(c fiber.Ctx) error {
	data, err := newReadModelService().ListCollectTaskResults(context.Background(), v2models.GameV2CollectTaskResultQuery{
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

func (api *gameV2Api) GetGameCollectStatus(c fiber.Ctx) error {
	id := parseInt64(c.Params("id"))
	appid := parseInt64(c.Query("appid", "0"))
	data, err := newReadModelService().GetGameCollectStatus(context.Background(), id, appid)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}
