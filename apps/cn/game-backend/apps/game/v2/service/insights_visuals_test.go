package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"reflect"
	"slices"
	"testing"
	"time"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
)

func visualFixture(id int64) v2models.InsightVisualCandidateRecord {
	return v2models.InsightVisualCandidateRecord{GameID: id, AppID: id + 10000, AssetGameID: id, AssetAppID: id + 10000, AssetID: id, HasClassification: true, Name: " 标题 ", NameEn: " Title ", AssetType: "header", Lang: "zh", Asset: fmt.Sprintf("https://shared.steamstatic.com/store_item_assets/steam/apps/%d/header.jpg", id+10000)}
}

func TestInsightSteamHeaderIdentity(t *testing.T) {
	data, err := os.ReadFile("testdata/insight-header-urls.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name, URL string
		Valid     bool
	}
	if err = json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			if got := insightSteamHeader(tc.URL, 12345); got != tc.Valid {
				t.Fatalf("valid=%v want=%v: %s", got, tc.Valid, tc.URL)
			}
		})
	}
}

func TestInsightVisualAdmissionAndEmptyArrays(t *testing.T) {
	cases := map[string]func(*v2models.InsightVisualCandidateRecord){
		"no current classification": func(r *v2models.InsightVisualCandidateRecord) { r.HasClassification = false },
		"adult":                     func(r *v2models.InsightVisualCandidateRecord) { r.HasAdult = true },
		"game id":                   func(r *v2models.InsightVisualCandidateRecord) { r.GameID = 0 },
		"app id":                    func(r *v2models.InsightVisualCandidateRecord) { r.AppID = -1 },
		"asset id":                  func(r *v2models.InsightVisualCandidateRecord) { r.AssetID = 0 },
		"foreign game":              func(r *v2models.InsightVisualCandidateRecord) { r.AssetGameID++ },
		"foreign app":               func(r *v2models.InsightVisualCandidateRecord) { r.AssetAppID++ },
		"absent":                    func(r *v2models.InsightVisualCandidateRecord) { v := false; r.Exists = &v },
		"capsule":                   func(r *v2models.InsightVisualCandidateRecord) { r.AssetType = "capsule" },
		"unsupported language":      func(r *v2models.InsightVisualCandidateRecord) { r.Lang = "fr" },
		"unnamed":                   func(r *v2models.InsightVisualCandidateRecord) { r.Name = "\t\u0085\u00a0\u3000"; r.NameEn = "\n" },
		"missing asset":             func(r *v2models.InsightVisualCandidateRecord) { r.Asset = "" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			r := visualFixture(1)
			mutate(&r)
			got := selectInsightVisuals([]v2models.InsightVisualCandidateRecord{r}, "2026-10-09")
			data, _ := json.Marshal(got)
			if string(data) != "[]" {
				t.Fatalf("unsafe candidate admitted: %s", data)
			}
		})
	}
	r := visualFixture(1)
	r.Name = ""
	r.NameEn = " English "
	got := selectInsightVisuals([]v2models.InsightVisualCandidateRecord{r}, "2026-10-09")
	if len(got) != 1 || got[0].Name != "English" || got[0].NameEn != "English" {
		t.Fatalf("name fallback: %+v", got)
	}
	r.Name = "中文"
	r.NameEn = ""
	got = selectInsightVisuals([]v2models.InsightVisualCandidateRecord{r}, "2026-10-09")
	if got[0].NameEn != "" {
		t.Fatal("must not invent translation")
	}
}

func TestInsightVisualDailyOrderLimitAndPreference(t *testing.T) {
	rows := []v2models.InsightVisualCandidateRecord{}
	for n := 0; n <= 20; n++ {
		got := selectInsightVisuals(rows, "2026-10-09")
		if len(got) != min(n, 3) || got == nil {
			t.Fatalf("size %d: %+v", n, got)
		}
		shuffled := slices.Clone(rows)
		slices.Reverse(shuffled)
		if !reflect.DeepEqual(got, selectInsightVisuals(shuffled, "2026-10-09")) {
			t.Fatal("input order affects rotation")
		}
		rows = append(rows, visualFixture(int64(n+1)))
	}
	if reflect.DeepEqual(selectInsightVisuals(rows, "2026-10-09"), selectInsightVisuals(rows, "2026-10-10")) {
		t.Fatal("daily key did not rotate fixture")
	}
	best := visualFixture(1)
	worseLang := best
	worseLang.Lang = "en"
	worseLang.AssetID = 2
	worseLang.Asset += "?t=2"
	worseType := best
	worseType.AssetType = "header_2x"
	worseType.AssetID = 3
	worseType.Asset += "?t=3"
	worseFamily := best
	worseFamily.AssetFamily = "z"
	worseFamily.AssetID = 4
	worseFamily.Asset += "?t=4"
	worseSort := best
	worseSort.SortOrder = 1
	worseSort.AssetID = 5
	worseSort.Asset += "?t=5"
	worseID := best
	worseID.AssetID = 6
	worseID.Asset += "?t=6"
	got := selectInsightVisuals([]v2models.InsightVisualCandidateRecord{worseType, worseLang, worseFamily, worseSort, worseID, best, best}, "2026-10-09")
	if len(got) != 1 || got[0].Visual.Asset != best.Asset {
		t.Fatalf("asset preference/dedupe: %+v", got)
	}
	data, _ := json.Marshal(got[0])
	var shape map[string]json.RawMessage
	_ = json.Unmarshal(data, &shape)
	if len(shape) != 4 || string(shape["game_id"]) != "1" || got[0].Visual.Kind != "game_header" {
		t.Fatalf("public contract: %s", data)
	}
}

