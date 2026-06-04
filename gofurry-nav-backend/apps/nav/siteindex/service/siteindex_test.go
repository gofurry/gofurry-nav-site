package service

import (
	"testing"
	"time"

	navmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
	"github.com/gofurry/gofurry-nav-backend/common"
	cm "github.com/gofurry/gofurry-nav-backend/common/models"
)

func TestGetSiteIndexBuildsReadyResponse(t *testing.T) {
	now := time.Date(2026, 6, 4, 15, 0, 0, 0, time.UTC)
	svc := newSiteIndexService(&fakeSiteIndexStore{
		records: []navmodels.GfnSiteIndex{
			{ID: 1, Domain: `{"domain":["a.example.com","b.example.com"]}`, UpdateTime: cm.LocalTime(now)},
		},
	}, func() time.Time { return now })

	response := svc.GetSiteIndex()
	if response.State != "ready" || len(response.Items) != 1 {
		t.Fatalf("response = %#v", response)
	}
	if len(response.Items[0].Domains) != 2 || response.Items[0].Domains[0] != "a.example.com" {
		t.Fatalf("domains = %v", response.Items[0].Domains)
	}
}

func TestParseDomainsFallsBackToRawString(t *testing.T) {
	items := parseDomains("example.com")
	if len(items) != 1 || items[0] != "example.com" {
		t.Fatalf("items = %v", items)
	}
}

type fakeSiteIndexStore struct {
	records []navmodels.GfnSiteIndex
	err     common.GFError
}

func (store *fakeSiteIndexStore) GetSiteIndexList() ([]navmodels.GfnSiteIndex, common.GFError) {
	return store.records, store.err
}
