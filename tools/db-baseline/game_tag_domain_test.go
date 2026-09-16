package main

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pressly/goose/v3"
)

const legacyTagFixture = `
INSERT INTO gfg_tag(id,name,name_en,info,info_en,prefix,create_time,update_time) VALUES
(100,'分类','Categories','','',-1,'2020-01-01','2020-01-02'),
(200,'物种','Species','','',-1,'2020-01-01','2020-01-02'),
(300,'平台','Platforms','','',-1,'2020-01-01','2020-01-02'),
(900,'其他','Other','','',-1,'2020-01-01','2020-01-02'),
(1001,'FVN','Changed display name','','',100,'2020-01-01','2020-01-02'),
(1014,'成人','Adult','','',100,'2020-01-01','2020-01-02'),
(2014,'狼','Wolf','','',200,'2020-01-01','2020-01-02');
INSERT INTO gfg_game(id,name,name_en,info,info_en,create_time,update_time,developers,publishers,appid,header,weight,primary_tag,secondary_tag)
VALUES(116,'test','test','','','2020-01-01','2021-01-01','[]','[]',116,'',0,1001,2014),
(117,'missing','missing','','','2020-01-01','2021-01-01','[]','[]',117,'',0,2014,1001);
INSERT INTO gfg_tag_map(id,game_id,tag_id,create_time,update_time) VALUES
(1,116,1001,'2020-02-01','2020-03-01'),(2,116,1014,'2020-02-01','2020-03-01');
INSERT INTO gfg_game_tracking_periods(game_id,appid,tracked_from,tracking_basis,opened_reason)
VALUES(116,116,transaction_timestamp()-interval '1 day','explicit','created');
SELECT gfg_refresh_current_game_daily(116,'observed');
INSERT INTO gfg_game_daily
SELECT (jsonb_populate_record(NULL::gfg_game_daily,to_jsonb(d) || jsonb_build_object(
 'fact_date',(transaction_timestamp() AT TIME ZONE 'UTC')::date-1,'finalized_at',transaction_timestamp()))).*
FROM gfg_game_daily d WHERE game_id=116;

`

