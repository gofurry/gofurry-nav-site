package contentsync

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/gofurry/gofurry-rag/config"
	"github.com/gofurry/gofurry-rag/internal/db"
)

func TestRunNowCreatesAndSkipsAndUpdatesDocuments(t *testing.T) {
	repo := newFakeRepo()
	client := &fakeSyncClient{
		sites: map[string][]NavSite{
			"zh-CN": {{ID: "1", Name: "站点", Domain: "example.com", Info: "简介", Country: "CN", NSFW: "no", Welfare: "no"}},
			"en-US": {{ID: "1", Name: "Site", Domain: "example.com", Info: "Intro", Country: "CN", NSFW: "no", Welfare: "no"}},
		},
		groups: map[string][]NavGroup{
			"zh-CN": {{ID: "g1", Name: "社区", Sites: []string{"1"}}},
			"en-US": {{ID: "g1", Name: "Community", Sites: []string{"1"}}},
		},
		details: map[string]NavSiteDetail{
			"1:zh-CN": {Name: "站点", Info: "简介", Country: "CN", NSFW: "no", Welfare: "no"},
			"1:en-US": {Name: "Site", Info: "Intro", Country: "CN", NSFW: "no", Welfare: "no"},
		},
		httpRecords: map[string]NavHTTPRecord{
			"example.com": {URL: "https://example.com", Title: "Example", Meta: struct {
				Description string `json:"description"`
			}{Description: "desc"}},
		},
	}
	manager := NewManager(config.Config{SyncTimeoutSeconds: 30}, repo, client, client)

	if err := manager.RunNow(context.Background(), SourceNavSites, TriggerManual); err != nil {
		t.Fatal(err)
	}
	if repo.created != 2 || repo.updated != 0 || repo.skipped != 0 {
		t.Fatalf("first sync counts = created:%d updated:%d skipped:%d", repo.created, repo.updated, repo.skipped)
	}

	if err := manager.RunNow(context.Background(), SourceNavSites, TriggerManual); err != nil {
		t.Fatal(err)
	}
	if repo.skipped != 2 {
		t.Fatalf("expected skips, got %d", repo.skipped)
	}

	client.sites["zh-CN"][0].Info = "新的简介"
	client.details["1:zh-CN"] = NavSiteDetail{Name: "站点", Info: "新的简介", Country: "CN", NSFW: "no", Welfare: "no"}
	if err := manager.RunNow(context.Background(), SourceNavSites, TriggerManual); err != nil {
		t.Fatal(err)
	}
	if repo.updated == 0 {
		t.Fatalf("expected update, got %+v", repo.docs)
	}
}

func TestRunNowFallbacksWhenNavDetailMissing(t *testing.T) {
	repo := newFakeRepo()
	client := &fakeSyncClient{
		sites: map[string][]NavSite{
			"zh-CN": {{ID: "101", Name: "站点", Domain: "example.com", Info: "来自首页"}},
			"en-US": {{ID: "101", Name: "Site", Domain: "example.com", Info: "From home"}},
		},
		groups: map[string][]NavGroup{
			"zh-CN": {{ID: "g1", Name: "社区", Sites: []string{"101"}}},
			"en-US": {{ID: "g1", Name: "Community", Sites: []string{"101"}}},
		},
		detailErrs: map[string]error{
			"101:zh-CN": errors.New("not found"),
			"101:en-US": errors.New("not found"),
		},
	}
	manager := NewManager(config.Config{SyncTimeoutSeconds: 30}, repo, client, client)

	if err := manager.RunNow(context.Background(), SourceNavSites, TriggerManual); err != nil {
		t.Fatal(err)
	}
	if repo.created != 2 {
		t.Fatalf("created = %d", repo.created)
	}
	run := repo.runs[SourceNavSites]
	if run.Status != "success" || run.FailedCount != 0 {
		t.Fatalf("run = %+v", run)
	}
}

func TestTriggerRejectsConcurrentRun(t *testing.T) {
	repo := newFakeRepo()
	client := &fakeSyncClient{
		sites: map[string][]NavSite{
			"zh-CN": {{ID: "1", Name: "站点", Domain: "example.com"}},
			"en-US": {{ID: "1", Name: "Site", Domain: "example.com"}},
		},
		groups: map[string][]NavGroup{
			"zh-CN": nil,
			"en-US": nil,
		},
		details: map[string]NavSiteDetail{
			"1:zh-CN": {Name: "站点"},
			"1:en-US": {Name: "Site"},
		},
		block: make(chan struct{}),
	}
	manager := NewManager(config.Config{SyncTimeoutSeconds: 30}, repo, client, client)
	if err := manager.Trigger(context.Background(), SourceNavSites, TriggerManual); err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)
	err := manager.Trigger(context.Background(), SourceGameDetails, TriggerManual)
	if err == nil {
		t.Fatal("expected conflict error")
	}
	var syncErr interface{ HTTPStatus() int }
	if !errors.As(err, &syncErr) || syncErr.HTTPStatus() != 409 {
		t.Fatalf("err = %v", err)
	}
	close(client.block)
}

