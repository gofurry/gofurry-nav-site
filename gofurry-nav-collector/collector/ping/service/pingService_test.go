package service

import (
	"testing"

	models2 "github.com/gofurry/gofurry-nav-collector/collector/ping/models"
	cm "github.com/gofurry/gofurry-nav-collector/common/models"
)

func TestBuildPingTargetsKeepsSiteID(t *testing.T) {
	www := "www."
	domains, siteIDByDomain := buildPingTargets([]models2.GfnCollectorDomain{
		{ID: 1, SiteID: 101, Name: "example.com"},
		{ID: 2, SiteID: 101, Name: "example.com", Prefix: &www},
		{ID: 3, SiteID: 202, Name: "example.net"},
		{ID: 4, SiteID: 0, Name: "missing-site-id.example"},
	})
	if len(domains) != 3 {
		t.Fatalf("domains length = %d, want 3", len(domains))
	}
	if siteIDByDomain["example.com"] != 101 {
		t.Fatalf("example.com site id = %d, want 101", siteIDByDomain["example.com"])
	}
	if siteIDByDomain["www.example.com"] != 101 {
		t.Fatalf("www.example.com site id = %d, want 101", siteIDByDomain["www.example.com"])
	}
	if siteIDByDomain["example.net"] != 202 {
		t.Fatalf("example.net site id = %d, want 202", siteIDByDomain["example.net"])
	}
}

func TestBuildPingObservationPayloadSuccess(t *testing.T) {
	payload, errorCode := buildPingObservationPayload(models2.PingModel{
		PingTime:        cm.LocalTime{},
		AvgLossRate:     0,
		AvgDelayTime:    42,
		ProbeDurationMS: 5100,
	}, &models2.PingSaveModel{
		Status: "up",
		Loss:   "0",
		Delay:  "42ms",
	}, "up")

	if errorCode != "" {
		t.Fatalf("errorCode = %q, want empty", errorCode)
	}
	if payload["icmp_status"] != "reachable" {
		t.Fatalf("icmp_status = %v, want reachable", payload["icmp_status"])
	}
	if payload["avg_rtt_ms"] != int64(42) {
		t.Fatalf("avg_rtt_ms = %v, want 42", payload["avg_rtt_ms"])
	}
	if payload["duration_ms"] != int64(5100) {
		t.Fatalf("duration_ms = %v, want 5100", payload["duration_ms"])
	}
	if payload["legacy_delay"] != "42ms" || payload["legacy_loss"] != "0" || payload["legacy_status"] != "up" {
		t.Fatalf("legacy fields missing or changed: %#v", payload)
	}
}

func TestBuildPingObservationPayloadFailure(t *testing.T) {
	payload, errorCode := buildPingObservationPayload(models2.PingModel{
		AvgLossRate:     100,
		AvgDelayTime:    100000000,
		ProbeDurationMS: 5002,
	}, &models2.PingSaveModel{
		Status: "down",
		Loss:   "100",
		Delay:  "100000000ms",
	}, "down")

	if errorCode != "ping_unreachable" {
		t.Fatalf("errorCode = %q, want ping_unreachable", errorCode)
	}
	if payload["icmp_status"] != "unreachable" {
		t.Fatalf("icmp_status = %v, want unreachable", payload["icmp_status"])
	}
	if payload["avg_rtt_ms"] != nil {
		t.Fatalf("avg_rtt_ms = %v, want nil", payload["avg_rtt_ms"])
	}
	if payload["error_code"] != "ping_unreachable" {
		t.Fatalf("payload error_code = %v, want ping_unreachable", payload["error_code"])
	}
	if payload["legacy_delay"] != "100000000ms" || payload["legacy_loss"] != "100" || payload["legacy_status"] != "down" {
		t.Fatalf("legacy fields missing or changed: %#v", payload)
	}
}

func TestBuildPingObservationPayloadKeepsSpecificErrorCode(t *testing.T) {
	payload, errorCode := buildPingObservationPayload(models2.PingModel{
		AvgLossRate:     100,
		AvgDelayTime:    100000000,
		ProbeDurationMS: 3,
		ErrorCode:       "ping_run_failed",
	}, &models2.PingSaveModel{
		Status: "down",
		Loss:   "100",
		Delay:  "100000000ms",
	}, "down")

	if errorCode != "ping_run_failed" {
		t.Fatalf("errorCode = %q, want ping_run_failed", errorCode)
	}
	if payload["error_code"] != "ping_run_failed" {
		t.Fatalf("payload error_code = %v, want ping_run_failed", payload["error_code"])
	}
}
