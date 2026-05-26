package service

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/readmodel/models"
	"github.com/gofurry/gofurry-nav-backend/common"
)

type fakeObservationStore struct {
	rows      []models.GfnCollectorObservation
	err       common.GFError
	siteID    int64
	target    string
	protocol  string
	limit     int
	callCount int
}

func (store *fakeObservationStore) ListObservations(siteID int64, target string, protocol string, limit int) ([]models.GfnCollectorObservation, common.GFError) {
	store.siteID = siteID
	store.target = target
	store.protocol = protocol
	store.limit = limit
	store.callCount++
	if store.err != nil {
		return nil, store.err
	}
	return store.rows, nil
}

func mapGetter(values map[string]string) redisGetter {
	return func(key string) (string, common.GFError) {
		return values[key], nil
	}
}

func failingGetter(err common.GFError) redisGetter {
	return func(string) (string, common.GFError) {
		return "", err
	}
}

func stringPtr(value string) *string {
	return &value
}

func TestReadModelKeysAndProtocols(t *testing.T) {
	if got := LatestKey(models.ProtocolPing, 123); got != "collector:v2:latest:ping:123" {
		t.Fatalf("LatestKey() = %q", got)
	}
	if got := TargetLatestKey(models.ProtocolHTTP, 123, "example.com"); got != "collector:v2:latest:http:123:example.com" {
		t.Fatalf("TargetLatestKey() = %q", got)
	}
	if got := TargetTrendKey(123, "example.com"); got != "collector:v2:trend:target:123:example.com" {
		t.Fatalf("TargetTrendKey() = %q", got)
	}
	if got := TargetChangeKey(123, "example.com"); got != "collector:v2:change:target:123:example.com" {
		t.Fatalf("TargetChangeKey() = %q", got)
	}
	if got := RunStateLatestKey(models.ProtocolDNS); got != "collector:v2:run:dns:latest" {
		t.Fatalf("RunStateLatestKey() = %q", got)
	}
	if len(models.CoreProtocols()) != 3 || len(models.LightProbeProtocols()) != 6 || len(models.AllProtocols()) != 9 {
		t.Fatalf("unexpected protocol groups: core=%v light=%v all=%v", models.CoreProtocols(), models.LightProbeProtocols(), models.AllProtocols())
	}
	if !models.IsProtocolAllowed(models.ProtocolWAFCanary) || models.IsProtocolAllowed("ftp") {
		t.Fatalf("protocol whitelist mismatch")
	}
}

func TestGetReadModelServiceConcurrentInitialization(t *testing.T) {
	previous := readModelSingleton
	fake := &readModelService{get: mapGetter(nil), observations: &fakeObservationStore{}}
	readModelSingleton = fake
	t.Cleanup(func() {
		readModelSingleton = previous
	})

	var wg sync.WaitGroup
	mismatches := make(chan struct{}, 64)
	for range 64 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if GetReadModelService() != fake {
				mismatches <- struct{}{}
			}
		}()
	}
	wg.Wait()
	close(mismatches)
	if len(mismatches) > 0 {
		t.Fatalf("GetReadModelService returned mismatched service %d times", len(mismatches))
	}
}

