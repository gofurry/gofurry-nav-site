package controller

import (
	"github.com/gofiber/fiber/v3"
	collectmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/collect/models"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/collect/service"
	"github.com/gofurry/gofurry-nav-backend/common"
	"github.com/gofurry/gofurry-nav-backend/common/util"
)

type collectReader interface {
	GetStatus() (collectmodels.CollectStatus, common.GFError)
	ListObservations(query collectmodels.ObservationQuery) ([]collectmodels.ObservationItem, common.GFError)
	GetSiteStatus(siteID int64) (collectmodels.SiteCollectStatus, common.GFError)
	GetTargetStatus(siteID int64, target string) (collectmodels.TargetCollectStatus, common.GFError)
}

type collectApi struct{ reader collectReader }

var CollectApi = &collectApi{}

func New(reader collectReader) *collectApi { return &collectApi{reader: reader} }

func (api *collectApi) service() collectReader {
	if api.reader != nil {
		return api.reader
	}
	return service.GetCollectService()
}

func (api *collectApi) GetStatus(c fiber.Ctx) error {
	data, err := api.service().GetStatus()
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *collectApi) ListObservations(c fiber.Ctx) error {
	data, err := api.service().ListObservations(collectmodels.ObservationQuery{
		SiteID:   parseInt64(c.Query("site_id", "0")),
		Target:   c.Query("target", ""),
		Protocol: c.Query("protocol", ""),
		Status:   c.Query("status", ""),
		Limit:    parseInt(c.Query("limit", "50")),
		Offset:   parseInt(c.Query("offset", "0")),
	})
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *collectApi) GetSiteStatus(c fiber.Ctx) error {
	siteID, parseErr := util.String2Int64(c.Params("siteId"))
	if parseErr != nil || siteID <= 0 {
		return common.NewResponse(c).Error("siteId 参数非法")
	}
	data, err := api.service().GetSiteStatus(siteID)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api *collectApi) GetTargetStatus(c fiber.Ctx) error {
	siteID, parseErr := util.String2Int64(c.Params("siteId"))
	if parseErr != nil || siteID <= 0 {
		return common.NewResponse(c).Error("siteId 参数非法")
	}
	data, err := api.service().GetTargetStatus(siteID, c.Params("target"))
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func parseInt(value string) int {
	parsed, err := util.String2Int(value)
	if err != nil {
		return 0
	}
	return parsed
}

func parseInt64(value string) int64 {
	parsed, err := util.String2Int64(value)
	if err != nil {
		return 0
	}
	return parsed
}
