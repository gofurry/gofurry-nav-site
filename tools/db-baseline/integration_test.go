package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-site/tools/internal/schema"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"gopkg.in/yaml.v3"
)

type integrationPostgresConfig struct {
	DSN      string `yaml:"dsn"`
	Name     string `yaml:"db_name"`
	Username string `yaml:"db_username"`
	Password string `yaml:"db_password"`
	Host     string `yaml:"db_host"`
	Port     string `yaml:"db_port"`
}

type integrationConfig struct {
	Database struct {
		Postgres integrationPostgresConfig `yaml:"postgres"`
	} `yaml:"database"`
}

func TestPostgresFreshAndBaselineAdoption(t *testing.T) {
	adminDSN := integrationAdminDSN(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	adminDB := openDatabase(t, adminDSN, "postgres")
	defer adminDB.Close()
	if err := adminDB.PingContext(ctx); err != nil {
		t.Fatalf("connect to development PostgreSQL: %v", err)
	}
	var serverVersion string
	var serverVersionNumber int
	if err := adminDB.QueryRowContext(ctx, `SHOW server_version`).Scan(&serverVersion); err != nil {
		t.Fatal(err)
	}
	if err := adminDB.QueryRowContext(ctx, `SHOW server_version_num`).Scan(&serverVersionNumber); err != nil {
		t.Fatal(err)
	}
	t.Logf("PostgreSQL server version: %s", serverVersion)
	if required := strings.TrimSpace(os.Getenv("GOFURRY_REQUIRE_POSTGRES_MAJOR")); required != "" {
		actualMajor := serverVersionNumber / 10000
		if fmt.Sprint(actualMajor) != required {
			t.Fatalf("PostgreSQL major version=%d, want %s", actualMajor, required)
		}
	}
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatal(err)
	}

	repositoryRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	updateFinal := os.Getenv("GOFURRY_UPDATE_SCHEMA_SNAPSHOTS") == "1"
	updateBaseline := os.Getenv("GOFURRY_UPDATE_BASELINE_SNAPSHOTS") == "1"
	tests := []struct {
		label      string
		owner      string
		driftTable string
	}{
		{label: "gfg", owner: "game", driftTable: "gfg_game"},
		{label: "gfn", owner: "nav", driftTable: "gfn_site"},
		{label: "gfa", owner: "admin", driftTable: "gfa_admin_account"},
	}
	for _, test := range tests {
		t.Run(test.label, func(t *testing.T) {
			freshName := temporaryDatabaseName(test.label, "fresh")
			createDatabase(t, ctx, adminDB, freshName)
			defer dropDatabase(t, adminDB, freshName)
			freshDB := openDatabase(t, adminDSN, freshName)
			defer freshDB.Close()

			migrationDir := filepath.Join(repositoryRoot, "db", test.owner, "migrations")
			if updateBaseline {
				if err := goose.UpToContext(ctx, freshDB, migrationDir, expectedBaselineVersion); err != nil {
					t.Fatalf("goose baseline on empty %s database: %v", test.label, err)
				}
				actual, err := schema.Inspect(ctx, freshDB)
				if err != nil {
					t.Fatalf("inspect baseline %s schema: %v", test.label, err)
				}
				writeExpectedSnapshot(t, repositoryRoot, test.label, actual)
				return
			}
			if err := goose.UpContext(ctx, freshDB, migrationDir); err != nil {
				t.Fatalf("goose up on empty %s database: %v", test.label, err)
			}
			actual, err := schema.Inspect(ctx, freshDB)
			if err != nil {
				t.Fatalf("inspect fresh %s schema: %v", test.label, err)
			}
			if updateFinal {
				writeFinalExpectedSnapshot(t, repositoryRoot, test.label, actual)
				return
			}
			expected, err := loadFinalExpected(repositoryRoot, test.label)
			if err != nil {
				t.Fatal(err)
			}
			if difference := schema.Difference(expected, actual); difference != "" {
				t.Fatalf("fresh %s schema drift: %s", test.label, difference)
			}

			adoptName := temporaryDatabaseName(test.label, "adopt")
			createDatabase(t, ctx, adminDB, adoptName)
			defer dropDatabase(t, adminDB, adoptName)
			adoptDB := openDatabase(t, adminDSN, adoptName)
			defer adoptDB.Close()
			if err := goose.UpToContext(ctx, adoptDB, migrationDir, expectedBaselineVersion); err != nil {
				t.Fatalf("prepare exact pre-Goose %s fixture: %v", test.label, err)
			}
			if test.label == "gfg" || test.label == "gfn" {
				seedAndBackupDeprecatedTables(t, ctx, adoptDB, test.label)
			}
			if _, err := adoptDB.ExecContext(ctx, `drop table public.goose_db_version`); err != nil {
				t.Fatalf("prepare exact pre-Goose fixture: %v", err)
			}
			if err := adopt(ctx, adoptDB, adoptOptions{DatabaseLabel: test.label, BaselineVersion: expectedBaselineVersion}); err != nil {
				t.Fatalf("adopt exact %s schema: %v", test.label, err)
			}
			version, err := goose.GetDBVersionContext(ctx, adoptDB)
			if err != nil {
				t.Fatalf("read adopted %s version: %v", test.label, err)
			}
			if version != expectedBaselineVersion {
				t.Fatalf("adopted version = %d, want %d", version, expectedBaselineVersion)
			}
			if err := goose.UpContext(ctx, adoptDB, migrationDir); err != nil {
				t.Fatalf("normal goose up after %s adoption: %v", test.label, err)
			}
			if test.label == "gfg" || test.label == "gfn" {
				assertDeprecatedCleanupAndBackup(t, ctx, adoptDB, test.label)
			}

			driftName := temporaryDatabaseName(test.label, "drift")
			createDatabase(t, ctx, adminDB, driftName)
			defer dropDatabase(t, adminDB, driftName)
			driftDB := openDatabase(t, adminDSN, driftName)
			defer driftDB.Close()
			if err := goose.UpToContext(ctx, driftDB, migrationDir, expectedBaselineVersion); err != nil {
				t.Fatalf("prepare drift %s fixture: %v", test.label, err)
			}
			if _, err := driftDB.ExecContext(ctx, `drop table public.goose_db_version`); err != nil {
				t.Fatalf("remove Goose table from drift fixture: %v", err)
			}
			alter := fmt.Sprintf(`alter table public.%s add column foundation_drift_probe text`, quoteIdentifier(test.driftTable))
			if _, err := driftDB.ExecContext(ctx, alter); err != nil {
				t.Fatalf("create intentional %s drift: %v", test.label, err)
			}
			if err := adopt(ctx, driftDB, adoptOptions{DatabaseLabel: test.label, BaselineVersion: expectedBaselineVersion}); err == nil {
				t.Fatalf("intentional %s drift was accepted", test.label)
			}
			var versionTableExists bool
			if err := driftDB.QueryRowContext(ctx, `select to_regclass('public.goose_db_version') is not null`).Scan(&versionTableExists); err != nil {
				t.Fatal(err)
			}
			if versionTableExists {
				t.Fatalf("failed %s adoption left a Goose version table", test.label)
			}
		})
	}
}

