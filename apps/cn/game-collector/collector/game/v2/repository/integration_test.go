package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gofurry/gofurry-game-collector/collector/facts"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/backfill"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	metricengine "github.com/gofurry/gofurry-game-collector/collector/metrics"
	gamesqlc "github.com/gofurry/gofurry-game-collector/internal/db/game/sqlc"
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

func TestPostgresRepositorySemantics(t *testing.T) {
	configPath := os.Getenv("GOFURRY_GAME_COLLECTOR_INTEGRATION_CONFIG")
	if configPath == "" {
		t.Skip("set GOFURRY_GAME_COLLECTOR_INTEGRATION_CONFIG for PostgreSQL integration tests")
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Database integrationDatabaseConfig `yaml:"data_base"`
	}
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		t.Fatal(err)
	}
	baseDSN := integrationDSN(cfg.Database)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	adminDB := integrationSQLDB(t, baseDSN, "postgres")
	defer adminDB.Close()
	databaseName := integrationDatabaseName()
	createIntegrationDatabase(t, ctx, adminDB, databaseName)
	defer dropIntegrationDatabase(t, adminDB, databaseName)

	testConfig, err := pgxpool.ParseConfig(baseDSN)
	if err != nil {
		t.Fatal(err)
	}
	testConfig.ConnConfig.Database = databaseName
	testDSN := integrationDatabaseDSN(baseDSN, databaseName)
	applyGameBaseline(t, testDSN)
	pool, err := pgxpool.NewWithConfig(ctx, testConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		t.Fatal(err)
	}

	clock, err := gamesqlc.New(pool).GameCollectionClock(ctx)
	if err != nil || !clock.Valid {
		t.Fatalf("read integration database clock: %v", err)
	}
	now := clock.Time.UTC().Truncate(time.Microsecond)
	seedGameTarget(t, ctx, pool, now)
	targets, err := gamesqlc.New(pool).ListGameTargets(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 1 || targets[0].ID != 91001 || targets[0].Appid != 92001 {
		t.Fatalf("unexpected target list: %+v", targets)
	}

	detailsRepository := NewDetailsRepository(pool)
	invalid := detailsFixture(now, []domain.RawSnapshot{{
		GameID: 91001, AppID: 92001, Language: domain.StoreLocale("zh-CN"),
		Region: domain.Region("CN"), Source: domain.Source("steam"), RawPayload: []byte("{"), CollectedAt: now,
	}})
	if err := detailsRepository.SaveDetails(ctx, invalid); err == nil {
		t.Fatal("invalid snapshot JSON should roll back the details transaction")
	}
	assertCount(t, ctx, pool, `select count(*) from gfg_game_details where game_id=$1`, 0, int64(91001))

	snapshots := make([]domain.RawSnapshot, 6)
	for i := range snapshots {
		snapshots[i] = domain.RawSnapshot{
			GameID: 91001, AppID: 92001, Language: domain.StoreLocale("zh-CN"),
			Region: domain.Region("CN"), Source: domain.Source("steam"),
			RawPayload: []byte(fmt.Sprintf(`{"sequence":%d}`, i)), CollectedAt: now.Add(time.Duration(i) * time.Second),
		}
	}
	initial := detailsFixture(now, snapshots)
	initial.CanonicalRelease = releaseFixture(now, domain.ReleaseUpcoming, domain.ReleasePrecisionMonth, "September 2026", time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC))
	initial.CanonicalLanguages = languageFixture(now, "en", "English")
	if err := detailsRepository.SaveDetails(ctx, initial); err != nil {
		t.Fatalf("save details: %v", err)
	}
	assertCount(t, ctx, pool, `select count(*) from gfg_game_detail_snapshots where game_id=$1`, 5, int64(91001))
	assertCount(t, ctx, pool, `select count(*) from gfg_game_media where game_id=$1`, 1, int64(91001))
	assertCount(t, ctx, pool, `select count(*) from gfg_game_assets where game_id=$1`, 1, int64(91001))
	assertCount(t, ctx, pool, `select count(*) from gfg_game_release_state where game_id=$1 and availability='upcoming'`, 1, int64(91001))
	assertCount(t, ctx, pool, `select count(*) from gfg_game_release_history where game_id=$1`, 1, int64(91001))
	assertCount(t, ctx, pool, `select count(*) from gfg_game_first_available where game_id=$1`, 0, int64(91001))
	assertCount(t, ctx, pool, `select count(*) from gfg_game_languages where game_id=$1`, 1, int64(91001))
	assertCount(t, ctx, pool, `select count(*) from gfg_game_price_daily where game_id=$1 and price_state='priced' and materialization_source='observed' and observed_at=$2`, 1, int64(91001), now)
	assertCount(t, ctx, pool, `select count(*) from gfg_game_daily where game_id=$1 and fact_date=($2 at time zone 'UTC')::date and materialization_source='observed'`, 1, int64(91001), now)

	// A raw-text-only observation updates current provenance but not semantic history.
	rawOnly := detailsFixture(now.Add(time.Minute), nil)
	rawOnly.CanonicalRelease = releaseFixture(now.Add(time.Minute), domain.ReleaseUpcoming, domain.ReleasePrecisionMonth, "Sep 2026", time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC))
	rawOnly.CanonicalLanguages = languageFixture(now.Add(time.Minute), "zh-Hans", "Simplified Chinese")
	if err := detailsRepository.SaveDetails(ctx, rawOnly); err != nil {
		t.Fatalf("save semantic-equivalent details: %v", err)
	}
	assertCount(t, ctx, pool, `select count(*) from gfg_game_release_history where game_id=$1`, 1, int64(91001))
	assertCount(t, ctx, pool, `select count(*) from gfg_game_languages where game_id=$1 and language_code='zh-Hans'`, 1, int64(91001))
	assertCount(t, ctx, pool, `select count(*) from gfg_game_languages where game_id=$1 and language_code='en'`, 0, int64(91001))

	// A non-authoritative collection preserves both canonical release and languages.
	if err := detailsRepository.SaveDetails(ctx, detailsFixture(now.Add(2*time.Minute), nil)); err != nil {
		t.Fatalf("save non-authoritative details: %v", err)
	}
	assertCount(t, ctx, pool, `select count(*) from gfg_game_release_history where game_id=$1`, 1, int64(91001))
	assertCount(t, ctx, pool, `select count(*) from gfg_game_languages where game_id=$1 and language_code='zh-Hans'`, 1, int64(91001))

	// The observed upcoming -> available transition establishes a non-inferred fact.
	releasedDate := time.Date(2026, time.August, 24, 0, 0, 0, 0, time.UTC)
	released := detailsFixture(now.Add(3*time.Minute), nil)
	released.CanonicalRelease = releaseFixture(now.Add(3*time.Minute), domain.ReleaseAvailable, domain.ReleasePrecisionDay, "24 Aug, 2026", releasedDate, releasedDate)
	released.CanonicalRelease.ExactDate = &releasedDate
	if err := detailsRepository.SaveDetails(ctx, released); err != nil {
		t.Fatalf("save available details: %v", err)
	}
	assertCount(t, ctx, pool, `select count(*) from gfg_game_release_history where game_id=$1`, 2, int64(91001))
	assertCount(t, ctx, pool, `select count(*) from gfg_game_first_available where game_id=$1 and source='observed_transition' and inferred=false and exact_date='2026-08-24'`, 1, int64(91001))

	// Later observations can change current state/history but never rewrite First Available.
	laterDate := time.Date(2026, time.August, 25, 0, 0, 0, 0, time.UTC)
	later := detailsFixture(now.Add(4*time.Minute), nil)
	later.CanonicalRelease = releaseFixture(now.Add(4*time.Minute), domain.ReleaseAvailable, domain.ReleasePrecisionDay, "25 Aug, 2026", laterDate, laterDate)
	later.CanonicalRelease.ExactDate = &laterDate
	if err := detailsRepository.SaveDetails(ctx, later); err != nil {
		t.Fatalf("save later available details: %v", err)
	}
	assertCount(t, ctx, pool, `select count(*) from gfg_game_release_history where game_id=$1`, 3, int64(91001))
	assertCount(t, ctx, pool, `select count(*) from gfg_game_first_available where game_id=$1 and exact_date='2026-08-24'`, 1, int64(91001))

	news := domain.GameNews{
		GameID: 91001, AppID: 92001, Language: domain.StoreLocale("zh-CN"), EventGID: "event-1",
		AnnouncementGID: "announcement-1", Headline: "headline", CollectedAt: now,
	}
	newsRepository := NewNewsRepository(pool)
	if err := newsRepository.SaveNews(ctx, []domain.GameNews{news, news}); err != nil {
		t.Fatalf("upsert news: %v", err)
	}
	assertCount(t, ctx, pool, `select count(*) from gfg_game_news where game_id=$1`, 1, int64(91001))

	playerRepository := NewPlayerRepository(pool)
	if err := playerRepository.SavePlayerCount(ctx, domain.PlayerCount{RunID: "current", GameID: 91001, AppID: 92001, Count: 42, Status: domain.StatusSuccess, CollectedAt: now}); err != nil {
		t.Fatalf("insert current player count: %v", err)
	}
	old := now.Add(-48 * time.Hour)
	if err := playerRepository.SavePlayerCount(ctx, domain.PlayerCount{RunID: "old", GameID: 91001, AppID: 92001, Count: 1, Status: domain.StatusSuccess, CollectedAt: old}); err != nil {
		t.Fatalf("insert old player count: %v", err)
	}

	if err := NewRetentionRepository(pool).Prune(ctx, RetentionConfig{PlayerCountsDays: 1, CollectRunsDays: 1, CollectTaskResultsDays: 1}); err != nil {
		t.Fatalf("prune retention: %v", err)
	}
	// P0.1.1 freezes destructive player-count pruning until the P0.2
	// retention/aggregation design is available.
	assertCount(t, ctx, pool, `select count(*) from gfg_game_player_counts where run_id=$1`, 1, "old")

	if _, err := pool.Exec(ctx, `update gfg_game set appid=92002 where id=91001`); err != nil {
		t.Fatal(err)
	}
	if err := detailsRepository.SaveDetails(ctx, detailsFixture(now.Add(5*time.Minute), nil)); !errors.Is(err, ErrStaleCollection) {
		t.Fatalf("expected stale collection error, got %v", err)
	}

	seedAdditionalGameTarget(t, ctx, pool, 91003, 92003, now, "")
	directDate := time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC)
	direct := minimalDetailsFixture(91003, 92003, now)
	direct.CanonicalRelease = releaseFixtureForGame(91003, now, domain.ReleaseAvailable, domain.ReleasePrecisionDay, "2 Jan, 2020", directDate, directDate)
	direct.CanonicalRelease.ExactDate = &directDate
	if err := detailsRepository.SaveDetails(ctx, direct); err != nil {
		t.Fatalf("save first observed available details: %v", err)
	}
	assertCount(t, ctx, pool, `select count(*) from gfg_game_first_available where game_id=$1 and source='steam_backfill' and inferred=true`, 1, int64(91003))

	seedAdditionalGameTarget(t, ctx, pool, 91004, 92004, now, "")
	unknown := minimalDetailsFixture(91004, 92004, now)
	unknown.CanonicalRelease = &domain.GameReleaseState{GameID: 91004, Availability: domain.ReleaseAvailable, Precision: domain.ReleasePrecisionUnknown, RawText: "Fall 2020", Source: domain.SourceSteam, SourceRegion: domain.RegionUS, SourceLocale: domain.StoreLocaleEN, Normalizer: "steam-go/v1.3.9", ObservedAt: now}
	if err := detailsRepository.SaveDetails(ctx, unknown); err != nil {
		t.Fatalf("save unknown precision details: %v", err)
	}
	assertCount(t, ctx, pool, `select count(*) from gfg_game_first_available where game_id=$1`, 0, int64(91004))

	seedAdditionalGameTarget(t, ctx, pool, 91005, 92005, now, "")
	rollback := minimalDetailsFixture(91005, 92005, now)
	rollback.CanonicalRelease = releaseFixtureForGame(91005, now, domain.ReleaseAvailable, domain.ReleasePrecisionDay, "2 Jan, 2020", directDate, directDate)
	rollback.CanonicalRelease.ExactDate = &directDate
	rollback.Snapshots = []domain.RawSnapshot{{GameID: 91005, AppID: 92005, Language: domain.StoreLocaleEN, Region: domain.RegionUS, Source: domain.SourceSteam, RawPayload: []byte("{"), CollectedAt: now}}
	if err := detailsRepository.SaveDetails(ctx, rollback); err == nil {
		t.Fatal("expected invalid snapshot to roll back canonical writes")
	}
	assertCount(t, ctx, pool, `select count(*) from gfg_game_release_state where game_id=$1`, 0, int64(91005))

	seedAdditionalGameTarget(t, ctx, pool, 91002, 93002, now, "2020.01.02")
	if _, err := pool.Exec(ctx, `insert into gfg_game_details (game_id,appid,release_coming_soon,collected_at) values (91002,93002,false,$1)`, now); err != nil {
		t.Fatal(err)
	}
	seedAdditionalGameTarget(t, ctx, pool, 91006, 93006, now, "2020")
	seedAdditionalGameTarget(t, ctx, pool, 91007, 93007, now, "2020")
	seedAdditionalGameTarget(t, ctx, pool, 91008, 93008, now, "Spring 2020")
	seedAdditionalGameTarget(t, ctx, pool, 91009, 93009, now, "2999-01-01")
	for _, candidate := range []struct {
		gameID, appID int64
		comingSoon    bool
	}{{91007, 93007, true}, {91008, 93008, false}, {91009, 93009, false}} {
		if _, err := pool.Exec(ctx, `insert into gfg_game_details (game_id,appid,release_coming_soon,collected_at) values ($1,$2,$3,$4)`, candidate.gameID, candidate.appID, candidate.comingSoon, now); err != nil {
			t.Fatal(err)
		}
	}
	drySummary, err := backfill.New(pool).Run(ctx, true)
	if err != nil || drySummary.Scanned != 5 || drySummary.Eligible != 1 || drySummary.Inserted != 0 || drySummary.SkippedNoCurrentState != 1 || drySummary.SkippedUpcoming != 1 || drySummary.SkippedInvalid != 1 || drySummary.SkippedFuture != 1 {
		t.Fatalf("unexpected dry-run summary: %+v err=%v", drySummary, err)
	}
	actualSummary, err := backfill.New(pool).Run(ctx, false)
	if err != nil || actualSummary.Inserted != 1 {
		t.Fatalf("unexpected backfill summary: %+v err=%v", actualSummary, err)
	}
	assertCount(t, ctx, pool, `select count(*) from gfg_game_first_available where game_id=$1 and source='legacy_manual' and inferred=false and normalizer_version='gofurry-legacy-release/v1'`, 1, int64(91002))
	testHistoricalPlayerFacts(t, ctx, pool)
	testHistoricalGameMetrics(t, ctx, pool)
}

