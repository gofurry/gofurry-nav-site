package service

import (
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/summary/models"
	"github.com/gofurry/gofurry-nav-backend/common"
)

func TestSummaryKeys(t *testing.T) {
	if got := SiteSummaryKey(123); got != "collector:v2:summary:site:123" {
		t.Fatalf("SiteSummaryKey() = %q", got)
	}
	if got := TargetSummaryKey(123, "example.com"); got != "collector:v2:summary:target:123:example.com" {
		t.Fatalf("TargetSummaryKey() = %q", got)
	}
}

func TestReadSiteSummaryMissing(t *testing.T) {
	summary, err := readSiteSummary(func(string) (string, common.GFError) {
		return "", nil
	}, 123, time.Hour, time.Now())
	if err != nil {
		t.Fatalf("readSiteSummary() error = %v", err)
	}
	if summary.State != models.SummaryStateMissing || summary.Status != models.StatusUnknown {
		t.Fatalf("summary = %+v", summary)
	}
	if len(summary.ReasonCodes) != 1 || summary.ReasonCodes[0] != "summary_missing" {
		t.Fatalf("reason_codes = %v", summary.ReasonCodes)
	}
}

func TestReadTargetSummaryStale(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	raw, _ := sonic.MarshalString(models.TargetSummaryResponse{
		SiteID:      123,
		Target:      "example.com",
		Status:      models.StatusHealthy,
		GeneratedAt: now.Add(-2 * time.Hour),
	})
	summary, err := readTargetSummary(func(string) (string, common.GFError) {
		return raw, nil
	}, 123, "example.com", time.Hour, now)
	if err != nil {
		t.Fatalf("readTargetSummary() error = %v", err)
	}
	if summary.State != models.SummaryStateStale || summary.Status != models.StatusUnknown {
		t.Fatalf("summary = %+v", summary)
	}
	if len(summary.ReasonCodes) != 1 || summary.ReasonCodes[0] != "summary_stale" {
		t.Fatalf("reason_codes = %v", summary.ReasonCodes)
	}
}

func TestReadTargetSummaryReady(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	raw, _ := sonic.MarshalString(models.TargetSummaryResponse{
		SiteID:      123,
		Target:      "example.com",
		Status:      models.StatusHealthy,
		GeneratedAt: now,
	})
	summary, err := readTargetSummary(func(string) (string, common.GFError) {
		return raw, nil
	}, 123, "example.com", time.Hour, now)
	if err != nil {
		t.Fatalf("readTargetSummary() error = %v", err)
	}
	if summary.State != models.SummaryStateReady || summary.Status != models.StatusHealthy {
		t.Fatalf("summary = %+v", summary)
	}
}

func TestReadSiteSummaryDecodeFailure(t *testing.T) {
	_, err := readSiteSummary(func(string) (string, common.GFError) {
		return "{bad-json", nil
	}, 123, time.Hour, time.Now())
	if err == nil {
		t.Fatal("readSiteSummary() expected decode error")
	}
}

func TestReadSummaryInvalidParams(t *testing.T) {
	if _, err := readSiteSummary(nil, 0, time.Hour, time.Now()); err == nil {
		t.Fatal("readSiteSummary() expected invalid siteID error")
	}
	if _, err := readTargetSummary(nil, 1, "", time.Hour, time.Now()); err == nil {
		t.Fatal("readTargetSummary() expected empty target error")
	}
}
