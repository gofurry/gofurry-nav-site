package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/report"
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

	now := time.Now().UTC().Truncate(time.Microsecond)
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
		GameID: 91001, AppID: 92001, Language: domain.Language("zh-CN"),
		Region: domain.Region("CN"), Source: domain.Source("steam"), RawPayload: []byte("{"), CollectedAt: now,
	}})
	if err := detailsRepository.SaveDetails(ctx, invalid); err == nil {
		t.Fatal("invalid snapshot JSON should roll back the details transaction")
	}
	assertCount(t, ctx, pool, `select count(*) from gfg_game_v2_details where game_id=$1`, 0, int64(91001))

	snapshots := make([]domain.RawSnapshot, 6)
	for i := range snapshots {
		snapshots[i] = domain.RawSnapshot{
			GameID: 91001, AppID: 92001, Language: domain.Language("zh-CN"),
			Region: domain.Region("CN"), Source: domain.Source("steam"),
			RawPayload: []byte(fmt.Sprintf(`{"sequence":%d}`, i)), CollectedAt: now.Add(time.Duration(i) * time.Second),
		}
	}
	if err := detailsRepository.SaveDetails(ctx, detailsFixture(now, snapshots)); err != nil {
		t.Fatalf("save details: %v", err)
	}
	assertCount(t, ctx, pool, `select count(*) from gfg_game_v2_detail_snapshots where game_id=$1`, 5, int64(91001))
	assertCount(t, ctx, pool, `select count(*) from gfg_game_v2_media where game_id=$1`, 1, int64(91001))
	assertCount(t, ctx, pool, `select count(*) from gfg_game_v2_assets where game_id=$1`, 1, int64(91001))

	news := domain.GameNews{
		GameID: 91001, AppID: 92001, Language: domain.Language("zh-CN"), EventGID: "event-1",
		AnnouncementGID: "announcement-1", Headline: "headline", CollectedAt: now,
	}
	newsRepository := NewNewsRepository(pool)
	if err := newsRepository.SaveNews(ctx, []domain.GameNews{news, news}); err != nil {
		t.Fatalf("upsert news: %v", err)
	}
	assertCount(t, ctx, pool, `select count(*) from gfg_game_v2_news where game_id=$1`, 1, int64(91001))

	playerRepository := NewPlayerRepository(pool)
	if err := playerRepository.SavePlayerCount(ctx, domain.PlayerCount{RunID: "current", GameID: 91001, AppID: 92001, Count: 42, Status: domain.StatusSuccess, CollectedAt: now}); err != nil {
		t.Fatalf("insert current player count: %v", err)
	}
	old := now.Add(-48 * time.Hour)
	if err := playerRepository.SavePlayerCount(ctx, domain.PlayerCount{RunID: "old", GameID: 91001, AppID: 92001, Count: 1, Status: domain.StatusSuccess, CollectedAt: old}); err != nil {
		t.Fatalf("insert old player count: %v", err)
	}

	runRepository := NewRunRepository(pool)
	currentRun := runFixture("current", now)
	if err := runRepository.SaveRunSummary(ctx, currentRun); err != nil {
		t.Fatalf("save current run: %v", err)
	}
	currentRun.Results = nil
	if err := runRepository.SaveRunSummary(ctx, currentRun); err != nil {
		t.Fatalf("replace current run results: %v", err)
	}
	assertCount(t, ctx, pool, `select count(*) from gfg_game_v2_collect_task_results where run_id=$1`, 0, "current")
	if err := runRepository.SaveRunSummary(ctx, runFixture("old", old)); err != nil {
		t.Fatalf("save old run: %v", err)
	}

	if err := NewRetentionRepository(pool).Prune(ctx, RetentionConfig{PlayerCountsDays: 1, CollectRunsDays: 1, CollectTaskResultsDays: 1}); err != nil {
		t.Fatalf("prune retention: %v", err)
	}
	assertCount(t, ctx, pool, `select count(*) from gfg_game_v2_player_counts where run_id=$1`, 0, "old")
	assertCount(t, ctx, pool, `select count(*) from gfg_game_v2_collect_runs where id=$1`, 0, "old")
	assertCount(t, ctx, pool, `select count(*) from gfg_game_v2_collect_runs where id=$1`, 1, "current")
}

func detailsFixture(now time.Time, snapshots []domain.RawSnapshot) domain.DetailsCollection {
	return domain.DetailsCollection{
		Details:      domain.GameDetails{GameID: 91001, AppID: 92001, Type: "game", Name: "test", CollectedAt: now},
		Localized:    []domain.GameLocalizedDetails{{GameID: 91001, AppID: 92001, Language: domain.Language("zh-CN"), Name: "test", CollectedAt: now}},
		Prices:       []domain.GamePrice{{GameID: 91001, AppID: 92001, Region: domain.Region("CN"), Currency: "CNY", CollectedAt: now}},
		Media:        domain.GameMedia{GameID: 91001, AppID: 92001, HeaderURL: "https://example.test/header.jpg", CollectedAt: now},
		Assets:       []domain.GameMediaAsset{{GameID: 91001, AppID: 92001, AssetType: "header", AssetFamily: "store", Source: "steam", MediaKey: "header", URL: "https://example.test/header.jpg", CollectedAt: now}},
		Requirements: domain.SystemRequirements{GameID: 91001, AppID: 92001, CollectedAt: now},
		Snapshots:    snapshots,
	}
}

func runFixture(id string, startedAt time.Time) report.RunSummary {
	endedAt := startedAt.Add(time.Minute)
	return report.RunSummary{
		ID: id, Status: domain.StatusSuccess, StartedAt: startedAt, EndedAt: endedAt,
		TotalCount: 1, SuccessCount: 1,
		TaskSummaries: []report.TaskSummary{{Task: domain.TaskDetails, TotalCount: 1, SuccessCount: 1, DurationMillis: 60000}},
		Results:       []report.TaskResult{{RunID: id, Task: domain.TaskDetails, Status: domain.StatusSuccess, GameID: 91001, AppID: 92001, StartedAt: startedAt, EndedAt: endedAt, DurationMillis: 60000}},
	}
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
	repositoryRoot, err := filepath.Abs(filepath.Join("..", "..", "..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("go", "tool", "goose", "-dir", filepath.Join(repositoryRoot, "db", "game", "migrations"), "postgres", dsn, "up")
	command.Dir = filepath.Join(repositoryRoot, "tools")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("apply Game baseline: %v\n%s", err, output)
	}
}
