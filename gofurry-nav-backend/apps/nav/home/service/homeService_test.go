package service

import (
	"errors"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/home/models"
	navmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
	"github.com/gofurry/gofurry-nav-backend/common"
)

func TestGetHomeAggregatesNavPageData(t *testing.T) {
	now := time.Date(2026, 6, 3, 12, 0, 0, 0, time.UTC)
	reader := &fakeHomeReader{
		sites: []navmodels.SiteVo{{ID: "1", Name: "GoFurry"}},
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
	if len(response.Sites) != 1 || len(response.Groups) != 1 || len(response.Ping) != 1 {
		t.Fatalf("home data was not aggregated: %#v", response)
	}
	if len(response.Groups[0].Sites) != 1 || response.Groups[0].Sites[0].ID != "1" {
		t.Fatalf("group sites were not expanded: %#v", response.Groups)
	}
	if response.Saying == nil || response.Saying.Content != "hello" {
		t.Fatalf("missing saying: %#v", response.Saying)
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

	if len(response.Sites) != 1 {
		t.Fatalf("expected sites to stay available: %#v", response.Sites)
	}
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

func TestBuildHomeGroupsPreservesGroupSiteOrder(t *testing.T) {
	sites := []navmodels.SiteVo{
		{ID: "1", Name: "A"},
		{ID: "2", Name: "B"},
		{ID: "3", Name: "C"},
	}
	groups := []navmodels.GroupVo{
		{ID: "10", Name: "Forums", Priority: 1, Sites: []string{"3", "1", "2", "404"}},
	}

	result := buildHomeGroups(sites, groups)
	if len(result) != 1 {
		t.Fatalf("group count = %d", len(result))
	}
	if got := []string{result[0].Sites[0].ID, result[0].Sites[1].ID, result[0].Sites[2].ID}; got[0] != "3" || got[1] != "1" || got[2] != "2" {
		t.Fatalf("unexpected expanded site order: %#v", result[0].Sites)
	}
}

type fakeHomeReader struct {
	sites      []navmodels.SiteVo
	siteErr    common.GFError
	groups     []navmodels.GroupVo
	groupErr   common.GFError
	ping       map[string]string
	pingErr    common.GFError
	saying     navmodels.SayingModel
	sayingErr  common.GFError
	desktopURL string
	mobileURL  string
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

func (f *fakeHomeReader) GetPingList() (map[string]string, common.GFError) {
	if f.pingErr != nil {
		return nil, f.pingErr
	}
	if f.ping == nil {
		return map[string]string{}, nil
	}
	return f.ping, nil
}

func (f *fakeHomeReader) GetSayingService() (navmodels.SayingModel, common.GFError) {
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
