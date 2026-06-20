package service

import (
	"context"
	"testing"
	"time"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	"github.com/gofurry/gofurry-game-backend/common"
	cm "github.com/gofurry/gofurry-game-backend/common/models"
)

type fakeDetailReader struct {
	aggregate      v2models.GameV2Aggregate
	query          v2models.GameV2DetailQuery
	listQuery      v2models.GameV2ListQuery
	listQueries    []v2models.GameV2ListQuery
	searchQuery    v2models.GameV2SearchPageQuery
	panelQuery     v2models.GameV2PanelQuery
	topOnlineQuery v2models.GameV2PanelQuery
	popularQuery   v2models.GameV2PanelQuery
	topPriceQuery  v2models.GameV2PanelQuery
	discountQuery  v2models.GameV2PanelQuery
	lowPriceQuery  v2models.GameV2PanelQuery
	searchItems    []v2models.GameV2SearchPageItem
	tags           []v2models.GameV2TagRecord
	reviews        v2models.GameV2ReviewList
	latestReviews  []v2models.GameV2LatestReview
	randomGameID   string
	similarRows    []v2models.GameV2RecommendationRow
	features       []v2models.GameV2RecommendationFeature
	savedRecs      []v2models.GfgGameV2Recommendation
	err            common.GFError
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
	reader.listQueries = append(reader.listQueries, query)
	if reader.err != nil {
		return nil, reader.err
	}
	return []v2models.GameV2Aggregate{reader.aggregate}, nil
}

func (reader *fakeDetailReader) SearchGames(_ context.Context, query v2models.GameV2SearchPageQuery) (cm.PageResponse, common.GFError) {
	reader.searchQuery = query
	if reader.err != nil {
		return cm.PageResponse{}, reader.err
	}
	return cm.PageResponse{
		Total: int64(len(reader.searchItems)),
		Data:  reader.searchItems,
	}, nil
}

func (reader *fakeDetailReader) ListTags(_ context.Context, _ string) ([]v2models.GameV2TagRecord, common.GFError) {
	if reader.err != nil {
		return nil, reader.err
	}
	return reader.tags, nil
}

func (reader *fakeDetailReader) GetGameReviews(_ context.Context, _ v2models.GameV2ReviewQuery) (v2models.GameV2ReviewList, common.GFError) {
	if reader.err != nil {
		return v2models.GameV2ReviewList{}, reader.err
	}
	return reader.reviews, nil
}

func (reader *fakeDetailReader) ListLatestReviews(_ context.Context, _ string, _ int) ([]v2models.GameV2LatestReview, common.GFError) {
	if reader.err != nil {
		return nil, reader.err
	}
	return reader.latestReviews, nil
}

func (reader *fakeDetailReader) GetRandomGameID(_ context.Context) (string, common.GFError) {
	if reader.err != nil {
		return "", reader.err
	}
	return reader.randomGameID, nil
}

func (reader *fakeDetailReader) ListSimilarRecommendations(_ context.Context, _ v2models.GameV2SimilarRecommendationQuery) ([]v2models.GameV2RecommendationRow, common.GFError) {
	if reader.err != nil {
		return nil, reader.err
	}
	return reader.similarRows, nil
}

func (reader *fakeDetailReader) SaveSimilarRecommendations(_ context.Context, _ int64, rows []v2models.GfgGameV2Recommendation) common.GFError {
	if reader.err != nil {
		return reader.err
	}
	reader.savedRecs = rows
	return nil
}

func (reader *fakeDetailReader) ListRecommendationFeatures(_ context.Context, _ string, _ string) ([]v2models.GameV2RecommendationFeature, common.GFError) {
	if reader.err != nil {
		return nil, reader.err
	}
	return reader.features, nil
}

func (reader *fakeDetailReader) GetGameNews(_ context.Context, _ v2models.GameV2NewsQuery) ([]v2models.GameV2NewsRow, common.GFError) {
	if reader.err != nil {
		return nil, reader.err
	}
	return []v2models.GameV2NewsRow{}, nil
}

func (reader *fakeDetailReader) ListTopOnlineAggregates(_ context.Context, query v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError) {
	reader.topOnlineQuery = query
	return reader.ListGameAggregates(context.Background(), v2models.GameV2ListQuery{})
}

func (reader *fakeDetailReader) ListPopularGameAggregates(_ context.Context, query v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError) {
	reader.popularQuery = query
	return reader.ListGameAggregates(context.Background(), v2models.GameV2ListQuery{})
}

func (reader *fakeDetailReader) ListFreeGameAggregates(_ context.Context, query v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError) {
	reader.panelQuery = query
	return reader.ListGameAggregates(context.Background(), v2models.GameV2ListQuery{})
}

func (reader *fakeDetailReader) ListHighestPriceAggregates(_ context.Context, query v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError) {
	reader.topPriceQuery = query
	return reader.ListGameAggregates(context.Background(), v2models.GameV2ListQuery{})
}