func TestGetTargetLatestReadyPartialMissingAndDedup(t *testing.T) {
	now := "2026-05-27T10:00:00Z"
	svc := newReadModelService(mapGetter(map[string]string{
		TargetLatestKey(models.ProtocolPing, 123, "example.com"): `{
			"site_id": 123,
			"target": "example.com",
			"protocol": "ping",
			"status": "success",
			"observed_at": "` + now + `",
			"duration_ms": 42,
			"payload": {"avg_rtt_ms": 20},
			"schema_version": 1,
			"collector_id": "collector-a",
			"job_id": "ping-1"
		}`,
		TargetLatestKey(models.ProtocolHTTP, 123, "example.com"): `{
			"site_id": 123,
			"target": "example.com",
			"protocol": "http",
			"status": "success",
			"observed_at": "` + now + `",
			"duration_ms": 120,
			"payload": {"status_code": 200},
			"schema_version": 1
		}`,
	}), nil)

	latest, err := svc.GetTargetLatest(123, " example.com ", []string{models.ProtocolPing, models.ProtocolHTTP, models.ProtocolDNS, models.ProtocolPing})
	if err != nil {
		t.Fatalf("GetTargetLatest() error = %v", err)
	}
	if latest.State != models.SummaryStateReady || latest.Target != "example.com" {
		t.Fatalf("latest = %+v", latest)
	}
	if len(latest.Protocols) != 2 {
		t.Fatalf("protocols = %+v", latest.Protocols)
	}
	if latest.Protocols[models.ProtocolPing].CollectorID != "collector-a" || latest.Protocols[models.ProtocolPing].JobID != "ping-1" {
		t.Fatalf("ping latest identity = %+v", latest.Protocols[models.ProtocolPing])
	}
	if !strings.Contains(string(latest.Protocols[models.ProtocolHTTP].Payload), `"status_code": 200`) {
		t.Fatalf("http payload = %s", latest.Protocols[models.ProtocolHTTP].Payload)
	}
}

func TestGetTargetLatestMissingAndBadJSON(t *testing.T) {
	svc := newReadModelService(mapGetter(nil), nil)
	latest, err := svc.GetTargetLatest(123, "example.com", nil)
	if err != nil {
		t.Fatalf("GetTargetLatest() error = %v", err)
	}
	if latest.State != models.SummaryStateMissing || len(latest.Protocols) != 0 {
		t.Fatalf("latest = %+v", latest)
	}

	svc = newReadModelService(mapGetter(map[string]string{
		TargetLatestKey(models.ProtocolPing, 123, "example.com"): "{bad-json",
	}), nil)
	if _, err := svc.GetTargetLatest(123, "example.com", []string{models.ProtocolPing}); err == nil {
		t.Fatal("GetTargetLatest() expected JSON decode error")
	}
}

func TestGetLightProbeLatestAllowsPartialMissing(t *testing.T) {
	svc := newReadModelService(mapGetter(map[string]string{
		TargetLatestKey(models.ProtocolRDAP, 123, "example.com"):      `{"site_id":123,"target":"example.com","protocol":"rdap","status":"success","observed_at":"2026-05-27T10:00:00Z","duration_ms":30,"payload":{"registrar":"Example"},"schema_version":1}`,
		TargetLatestKey(models.ProtocolWAFCanary, 123, "example.com"): `{"site_id":123,"target":"example.com","protocol":"waf_canary","status":"success","observed_at":"2026-05-27T10:00:00Z","duration_ms":300,"payload":{"blocked_count":3},"schema_version":1}`,
	}), nil)

	latest, err := svc.GetLightProbeLatest(123, "example.com")
	if err != nil {
		t.Fatalf("GetLightProbeLatest() error = %v", err)
	}
	if latest.State != models.SummaryStateReady || len(latest.Protocols) != 2 {
		t.Fatalf("latest = %+v", latest)
	}
	if _, ok := latest.Protocols[models.ProtocolRobots]; ok {
		t.Fatalf("missing robots latest should not be synthesized: %+v", latest.Protocols)
	}
}

