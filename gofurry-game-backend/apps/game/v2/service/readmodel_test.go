package service

import (
	"context"
	"testing"
	"time"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	"github.com/gofurry/gofurry-game-backend/common"
)

type fakeDetailReader struct {
	aggregate  v2models.GameV2Aggregate
	query      v2models.GameV2DetailQuery
	listQuery  v2models.GameV2ListQuery
	panelQuery v2models.GameV2PanelQuery
	err        common.GFError
}

func (reader *fakeDetailReader) GetGameDetailAggregate(_ context.Context, query v2models.GameV2DetailQuery) (v2models.GameV2Aggregate, common.GFError) {
	reader.query = query
	if reader.err != nil {
		return v2models.GameV2Aggregate{}, reader.err
	}
	return reader.aggregate, nil
}

func (reader *fakeDetailReader) ListGameAggregates(_ context.Context, query v2models.GameV2ListQuery) ([]v2models.GameV2Aggregate, common.GFError) {
	reader.listQuery = query
	if reader.err != nil {
		return nil, reader.err
	}
	return []v2models.GameV2Aggregate{reader.aggregate}, nil
}

func (reader *fakeDetailReader) GetGameNews(_ context.Context, _ v2models.GameV2NewsQuery) ([]v2models.GameV2NewsRow, common.GFError) {
	if reader.err != nil {
		return nil, reader.err
	}
	return []v2models.GameV2NewsRow{}, nil
}

func (reader *fakeDetailReader) ListTopOnlineAggregates(_ context.Context, query v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError) {
	reader.panelQuery = query
	return reader.ListGameAggregates(context.Background(), v2models.GameV2ListQuery{})
}

func (reader *fakeDetailReader) ListFreeGameAggregates(_ context.Context, _ v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError) {
	return reader.ListGameAggregates(context.Background(), v2models.GameV2ListQuery{})
}

func (reader *fakeDetailReader) ListHighestDiscountAggregates(_ context.Context, _ v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError) {
	return reader.ListGameAggregates(context.Background(), v2models.GameV2ListQuery{})
}

func (reader *fakeDetailReader) ListLowPriceAggregates(_ context.Context, _ v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError) {
	return reader.ListGameAggregates(context.Background(), v2models.GameV2ListQuery{})
}

func (reader *fakeDetailReader) GetLatestGameNews(_ context.Context, _ v2models.GameV2NewsQuery) ([]v2models.GameV2NewsRow, common.GFError) {
	if reader.err != nil {
		return nil, reader.err
	}
	return []v2models.GameV2NewsRow{}, nil
}

func (reader *fakeDetailReader) GetCollectStatus(_ context.Context) (v2models.GameV2CollectStatus, common.GFError) {
	if reader.err != nil {
		return v2models.GameV2CollectStatus{}, reader.err
	}
	return v2models.GameV2CollectStatus{}, nil
}

func (reader *fakeDetailReader) ListCollectRuns(_ context.Context, _ v2models.GameV2CollectRunQuery) ([]v2models.GfgGameV2CollectRun, common.GFError) {
	if reader.err != nil {
		return nil, reader.err
	}
	return []v2models.GfgGameV2CollectRun{}, nil
}

func (reader *fakeDetailReader) GetCollectRun(_ context.Context, _ string) (*v2models.GfgGameV2CollectRun, common.GFError) {
	if reader.err != nil {
		return nil, reader.err
	}
	return &v2models.GfgGameV2CollectRun{}, nil
}

func (reader *fakeDetailReader) ListCollectTaskResults(_ context.Context, _ v2models.GameV2CollectTaskResultQuery) ([]v2models.GfgGameV2CollectTaskResult, common.GFError) {
	if reader.err != nil {
		return nil, reader.err
	}
	return []v2models.GfgGameV2CollectTaskResult{}, nil
}

func (reader *fakeDetailReader) GetGameCollectStatus(_ context.Context, _ int64, _ int64) (v2models.GameV2CollectGameStatus, common.GFError) {
	if reader.err != nil {
		return v2models.GameV2CollectGameStatus{}, reader.err
	}
	return v2models.GameV2CollectGameStatus{}, nil
}

func (reader *fakeDetailReader) ListSyncCreators(_ context.Context, _ string) ([]v2models.GameV2SyncCreatorRow, common.GFError) {
	if reader.err != nil {
		return nil, reader.err
	}
	return []v2models.GameV2SyncCreatorRow{}, nil
}