func integrationAdminDSN(t *testing.T) string {
	t.Helper()
	if dsn := os.Getenv("GOFURRY_TEST_POSTGRES_ADMIN_URL"); dsn != "" {
		return dsn
	}
	configPath := os.Getenv("GOFURRY_DEV_ADMIN_CONFIG")
	if configPath == "" {
		t.Skip("set GOFURRY_TEST_POSTGRES_ADMIN_URL or GOFURRY_DEV_ADMIN_CONFIG for PostgreSQL integration tests")
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	var cfg integrationConfig
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		t.Fatal(err)
	}
	postgres := cfg.Database.Postgres
	if postgres.DSN != "" {
		return postgres.DSN
	}
	u := &url.URL{Scheme: "postgres", Host: postgres.Host + ":" + postgres.Port, Path: postgres.Name}
	u.User = url.UserPassword(postgres.Username, postgres.Password)
	q := u.Query()
	q.Set("sslmode", "prefer")
	u.RawQuery = q.Encode()
	return u.String()
}

func openDatabase(t *testing.T, dsn, databaseName string) *sql.DB {
	t.Helper()
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.Database = databaseName
	config.RuntimeParams["application_name"] = "gofurry-foundation-integration"
	db := stdlib.OpenDB(*config)
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)
	return db
}