func testHistoricalGameMetrics(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	day := time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC)
	dayEnd := day.Add(24 * time.Hour)
	if _, err := pool.Exec(ctx, `
INSERT INTO gfg_game_tracking_periods
    (game_id,appid,tracked_from,tracking_basis,opened_reason)
SELECT game_id, game_id + 1000, $1::timestamptz - interval '1 day', 'legacy_observed', 'metric_integration'
FROM unnest(ARRAY[99200,99201,99202,99203,99204,99205]::bigint[]) game_id;

INSERT INTO gfg_game_daily (
	game_id,fact_date,tracking_period_id,appid,snapshot_at,tracked_at_end,
	name,name_en,view_count,is_free,windows,linux,release_availability,
	release_observed_at,primary_tag_id,tag_ids,details_observed_at,developers,publishers,
	materialization_source,projection_version,finalized_at,created_at,updated_at
)
SELECT input.game_id,$1::date,period.id,input.game_id + 1000,$2,true,
	   'metric game ' || input.game_id,'metric game ' || input.game_id,0,
	   input.is_free,input.windows,input.linux,input.availability,input.release_observed_at,
	   input.primary_tag_id,input.tag_ids,input.details_observed_at,
	   ARRAY[]::text[],ARRAY[]::text[],'observed',1,$2,$2,$2
FROM (VALUES
	(99200::bigint,true, true, false,'available'::text,100::bigint,ARRAY[100,101,101]::bigint[],$2::timestamptz - interval '1 hour',$2::timestamptz - interval '10 hours'),
	(99201::bigint,false,false,true, 'available'::text,NULL::bigint,ARRAY[]::bigint[],$2::timestamptz - interval '2 hours',$2::timestamptz - interval '11 hours'),
	(99202::bigint,false,true, true, 'upcoming'::text,102::bigint,ARRAY[102]::bigint[],$2::timestamptz - interval '1 hour',$2::timestamptz - interval '6 hours'),
	(99203::bigint,true, true, true, 'unknown'::text,NULL::bigint,ARRAY[]::bigint[],$2::timestamptz - interval '1 hour',$2::timestamptz - interval '7 hours'),
	(99204::bigint,true, true, true, 'available'::text,104::bigint,ARRAY[104]::bigint[],$2::timestamptz - interval '4 days',$2::timestamptz - interval '12 hours'),
	(99205::bigint,NULL::boolean,NULL::boolean,NULL::boolean,'available'::text,NULL::bigint,ARRAY[]::bigint[],NULL::timestamptz,$2::timestamptz - interval '9 hours')
) input(game_id,is_free,windows,linux,availability,primary_tag_id,tag_ids,details_observed_at,release_observed_at)
JOIN gfg_game_tracking_periods period ON period.game_id=input.game_id;

UPDATE gfg_fact_rollup_checkpoints
SET source_start_date=$1::date,processed_through=$1::date,updated_at=$2
WHERE pipeline_key='game.state_facts';
UPDATE gfg_metric_checkpoints
SET source_start_date=$1::date,processed_through=NULL,updated_at=$2;
`, pgx.QueryExecModeSimpleProtocol, day, dayEnd); err != nil {
		t.Fatal(err)
	}

	engine := metricengine.New(pool, metricengine.Options{})
	if err := engine.ValidateCatalog(ctx); err != nil {
		t.Fatalf("Game registry/evaluator drift: %v", err)
	}
	assertCount(t, ctx, pool, `SELECT count(*) FROM gfg_metric_registry`, 4)
	assertCount(t, ctx, pool, `SELECT count(*) FROM gfg_metric_checkpoints`, 4)

	// Two contenders serialize on the per-version checkpoint; exactly one can
	// advance the only upstream-ready day.
	type concurrentResult struct {
		processed bool
		err       error
	}
	results := make(chan concurrentResult, 2)
	for range 2 {
		go func() {
			result, err := engine.RunNext(ctx, "free_game_share", 1, false)
			results <- concurrentResult{processed: result.Processed, err: err}
		}()
	}
	processed := 0
	for range 2 {
		result := <-results
		if result.err != nil {
			t.Fatalf("concurrent Game checkpoint run: %v", result.err)
		}
		if result.processed {
			processed++
		}
	}
	if processed != 1 {
		t.Fatalf("concurrent Game checkpoint processed=%d, want 1", processed)
	}
	for _, key := range []string{"windows_support", "linux_support"} {
		result, err := engine.RunNext(ctx, key, 1, false)
		if err != nil || !result.Processed {
			t.Fatalf("run Game metric %s: result=%+v err=%v", key, result, err)
		}
	}

	wantFree := map[int64]string{99200: "positive/game_is_free", 99201: "negative/game_is_paid", 99202: "not_applicable/game_not_available", 99203: "unknown/game_availability_unknown", 99204: "stale/details_state_stale", 99205: "unknown/is_free_unknown"}
	for gameID, expected := range wantFree {
		var state, reason string
		if err := pool.QueryRow(ctx, `SELECT state,reason_code FROM gfg_metric_entity_daily WHERE metric_key='free_game_share' AND metric_version=1 AND fact_date=$1 AND game_id=$2`, day, gameID).Scan(&state, &reason); err != nil {
			t.Fatal(err)
		}
		if state+"/"+reason != expected {
			t.Fatalf("free_game_share game=%d got=%s/%s want=%s", gameID, state, reason, expected)
		}
	}
	wantEvidence := map[int64]time.Time{
		99200: dayEnd.Add(-time.Hour),
		99201: dayEnd.Add(-2 * time.Hour),
		99202: dayEnd.Add(-6 * time.Hour),
		99203: dayEnd.Add(-7 * time.Hour),
		99204: dayEnd.Add(-4 * 24 * time.Hour),
	}
	for gameID, expected := range wantEvidence {
		var observed time.Time
		if err := pool.QueryRow(ctx, `SELECT source_observed_at FROM gfg_metric_entity_daily WHERE metric_key='free_game_share' AND metric_version=1 AND fact_date=$1 AND game_id=$2`, day, gameID).Scan(&observed); err != nil {
			t.Fatal(err)
		}
		if !observed.Equal(expected) {
			t.Fatalf("free_game_share provenance game=%d got=%s want=%s", gameID, observed, expected)
		}
	}
	var missingEvidence *time.Time
	if err := pool.QueryRow(ctx, `SELECT source_observed_at FROM gfg_metric_entity_daily WHERE metric_key='free_game_share' AND metric_version=1 AND fact_date=$1 AND game_id=99205`, day).Scan(&missingEvidence); err != nil {
		t.Fatal(err)
	}
	if missingEvidence != nil {
		t.Fatalf("is_free_unknown provenance=%s, want NULL details evidence", *missingEvidence)
	}

	var population, eligible, notApplicable, positive, negative, stale, unknown int64
	if err := pool.QueryRow(ctx, `
SELECT population_count,eligible_count,not_applicable_count,positive_count,negative_count,stale_count,unknown_count
FROM gfg_metric_daily
WHERE metric_key='free_game_share' AND metric_version=1 AND fact_date=$1
  AND dimension_key='global' AND dimension_value='all'`, day).Scan(
		&population, &eligible, &notApplicable, &positive, &negative, &stale, &unknown,
	); err != nil {
		t.Fatal(err)
	}
	if population != 6 || eligible != 5 || notApplicable != 1 || positive != 1 || negative != 1 || stale != 1 || unknown != 2 {
		t.Fatalf("unexpected Game global counts pop=%d eligible=%d na=%d +%d -%d stale=%d unknown=%d", population, eligible, notApplicable, positive, negative, stale, unknown)
	}
	assertCount(t, ctx, pool, `SELECT population_count FROM gfg_metric_daily WHERE metric_key='free_game_share' AND fact_date=$1 AND dimension_key='tag_id' AND dimension_value='101'`, 1, day)
	assertCount(t, ctx, pool, `SELECT population_count FROM gfg_metric_daily WHERE metric_key='free_game_share' AND fact_date=$1 AND dimension_key='primary_tag_id' AND dimension_value='unknown'`, 3, day)

	if _, err := engine.Rebuild(ctx, "windows_support", 1, day, day, 1, true); err != nil {
		t.Fatalf("Game metric rebuild dry-run: %v", err)
	}
	if _, err := engine.Rebuild(ctx, "windows_support", 1, day, day, 1, false); err != nil {
		t.Fatalf("Game metric rebuild: %v", err)
	}
	assertCount(t, ctx, pool, `SELECT count(*) FROM gfg_metric_entity_daily WHERE metric_key='windows_support' AND fact_date=$1`, 6, day)

	wantSupport := map[string]map[int64]string{
		"windows_support": {
			99200: "positive/windows_supported", 99201: "negative/windows_not_supported",
			99202: "positive/windows_supported", 99203: "positive/windows_supported",
			99204: "stale/details_state_stale", 99205: "unknown/windows_support_unknown",
		},
		"linux_support": {
			99200: "negative/linux_not_supported", 99201: "positive/linux_supported",
			99202: "positive/linux_supported", 99203: "positive/linux_supported",
			99204: "stale/details_state_stale", 99205: "unknown/linux_support_unknown",
		},
	}
	for metricKey, entities := range wantSupport {
		for gameID, expected := range entities {
			var state, reason string
			if err := pool.QueryRow(ctx, `SELECT state,reason_code FROM gfg_metric_entity_daily WHERE metric_key=$1 AND metric_version=1 AND fact_date=$2 AND game_id=$3`, metricKey, day, gameID).Scan(&state, &reason); err != nil {
				t.Fatal(err)
			}
			if state+"/"+reason != expected {
				t.Fatalf("%s game=%d got=%s/%s want=%s", metricKey, gameID, state, reason, expected)
			}
		}
	}

	// Runtime failure of one metric must not block another active metric.
	nextDay := day.AddDate(0, 0, 1)
	if _, err := pool.Exec(ctx, `
DELETE FROM gfg_metric_checkpoints WHERE metric_key='free_game_share' AND metric_version=1;
INSERT INTO gfg_game_daily (
 game_id,fact_date,tracking_period_id,appid,snapshot_at,tracked_at_end,name,name_en,view_count,
 is_free,windows,linux,release_availability,tag_ids,details_observed_at,developers,publishers,
 materialization_source,projection_version,finalized_at,created_at,updated_at
)
SELECT 99200,$1::date,id,100200,$2,true,'metric game 99200','metric game 99200',0,
       true,true,false,'available',ARRAY[]::bigint[],$2::timestamptz - interval '1 hour',ARRAY[]::text[],ARRAY[]::text[],
       'observed',1,$2,$2,$2
FROM gfg_game_tracking_periods WHERE game_id=99200;
UPDATE gfg_fact_rollup_checkpoints SET processed_through=$1::date WHERE pipeline_key='game.state_facts';
`, pgx.QueryExecModeSimpleProtocol, nextDay, nextDay.Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := engine.Reconcile(ctx); err == nil {
		t.Fatal("Game metric reconcile should report the missing free_game_share checkpoint")
	}
	assertCount(t, ctx, pool, `SELECT count(*) FROM gfg_metric_entity_daily WHERE metric_key='windows_support' AND fact_date=$1`, 1, nextDay)

	// Future source evidence rolls back the whole metric day and leaves its checkpoint unchanged.
	futureDay := nextDay.AddDate(0, 0, 1)
	if _, err := pool.Exec(ctx, `
INSERT INTO gfg_metric_checkpoints (metric_key,metric_version,source_start_date,processed_through)
VALUES ('free_game_share',1,$2::date - 1,$2::date);
INSERT INTO gfg_game_daily (
 game_id,fact_date,tracking_period_id,appid,snapshot_at,tracked_at_end,name,name_en,view_count,
 is_free,windows,linux,release_availability,release_observed_at,tag_ids,details_observed_at,developers,publishers,
 materialization_source,projection_version,finalized_at,created_at,updated_at
)
SELECT 99200,$1::date,id,100200,$3,true,'metric game 99200','metric game 99200',0,
	   true,true,false,'upcoming',$3::timestamptz + interval '1 second',ARRAY[]::bigint[],$3::timestamptz - interval '1 hour',ARRAY[]::text[],ARRAY[]::text[],
	   'observed',1,$3,$3,$3
FROM gfg_game_tracking_periods WHERE game_id=99200;
UPDATE gfg_fact_rollup_checkpoints SET processed_through=$1::date WHERE pipeline_key='game.state_facts';
`, pgx.QueryExecModeSimpleProtocol, futureDay, nextDay, futureDay.Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.RunNext(ctx, "free_game_share", 1, false); err == nil {
		t.Fatal("future Game release evidence did not fail")
	}
	assertCount(t, ctx, pool, `SELECT count(*) FROM gfg_metric_entity_daily WHERE metric_key='free_game_share' AND fact_date=$1`, 0, futureDay)
	assertCount(t, ctx, pool, `SELECT count(*) FROM gfg_metric_checkpoints WHERE metric_key='free_game_share' AND processed_through=$1::date`, 1, nextDay)
	if _, err := pool.Exec(ctx, `
UPDATE gfg_game_daily
SET release_availability='available',
    release_observed_at=$2::timestamptz - interval '1 hour',
    details_observed_at=$2::timestamptz + interval '1 second'
WHERE game_id=99200 AND fact_date=$1::date`, futureDay, futureDay.Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.RunNext(ctx, "free_game_share", 1, false); err == nil {
		t.Fatal("future Game details evidence did not fail")
	}
	assertCount(t, ctx, pool, `SELECT count(*) FROM gfg_metric_entity_daily WHERE metric_key='free_game_share' AND fact_date=$1`, 0, futureDay)
	assertCount(t, ctx, pool, `SELECT count(*) FROM gfg_metric_checkpoints WHERE metric_key='free_game_share' AND processed_through=$1::date`, 1, nextDay)
}

func testHistoricalPlayerFacts(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	day := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	if _, err := pool.Exec(ctx, `
INSERT INTO gfg_game_tracking_periods (game_id,appid,tracked_from,tracked_until,tracking_basis,opened_reason,closed_reason)
VALUES (99001,99002,$1,$2,'legacy_observed','integration','integration'),
       (99003,99004,$1,$2,'explicit','integration','integration');
INSERT INTO gfg_game_daily (
    game_id,fact_date,tracking_period_id,appid,snapshot_at,tracked_at_end,
    name,name_en,view_count,developers,publishers,tag_ids,
    materialization_source,projection_version,created_at,updated_at
)
SELECT 99003,$1::date,id,99004,$2,true,'state fixture','state fixture',0,
       ARRAY[]::text[],ARRAY[]::text[],ARRAY[]::bigint[],'observed',1,$1,$1
FROM gfg_game_tracking_periods WHERE game_id=99003;
INSERT INTO gfg_game_price_daily (
    tracking_period_id,game_id,appid,region,fact_date,price_state,currency,
    initial_amount,final_amount,discount_percent,observed_at,
    materialization_source,projection_version,created_at,updated_at
)
SELECT id,99003,99004,'CN',$1::date,'priced','CNY',100,80,20,$1,
       'observed',1,$1,$1
FROM gfg_game_tracking_periods WHERE game_id=99003;
INSERT INTO gfg_collection_schedules (job_key,name,enabled,schedule_kind,interval_seconds,anchor_at,timezone,misfire_policy,misfire_grace_seconds,priority,concurrency_key,next_scheduled_for)
VALUES ('game.players','facts fixture',true,'interval',3600,$1,'UTC','skip',0,1,'steam',$2)
ON CONFLICT (job_key) DO UPDATE SET enabled=true,next_scheduled_for=EXCLUDED.next_scheduled_for;
INSERT INTO gfg_collection_schedules (job_key,name,enabled,schedule_kind,interval_seconds,anchor_at,timezone,misfire_policy,misfire_grace_seconds,priority,concurrency_key,next_scheduled_for)
VALUES ('game.metadata','state facts fixture',true,'interval',3600,$1,'UTC','skip',0,1,'steam',$2)
ON CONFLICT (job_key) DO UPDATE SET enabled=true,next_scheduled_for=EXCLUDED.next_scheduled_for;
INSERT INTO gfg_collector_instances (instance_id,collector_id,hostname,version,commit_sha,capabilities,started_at,last_heartbeat_at)
VALUES ('facts-fixture','facts','localhost','test','',ARRAY['game.players'],$1,$2)
ON CONFLICT (instance_id) DO NOTHING;
UPDATE gfg_fact_rollup_checkpoints SET source_start_date=$1::date,processed_through=NULL,quality_cutover_at=$1 WHERE pipeline_key='game.player_facts';
UPDATE gfg_fact_rollup_checkpoints SET source_start_date=$1::date,processed_through=NULL,quality_cutover_at=$1 WHERE pipeline_key='game.state_facts';
`, pgx.QueryExecModeSimpleProtocol, day, day.Add(25*time.Hour)); err != nil {
		t.Fatal(err)
	}

	counts := []int64{1, 100, 2}
	for index := 0; index < 4; index++ {
		slot := day.Add(time.Duration(index) * time.Hour)
		jobID := int64(99010 + index)
		runID := fmt.Sprintf("facts-player-%d", index)
		status := "success"
		if index == 3 {
			status = "failed"
		}
		if _, err := pool.Exec(ctx, `
INSERT INTO gfg_collection_jobs (id,schedule_id,schedule_version,job_key,trigger,scope_type,tasks,priority,concurrency_key,scheduled_for,status,requested_by,created_at,updated_at,completed_at)
SELECT $1,id,version,'game.players','scheduled','all',ARRAY['players'],1,'steam',$2,'success','fixture',$2,$2,$2 FROM gfg_collection_schedules WHERE job_key='game.players';
INSERT INTO gfg_collection_runs (id,job_id,attempt_no,collector_instance_id,status,scheduled_for,started_at,ended_at,expected_count,attempted_count,success_count,partial_count,failure_count,skipped_count)
VALUES ($3,$1,1,'facts-fixture',CASE WHEN $4='success' THEN 'success' ELSE 'failed' END,$2,$2,$2,1,1,CASE WHEN $4='success' THEN 1 ELSE 0 END,0,CASE WHEN $4='failed' THEN 1 ELSE 0 END,0);
INSERT INTO gfg_collection_task_results (run_id,task_type,game_id,appid,status,error_kind,started_at,ended_at)
VALUES ($3,'players',99001,99002,$4,CASE WHEN $4='failed' THEN 'upstream' ELSE '' END,$2,$2);
`, pgx.QueryExecModeSimpleProtocol, jobID, slot, runID, status); err != nil {
			t.Fatal(err)
		}
		if index < len(counts) {
			if _, err := pool.Exec(ctx, `INSERT INTO gfg_game_player_counts (game_id,appid,count,status,collected_at,run_id) VALUES (99001,99002,$1,'success',$2,$3)`, counts[index], slot.Add(time.Minute), runID); err != nil {
				t.Fatal(err)
			}
		}
	}

	retentionEngine := facts.New(pool, facts.Options{Now: func() time.Time { return day.Add(26 * time.Hour) }, FinalizationGrace: time.Minute, RetentionEnabled: true, PlayerRawAge: time.Hour, RetentionBatch: 100})
	if deleted, err := retentionEngine.PrunePlayerRaw(ctx); err != nil || deleted != 0 {
		t.Fatalf("retention without checkpoint deleted=%d err=%v", deleted, err)
	}
	engine := facts.New(pool, facts.Options{Now: func() time.Time { return day.Add(26 * time.Hour) }, FinalizationGrace: time.Minute, RetentionEnabled: false})
	result, err := engine.RunNext(ctx, facts.PlayerPipeline, false)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Processed {
		t.Fatalf("player facts not processed: %+v", result)
	}
	var median float64
	var expected, attempted, succeeded, failed int
	if err := pool.QueryRow(ctx, `SELECT median_players,expected_samples,attempted_samples,successful_samples,failed_samples FROM gfg_game_player_daily WHERE game_id=99001 AND fact_date=$1`, day).Scan(&median, &expected, &attempted, &succeeded, &failed); err != nil {
		t.Fatal(err)
	}
	if median != 2 || expected != 4 || attempted != 4 || succeeded != 3 || failed != 1 {
		t.Fatalf("unexpected daily fact median=%v expected=%d attempted=%d success=%d failure=%d", median, expected, attempted, succeeded, failed)
	}
	if _, err := engine.Rebuild(ctx, facts.PlayerPipeline, day, day, true); err != nil {
		t.Fatalf("player rebuild dry-run: %v", err)
	}
	for attempt := 0; attempt < 2; attempt++ {
		if _, err := engine.Rebuild(ctx, facts.PlayerPipeline, day, day, false); err != nil {
			t.Fatalf("idempotent player rebuild attempt %d: %v", attempt+1, err)
		}
	}
	stateResult, err := engine.RunNext(ctx, facts.StatePipeline, false)
	if err != nil || !stateResult.Processed {
		t.Fatalf("state facts result=%+v err=%v", stateResult, err)
	}
	assertCount(t, ctx, pool, `select count(*) from gfg_game_daily where game_id=99003 and fact_date=$1 and finalized_at is not null`, 1, day)
	assertCount(t, ctx, pool, `select count(*) from gfg_game_price_daily where game_id=99003 and fact_date=$1`, 3, day)
	assertCount(t, ctx, pool, `select count(*) from gfg_game_price_daily where game_id=99003 and fact_date=$1::date and region='CN' and materialization_source='observed' and observed_at=$2::timestamptz`, 1, day, day)
	if _, err := engine.Rebuild(ctx, facts.StatePipeline, day, day, false); err != nil {
		t.Fatalf("idempotent state rebuild: %v", err)
	}
	if deleted, err := engine.PrunePlayerRaw(ctx); err != nil || deleted != 0 {
		t.Fatalf("retention disabled deleted=%d err=%v", deleted, err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO gfg_game_player_counts (game_id,appid,count,status,collected_at,run_id) VALUES (99001,99002,9,'success',$1,'facts-player-0')`, day); err == nil {
		t.Fatal("duplicate (run_id,game_id) raw guard did not reject duplicate")
	}
	if deleted, err := retentionEngine.PrunePlayerRaw(ctx); err != nil || deleted < 3 {
		t.Fatalf("checkpoint-safe player retention deleted=%d err=%v", deleted, err)
	}
	if _, err := retentionEngine.Rebuild(ctx, facts.PlayerPipeline, day, day, false); err == nil {
		t.Fatal("rebuild pretended completeness after successful raw source was pruned")
	}
}

func detailsFixture(now time.Time, snapshots []domain.RawSnapshot) domain.DetailsCollection {
	return domain.DetailsCollection{
		Details:      domain.GameDetails{GameID: 91001, AppID: 92001, Type: "game", Name: "test", CollectedAt: now},
		Localized:    []domain.GameLocalizedDetails{{GameID: 91001, AppID: 92001, Language: domain.StoreLocale("zh-CN"), Name: "test", CollectedAt: now}},
		Prices:       []domain.GamePrice{{GameID: 91001, AppID: 92001, Region: domain.Region("CN"), Currency: "CNY", CollectedAt: now}},
		Media:        domain.GameMedia{GameID: 91001, AppID: 92001, HeaderURL: "https://example.test/header.jpg", CollectedAt: now},
		Assets:       []domain.GameMediaAsset{{GameID: 91001, AppID: 92001, AssetType: "header", AssetFamily: "store", Source: "steam", MediaKey: "header", URL: "https://example.test/header.jpg", CollectedAt: now}},
		Requirements: domain.SystemRequirements{GameID: 91001, AppID: 92001, CollectedAt: now},
		Snapshots:    snapshots,
	}
}

func releaseFixture(observedAt time.Time, availability domain.ReleaseAvailability, precision domain.ReleasePrecision, raw string, start, end time.Time) *domain.GameReleaseState {
	return releaseFixtureForGame(91001, observedAt, availability, precision, raw, start, end)
}

func releaseFixtureForGame(gameID int64, observedAt time.Time, availability domain.ReleaseAvailability, precision domain.ReleasePrecision, raw string, start, end time.Time) *domain.GameReleaseState {
	year := start.Year()
	month := int(start.Month())
	return &domain.GameReleaseState{
		GameID: gameID, Availability: availability, Precision: precision,
		Year: &year, Month: &month, WindowStart: &start, WindowEnd: &end,
		RawText: raw, Source: domain.SourceSteam, SourceRegion: domain.RegionUS,
		SourceLocale: domain.StoreLocaleEN, Normalizer: "steam-go/v1.3.9", ObservedAt: observedAt,
	}
}

func minimalDetailsFixture(gameID int64, appID uint32, now time.Time) domain.DetailsCollection {
	return domain.DetailsCollection{
		Details:      domain.GameDetails{GameID: gameID, AppID: appID, Type: "game", Name: "test", CollectedAt: now},
		Media:        domain.GameMedia{GameID: gameID, AppID: appID, CollectedAt: now},
		Requirements: domain.SystemRequirements{GameID: gameID, AppID: appID, CollectedAt: now},
	}
}

func languageFixture(observedAt time.Time, code, name string) *domain.GameLanguages {
	return &domain.GameLanguages{Items: []domain.GameLanguage{{
		Code: &code, SteamName: name, Tier: "platform", SortOrder: 0,
		Source: domain.SourceSteam, SourceRegion: domain.RegionUS, SourceLocale: domain.StoreLocaleEN,
		Normalizer: "steam-go/v1.3.9", ObservedAt: observedAt,
	}}}
}

func seedGameTarget(t *testing.T, ctx context.Context, pool *pgxpool.Pool, now time.Time) {
	t.Helper()
	_, err := pool.Exec(ctx, `insert into gfg_game (
id,name,name_en,info,info_en,create_time,update_time,resources,groups,release_date,
developers,publishers,appid,header,links,weight,primary_tag,secondary_tag,view_count
) values ($1,'test','test','test','test',$2,$2,null,null,'', '[]'::jsonb,'[]'::jsonb,$3,'',null,0,0,0,0)`, int64(91001), now, int64(92001))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `insert into gfg_game_tracking_periods (game_id,appid,tracked_from,tracking_basis,opened_reason) values ($1,$2,$3,'explicit','integration')`, int64(91001), int64(92001), now.Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
}

func seedAdditionalGameTarget(t *testing.T, ctx context.Context, pool *pgxpool.Pool, gameID, appID int64, now time.Time, releaseDate string) {
	t.Helper()
	_, err := pool.Exec(ctx, `insert into gfg_game (
id,name,name_en,info,info_en,create_time,update_time,resources,groups,release_date,
developers,publishers,appid,header,links,weight,primary_tag,secondary_tag,view_count
) values ($1,'test','test','test','test',$3,$3,null,null,$4,'[]'::jsonb,'[]'::jsonb,$2,'',null,0,0,0,0)`, gameID, appID, now, releaseDate)
	if err != nil {
		t.Fatal(err)
	}
}

func assertCount(t *testing.T, ctx context.Context, pool *pgxpool.Pool, query string, want int64, args ...any) {
	t.Helper()
	var got int64
	if err := pool.QueryRow(ctx, query, args...).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("count = %d, want %d for %s", got, want, query)
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
	return "gofurry_game_collector_it_" + hex.EncodeToString(value[:])
}

func createIntegrationDatabase(t *testing.T, ctx context.Context, adminDB *sql.DB, name string) {
	t.Helper()
	validateIntegrationDatabaseName(t, name)
	if _, err := adminDB.ExecContext(ctx, `create database `+quoteIntegrationIdentifier(name)); err != nil {
		t.Fatal(err)
	}
}

func dropIntegrationDatabase(t *testing.T, adminDB *sql.DB, name string) {
	t.Helper()
	validateIntegrationDatabaseName(t, name)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := adminDB.ExecContext(ctx, `drop database if exists `+quoteIntegrationIdentifier(name)+` with (force)`); err != nil {
		t.Errorf("drop temporary database: %v", err)
	}
}

func validateIntegrationDatabaseName(t *testing.T, name string) {
	t.Helper()
	if !strings.HasPrefix(name, "gofurry_game_collector_it_") || strings.ContainsAny(name, `"'; \\`) {
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
