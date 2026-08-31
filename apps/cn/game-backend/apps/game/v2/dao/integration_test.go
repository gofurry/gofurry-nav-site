package dao_test

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	v2dao "github.com/gofurry/gofurry-game-backend/apps/game/v2/dao"
	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	v2service "github.com/gofurry/gofurry-game-backend/apps/game/v2/service"
	prizedao "github.com/gofurry/gofurry-game-backend/apps/prize/dao"
	prizemodels "github.com/gofurry/gofurry-game-backend/apps/prize/models"
	reviewdao "github.com/gofurry/gofurry-game-backend/apps/review/dao"
	reviewmodels "github.com/gofurry/gofurry-game-backend/apps/review/models"
	"github.com/gofurry/gofurry-game-backend/common"
	cm "github.com/gofurry/gofurry-game-backend/common/models"
	gamesqlc "github.com/gofurry/gofurry-game-backend/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"gopkg.in/yaml.v3"
)

type integrationDatabaseConfig struct {
	Name     string `yaml:"db_name"`
	Username string `yaml:"db_username"`
	Password string `yaml:"db_password"`
	Host     string `yaml:"db_host"`
	Port     string `yaml:"db_port"`
}

func TestPostgresReadModelSemantics(t *testing.T) {
	configPath := os.Getenv("GOFURRY_GAME_BACKEND_INTEGRATION_CONFIG")
	if configPath == "" {
		t.Skip("set GOFURRY_GAME_BACKEND_INTEGRATION_CONFIG for PostgreSQL integration tests")
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Database integrationDatabaseConfig `yaml:"database"`
		DataBase integrationDatabaseConfig `yaml:"data_base"`
	}
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Database.Host == "" {
		cfg.Database = cfg.DataBase
	}
	baseDSN := integrationDSN(cfg.Database)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	adminDB := integrationSQLDB(t, baseDSN, "postgres")
	defer adminDB.Close()
	databaseName := integrationDatabaseName()
	createIntegrationDatabase(t, ctx, adminDB, databaseName)
	defer dropIntegrationDatabase(t, adminDB, databaseName)
	testDSN := integrationDatabaseDSN(baseDSN, databaseName)
	applyGameBaseline(t, testDSN)
	pool, err := pgxpool.New(ctx, testDSN)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	seedReadModel(t, ctx, pool, now)
	seedGameInsights(t, ctx, pool, now)
	assertGameInsights(t, ctx, pool)
	readDAO := v2dao.NewReadModelDAO(pool)

	detail, gfErr := readDAO.GetGameDetailAggregate(ctx, v2models.GameV2DetailQuery{GameID: 91001, Lang: "en", NewsLimit: 5})
	if gfErr != nil {
		t.Fatalf("detail: %s", gfErr.GetMsg())
	}
	if detail.Site.ID != 91001 || detail.Localized == nil || detail.Localized.Name != "English Game" || len(detail.Prices) != 1 || len(detail.Tags) != 1 {
		t.Fatalf("unexpected detail aggregate: %+v", detail)
	}
	if detail.ReleaseState == nil || detail.ReleaseState.Availability != "available" || detail.ReleaseState.ExactDate == nil || *detail.ReleaseState.ExactDate != "2026-08-01" {
		t.Fatalf("unexpected structured release state: %+v", detail.ReleaseState)
	}
	if detail.FirstAvailable == nil || detail.FirstAvailable.Precision != "month" || detail.FirstAvailable.ExactDate != nil || detail.FirstAvailable.WindowStart != "2026-08-01" || detail.FirstAvailable.WindowEnd != "2026-08-31" || len(detail.Languages) != 2 || detail.Languages[1].Code != nil || detail.Languages[1].SteamName != "Klingon" {
		t.Fatalf("unexpected canonical domain aggregate: first=%+v languages=%+v", detail.FirstAvailable, detail.Languages)
	}
	if _, gfErr = readDAO.GetGameDetailAggregate(ctx, v2models.GameV2DetailQuery{GameID: 999999}); gfErr == nil {
		t.Fatal("missing detail should preserve not-found error behavior")
	}

	list, gfErr := readDAO.ListGameAggregates(ctx, v2models.GameV2ListQuery{Lang: "zh", Limit: 1, Sort: "weight"})
	if gfErr != nil || len(list) != 1 || list[0].Site.ID != 91001 {
		t.Fatalf("list: rows=%+v err=%v", list, gfErr)
	}
	newestGames, gfErr := readDAO.ListGameAggregates(ctx, v2models.GameV2ListQuery{Lang: "zh", Limit: 10, Sort: "newest"})
	if gfErr != nil || len(newestGames) != 2 || newestGames[0].Site.ID != 91002 || newestGames[0].ReleaseState == nil || newestGames[0].ReleaseState.Availability != "upcoming" {
		t.Fatalf("recently collected must include upcoming games: rows=%+v err=%v", newestGames, gfErr)
	}
	latestGames, gfErr := readDAO.ListGameAggregates(ctx, v2models.GameV2ListQuery{Lang: "zh", Limit: 10, Sort: "release_date"})
	if gfErr != nil || len(latestGames) != 1 || latestGames[0].Site.ID != 91001 || latestGames[0].FirstAvailable == nil {
		t.Fatalf("canonical latest games: rows=%+v err=%v", latestGames, gfErr)
	}
	page, gfErr := readDAO.SearchGames(ctx, v2models.GameV2SearchPageQuery{Lang: "en", Content: "English", PageNum: 1, PageSize: 1})
	if gfErr != nil || page.Total != 1 {
		t.Fatalf("search: page=%+v err=%v", page, gfErr)
	}
	items, ok := page.Data.([]v2models.GameV2SearchPageItem)
	if !ok || len(items) != 1 || items[0].AppID != 92001 || items[0].Name != "English Game" {
		t.Fatalf("unexpected search items: %#v", page.Data)
	}
	if items[0].Release == nil || items[0].Release.Availability != "available" || items[0].FirstAvailable == nil || items[0].FirstAvailable.Precision != "month" || items[0].FirstAvailable.ExactDate != nil || items[0].FirstAvailable.WindowEnd != "2026-08-31" {
		t.Fatalf("search item missing structured first available: %#v", items[0])
	}
	upcomingPage, gfErr := readDAO.SearchGames(ctx, v2models.GameV2SearchPageQuery{
		Lang: "en", Content: "Second", Availability: "upcoming", TimeOrder: true,
		PlannedStartTime: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
		PlannedEndTime:   time.Date(2026, 11, 30, 23, 59, 59, 0, time.UTC),
		PageNum:          1, PageSize: 10,
	})
	if gfErr != nil || upcomingPage.Total != 1 {
		t.Fatalf("upcoming planned-window search: page=%+v err=%v", upcomingPage, gfErr)
	}
	upcomingItems := upcomingPage.Data.([]v2models.GameV2SearchPageItem)
	if len(upcomingItems) != 1 || upcomingItems[0].ID != "91002" || upcomingItems[0].Release == nil || upcomingItems[0].Release.Precision != "quarter" || upcomingItems[0].FirstAvailable != nil {
		t.Fatalf("unexpected upcoming search items: %#v", upcomingItems)
	}
	intervalPage, gfErr := readDAO.SearchGames(ctx, v2models.GameV2SearchPageQuery{
		Lang: "zh", PubStartTime: time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC),
		PubEndTime: time.Date(2026, 8, 15, 23, 59, 59, 0, time.UTC), PageNum: 1, PageSize: 10,
	})
	if gfErr != nil || intervalPage.Total != 1 || intervalPage.Data.([]v2models.GameV2SearchPageItem)[0].ID != "91001" {
		t.Fatalf("canonical first-available interval search: page=%+v err=%v", intervalPage, gfErr)
	}
	tagPage, gfErr := readDAO.SearchGames(ctx, v2models.GameV2SearchPageQuery{Lang: "zh", TagList: []int64{1}, PageNum: 2, PageSize: 1})
	if gfErr != nil || tagPage.Total != 2 || len(tagPage.Data.([]v2models.GameV2SearchPageItem)) != 1 {
		t.Fatalf("tag search pagination: %+v err=%v", tagPage, gfErr)
	}

	tags, gfErr := readDAO.ListTags(ctx, "en")
	if gfErr != nil || len(tags) != 1 || tags[0].Name != "Adventure" || tags[0].GameCount != 2 {
		t.Fatalf("tags: %+v err=%v", tags, gfErr)
	}
	reviews, gfErr := readDAO.GetGameReviews(ctx, v2models.GameV2ReviewQuery{GameID: 91001, PageNum: 1, PageSize: 1})
	if gfErr != nil || reviews.Total != 2 || len(reviews.Remarks) != 1 || reviews.Remarks[0].Content != "new review" {
		t.Fatalf("reviews: %+v err=%v", reviews, gfErr)
	}
	latest, gfErr := readDAO.ListLatestReviews(ctx, "en", 1)
	if gfErr != nil || len(latest) != 1 || latest[0].GameName != "English Game" {
		t.Fatalf("latest reviews: %+v err=%v", latest, gfErr)
	}
	news, gfErr := readDAO.GetGameNews(ctx, v2models.GameV2NewsQuery{GameID: 91001, Lang: "en", Limit: 5})
	if gfErr != nil || len(news) != 1 || news[0].GameName != "中文游戏" {
		t.Fatalf("game news: %+v err=%v", news, gfErr)
	}
	latestNews, gfErr := readDAO.GetLatestGameNews(ctx, v2models.GameV2NewsQuery{Lang: "en", Limit: 5})
	if gfErr != nil || len(latestNews) != 1 {
		t.Fatalf("latest news: %+v err=%v", latestNews, gfErr)
	}
	if randomID, gfErr := readDAO.GetRandomGameID(ctx); gfErr != nil || (randomID != "91001" && randomID != "91002") {
		t.Fatalf("random game: id=%q err=%v", randomID, gfErr)
	}
	panelQuery := v2models.GameV2PanelQuery{Lang: "zh", Region: "CN", Limit: 2}
	for name, load := range map[string]func(context.Context, v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError){
		"top-online": readDAO.ListTopOnlineAggregates,
		"popular":    readDAO.ListPopularGameAggregates,
		"top-price":  readDAO.ListHighestPriceAggregates,
		"discount":   readDAO.ListHighestDiscountAggregates,
		"low-price":  readDAO.ListLowPriceAggregates,
	} {
		if rows, gfErr := load(ctx, panelQuery); gfErr != nil || len(rows) == 0 {
			t.Fatalf("%s panel: rows=%d err=%v", name, len(rows), gfErr)
		}
	}

	recommendations := []v2models.GfgGameV2Recommendation{{
		SourceGameID: 91001, TargetGameID: 91002, Score: .8, DisplayScore: .8,
		Rank: 1, ReasonJSON: `[{"type":"tag"}]`, AlgorithmVersion: "test-v1", ComputedAt: now,
	}}
	if gfErr := readDAO.SaveSimilarRecommendations(ctx, 91001, recommendations); gfErr != nil {
		t.Fatalf("save recommendations: %s", gfErr.GetMsg())
	}
	recRows, gfErr := readDAO.ListSimilarRecommendations(ctx, v2models.GameV2SimilarRecommendationQuery{
		GameID: 91001, Lang: "en", Region: "CN", Limit: 8, AlgorithmVersion: "test-v1",
	})
	if gfErr != nil || len(recRows) != 1 || recRows[0].TargetGameID != 91002 || recRows[0].Name != "Second Game" {
		t.Fatalf("recommendations: %+v err=%v", recRows, gfErr)
	}
	features, gfErr := readDAO.ListRecommendationFeatures(ctx, "en", "CN")
	if gfErr != nil || len(features) != 2 {
		t.Fatalf("recommendation features: %d err=%v", len(features), gfErr)
	}
	if gfErr := readDAO.SaveSimilarRecommendations(ctx, 91001, nil); gfErr != nil {
		t.Fatalf("clear recommendations: %s", gfErr.GetMsg())
	}
	if recRows, gfErr = readDAO.ListSimilarRecommendations(ctx, v2models.GameV2SimilarRecommendationQuery{GameID: 91001, AlgorithmVersion: "test-v1"}); gfErr != nil || len(recRows) != 0 {
		t.Fatalf("cleared recommendations: %+v err=%v", recRows, gfErr)
	}

	prizeDAO := prizedao.New(pool)
	var prize prizemodels.GfgPrize
	if gfErr := prizeDAO.GetById(93001, &prize); gfErr != nil {
		t.Fatalf("get prize: %s", gfErr.GetMsg())
	}
	prize.Status = false
	if affected, gfErr := prizeDAO.Save(prize.ID, prize); gfErr != nil || affected != 1 {
		t.Fatalf("save false prize status: affected=%d err=%v", affected, gfErr)
	}
	var statusValue bool
	if err := pool.QueryRow(ctx, `SELECT status FROM gfg_prize WHERE id=$1`, prize.ID).Scan(&statusValue); err != nil || statusValue {
		t.Fatalf("false prize status was not persisted: status=%v err=%v", statusValue, err)
	}
	member := prizemodels.GfgPrizeMember{ID: 94001, PrizeID: prize.ID, Name: "member", Email: "one@example.test", IP: "127.0.0.1", Agent: "test", CreateTime: cm.LocalTime(now)}
	if gfErr := prizeDAO.Add(&member); gfErr != nil {
		t.Fatalf("add prize member: %s", gfErr.GetMsg())
	}
	if gfErr := prizeDAO.Add(&member); gfErr == nil || gfErr.GetMsg() != "数据重复，入库失败" {
		t.Fatalf("unique violation mapping: %v", gfErr)
	}

	reviewDAO := reviewdao.New(pool)
	if _, gfErr := reviewDAO.GetReviewByIPAndName("91001", "missing", "missing"); gfErr == nil || gfErr.GetMsg() != common.RETURN_RECORD_NOT_FOUND {
		t.Fatalf("review no-row mapping: %v", gfErr)
	}
	newReview := reviewmodels.GfgGameComment{ID: 95001, Region: "test", Content: "inserted", Score: 4, CreateTime: cm.LocalTime(now), GameID: 91002, IP: "127.0.0.2", Name: "tester"}
	if gfErr := reviewDAO.Add(&newReview); gfErr != nil {
		t.Fatalf("insert review: %s", gfErr.GetMsg())
	}
}

