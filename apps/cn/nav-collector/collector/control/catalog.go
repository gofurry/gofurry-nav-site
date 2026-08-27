package control

import (
	"context"
	"fmt"
	"time"

	"github.com/gofurry/gofurry-nav-collector/collector/observation"
	navsqlc "github.com/gofurry/gofurry-nav-collector/internal/db/nav/sqlc"
	"github.com/gofurry/gofurry-nav-collector/roof/env"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	priorityScheduled = int32(100)
	priorityCatchUp   = int32(150)
	priorityManual    = int32(200)
	priorityEntity    = int32(300)
)

type Capability struct {
	JobKey      string
	Protocol    string
	PointInTime bool
	Interval    time.Duration
	Offset      time.Duration
	Enabled     bool
}

func catalog() []Capability {
	cfg := env.GetServerConfig().Collector
	pingInterval := time.Duration(cfg.Ping.PingInterval) * time.Second
	if pingInterval <= 0 {
		pingInterval = time.Minute
	}
	httpInterval := time.Duration(cfg.Request.RequestInterval) * time.Hour
	if httpInterval <= 0 {
		httpInterval = time.Hour
	}
	dnsInterval := time.Duration(cfg.Dns.DnsInterval) * time.Hour
	if dnsInterval <= 0 {
		dnsInterval = time.Hour
	}
	return []Capability{
		{JobKey: "nav.ping", Protocol: observation.ProtocolPing, PointInTime: true, Interval: pingInterval, Enabled: cfg.V2.ProtocolEnabled(observation.ProtocolPing)},
		{JobKey: "nav.http", Protocol: observation.ProtocolHTTP, PointInTime: true, Interval: httpInterval, Offset: 5 * time.Minute, Enabled: cfg.V2.ProtocolEnabled(observation.ProtocolHTTP)},
		{JobKey: "nav.dns", Protocol: observation.ProtocolDNS, PointInTime: true, Interval: dnsInterval, Offset: 15 * time.Minute, Enabled: cfg.V2.ProtocolEnabled(observation.ProtocolDNS)},
		{JobKey: "nav.rdap", Protocol: observation.ProtocolRDAP, Interval: cfg.V2.LightProbe.RDAP.Interval(), Offset: 25 * time.Minute, Enabled: cfg.V2.ProtocolEnabled(observation.ProtocolRDAP)},
		{JobKey: "nav.robots", Protocol: observation.ProtocolRobots, Interval: cfg.V2.LightProbe.Robots.Interval(), Offset: 30 * time.Minute, Enabled: cfg.V2.ProtocolEnabled(observation.ProtocolRobots)},
		{JobKey: "nav.security_txt", Protocol: observation.ProtocolSecurityTXT, Interval: cfg.V2.LightProbe.SecurityTXT.Interval(), Offset: 35 * time.Minute, Enabled: cfg.V2.ProtocolEnabled(observation.ProtocolSecurityTXT)},
		{JobKey: "nav.llms_txt", Protocol: observation.ProtocolLLMSTXT, Interval: cfg.V2.LightProbe.LLMSTXT.Interval(), Offset: 40 * time.Minute, Enabled: cfg.V2.ProtocolEnabled(observation.ProtocolLLMSTXT)},
		{JobKey: "nav.page_assets", Protocol: observation.ProtocolPageAssets, Interval: cfg.V2.LightProbe.PageAssets.Interval(), Offset: 45 * time.Minute, Enabled: cfg.V2.ProtocolEnabled(observation.ProtocolPageAssets)},
		{JobKey: "nav.port_check", Protocol: observation.ProtocolPortCheck, PointInTime: true, Interval: cfg.V2.LightProbe.PortCheck.Interval(), Offset: 50 * time.Minute, Enabled: cfg.V2.ProtocolEnabled(observation.ProtocolPortCheck)},
		{JobKey: "nav.waf_canary", Protocol: observation.ProtocolWAFCanary, Interval: cfg.V2.LightProbe.WAFCanary.Interval(), Offset: 55 * time.Minute, Enabled: cfg.V2.ProtocolEnabled(observation.ProtocolWAFCanary)},
	}
}

func capability(jobKey string) (Capability, bool) {
	for _, item := range catalog() {
		if item.JobKey == jobKey {
			return item, true
		}
	}
	return Capability{}, false
}

func ensureSchedules(ctx context.Context, queries *navsqlc.Queries) error {
	baseAnchor := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, item := range catalog() {
		seconds := int64(item.Interval / time.Second)
		if seconds <= 0 {
			return fmt.Errorf("invalid interval for %s", item.JobKey)
		}
		policy := "catch_up_once"
		grace := int32(300)
		if item.PointInTime {
			policy, grace = "skip", 90
		}
		if err := queries.EnsureNavCollectionSchedule(ctx, navsqlc.EnsureNavCollectionScheduleParams{
			JobKey: item.JobKey, Name: "Nav " + item.Protocol + " collection",
			Enabled: item.Enabled, ScheduleKind: "interval", IntervalSeconds: &seconds,
			AnchorAt: pgtype.Timestamptz{Time: baseAnchor.Add(item.Offset), Valid: true},
			Timezone: "UTC", MisfirePolicy: policy, MisfireGraceSeconds: grace,
			Priority: priorityScheduled, ConcurrencyKey: item.Protocol,
		}); err != nil {
			return fmt.Errorf("ensure %s schedule: %w", item.JobKey, err)
		}
	}
	return nil
}
