package service

import (
	"strings"
	"time"

	collectdao "github.com/gofurry/gofurry-nav-backend/apps/nav/collect/dao"
	collectmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/collect/models"
	readmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/readmodel/models"
	readservice "github.com/gofurry/gofurry-nav-backend/apps/nav/readmodel/service"
	summaryservice "github.com/gofurry/gofurry-nav-backend/apps/nav/summary/service"
	"github.com/gofurry/gofurry-nav-backend/common"
)

type collectStore interface {
	ListObservationSummary() ([]collectmodels.ObservationStatusSummary, common.GFError)
	ListObservations(query collectmodels.ObservationQuery) ([]collectmodels.ObservationItem, common.GFError)
}

type runStateReader interface {
	GetRunState(protocol string) (readmodels.RunStateResponse, common.GFError)
	GetTargetLatest(siteID int64, target string, protocols []string) (readmodels.TargetLatestResponse, common.GFError)
	GetLightProbeLatest(siteID int64, target string) (readmodels.TargetLatestResponse, common.GFError)
	GetTargetTrend(siteID int64, target string) (readmodels.TargetTrendResponse, common.GFError)
	GetTargetChanges(siteID int64, target string) (readmodels.TargetChangesResponse, common.GFError)
}

type collectService struct {
	store collectStore
	read  runStateReader
	now   func() time.Time
}

var collectSingleton = &collectService{}

func GetCollectService() *collectService { return collectSingleton }

func (svc *collectService) GetStatus() (collectmodels.CollectStatus, common.GFError) {
	runs := make([]readmodels.RunStateResponse, 0, len(readmodels.AllProtocols()))
	for _, protocol := range readmodels.AllProtocols() {
		state, err := svc.readModels().GetRunState(protocol)
		if err != nil {
			return collectmodels.CollectStatus{}, err
		}
		runs = append(runs, state)
	}
	summary, err := svc.storeDAO().ListObservationSummary()
	if err != nil {
		return collectmodels.CollectStatus{}, err
	}
	return collectmodels.CollectStatus{
		LatestRuns:  runs,
		Summary:     summary,
		GeneratedAt: svc.clock()(),
	}, nil
}

func (svc *collectService) ListObservations(query collectmodels.ObservationQuery) ([]collectmodels.ObservationItem, common.GFError) {
	query.Target = strings.TrimSpace(query.Target)
	query.Protocol = strings.TrimSpace(query.Protocol)
	query.Status = strings.TrimSpace(query.Status)
	if query.Protocol != "" && !readmodels.IsProtocolAllowed(query.Protocol) {
		return nil, common.NewServiceError("protocol 参数非法")
	}
	return svc.storeDAO().ListObservations(query)
}

func (svc *collectService) GetSiteStatus(siteID int64) (collectmodels.SiteCollectStatus, common.GFError) {
	if siteID <= 0 {
		return collectmodels.SiteCollectStatus{}, common.NewServiceError("siteId 参数非法")
	}
	summary, err := summaryservice.GetSummaryService().GetSiteSummary(siteID)
	if err != nil {
		return collectmodels.SiteCollectStatus{}, err
	}
	return collectmodels.SiteCollectStatus{
		SiteID:      siteID,
		Summary:     summary,
		Targets:     summary.Targets,
		GeneratedAt: svc.clock()(),
	}, nil
}

func (svc *collectService) GetTargetStatus(siteID int64, target string) (collectmodels.TargetCollectStatus, common.GFError) {
	target = strings.TrimSpace(target)
	if siteID <= 0 {
		return collectmodels.TargetCollectStatus{}, common.NewServiceError("siteId 参数非法")
	}
	if target == "" {
		return collectmodels.TargetCollectStatus{}, common.NewServiceError("target 参数不能为空")
	}
	summary, err := summaryservice.GetSummaryService().GetTargetSummary(siteID, target)
	if err != nil {
		return collectmodels.TargetCollectStatus{}, err
	}
	core, err := svc.readModels().GetTargetLatest(siteID, target, readmodels.CoreProtocols())
	if err != nil {
		return collectmodels.TargetCollectStatus{}, err
	}
	light, err := svc.readModels().GetLightProbeLatest(siteID, target)
	if err != nil {
		return collectmodels.TargetCollectStatus{}, err
	}
	trend, err := svc.readModels().GetTargetTrend(siteID, target)
	if err != nil {
		return collectmodels.TargetCollectStatus{}, err
	}
	changes, err := svc.readModels().GetTargetChanges(siteID, target)
	if err != nil {
		return collectmodels.TargetCollectStatus{}, err
	}
	return collectmodels.TargetCollectStatus{
		SiteID:      siteID,
		Target:      target,
		Summary:     summary,
		LatestCore:  core,
		LatestLight: light,
		Trend:       trend,
		Changes:     changes,
		GeneratedAt: svc.clock()(),
	}, nil
}

func (svc *collectService) storeDAO() collectStore {
	if svc != nil && svc.store != nil {
		return svc.store
	}
	return collectdao.GetCollectDao()
}

func (svc *collectService) readModels() runStateReader {
	if svc != nil && svc.read != nil {
		return svc.read
	}
	return readservice.GetReadModelService()
}

func (svc *collectService) clock() func() time.Time {
	if svc != nil && svc.now != nil {
		return svc.now
	}
	return time.Now
}
