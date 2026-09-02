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
	site                 *models.SiteRecord
	sites                map[int64]*models.SiteRecord
	summaries            map[string]*models.MetricSummaryRecord
	trend                []models.MetricTrendRecord
	siteMetrics          map[string]*models.SiteMetricRecord
	overview             []models.ChangeRecord
	entityChanges        []models.ChangeRecord
	breakdown            []models.DimensionRecord
	sliceAvailability    models.DimensionAvailabilityRecord
	sliceTrend           []models.DimensionTrendRecord
	explorer             []models.ChangeRecord
	lastBreakdownDate    time.Time
	lastSliceThrough     time.Time
	lastSliceValue       string
	lastExplorer         models.ChangeExplorerConditions
	certificateSummary   *models.CertificateOverviewRecord
	certificateAttention []models.CertificateItemRecord
	certificateIssues    []models.CertificateItemRecord
	compareHorizon       *time.Time
	compareCapabilities  []models.SiteCompareCapabilityRecord
	compareCertificates  []models.CertificateItemRecord
}

func (f *fakeStore) CountEntities(context.Context) (int64, error) { return 2, nil }
func (f *fakeStore) GetSite(_ context.Context, id int64) (*models.SiteRecord, error) {
	if f.sites != nil {
		return f.sites[id], nil
	}
	return f.site, nil
}
func (f *fakeStore) GetMetricSummary(_ context.Context, c models.MetricContract) (*models.MetricSummaryRecord, error) {
	return f.summaries[c.PublicKey], nil
}
func (f *fakeStore) ListMetricTrend(context.Context, models.MetricContract, int32) ([]models.MetricTrendRecord, error) {
	return f.trend, nil
}
func (f *fakeStore) ListMetricBreakdown(_ context.Context, _ models.MetricContract, _ models.DimensionContract, date time.Time) ([]models.DimensionRecord, error) {
	f.lastBreakdownDate = date
	return f.breakdown, nil
}
func (f *fakeStore) GetMetricSliceAvailability(context.Context, models.MetricContract, models.DimensionContract, string) (models.DimensionAvailabilityRecord, error) {
	return f.sliceAvailability, nil
}
func (f *fakeStore) ListMetricSliceTrend(_ context.Context, _ models.MetricContract, _ models.DimensionContract, value string, through time.Time, _ int32) ([]models.DimensionTrendRecord, error) {
	f.lastSliceValue, f.lastSliceThrough = value, through
	return f.sliceTrend, nil
}
func (f *fakeStore) GetSiteMetric(_ context.Context, _ int64, c models.MetricContract) (*models.SiteMetricRecord, error) {
	return f.siteMetrics[c.PublicKey], nil
}
func (f *fakeStore) GetSiteCompareHorizon(context.Context, []int64, []models.MetricContract) (*time.Time, error) {
	return f.compareHorizon, nil
}
func (f *fakeStore) ListSiteCompareCapabilities(context.Context, []int64, []models.MetricContract, time.Time) ([]models.SiteCompareCapabilityRecord, error) {
	return f.compareCapabilities, nil
}
func (f *fakeStore) ListSiteCompareCertificates(context.Context, []int64, time.Time) ([]models.CertificateItemRecord, error) {
	return f.compareCertificates, nil
}
func (f *fakeStore) GetCertificateOverviewSummary(context.Context) (*models.CertificateOverviewRecord, error) {
	return f.certificateSummary, nil
}
func (f *fakeStore) ListCertificateExpiryAttention(context.Context, int32) ([]models.CertificateItemRecord, error) {
	return f.certificateAttention, nil
}
func (f *fakeStore) ListCertificateVerificationIssues(context.Context, int32) ([]models.CertificateItemRecord, error) {
	return f.certificateIssues, nil
}
func (f *fakeStore) CountOverviewChanges(context.Context, []string, []string) (int64, error) {
	return int64(len(f.overview)), nil
}
func (f *fakeStore) ListOverviewChanges(context.Context, []string, []string, int32) ([]models.ChangeRecord, error) {
	return f.overview, nil
}
func (f *fakeStore) ListExplorerChanges(_ context.Context, conditions models.ChangeExplorerConditions) ([]models.ChangeRecord, error) {
	f.lastExplorer = conditions
	return f.explorer, nil
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
		"http2": {"http2_adoption", 1}, "hsts": {"hsts_adoption", 1},
		"csp": {"csp_adoption", 1}, "security_txt": {"security_txt_adoption", 2},
		"certificate_verified": {"tls_certificate_verification", 1},
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

func TestCertificateOverviewUsesFactDayAndPreservesCertificateSemantics(t *testing.T) {
	factDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	referenceAt := factDate.AddDate(0, 0, 1)
	expired := referenceAt
	sevenDays := referenceAt.Add(7 * 24 * time.Hour)
	unknownIssue := "raw x509 text"
	verified, failed := true, false
	store := &fakeStore{
		certificateSummary: &models.CertificateOverviewRecord{
			FactDate: factDate, ReferenceAt: referenceAt, FreshnessSeconds: 172800,
			Population: 10, Eligible: 8, Verified: 4, Failed: 2,
			NotApplicable: 2, Stale: 1, NotProbed: 0, ProbeFailed: 0, Unknown: 1,
			Expired: 1, ExpiresWithin7D: 1, ExpiresIn8To30D: 1, Later: 2,
		},
		certificateAttention: []models.CertificateItemRecord{
			{SiteID: 1, SiteName: "Expired", Target: "expired.example", NotAfter: &expired, Verified: &failed, VerificationIssue: &unknownIssue},
			{SiteID: 2, SiteName: "Boundary", Target: "boundary.example", NotAfter: &sevenDays, Verified: &verified},
		},
		certificateIssues: []models.CertificateItemRecord{
			{SiteID: 1, SiteName: "Expired", Target: "expired.example", NotAfter: &expired, Verified: &failed, VerificationIssue: &unknownIssue},
		},
	}
	got, err := New(store).GetCertificateOverview(context.Background(), 20)
	if err != nil {
		t.Fatal(err)
	}
	if got.AsOf == nil || *got.AsOf != "2026-09-01" || got.ReferenceAt == nil || !got.ReferenceAt.Equal(referenceAt) {
		t.Fatalf("fact-day horizon = %#v", got)
	}
	if got.Verification.Known != 6 || got.Verification.Coverage == nil || *got.Verification.Coverage != .75 ||
		got.Expiry.Known != 5 || got.Expiry.Coverage == nil || *got.Expiry.Coverage != .625 {
		t.Fatalf("certificate denominator math = %#v", got)
	}
	if len(got.ExpiryAttention) != 2 || got.ExpiryAttention[0].ExpiryStatus == nil ||
		*got.ExpiryAttention[0].ExpiryStatus != "expired" || got.ExpiryAttention[1].ExpiryStatus == nil ||
		*got.ExpiryAttention[1].ExpiryStatus != "expires_within_7d" {
		t.Fatalf("expiry boundaries = %#v", got.ExpiryAttention)
	}
	if got.VerificationIssues[0].VerificationIssue == nil || *got.VerificationIssues[0].VerificationIssue != "other" {
		t.Fatalf("raw verification issue was not whitelisted: %#v", got.VerificationIssues)
	}
	if _, err := New(store).GetCertificateOverview(context.Background(), 101); !errors.Is(err, ErrInvalidCertificateLimit) {
		t.Fatalf("invalid limit error = %v", err)
	}
	empty, err := New(&fakeStore{}).GetCertificateOverview(context.Background(), 0)
	if err != nil || empty.AsOf != nil || empty.Verification.Coverage != nil || empty.Expiry.Coverage != nil {
		t.Fatalf("empty overview = %#v, %v", empty, err)
	}
}

func TestP23ChangeContractsKeepOverviewWhitelistAndCategories(t *testing.T) {
	overviewDetectors, overviewIDs := changeQueryContracts(true)
	joined := strings.Join(append(overviewDetectors, overviewIDs...), "|")
	for _, required := range []string{
		"http2_transition", "hsts_transition", "csp_transition",
		"tls_certificate_verification_transition", "tls_certificate_verification_failed",
	} {
		if !strings.Contains(joined, required) {
			t.Fatalf("overview contracts missing %q: %s", required, joined)
		}
	}
	if strings.Contains(joined, "tls_certificate_changed") {
		t.Fatalf("routine certificate replacement entered overview: %s", joined)
	}
	certificate, ok := explorerContracts("certificate", "")
	if !ok || len(certificate) != 3 {
		t.Fatalf("certificate explorer contracts = %#v, %v", certificate, ok)
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

func TestDimensionBreakdownUsesGlobalHorizonAndPreservesNullUnknownAndFallback(t *testing.T) {
	horizon := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	store := &fakeStore{
		summaries: map[string]*models.MetricSummaryRecord{"ipv6": {FactDate: horizon}},
		breakdown: []models.DimensionRecord{
			{Value: "true", Population: 5, Eligible: 5, PositiveCount: 3, NegativeCount: 1},
			{Value: "unknown", Population: 2, Eligible: 0},
		},
	}
	got, err := New(store).GetMetricBreakdown(context.Background(), "ipv6", "nsfw")
	if err != nil {
		t.Fatal(err)
	}
	if got.AsOf == nil || *got.AsOf != "2026-09-01" || !store.lastBreakdownDate.Equal(horizon) {
		t.Fatalf("breakdown did not use global horizon: %#v date=%v", got, store.lastBreakdownDate)
	}
	if len(got.Items) != 2 || got.Items[0].Value != "nsfw" || got.Items[0].Known != 4 ||
		got.Items[0].MetricValue == nil || *got.Items[0].MetricValue != .75 || got.Items[0].Coverage == nil || *got.Items[0].Coverage != .8 {
		t.Fatalf("breakdown mapping/math = %#v", got.Items)
	}
	if got.Items[1].Value != "unknown" || got.Items[1].MetricValue != nil || got.Items[1].Coverage != nil {
		t.Fatalf("unknown/zero denominator collapsed: %#v", got.Items[1])
	}
	fallback := publicDimensionSliceRef(models.DimensionContract{PublicKey: "group"}, "12", nil, nil)
	if fallback.Label == nil || *fallback.Label != "分组 #12" || fallback.LabelEn == nil || *fallback.LabelEn != "Group #12" {
		t.Fatalf("deleted group fallback = %#v", fallback)
	}
}

func TestSliceTrendAnchorsToGlobalAndDoesNotFillMissingDates(t *testing.T) {
	horizon := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	from := horizon.AddDate(0, 0, -20)
	through := horizon.AddDate(0, 0, -2)
	store := &fakeStore{
		summaries:         map[string]*models.MetricSummaryRecord{"ipv6": {FactDate: horizon}},
		sliceAvailability: models.DimensionAvailabilityRecord{AvailableFrom: &from, AvailableThrough: &through},
		sliceTrend: []models.DimensionTrendRecord{
			{FactDate: horizon.AddDate(0, 0, -2), Population: 4, Eligible: 4, PositiveCount: 2, NegativeCount: 2},
			{FactDate: horizon, Population: 4, Eligible: 0},
		},
	}
	got, err := New(store).GetMetricSliceTrend(context.Background(), "ipv6", "nsfw", "sfw", "30d")
	if err != nil {
		t.Fatal(err)
	}
	if !store.lastSliceThrough.Equal(horizon) || store.lastSliceValue != "false" || len(got.Points) != 2 {
		t.Fatalf("slice trend anchor/value/fill = %#v through=%v value=%q", got, store.lastSliceThrough, store.lastSliceValue)
	}
	if got.Points[1].MetricValue != nil || got.Points[1].Coverage != nil || got.AvailableThrough == nil || *got.AvailableThrough != "2026-08-30" {
		t.Fatalf("slice null/availability semantics = %#v", got)
	}
}

func TestExplorerCursorBindsFiltersAndFrozenRangeWithoutDedupe(t *testing.T) {
	day := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	exact := day.Add(8 * time.Hour)
	store := &fakeStore{explorer: []models.ChangeRecord{
		{EntityID: 1, DetectorKey: "ipv6_transition", DetectorVersion: 2, EventCode: "ipv6_enabled", ProjectionDate: day, TimeBasis: "observed", EventAt: &exact, PrecisionRank: 1, EventSortAt: exact, OpaqueTie: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		{EntityID: 1, DetectorKey: "tls13_transition", DetectorVersion: 1, EventCode: "tls13_disabled", ProjectionDate: day, TimeBasis: "day", PrecisionRank: 0, EventSortAt: day, OpaqueTie: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
		{EntityID: 2, DetectorKey: "security_txt_transition", DetectorVersion: 2, EventCode: "security_txt_added", ProjectionDate: day.AddDate(0, 0, -1), TimeBasis: "day", PrecisionRank: 0, EventSortAt: day.AddDate(0, 0, -1), OpaqueTie: "cccccccccccccccccccccccccccccccc"},
	}}
	service := New(store)
	service.now = func() time.Time { return day.Add(12 * time.Hour) }
	first, err := service.GetChanges(context.Background(), models.ChangeExplorerQuery{Range: "30d", Category: "capability", Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Items) != 2 || first.Items[0].Entity.ID != first.Items[1].Entity.ID || first.Items[1].OccurredAt != nil || first.NextCursor == nil {
		t.Fatalf("explorer deduplicated or fabricated precision: %#v", first)
	}
	payload, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"detector_key", "detector_version", "event_code", "event_key", "source_key", "ipv6_transition"} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("explorer leaked %q: %s", forbidden, payload)
		}
	}
	service.now = func() time.Time { return day.AddDate(0, 0, 1) }
	_, err = service.GetChanges(context.Background(), models.ChangeExplorerQuery{Range: "30d", Category: "capability", Cursor: *first.NextCursor, Limit: 2})
	if err != nil || !store.lastExplorer.RangeThrough.Equal(day) || store.lastExplorer.Position == nil {
		t.Fatalf("cursor did not freeze range/position: err=%v conditions=%#v", err, store.lastExplorer)
	}
	if _, err := service.GetChanges(context.Background(), models.ChangeExplorerQuery{Range: "7d", Category: "capability", Cursor: *first.NextCursor, Limit: 2}); !errors.Is(err, ErrInvalidCursor) {
		t.Fatalf("cursor/filter mismatch = %v", err)
	}
	if contracts, ok := explorerContracts("", "site.tls13.enabled"); !ok || len(contracts) != 1 || contracts[0].category != "capability" {
		t.Fatalf("exact public type filter = %#v, %v", contracts, ok)
	}
}
