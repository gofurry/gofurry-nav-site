package service

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/home/models"
	navmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
	"github.com/gofurry/gofurry-nav-backend/common"
)

func TestGetHomeAggregatesNavPageData(t *testing.T) {
	now := time.Date(2026, 6, 3, 12, 0, 0, 0, time.UTC)
	reader := &fakeHomeReader{
		sites: []navmodels.SiteVo{{ID: "1", Name: "GoFurry", ViewCount: 8, CreateTime: "2026-06-01 00:00:00"}},
		featured: []navmodels.FeaturedSiteVo{{
			ID:     "1",
			SiteID: "1",
			Weight: 10,
		}},
		groups: []navmodels.GroupVo{{
			ID:       "10",
			Name:     "Community",
			Priority: 1,
			Sites:    []string{"1"},
		}},
		ping:       map[string]string{"example.com": `{"status":"up"}`},
		saying:     navmodels.SayingModel{Content: "hello"},
		desktopURL: "desktop.avif",
		mobileURL:  "mobile.avif",
	}

	response := newHomeService(reader, func() time.Time { return now }).GetHome("en")

	if response.SchemaVersion != models.HomeSchemaVersion {
		t.Fatalf("unexpected schema version: %d", response.SchemaVersion)
	}
	if !response.GeneratedAt.Equal(now) {
		t.Fatalf("unexpected generated_at: %v", response.GeneratedAt)
	}
	if len(response.Groups) != 1 || len(response.Ping) != 1 {
		t.Fatalf("home data was not aggregated: %#v", response)
	}
	if len(response.Spotlight.Featured) != 1 || len(response.Spotlight.Popular) != 1 || response.Spotlight.PageSize != 6 {
		t.Fatalf("spotlight data was not aggregated: %#v", response.Spotlight)
	}
	if len(response.Groups[0].Sites) != 1 || response.Groups[0].Sites[0].ID != "1" {
		t.Fatalf("group sites were not expanded: %#v", response.Groups)
	}
	if response.Sites != nil {
		t.Fatalf("home response should not carry full sites payload: %#v", response.Sites)
	}
	if response.Saying == nil || response.Saying.Content != "hello" {
		t.Fatalf("missing saying: %#v", response.Saying)
	}
	if reader.sayingLang != "en" {
		t.Fatalf("saying lang = %q, want en", reader.sayingLang)
	}
	if response.Backgrounds.Desktop != "desktop.avif" || response.Backgrounds.Mobile != "mobile.avif" {
		t.Fatalf("unexpected backgrounds: %#v", response.Backgrounds)
	}
	if response.CacheState["sites"] != models.HomeStateReady || response.CacheState["backgrounds"] != models.HomeStateReady {
		t.Fatalf("unexpected cache state: %#v", response.CacheState)
	}
	if response.ReasonMessages != nil {
		t.Fatalf("unexpected reason messages: %#v", response.ReasonMessages)
	}
}

func TestGetHomeKeepsPartialDataWhenOptionalBlockFails(t *testing.T) {
	reader := &fakeHomeReader{
		sites:      []navmodels.SiteVo{{ID: "1"}},
		groupErr:   common.NewServiceError("groups unavailable"),
		pingErr:    common.NewServiceError("ping unavailable"),
		sayingErr:  common.NewServiceError("saying unavailable"),
		desktopURL: "desktop.avif",
	}

	response := newHomeService(reader, time.Now).GetHome("bad-lang")

	if response.CacheState["sites"] != models.HomeStateReady {
		t.Fatalf("expected sites ready, got %q", response.CacheState["sites"])
	}
	if response.CacheState["groups"] != models.HomeStateMissing || response.CacheState["ping"] != models.HomeStateMissing {
		t.Fatalf("unexpected cache state: %#v", response.CacheState)
	}
	if response.ReasonMessages["groups"] == "" || response.ReasonMessages["ping"] == "" {
		t.Fatalf("missing reason messages: %#v", response.ReasonMessages)
	}
}

func TestGetHomePingReportsMissingState(t *testing.T) {
	reader := &fakeHomeReader{pingErr: common.NewServiceError("redis unavailable")}

	response := newHomeService(reader, time.Now).GetHomePing()

	if response.State != models.HomeStateMissing {
		t.Fatalf("expected missing state, got %q", response.State)
	}
	if len(response.Ping) != 0 || len(response.ReasonMessages) != 1 {
		t.Fatalf("unexpected ping response: %#v", response)
	}
}

