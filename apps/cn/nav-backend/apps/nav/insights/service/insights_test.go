package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/insights/models"
)

type fakeStore struct {
	site          *models.SiteRecord
	summaries     map[string]*models.MetricSummaryRecord
	trend         []models.MetricTrendRecord
	siteMetrics   map[string]*models.SiteMetricRecord
	overview      []models.ChangeRecord
	entityChanges []models.ChangeRecord
}

func (f *fakeStore) CountEntities(context.Context) (int64, error)               { return 2, nil }
func (f *fakeStore) GetSite(context.Context, int64) (*models.SiteRecord, error) { return f.site, nil }
func (f *fakeStore) GetMetricSummary(_ context.Context, c models.MetricContract) (*models.MetricSummaryRecord, error) {
	return f.summaries[c.PublicKey], nil
}
func (f *fakeStore) ListMetricTrend(context.Context, models.MetricContract, int32) ([]models.MetricTrendRecord, error) {
	return f.trend, nil
}
func (f *fakeStore) GetSiteMetric(_ context.Context, _ int64, c models.MetricContract) (*models.SiteMetricRecord, error) {
	return f.siteMetrics[c.PublicKey], nil
}
func (f *fakeStore) CountOverviewChanges(context.Context, []string, []string) (int64, error) {
	return int64(len(f.overview)), nil
}
func (f *fakeStore) ListOverviewChanges(context.Context, []string, []string, int32) ([]models.ChangeRecord, error) {
	return f.overview, nil
}
func (f *fakeStore) ListSiteChanges(context.Context, models.SiteRecord, []string, []string, int32) ([]models.ChangeRecord, error) {
	return f.entityChanges, nil
}

func TestPublicMetricAndStateMappingsAreFrozen(t *testing.T) {
	want := map[string]struct {
		key     string
		version int32
	}{
		"ipv6": {"ipv6_adoption", 2}, "tls13": {"tls13_adoption", 1},
		"security_txt": {"security_txt_adoption", 2},
	}
	for publicKey, expected := range want {
		contract, ok := resolveMetric(publicKey)
		if !ok || contract.InternalKey != expected.key || contract.Version != expected.version {
			t.Fatalf("mapping %q = %#v, %v", publicKey, contract, ok)
		}
	}
	if _, ok := resolveMetric("ipv6_adoption"); ok {
		t.Fatal("internal metric key must not be public")
	}
	states := map[string]string{
		"positive": "supported", "negative": "unsupported", "probe_failed": "unavailable",
		"unknown": "unknown", "not_probed": "not_probed", "stale": "stale", "not_applicable": "not_applicable",
	}
	for internal, expected := range states {
		if actual, ok := publicState(internal); !ok || actual != expected {
			t.Fatalf("state %q = %q, %v", internal, actual, ok)
		}
	}
	if state, _ := publicState("unknown"); state == "unsupported" {
		t.Fatal("unknown must not collapse to unsupported")
	}
}

func TestMetricMathRequiresExact30DayComparison(t *testing.T) {
	now := time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC)
	record := models.MetricSummaryRecord{FactDate: now, PositiveCount: 40, NegativeCount: 60, EligibleCount: 125}
	metric := publicMetric("ipv6", record)
	if metric.Delta30D != nil || metric.Value == nil || *metric.Value != 0.4 || metric.Coverage == nil || *metric.Coverage != 0.8 {
		t.Fatalf("unexpected metric without exact comparison: %#v", metric)
	}
	positive, negative := int64(20), int64(80)
	record.PreviousPositiveCount, record.PreviousNegativeCount = &positive, &negative
	metric = publicMetric("ipv6", record)
	if metric.Delta30D == nil || *metric.Delta30D != 0.2 {
		t.Fatalf("delta = %v", metric.Delta30D)
	}
}

func TestSiteInsightsPreservesUncertaintySameDayDataAndEntityTimeline(t *testing.T) {
	date := time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC)
	store := &fakeStore{
		site: &models.SiteRecord{ID: 7, Name: "Site"},
		siteMetrics: map[string]*models.SiteMetricRecord{
			"ipv6": {FactDate: date, State: "probe_failed", EligibleCount: 10, PositiveCount: 4, NegativeCount: 4},
		},
		entityChanges: []models.ChangeRecord{
			{EntityID: 7, EntityName: "Site", DetectorKey: "primary_target_transition", DetectorVersion: 1, EventCode: "primary_target_changed", ProjectionDate: date, TimeBasis: "day"},
			{EntityID: 7, EntityName: "Site", DetectorKey: "ipv6_transition", DetectorVersion: 2, EventCode: "ipv6_enabled", ProjectionDate: date.AddDate(0, 0, -1), TimeBasis: "day"},
		},
	}
	got, err := New(store).GetSiteInsights(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Capabilities) != 1 || got.Capabilities[0].State != "unavailable" || got.Capabilities[0].AsOf != "2026-08-30" {
		t.Fatalf("capabilities = %#v", got.Capabilities)
	}
	if len(got.RecentChanges) != 2 || got.RecentChanges[0].OccurredAt != nil {
		t.Fatalf("entity timeline was filtered/deduplicated or day time fabricated: %#v", got.RecentChanges)
	}
	payload, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"detector_key", "detector_version", "event_code", "old_value", "new_value", "ipv6_adoption"} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("public payload leaked %q: %s", forbidden, payload)
		}
	}
}

func TestTrendValidationAndMissingEntity(t *testing.T) {
	store := &fakeStore{summaries: map[string]*models.MetricSummaryRecord{}, siteMetrics: map[string]*models.SiteMetricRecord{}}
	service := New(store)
	if _, err := service.GetMetricTrend(context.Background(), "bad", "30d"); !errors.Is(err, ErrInvalidMetricKey) {
		t.Fatalf("invalid key error = %v", err)
	}
	if _, err := service.GetMetricTrend(context.Background(), "ipv6", "7d"); !errors.Is(err, ErrInvalidRange) {
		t.Fatalf("invalid range error = %v", err)
	}
	if _, err := service.GetSiteInsights(context.Background(), 99); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing site error = %v", err)
	}
}