func TestGetGameDetailUsesLocalizedFallback(t *testing.T) {
	reader := &fakeDetailReader{
		aggregate: v2models.GameV2Aggregate{
			Site: v2models.GameV2SiteRecord{
				ID:     1,
				AppID:  440,
				Name:   "军团要塞2",
				NameEn: "Team Fortress 2",
				Info:   "中文站内简介",
				InfoEn: "English site summary",
			},
			Details: &v2models.GfgGameV2Details{
				GameID: 1,
				AppID:  440,
				Name:   "Team Fortress 2",
				Type:   "game",
			},
			Localized: &v2models.GfgGameV2LocalizedDetails{
				GameID:           1,
				AppID:            440,
				Lang:             "zh",
				Name:             "军团要塞2",
				ShortDescription: strPtr("中文采集简介"),
			},
		},
	}

	svc := NewReadModelServiceWithReader(reader)
	res, err := svc.GetGameDetail(context.Background(), v2models.GameV2DetailRequest{GameID: 1, Lang: "en"})
	if err != nil {
		t.Fatalf("GetGameDetail returned error: %s", err.GetMsg())
	}

	if reader.query.Lang != "en" {
		t.Fatalf("expected DAO query lang en, got %s", reader.query.Lang)
	}
	if res.RequestedLang != "en" {
		t.Fatalf("expected requested lang en, got %s", res.RequestedLang)
	}
	if res.Lang != "zh" {
		t.Fatalf("expected fallback lang zh, got %s", res.Lang)
	}
	if res.Name != "军团要塞2" {
		t.Fatalf("expected zh localized name, got %s", res.Name)
	}
	if res.Summary != "中文采集简介" {
		t.Fatalf("expected zh localized summary, got %s", res.Summary)
	}
}

func TestGetGameDetailMarksCNPriceUnavailableWithoutHKFallback(t *testing.T) {
	reader := &fakeDetailReader{
		aggregate: v2models.GameV2Aggregate{
			Site: v2models.GameV2SiteRecord{ID: 1, AppID: 570, Name: "Dota 2"},
			Prices: []v2models.GfgGameV2Price{
				{
					GameID:         1,
					AppID:          570,
					Region:         "CN",
					IsFree:         false,
					FinalAmount:    0,
					Currency:       strPtr(""),
					FinalFormatted: strPtr(""),
				},
				{
					GameID:          1,
					AppID:           570,
					Region:          "HK",
					IsFree:          false,
					Currency:        strPtr("HKD"),
					FinalAmount:     8800,
					FinalFormatted:  strPtr("HK$ 88.00"),
					DiscountPercent: 10,
				},
			},
		},
	}

	svc := NewReadModelServiceWithReader(reader)
	res, err := svc.GetGameDetail(context.Background(), v2models.GameV2DetailRequest{GameID: 1, Lang: "zh", Region: "CN"})
	if err != nil {
		t.Fatalf("GetGameDetail returned error: %s", err.GetMsg())
	}

	if res.Price.Region != "CN" {
		t.Fatalf("expected selected price region CN, got %s", res.Price.Region)
	}
	if res.Price.Available {
		t.Fatal("expected CN price to be unavailable")
	}
	if res.Price.UnavailableReason != priceUnavailable {
		t.Fatalf("expected unavailable reason %s, got %s", priceUnavailable, res.Price.UnavailableReason)
	}
	if res.Price.Currency == "HKD" || res.Price.FinalAmount == 8800 {
		t.Fatal("expected CN price not to fallback to HK price")
	}
}

func TestGetGameDetailUsesLatestSuccessfulOnlineCount(t *testing.T) {
	collectedAt := time.Date(2026, 6, 7, 21, 30, 0, 0, time.UTC)
	reader := &fakeDetailReader{
		aggregate: v2models.GameV2Aggregate{
			Site: v2models.GameV2SiteRecord{ID: 1, AppID: 730, Name: "Counter-Strike 2"},
			OnlineCount: &v2models.GfgGameV2PlayerCount{
				GameID:      1,
				AppID:       730,
				Status:      "success",
				Count:       123456,
				CollectedAt: collectedAt,
			},
		},
	}

	svc := NewReadModelServiceWithReader(reader)
	res, err := svc.GetGameDetail(context.Background(), v2models.GameV2DetailRequest{GameID: 1, Lang: "zh"})
	if err != nil {
		t.Fatalf("GetGameDetail returned error: %s", err.GetMsg())
	}

	if res.OnlineCount.Status != "success" {
		t.Fatalf("expected online status success, got %s", res.OnlineCount.Status)
	}
	if res.OnlineCount.Count != 123456 {
		t.Fatalf("expected online count 123456, got %d", res.OnlineCount.Count)
	}
	if !res.OnlineCount.CollectedAt.Equal(collectedAt) {
		t.Fatalf("expected collected_at %s, got %s", collectedAt, res.OnlineCount.CollectedAt)
	}
}