func TestGameTagDomainMigration(t *testing.T) {
	dsn := integrationAdminDSN(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	admin := openDatabase(t, dsn, "postgres")
	defer admin.Close()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join("..", "..", "db", "game", "migrations")
	for _, tc := range []struct{ name, bad, want string }{
		{"preserves identities and history", "", ""},
		{"unmapped", `UPDATE gfg_tag SET id=8888 WHERE id=1014`, "unmapped legacy leaf"},
		{"orphan", `UPDATE gfg_tag SET prefix=999 WHERE id=1014`, "orphan"},
		{"parent assigned", `UPDATE gfg_tag_map SET tag_id=100 WHERE id=2`, "parent assigned"},
		{"duplicate", `INSERT INTO gfg_tag_map SELECT 3,game_id,tag_id,create_time,update_time FROM gfg_tag_map WHERE id=1`, "duplicate"},
		{"invalid primary", `UPDATE gfg_game SET primary_tag=9999`, "invalid primary"},
		{"equal roles", `UPDATE gfg_game SET secondary_tag=primary_tag`, "primary equals secondary"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			name := temporaryDatabaseName("gfg", "tags")
			createDatabase(t, ctx, admin, name)
			defer dropDatabase(t, admin, name)
			db := openDatabase(t, dsn, name)
			defer db.Close()
			if err := goose.UpToContext(ctx, db, dir, 20260916010000); err != nil {
				t.Fatal(err)
			}
			if _, err := db.ExecContext(ctx, legacyTagFixture); err != nil {
				t.Fatal(err)
			}
			var history string
			if err := db.QueryRowContext(ctx, `SELECT jsonb_agg(to_jsonb(d) ORDER BY fact_date)::text FROM gfg_game_daily d WHERE game_id=116`).Scan(&history); err != nil {
				t.Fatal(err)
			}
			if tc.bad != "" {
				if _, err := db.ExecContext(ctx, tc.bad); err != nil {
					t.Fatal(err)
				}
			}
			err := goose.UpContext(ctx, db, dir)
			if tc.want != "" {
				if err == nil || !strings.Contains(err.Error(), tc.want) {
					t.Fatalf("got %v, want rejection %s", err, tc.want)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var after string
			if err := db.QueryRowContext(ctx, `SELECT jsonb_agg(to_jsonb(d) ORDER BY fact_date)::text FROM gfg_game_daily d WHERE game_id=116`).Scan(&after); err != nil {
				t.Fatal(err)
			}
			if history != after {
				t.Fatal("migration rewrote historical fact")
			}
			for _, q := range []string{
				`SELECT count(*)=3 FROM gfg_tag`,
				`SELECT code='fvn' AND id=1001 AND category_id=1 FROM gfg_tag WHERE id=1001`,
				`SELECT count(*)=3 FROM gfg_game_tag WHERE game_id=116`,
				`SELECT count(*)=5 FROM gfg_game_tag`,
				`SELECT role='primary' AND create_time='2021-01-01' FROM gfg_game_tag WHERE game_id=117 AND tag_id=2014`,
				`SELECT role='primary' AND create_time='2020-02-01' AND update_time='2021-01-01' FROM gfg_game_tag WHERE game_id=116 AND tag_id=1001`,
				`SELECT role='secondary' AND create_time='2021-01-01' AND update_time='2021-01-01' FROM gfg_game_tag WHERE game_id=116 AND tag_id=2014`,
				`SELECT role='normal' AND update_time='2020-03-01' FROM gfg_game_tag WHERE game_id=116 AND tag_id=1014`,
				`SELECT to_regclass('public.gfg_tag_map') IS NULL`,
				`SELECT count(*)=0 FROM information_schema.columns WHERE table_schema='public' AND ((table_name='gfg_tag' AND column_name='prefix') OR (table_name='gfg_game' AND column_name IN ('primary_tag','secondary_tag')))`,
			} {
				assertTagSQL(t, ctx, db, q)
			}
			t.Log("preserved 3 leaf IDs and 2 old pairs; synthesized 1 primary + 2 secondary relations; final pairs=5; existing history unchanged")
			var nextID int64
			if err := db.QueryRowContext(ctx, `INSERT INTO gfg_tag(code,category_id,name,name_en,info,info_en,create_time,update_time) VALUES('future-tag',1,'new','new','','',NOW(),NOW()) RETURNING id`).Scan(&nextID); err != nil {
				t.Fatal(err)
			}
			if nextID <= 2014 {
				t.Fatalf("new id reused historical identity: %d", nextID)
			}
			if _, err := db.ExecContext(ctx, `SELECT gfg_refresh_current_game_daily(116,'observed')`); err != nil {
				t.Fatal(err)
			}
			assertTagSQL(t, ctx, db, `SELECT projection_version=2 AND primary_tag_id=1001 AND secondary_tag_id=2014 AND tag_ids=ARRAY[1001,1014,2014]::bigint[] FROM gfg_game_daily WHERE game_id=116 AND fact_date=(transaction_timestamp() AT TIME ZONE 'UTC')::date`)
			if _, err := db.ExecContext(ctx, `SELECT * FROM gfg_project_state_fact_day((transaction_timestamp() AT TIME ZONE 'UTC')::date)`); err != nil {
				t.Fatal(err)
			}
			assertTagSQL(t, ctx, db, `SELECT projection_version=2 FROM gfg_game_daily WHERE game_id=116 AND fact_date=(transaction_timestamp() AT TIME ZONE 'UTC')::date`)
			assertTagSQL(t, ctx, db, `SELECT projection_version=1 AND tag_ids=ARRAY[1001,1014]::bigint[] FROM gfg_game_daily WHERE game_id=116 AND fact_date=(transaction_timestamp() AT TIME ZONE 'UTC')::date-1`)
			if err := goose.DownContext(ctx, db, dir); err == nil || !strings.Contains(err.Error(), "irreversible") {
				t.Fatalf("expected explicit irreversible Down: %v", err)
			}
		})
	}
}

func assertTagSQL(t *testing.T, ctx context.Context, db *sql.DB, q string) {
	t.Helper()
	var ok bool
	if err := db.QueryRowContext(ctx, q).Scan(&ok); err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("invariant failed: %s", q)
	}
}
