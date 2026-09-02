package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/insights/models"
)

func TestSiteComparePreservesOrderDeduplicatesAndUsesOneCompleteSnapshot(t *testing.T) {
	horizon := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	unknownIssue := "raw x509 error"
	failed := false
	notAfter := horizon.AddDate(0, 0, 5)
	store := &fakeStore{
		sites: map[int64]*models.SiteRecord{
			2: {ID: 2, Name: "Second"},
			1: {ID: 1, Name: "First"},
		},
		compareHorizon: &horizon,
		compareCertificates: []models.CertificateItemRecord{
			{SiteID: 2, SiteName: "Second", Target: "second.example", NotAfter: &notAfter, Verified: &failed, VerificationIssue: &unknownIssue},
		},
	}
	for _, siteID := range []int64{2, 1} {
		for index, contract := range metricContracts {
			state := "positive"
			if index == 1 {
				state = "unknown"
			}
			store.compareCapabilities = append(store.compareCapabilities, models.SiteCompareCapabilityRecord{
				SiteID: siteID, MetricKey: contract.InternalKey, MetricVersion: contract.Version, State: state,
			})
		}
	}

	got, err := New(store).GetSiteCompare(context.Background(), "2,1,2")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "ready" || got.AsOf == nil || *got.AsOf != "2026-09-01" {
		t.Fatalf("snapshot = %#v", got)
	}
	if len(got.Sites) != 2 || got.Sites[0].Site.ID != 2 || got.Sites[1].Site.ID != 1 {
		t.Fatalf("order/dedup = %#v", got.Sites)
	}
	if got.Sites[0].Capabilities[1].State != "unknown" {
		t.Fatalf("unknown capability was changed: %#v", got.Sites[0].Capabilities)
	}
	if got.Sites[0].Certificate == nil || got.Sites[0].Certificate.VerificationIssue == nil || *got.Sites[0].Certificate.VerificationIssue != "other" {
		t.Fatalf("certificate semantics not reused: %#v", got.Sites[0].Certificate)
	}
	if got.Sites[1].Certificate != nil {
		t.Fatalf("invalid/missing certificate evidence should stay absent: %#v", got.Sites[1].Certificate)
	}
}

func TestSiteCompareValidationNotFoundAndInsufficientData(t *testing.T) {
	for _, ids := range []string{"", "1", "1,1", "1,bad", "1,2,3,4,5"} {
		if _, err := New(&fakeStore{}).GetSiteCompare(context.Background(), ids); !errors.Is(err, ErrInvalidCompare) {
			t.Fatalf("ids=%q err=%v", ids, err)
		}
	}
	store := &fakeStore{sites: map[int64]*models.SiteRecord{1: {ID: 1, Name: "First"}}}
	if _, err := New(store).GetSiteCompare(context.Background(), "1,2"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing site err=%v", err)
	}
	store.sites[2] = &models.SiteRecord{ID: 2, Name: "Second"}
	got, err := New(store).GetSiteCompare(context.Background(), "1,2")
	if err != nil || got.Status != "insufficient_data" || got.AsOf != nil || len(got.Sites) != 2 {
		t.Fatalf("insufficient snapshot = %#v err=%v", got, err)
	}
}
