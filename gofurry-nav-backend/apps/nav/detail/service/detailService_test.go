package service

import (
	"encoding/json"
	"testing"
	"time"

	detailmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/detail/models"
	navmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
	readmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/readmodel/models"
	summarymodels "github.com/gofurry/gofurry-nav-backend/apps/nav/summary/models"
	"github.com/gofurry/gofurry-nav-backend/common"
)

func TestGetSiteDetailAggregatesReadModelAndSummary(t *testing.T) {
	now := time.Date(2026, 5, 27, 10, 0, 0, 0, time.UTC)
	store := &fakeSiteStore{
		site: navmodels.GfnSite{
			ID:        7,
			Name:      "中文站点",
			NameEn:    "English Site",
			Info:      "中文介绍",
			InfoEn:    "English intro",
			ViewCount: 42,
		},
		targets: []detailmodels.CollectorDomain{
			{ID: 11, SiteID: 7, Name: "example.com", TLS: "1", Proxy: "0"},
			{ID: 12, SiteID: 7, Name: "example.com", Prefix: strPtr("www."), TLS: "1", Proxy: "1"},
		},
	}
	summaries := &fakeSummaryReader{
		site: summarymodels.SiteSummaryResponse{
			State:  summarymodels.SummaryStateReady,
			SiteID: 7,
			Status: summarymodels.StatusWarning,
			Targets: []summarymodels.TargetSummaryItem{
				{Target: "example.com", Status: summarymodels.StatusHealthy},
				{Target: "www.example.com", Status: summarymodels.StatusWarning},
			},
		},
		target: summarymodels.TargetSummaryResponse{
			State:     summarymodels.SummaryStateReady,
			SiteID:    7,
			Target:    "example.com",
			Status:    summarymodels.StatusHealthy,
			Protocols: map[string]summarymodels.ProtocolSummary{},
		},
	}
	readModel := &fakeReadModelReader{
		latest: readmodels.TargetLatestResponse{
			State:  readmodels.SummaryStateReady,
			SiteID: 7,
			Target: "example.com",
			Protocols: map[string]readmodels.CollectorEnvelope{
				readmodels.ProtocolHTTP: {SiteID: 7, Target: "example.com", Protocol: readmodels.ProtocolHTTP},
			},
		},
		light: readmodels.TargetLatestResponse{
			State:  readmodels.SummaryStateReady,
			SiteID: 7,
			Target: "example.com",
			Protocols: map[string]readmodels.CollectorEnvelope{
				readmodels.ProtocolRDAP: {SiteID: 7, Target: "example.com", Protocol: readmodels.ProtocolRDAP},
			},
		},
		trend: readmodels.TargetTrendResponse{
			State:         readmodels.SummaryStateReady,
			SiteID:        7,
			Target:        "example.com",
			Windows:       json.RawMessage(`{"24h":{"protocols":{"http":{"success_rate":1}}}}`),
			SchemaVersion: 9,
		},
		changes: readmodels.TargetChangesResponse{
			State:         readmodels.SummaryStateReady,
			SiteID:        7,
			Target:        "example.com",
			Events:        json.RawMessage(`[{"event_id":"evt-1"}]`),
			SchemaVersion: 8,
		},
	}
	svc := newDetailService(store, summaries, readModel, func() time.Time { return now })

	response, err := svc.GetSiteDetail(7, "en", "")
	if err != nil {
		t.Fatalf("GetSiteDetail() error = %v", err)
	}
	if response.Site.Name != "English Site" || response.Site.Info != "English intro" {
		t.Fatalf("site localization failed: %+v", response.Site)
	}
	if response.SelectedTarget != "example.com" {
		t.Fatalf("SelectedTarget = %q", response.SelectedTarget)
	}
	if len(response.Targets) != 2 {
		t.Fatalf("Targets len = %d", len(response.Targets))
	}
	if response.Targets[0].SummaryState != summarymodels.SummaryStateReady || response.Targets[0].Status != summarymodels.StatusHealthy {
		t.Fatalf("first target summary state = %+v", response.Targets[0])
	}
	if response.LatestCore.SchemaVersion != detailmodels.DetailSchemaVersion || response.LightProbeState.SchemaVersion != detailmodels.DetailSchemaVersion {
		t.Fatalf("latest schema version mismatch: %+v %+v", response.LatestCore, response.LightProbeState)
	}
	if response.Derived.Trend.SchemaVersion != 9 || response.Derived.Changes.SchemaVersion != 8 {
		t.Fatalf("derived payload mismatch: %+v", response.Derived)
	}
	if len(readModel.latestProtocols) != len(readmodels.CoreProtocols()) {
		t.Fatalf("latest protocols = %v", readModel.latestProtocols)
	}
}

