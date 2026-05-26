package service

import (
	"encoding/json"
	"strings"
	"time"

	detaildao "github.com/gofurry/gofurry-nav-backend/apps/nav/detail/dao"
	detailmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/detail/models"
	navmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
	readmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/readmodel/models"
	readservice "github.com/gofurry/gofurry-nav-backend/apps/nav/readmodel/service"
	summarymodels "github.com/gofurry/gofurry-nav-backend/apps/nav/summary/models"
	summaryservice "github.com/gofurry/gofurry-nav-backend/apps/nav/summary/service"
	"github.com/gofurry/gofurry-nav-backend/common"
)

type siteStore interface {
	GetSiteByID(siteID int64) (navmodels.GfnSite, common.GFError)
	ListCollectorDomains(siteID int64) ([]detailmodels.CollectorDomain, common.GFError)
}

type summaryReader interface {
	GetSiteSummary(siteID int64) (summarymodels.SiteSummaryResponse, common.GFError)
	GetTargetSummary(siteID int64, target string) (summarymodels.TargetSummaryResponse, common.GFError)
}

type readModelReader interface {
	GetTargetLatest(siteID int64, target string, protocols []string) (readmodels.TargetLatestResponse, common.GFError)
	GetLightProbeLatest(siteID int64, target string) (readmodels.TargetLatestResponse, common.GFError)
	ListObservations(siteID int64, target string, protocol string, limit int) (readmodels.ObservationsResponse, common.GFError)
	GetTargetTrend(siteID int64, target string) (readmodels.TargetTrendResponse, common.GFError)
	GetTargetChanges(siteID int64, target string) (readmodels.TargetChangesResponse, common.GFError)
}

type detailService struct {
	sites     siteStore
	summaries summaryReader
	readModel readModelReader
	now       func() time.Time
}

var detailSingleton = &detailService{}

func GetDetailService() *detailService {
	if detailSingleton.sites == nil {
		detailSingleton.sites = detaildao.GetDetailDao()
	}
	if detailSingleton.summaries == nil {
		detailSingleton.summaries = summaryservice.GetSummaryService()
	}
	if detailSingleton.readModel == nil {
		detailSingleton.readModel = readservice.GetReadModelService()
	}
	if detailSingleton.now == nil {
		detailSingleton.now = time.Now
	}
	return detailSingleton
}

func newDetailService(sites siteStore, summaries summaryReader, readModel readModelReader, now func() time.Time) *detailService {
	return &detailService{
		sites:     sites,
		summaries: summaries,
		readModel: readModel,
		now:       now,
	}
}

func (svc *detailService) GetSiteDetail(siteID int64, lang string, target string) (detailmodels.SiteDetailResponse, common.GFError) {
	site, targets, siteSummary, err := svc.loadSiteContext(siteID)
	if err != nil {
		return detailmodels.SiteDetailResponse{}, err
	}

	selectedTarget, targetIndexErr := selectTarget(targets, strings.TrimSpace(target))
	if targetIndexErr != nil {
		return detailmodels.SiteDetailResponse{}, targetIndexErr
	}

	response := detailmodels.SiteDetailResponse{
		Site:           buildSiteInfo(site, lang),
		Targets:        targets,
		SelectedTarget: selectedTarget,
		SiteSummary:    siteSummary,
		GeneratedAt:    svc.clock()(),
		SchemaVersion:  detailmodels.DetailSchemaVersion,
	}

	if selectedTarget == "" {
		response.TargetSummary = missingTargetSummary(siteID, selectedTarget, "站点未配置可用 target")
		response.LatestCore = missingLatest(siteID, selectedTarget)
		response.Derived = detailmodels.DerivedState{
			Trend:   missingTrend(siteID, selectedTarget),
			Changes: missingChanges(siteID, selectedTarget),
		}
		response.LightProbeState = missingLatest(siteID, selectedTarget)
		return response, nil
	}

	targetSummary, err := svc.summaryService().GetTargetSummary(siteID, selectedTarget)
	if err != nil {
		return detailmodels.SiteDetailResponse{}, err
	}
	updateSelectedTargetSummary(targets, selectedTarget, targetSummary)

	latestCore, err := svc.readModels().GetTargetLatest(siteID, selectedTarget, readmodels.CoreProtocols())
	if err != nil {
		return detailmodels.SiteDetailResponse{}, err
	}
	lightProbeState, err := svc.readModels().GetLightProbeLatest(siteID, selectedTarget)
	if err != nil {
		return detailmodels.SiteDetailResponse{}, err
	}
	trend, err := svc.readModels().GetTargetTrend(siteID, selectedTarget)
	if err != nil {
		return detailmodels.SiteDetailResponse{}, err
	}
	changes, err := svc.readModels().GetTargetChanges(siteID, selectedTarget)
	if err != nil {
		return detailmodels.SiteDetailResponse{}, err
	}

	response.TargetSummary = normalizeTargetSummary(targetSummary)
	response.LatestCore = toLatestResponse(latestCore, response.GeneratedAt)
	response.Derived = detailmodels.DerivedState{
		Trend:   normalizeTrend(trend),
		Changes: normalizeChanges(changes),
	}
	response.LightProbeState = toLatestResponse(lightProbeState, response.GeneratedAt)
	return response, nil
}

