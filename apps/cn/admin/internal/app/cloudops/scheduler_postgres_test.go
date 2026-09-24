package cloudops

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	env "github.com/gofurry/gofurry-admin/config"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Opt-in development acceptance: no DDL, no cloud call, no committed audit
// rows. Each independent pool has one connection, as two Admin instances would.
func TestScheduledPurgePostgres(t *testing.T) {
	path := os.Getenv("GOFURRY_ADMIN_SCHEDULER_TEST_CONFIG")
	if path == "" {
		t.Skip("set GOFURRY_ADMIN_SCHEDULER_TEST_CONFIG to ignored development Admin YAML")
	}
	var cfg env.ServerConfigHolder
	if err := env.InitConfig("gofurry-admin", "server.yaml", path, &cfg); err != nil {
		t.Fatal("development YAML could not be loaded")
	}
	if cfg.DataBase.Postgres.DBName != "gfa" || cfg.DataBase.Postgres.DBUser != "gofurry_app" {
		t.Fatal("requires the development GFA runtime account")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	newStore := func() *pgScheduledPurgeStore {
		config, err := pgxpool.ParseConfig(cfg.DataBase.Postgres.ConnectionString())
		if err != nil {
			t.Fatal("development GFA config invalid")
		}
		config.MinConns, config.MaxConns = 0, 1
		config.ConnConfig.ConnectTimeout = 5 * time.Second
		config.ConnConfig.RuntimeParams["application_name"] = "admin-scheduled-purge-acceptance"
		pool, err := pgxpool.NewWithConfig(ctx, config)
		if err != nil {
			t.Fatal("development GFA pool failed")
		}
		t.Cleanup(pool.Close)
		if err := pool.Ping(ctx); err != nil {
			t.Fatal("development GFA connection failed")
		}
		return &pgScheduledPurgeStore{pool: pool, logger: audit.New(pool)}
	}
	storeA, storeB := newStore(), newStore()
	acquire := func(store *pgScheduledPurgeStore) (scheduledPurgeLease, func()) {
		t.Helper()
		lease, err := store.Acquire(ctx)
		if err != nil || lease == nil {
			t.Fatalf("acquire: lease=%v err=%v", lease != nil, err)
		}
		var once sync.Once
		release := func() {
			once.Do(func() {
				cleanupCtx, done := context.WithTimeout(context.Background(), 3*time.Second)
				defer done()
				if err := lease.Release(cleanupCtx); err != nil {
					t.Error(err)
				}
			})
		}
		t.Cleanup(release)
		return lease, release
	}
	_, releaseA := acquire(storeA)
	if lease, err := storeB.Acquire(ctx); err != nil || lease != nil {
		if lease != nil {
			_ = lease.Release(ctx)
		}
		t.Fatalf("B must not acquire A's lock: %v", err)
	}
	releaseA()
	_, releaseB := acquire(storeB)
	releaseB()
	t.Log("separate sessions: A acquired, B skipped, A unlocked, B acquired")

	// Exercise sqlc + existing audit logger on the lock-owning connection. The
	// explicit transaction is test-only; production intent is auto-committed.
	lease, release := acquire(storeA)
	pgLease := lease.(*pgScheduledPurgeLease)
	tx, err := pgLease.conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	slot := "edgeone-scheduled-purge:test-rollback-" + time.Now().Format("20060102T150405.000000000")
	host := "scheduled-purge-acceptance.example.test"
	if found, err := lease.Recorded(ctx, slot, host); err != nil || found {
		t.Fatalf("unexpected initial intent: %v / %v", found, err)
	}
	meta := audit.SystemMeta(slot)
	for _, detail := range []map[string]any{
		{"status": "requested", "input": map[string]any{"type": "host", "targets": []string{host}, "scheduled_for": testSlot().Format(time.RFC3339)}},
		{"status": "completed", "result": map[string]any{"job_id": "fake-job", "status": "submitted"}},
	} {
		if err := lease.Audit(ctx, meta, host, detail); err != nil {
			t.Fatal(err)
		}
	}
	if found, err := lease.Recorded(ctx, slot, host); err != nil || !found {
		t.Fatalf("intent missing: %v / %v", found, err)
	}
	rows, err := tx.Query(ctx, `SELECT action, resource, target_id, operator, request_id, after_data FROM gfa_admin_audit_log WHERE request_id=$1 ORDER BY id`, slot)
	if err != nil {
		t.Fatal(err)
	}
	var statuses []string
	for rows.Next() {
		var action, resource, target, operator, requestID, after string
		if err := rows.Scan(&action, &resource, &target, &operator, &requestID, &after); err != nil {
			t.Fatal(err)
		}
		if action != scheduledPurgeAction || resource != "cloud" || target != host || operator != "system" || requestID != slot {
			t.Fatal("incorrect stored audit metadata")
		}
		var detail struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal([]byte(after), &detail); err != nil {
			t.Fatal(err)
		}
		statuses = append(statuses, detail.Status)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(statuses) != 2 || statuses[0] != "requested" || statuses[1] != "completed" {
		t.Fatalf("stored audit states: %v", statuses)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	if found, err := lease.Recorded(ctx, slot, host); err != nil || found {
		t.Fatalf("test audit rollback failed: %v / %v", found, err)
	}
	release()
	t.Log("single-connection audit and slot lookup verified; test audit rows rolled back")

	// If unlock cannot be confirmed (e.g. canceled connection), discard that
	// session rather than returning a potentially held lock to the pool.
	uncertain, err := storeA.Acquire(ctx)
	if err != nil || uncertain == nil {
		t.Fatal("could not acquire for canceled unlock check")
	}
	canceledCtx, canceled := context.WithCancel(ctx)
	canceled()
	if err := uncertain.Release(canceledCtx); err == nil {
		t.Fatal("canceled unlock should fail and close the session")
	}
	_, releaseAfterClose := acquire(storeB)
	releaseAfterClose()
	t.Log("failed unlock discarded its session; another instance can acquire")
}