func TestGetSiteDetailKeepsMissingSemantics(t *testing.T) {
	store := &fakeSiteStore{
		site: navmodels.GfnSite{ID: 9, Name: "site"},
		targets: []detailmodels.CollectorDomain{
			{ID: 19, SiteID: 9, Name: "example.com"},
		},
	}
	summaries := &fakeSummaryReader{
		site: summarymodels.SiteSummaryResponse{
			State:  summarymodels.SummaryStateMissing,
			SiteID: 9,
			Status: summarymodels.StatusUnknown,
		},
		target: summarymodels.TargetSummaryResponse{
			State:  summarymodels.SummaryStateMissing,
			SiteID: 9,
			Target: "example.com",
			Status: summarymodels.StatusUnknown,
		},
	}
	readModel := &fakeReadModelReader{
		latest:  readmodels.TargetLatestResponse{State: readmodels.SummaryStateMissing, SiteID: 9, Target: "example.com", Protocols: map[string]readmodels.CollectorEnvelope{}},
		light:   readmodels.TargetLatestResponse{State: readmodels.SummaryStateMissing, SiteID: 9, Target: "example.com", Protocols: map[string]readmodels.CollectorEnvelope{}},
		trend:   readmodels.TargetTrendResponse{State: readmodels.SummaryStateMissing, SiteID: 9, Target: "example.com"},
		changes: readmodels.TargetChangesResponse{State: readmodels.SummaryStateMissing, SiteID: 9, Target: "example.com"},
	}
	svc := newDetailService(store, summaries, readModel, time.Now)

	response, err := svc.GetSiteDetail(9, "zh", "")
	if err != nil {
		t.Fatalf("GetSiteDetail() error = %v", err)
	}
	if response.SiteSummary.State != summarymodels.SummaryStateMissing || response.TargetSummary.State != summarymodels.SummaryStateMissing {
		t.Fatalf("summary states = %+v %+v", response.SiteSummary, response.TargetSummary)
	}
	if response.LatestCore.State != summarymodels.SummaryStateMissing || response.LightProbeState.State != summarymodels.SummaryStateMissing {
		t.Fatalf("latest states = %+v %+v", response.LatestCore, response.LightProbeState)
	}
	if string(response.Derived.Trend.Windows) != "{}" || string(response.Derived.Changes.Events) != "[]" {
		t.Fatalf("derived missing payload mismatch: %+v", response.Derived)
	}
}

func TestGetSiteDetailRejectsTargetOutsideSite(t *testing.T) {
	svc := newDetailService(&fakeSiteStore{
		site: navmodels.GfnSite{ID: 1, Name: "site"},
		targets: []detailmodels.CollectorDomain{
			{ID: 1, SiteID: 1, Name: "example.com"},
		},
	}, &fakeSummaryReader{
		site: summarymodels.SiteSummaryResponse{
			State:  summarymodels.SummaryStateMissing,
			SiteID: 1,
			Status: summarymodels.StatusUnknown,
		},
	}, &fakeReadModelReader{}, time.Now)

	_, err := svc.GetSiteDetail(1, "zh", "other.example.com")
	if err == nil || err.GetMsg() != "target 不属于当前 site" {
		t.Fatalf("GetSiteDetail() err = %v", err)
	}
}

func TestGetTargetLatestWrapsAllProtocols(t *testing.T) {
	now := time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
	readModel := &fakeReadModelReader{
		latest: readmodels.TargetLatestResponse{
			State:  readmodels.SummaryStateReady,
			SiteID: 5,
			Target: "example.com",
			Protocols: map[string]readmodels.CollectorEnvelope{
				readmodels.ProtocolHTTP: {Protocol: readmodels.ProtocolHTTP},
				readmodels.ProtocolRDAP: {Protocol: readmodels.ProtocolRDAP},
			},
		},
	}
	svc := newDetailService(&fakeSiteStore{
		site: navmodels.GfnSite{ID: 5, Name: "site"},
		targets: []detailmodels.CollectorDomain{
			{ID: 5, SiteID: 5, Name: "example.com"},
		},
	}, &fakeSummaryReader{
		site: summarymodels.SiteSummaryResponse{
			State:  summarymodels.SummaryStateMissing,
			SiteID: 5,
			Status: summarymodels.StatusUnknown,
		},
	}, readModel, func() time.Time { return now })

	response, err := svc.GetTargetLatest(5, "example.com")
	if err != nil {
		t.Fatalf("GetTargetLatest() error = %v", err)
	}
	if response.GeneratedAt != now || response.SchemaVersion != detailmodels.DetailSchemaVersion {
		t.Fatalf("latest wrapper mismatch: %+v", response)
	}
	if len(readModel.latestProtocols) != len(readmodels.AllProtocols()) {
		t.Fatalf("latest protocols = %v", readModel.latestProtocols)
	}
}