func seedGameInsights(t *testing.T, ctx context.Context, pool *pgxpool.Pool, now time.Time) {
	t.Helper()
	latestDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1)
	_, err := pool.Exec(ctx, `
INSERT INTO gfg_game_tracking_periods
    (id,game_id,appid,tracked_from,tracking_basis,opened_reason)
VALUES (99001,91001,92001,$1::timestamptz - interval '60 days','explicit','insights_test');
INSERT INTO gfg_game_daily
    (game_id,fact_date,tracking_period_id,appid,snapshot_at,tracked_at_end,name,name_en,view_count,
     is_free,windows,linux,release_availability,developers,publishers,tag_ids,details_observed_at,
     materialization_source,projection_version,finalized_at)
VALUES (91001,$2::date,99001,92001,$1,true,'中文游戏','English Game',3,false,true,true,'available',
        ARRAY['Dev'],ARRAY['Pub'],ARRAY[1::bigint],$1,'observed',1,$1);
INSERT INTO gfg_metric_daily
    (metric_key,metric_version,fact_date,dimension_key,dimension_value,population_count,eligible_count,
     not_applicable_count,positive_count,negative_count,stale_count,not_probed_count,probe_failed_count,unknown_count,computed_at)
VALUES
    ('free_game_share',1,$3::date,'global','all',10,10,0,2,8,0,0,0,0,$1),
    ('free_game_share',1,$2::date,'global','all',10,10,0,5,5,0,0,0,0,$1),
    ('windows_support',1,$2::date,'global','all',10,10,0,8,2,0,0,0,0,$1);
INSERT INTO gfg_game_player_counts (run_id,game_id,appid,count,status,collected_at)
VALUES ('insights-zero',91001,92001,0,'success',$1::timestamptz + interval '1 second'),
       ('insights-failed',91001,92001,0,'failure',$1::timestamptz + interval '2 seconds');
INSERT INTO gfg_game_player_daily
    (tracking_period_id,game_id,appid,fact_date,min_players,max_players,avg_players,median_players,
     attempted_samples,successful_samples,partial_samples,failed_samples,failure_kind_counts,
     quality_basis,projection_version,finalized_at)
VALUES
    (99001,91001,92001,$2::date - 1,0,0,0,0,1,1,0,0,'{}','legacy_observed_only',1,$1),
    (99001,91001,92001,$2::date, NULL,NULL,NULL,NULL,1,0,0,1,'{"request_failed":1}','legacy_observed_only',1,$1);
INSERT INTO gfg_game_price_daily
    (tracking_period_id,game_id,appid,region,fact_date,price_state,currency,initial_amount,final_amount,discount_percent,
     materialization_source,projection_version,finalized_at)
VALUES
    (99001,91001,92001,'CN',$2::date - 1,'free',NULL,NULL,NULL,NULL,'observed',1,$1),
    (99001,91001,92001,'CN',$2::date,'priced','CNY',1000,0,100,'observed',1,$1),
    (99001,91001,92001,'HK',$2::date + 1,'priced','HKD',2000,1800,10,'observed',1,$1);
INSERT INTO gfg_change_events
    (event_key,detector_key,detector_version,game_id,projection_date,event_at,time_basis,event_code,scope_kind,scope_key,
     old_value,new_value,source_event_key,source_before_key,source_after_key,source_versions,materialized_at)
VALUES
    ('insights-free-old','free_game_transition',1,91001,$2::date - 1,NULL,'day','game_became_paid','global','all','{}','{}','game-src-old','before-old','after-old','{}',$1),
    ('insights-free-new','free_game_transition',1,91001,$2::date,NULL,'day','game_became_free','global','all','{}','{}','game-src-new','before-new','after-new','{}',$1),
    ('insights-price-cn','game_price_transition',1,91001,$2::date,NULL,'day','game_price_decreased','region','CN','{}','{}','game-src-price-cn','before-cn','after-cn','{}',$1),
    ('insights-price-hk','game_price_transition',1,91001,$2::date + 1,NULL,'day','game_price_increased','region','HK','{}','{}','game-src-price-hk','before-hk','after-hk','{}',$1);
`, pgx.QueryExecModeSimpleProtocol, now, latestDay, latestDay.AddDate(0, 0, -30))
	if err != nil {
		t.Fatal(err)
	}
}

