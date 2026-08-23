package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/gofurry/gofurry-nav-site/tools/internal/schema"
	"github.com/pressly/goose/v3"
	goosedb "github.com/pressly/goose/v3/database"
)

const expectedBaselineVersion int64 = 20260823000000

//go:embed expected/*.json
var expectedSchemas embed.FS

type adoptOptions struct {
	DatabaseLabel   string
	BaselineVersion int64
	RequireDBName   bool
}

func adopt(ctx context.Context, db *sql.DB, options adoptOptions) error {
	if options.BaselineVersion != expectedBaselineVersion {
		return fmt.Errorf("baseline version %d is not the expected version %d", options.BaselineVersion, expectedBaselineVersion)
	}
	if options.DatabaseLabel != "gfg" && options.DatabaseLabel != "gfn" && options.DatabaseLabel != "gfa" {
		return fmt.Errorf("unsupported database %q", options.DatabaseLabel)
	}

	expected, err := loadExpected(options.DatabaseLabel)
	if err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin adoption transaction: %w", err)
	}
	defer tx.Rollback()

	// Serialize the one-time action even if two operators accidentally start it.
	if _, err := tx.ExecContext(ctx, `select pg_catalog.pg_advisory_xact_lock($1)`, int64(0x474f465552525942)); err != nil {
		return fmt.Errorf("acquire adoption lock: %w", err)
	}
	var actualDatabase string
	if err := tx.QueryRowContext(ctx, `select current_database()`).Scan(&actualDatabase); err != nil {
		return fmt.Errorf("read current database: %w", err)
	}
	if options.RequireDBName && actualDatabase != options.DatabaseLabel {
		return fmt.Errorf("connected database is %q; expected %q", actualDatabase, options.DatabaseLabel)
	}

	var versionTableExists bool
	if err := tx.QueryRowContext(ctx, `select to_regclass('public.goose_db_version') is not null`).Scan(&versionTableExists); err != nil {
		return fmt.Errorf("check Goose version table: %w", err)
	}
	if versionTableExists {
		return errors.New("database is not eligible: public.goose_db_version already exists")
	}

	actual, err := schema.Inspect(ctx, tx)
	if err != nil {
		return fmt.Errorf("inspect existing schema: %w", err)
	}
	if difference := schema.Difference(expected, actual); difference != "" {
		expectedFingerprint, _ := schema.Fingerprint(expected)
		actualFingerprint, _ := schema.Fingerprint(actual)
		return fmt.Errorf("database schema does not exactly match the audited %s baseline: %s (expected fingerprint %s, actual %s)", options.DatabaseLabel, difference, expectedFingerprint, actualFingerprint)
	}

	store, err := goosedb.NewStore(goosedb.DialectPostgres, goose.DefaultTablename)
	if err != nil {
		return fmt.Errorf("create Goose version store: %w", err)
	}
	if err := store.CreateVersionTable(ctx, tx); err != nil {
		return fmt.Errorf("create Goose version table: %w", err)
	}
	if err := store.Insert(ctx, tx, goosedb.InsertRequest{Version: 0}); err != nil {
		return fmt.Errorf("record Goose initial version: %w", err)
	}
	if err := store.Insert(ctx, tx, goosedb.InsertRequest{Version: options.BaselineVersion}); err != nil {
		return fmt.Errorf("record baseline version: %w", err)
	}
	latest, err := store.GetLatestVersion(ctx, tx)
	if err != nil {
		return fmt.Errorf("verify Goose baseline version: %w", err)
	}
	if latest != options.BaselineVersion {
		return fmt.Errorf("Goose version verification returned %d; expected %d", latest, options.BaselineVersion)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit baseline adoption: %w", err)
	}
	return nil
}

func loadExpected(databaseLabel string) (schema.Snapshot, error) {
	data, err := expectedSchemas.ReadFile("expected/" + databaseLabel + ".json")
	if err != nil {
		return schema.Snapshot{}, fmt.Errorf("load expected %s schema: %w", databaseLabel, err)
	}
	snapshot, err := schema.Unmarshal(data)
	if err != nil {
		return schema.Snapshot{}, fmt.Errorf("decode expected %s schema: %w", databaseLabel, err)
	}
	return snapshot, nil
}