func TestListTargetObservationsUsesNormalizedLimit(t *testing.T) {
	now := time.Date(2026, 5, 27, 12, 30, 0, 0, time.UTC)
	readModel := &fakeReadModelReader{
		observations: readmodels.ObservationsResponse{
			State:    readmodels.SummaryStateReady,
			SiteID:   8,
			Target:   "example.com",
			Protocol: readmodels.ProtocolHTTP,
			Limit:    readmodels.MaxObservationLimit,
			Items: []readmodels.CollectorEnvelope{
				{Protocol: readmodels.ProtocolHTTP},
			},
		},
	}
	svc := newDetailService(&fakeSiteStore{
		site: navmodels.GfnSite{ID: 8, Name: "site"},
		targets: []detailmodels.CollectorDomain{
			{ID: 8, SiteID: 8, Name: "example.com"},
		},
	}, &fakeSummaryReader{
		site: summarymodels.SiteSummaryResponse{
			State:  summarymodels.SummaryStateMissing,
			SiteID: 8,
			Status: summarymodels.StatusUnknown,
		},
	}, readModel, func() time.Time { return now })

	response, err := svc.ListTargetObservations(8, "example.com", readmodels.ProtocolHTTP, 999)
	if err != nil {
		t.Fatalf("ListTargetObservations() error = %v", err)
	}
	if response.GeneratedAt != now || response.Limit != readmodels.MaxObservationLimit {
		t.Fatalf("observations wrapper mismatch: %+v", response)
	}
}

func TestEnsureSiteTargetTranslatesSiteMissing(t *testing.T) {
	svc := newDetailService(&fakeSiteStore{
		siteErr: common.NewDaoError("404"),
	}, &fakeSummaryReader{}, &fakeReadModelReader{}, time.Now)

	_, err := svc.GetTargetTrend(404, "example.com")
	if err == nil || err.GetMsg() != "站点不存在" {
		t.Fatalf("GetTargetTrend() err = %v", err)
	}
}

type fakeSiteStore struct {
	site    navmodels.GfnSite
	siteErr common.GFError
	targets []detailmodels.CollectorDomain
}

func (f *fakeSiteStore) GetSiteByID(siteID int64) (navmodels.GfnSite, common.GFError) {
	if f.siteErr != nil {
		return navmodels.GfnSite{}, f.siteErr
	}
	return f.site, nil
}

func (f *fakeSiteStore) ListCollectorDomains(siteID int64) ([]detailmodels.CollectorDomain, common.GFError) {
	return append([]detailmodels.CollectorDomain(nil), f.targets...), nil
}

type fakeSummaryReader struct {
	site      summarymodels.SiteSummaryResponse
	siteErr   common.GFError
	target    summarymodels.TargetSummaryResponse
	targetErr common.GFError
}

func (f *fakeSummaryReader) GetSiteSummary(siteID int64) (summarymodels.SiteSummaryResponse, common.GFError) {
	if f.siteErr != nil {
		return summarymodels.SiteSummaryResponse{}, f.siteErr
	}
	return f.site, nil
}

func (f *fakeSummaryReader) GetTargetSummary(siteID int64, target string) (summarymodels.TargetSummaryResponse, common.GFError) {
	if f.targetErr != nil {
		return summarymodels.TargetSummaryResponse{}, f.targetErr
	}
	return f.target, nil
}

type fakeReadModelReader struct {
	latest          readmodels.TargetLatestResponse
	latestErr       common.GFError
	light           readmodels.TargetLatestResponse
	lightErr        common.GFError
	observations    readmodels.ObservationsResponse
	observationsErr common.GFError
	trend           readmodels.TargetTrendResponse
	trendErr        common.GFError
	changes         readmodels.TargetChangesResponse
	changesErr      common.GFError
	latestProtocols []string
}

func (f *fakeReadModelReader) GetTargetLatest(siteID int64, target string, protocols []string) (readmodels.TargetLatestResponse, common.GFError) {
	f.latestProtocols = append([]string(nil), protocols...)
	if f.latestErr != nil {
		return readmodels.TargetLatestResponse{}, f.latestErr
	}
	return f.latest, nil
}

func (f *fakeReadModelReader) GetLightProbeLatest(siteID int64, target string) (readmodels.TargetLatestResponse, common.GFError) {
	if f.lightErr != nil {
		return readmodels.TargetLatestResponse{}, f.lightErr
	}
	return f.light, nil
}

func (f *fakeReadModelReader) ListObservations(siteID int64, target string, protocol string, limit int) (readmodels.ObservationsResponse, common.GFError) {
	if f.observationsErr != nil {
		return readmodels.ObservationsResponse{}, f.observationsErr
	}
	response := f.observations
	response.Limit = readmodels.NormalizeObservationLimit(limit)
	return response, nil
}

func (f *fakeReadModelReader) GetTargetTrend(siteID int64, target string) (readmodels.TargetTrendResponse, common.GFError) {
	if f.trendErr != nil {
		return readmodels.TargetTrendResponse{}, f.trendErr
	}
	return f.trend, nil
}

func (f *fakeReadModelReader) GetTargetChanges(siteID int64, target string) (readmodels.TargetChangesResponse, common.GFError) {
	if f.changesErr != nil {
		return readmodels.TargetChangesResponse{}, f.changesErr
	}
	return f.changes, nil
}

func strPtr(value string) *string {
	return &value
}
