package control

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	gamesqlc "github.com/gofurry/gofurry-game-collector/internal/db/game/sqlc"
	"github.com/gofurry/gofurry-game-collector/roof/env"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestGameExpiredLeaseRecoveryIntegration(t *testing.T) {
	configPath := os.Getenv("GOFURRY_GAME_COLLECTOR_INTEGRATION_CONFIG")
	if configPath == "" {
		t.Skip("set GOFURRY_GAME_COLLECTOR_INTEGRATION_CONFIG for PostgreSQL integration tests")
	}
	if err := env.LoadServerConfig(configPath); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pool := newGameRecoveryDatabase(t, ctx, env.GetServerConfig().DataBase.ConnectionString())

	if _, err := pool.Exec(ctx, `
INSERT INTO gfg_collector_instances
    (instance_id,collector_id,hostname,version,commit_sha,capabilities,started_at,last_heartbeat_at)
VALUES ('lost-game-instance','game-recovery','host','test','test',ARRAY['game.metadata'],now()-interval '2 minutes',now()-interval '2 minutes')`); err != nil {
		t.Fatal(err)
	}
	var lostJobID int64
	if err := pool.QueryRow(ctx, `
INSERT INTO gfg_collection_jobs
    (job_key,trigger,scope_type,tasks,priority,concurrency_key,status,requested_by,claimed_by,lease_until)
VALUES ('game.metadata','manual','all',ARRAY['details'],100,'steam','running','recovery-test','lost-game-instance',now()-interval '1 minute')
RETURNING id`).Scan(&lostJobID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO gfg_collection_runs
    (id,job_id,attempt_no,collector_instance_id,status,started_at,expected_count)
VALUES ('lost-game-run',$1,1,'lost-game-instance','running',now()-interval '2 minutes',1)`, lostJobID); err != nil {
		t.Fatal(err)
	}
	var queuedJobID int64
	if err := pool.QueryRow(ctx, `
INSERT INTO gfg_collection_jobs
    (job_key,trigger,scope_type,tasks,priority,concurrency_key,status,requested_by)
VALUES ('game.metadata','manual','all',ARRAY['details'],90,'steam','queued','recovery-test')
RETURNING id`).Scan(&queuedJobID); err != nil {
		t.Fatal(err)
	}

	worker := NewWorker(pool, nil, "replacement-game-instance")
	if err := worker.RecoverExpired(ctx); err != nil {
		t.Fatalf("recover expired Game lease: %v", err)
	}
	var jobStatus, runStatus, errorKind string
	var leaseCleared, runEnded bool
	if err := pool.QueryRow(ctx, `
SELECT job.status, job.lease_until IS NULL, run.status, run.error_kind, run.ended_at IS NOT NULL
FROM gfg_collection_jobs job
JOIN gfg_collection_runs run ON run.job_id=job.id
WHERE job.id=$1`, lostJobID).Scan(&jobStatus, &leaseCleared, &runStatus, &errorKind, &runEnded); err != nil {
		t.Fatal(err)
	}
	if jobStatus != "failed" || !leaseCleared || runStatus != "failed" || errorKind != "worker_lost" || !runEnded {
		t.Fatalf("recovered Game job/run=%s/%t %s/%s/%t", jobStatus, leaseCleared, runStatus, errorKind, runEnded)
	}

	queries := gamesqlc.New(pool)
	replacement := "replacement-game-instance"
	claimed, err := queries.ClaimNextGameCollectionJob(ctx, gamesqlc.ClaimNextGameCollectionJobParams{
		InstanceID: &replacement, LeaseSeconds: leaseSeconds,
	})
	if err != nil {
		t.Fatalf("claim queued Game work after recovery: %v", err)
	}
	if claimed.ID != queuedJobID {
		t.Fatalf("claimed Game job=%d, want %d", claimed.ID, queuedJobID)
	}
	if _, err := queries.ClaimNextGameCollectionJob(ctx, gamesqlc.ClaimNextGameCollectionJobParams{
		InstanceID: &replacement, LeaseSeconds: leaseSeconds,
	}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("second same-lane Game claim error=%v, want no rows", err)
	}
	assertGameRecoveryCount(t, ctx, pool, `SELECT count(*) FROM gfg_collection_jobs WHERE concurrency_key='steam' AND status='running'`, 1)
	assertGameRecoveryCount(t, ctx, pool, `SELECT count(*) FROM gfg_collection_runs run LEFT JOIN gfg_collection_jobs job ON job.id=run.job_id WHERE job.id IS NULL`, 0)
}

func newGameRecoveryDatabase(t *testing.T, ctx context.Context, dsn string) *pgxpool.Pool {
	t.Helper()
	baseConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	adminConfig := baseConfig.Copy()
	adminConfig.ConnConfig.Database = "postgres"
	adminPool, err := pgxpool.NewWithConfig(ctx, adminConfig)
	if err != nil {
		t.Fatal(err)
	}
	if err := adminPool.Ping(ctx); err != nil {
		adminPool.Close()
		t.Fatal(err)
	}
	databaseName := recoveryDatabaseName("game")
	if _, err := adminPool.Exec(ctx, `CREATE DATABASE `+pgx.Identifier{databaseName}.Sanitize()); err != nil {
		adminPool.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if _, err := adminPool.Exec(cleanupCtx, `DROP DATABASE IF EXISTS `+pgx.Identifier{databaseName}.Sanitize()+` WITH (FORCE)`); err != nil {
			t.Errorf("drop Game recovery database: %v", err)
		}
		adminPool.Close()
	})

	testConfig := baseConfig.Copy()
	testConfig.ConnConfig.Database = databaseName
	applyRecoveryMigrations(t, "game", recoveryDatabaseDSN(dsn, databaseName))
	pool, err := pgxpool.NewWithConfig(ctx, testConfig)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func recoveryDatabaseDSN(baseDSN, databaseName string) string {
	u, err := url.Parse(baseDSN)
	if err != nil {
		panic(err)
	}
	u.Path = databaseName
	return u.String()
}

func recoveryDatabaseName(domain string) string {
	var value [5]byte
	if _, err := rand.Read(value[:]); err != nil {
		panic(err)
	}
	return "gofurry_" + domain + "_recovery_it_" + hex.EncodeToString(value[:])
}

func applyRecoveryMigrations(t *testing.T, domain, dsn string) {
	t.Helper()
	root := recoveryRepositoryRoot(t)
	command := exec.Command("go", "tool", "goose", "-dir", filepath.Join(root, "db", domain, "migrations"), "postgres", dsn, "up")
	command.Dir = filepath.Join(root, "tools")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("apply %s recovery migrations: %v\n%s", domain, err, output)
	}
}

func recoveryRepositoryRoot(t *testing.T) string {
	t.Helper()
	current, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(current, "sqlc.yaml")); err == nil {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			t.Fatal("repository root containing sqlc.yaml not found")
		}
		current = parent
	}
}

func assertGameRecoveryCount(t *testing.T, ctx context.Context, pool *pgxpool.Pool, query string, want int64) {
	t.Helper()
	var got int64
	if err := pool.QueryRow(ctx, query).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("count=%d, want %d for %s", got, want, query)
	}
}