func TestRunNowRejectsRemovedChangelogSource(t *testing.T) {
	manager := NewManager(config.Config{SyncTimeoutSeconds: 30}, newFakeRepo(), &fakeSyncClient{}, &fakeSyncClient{})

	err := manager.RunNow(context.Background(), "site_changelog", TriggerManual)
	if err == nil {
		t.Fatal("expected unsupported source error")
	}
	var syncErr interface{ HTTPStatus() int }
	if !errors.As(err, &syncErr) || syncErr.HTTPStatus() != 400 {
		t.Fatalf("err = %v", err)
	}
}

func TestRunNowSyncsGameSources(t *testing.T) {
	repo := newFakeRepo()
	client := &fakeSyncClient{
		gameLists: map[string][]GameSummary{
			"zh": {{ID: "7", Name: "星火", Info: "中文简介", ReleaseDate: "2026-01-01", Developers: []string{"Team A"}, Publishers: []string{"Pub A"}}},
			"en": {{ID: "7", Name: "Spark", Info: "English intro", ReleaseDate: "2026-01-01", Developers: []string{"Team A"}, Publishers: []string{"Pub A"}}},
		},
		gameDetails: map[string]GameDetail{
			"7:zh": {
				Name:                "星火",
				Info:                "中文简介",
				Platform:            "Windows",
				SupportedLanguages:  "简体中文, English",
				Website:             "https://example.com/games/7",
				DetailedDescription: "很长的介绍",
				AboutTheGame:        "关于内容",
				Developers:          []string{"Team A"},
				Publishers:          []string{"Pub A"},
				Groups:              []GameKV{{Key: "group", Value: "剧情"}},
				Tags:                []GameTag{{ID: "1", Name: "RPG"}},
			},
			"7:en": {
				Name:                "Spark",
				Info:                "English intro",
				Platform:            "Windows",
				SupportedLanguages:  "English, Simplified Chinese",
				Website:             "https://example.com/games/7",
				DetailedDescription: "Long description",
				AboutTheGame:        "About text",
				Developers:          []string{"Team A"},
				Publishers:          []string{"Pub A"},
				Groups:              []GameKV{{Key: "group", Value: "Story"}},
				Tags:                []GameTag{{ID: "1", Name: "RPG"}},
			},
		},
		gameNews: map[string][]GameNews{
			"zh": {{Name: "星火", Headline: "更新 1", PostTime: "2026-05-01", Author: "福狼", Content: "中文内容", URL: "https://example.com/news/1"}},
			"en": {{Name: "Spark", Headline: "Update 1", PostTime: "2026-05-01", Author: "Furry", Content: "English content", URL: ""}},
		},
	}
	manager := NewManager(config.Config{SyncTimeoutSeconds: 30}, repo, client, client)

	if err := manager.RunNow(context.Background(), SourceGameDetails, TriggerManual); err != nil {
		t.Fatal(err)
	}
	if err := manager.RunNow(context.Background(), SourceGameNews, TriggerManual); err != nil {
		t.Fatal(err)
	}
	if repo.created != 4 {
		t.Fatalf("created = %d", repo.created)
	}

	if err := manager.RunNow(context.Background(), SourceAll, TriggerManual); err != nil {
		t.Fatal(err)
	}
	if repo.skipped < 4 {
		t.Fatalf("expected game documents to skip on second sync, skipped = %d", repo.skipped)
	}

	client.gameDetails["7:zh"] = GameDetail{
		Name:                "星火",
		Info:                "中文简介更新",
		Platform:            "Windows",
		SupportedLanguages:  "简体中文, English",
		Website:             "https://example.com/games/7",
		DetailedDescription: "很长的介绍",
		AboutTheGame:        "关于内容",
		Developers:          []string{"Team A"},
		Publishers:          []string{"Pub A"},
		Groups:              []GameKV{{Key: "group", Value: "剧情"}},
		Tags:                []GameTag{{ID: "1", Name: "RPG"}},
	}
	if err := manager.RunNow(context.Background(), SourceGameDetails, TriggerManual); err != nil {
		t.Fatal(err)
	}
	if repo.updated == 0 {
		t.Fatalf("expected updates, got %+v", repo.docs)
	}
}

type fakeRepo struct {
	docs    map[string]db.Document
	runs    map[string]db.SyncRun
	nextID  int64
	nextRun int64
	created int
	updated int
	skipped int
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		docs:    make(map[string]db.Document),
		runs:    make(map[string]db.SyncRun),
		nextID:  1,
		nextRun: 1,
	}
}