func temporaryDatabaseName(label, purpose string) string {
	var random [5]byte
	if _, err := rand.Read(random[:]); err != nil {
		panic(err)
	}
	return fmt.Sprintf("gofurry_foundation_it_%s_%s_%s", label, purpose, hex.EncodeToString(random[:]))
}

func createDatabase(t *testing.T, ctx context.Context, adminDB *sql.DB, name string) {
	t.Helper()
	validateTemporaryDatabaseName(t, name)
	if _, err := adminDB.ExecContext(ctx, `create database `+quoteIdentifier(name)); err != nil {
		t.Fatalf("create temporary database %s: %v", name, err)
	}
}

func dropDatabase(t *testing.T, adminDB *sql.DB, name string) {
	t.Helper()
	validateTemporaryDatabaseName(t, name)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := adminDB.ExecContext(ctx, `drop database if exists `+quoteIdentifier(name)+` with (force)`); err != nil {
		t.Errorf("drop temporary database %s: %v", name, err)
	}
}

func validateTemporaryDatabaseName(t *testing.T, name string) {
	t.Helper()
	if !strings.HasPrefix(name, "gofurry_foundation_it_") || strings.ContainsAny(name, `"'; \\`) {
		t.Fatalf("refusing unsafe temporary database name %q", name)
	}
}

func quoteIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func writeExpectedSnapshot(t *testing.T, repositoryRoot, label string, snapshot schema.Snapshot) {
	writeSnapshot(t, filepath.Join(repositoryRoot, "tools", "db-baseline", "expected", label+".json"), snapshot)
}