func (svc *detailService) GetTargetLatest(siteID int64, target string) (detailmodels.TargetLatestResponse, common.GFError) {
	target, err := svc.ensureSiteTarget(siteID, target)
	if err != nil {
		return detailmodels.TargetLatestResponse{}, err
	}
	response, err := svc.readModels().GetTargetLatest(siteID, target, readmodels.AllProtocols())
	if err != nil {
		return detailmodels.TargetLatestResponse{}, err
	}
	return toLatestResponse(response, svc.clock()()), nil
}

func (svc *detailService) ListTargetObservations(siteID int64, target string, protocol string, limit int) (detailmodels.TargetObservationsResponse, common.GFError) {
	target, err := svc.ensureSiteTarget(siteID, target)
	if err != nil {
		return detailmodels.TargetObservationsResponse{}, err
	}
	response, err := svc.readModels().ListObservations(siteID, target, protocol, limit)
	if err != nil {
		return detailmodels.TargetObservationsResponse{}, err
	}
	return detailmodels.TargetObservationsResponse{
		State:         response.State,
		SiteID:        response.SiteID,
		Target:        response.Target,
		Protocol:      response.Protocol,
		Limit:         response.Limit,
		Items:         response.Items,
		GeneratedAt:   svc.clock()(),
		SchemaVersion: detailmodels.DetailSchemaVersion,
	}, nil
}

func (svc *detailService) GetTargetTrend(siteID int64, target string) (detailmodels.TargetTrendResponse, common.GFError) {
	target, err := svc.ensureSiteTarget(siteID, target)
	if err != nil {
		return detailmodels.TargetTrendResponse{}, err
	}
	response, err := svc.readModels().GetTargetTrend(siteID, target)
	if err != nil {
		return detailmodels.TargetTrendResponse{}, err
	}
	return normalizeTrend(response), nil
}

func (svc *detailService) GetTargetChanges(siteID int64, target string) (detailmodels.TargetChangesResponse, common.GFError) {
	target, err := svc.ensureSiteTarget(siteID, target)
	if err != nil {
		return detailmodels.TargetChangesResponse{}, err
	}
	response, err := svc.readModels().GetTargetChanges(siteID, target)
	if err != nil {
		return detailmodels.TargetChangesResponse{}, err
	}
	return normalizeChanges(response), nil
}

func (svc *detailService) GetTargetLightProbes(siteID int64, target string) (detailmodels.TargetLatestResponse, common.GFError) {
	target, err := svc.ensureSiteTarget(siteID, target)
	if err != nil {
		return detailmodels.TargetLatestResponse{}, err
	}
	response, err := svc.readModels().GetLightProbeLatest(siteID, target)
	if err != nil {
		return detailmodels.TargetLatestResponse{}, err
	}
	return toLatestResponse(response, svc.clock()()), nil
}

func (svc *detailService) loadSiteContext(siteID int64) (navmodels.GfnSite, []detailmodels.SiteTarget, summarymodels.SiteSummaryResponse, common.GFError) {
	site, err := svc.loadSite(siteID)
	if err != nil {
		return navmodels.GfnSite{}, nil, summarymodels.SiteSummaryResponse{}, err
	}
	domains, err := svc.siteStore().ListCollectorDomains(siteID)
	if err != nil {
		return navmodels.GfnSite{}, nil, summarymodels.SiteSummaryResponse{}, err
	}
	siteSummary, err := svc.summaryService().GetSiteSummary(siteID)
	if err != nil {
		return navmodels.GfnSite{}, nil, summarymodels.SiteSummaryResponse{}, err
	}
	return site, mergeTargets(domains, siteSummary), normalizeSiteSummary(siteSummary), nil
}

