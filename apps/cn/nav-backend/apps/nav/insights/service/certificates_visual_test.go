package service

import (
	"context"
	"encoding/json"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/insights/models"
	"strings"
	"testing"
	"time"
)

func TestCertificateListsOptionalVisualKeepsEvidence(t *testing.T) {
	reference := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)
	expiry, verified, issue := reference.Add(48*time.Hour), false, "hostname_mismatch"
	records := []models.CertificateItemRecord{
		{SiteID: 1, SiteName: "Icon", VisualAsset: " site-logo.png ", Target: "example.test:443", NotAfter: &expiry, Verified: &verified, VerificationIssue: &issue, ObservedAt: &reference},
		{SiteID: 2, SiteName: "Missing", VisualAsset: " "},
	}
	store := &fakeStore{certificateSummary: &models.CertificateOverviewRecord{FactDate: reference, ReferenceAt: reference}, certificateAttention: records, certificateIssues: records}
	got, err := New(store).GetCertificateOverview(context.Background(), 20)
	if err != nil {
		t.Fatal(err)
	}
	for _, items := range [][]models.CertificateItem{got.ExpiryAttention, got.VerificationIssues} {
		first := items[0]
		if first.Site.Visual == nil || first.Site.Visual.Kind != "site_icon" || first.Site.Visual.Asset != "site-logo.png" {
			t.Fatalf("icon reference = %#v", first.Site)
		}
		if first.Target != records[0].Target || first.NotAfter != &expiry || *first.DaysToExpiry != 2 || *first.ExpiryStatus != "expires_within_7d" || *first.Verified || *first.VerificationIssue != issue || first.ObservedAt != &reference {
			t.Fatalf("certificate evidence changed: %#v", first)
		}
		encoded, err := json.Marshal(items[1].Site)
		if err != nil || strings.Contains(string(encoded), "visual") {
			t.Fatalf("absent icon must be omitted: %s", encoded)
		}
	}
}
