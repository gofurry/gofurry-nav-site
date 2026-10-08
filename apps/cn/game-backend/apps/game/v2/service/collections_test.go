package service

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	"github.com/redis/go-redis/v9"
)

type collectionReaderFake struct {
	records         []v2models.CollectionRecord
	batch           v2models.CollectionGames
	err             error
	reads           atomic.Int64
	loads           atomic.Int64
	entered         chan string
	release         chan struct{}
	decorations     map[int64]v2models.CollectionTimelineDecoration
	decorationLoads atomic.Int64
	decorationIDs   []int64
	decorationMu    sync.Mutex
	decorationErr   error
}

func (r *collectionReaderFake) LoadCollectionTimelineDecorations(_ context.Context, ids []int64, _ string) (map[int64]v2models.CollectionTimelineDecoration, error) {
	r.decorationLoads.Add(1)
	r.decorationMu.Lock()
	r.decorationIDs = append([]int64(nil), ids...)
	r.decorationMu.Unlock()
	return r.decorations, r.decorationErr
}

func (r *collectionReaderFake) CountPublishedCollections(context.Context, v2models.CollectionQuery) (int64, error) {
	return int64(len(r.records)), r.err
}
func (r *collectionReaderFake) ListPublishedCollections(_ context.Context, query v2models.CollectionQuery) ([]v2models.CollectionRecord, error) {
	limit, offset := query.PageSize, (query.Page-1)*query.PageSize
	r.reads.Add(1)
	if offset >= int64(len(r.records)) {
		return nil, r.err
	}
	return r.records[offset:min(offset+limit, int64(len(r.records)))], r.err
}
func (r *collectionReaderFake) GetPublishedCollection(ctx context.Context, code string) (*v2models.CollectionRecord, error) {
	r.reads.Add(1)
	if r.entered != nil {
		r.entered <- code
		select {
		case <-r.release:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	for _, record := range r.records {
		if record.Code == code {
			return &record, r.err
		}
	}
	return nil, r.err
}
func (r *collectionReaderFake) ListPublishedCollectionHomeSlots(_ context.Context, mode string) ([]v2models.CollectionRecord, error) {
	r.reads.Add(1)
	records := append([]v2models.CollectionRecord(nil), r.records...)
	for i := range records {
		for _, member := range r.batch.Memberships {
			if member.CollectionID != records[i].ID {
				continue
			}
			for _, game := range r.batch.Games {
				if game.Site.ID == member.GameID && (mode == "nsfw" || !game.Adult) {
					records[i].VisibleGameCount++
				}
			}
		}
	}
	return records, r.err
}
func (r *collectionReaderFake) LoadCollectionProjectionGames(_ context.Context, ids []int64, _ string) (v2models.CollectionGames, error) {
	r.loads.Add(1)
	result := v2models.CollectionGames{Games: r.batch.Games}
	for _, m := range r.batch.Memberships {
		for _, id := range ids {
			if m.CollectionID == id {
				result.Memberships = append(result.Memberships, m)
			}
		}
	}
	return result, r.err
}

type collectionCacheEntry struct {
	data  string
	until time.Time
}
type collectionCacheFake struct {
	mu      sync.Mutex
	entries map[string]collectionCacheEntry
	now     time.Time
	err     error
	reads   int
	writes  int
	ttls    []time.Duration
}

func (c *collectionCacheFake) Get(_ context.Context, key string) *redis.StringCmd {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reads++
	if c.err != nil {
		return redis.NewStringResult("", c.err)
	}
	entry, ok := c.entries[key]
	if !ok || !c.now.Before(entry.until) {
		return redis.NewStringResult("", redis.Nil)
	}
	return redis.NewStringResult(entry.data, nil)
}
func (c *collectionCacheFake) Set(_ context.Context, key string, value interface{}, ttl time.Duration) *redis.StatusCmd {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.writes++
	c.ttls = append(c.ttls, ttl)
	if c.err == nil {
		c.entries[key] = collectionCacheEntry{string(value.([]byte)), c.now.Add(ttl)}
	}
	return redis.NewStatusResult("OK", c.err)
}

func collectionFixture() (*collectionReaderFake, *collectionCacheFake, *CollectionService) {
	now := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	r := &collectionReaderFake{records: []v2models.CollectionRecord{
		{ID: 1, Code: "mixed", Name: "中文", NameEn: "English", Info: "中文简介", PublishedAt: now, Slot: 1},
		{ID: 2, Code: "adult-only", Name: "纯成人", NameEn: "Adults", PublishedAt: now, Slot: 2},
		{ID: 3, Code: "empty", Name: "空", NameEn: "Empty", PublishedAt: now, Slot: 3},
	}}
	for id := int64(1); id <= 5; id++ {
		g := v2models.CollectionProjectionGame{Site: v2models.GameV2SiteRecord{ID: id, Name: "游戏", NameEn: "Game", Info: "简介", InfoEn: "Summary", Header: "https://example.test/header.jpg"}}
		if id == 2 {
			g.Adult = true
		}
		r.batch.Games = append(r.batch.Games, g)
		r.batch.Memberships = append(r.batch.Memberships, v2models.CollectionMembership{CollectionID: 1, GameID: id})
	}
	r.batch.Memberships = append(r.batch.Memberships, v2models.CollectionMembership{CollectionID: 2, GameID: 2})
	c := &collectionCacheFake{now: now, entries: map[string]collectionCacheEntry{}}
	s := NewCollectionService(r, c)
	s.Now = func() time.Time { return c.now }
	return r, c, s
}

func TestCollectionVisibilityAndBatchProjection(t *testing.T) {
	r, _, s := collectionFixture()
	ctx := context.Background()
	index, err := s.List(ctx, v2models.CollectionQuery{})
	if err != nil || index.Total != 3 || len(index.Items) != 3 || index.Items[0].VisibleGameCount != 4 || index.Items[1].VisibleGameCount != 0 || index.Items[1].PreviewGames == nil {
		t.Fatalf("index=%+v err=%v", index, err)
	}
	if r.loads.Load() != 1 {
		t.Fatal("must load all collection games once")
	}
	want := []string{"1", "4", "5"}
	got := []string{}
	for _, p := range index.Items[0].PreviewGames {
		got = append(got, p.GameID)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatal("preview must use filtered timeline", got)
	}
	for _, mode := range []string{"sfw", "nsfw"} {
		q := v2models.CollectionQuery{Mode: mode, Lang: "en"}
		detail, e := s.Detail(ctx, "mixed", q)
		wantCount := 4
		if mode == "nsfw" {
			wantCount = 5
		}
		if e != nil || len(detail.Items) != wantCount || detail.Collection.VisibleGameCount != wantCount || detail.Collection.Name != "English" || detail.Collection.Info != "中文简介" {
			t.Fatalf("detail=%+v err=%v", detail, e)
		}
		if detail.Items[0].GameID != "1" || detail.Items[0].Name != "Game" || detail.Items[0].Summary != "Summary" {
			t.Fatal("numeric adult ID or V2 localization drift", detail.Items[0])
		}
		adult, e := s.Detail(ctx, "adult-only", q)
		adultCount := 0
		if mode == "nsfw" {
			adultCount = 1
		}
		if e != nil || adult.Items == nil || len(adult.Items) != adultCount || adult.Collection.VisibleGameCount != adultCount {
			t.Fatalf("adult-only=%+v err=%v", adult, e)
		}
		home, e := s.Home(ctx, q)
		slots := 1
		if mode == "nsfw" {
			slots = 2
		}
		if e != nil || len(home.Slots) != slots || home.Slots[0].Slot != 1 {
			t.Fatalf("home=%+v err=%v", home, e)
		}
		body, _ := json.Marshal(detail)
		for _, forbidden := range []string{"hidden_adult_count", "raw_count", "archived_at", "\"status\"", "\"version\"", "\"id\""} {
			if strings.Contains(string(body), forbidden) {
				t.Fatal("private metadata leaked", forbidden)
			}
		}
	}
	for _, code := range []string{"missing", "draft", "archived", "a--b", "A", "../bad", "", strings.Repeat("a", 65)} {
		if _, err := s.Detail(ctx, code, v2models.CollectionQuery{}); !errors.Is(err, ErrCollectionNotFound) {
			t.Fatalf("%q err=%v", code, err)
		}
	}
	page, err := s.List(ctx, v2models.CollectionQuery{Page: 2, PageSize: 1})
	if err != nil || page.Items[0].Code != "adult-only" || page.Total != 3 || !page.HasMore {
		t.Fatalf("page=%+v %v", page, err)
	}
	page, _ = s.List(ctx, v2models.CollectionQuery{Page: 4, PageSize: 1})
	if page.Items == nil || len(page.Items) != 0 || page.HasMore {
		t.Fatalf("out of range=%+v", page)
	}
	// Defensive collection locale fallback, including data that future publication validation will reject.
	for _, lang := range []string{"zh", "en"} {
		record := v2models.CollectionRecord{Name: " ", NameEn: "Fallback", Info: "中文"}
		if got := collectionInfo(record, lang, 0); got.Name != "Fallback" || got.Info != "中文" {
			t.Fatal("locale fallback", got)
		}
	}
}

func TestCollectionQueryNormalization(t *testing.T) {
	for _, q := range []v2models.CollectionQuery{{}, {Lang: "fr", Mode: "NSFW", Page: -1, PageSize: -2}, {Page: math.MaxInt64, PageSize: 24}} {
		got := NormalizeCollectionQuery(q)
		if got.Lang != "zh" || got.Mode != "sfw" || got.Page != 1 || got.PageSize != 24 {
			t.Fatalf("query=%+v", got)
		}
	}
	if got := NormalizeCollectionQuery(v2models.CollectionQuery{Lang: "en", Mode: "nsfw", Page: 2, PageSize: 100}); got.PageSize != 60 || got.Page != 2 || got.Lang != "en" || got.Mode != "nsfw" {
		t.Fatal(got)
	}
}

func TestCollectionCacheKeysAndTTL(t *testing.T) {
	r, c, s := collectionFixture()
	ctx := context.Background()
	first, err := s.Detail(ctx, "mixed", v2models.CollectionQuery{})
	if err != nil {
		t.Fatal(err)
	}
	first.Items[0].Name = "mutated by caller"
	second, err := s.Detail(ctx, "mixed", v2models.CollectionQuery{})
	if err != nil || r.reads.Load() != 1 || second.Items[0].Name == first.Items[0].Name {
		t.Fatal("cache miss or shared mutation", err)
	}
	_, _ = s.Detail(ctx, "mixed", v2models.CollectionQuery{Lang: "en"})
	_, _ = s.Detail(ctx, "mixed", v2models.CollectionQuery{Mode: "nsfw"})
	_, _ = s.Detail(ctx, "empty", v2models.CollectionQuery{})
	_, _ = s.List(ctx, v2models.CollectionQuery{})
	_, _ = s.List(ctx, v2models.CollectionQuery{Page: 2})
	_, _ = s.List(ctx, v2models.CollectionQuery{PageSize: 1})
	_, _ = s.Home(ctx, v2models.CollectionQuery{})
	for _, key := range []string{
		"detail:2026-10-06:zh:sfw:mixed", "detail:2026-10-06:en:sfw:mixed", "detail:2026-10-06:zh:nsfw:mixed", "detail:2026-10-06:zh:sfw:empty",
		"list:2026-10-06:zh:sfw:1:24", "list:2026-10-06:zh:sfw:2:24", "list:2026-10-06:zh:sfw:1:1", "home:2026-10-06:zh:sfw",
	} {
		if _, ok := c.entries["game:v2:collections:v3:"+key]; !ok {
			t.Fatal("missing key", key)
		}
	}
	for _, ttl := range c.ttls {
		if ttl != time.Hour {
			t.Fatal("TTL", ttl)
		}
	}
	before := r.reads.Load()
	c.now = c.now.Add(time.Hour - time.Nanosecond)
	_, _ = s.Detail(ctx, "mixed", v2models.CollectionQuery{})
	if r.reads.Load() != before {
		t.Fatal("expired too early")
	}
	c.now = c.now.Add(time.Nanosecond)
	_, _ = s.Detail(ctx, "mixed", v2models.CollectionQuery{})
	if r.reads.Load() != before+1 {
		t.Fatal("TTL did not expire")
	}
	c.now = time.Date(2026, 10, 6, 23, 59, 0, 0, time.UTC)
	_, _ = s.Detail(ctx, "mixed", v2models.CollectionQuery{})
	c.now = c.now.Add(time.Minute)
	before = r.reads.Load()
	_, _ = s.Detail(ctx, "mixed", v2models.CollectionQuery{})
	if r.reads.Load() != before+1 {
		t.Fatal("new UTC day reused previous cache")
	}
	if _, ok := c.entries["game:v2:collections:v3:detail:2026-10-07:zh:sfw:mixed"]; !ok {
		t.Fatal("missing new date key")
	}
}

func TestCollectionCacheFailureFallback(t *testing.T) {
	for _, body := range []string{"broken json", "null", "{}", `{"schema_version":99,"as_of_date":"2026-10-06","items":[]}`, `{"schema_version":1,"generated_at":"2026-10-06T00:00:00Z","as_of_date":"2026-10-05","items":[]}`} {
		r, c, s := collectionFixture()
		c.entries["game:v2:collections:v3:detail:2026-10-06:zh:sfw:mixed"] = collectionCacheEntry{body, c.now.Add(time.Hour)}
		if _, e := s.Detail(context.Background(), "mixed", v2models.CollectionQuery{}); e != nil || r.reads.Load() != 1 || c.writes != 1 {
			t.Fatalf("malformed fallback %q: %v", body, e)
		}
	}
	r, c, s := collectionFixture()
	c.err = errors.New("redis unavailable")
	for i := 0; i < 2; i++ {
		if _, err := s.Detail(context.Background(), "mixed", v2models.CollectionQuery{}); err != nil {
			t.Fatal(err)
		}
	}
	if r.reads.Load() != 2 {
		t.Fatal("Redis failure didn't use DB")
	}
	c.err = nil
	r.err = errors.New("database unavailable")
	before := c.writes
	if _, err := s.Detail(context.Background(), "mixed", v2models.CollectionQuery{}); err == nil || c.writes != before {
		t.Fatal("DB error cached")
	}
	s.cache = nil
	r.err = nil
	if _, err := s.Detail(context.Background(), "mixed", v2models.CollectionQuery{}); err != nil {
		t.Fatal("nil Redis", err)
	}
}

func TestCollectionSingleflightAndIndependentKeys(t *testing.T) {
	r, _, s := collectionFixture()
	r.entered = make(chan string, 32)
	r.release = make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	results := make(chan error, 33)
	for i := 0; i < 32; i++ {
		go func() {
			v, e := s.Detail(ctx, "mixed", v2models.CollectionQuery{})
			if e == nil && v.Collection.Code != "mixed" {
				e = errors.New("wrong shared payload")
			}
			results <- e
		}()
	}
	select {
	case <-r.entered:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	go func() {
		v, e := s.Detail(ctx, "empty", v2models.CollectionQuery{})
		if e == nil && v.Collection.Code != "empty" {
			e = errors.New("different key shared payload")
		}
		results <- e
	}()
	select {
	case code := <-r.entered:
		if code != "empty" {
			t.Fatal("duplicate same-key DB build", code)
		}
	case <-ctx.Done():
		t.Fatal("different key blocked by same-key flight")
	}
	close(r.release)
	for i := 0; i < 33; i++ {
		select {
		case err := <-results:
			if err != nil {
				t.Fatal(err)
			}
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
	if r.reads.Load() != 2 {
		t.Fatalf("DB builds=%d", r.reads.Load())
	}
}

func TestCollectionEmptyResponses(t *testing.T) {
	r, _, s := collectionFixture()
	r.records = nil
	home, e := s.Home(context.Background(), v2models.CollectionQuery{})
	if e != nil || home.Slots == nil || len(home.Slots) != 0 {
		t.Fatalf("home=%+v %v", home, e)
	}
	index, e := s.List(context.Background(), v2models.CollectionQuery{})
	if e != nil || index.Items == nil || len(index.Items) != 0 || index.Total != 0 {
		t.Fatalf("index=%+v %v", index, e)
	}
}

func TestCollectionRedisClientOutageAndNil(t *testing.T) {
	r, _, _ := collectionFixture()
	var absent *redis.Client
	closed := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	_ = closed.Close()
	for _, client := range []*redis.Client{absent, closed} {
		s := NewCollectionService(r, client)
		if _, err := s.Detail(context.Background(), "mixed", v2models.CollectionQuery{}); err != nil {
			t.Fatal("unavailable Redis prevented database read", err)
		}
	}
}

func TestCollectionCanceledWaiterDoesNotCancelSharedBuild(t *testing.T) {
	r, _, s := collectionFixture()
	r.entered = make(chan string, 1)
	r.release = make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() { _, e := s.Detail(ctx, "mixed", v2models.CollectionQuery{}); result <- e }()
	select {
	case <-r.entered:
	case <-time.After(3 * time.Second):
		t.Fatal("build did not start")
	}
	cancel()
	if e := <-result; !errors.Is(e, context.Canceled) {
		t.Fatal("waiter ignored cancellation", e)
	}
	// Register a second waiter while the first build is still blocked. The
	// singleflight callback below must never execute; it observes the same work.
	key := collectionCacheKey("detail", s.metadata(), NormalizeCollectionQuery(v2models.CollectionQuery{})) + ":mixed"
	joined := s.flights.DoChan(key, func() (any, error) { return nil, errors.New("first caller canceled shared work") })
	close(r.release)
	select {
	case out := <-joined:
		if out.Err != nil || !out.Shared {
			t.Fatal("shared build failed", out.Err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("shared build did not finish")
	}
	if _, e := s.Detail(context.Background(), "mixed", v2models.CollectionQuery{}); e != nil || r.reads.Load() != 1 {
		t.Fatal("completed cache missing", e)
	}
}

func TestCollectionHomeNeverLoadsAggregates(t *testing.T) {
	r, c, s := collectionFixture()
	for _, mode := range []string{"sfw", "nsfw"} {
		home, err := s.Home(context.Background(), v2models.CollectionQuery{Mode: mode, Lang: "en"})
		if err != nil {
			t.Fatal(err)
		}
		wantSlots, wantCount := 1, 4
		if mode == "nsfw" {
			wantSlots, wantCount = 2, 5
		}
		if len(home.Slots) != wantSlots || home.Slots[0].Collection.VisibleGameCount != wantCount || home.Slots[0].Collection.Name != "English" {
			t.Fatalf("home: %+v", home)
		}
		for _, slot := range home.Slots {
			if slot.Collection.PreviewGames == nil || len(slot.Collection.PreviewGames) != 0 {
				t.Fatal("Home must not project previews")
			}
		}
		_, err = s.Home(context.Background(), v2models.CollectionQuery{Mode: mode, Lang: "en"})
		if err != nil {
			t.Fatal(err)
		}
	}
	if r.loads.Load() != 0 || r.reads.Load() != 2 || c.writes != 2 {
		t.Fatalf("aggregate=%d reads=%d writes=%d", r.loads.Load(), r.reads.Load(), c.writes)
	}
	for _, ttl := range c.ttls {
		if ttl != time.Hour {
			t.Fatal(ttl)
		}
	}
}

func TestCollectionHomeRebuildsLegacyPreviewCacheInSameNamespace(t *testing.T) {
	r, c, s := collectionFixture()
	query := v2models.CollectionQuery{Lang: "en", Mode: "sfw"}
	home, err := s.Home(context.Background(), query)
	if err != nil {
		t.Fatal(err)
	}
	key := "game:v2:collections:v3:home:2026-10-06:en:sfw"
	if _, ok := c.entries[key]; !ok {
		t.Fatal("Home cache namespace changed")
	}
	home.Slots[0].Collection.PreviewGames = []v2models.CollectionPreviewGame{{GameID: "1"}}
	encoded, _ := json.Marshal(home)
	c.entries[key] = collectionCacheEntry{string(encoded), c.now.Add(CollectionCacheTTL)}
	rebuilt, err := s.Home(context.Background(), query)
	if err != nil || len(rebuilt.Slots[0].Collection.PreviewGames) != 0 || r.loads.Load() != 0 || r.reads.Load() != 2 {
		t.Fatalf("legacy cache did not rebuild through light path: %+v %v reads=%d loads=%d", rebuilt, err, r.reads.Load(), r.loads.Load())
	}
}

func TestCollectionDiscoveryBypassesResultCache(t *testing.T) {
	for _, q := range []v2models.CollectionQuery{{Q: "游戏"}, {Phase: "released"}, {Phase: "upcoming"}, {Phase: "mixed"}, {Sort: "count_desc"}, {Sort: "count_asc"}, {Sort: "name_asc"}, {Sort: "name_desc"}} {
		r, cache, service := collectionFixture()
		for range 2 {
			if _, err := service.List(context.Background(), q); err != nil {
				t.Fatal(err)
			}
		}
		if cache.reads != 0 || cache.writes != 0 || len(cache.entries) != 0 || r.reads.Load() != 2 {
			t.Fatalf("discovery must bypass cache: %+v", q)
		}
	}
	r, cache, service := collectionFixture()
	for range 2 {
		if _, err := service.List(context.Background(), v2models.CollectionQuery{}); err != nil {
			t.Fatal(err)
		}
	}
	if cache.writes != 1 || r.reads.Load() != 1 || cache.ttls[0] != time.Hour {
		t.Fatal("browse cache contract")
	}
	for _, q := range []v2models.CollectionQuery{{Phase: "bad"}, {Sort: "bad"}} {
		if _, err := service.List(context.Background(), q); !errors.Is(err, ErrCollectionQuery) {
			t.Fatal(err)
		}
	}
}

func TestCollectionProjectionPreservesGamePresentation(t *testing.T) {
	pointer := func(s string) *string { return &s }
	exists := false
	for _, lang := range []string{"zh", "en"} {
		for _, game := range []v2models.CollectionProjectionGame{
			{Site: v2models.GameV2SiteRecord{ID: 1, NameEn: "Fallback", InfoEn: "Summary", Header: "https://example.test/site"}},
			{Site: v2models.GameV2SiteRecord{ID: 1}, Details: &v2models.GfgGameV2Details{Name: "Detail", HeaderURL: pointer("https://example.test/detail")}, Localized: &v2models.GfgGameV2LocalizedDetails{Lang: "en", Name: "Localized", ShortDescription: pointer("Localized summary")}},
			{Site: v2models.GameV2SiteRecord{ID: 1}, Media: []v2models.GfgGameV2Media{{MediaType: "header", URL: pointer("https://example.test/first")}, {MediaType: "header", URL: pointer("https://example.test/last")}}, Assets: []v2models.GfgGameV2Asset{{AssetType: "header", Lang: "zh", URL: "https://example.test/missing", Exists: &exists}, {AssetType: "header", Lang: "en", URL: "https://example.test/en"}, {AssetType: "header_2x", Lang: "zh", URL: "https://example.test/2x"}}},
		} {
			r, _, service := collectionFixture()
			r.batch = v2models.CollectionGames{Games: []v2models.CollectionProjectionGame{game}, Memberships: []v2models.CollectionMembership{{CollectionID: 1, GameID: 1}}}
			got, err := service.Detail(context.Background(), "mixed", v2models.CollectionQuery{Lang: lang})
			if err != nil {
				t.Fatal(err)
			}
			want := buildListItem(v2models.GameV2Aggregate{Site: game.Site, Details: game.Details, Localized: game.Localized, Media: game.Media, Assets: game.Assets}, lang, defaultRegion)
			if got.Items[0].Name != want.Name || got.Items[0].Summary != want.Summary || got.Items[0].HeaderURL != want.HeaderURL {
				t.Fatalf("projection drift: %+v vs %+v", got.Items[0], want)
			}
		}
	}
}

func TestCollectionDetailDecoratesOnlyVisibleMembersAndCachesResult(t *testing.T) {
	r, c, s := collectionFixture()
	r.decorations = map[int64]v2models.CollectionTimelineDecoration{
		1: {PrimaryTag: &v2models.CollectionTimelineTag{Code: "story", Name: "剧情"}, Rating: &v2models.CollectionTimelineRating{Average: 4.5, Count: 2}, Online: &v2models.CollectionTimelineOnline{Count: 0, CollectedAt: c.now}, CommunityCount: 3},
		2: {PrimaryTag: &v2models.CollectionTimelineTag{Code: "hidden", Name: "Hidden adult metadata"}},
	}
	ctx := context.Background()
	_, _ = s.List(ctx, v2models.CollectionQuery{})
	_, _ = s.Home(ctx, v2models.CollectionQuery{})
	if r.decorationLoads.Load() != 0 {
		t.Fatal("Index/Home must not decorate")
	}
	detail, err := s.Detail(ctx, "mixed", v2models.CollectionQuery{})
	if err != nil || r.decorationLoads.Load() != 1 || !reflect.DeepEqual(r.decorationIDs, []int64{1, 3, 4, 5}) {
		t.Fatal("SFW decoration batch", detail, err, r.decorationIDs)
	}
	if detail.Items[0].Rating.Average != 4.5 || detail.Items[0].Online.Count != 0 || detail.Items[0].CommunityCount != 3 {
		t.Fatal(detail.Items[0])
	}
	encoded, _ := json.Marshal(detail)
	if strings.Contains(string(encoded), "Hidden adult") || !strings.Contains(string(encoded), `"rating":null`) || !strings.Contains(string(encoded), `"schema_version":1`) {
		t.Fatal("additive DTO/null/privacy", string(encoded))
	}
	beforeReads, beforeLoads := r.reads.Load(), r.loads.Load()
	_, err = s.Detail(ctx, "mixed", v2models.CollectionQuery{})
	if err != nil || r.reads.Load() != beforeReads || r.loads.Load() != beforeLoads || r.decorationLoads.Load() != 1 {
		t.Fatal("warm hit rebuilt DB", err)
	}
	_, err = s.Detail(ctx, "mixed", v2models.CollectionQuery{Mode: "nsfw"})
	if err != nil || !reflect.DeepEqual(r.decorationIDs, []int64{1, 2, 3, 4, 5}) {
		t.Fatal("NSFW batch", err, r.decorationIDs)
	}
	before := r.decorationLoads.Load()
	_, err = s.Detail(ctx, "adult-only", v2models.CollectionQuery{})
	if err != nil || r.decorationLoads.Load() != before {
		t.Fatal("empty SFW must not decorate", err)
	}
}

func TestCollectionDecorationFailureNotCached(t *testing.T) {
	r, c, s := collectionFixture()
	r.decorationErr = errors.New("decoration DB failure")
	_, err := s.Detail(context.Background(), "mixed", v2models.CollectionQuery{})
	if err == nil || c.writes != 0 {
		t.Fatal("partial DTO must not be cached", err, c.writes)
	}
	r.decorationErr = nil
	if _, err = s.Detail(context.Background(), "mixed", v2models.CollectionQuery{}); err != nil || r.decorationLoads.Load() != 2 {
		t.Fatal("must rebuild after failure", err)
	}
}

func TestCollectionCacheRevisionRetiresManualOnlyDetailPayload(t *testing.T) {
	r, cache, service := collectionFixture()
	r.decorations = map[int64]v2models.CollectionTimelineDecoration{1: {CommunityCount: 2}}
	detail, err := service.Detail(context.Background(), "mixed", v2models.CollectionQuery{})
	if err != nil {
		t.Fatal(err)
	}
	const oldKey = "game:v2:collections:v2:detail:2026-10-06:zh:sfw:mixed"
	const newKey = "game:v2:collections:v3:detail:2026-10-06:zh:sfw:mixed"
	detail.Items = nil // The old manual-only cache omitted the effective members.
	encoded, _ := json.Marshal(detail)
	cache.entries[oldKey] = collectionCacheEntry{string(encoded), cache.now.Add(time.Hour)}
	delete(cache.entries, newKey)
	reads, loads := r.reads.Load(), r.decorationLoads.Load()
	fresh, err := service.Detail(context.Background(), "mixed", v2models.CollectionQuery{})
	if err != nil || len(fresh.Items) == 0 || fresh.Items[0].CommunityCount != 2 || r.reads.Load() != reads+1 || r.decorationLoads.Load() != loads+1 {
		t.Fatalf("old payload reused: %+v %v", fresh, err)
	}
	if _, ok := cache.entries[oldKey]; !ok {
		t.Fatal("old keys must expire naturally")
	}
	if _, ok := cache.entries[newKey]; !ok {
		t.Fatal("new namespace not written")
	}
	body, _ := json.Marshal(fresh)
	if !strings.Contains(string(body), `"schema_version":1`) {
		t.Fatal("internal revision changed public schema")
	}
	_, err = service.Detail(context.Background(), "mixed", v2models.CollectionQuery{})
	if err != nil || r.reads.Load() != reads+1 || r.decorationLoads.Load() != loads+1 {
		t.Fatal("new warm cache rebuilt", err)
	}
}