func (svc *detailService) ensureSiteTarget(siteID int64, target string) (string, common.GFError) {
	if siteID <= 0 {
		return "", common.NewServiceError("siteId 参数非法")
	}
	target = strings.TrimSpace(target)
	if target == "" {
		return "", common.NewServiceError("target 参数不能为空")
	}

	_, err := svc.loadSite(siteID)
	if err != nil {
		return "", err
	}
	domains, err := svc.siteStore().ListCollectorDomains(siteID)
	if err != nil {
		return "", err
	}
	for _, domain := range domains {
		if domain.TargetName() == target {
			return target, nil
		}
	}
	siteSummary, summaryErr := svc.summaryService().GetSiteSummary(siteID)
	if summaryErr == nil {
		for _, item := range siteSummary.Targets {
			if item.Target == target {
				return target, nil
			}
		}
	}
	return "", common.NewServiceError("target 不属于当前 site")
}

func (svc *detailService) loadSite(siteID int64) (navmodels.GfnSite, common.GFError) {
	if siteID <= 0 {
		return navmodels.GfnSite{}, common.NewServiceError("siteId 参数非法")
	}
	record, err := svc.siteStore().GetSiteByID(siteID)
	if err != nil {
		if err.GetMsg() == "404" {
			return navmodels.GfnSite{}, common.NewServiceError("站点不存在")
		}
		return navmodels.GfnSite{}, err
	}
	return record, nil
}

func (svc *detailService) siteStore() siteStore {
	if svc != nil && svc.sites != nil {
		return svc.sites
	}
	return detaildao.GetDetailDao()
}

func (svc *detailService) summaryService() summaryReader {
	if svc != nil && svc.summaries != nil {
		return svc.summaries
	}
	return summaryservice.GetSummaryService()
}

func (svc *detailService) readModels() readModelReader {
	if svc != nil && svc.readModel != nil {
		return svc.readModel
	}
	return readservice.GetReadModelService()
}

func (svc *detailService) clock() func() time.Time {
	if svc != nil && svc.now != nil {
		return svc.now
	}
	return time.Now
}

func buildSiteInfo(site navmodels.GfnSite, lang string) detailmodels.SiteInfo {
	name := site.Name
	info := site.Info
	if strings.EqualFold(strings.TrimSpace(lang), "en") {
		name = site.NameEn
		info = site.InfoEn
	}
	return detailmodels.SiteInfo{
		ID:        site.ID,
		Name:      name,
		Info:      info,
		Icon:      site.Icon,
		Country:   site.Country,
		Nsfw:      site.Nsfw,
		Welfare:   site.Welfare,
		ViewCount: site.ViewCount,
	}
}

func mergeTargets(domains []detailmodels.CollectorDomain, siteSummary summarymodels.SiteSummaryResponse) []detailmodels.SiteTarget {
	items := make([]detailmodels.SiteTarget, 0, len(domains)+len(siteSummary.Targets))
	index := map[string]int{}
	for _, domain := range domains {
		target := domain.TargetName()
		if target == "" {
			continue
		}
		if _, exists := index[target]; exists {
			continue
		}
		index[target] = len(items)
		items = append(items, detailmodels.SiteTarget{
			Target:       target,
			DomainID:     domain.ID,
			Name:         domain.Name,
			Prefix:       domain.Prefix,
			TLS:          domain.TLS,
			Proxy:        domain.Proxy,
			SummaryState: summarymodels.SummaryStateMissing,
			Status:       summarymodels.StatusUnknown,
		})
	}
	for _, targetSummary := range siteSummary.Targets {
		idx, exists := index[targetSummary.Target]
		if !exists {
			index[targetSummary.Target] = len(items)
			items = append(items, detailmodels.SiteTarget{
				Target:       targetSummary.Target,
				Name:         targetSummary.Target,
				SummaryState: siteSummary.State,
				Status:       normalizeStatus(targetSummary.Status),
			})
			continue
		}
		items[idx].SummaryState = normalizeSummaryState(siteSummary.State)
		items[idx].Status = normalizeStatus(targetSummary.Status)
	}
	return items
}

func updateSelectedTargetSummary(targets []detailmodels.SiteTarget, selectedTarget string, summary summarymodels.TargetSummaryResponse) {
	for idx := range targets {
		if targets[idx].Target != selectedTarget {
			continue
		}
		targets[idx].SummaryState = normalizeSummaryState(summary.State)
		targets[idx].Status = normalizeStatus(summary.Status)
		return
	}
}

func selectTarget(targets []detailmodels.SiteTarget, preferred string) (string, common.GFError) {
	if preferred == "" {
		if len(targets) == 0 {
			return "", nil
		}
		return targets[0].Target, nil
	}
	for _, target := range targets {
		if target.Target == preferred {
			return preferred, nil
		}
	}
	return "", common.NewServiceError("target 不属于当前 site")
}

