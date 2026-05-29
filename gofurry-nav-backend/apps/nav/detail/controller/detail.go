package controller

import (
	"net/url"
	"sync"

	"github.com/gofiber/fiber/v3"
	detailmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/detail/models"
	detailservice "github.com/gofurry/gofurry-nav-backend/apps/nav/detail/service"
	"github.com/gofurry/gofurry-nav-backend/common"
	"github.com/gofurry/gofurry-nav-backend/common/util"
)

type detailReader interface {
	GetSiteDetail(siteID int64, lang string, target string, payloadMode string) (detailmodels.SiteDetailResponse, common.GFError)
	GetTargetLatest(siteID int64, target string, payloadMode string) (detailmodels.TargetLatestResponse, common.GFError)
	ListTargetObservations(siteID int64, target string, protocol string, limit int, payloadMode string) (detailmodels.TargetObservationsResponse, common.GFError)
	GetTargetTrend(siteID int64, target string) (detailmodels.TargetTrendResponse, common.GFError)
	GetTargetChanges(siteID int64, target string) (detailmodels.TargetChangesResponse, common.GFError)
	GetTargetLightProbes(siteID int64, target string, payloadMode string) (detailmodels.TargetLatestResponse, common.GFError)
}

type detailApi struct{}

var DetailApi *detailApi
var detailSvc detailReader
var detailSvcMu sync.RWMutex

func init() {
	DetailApi = &detailApi{}
}

func (api detailApi) GetSiteDetail(c fiber.Ctx) error {
	siteID, parseErr := util.String2Int64(c.Params("siteId"))
	if parseErr != nil || siteID <= 0 {
		return common.NewResponse(c).Error("siteId 参数非法")
	}

	data, err := currentDetailService().GetSiteDetail(siteID, c.Query("lang", "zh"), c.Query("target"), c.Query("payload_mode"))
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api detailApi) GetTargetLatest(c fiber.Ctx) error {
	siteID, parseErr := util.String2Int64(c.Params("siteId"))
	if parseErr != nil || siteID <= 0 {
		return common.NewResponse(c).Error("siteId 参数非法")
	}

	data, err := currentDetailService().GetTargetLatest(siteID, targetParam(c), c.Query("payload_mode"))
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api detailApi) ListTargetObservations(c fiber.Ctx) error {
	siteID, parseErr := util.String2Int64(c.Params("siteId"))
	if parseErr != nil || siteID <= 0 {
		return common.NewResponse(c).Error("siteId 参数非法")
	}

	limit := 0
	limitValue := c.Query("limit")
	if limitValue != "" {
		parsedLimit, limitErr := util.String2Int(limitValue)
		if limitErr != nil {
			return common.NewResponse(c).Error("limit 参数非法")
		}
		limit = parsedLimit
	}

	data, err := currentDetailService().ListTargetObservations(siteID, targetParam(c), c.Query("protocol"), limit, c.Query("payload_mode"))
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api detailApi) GetTargetTrend(c fiber.Ctx) error {
	siteID, parseErr := util.String2Int64(c.Params("siteId"))
	if parseErr != nil || siteID <= 0 {
		return common.NewResponse(c).Error("siteId 参数非法")
	}

	data, err := currentDetailService().GetTargetTrend(siteID, targetParam(c))
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api detailApi) GetTargetChanges(c fiber.Ctx) error {
	siteID, parseErr := util.String2Int64(c.Params("siteId"))
	if parseErr != nil || siteID <= 0 {
		return common.NewResponse(c).Error("siteId 参数非法")
	}

	data, err := currentDetailService().GetTargetChanges(siteID, targetParam(c))
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func (api detailApi) GetTargetLightProbes(c fiber.Ctx) error {
	siteID, parseErr := util.String2Int64(c.Params("siteId"))
	if parseErr != nil || siteID <= 0 {
		return common.NewResponse(c).Error("siteId 参数非法")
	}

	data, err := currentDetailService().GetTargetLightProbes(siteID, targetParam(c), c.Query("payload_mode"))
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).SuccessWithData(data)
}

func currentDetailService() detailReader {
	detailSvcMu.RLock()
	svc := detailSvc
	detailSvcMu.RUnlock()
	if svc != nil {
		return svc
	}

	detailSvcMu.Lock()
	defer detailSvcMu.Unlock()
	if detailSvc == nil {
		detailSvc = detailservice.GetDetailService()
	}
	return detailSvc
}

func targetParam(c fiber.Ctx) string {
	target := c.Params("target")
	decoded, err := url.PathUnescape(target)
	if err != nil {
		return target
	}
	return decoded
}

func setDetailReaderForTest(reader detailReader) func() {
	detailSvcMu.Lock()
	previous := detailSvc
	detailSvc = reader
	detailSvcMu.Unlock()
	return func() {
		detailSvcMu.Lock()
		detailSvc = previous
		detailSvcMu.Unlock()
	}
}