type visualStore struct {
	fakeInsightsStore
	read  func(context.Context, v2models.InsightVisualQuery) ([]v2models.InsightVisualCandidateRecord, error)
	calls int
	fail  string
}

var visualBaseError = errors.New("original overview read failed")

func (s *visualStore) ListInsightVisualCandidates(ctx context.Context, query v2models.InsightVisualQuery) ([]v2models.InsightVisualCandidateRecord, error) {
	s.calls++
	return s.read(ctx, query)
}
func (s *visualStore) CountInsightEntities(ctx context.Context) (int64, error) {
	if s.fail == "count" {
		return 0, visualBaseError
	}
	return s.fakeInsightsStore.CountInsightEntities(ctx)
}
func (s *visualStore) CountInsightOverviewChanges(ctx context.Context, a, b []string) (int64, error) {
	if s.fail == "change count" {
		return 0, visualBaseError
	}
	return s.fakeInsightsStore.CountInsightOverviewChanges(ctx, a, b)
}
func (s *visualStore) GetInsightMetricSummary(ctx context.Context, c v2models.InsightMetricContract) (*v2models.InsightMetricSummaryRecord, error) {
	if s.fail == "metric" {
		return nil, visualBaseError
	}
	return s.fakeInsightsStore.GetInsightMetricSummary(ctx, c)
}
func (s *visualStore) ListInsightOverviewChanges(ctx context.Context, a, b []string, n int32) ([]v2models.InsightChangeRecord, error) {
	if s.fail == "feed" {
		return nil, visualBaseError
	}
	return s.fakeInsightsStore.ListInsightOverviewChanges(ctx, a, b, n)
}