func writeFinalExpectedSnapshot(t *testing.T, repositoryRoot, label string, snapshot schema.Snapshot) {
	path := filepath.Join(repositoryRoot, "tools", "db-baseline", "expected-final", label+".json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	writeSnapshot(t, path, snapshot)
}

func writeSnapshot(t *testing.T, path string, snapshot schema.Snapshot) {
	t.Helper()
	data, err := schema.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func loadFinalExpected(repositoryRoot, label string) (schema.Snapshot, error) {
	data, err := os.ReadFile(filepath.Join(repositoryRoot, "tools", "db-baseline", "expected-final", label+".json"))
	if err != nil {
		return schema.Snapshot{}, err
	}
	return schema.Unmarshal(data)
}

func seedAndBackupDeprecatedTables(t *testing.T, ctx context.Context, db *sql.DB, label string) {
	t.Helper()
	statements := map[string]string{
		"gfg": `
INSERT INTO public.gfg_game_creator_deprecated_20260614
  (id,name,info,main_url,links,cover,contact,create_time,update_time,type,name_en,info_en,deleted)
VALUES (1,'creator','info','https://example.test','[]','cover','[]',NOW(),NOW(),1,'creator','info',false);
INSERT INTO public.gfg_game_record
  (id,game_id,language,release_date,platform,developer,publisher,info,cover,lang,price_list,initial,final,discount)
VALUES (1,1,'zh','2026.08.23','windows','dev','pub','info','cover','zh','[]',100,50,50);
INSERT INTO public.gfg_game_news
  (id,game_id,headline,content,"index",post_time,create_time,author,url,total,lang)
VALUES (1,1,'headline','content',1,NOW(),NOW(),'author','https://example.test',1,'zh');
INSERT INTO public.gfg_game_player_count (id,game_id,count,create_time) VALUES (1,1,10,NOW());
INSERT INTO public.gfg_game_v2_collect_runs
  (id,task_type,status,total_count,success_count,failed_count,skipped_count,partial_count,task_summary,duration_millis,started_at,ended_at)
VALUES ('legacy-run-integration','details','partial',2,1,0,0,1,'[]',123,NOW() - INTERVAL '1 minute',NOW());
INSERT INTO public.gfg_game_v2_collect_task_results
  (run_id,task_type,status,game_id,appid,duration_millis,started_at,ended_at)
VALUES
  ('legacy-run-integration','details','success',900001,900101,60,NOW() - INTERVAL '1 minute',NOW()),
  ('legacy-run-integration','details','partial',900002,900102,63,NOW() - INTERVAL '1 minute',NOW());
CREATE SCHEMA foundation_cleanup_backup;
CREATE TABLE foundation_cleanup_backup.gfg_game_creator_deprecated_20260614 AS TABLE public.gfg_game_creator_deprecated_20260614 WITH DATA;
CREATE TABLE foundation_cleanup_backup.gfg_game_record AS TABLE public.gfg_game_record WITH DATA;
CREATE TABLE foundation_cleanup_backup.gfg_game_news AS TABLE public.gfg_game_news WITH DATA;
CREATE TABLE foundation_cleanup_backup.gfg_game_player_count AS TABLE public.gfg_game_player_count WITH DATA;`,
		"gfn": `
INSERT INTO public.gfn_log_update (id,title,url,create_time,update_time,deleted)
VALUES (1,'legacy','https://example.test',NOW(),NOW(),false);
CREATE SCHEMA foundation_cleanup_backup;
CREATE TABLE foundation_cleanup_backup.gfn_log_update AS TABLE public.gfn_log_update WITH DATA;`,
	}
	if _, err := db.ExecContext(ctx, statements[label]); err != nil {
		t.Fatalf("seed and simulate %s deprecated-table backup: %v", label, err)
	}
	counts := deprecatedTableCounts(t, ctx, db, label, "public")
	t.Logf("%s deprecated row counts before cleanup: %v", label, counts)
}

func assertDeprecatedCleanupAndBackup(t *testing.T, ctx context.Context, db *sql.DB, label string) {
	t.Helper()
	tables := deprecatedTables(label)
	for _, table := range tables {
		// V3-P0.1.1 intentionally reuses gfg_game_news for the canonical
		// Steam news table after the deprecated table is backed up and dropped.
		if label == "gfg" && table == "gfg_game_news" {
			var oldShape bool
			if err := db.QueryRowContext(ctx, `
SELECT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'gfg_game_news' AND column_name = 'content'
)`).Scan(&oldShape); err != nil {
				t.Fatal(err)
			}
			if oldShape {
				t.Fatal("deprecated public.gfg_game_news shape still exists")
			}
			continue
		}
		var exists bool
		if err := db.QueryRowContext(ctx, `SELECT to_regclass($1) IS NOT NULL`, "public."+table).Scan(&exists); err != nil {
			t.Fatal(err)
		}
		if exists {
			t.Fatalf("deprecated table public.%s still exists", table)
		}
	}
	counts := deprecatedTableCounts(t, ctx, db, label, "foundation_cleanup_backup")
	for table, count := range counts {
		if count != 1 {
			t.Fatalf("backup %s.%s row count=%d, want 1", "foundation_cleanup_backup", table, count)
		}
	}
	if label == "gfg" {
		var runs, results int64
		if err := db.QueryRowContext(ctx, `
SELECT count(*)
FROM public.gfg_collection_runs r
JOIN public.gfg_collection_jobs j ON j.id = r.job_id
WHERE r.id = 'legacy-run-integration'
  AND j.trigger = 'legacy_import'
  AND j.requested_by = 'migration'
  AND r.expected_count = 2
  AND r.attempted_count = 2
  AND r.success_count = 1
  AND r.partial_count = 1`).Scan(&runs); err != nil {
			t.Fatal(err)
		}
		if err := db.QueryRowContext(ctx, `
SELECT count(*) FROM public.gfg_collection_task_results
WHERE run_id = 'legacy-run-integration'`).Scan(&results); err != nil {
			t.Fatal(err)
		}
		if runs != 1 || results != 2 {
			t.Fatalf("legacy Game collection history migration: runs=%d results=%d", runs, results)
		}
	}
}

func deprecatedTableCounts(t *testing.T, ctx context.Context, db *sql.DB, label, schemaName string) map[string]int64 {
	t.Helper()
	counts := make(map[string]int64)
	for _, table := range deprecatedTables(label) {
		query := `SELECT COUNT(*) FROM ` + quoteIdentifier(schemaName) + `.` + quoteIdentifier(table)
		var count int64
		if err := db.QueryRowContext(ctx, query).Scan(&count); err != nil {
			t.Fatal(err)
		}
		counts[table] = count
	}
	return counts
}

func deprecatedTables(label string) []string {
	if label == "gfg" {
		return []string{"gfg_game_creator_deprecated_20260614", "gfg_game_record", "gfg_game_news", "gfg_game_player_count"}
	}
	return []string{"gfn_log_update"}
}
