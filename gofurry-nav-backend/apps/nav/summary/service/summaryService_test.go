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
		SiteID: 123,
		Target: "example.com",
		Status: models.StatusHealthy,
		CanonicalTarget: &models.CanonicalTargetHint{
			TargetHost: "example.com",
			FinalHost:  "www.example.com",
			Relation:   "redirect_to_www",
			Source:     "final_url",
		},
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
	if summary.CanonicalTarget == nil || summary.CanonicalTarget.FinalHost != "www.example.com" {
		t.Fatalf("stale summary dropped canonical hint: %+v", summary.CanonicalTarget)
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

func TestReadTargetSummaryPreservesHintsAndUnknownFields(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	raw := `{
		"site_id": 123,
		"target": "example.com",
		"status": "healthy",
		"reason_codes": ["http_ok"],
		"reason_messages": ["HTTP 可用"],
		"protocols": {},
		"canonical_target_hint": {
			"target_host": "example.com",
			"final_host": "www.example.com",
			"canonical_host": "www.example.com",
			"preferred_host": "www.example.com",
			"relation": "redirect_to_www",
			"source": "canonical_url",
			"final_url": "https://www.example.com/home",
			"canonical_url": "https://www.example.com/"
		},
		"target_relation_hints": [
			{
				"relation": "redirect_to_www",
				"source": "final_url",
				"target_host": "example.com",
				"related_host": "www.example.com",
				"value": "https://www.example.com/home"
			}
		],
		"edge_provider_hints": [
			{
				"provider": "cloudflare",
				"hint_type": "cdn",
				"confidence": "high",
				"evidence": [
					{"source": "http", "field": "headers.cf-ray", "value": "abc-HKG"}
				]
			}
		],
		"generated_at": "2026-05-24T10:00:00Z",
		"schema_version": 1,
		"collector_future_field": {"kept_by_collector": true}
	}`

	summary, err := readTargetSummary(func(string) (string, common.GFError) {
		return raw, nil
	}, 123, "example.com", time.Hour, now)
	if err != nil {
		t.Fatalf("readTargetSummary() error = %v", err)
	}
	if summary.State != models.SummaryStateReady || summary.Status != models.StatusHealthy {
		t.Fatalf("summary = %+v", summary)
	}
	if summary.CanonicalTarget == nil || summary.CanonicalTarget.CanonicalURL != "https://www.example.com/" {
		t.Fatalf("canonical_target_hint = %+v", summary.CanonicalTarget)
	}
	if len(summary.TargetRelations) != 1 || summary.TargetRelations[0].RelatedHost != "www.example.com" {
		t.Fatalf("target_relation_hints = %+v", summary.TargetRelations)
	}
	if len(summary.EdgeProviderHints) != 1 || len(summary.EdgeProviderHints[0].Evidence) != 1 {
		t.Fatalf("edge_provider_hints = %+v", summary.EdgeProviderHints)
	}
	if summary.EdgeProviderHints[0].Evidence[0].Value != "abc-HKG" {
		t.Fatalf("edge evidence = %+v", summary.EdgeProviderHints[0].Evidence)
	}
}

func TestReadSiteSummaryPreservesRelationAndTargetHints(t *testing.T) {
	now := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	raw, _ := sonic.MarshalString(models.SiteSummaryResponse{
		SiteID:       123,
		Status:       models.StatusHealthy,
		TargetCount:  1,
		StatusCounts: map[string]int{models.StatusHealthy: 1},
		Targets: []models.TargetSummaryItem{
			{
				Target: "example.com",
				Status: models.StatusHealthy,
				CanonicalTarget: &models.CanonicalTargetHint{
					TargetHost:    "example.com",
					FinalHost:     "www.example.com",
					CanonicalHost: "www.example.com",
					Relation:      "redirect_to_www",
					Source:        "final_url",
				},
				TargetRelations: []models.TargetRelationHint{
					{
						Relation:    "redirect_to_www",
						Source:      "final_url",
						TargetHost:  "example.com",
						RelatedHost: "www.example.com",
						Value:       "https://www.example.com/",
					},
				},
				EdgeProviderHints: []models.EdgeProviderHint{
					{
						Provider:   "cloudflare",
						HintType:   "cdn",
						Confidence: "high",
						Evidence: []models.EdgeProviderEvidence{
							{Source: "dns", Field: "asn", Value: "AS13335"},
						},
					},
				},
			},
		},
		TargetRelations: []models.SiteTargetRelationHint{
			{
				Relation: "duplicate_final_host",
				Host:     "www.example.com",
				Targets:  []string{"example.com", "www.example.com"},
			},
		},
		GeneratedAt:   now,
		SchemaVersion: 1,
	})

	summary, err := readSiteSummary(func(string) (string, common.GFError) {
		return raw, nil
	}, 123, time.Hour, now)
	if err != nil {
		t.Fatalf("readSiteSummary() error = %v", err)
	}
	if summary.State != models.SummaryStateReady || summary.Status != models.StatusHealthy {
		t.Fatalf("summary = %+v", summary)
	}
	if len(summary.TargetRelations) != 1 || summary.TargetRelations[0].Host != "www.example.com" {
		t.Fatalf("site target_relation_hints = %+v", summary.TargetRelations)
	}
	if len(summary.Targets) != 1 || summary.Targets[0].CanonicalTarget == nil {
		t.Fatalf("targets = %+v", summary.Targets)
	}
	if len(summary.Targets[0].TargetRelations) != 1 {
		t.Fatalf("target relation hints = %+v", summary.Targets[0].TargetRelations)
	}
	if len(summary.Targets[0].EdgeProviderHints) != 1 || len(summary.Targets[0].EdgeProviderHints[0].Evidence) != 1 {
		t.Fatalf("target edge hints = %+v", summary.Targets[0].EdgeProviderHints)
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
