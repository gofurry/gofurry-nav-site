package main

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pressly/goose/v3"
)

func TestPostgresSiteArchivalUpgrade(t *testing.T) {
	dsn := integrationAdminDSN(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	admin := openDatabase(t, dsn, "postgres")
	defer admin.Close()
	var major int
	if err := admin.QueryRowContext(ctx, `SELECT current_setting('server_version_num')::integer/10000`).Scan(&major); err != nil || major != 18 {
		t.Fatalf("isolated PG18 required: %d %v", major, err)
	}
	name := temporaryDatabaseName("gfn", "site_archival_upgrade")
	createDatabase(t, ctx, admin, name)
	defer dropDatabase(t, admin, name)
	db := openDatabase(t, dsn, name)
	defer db.Close()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatal(err)
	}
	migrations, err := filepath.Abs("../../db/nav/migrations")
	if err != nil {
		t.Fatal(err)
	}
	if err := goose.UpToContext(ctx, db, migrations, 20261018010000); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO gfn_site(id,name,name_en,info,info_en,create_time,update_time,nsfw,welfare,deleted,deleted_at,view_count) VALUES(1,'Active','Active','original','original',now(),now(),'0','0',false,NULL,5),(2,'Legacy','Legacy','history','history',now(),now(),'0','0',true,'2026-09-01',8)`); err != nil {
		t.Fatal(err)
	}
	var before, after string
	if err := db.QueryRowContext(ctx, `SELECT jsonb_agg(to_jsonb(s) ORDER BY id)::text FROM gfn_site s`).Scan(&before); err != nil {
		t.Fatal(err)
	}
	if err := goose.UpContext(ctx, db, migrations); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT jsonb_agg(to_jsonb(s)-'deletion_reason_code'-'deletion_reason' ORDER BY id)::text FROM gfn_site s`).Scan(&after); err != nil || before != after {
		t.Fatalf("migration rewrote content/history: %v", err)
	}
	var count int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM gfn_site WHERE deletion_reason_code IS NULL AND deletion_reason IS NULL`).Scan(&count); err != nil || count != 2 {
		t.Fatal("invented historical reasons", err)
	}
	for _, statement := range []string{
		`UPDATE gfn_site SET deletion_reason_code='closed' WHERE id=1`,
		`UPDATE gfn_site SET deletion_reason='active note' WHERE id=1`,
		`UPDATE gfn_site SET deletion_reason_code='invalid' WHERE id=2`,
		`UPDATE gfn_site SET deletion_reason_code='' WHERE id=2`,
		`UPDATE gfn_site SET deletion_reason=repeat('字',1001) WHERE id=2`,
	} {
		if _, err := db.ExecContext(ctx, statement); err == nil {
			t.Fatalf("constraint accepted %s", statement)
		}
	}
	if _, err := db.ExecContext(ctx, `UPDATE gfn_site SET deletion_reason_code='other',deletion_reason=repeat('🦊',1000) WHERE id=2`); err != nil {
		t.Fatal("valid Unicode boundary", err)
	}
	if err := goose.DownContext(ctx, db, migrations); err == nil || !strings.Contains(err.Error(), "retained history") {
		t.Fatalf("destructive down must reject: %v", err)
	}
	if version, err := goose.GetDBVersionContext(ctx, db); err != nil || version != 20261018020000 {
		t.Fatalf("failed Down changed version: %d %v", version, err)
	}
}