func (r *fakeRepo) UpsertSyncedDocument(ctx context.Context, params db.SyncDocumentParams) (db.SyncDocumentResult, error) {
	key := params.SourceType + ":" + params.SourceID
	if len(params.Metadata) == 0 {
		params.Metadata = json.RawMessage(`{}`)
	}
	doc, ok := r.docs[key]
	if !ok {
		doc = db.Document{
			ID:         r.nextID,
			SourceType: params.SourceType,
			SourceID:   params.SourceID,
			Title:      params.Title,
			URL:        params.URL,
			Checksum:   params.Checksum,
			Content:    params.Content,
			Metadata:   params.Metadata,
			Status:     db.StatusPending,
		}
		r.nextID++
		r.docs[key] = doc
		r.created++
		return db.SyncDocumentResult{Action: "created", Document: doc}, nil
	}
	if doc.Checksum == params.Checksum {
		r.skipped++
		return db.SyncDocumentResult{Action: "skipped", Document: doc}, nil
	}
	doc.Title = params.Title
	doc.URL = params.URL
	doc.Checksum = params.Checksum
	doc.Content = params.Content
	doc.Metadata = params.Metadata
	doc.Status = db.StatusPending
	r.docs[key] = doc
	r.updated++
	return db.SyncDocumentResult{Action: "updated", Document: doc}, nil
}

func (r *fakeRepo) CreateSyncRun(ctx context.Context, params db.CreateSyncRunParams) (db.SyncRun, error) {
	run := db.SyncRun{
		ID:        r.nextRun,
		Source:    params.Source,
		Trigger:   params.Trigger,
		Status:    "running",
		StartedAt: time.Now(),
	}
	r.nextRun++
	r.runs[params.Source] = run
	return run, nil
}

func (r *fakeRepo) CompleteSyncRun(ctx context.Context, id int64, params db.CompleteSyncRunParams) error {
	for key, run := range r.runs {
		if run.ID == id {
			completed := time.Now()
			run.Status = params.Status
			run.CompletedAt = &completed
			run.SourceTotalCount = params.SourceTotalCount
			run.AddedCount = params.AddedCount
			run.UpdatedCount = params.UpdatedCount
			run.SkippedCount = params.SkippedCount
			run.FailedCount = params.FailedCount
			run.Message = params.Message
			r.runs[key] = run
			return nil
		}
	}
	return nil
}

func (r *fakeRepo) LatestSyncRuns(ctx context.Context) (map[string]db.SyncRun, error) {
	result := make(map[string]db.SyncRun, len(r.runs))
	for key, run := range r.runs {
		result[key] = run
	}
	return result, nil
}

func (r *fakeRepo) CountDocumentsBySourceType(ctx context.Context) (map[string]int64, error) {
	result := make(map[string]int64)
	for _, doc := range r.docs {
		result[doc.SourceType]++
	}
	return result, nil
}

type fakeSyncClient struct {
	sites       map[string][]NavSite
	groups      map[string][]NavGroup
	details     map[string]NavSiteDetail
	detailErrs  map[string]error
	httpRecords map[string]NavHTTPRecord
	gameLists   map[string][]GameSummary
	gameDetails map[string]GameDetail
	gameNews    map[string][]GameNews
	block       chan struct{}
}

func (f *fakeSyncClient) ListSites(ctx context.Context, locale string) ([]NavSite, error) {
	if f.block != nil {
		<-f.block
	}
	return append([]NavSite(nil), f.sites[locale]...), nil
}

func (f *fakeSyncClient) ListGroups(ctx context.Context, locale string) ([]NavGroup, error) {
	return append([]NavGroup(nil), f.groups[locale]...), nil
}

func (f *fakeSyncClient) GetSiteDetail(ctx context.Context, id, locale string) (NavSiteDetail, error) {
	if err := f.detailErrs[id+":"+locale]; err != nil {
		return NavSiteDetail{}, err
	}
	return f.details[id+":"+locale], nil
}

func (f *fakeSyncClient) GetSiteHTTP(ctx context.Context, domain string) (NavHTTPRecord, error) {
	return f.httpRecords[domain], nil
}

func (f *fakeSyncClient) ListGames(ctx context.Context, locale string) ([]GameSummary, error) {
	return append([]GameSummary(nil), f.gameLists[locale]...), nil
}

func (f *fakeSyncClient) GetGameInfo(ctx context.Context, id, locale string) (GameDetail, error) {
	return f.gameDetails[id+":"+locale], nil
}

func (f *fakeSyncClient) ListGameNews(ctx context.Context, locale string) ([]GameNews, error) {
	return append([]GameNews(nil), f.gameNews[locale]...), nil
}
