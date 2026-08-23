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
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatal(err)
	}

	repositoryRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	update := os.Getenv("GOFURRY_UPDATE_SCHEMA_SNAPSHOTS") == "1"
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
			if err := goose.UpContext(ctx, freshDB, migrationDir); err != nil {
				t.Fatalf("goose up on empty %s database: %v", test.label, err)
			}
			actual, err := schema.Inspect(ctx, freshDB)
			if err != nil {
				t.Fatalf("inspect fresh %s schema: %v", test.label, err)
			}
			if update {
				writeExpectedSnapshot(t, repositoryRoot, test.label, actual)
				return
			}
			expected, err := loadExpected(test.label)
			if err != nil {
				t.Fatal(err)
			}
			if difference := schema.Difference(expected, actual); difference != "" {
				t.Fatalf("fresh %s schema drift: %s", test.label, difference)
			}

			if _, err := freshDB.ExecContext(ctx, `drop table public.goose_db_version`); err != nil {
				t.Fatalf("prepare exact pre-Goose fixture: %v", err)
			}
			if err := adopt(ctx, freshDB, adoptOptions{DatabaseLabel: test.label, BaselineVersion: expectedBaselineVersion}); err != nil {
				t.Fatalf("adopt exact %s schema: %v", test.label, err)
			}
			version, err := goose.GetDBVersionContext(ctx, freshDB)
			if err != nil {
				t.Fatalf("read adopted %s version: %v", test.label, err)
			}
			if version != expectedBaselineVersion {
				t.Fatalf("adopted version = %d, want %d", version, expectedBaselineVersion)
			}
			if err := goose.UpContext(ctx, freshDB, migrationDir); err != nil {
				t.Fatalf("normal goose up after %s adoption: %v", test.label, err)
			}

			driftName := temporaryDatabaseName(test.label, "drift")
			createDatabase(t, ctx, adminDB, driftName)
			defer dropDatabase(t, adminDB, driftName)
			driftDB := openDatabase(t, adminDSN, driftName)
			defer driftDB.Close()
			if err := goose.UpContext(ctx, driftDB, migrationDir); err != nil {
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
	t.Helper()
	data, err := schema.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, '\n')
	path := filepath.Join(repositoryRoot, "tools", "db-baseline", "expected", label+".json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
