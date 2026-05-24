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
		PingTime:              cm.LocalTime{},
		AvgLossRate:           0,
		AvgDelayTime:          42,
		MinRTTMS:              35,
		MaxRTTMS:              50,
		StdDevRTTMS:           4,
		PacketsSent:           5,
		PacketsRecv:           5,
		PacketsRecvDuplicates: 1,
		ResolvedIP:            "203.0.113.10",
		ResolvedIPs:           []string{"203.0.113.10"},
		SelectedIP:            "203.0.113.10",
		IPFamily:              "ipv4",
		ResolutionSource:      "go-ping",
		ProbeDurationMS:       5100,
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
	if payload["min_rtt_ms"] != int64(35) || payload["max_rtt_ms"] != int64(50) || payload["stddev_rtt_ms"] != int64(4) {
		t.Fatalf("RTT fields missing or changed: %#v", payload)
	}
	if payload["jitter_ms"] != int64(4) {
		t.Fatalf("jitter_ms = %v, want 4", payload["jitter_ms"])
	}
	if payload["packets_sent"] != 5 || payload["packets_recv"] != 5 || payload["packets_recv_duplicates"] != 1 {
		t.Fatalf("packet fields missing or changed: %#v", payload)
	}
	if payload["resolved_ip"] != "203.0.113.10" {
		t.Fatalf("resolved_ip = %v, want 203.0.113.10", payload["resolved_ip"])
	}
	resolvedIPs := payload["resolved_ips"].([]string)
	if len(resolvedIPs) != 1 || resolvedIPs[0] != "203.0.113.10" {
		t.Fatalf("resolved_ips = %+v", resolvedIPs)
	}
	if payload["selected_ip"] != "203.0.113.10" || payload["ip_family"] != "ipv4" || payload["resolution_source"] != "go-ping" {
		t.Fatalf("IP 解释字段错误: selected=%v family=%v source=%v", payload["selected_ip"], payload["ip_family"], payload["resolution_source"])
	}
	if payload["icmp_blocked_suspected"] != false {
		t.Fatalf("成功 ping 不应标记 icmp_blocked_suspected，got %v", payload["icmp_blocked_suspected"])
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
		AvgLossRate:      100,
		AvgDelayTime:     100000000,
		PacketsSent:      5,
		PacketsRecv:      0,
		ResolvedIP:       "2001:db8::1",
		ResolvedIPs:      []string{"2001:db8::1"},
		SelectedIP:       "2001:db8::1",
		IPFamily:         "ipv6",
		ResolutionSource: "go-ping",
		ProbeDurationMS:  5002,
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
	if payload["min_rtt_ms"] != nil || payload["max_rtt_ms"] != nil || payload["stddev_rtt_ms"] != nil || payload["jitter_ms"] != nil {
		t.Fatalf("failed ping RTT fields should be nil: %#v", payload)
	}
	if payload["error_code"] != "ping_unreachable" {
		t.Fatalf("payload error_code = %v, want ping_unreachable", payload["error_code"])
	}
	if payload["selected_ip"] != "2001:db8::1" || payload["ip_family"] != "ipv6" {
		t.Fatalf("失败 ping IP 字段错误: selected=%v family=%v", payload["selected_ip"], payload["ip_family"])
	}
	if payload["icmp_blocked_suspected"] != true {
		t.Fatalf("全丢包且已解析 IP 时应标记 icmp_blocked_suspected，got %v", payload["icmp_blocked_suspected"])
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