func TestListObservationsReadyPromotesRunIdentityAndNormalizesLimit(t *testing.T) {
	now := time.Date(2026, 5, 27, 10, 0, 0, 0, time.UTC)
	store := &fakeObservationStore{
		rows: []models.GfnCollectorObservation{
			{
				ID:            1,
				SiteID:        123,
				Target:        "example.com",
				Protocol:      models.ProtocolHTTP,
				Status:        "success",
				ObservedAt:    now,
				DurationMS:    120,
				ErrorCode:     stringPtr(""),
				ErrorMessage:  stringPtr(""),
				Payload:       `{"collector_id":"collector-a","job_id":"http-1","status_code":200}`,
				SchemaVersion: 1,
			},
		},
	}
	svc := newReadModelService(nil, store)

	observations, err := svc.ListObservations(123, " example.com ", models.ProtocolHTTP, 9999)
	if err != nil {
		t.Fatalf("ListObservations() error = %v", err)
	}
	if store.limit != models.MaxObservationLimit {
		t.Fatalf("limit = %d, want %d", store.limit, models.MaxObservationLimit)
	}
	if observations.State != models.SummaryStateReady || len(observations.Items) != 1 {
		t.Fatalf("observations = %+v", observations)
	}
	item := observations.Items[0]
	if item.CollectorID != "collector-a" || item.JobID != "http-1" {
		t.Fatalf("run identity not promoted: %+v", item)
	}
	if !strings.Contains(string(item.Payload), `"status_code":200`) {
		t.Fatalf("payload = %s", item.Payload)
	}
}

func TestListObservationsMissingDefaultLimitAndErrors(t *testing.T) {
	store := &fakeObservationStore{}
	svc := newReadModelService(nil, store)
	observations, err := svc.ListObservations(123, "example.com", models.ProtocolDNS, 0)
	if err != nil {
		t.Fatalf("ListObservations() error = %v", err)
	}
	if store.limit != models.DefaultObservationLimit {
		t.Fatalf("limit = %d, want %d", store.limit, models.DefaultObservationLimit)
	}
	if observations.State != models.SummaryStateMissing || len(observations.Items) != 0 {
		t.Fatalf("observations = %+v", observations)
	}

	store = &fakeObservationStore{err: common.NewDaoError("db failed")}
	svc = newReadModelService(nil, store)
	if _, err := svc.ListObservations(123, "example.com", models.ProtocolDNS, 10); err == nil {
		t.Fatal("ListObservations() expected DAO error")
	}
	if _, err := svc.ListObservations(0, "example.com", models.ProtocolDNS, 10); err == nil {
		t.Fatal("ListObservations() expected invalid siteID error")
	}
	if _, err := svc.ListObservations(123, "", models.ProtocolDNS, 10); err == nil {
		t.Fatal("ListObservations() expected empty target error")
	}
	if _, err := svc.ListObservations(123, "example.com", "ftp", 10); err == nil {
		t.Fatal("ListObservations() expected invalid protocol error")
	}
}

func TestListObservationsRejectsInvalidPayload(t *testing.T) {
	store := &fakeObservationStore{
		rows: []models.GfnCollectorObservation{
			{SiteID: 123, Target: "example.com", Protocol: models.ProtocolHTTP, Payload: "{bad-json", SchemaVersion: 1},
		},
	}
	svc := newReadModelService(nil, store)
	if _, err := svc.ListObservations(123, "example.com", models.ProtocolHTTP, 10); err == nil {
		t.Fatal("ListObservations() expected invalid payload error")
	}
}

func TestGetTargetTrendReadyMissingAndBadJSON(t *testing.T) {
	svc := newReadModelService(mapGetter(map[string]string{
		TargetTrendKey(123, "example.com"): `{"site_id":123,"target":"example.com","windows":{"24h":{"protocols":{}}},"generated_at":"2026-05-27T10:00:00Z","schema_version":1}`,
	}), nil)
	trend, err := svc.GetTargetTrend(123, "example.com")
	if err != nil {
		t.Fatalf("GetTargetTrend() error = %v", err)
	}
	if trend.State != models.SummaryStateReady || !strings.Contains(string(trend.Windows), `"24h"`) {
		t.Fatalf("trend = %+v", trend)
	}

	svc = newReadModelService(mapGetter(nil), nil)
	trend, err = svc.GetTargetTrend(123, "example.com")
	if err != nil {
		t.Fatalf("GetTargetTrend() missing error = %v", err)
	}
	if trend.State != models.SummaryStateMissing || string(trend.Windows) != "{}" {
		t.Fatalf("missing trend = %+v", trend)
	}

	svc = newReadModelService(mapGetter(map[string]string{TargetTrendKey(123, "example.com"): "{bad-json"}), nil)
	if _, err := svc.GetTargetTrend(123, "example.com"); err == nil {
		t.Fatal("GetTargetTrend() expected decode error")
	}
}