func assertGameInsights(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	queries := gamesqlc.New(pool)
	svc := v2service.NewInsightsService(v2dao.NewInsightsDAO(queries))
	overview, err := svc.GetInsightsOverview(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(overview.Metrics) != 2 || overview.Metrics[0].Key != "free" || overview.Metrics[0].Delta30D == nil || *overview.Metrics[0].Delta30D != .3 {
		t.Fatalf("metric mapping/exact delta: %#v", overview.Metrics)
	}
	if len(overview.RecentChanges) != 1 || overview.RecentChanges[0].Type != "game.price.decreased" {
		t.Fatalf("overview dedup/CN whitelist: %#v", overview.RecentChanges)
	}
	summary, err := svc.GetGameInsights(ctx, 91001)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Players.Current == nil || *summary.Players.Current != 0 || summary.Players.Peak30D == nil || *summary.Players.Peak30D != 0 {
		t.Fatalf("real zero player summary: %#v", summary.Players)
	}
	if summary.Price == nil || summary.Price.Region != "CN" || summary.Price.State != "priced" || summary.Price.FinalAmount == nil || *summary.Price.FinalAmount != 0 {
		t.Fatalf("CN priced-zero summary: %#v", summary.Price)
	}
	players, err := svc.GetGamePlayerInsights(ctx, 91001, "all")
	if err != nil {
		t.Fatal(err)
	}
	if len(players.Points) != 1 || players.Points[0].Max != 0 {
		t.Fatalf("failed day became zero or real zero was lost: %#v", players.Points)
	}
	prices, err := svc.GetGamePriceInsights(ctx, 91001, "all")
	if err != nil {
		t.Fatal(err)
	}
	if len(prices.Points) != 2 || prices.Points[0].State != "free" || prices.Points[0].FinalAmount != nil || prices.Points[1].State != "priced" || prices.Points[1].FinalAmount == nil || *prices.Points[1].FinalAmount != 0 {
		t.Fatalf("free/priced-zero/CN-only semantics: %#v", prices.Points)
	}
}

func seedReadModel(t *testing.T, ctx context.Context, pool *pgxpool.Pool, now time.Time) {
	t.Helper()
	statements := []string{
		`INSERT INTO gfg_tag (id,name,name_en,info,info_en,prefix,create_time,update_time) VALUES (1,'冒险','Adventure','冒险','Adventure',0,$1,$1)`,
		`INSERT INTO gfg_game (id,name,name_en,info,info_en,create_time,update_time,resources,groups,release_date,developers,publishers,appid,header,links,weight,primary_tag,secondary_tag,view_count) VALUES
(91001,'中文游戏','English Game','中文简介','English summary',$1,$1,'[]','[]','2026-08-01','["Dev"]','["Pub"]',92001,'header-1','[]',1,1,0,3),
(91002,'第二游戏','Second Game','第二简介','Second summary',$1,$1,'[]','[]','2026-08-02','["Dev"]','["Pub"]',92002,'header-2','[]',2,1,0,1)`,
		`INSERT INTO gfg_tag_map (id,game_id,tag_id,create_time,update_time) VALUES (1,91001,1,$1,$1),(2,91002,1,$1,$1)`,
		`INSERT INTO gfg_game_details (game_id,appid,source,type,name,is_free,developers,publishers,release_date_text,platforms,collected_at,updated_at) VALUES
(91001,92001,'steam','game','English Game',false,'["Dev"]','["Pub"]','2026-08-01','{"windows":true}',$1,$1),
(91002,92002,'steam','game','Second Game',false,'["Dev"]','["Pub"]','2026-08-02','{"windows":true}',$1,$1)`,
		`INSERT INTO gfg_game_localized_details (game_id,appid,lang,name,short_description,collected_at,updated_at) VALUES
(91001,92001,'zh','中文游戏','中文简介',$1,$1),(91001,92001,'en','English Game','English summary',$1,$1),
(91002,92002,'zh','第二游戏','第二简介',$1,$1),(91002,92002,'en','Second Game','Second summary',$1,$1)`,
		`INSERT INTO gfg_game_prices (game_id,appid,region,is_free,currency,initial_amount,final_amount,discount_percent,initial_formatted,final_formatted,collected_at,updated_at) VALUES
(91001,92001,'CN',false,'CNY',1000,800,20,'¥10','¥8',$1,$1),(91002,92002,'CN',false,'CNY',2000,1800,10,'¥20','¥18',$1,$1)`,
		`INSERT INTO gfg_game_assets (game_id,appid,asset_type,asset_family,source,lang,media_key,title,url,thumbnail_url,format,exists,collected_at,updated_at) VALUES
(91001,92001,'header','store','store_browse','','header','','asset-header-1','','jpg',true,$1,$1),
(91002,92002,'header','store','store_browse','','header','','asset-header-2','','jpg',true,$1,$1)`,
		`INSERT INTO gfg_game_requirements (game_id,appid,pc,mac,linux,collected_at,updated_at) VALUES (91001,92001,'{}','{}','{}',$1,$1),(91002,92002,'{}','{}','{}',$1,$1)`,
		`INSERT INTO gfg_game_comment (id,region,content,score,create_time,game_id,ip,name) VALUES (1,'CN','old review',3,$2,91001,'127.0.0.1','old'),(2,'CN','new review',5,$1,91001,'127.0.0.2','new')`,
		`INSERT INTO gfg_game_news (game_id,appid,lang,event_gid,headline,summary,collected_at,published_at,updated_at) VALUES (91001,92001,'en','event-1','news','summary',$1,$1,$1)`,
		`INSERT INTO gfg_game_player_counts (run_id,game_id,appid,count,status,collected_at) VALUES ('run-1',91001,92001,42,'success',$1)`,
		`INSERT INTO gfg_game_release_state (game_id,availability,precision,exact_date,release_year,release_month,release_quarter,window_start,window_end,raw_text,source,source_region,source_locale,normalizer_version,observed_at) VALUES
(91001,'available','day','2026-08-01',2026,8,NULL,'2026-08-01','2026-08-01','1 Aug, 2026','steam','US','en','steam-go/v1.3.9',$1),
(91002,'upcoming','quarter',NULL,2026,NULL,4,'2026-10-01','2026-12-31','Q4 2026','steam','US','en','steam-go/v1.3.9',$1)`,
		`INSERT INTO gfg_game_first_available (game_id,precision,exact_date,release_year,release_month,window_start,window_end,source,inferred,source_raw,source_observed_at,normalizer_version) VALUES
(91001,'month',NULL,2026,8,'2026-08-01','2026-08-31','legacy_manual',false,'August 2026',$1,'gofurry-legacy-release/v1')`,
		`INSERT INTO gfg_game_languages (game_id,language_code,steam_name,steam_api_code,steam_web_code,tier,full_audio_supported,sort_order,source,source_region,source_locale,normalizer_version,observed_at) VALUES
(91001,'en','English','english','en','platform',true,0,'steam','US','en','steam-go/v1.3.9',$1),
(91001,NULL,'Klingon',NULL,NULL,'unknown',NULL,1,'steam','US','en','steam-go/v1.3.9',$1)`,
		`INSERT INTO gfg_prize (id,title,"desc",prize,"key",start_time,end_time,create_time,status) VALUES (93001,'Prize','Description','{"title":"Prize","platform":"Steam","keys":["key"]}','secret',$1,$2,$1,true)`,
	}
	for i, statement := range statements {
		args := []any{now}
		switch i {
		case 8:
			args = append(args, now.Add(-time.Hour))
		case 14:
			args = append(args, now.Add(24*time.Hour))
		}
		if _, err := pool.Exec(ctx, statement, args...); err != nil {
			t.Fatalf("seed statement %d: %v", i, err)
		}
	}
}

func integrationDSN(cfg integrationDatabaseConfig) string {
	u := &url.URL{Scheme: "postgres", Host: cfg.Host + ":" + cfg.Port, Path: cfg.Name}
	u.User = url.UserPassword(cfg.Username, cfg.Password)
	query := u.Query()
	query.Set("sslmode", "prefer")
	u.RawQuery = query.Encode()
	return u.String()
}

func integrationDatabaseDSN(baseDSN, databaseName string) string {
	u, err := url.Parse(baseDSN)
	if err != nil {
		panic(err)
	}
	u.Path = databaseName
	return u.String()
}

func integrationSQLDB(t *testing.T, dsn, databaseName string) *sql.DB {
	t.Helper()
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.Database = databaseName
	return stdlib.OpenDB(*config)
}

func integrationDatabaseName() string {
	var value [5]byte
	if _, err := rand.Read(value[:]); err != nil {
		panic(err)
	}
	return "gofurry_game_backend_it_" + hex.EncodeToString(value[:])
}

func createIntegrationDatabase(t *testing.T, ctx context.Context, adminDB *sql.DB, name string) {
	t.Helper()
	validateIntegrationDatabaseName(t, name)
	if _, err := adminDB.ExecContext(ctx, `CREATE DATABASE `+quoteIntegrationIdentifier(name)); err != nil {
		t.Fatal(err)
	}
}

func dropIntegrationDatabase(t *testing.T, adminDB *sql.DB, name string) {
	t.Helper()
	validateIntegrationDatabaseName(t, name)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := adminDB.ExecContext(ctx, `DROP DATABASE IF EXISTS `+quoteIntegrationIdentifier(name)+` WITH (FORCE)`); err != nil {
		t.Errorf("drop temporary database: %v", err)
	}
}

func validateIntegrationDatabaseName(t *testing.T, name string) {
	t.Helper()
	if !strings.HasPrefix(name, "gofurry_game_backend_it_") || strings.ContainsAny(name, `"'; \`) {
		t.Fatalf("unsafe temporary database name %q", name)
	}
}

func quoteIntegrationIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func applyGameBaseline(t *testing.T, dsn string) {
	t.Helper()
	repositoryRoot := integrationRepositoryRoot(t)
	command := exec.Command("go", "tool", "goose", "-dir", filepath.Join(repositoryRoot, "db", "game", "migrations"), "postgres", dsn, "up")
	command.Dir = filepath.Join(repositoryRoot, "tools")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("apply Game baseline: %v\n%s", err, output)
	}
}

func integrationRepositoryRoot(t *testing.T) string {
	t.Helper()
	current, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(current, "sqlc.yaml")); err == nil {
			if _, err := os.Stat(filepath.Join(current, "tools", "go.mod")); err == nil {
				return current
			}
		}
		parent := filepath.Dir(current)
		if parent == current {
			t.Fatal("repository root containing sqlc.yaml and tools/go.mod not found")
		}
		current = parent
	}
}
