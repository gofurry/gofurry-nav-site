package observation

import (
	"strings"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-collector/roof/env"
)

func TestBuildEdgeProviderHintsCloudflareFromHeaderDNSAndASN(t *testing.T) {
	docs := map[string]LatestDocument{
		ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, time.Now(), map[string]any{
			"headers": map[string]any{
				"CF-Ray":          "abc-HKG",
				"CF-Cache-Status": "HIT",
			},
		}),
		ProtocolDNS: latestDoc(ProtocolDNS, StatusSuccess, time.Now(), map[string]any{
			"cname_terminal": "example.cdn.cloudflare.net.",
			"A": []any{
				map[string]any{
					"value": "203.0.113.10",
					"asn":   "AS13335 (Cloudflare, Inc.)",
					"isp":   "Cloudflare",
				},
			},
		}),
	}

	hints := BuildEdgeProviderHints(docs)
	hint := findEdgeHint(hints, "cloudflare")
	if hint == nil {
		t.Fatalf("missing cloudflare hint: %+v", hints)
	}
	if hint.Confidence != "high" || len(hint.Evidence) < 3 {
		t.Fatalf("unexpected cloudflare hint: %+v", hint)
	}
}

func TestBuildEdgeProviderHintsCommonProviders(t *testing.T) {
	cases := []struct {
		name     string
		provider string
		payload  map[string]any
	}{
		{name: "vercel", provider: "vercel", payload: map[string]any{"headers": map[string]any{"x-vercel-id": "hkg1::abc"}}},
		{name: "netlify", provider: "netlify", payload: map[string]any{"headers": map[string]any{"x-nf-request-id": "abc"}}},
		{name: "cloudfront", provider: "aws_cloudfront", payload: map[string]any{"headers": map[string]any{"x-amz-cf-id": "abc"}}},
		{name: "fastly", provider: "fastly", payload: map[string]any{"headers": map[string]any{"x-fastly-request-id": "abc"}}},
		{name: "github", provider: "github_pages", payload: map[string]any{"headers": map[string]any{"server": "GitHub.com"}}},
		{name: "aliyun", provider: "aliyun", payload: map[string]any{"headers": map[string]any{"x-oss-request-id": "abc"}}},
		{name: "tencent", provider: "tencent_cloud", payload: map[string]any{"headers": map[string]any{"x-nws-log-uuid": "abc"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hints := BuildEdgeProviderHints(map[string]LatestDocument{
				ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, time.Now(), tc.payload),
			})
			if findEdgeHint(hints, tc.provider) == nil {
				t.Fatalf("missing %s hint: %+v", tc.provider, hints)
			}
		})
	}
}

func TestBuildEdgeProviderHintsWeakEvidenceLowConfidence(t *testing.T) {
	hints := BuildEdgeProviderHints(map[string]LatestDocument{
		ProtocolDNS: latestDoc(ProtocolDNS, StatusSuccess, time.Now(), map[string]any{
			"cname_terminal": "example.cloudfront.net.",
		}),
	})
	hint := findEdgeHint(hints, "aws_cloudfront")
	if hint == nil {
		t.Fatalf("missing cloudfront hint: %+v", hints)
	}
	if hint.Confidence != "low" {
		t.Fatalf("Confidence = %q, want low", hint.Confidence)
	}
}

func TestBuildEdgeProviderHintsEvidenceDedupAndLimit(t *testing.T) {
	longValue := "cloudflare " + strings.Repeat("x", edgeEvidenceValueLimit+20)
	hints := BuildEdgeProviderHints(map[string]LatestDocument{
		ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, time.Now(), map[string]any{
			"headers": map[string]any{
				"server": longValue,
				"Server": longValue,
			},
		}),
	})
	hint := findEdgeHint(hints, "cloudflare")
	if hint == nil || len(hint.Evidence) != 1 {
		t.Fatalf("dedup failed: %+v", hints)
	}
	if len([]rune(hint.Evidence[0].Value)) != edgeEvidenceValueLimit {
		t.Fatalf("evidence value not limited: %d", len([]rune(hint.Evidence[0].Value)))
	}
}

func TestBuildTargetSummaryEdgeHintsCanBeDisabled(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	oldV2 := env.GetServerConfig().Collector.V2
	disabled := false
	env.GetServerConfig().Collector.V2 = env.CollectorV2Config{
		EdgeHints: env.EdgeHintsConfig{Enabled: &disabled},
	}
	t.Cleanup(func() {
		env.GetServerConfig().Collector.V2 = oldV2
	})

	summary := BuildTargetSummary(1, "example.com", map[string]LatestDocument{
		ProtocolHTTP: latestDoc(ProtocolHTTP, StatusSuccess, now.Add(-time.Minute), map[string]any{
			"headers": map[string]any{"cf-ray": "abc"},
		}),
	}, now)
	if len(summary.EdgeProviderHints) != 0 {
		t.Fatalf("edge hints should be disabled: %+v", summary.EdgeProviderHints)
	}
	if summary.Status != StatusHealthy {
		t.Fatalf("status changed unexpectedly: %s", summary.Status)
	}
}

func TestBuildSiteSummaryCopiesEdgeHints(t *testing.T) {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	site := BuildSiteSummary(1, []TargetSummaryDocument{
		{
			SiteID: 1,
			Target: "example.com",
			Status: StatusHealthy,
			EdgeProviderHints: []EdgeProviderHint{
				{Provider: "cloudflare", HintType: "cdn", Confidence: "high"},
			},
		},
	}, now)
	if len(site.Targets) != 1 || len(site.Targets[0].EdgeProviderHints) != 1 {
		t.Fatalf("site summary did not copy edge hints: %+v", site.Targets)
	}
	if site.Status != StatusHealthy {
		t.Fatalf("site status changed unexpectedly: %s", site.Status)
	}
}

func findEdgeHint(hints []EdgeProviderHint, provider string) *EdgeProviderHint {
	for i := range hints {
		if hints[i].Provider == provider {
			return &hints[i]
		}
	}
	return nil
}