func TestGetTargetChangesReadyMissingAndBadJSON(t *testing.T) {
	svc := newReadModelService(mapGetter(map[string]string{
		TargetChangeKey(123, "example.com"): `{"site_id":123,"target":"example.com","events":[{"event_id":"1","protocol":"http","category":"http","field":"title","old_observed_at":"2026-05-27T09:00:00Z","new_observed_at":"2026-05-27T10:00:00Z","detected_at":"2026-05-27T10:00:00Z"}],"generated_at":"2026-05-27T10:00:00Z","schema_version":1}`,
	}), nil)
	changes, err := svc.GetTargetChanges(123, "example.com")
	if err != nil {
		t.Fatalf("GetTargetChanges() error = %v", err)
	}
	if changes.State != models.SummaryStateReady || !strings.Contains(string(changes.Events), `"event_id"`) {
		t.Fatalf("changes = %+v", changes)
	}

	svc = newReadModelService(mapGetter(nil), nil)
	changes, err = svc.GetTargetChanges(123, "example.com")
	if err != nil {
		t.Fatalf("GetTargetChanges() missing error = %v", err)
	}
	if changes.State != models.SummaryStateMissing || string(changes.Events) != "[]" {
		t.Fatalf("missing changes = %+v", changes)
	}

	svc = newReadModelService(mapGetter(map[string]string{TargetChangeKey(123, "example.com"): "{bad-json"}), nil)
	if _, err := svc.GetTargetChanges(123, "example.com"); err == nil {
		t.Fatal("GetTargetChanges() expected decode error")
	}
}

func TestGetRunStateReadyMissingInvalidProtocolAndBadJSON(t *testing.T) {
	svc := newReadModelService(mapGetter(map[string]string{
		RunStateLatestKey(models.ProtocolDNS): `{"collector_id":"collector-a","job_id":"dns-1","protocol":"dns","status":"complete","started_at":"2026-05-27T10:00:00Z","finished_at":"2026-05-27T10:00:01Z","duration_ms":1000,"target_count":10,"success_count":9,"failure_count":1,"skipped_count":0,"error_count":1}`,
	}), nil)
	state, err := svc.GetRunState(models.ProtocolDNS)
	if err != nil {
		t.Fatalf("GetRunState() error = %v", err)
	}
	if state.State != models.SummaryStateReady || state.JobID != "dns-1" || state.SuccessCount != 9 {
		t.Fatalf("run state = %+v", state)
	}

	svc = newReadModelService(mapGetter(nil), nil)
	state, err = svc.GetRunState(models.ProtocolHTTP)
	if err != nil {
		t.Fatalf("GetRunState() missing error = %v", err)
	}
	if state.State != models.SummaryStateMissing || state.Protocol != models.ProtocolHTTP {
		t.Fatalf("missing run state = %+v", state)
	}
	if _, err := svc.GetRunState("ftp"); err == nil {
		t.Fatal("GetRunState() expected invalid protocol error")
	}

	svc = newReadModelService(mapGetter(map[string]string{RunStateLatestKey(models.ProtocolDNS): "{bad-json"}), nil)
	if _, err := svc.GetRunState(models.ProtocolDNS); err == nil {
		t.Fatal("GetRunState() expected decode error")
	}
}

func TestRedisGetterErrorPropagates(t *testing.T) {
	svc := newReadModelService(failingGetter(common.NewServiceError("redis failed")), nil)
	if _, err := svc.GetTargetLatest(123, "example.com", []string{models.ProtocolPing}); err == nil {
		t.Fatal("GetTargetLatest() expected redis error")
	}
}