func (reader *fakeDetailReader) ListHighestDiscountAggregates(_ context.Context, query v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError) {
	reader.discountQuery = query
	return reader.ListGameAggregates(context.Background(), v2models.GameV2ListQuery{})
}

func (reader *fakeDetailReader) ListLowPriceAggregates(_ context.Context, query v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError) {
	reader.lowPriceQuery = query
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
	if res.Name != "Team Fortress 2" {
		t.Fatalf("expected english site name, got %s", res.Name)
	}
	if res.Summary != "English site summary" {
		t.Fatalf("expected english site summary, got %s", res.Summary)
	}
	if res.ShortDescription != "中文采集简介" {
		t.Fatalf("expected localized short description to stay available, got %s", res.ShortDescription)
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

func TestSimpleSearchNormalizesQuery(t *testing.T) {
	reader := &fakeDetailReader{
		searchItems: []v2models.GameV2SearchPageItem{
			{
				ID:    "1",
				Name:  "Team Fortress 2",
				Info:  "Class-based action",
				Cover: "https://cdn.example/header.jpg",
			},
		},
	}

	svc := NewReadModelServiceWithReader(reader)
	res, err := svc.SimpleSearch(context.Background(), v2models.GameV2SearchRequest{
		Txt:  "  furry game  ",
		Lang: "en-US",
	})
	if err != nil {
		t.Fatalf("SimpleSearch returned error: %s", err.GetMsg())
	}
	if reader.searchQuery.Lang != "en" {
		t.Fatalf("expected normalized lang en, got %s", reader.searchQuery.Lang)
	}
	if reader.searchQuery.Content != "furry game" {
		t.Fatalf("expected trimmed search content, got %q", reader.searchQuery.Content)
	}
	if reader.searchQuery.PageNum != 1 || reader.searchQuery.PageSize != 8 {
		t.Fatalf("expected simple search page 1 size 8, got page=%d size=%d", reader.searchQuery.PageNum, reader.searchQuery.PageSize)
	}
	if len(res) != 1 || res[0].ID != "1" || res[0].Cover == "" {
		t.Fatalf("expected mapped simple search result, got %#v", res)
	}
}

func TestGetGameReviewsDesensitizesIP(t *testing.T) {
	reader := &fakeDetailReader{
		reviews: v2models.GameV2ReviewList{
			Total:    1,
			AvgScore: 4.5,
			Remarks: []v2models.GameV2ReviewItem{
				{
					Region:  "Local",
					Content: "good",
					Score:   4.5,
					IP:      "192.168.1.42",
				},
			},
		},
	}

	svc := NewReadModelServiceWithReader(reader)
	res, err := svc.GetGameReviews(context.Background(), "1", 1, 5)
	if err != nil {
		t.Fatalf("GetGameReviews returned error: %s", err.GetMsg())
	}
	if res.Total != 1 || len(res.Remarks) != 1 {
		t.Fatalf("expected one review, got %#v", res)
	}
	if res.Remarks[0].IP == "192.168.1.42" {
		t.Fatal("expected review IP to be desensitized")
	}
}

func TestGetPanelMainBuildsAllSections(t *testing.T) {
	reader := &fakeDetailReader{
		aggregate: v2models.GameV2Aggregate{
			Site: v2models.GameV2SiteRecord{
				ID:     1,
				AppID:  440,
				Name:   "军团要塞2",
				NameEn: "Team Fortress 2",
				Info:   "中文站内简介",
				InfoEn: "English site summary",
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
			ReviewStats: v2models.GameV2ReviewStats{
				AvgScore:     4.2,
				CommentCount: 7,
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
	if reader.topOnlineQuery.Limit != 60 {
		t.Fatalf("expected top online limit clamp 60, got %d", reader.topOnlineQuery.Limit)
	}
	if reader.popularQuery.Limit != 24 {
		t.Fatalf("expected popular games limit clamp 24, got %d", reader.popularQuery.Limit)
	}
	if reader.topPriceQuery.Region != "US" || reader.topPriceQuery.Limit != 15 {
		t.Fatalf("expected top price to use US limit 15, got region=%s limit=%d", reader.topPriceQuery.Region, reader.topPriceQuery.Limit)
	}
	if reader.discountQuery.Region != "US" || reader.discountQuery.Limit != 15 {
		t.Fatalf("expected discount to use US limit 15, got region=%s limit=%d", reader.discountQuery.Region, reader.discountQuery.Limit)
	}
	if reader.lowPriceQuery.Region != "US" || reader.lowPriceQuery.Limit != 120 {
		t.Fatalf("expected low price to use US limit 120, got region=%s limit=%d", reader.lowPriceQuery.Region, reader.lowPriceQuery.Limit)
	}
	if len(res.LatestGames) != 1 ||
		len(res.UpdatedGames) != 1 ||
		len(res.TopOnline) != 1 ||
		len(res.PopularGames) != 1 ||
		len(res.FreeGames) != 1 ||
		len(res.TopPrice) != 1 ||
		len(res.HighestDiscount) != 1 ||
		len(res.LowPrice) != 1 {
		t.Fatal("expected all panel game sections to contain one item")
	}
	if len(reader.listQueries) < 2 || reader.listQueries[0].Sort != "release_date" || reader.listQueries[1].Sort != "newest" {
		t.Fatalf("expected latest sort release_date and updated sort newest, got %+v", reader.listQueries)
	}
	if res.LatestGames[0].CapsuleURL != "https://cdn.example/capsule.jpg" {
		t.Fatalf("expected panel item capsule url, got %s", res.LatestGames[0].CapsuleURL)
	}
	if res.LatestGames[0].AvgScore != 4.2 || res.LatestGames[0].CommentCount != 7 {
		t.Fatalf("expected panel item review stats, got avg_score=%v comment_count=%d", res.LatestGames[0].AvgScore, res.LatestGames[0].CommentCount)
	}
	if res.LatestGames[0].NameZh != "军团要塞2" || res.LatestGames[0].NameEn != "Team Fortress 2" {
		t.Fatalf("expected panel item bilingual names from site table, got zh=%s en=%s", res.LatestGames[0].NameZh, res.LatestGames[0].NameEn)
	}
	if res.LatestGames[0].SummaryZh != "中文站内简介" || res.LatestGames[0].SummaryEn != "English site summary" {
		t.Fatalf("expected panel item bilingual summaries from site table, got zh=%s en=%s", res.LatestGames[0].SummaryZh, res.LatestGames[0].SummaryEn)
	}
}

func TestGetSimilarRecommendationsComputesAndSavesHybridScore(t *testing.T) {
	tagsA := `[{"id":"1001","name":"RPG","prefix":"1000"},{"id":"2001","name":"Wolf","prefix":"2000"}]`
	tagsB := `[{"id":"1001","name":"RPG","prefix":"1000"},{"id":"2001","name":"Wolf","prefix":"2000"}]`
	tagsC := `[{"id":"3001","name":"Windows","prefix":"3000"}]`
	developersA := `["Studio A"]`
	developersB := `["Studio A"]`
	developersC := `["Studio C"]`
	platforms := `{"windows":true}`

	reader := &fakeDetailReader{
		features: []v2models.GameV2RecommendationFeature{
			{GameID: 1, AppID: 1001, Name: "Source", Summary: "furry rpg", Tags: &tagsA, Developers: &developersA, Platforms: &platforms, PrimaryTagID: 1001, PriceRegion: "CN", PriceAvailable: true, FinalAmount: 1000, OnlineCount: 100},
			{GameID: 2, AppID: 1002, Name: "Strong Match", Summary: "furry rpg", Tags: &tagsB, Developers: &developersB, Platforms: &platforms, PrimaryTagID: 1001, LibraryCoverURL: "library.jpg", LibraryCover2xURL: "library_2x.jpg", PriceRegion: "CN", PriceAvailable: true, FinalAmount: 1200, OnlineCount: 90},
			{GameID: 3, AppID: 1003, Name: "Weak Match", Summary: "space puzzle", Tags: &tagsC, Developers: &developersC, Platforms: &platforms, PriceRegion: "CN", PriceAvailable: true, FinalAmount: 8000, OnlineCount: 3},
		},
	}

	res, err := NewReadModelServiceWithReader(reader).GetSimilarRecommendations(context.Background(), v2models.GameV2SimilarRecommendationQuery{
		GameID: 1,
		Lang:   "zh",
		Region: "CN",
		Limit:  2,
	})
	if err != nil {
		t.Fatalf("expected no error, got %s", err.GetMsg())
	}
	if len(res) == 0 {
		t.Fatal("expected recommendations")
	}
	if res[0].ID != "2" {
		t.Fatalf("expected strong tag match first, got %s", res[0].ID)
	}
	if res[0].LibraryCoverURL != "library.jpg" || res[0].LibraryCover2xURL != "library_2x.jpg" {
		t.Fatalf("expected library covers to be carried through, got %#v", res[0])
	}
	if len(res[0].Reasons) == 0 || res[0].Reasons[0].Type != "tag" {
		t.Fatalf("expected tag reason first, got %#v", res[0].Reasons)
	}
	if len(reader.savedRecs) == 0 {
		t.Fatal("expected computed recommendations to be saved")
	}
	if reader.savedRecs[0].AlgorithmVersion != similarRecommendationAlgorithmVersion {
		t.Fatalf("expected algorithm version %s, got %s", similarRecommendationAlgorithmVersion, reader.savedRecs[0].AlgorithmVersion)
	}
}

func strPtr(value string) *string {
	return &value
}