func TestBuildHomeGroupsOrdersSitesByGroupWeightThenUpdateTime(t *testing.T) {
	sites := []navmodels.SiteVo{
		{ID: "1", Name: "A", UpdateTime: "2026-06-01 00:00:00"},
		{ID: "2", Name: "B", UpdateTime: "2026-06-02 00:00:00"},
		{ID: "3", Name: "C", UpdateTime: "2026-06-03 00:00:00"},
	}
	groups := []navmodels.GroupVo{
		{
			ID:       "10",
			Name:     "Forums",
			Priority: 1,
			Sites:    []string{"3", "1", "2", "404"},
			SiteWeights: map[string]int64{
				"1": 10,
				"2": 20,
				"3": 20,
			},
		},
	}

	result := buildHomeGroups(sites, groups)
	if len(result) != 1 {
		t.Fatalf("group count = %d", len(result))
	}
	if got := []string{result[0].Sites[0].ID, result[0].Sites[1].ID, result[0].Sites[2].ID}; got[0] != "3" || got[1] != "2" || got[2] != "1" {
		t.Fatalf("unexpected expanded site order: %#v", result[0].Sites)
	}
	if result[0].SiteCount != 3 || result[0].DetailPath != "/site-groups/10" {
		t.Fatalf("unexpected group metadata: %#v", result[0])
	}
}

func TestBuildHomeGroupsLimitsPreviewToEight(t *testing.T) {
	sites := make([]navmodels.SiteVo, 0, 10)
	siteIDs := make([]string, 0, 10)
	for i := 1; i <= 10; i++ {
		id := strconv.Itoa(i)
		sites = append(sites, navmodels.SiteVo{ID: id, Name: "Site " + id})
		siteIDs = append(siteIDs, id)
	}

	result := buildHomeGroups(sites, []navmodels.GroupVo{{ID: "10", Name: "Forums", Sites: siteIDs}})
	if len(result) != 1 {
		t.Fatalf("group count = %d", len(result))
	}
	if len(result[0].Sites) != 8 {
		t.Fatalf("preview site count = %d, want 8", len(result[0].Sites))
	}
	if result[0].SiteCount != 10 || !result[0].HasMore {
		t.Fatalf("unexpected group metadata: %#v", result[0])
	}
}

func TestBuildHomeSpotlightOrdersSections(t *testing.T) {
	sites := []navmodels.SiteVo{
		{ID: "1", Name: "A", ViewCount: 10, CreateTime: "2026-06-01 00:00:00"},
		{ID: "2", Name: "B", ViewCount: 30, CreateTime: "2026-06-03 00:00:00"},
		{ID: "3", Name: "C", ViewCount: 20, CreateTime: "2026-06-02 00:00:00"},
	}
	featured := []navmodels.FeaturedSiteVo{
		{ID: "10", SiteID: "3", Weight: 30},
		{ID: "11", SiteID: "1", Weight: 10},
	}

	result := buildHomeSpotlight(sites, featured, time.Unix(1, 0))
	if got := []string{result.Featured[0].ID, result.Featured[1].ID}; got[0] != "3" || got[1] != "1" {
		t.Fatalf("unexpected featured order: %#v", result.Featured)
	}
	if result.Popular[0].ID != "2" || result.Latest[0].ID != "2" {
		t.Fatalf("unexpected popular/latest order: %#v %#v", result.Popular, result.Latest)
	}
	if len(result.Random) != len(sites) {
		t.Fatalf("random sites were not included: %#v", result.Random)
	}
}

type fakeHomeReader struct {
	sites       []navmodels.SiteVo
	siteErr     common.GFError
	groups      []navmodels.GroupVo
	groupErr    common.GFError
	featured    []navmodels.FeaturedSiteVo
	featuredErr common.GFError
	ping        map[string]string
	pingErr     common.GFError
	saying      navmodels.SayingModel
	sayingErr   common.GFError
	sayingLang  string
	desktopURL  string
	mobileURL   string
}

func (f *fakeHomeReader) GetSiteList(lang string) ([]navmodels.SiteVo, common.GFError) {
	if f.siteErr != nil {
		return nil, f.siteErr
	}
	if lang != "zh" && lang != "en" {
		return nil, common.NewServiceError("unexpected lang")
	}
	return f.sites, nil
}

func (f *fakeHomeReader) GetGroupList(string) ([]navmodels.GroupVo, common.GFError) {
	if f.groupErr != nil {
		return nil, f.groupErr
	}
	return f.groups, nil
}

func (f *fakeHomeReader) GetFeaturedSiteList() ([]navmodels.FeaturedSiteVo, common.GFError) {
	if f.featuredErr != nil {
		return nil, f.featuredErr
	}
	return f.featured, nil
}

func (f *fakeHomeReader) GetPingList() (map[string]string, common.GFError) {
	if f.pingErr != nil {
		return nil, f.pingErr
	}
	if f.ping == nil {
		return map[string]string{}, nil
	}
	return f.ping, nil
}

func (f *fakeHomeReader) GetSayingService(lang string) (navmodels.SayingModel, common.GFError) {
	f.sayingLang = lang
	if f.sayingErr != nil {
		return navmodels.SayingModel{}, f.sayingErr
	}
	if f.saying.Content == "" {
		return navmodels.SayingModel{}, common.NewServiceError(errors.New("missing saying").Error())
	}
	return f.saying, nil
}

func (f *fakeHomeReader) GetImageUrl(t string) string {
	if t == "standard" {
		return f.desktopURL
	}
	if t == "mobile" {
		return f.mobileURL
	}
	return ""
}