func TestOverviewVisualReadIsolationAndUTC(t *testing.T) {
	for _, mode := range []string{"success", "empty", "error", "timeout", "canceled"} {
		t.Run(mode, func(t *testing.T) {
			store := &visualStore{fakeInsightsStore: fakeInsightsStore{summaries: map[string]*v2models.InsightMetricSummaryRecord{"free": {FactDate: time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC), EligibleCount: 1}}}}
			store.changes = []v2models.InsightChangeRecord{{EntityID: 82, EntityName: "Game", DetectorKey: "mac_support_transition", DetectorVersion: 1, EventCode: "mac_support_added", ProjectionDate: time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC), TimeBasis: "day"}}
			store.read = func(ctx context.Context, query v2models.InsightVisualQuery) ([]v2models.InsightVisualCandidateRecord, error) {
				if query.UTCDay != "2026-10-08" {
					t.Fatalf("not UTC: %s", query.UTCDay)
				}
				deadline, ok := ctx.Deadline()
				if !ok || time.Until(deadline) > insightVisualQueryTimeout {
					t.Fatal("optional read is unbounded")
				}
				switch mode {
				case "success":
					return []v2models.InsightVisualCandidateRecord{visualFixture(1)}, nil
				case "empty":
					return nil, nil
				case "timeout":
					<-ctx.Done()
					return nil, ctx.Err()
				case "canceled":
					return nil, context.Canceled
				default:
					return nil, errors.New("sensitive driver error")
				}
			}
			svc := NewInsightsService(store)
			svc.now = func() time.Time { return time.Date(2026, 10, 9, 0, 30, 0, 0, time.FixedZone("CST", 8*3600)) }
			got, err := svc.GetInsightsOverview(context.Background())
			if err != nil || got.EntityCount != 1 || len(got.Metrics) != 1 || got.RecentChanges == nil || store.calls != 1 {
				t.Fatalf("overview isolation: %+v %v calls=%d", got, err, store.calls)
			}
			if got.FeaturedVisuals == nil || (mode == "success") != (len(got.FeaturedVisuals) == 1) {
				t.Fatalf("visual result: %+v", got.FeaturedVisuals)
			}
			baselineService := NewInsightsService(&store.fakeInsightsStore)
			baselineService.now = svc.now
			baseline, baselineErr := baselineService.GetInsightsOverview(context.Background())
			withoutVisuals := got
			withoutVisuals.FeaturedVisuals = []v2models.InsightFeaturedVisual{}
			if baselineErr != nil || len(got.RecentChanges) != 1 || got.Changes7D != 1 || !reflect.DeepEqual(baseline, withoutVisuals) {
				t.Fatalf("optional branch changed existing facts: %+v baseline=%+v", got, baseline)
			}
			for _, metric := range got.Metrics {
				if metric.Value != nil {
					t.Fatal("unknown metric became zero")
				}
			}
			data, _ := json.Marshal(got)
			var shape map[string]json.RawMessage
			_ = json.Unmarshal(data, &shape)
			if string(shape["hero_visual"]) != "null" {
				t.Fatalf("missing facts must serialize null: %s", data)
			}
			if mode != "success" && string(shape["featured_visuals"]) != "[]" {
				t.Fatalf("not always array: %s", data)
			}
		})
	}
	for _, stage := range []string{"count", "change count", "metric", "feed"} {
		t.Run(stage, func(t *testing.T) {
			store := &visualStore{fail: stage}
			_, err := NewInsightsService(store).GetInsightsOverview(context.Background())
			if !errors.Is(err, visualBaseError) || store.calls != 0 {
				t.Fatalf("original error masked: %v calls=%d", err, store.calls)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	store := &visualStore{read: func(ctx context.Context, _ v2models.InsightVisualQuery) ([]v2models.InsightVisualCandidateRecord, error) {
		if ctx.Err() != context.Canceled {
			t.Fatal("parent cancellation lost")
		}
		return nil, ctx.Err()
	}}
	if got, hero := NewInsightsService(store).overviewVisuals(ctx, time.Now(), nil); got == nil || len(got) != 0 || hero != nil {
		t.Fatal("cancellation fallback")
	}
}

func TestInsightHeroMetricContext(t *testing.T) {
	date := time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC)
	keys := []string{"free", "windows", "mac", "linux"}
	internal := []string{"free_game_share", "windows_support", "mac_support", "linux_support"}
	for i, key := range keys {
		t.Run(key, func(t *testing.T) {
			zero := 0.0
			metrics := []v2models.InsightMetric{{Key: key, Value: &zero, AsOf: "2026-10-07"}}
			for _, other := range keys[i+1:] {
				metrics = append(metrics, v2models.InsightMetric{Key: other, Value: &zero, AsOf: "2026-10-08"})
			}
			for _, bad := range []float64{-1, 2, math.NaN(), math.Inf(1)} {
				// Invalid earlier keys (including a duplicate of the chosen key)
				// must not hide the first reliable value, matching Nav Web.
				input := append([]v2models.InsightMetric{{Key: "free", Value: &bad}}, metrics...)
				contract, gotDate := insightHeroContext(input)
				if contract.PublicKey != key || contract.InternalKey != internal[i] || contract.Version != 1 || !gotDate.Equal(date) {
					t.Fatalf("controlled mapping/zero/priority: %+v %s", contract, gotDate)
				}
			}
			for _, badDate := range []string{"", "2026-02-30", "2026-10-07T00:00:00Z", "2026-1-07"} {
				metrics[0].AsOf = badDate
				contract, gotDate := insightHeroContext(metrics)
				if contract.PublicKey != "" || !gotDate.IsZero() {
					t.Fatal("invalid selected date fell through to another metric")
				}
			}
		})
	}
}

func TestInsightHeroProofAndOrdinaryCandidates(t *testing.T) {
	value := 0.25
	metrics := []v2models.InsightMetric{{Key: "free", Value: &value, AsOf: "2026-10-07"}}
	for _, mode := range []string{"positive", "no match", "unsafe positive", "no context", "read failure"} {
		t.Run(mode, func(t *testing.T) {
			store := &visualStore{read: func(_ context.Context, q v2models.InsightVisualQuery) ([]v2models.InsightVisualCandidateRecord, error) {
				if mode != "no context" && (q.MetricKey != "free_game_share" || q.MetricVersion != 1 || q.FactDate.Format(time.DateOnly) != "2026-10-07" || q.UTCDay != "2026-10-09") {
					t.Fatalf("wrong query context: %+v", q)
				}
				if mode == "read failure" {
					return nil, errors.New("optional failure")
				}
				candidate := visualFixture(3)
				candidate.HeroMatch = mode != "no match"
				candidate.HasAdult = mode == "unsafe positive"
				return []v2models.InsightVisualCandidateRecord{visualFixture(1), visualFixture(2), candidate}, nil
			}}
			input := metrics
			if mode == "no context" {
				input = nil
			}
			visuals, hero := NewInsightsService(store).overviewVisuals(context.Background(), time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC), input)
			if store.calls != 1 || (hero != nil) != (mode == "positive") {
				t.Fatalf("proof/query isolation: %+v calls=%d", hero, store.calls)
			}
			if hero != nil && (hero.GameID != 3 || hero.MetricKey != "free" || hero.AsOf != "2026-10-07" || visuals[0].GameID != 3) {
				t.Fatalf("wrong proof: %+v", hero)
			}
			want := 3
			if mode == "unsafe positive" {
				want = 2
			}
			if mode == "read failure" {
				want = 0
			}
			if visuals == nil || len(visuals) != want {
				t.Fatalf("ordinary candidates lost: %+v", visuals)
			}
		})
	}
}
