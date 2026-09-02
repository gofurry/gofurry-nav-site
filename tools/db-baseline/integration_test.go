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

			// Exercise the deployed alpha.2 starting boundary through current. Game
			// additionally proves both the pre-P0.2 Raw floor and the follow-up
			// P0.2 Tracking/Fact floor before the final sequence correction.
			if alpha2Version, ok := map[string]int64{
				"gfg": 20260827010400,
				"gfn": 20260827010100,
			}[test.label]; ok {
				upgradeName := temporaryDatabaseName(test.label, "alpha2_upgrade")
				createDatabase(t, ctx, adminDB, upgradeName)
				defer dropDatabase(t, adminDB, upgradeName)
				upgradeDB := openDatabase(t, adminDSN, upgradeName)
				defer upgradeDB.Close()
				if err := goose.UpToContext(ctx, upgradeDB, migrationDir, alpha2Version); err != nil {
					t.Fatalf("prepare %s alpha.2 fixture: %v", test.label, err)
				}
				if test.label == "gfg" {
					if _, err := upgradeDB.ExecContext(ctx, `
INSERT INTO public.gfg_game_player_counts
    (game_id, appid, count, status, collected_at, run_id)
VALUES (950001, 950101, 1, 'success', transaction_timestamp(), 'alpha2-deleted-high-water');`); err != nil {
						t.Fatalf("seed alpha.2 deleted Game evidence: %v", err)
					}
					if err := goose.UpToContext(ctx, upgradeDB, migrationDir, 20260828010200); err != nil {
						t.Fatalf("upgrade gfg through initial P0.2 migrations: %v", err)
					}
					if _, err := upgradeDB.ExecContext(ctx, `
INSERT INTO public.gfg_game_tracking_periods
    (game_id, appid, tracked_from, tracked_until, tracking_basis, opened_reason, closed_reason)
VALUES
    (960001, 960101, transaction_timestamp() - interval '2 days',
     transaction_timestamp() - interval '1 day', 'explicit', 'sequence_test', 'sequence_test');`); err != nil {
						t.Fatalf("seed P0.2 historical Game evidence: %v", err)
					}
				}
				if err := goose.UpContext(ctx, upgradeDB, migrationDir); err != nil {
					t.Fatalf("upgrade %s alpha.2 to current: %v", test.label, err)
				}
				upgradeActual, err := schema.Inspect(ctx, upgradeDB)
				if err != nil {
					t.Fatalf("inspect upgraded %s schema: %v", test.label, err)
				}
				if difference := schema.Difference(expected, upgradeActual); difference != "" {
					t.Fatalf("alpha.2 -> current %s schema drift: %s", test.label, difference)
				}
				if test.label == "gfg" {
					var allocated int64
					if err := upgradeDB.QueryRowContext(ctx, `SELECT nextval('public.gfg_game_id_seq')`).Scan(&allocated); err != nil {
						t.Fatal(err)
					}
					if allocated <= 960001 {
						t.Fatalf("historical Game ID was reusable after alpha.2 upgrade: next=%d", allocated)
					}
				}
			}

			// Exercise the deployed alpha.3 starting boundary through current,
			// independently from the longer alpha.2 upgrade path.
			if alpha3Version, ok := map[string]int64{
				"gfg": 20260828010300,
				"gfn": 20260829010000,
			}[test.label]; ok {
				upgradeName := temporaryDatabaseName(test.label, "alpha3_upgrade")
				createDatabase(t, ctx, adminDB, upgradeName)
				defer dropDatabase(t, adminDB, upgradeName)
				upgradeDB := openDatabase(t, adminDSN, upgradeName)
				defer upgradeDB.Close()
				if err := goose.UpToContext(ctx, upgradeDB, migrationDir, alpha3Version); err != nil {
					t.Fatalf("prepare %s alpha.3 fixture: %v", test.label, err)
				}
				if err := goose.UpContext(ctx, upgradeDB, migrationDir); err != nil {
					t.Fatalf("upgrade %s alpha.3 to current: %v", test.label, err)
				}
				upgradeActual, err := schema.Inspect(ctx, upgradeDB)
				if err != nil {
					t.Fatalf("inspect alpha.3-upgraded %s schema: %v", test.label, err)
				}
				if difference := schema.Difference(expected, upgradeActual); difference != "" {
					t.Fatalf("alpha.3 -> current %s schema drift: %s", test.label, difference)
				}
			}

			// Exercise the deployed alpha.4 boundary through current independently,
			// including Goose-owned detector seeds and later compatibility repairs.
			if test.label == "gfg" || test.label == "gfn" {
				upgradeName := temporaryDatabaseName(test.label, "alpha4_upgrade")
				createDatabase(t, ctx, adminDB, upgradeName)
				defer dropDatabase(t, adminDB, upgradeName)
				upgradeDB := openDatabase(t, adminDSN, upgradeName)
				defer upgradeDB.Close()
				if err := goose.UpToContext(ctx, upgradeDB, migrationDir, 20260829020000); err != nil {
					t.Fatalf("prepare %s alpha.4 fixture: %v", test.label, err)
				}
				if err := goose.UpContext(ctx, upgradeDB, migrationDir); err != nil {
					t.Fatalf("upgrade %s alpha.4 to current: %v", test.label, err)
				}
				upgradeActual, err := schema.Inspect(ctx, upgradeDB)
				if err != nil {
					t.Fatalf("inspect alpha.5-upgraded %s schema: %v", test.label, err)
				}
				if difference := schema.Difference(expected, upgradeActual); difference != "" {
					t.Fatalf("alpha.4 -> alpha.5 %s schema drift: %s", test.label, difference)
				}
				var registryCount, checkpointCount int
				prefix := map[string]string{"gfg": "gfg", "gfn": "gfn"}[test.label]
				if err := upgradeDB.QueryRowContext(ctx, fmt.Sprintf(`SELECT count(*) FROM public.%s_change_registry`, prefix)).Scan(&registryCount); err != nil {
					t.Fatal(err)
				}
				if err := upgradeDB.QueryRowContext(ctx, fmt.Sprintf(`SELECT count(*) FROM public.%s_change_checkpoints`, prefix)).Scan(&checkpointCount); err != nil {
					t.Fatal(err)
				}
				expectedDetectors := 5
				if test.label == "gfg" {
					expectedDetectors = 6
				} else if test.label == "gfn" {
					expectedDetectors = 11
				}
				if registryCount != expectedDetectors || checkpointCount != expectedDetectors {
					t.Fatalf("%s current detector seeds registry=%d checkpoints=%d, want %d", test.label, registryCount, checkpointCount, expectedDetectors)
				}
			}

			// Exercise the staged P2.2 Game Mac rollout boundary: the Metric
			// contract is independently deployable/backfillable before the Change
			// contract, and the completed upgrade converges on the current schema.
			if test.label == "gfg" {
				upgradeName := temporaryDatabaseName(test.label, "p22_mac_upgrade")
				createDatabase(t, ctx, adminDB, upgradeName)
				defer dropDatabase(t, adminDB, upgradeName)
				upgradeDB := openDatabase(t, adminDSN, upgradeName)
				defer upgradeDB.Close()
				if err := goose.UpToContext(ctx, upgradeDB, migrationDir, 20260902010000); err != nil {
					t.Fatalf("prepare gfg P2.2 Mac Metric boundary: %v", err)
				}
				var metricContracts, changeContracts int
				if err := upgradeDB.QueryRowContext(ctx, `SELECT count(*) FROM gfg_metric_registry WHERE metric_key='mac_support' AND metric_version=1 AND status='active'`).Scan(&metricContracts); err != nil {
					t.Fatal(err)
				}
				if err := upgradeDB.QueryRowContext(ctx, `SELECT count(*) FROM gfg_change_registry WHERE detector_key='mac_support_transition'`).Scan(&changeContracts); err != nil {
					t.Fatal(err)
				}
				if metricContracts != 1 || changeContracts != 0 {
					t.Fatalf("staged Mac boundary metric=%d change=%d", metricContracts, changeContracts)
				}
				if err := goose.UpContext(ctx, upgradeDB, migrationDir); err != nil {
					t.Fatalf("upgrade gfg P2.2 Mac Change boundary: %v", err)
				}
				if err := upgradeDB.QueryRowContext(ctx, `SELECT count(*) FROM gfg_change_registry WHERE detector_key='mac_support_transition' AND detector_version=1 AND status='active'`).Scan(&changeContracts); err != nil {
					t.Fatal(err)
				}
				if changeContracts != 1 {
					t.Fatalf("completed Mac Change contracts=%d", changeContracts)
				}
				upgradeActual, err := schema.Inspect(ctx, upgradeDB)
				if err != nil {
					t.Fatalf("inspect P2.2-upgraded gfg schema: %v", err)
				}
				if difference := schema.Difference(expected, upgradeActual); difference != "" {
					t.Fatalf("P2.2 Mac staged upgrade schema drift: %s", difference)
				}
			}

			// Exercise the released alpha.5 -> P0.5.1 Nav semantics repair
			// boundary, including version retirement and v2 activation.
			if test.label == "gfn" {
				upgradeName := temporaryDatabaseName(test.label, "alpha5_repair_upgrade")
				createDatabase(t, ctx, adminDB, upgradeName)
				defer dropDatabase(t, adminDB, upgradeName)
				upgradeDB := openDatabase(t, adminDSN, upgradeName)
				defer upgradeDB.Close()
				if err := goose.UpToContext(ctx, upgradeDB, migrationDir, 20260830010000); err != nil {
					t.Fatalf("prepare gfn alpha.5 fixture: %v", err)
				}
				if err := goose.UpContext(ctx, upgradeDB, migrationDir); err != nil {
					t.Fatalf("upgrade gfn alpha.5 through P0.5.1 repair: %v", err)
				}
				upgradeActual, err := schema.Inspect(ctx, upgradeDB)
				if err != nil {
					t.Fatalf("inspect P0.5.1-upgraded gfn schema: %v", err)
				}
				if difference := schema.Difference(expected, upgradeActual); difference != "" {
					t.Fatalf("alpha.5 -> P0.5.1 gfn schema drift: %s", difference)
				}
				var activeMetricV2, retiredMetricV1, activeDetectorV2, retiredDetectorV1 int
				if err := upgradeDB.QueryRowContext(ctx, `
SELECT
  count(*) FILTER (WHERE metric_version=2 AND status='active'),
  count(*) FILTER (WHERE metric_version=1 AND status='retired')
FROM gfn_metric_registry
WHERE metric_key IN ('ipv6_adoption','security_txt_adoption')`).Scan(&activeMetricV2, &retiredMetricV1); err != nil {
					t.Fatal(err)
				}
				if err := upgradeDB.QueryRowContext(ctx, `
SELECT
  count(*) FILTER (WHERE detector_version=2 AND status='active'),
  count(*) FILTER (WHERE detector_version=1 AND status='retired')
FROM gfn_change_registry
WHERE detector_key IN ('ipv6_transition','security_txt_transition')`).Scan(&activeDetectorV2, &retiredDetectorV1); err != nil {
					t.Fatal(err)
				}
				if activeMetricV2 != 2 || retiredMetricV1 != 2 || activeDetectorV2 != 2 || retiredDetectorV1 != 2 {
					t.Fatalf("gfn version cutover metrics active-v2=%d retired-v1=%d detectors active-v2=%d retired-v1=%d", activeMetricV2, retiredMetricV1, activeDetectorV2, retiredDetectorV1)
				}
			}

			// Exercise the released single-account Admin boundary through the
			// multi-account identity migration. Values that carry security or
			// historical meaning must survive unchanged, while ambiguous legacy
			// databases must fail rather than receive guessed Owner privilege.
			if test.label == "gfa" {
				upgradeName := temporaryDatabaseName(test.label, "identity_upgrade")
				createDatabase(t, ctx, adminDB, upgradeName)
				defer dropDatabase(t, adminDB, upgradeName)
				upgradeDB := openDatabase(t, adminDSN, upgradeName)
				defer upgradeDB.Close()
				if err := goose.UpToContext(ctx, upgradeDB, migrationDir, expectedBaselineVersion); err != nil {
					t.Fatalf("prepare gfa legacy identity fixture: %v", err)
				}
				if _, err := upgradeDB.ExecContext(ctx, `
INSERT INTO gfa_admin_account(id,password_hash,session_version,created_at,updated_at,password_updated_at)
VALUES (1,'legacy-password-hash',7,'2026-08-01','2026-08-02','2026-08-03');
INSERT INTO gfa_admin_audit_log(action,resource,target_id,operator,session_version,created_at)
VALUES ('legacy-action','legacy-resource','1','admin',7,'2026-08-04');`); err != nil {
					t.Fatal(err)
				}
				if err := goose.UpContext(ctx, upgradeDB, migrationDir); err != nil {
					t.Fatalf("upgrade legacy gfa identity: %v", err)
				}
				upgradeActual, err := schema.Inspect(ctx, upgradeDB)
				if err != nil {
					t.Fatal(err)
				}
				if difference := schema.Difference(expected, upgradeActual); difference != "" {
					t.Fatalf("legacy identity -> current gfa schema drift: %s", difference)
				}
				var preserved int
				if err := upgradeDB.QueryRowContext(ctx, `
SELECT count(*)
FROM gfa_admin_account account
JOIN gfa_admin_audit_log audit ON audit.operator_account_id=account.id
WHERE account.id=1 AND account.username='owner' AND account.display_name='Owner'
  AND account.role='owner' AND account.status='active'
  AND account.password_hash='legacy-password-hash' AND account.session_version=7
  AND account.created_at='2026-08-01' AND account.updated_at='2026-08-02'
  AND account.password_updated_at='2026-08-03'
  AND audit.action='legacy-action' AND audit.operator_name='Owner' AND audit.operator_role='owner'`).Scan(&preserved); err != nil {
					t.Fatal(err)
				}
				if preserved != 1 {
					t.Fatal("legacy Admin identity or audit snapshot was not preserved")
				}
				var nextID int64
				if err := upgradeDB.QueryRowContext(ctx, `SELECT nextval('gfa_admin_account_id_seq')`).Scan(&nextID); err != nil {
					t.Fatal(err)
				}
				if nextID != 2 {
					t.Fatalf("Admin account sequence next=%d, want 2", nextID)
				}

				ambiguousName := temporaryDatabaseName(test.label, "identity_ambiguous")
				createDatabase(t, ctx, adminDB, ambiguousName)
				defer dropDatabase(t, adminDB, ambiguousName)
				ambiguousDB := openDatabase(t, adminDSN, ambiguousName)
				defer ambiguousDB.Close()
				if err := goose.UpToContext(ctx, ambiguousDB, migrationDir, expectedBaselineVersion); err != nil {
					t.Fatal(err)
				}
				if _, err := ambiguousDB.ExecContext(ctx, `
INSERT INTO gfa_admin_account(id,password_hash,session_version,created_at,updated_at)
VALUES (1,'one',1,now(),now()),(2,'two',1,now(),now())`); err != nil {
					t.Fatal(err)
				}
				if err := goose.UpContext(ctx, ambiguousDB, migrationDir); err == nil {
					t.Fatal("ambiguous multi-row legacy Admin database was migrated")
				}
				var usernameColumnExists bool
				if err := ambiguousDB.QueryRowContext(ctx, `SELECT EXISTS (
  SELECT 1 FROM information_schema.columns
  WHERE table_schema='public' AND table_name='gfa_admin_account' AND column_name='username'
)`).Scan(&usernameColumnExists); err != nil {
					t.Fatal(err)
				}
				if usernameColumnExists {
					t.Fatal("failed ambiguous migration left partial identity columns")
				}
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
