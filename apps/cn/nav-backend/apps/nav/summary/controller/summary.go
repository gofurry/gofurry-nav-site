package controller

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/summary/service"
	"github.com/gofurry/gofurry-nav-backend/common"
	"github.com/gofurry/gofurry-nav-backend/common/util"
)

type summaryApi struct{}

var SummaryApi *summaryApi

func init() {
	SummaryApi = &summaryApi{}
}

func (api summaryApi) GetSiteSummary(c fiber.Ctx) error {
	siteID, parseErr := util.String2Int64(c.Params("siteId"))
	if parseErr != nil || siteID <= 0 {
		return common.NewResponse(c).Error("siteId 参数非法")
	}

	data, err := service.GetSummaryService().GetSiteSummary(siteID)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api summaryApi) GetTargetSummary(c fiber.Ctx) error {
	siteID, parseErr := util.String2Int64(c.Params("siteId"))
	if parseErr != nil || siteID <= 0 {
		return common.NewResponse(c).Error("siteId 参数非法")
	}
	target := c.Params("target")

	data, err := service.GetSummaryService().GetTargetSummary(siteID, target)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}