func TestGetGameDetailDefaultsInvalidLangAndMissingOnlineCount(t *testing.T) {
	reader := &fakeDetailReader{
		aggregate: v2models.GameV2Aggregate{
			Site: v2models.GameV2SiteRecord{ID: 1, AppID: 440, Name: "军团要塞2"},
		},
	}

	svc := NewReadModelServiceWithReader(reader)
	res, err := svc.GetGameDetail(context.Background(), v2models.GameV2DetailRequest{GameID: 1, Lang: "fr"})
	if err != nil {
		t.Fatalf("GetGameDetail returned error: %s", err.GetMsg())
	}

	if reader.query.Lang != "zh" {
		t.Fatalf("expected normalized query lang zh, got %s", reader.query.Lang)
	}
	if res.Lang != "zh" {
		t.Fatalf("expected response lang zh, got %s", res.Lang)
	}
	if res.OnlineCount.Status != onlineUnknown {
		t.Fatalf("expected missing online count status %s, got %s", onlineUnknown, res.OnlineCount.Status)
	}
}

func TestGetPanelMainBuildsAllSections(t *testing.T) {
	reader := &fakeDetailReader{
		aggregate: v2models.GameV2Aggregate{
			Site: v2models.GameV2SiteRecord{
				ID:     1,
				AppID:  440,
				Name:   "军团要塞2",
				Info:   "中文站内简介",
				Header: "https://cdn.example/header.jpg",
			},
			Details: &v2models.GfgGameV2Details{
				GameID:     1,
				AppID:      440,
				Name:       "Team Fortress 2",
				Type:       "game",
				Developers: strPtr(`["Valve"]`),
				Publishers: strPtr(`["Valve"]`),
				Platforms:  strPtr(`{"windows":true,"mac":true,"linux":true}`),
			},
			Prices: []v2models.GfgGameV2Price{
				{
					GameID:         1,
					AppID:          440,
					Region:         "CN",
					IsFree:         true,
					FinalFormatted: strPtr("Free To Play"),
				},
			},
			Media: []v2models.GfgGameV2Media{
				{
					GameID:    1,
					AppID:     440,
					MediaType: "capsule",
					MediaKey:  "capsule",
					URL:       strPtr("https://cdn.example/capsule.jpg"),
				},
			},
		},
	}

	svc := NewReadModelServiceWithReader(reader)
	res, err := svc.GetPanelMain(context.Background(), v2models.GameV2PanelQuery{
		Lang:      "fr",
		Region:    "",
		Limit:     999,
		NewsLimit: 999,
	})
	if err != nil {
		t.Fatalf("GetPanelMain returned error: %s", err.GetMsg())
	}

	if reader.panelQuery.Lang != "zh" {
		t.Fatalf("expected normalized panel lang zh, got %s", reader.panelQuery.Lang)
	}
	if reader.panelQuery.Region != "CN" {
		t.Fatalf("expected normalized panel region CN, got %s", reader.panelQuery.Region)
	}
	if reader.panelQuery.Limit != 24 {
		t.Fatalf("expected panel limit clamp 24, got %d", reader.panelQuery.Limit)
	}
	if len(res.LatestGames) != 1 ||
		len(res.UpdatedGames) != 1 ||
		len(res.TopOnline) != 1 ||
		len(res.FreeGames) != 1 ||
		len(res.HighestDiscount) != 1 ||
		len(res.LowPrice) != 1 {
		t.Fatal("expected all panel game sections to contain one item")
	}
	if res.LatestGames[0].CapsuleURL != "https://cdn.example/capsule.jpg" {
		t.Fatalf("expected panel item capsule url, got %s", res.LatestGames[0].CapsuleURL)
	}
}

func strPtr(value string) *string {
	return &value
}
