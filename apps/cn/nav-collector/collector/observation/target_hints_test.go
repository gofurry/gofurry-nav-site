package observation

import (
	"testing"
	"time"
)

func TestBuildTargetSummaryAddsCanonicalTargetHints(t *testing.T) {
	now := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	summary := BuildTargetSummary(1, "example.com", map[string]LatestDocument{
		ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, now.Add(-time.Minute), map[string]any{
			"final_url":     "https://www.example.com/home",
			"canonical_url": "https://www.example.com/",
			"tls_handshake": "not_tls",
		}),
	}, now)

	if summary.CanonicalTarget == nil {
		t.Fatal("CanonicalTarget should not be nil")
	}
	if summary.CanonicalTarget.TargetHost != "example.com" {
		t.Fatalf("target host = %q", summary.CanonicalTarget.TargetHost)
	}
	if summary.CanonicalTarget.FinalHost != "www.example.com" || summary.CanonicalTarget.CanonicalHost != "www.example.com" {
		t.Fatalf("unexpected canonical hint: %+v", summary.CanonicalTarget)
	}
	if summary.CanonicalTarget.Relation != "redirect_to_www" {
		t.Fatalf("relation = %q, want redirect_to_www", summary.CanonicalTarget.Relation)
	}
	if !hasTargetRelation(summary.TargetRelations, "redirect_to_www", "final_url", "www.example.com") {
		t.Fatalf("missing final_url redirect relation: %+v", summary.TargetRelations)
	}
	if !hasTargetRelation(summary.TargetRelations, "redirect_to_www", "canonical_url", "www.example.com") {
		t.Fatalf("missing canonical_url redirect relation: %+v", summary.TargetRelations)
	}
}

func TestBuildTargetSummaryMarksExternalFinalHost(t *testing.T) {
	now := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	summary := BuildTargetSummary(1, "example.com", map[string]LatestDocument{
		ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, now.Add(-time.Minute), map[string]any{
			"final_url":     "https://other.example.net/",
			"tls_handshake": "not_tls",
		}),
	}, now)

	if summary.CanonicalTarget == nil || summary.CanonicalTarget.Relation != "redirect_to_external" {
		t.Fatalf("unexpected canonical target hint: %+v", summary.CanonicalTarget)
	}
	if !hasTargetRelation(summary.TargetRelations, "redirect_to_external", "final_url", "other.example.net") {
		t.Fatalf("missing external redirect relation: %+v", summary.TargetRelations)
	}
}

func TestBuildSiteSummaryDetectsDuplicateFinalHost(t *testing.T) {
	now := time.Date(2026, 5, 25, 12, 0, 0, 0, time.UTC)
	summaries := []TargetSummaryDocument{
		BuildTargetSummary(1, "example.com", map[string]LatestDocument{
			ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, now.Add(-time.Minute), map[string]any{
				"final_url":     "https://www.example.com/",
				"tls_handshake": "not_tls",
			}),
		}, now),
		BuildTargetSummary(1, "www.example.com", map[string]LatestDocument{
			ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, now.Add(-time.Minute), map[string]any{
				"final_url":     "https://www.example.com/",
				"tls_handshake": "not_tls",
			}),
		}, now),
	}

	site := BuildSiteSummary(1, summaries, now)
	if len(site.TargetRelations) != 1 {
		t.Fatalf("site target relations = %+v, want one duplicate final host hint", site.TargetRelations)
	}
	hint := site.TargetRelations[0]
	if hint.Relation != "duplicate_final_host" || hint.Host != "www.example.com" {
		t.Fatalf("unexpected duplicate hint: %+v", hint)
	}
	if len(hint.Targets) != 2 || hint.Targets[0] != "example.com" || hint.Targets[1] != "www.example.com" {
		t.Fatalf("duplicate targets not stable: %+v", hint.Targets)
	}
	if site.Targets[0].CanonicalTarget == nil || len(site.Targets[0].TargetRelations) == 0 {
		t.Fatalf("site summary item did not copy target hints: %+v", site.Targets[0])
	}
}

func TestBuildTargetRelationHintsUsesRelatedHostForSameRegistrableDomain(t *testing.T) {
	canonical, relations := BuildTargetRelationHints("app.example.co.uk", map[string]LatestDocument{
		ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, time.Now(), map[string]any{
			"final_url": "https://static.example.co.uk/",
		}),
	})
	if canonical == nil || canonical.Relation != "redirect_to_related_host" {
		t.Fatalf("canonical relation = %+v, want related host", canonical)
	}
	if !hasTargetRelation(relations, "redirect_to_related_host", "final_url", "static.example.co.uk") {
		t.Fatalf("missing related host relation: %+v", relations)
	}
}

func hasTargetRelation(values []TargetRelationHint, relation string, source string, relatedHost string) bool {
	for _, value := range values {
		if value.Relation == relation && value.Source == source && value.RelatedHost == relatedHost {
			return true
		}
	}
	return false
}