func toLatestResponse(response readmodels.TargetLatestResponse, generatedAt time.Time) detailmodels.TargetLatestResponse {
	return detailmodels.TargetLatestResponse{
		State:         normalizeSummaryState(response.State),
		SiteID:        response.SiteID,
		Target:        response.Target,
		Protocols:     response.Protocols,
		GeneratedAt:   generatedAt,
		SchemaVersion: detailmodels.DetailSchemaVersion,
	}
}

func normalizeTrend(response readmodels.TargetTrendResponse) detailmodels.TargetTrendResponse {
	if len(response.Windows) == 0 {
		response.Windows = json.RawMessage(`{}`)
	}
	if response.SchemaVersion == 0 {
		response.SchemaVersion = detailmodels.DetailSchemaVersion
	}
	return detailmodels.TargetTrendResponse{
		State:         normalizeSummaryState(response.State),
		SiteID:        response.SiteID,
		Target:        response.Target,
		Windows:       response.Windows,
		GeneratedAt:   response.GeneratedAt,
		SchemaVersion: response.SchemaVersion,
	}
}

func normalizeChanges(response readmodels.TargetChangesResponse) detailmodels.TargetChangesResponse {
	if len(response.Events) == 0 {
		response.Events = json.RawMessage(`[]`)
	}
	if response.SchemaVersion == 0 {
		response.SchemaVersion = detailmodels.DetailSchemaVersion
	}
	return detailmodels.TargetChangesResponse{
		State:         normalizeSummaryState(response.State),
		SiteID:        response.SiteID,
		Target:        response.Target,
		Events:        response.Events,
		GeneratedAt:   response.GeneratedAt,
		SchemaVersion: response.SchemaVersion,
	}
}

func normalizeSiteSummary(summary summarymodels.SiteSummaryResponse) summarymodels.SiteSummaryResponse {
	summary.State = normalizeSummaryState(summary.State)
	if summary.Status == "" {
		summary.Status = summarymodels.StatusUnknown
	}
	return summary
}

func normalizeTargetSummary(summary summarymodels.TargetSummaryResponse) summarymodels.TargetSummaryResponse {
	summary.State = normalizeSummaryState(summary.State)
	if summary.Status == "" {
		summary.Status = summarymodels.StatusUnknown
	}
	if summary.Protocols == nil {
		summary.Protocols = map[string]summarymodels.ProtocolSummary{}
	}
	return summary
}

func normalizeSummaryState(state string) string {
	switch state {
	case summarymodels.SummaryStateReady, summarymodels.SummaryStateStale, summarymodels.SummaryStateMissing:
		return state
	default:
		return summarymodels.SummaryStateMissing
	}
}

func normalizeStatus(status string) string {
	if status == "" {
		return summarymodels.StatusUnknown
	}
	return status
}

func missingTargetSummary(siteID int64, target string, message string) summarymodels.TargetSummaryResponse {
	if message == "" {
		message = "目标健康摘要暂不可用"
	}
	return summarymodels.TargetSummaryResponse{
		State:          summarymodels.SummaryStateMissing,
		SiteID:         siteID,
		Target:         target,
		Status:         summarymodels.StatusUnknown,
		ReasonCodes:    []string{"summary_missing"},
		ReasonMessages: []string{message},
		Protocols:      map[string]summarymodels.ProtocolSummary{},
	}
}

func missingLatest(siteID int64, target string) detailmodels.TargetLatestResponse {
	return detailmodels.TargetLatestResponse{
		State:         summarymodels.SummaryStateMissing,
		SiteID:        siteID,
		Target:        target,
		Protocols:     map[string]readmodels.CollectorEnvelope{},
		SchemaVersion: detailmodels.DetailSchemaVersion,
	}
}

func missingTrend(siteID int64, target string) detailmodels.TargetTrendResponse {
	return detailmodels.TargetTrendResponse{
		State:         summarymodels.SummaryStateMissing,
		SiteID:        siteID,
		Target:        target,
		Windows:       json.RawMessage(`{}`),
		SchemaVersion: detailmodels.DetailSchemaVersion,
	}
}

func missingChanges(siteID int64, target string) detailmodels.TargetChangesResponse {
	return detailmodels.TargetChangesResponse{
		State:         summarymodels.SummaryStateMissing,
		SiteID:        siteID,
		Target:        target,
		Events:        json.RawMessage(`[]`),
		SchemaVersion: detailmodels.DetailSchemaVersion,
	}
}
