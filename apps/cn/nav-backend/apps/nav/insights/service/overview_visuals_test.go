package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/insights/models"
)

func siteVisualFixture(id int64) models.SiteVisualRecord {
	return models.SiteVisualRecord{SiteID: id, Name: " 站点 ", NSFW: "0", Asset: fmt.Sprintf("nav/sites/%d/icon/%s.png", id, strings.Repeat("a", 32))}
}

func TestOverviewSiteVisualAdmission(t *testing.T) {
	for name, mutate := range map[string]func(*models.SiteVisualRecord){
		"adult":                    func(r *models.SiteVisualRecord) { r.NSFW = "1" },
		"true":                     func(r *models.SiteVisualRecord) { r.NSFW = "true" },
		"false is not stored zero": func(r *models.SiteVisualRecord) { r.NSFW = "false" },
		"unknown":                  func(r *models.SiteVisualRecord) { r.NSFW = "unknown" },
		"missing":                  func(r *models.SiteVisualRecord) { r.NSFW = "" },
		"deleted":                  func(r *models.SiteVisualRecord) { r.Deleted = true },
		"id":                       func(r *models.SiteVisualRecord) { r.SiteID = 0 },
		"name":                     func(r *models.SiteVisualRecord) { r.Name = "\t\u3000" },
		"missing icon":             func(r *models.SiteVisualRecord) { r.Asset = "" },
		"foreign icon":             func(r *models.SiteVisualRecord) { r.Asset = strings.Replace(r.Asset, "/1/", "/2/", 1) },
		"url":                      func(r *models.SiteVisualRecord) { r.Asset = "https://example.test/" + r.Asset },
		"query":                    func(r *models.SiteVisualRecord) { r.Asset += "?x=1" },
		"short hash":               func(r *models.SiteVisualRecord) { r.Asset = "nav/sites/1/icon/abc.png" },
		"traversal":                func(r *models.SiteVisualRecord) { r.Asset = "nav/sites/1/icon/../logo.png" },
		"newline":                  func(r *models.SiteVisualRecord) { r.Asset += "\n" },
	} {
		t.Run(name, func(t *testing.T) {
			row := siteVisualFixture(1)
			mutate(&row)
			got := projectOverviewSiteVisuals([]int64{1}, []models.SiteVisualRecord{row})
			data, _ := json.Marshal(got)
			if string(data) != "[]" {
				t.Fatalf("unsafe projection: %s", data)
			}
		})
	}
	row := siteVisualFixture(1)
	row.Name = ""
	row.NameEn = " English "
	if got := projectOverviewSiteVisuals([]int64{1}, []models.SiteVisualRecord{row}); len(got) != 1 || got[0].Name != "English" || got[0].Visual.Kind != "site_icon" {
		t.Fatalf("fallback: %+v", got)
	}
	for _, ext := range []string{"", ".svg", ".ico", ".webp"} {
		row.Asset = "nav/sites/1/icon/" + strings.Repeat("b", 32) + ext
		if len(projectOverviewSiteVisuals([]int64{1}, []models.SiteVisualRecord{row})) != 1 {
			t.Fatal("valid managed key rejected", ext)
		}
	}
}

func TestOverviewSiteVisualOrderBoundsAndMembership(t *testing.T) {
	var rows []models.SiteVisualRecord
	var changes []models.Change
	for id := int64(9); id >= 1; id-- {
		rows = append(rows, siteVisualFixture(id))
		changes = append(changes, models.Change{Entity: models.EntityRef{ID: id}})
	}
	changes = append([]models.Change{{Entity: models.EntityRef{ID: 0}}, changes[0]}, changes...)
	before := slices.Clone(changes)
	ids := overviewVisualIDs(changes)
	if !reflect.DeepEqual(ids, []int64{9, 8, 7, 6, 5, 4, 3, 2}) {
		t.Fatalf("batch identity bound: %v", ids)
	}
	slices.Reverse(rows)
	for n := 0; n <= 8; n++ {
		got := projectOverviewSiteVisuals(ids[:n], rows)
		if got == nil || len(got) != min(n, 5) {
			t.Fatal("0..5 contract", got)
		}
		for i, r := range got {
			if r.SiteID != ids[i] {
				t.Fatal("event order changed", got)
			}
		}
	}
	if !reflect.DeepEqual(before, changes) {
		t.Fatal("projection mutated events")
	}
	if len(projectOverviewSiteVisuals([]int64{9, 9}, rows)) != 1 {
		t.Fatal("duplicate logo")
	}
	if len(projectOverviewSiteVisuals([]int64{99}, rows)) != 0 {
		t.Fatal("invented a non-event site")
	}
}

type overviewVisualStore struct {
	fakeStore
	read  func(context.Context, []int64) ([]models.SiteVisualRecord, error)
	calls int
	fail  string
}

var originalOverviewError = errors.New("existing overview error")

func (s *overviewVisualStore) ListOverviewSiteVisuals(ctx context.Context, ids []int64) ([]models.SiteVisualRecord, error) {
	s.calls++
	return s.read(ctx, ids)
}
func (s *overviewVisualStore) CountEntities(ctx context.Context) (int64, error) {
	if s.fail == "count" {
		return 0, originalOverviewError
	}
	return s.fakeStore.CountEntities(ctx)
}
func (s *overviewVisualStore) CountOverviewChanges(ctx context.Context, a, b []string) (int64, error) {
	if s.fail == "changes" {
		return 0, originalOverviewError
	}
	return s.fakeStore.CountOverviewChanges(ctx, a, b)
}
func (s *overviewVisualStore) GetMetricSummary(ctx context.Context, c models.MetricContract) (*models.MetricSummaryRecord, error) {
	if s.fail == "metric" {
		return nil, originalOverviewError
	}
	return s.fakeStore.GetMetricSummary(ctx, c)
}
func (s *overviewVisualStore) ListOverviewChanges(ctx context.Context, a, b []string, n int32) ([]models.ChangeRecord, error) {
	if s.fail == "feed" {
		return nil, originalOverviewError
	}
	return s.fakeStore.ListOverviewChanges(ctx, a, b, n)
}

func TestOverviewSiteVisualFailureIsolation(t *testing.T) {
	day := time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC)
	for _, mode := range []string{"success", "empty", "error", "timeout", "canceled"} {
		t.Run(mode, func(t *testing.T) {
			store := &overviewVisualStore{fakeStore: fakeStore{overview: []models.ChangeRecord{{EntityID: 1, EntityName: "Site", DetectorKey: "ipv6_transition", DetectorVersion: 2, EventCode: "ipv6_enabled", ProjectionDate: day, TimeBasis: "day"}}, summaries: map[string]*models.MetricSummaryRecord{"ipv6": {FactDate: day, EligibleCount: 1}}}}
			store.read = func(ctx context.Context, ids []int64) ([]models.SiteVisualRecord, error) {
				if !reflect.DeepEqual(ids, []int64{1}) {
					t.Fatal("not the returned event IDs", ids)
				}
				deadline, ok := ctx.Deadline()
				if !ok || time.Until(deadline) > overviewVisualTimeout {
					t.Fatal("unbounded read")
				}
				switch mode {
				case "success":
					return []models.SiteVisualRecord{siteVisualFixture(1)}, nil
				case "empty":
					return nil, nil
				case "timeout":
					<-ctx.Done()
					return nil, ctx.Err()
				case "canceled":
					return nil, context.Canceled
				default:
					return nil, errors.New("private driver details")
				}
			}
			svc := New(store)
			svc.now = func() time.Time { return day }
			got, err := svc.GetOverview(context.Background())
			if err != nil || store.calls != 1 || got.SiteVisuals == nil || (mode == "success") != (len(got.SiteVisuals) == 1) {
				t.Fatalf("optional result: %+v %v calls=%d", got, err, store.calls)
			}
			baselineSvc := New(&store.fakeStore)
			baselineSvc.now = svc.now
			baseline, _ := baselineSvc.GetOverview(context.Background())
			without := got
			without.SiteVisuals = []models.SiteVisual{}
			if !reflect.DeepEqual(without, baseline) || len(got.RecentChanges) != 1 || got.Metrics[0].Value != nil {
				t.Fatal("optional read changed existing semantics")
			}
			data, _ := json.Marshal(got)
			if mode != "success" && !strings.Contains(string(data), `"site_visuals":[]`) {
				t.Fatal("not an array", string(data))
			}
		})
	}
	for _, stage := range []string{"count", "changes", "metric", "feed"} {
		store := &overviewVisualStore{fail: stage}
		_, err := New(store).GetOverview(context.Background())
		if !errors.Is(err, originalOverviewError) || store.calls != 0 {
			t.Fatalf("%s error swallowed: %v", stage, err)
		}
	}
	empty := &overviewVisualStore{}
	if _, err := New(empty).GetOverview(context.Background()); err != nil || empty.calls != 0 {
		t.Fatal("empty events requested decorations")
	}
}
