package main

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pressly/goose/v3"
)

func assertHybridCollectionMigration(t *testing.T, ctx context.Context, admin *sql.DB, dsn, root string) {
	t.Helper()
	name := temporaryDatabaseName("gfg", "hybrid_upgrade")
	createDatabase(t, ctx, admin, name)
	defer dropDatabase(t, admin, name)
	db := openDatabase(t, dsn, name)
	defer db.Close()
	dir := filepath.Join(root, "db", "game", "migrations")
	if err := goose.UpToContext(ctx, db, dir, 20261006020000); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO gfg_game(id,name,name_en,info,info_en,appid,header,developers,publishers,weight,create_time,update_time)
VALUES(144,'旧成员','Existing member','','',144,'','[]','[]',0,now(),now());
INSERT INTO gfg_game_collection(id,code,name,name_en) VALUES(144,'old-manual','旧分区','Old collection');
INSERT INTO gfg_game_collection_item(collection_id,game_id,created_at) VALUES(144,144,'2026-10-06');
INSERT INTO gfg_tag_category(id,code,name,name_en,info,info_en,sort_order,create_time,update_time)
VALUES(144,'hybrid-upgrade','类别','Category','','',0,now(),now());
INSERT INTO gfg_tag(id,code,name,name_en,info,info_en,category_id,create_time,update_time)
VALUES(144,'hybrid-upgrade','标签','Tag','','',144,now(),now());`); err != nil {
		t.Fatal(err)
	}
	if err := goose.UpToContext(ctx, db, dir, 20261008010000); err != nil {
		t.Fatal(err)
	}
	var manual, rules, exclusions int
	if err := db.QueryRowContext(ctx, `SELECT
(SELECT count(*) FROM gfg_game_collection_item WHERE collection_id=144 AND game_id=144 AND created_at='2026-10-06'),
(SELECT count(*) FROM gfg_game_collection_tag),
(SELECT count(*) FROM gfg_game_collection_exclusion)`).Scan(&manual, &rules, &exclusions); err != nil || manual != 1 || rules != 0 || exclusions != 0 {
		t.Fatal("migration must preserve manual facts without seeding rules", manual, rules, exclusions, err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO gfg_game_collection_tag(collection_id,tag_id) VALUES(144,144);
INSERT INTO gfg_game_collection_exclusion(collection_id,game_id) VALUES(144,144);`); err != nil {
		t.Fatal(err)
	}
	if err := goose.DownContext(ctx, db, dir); err == nil || !strings.Contains(err.Error(), "destroy curated rules and exclusions") {
		t.Fatal("Down must explicitly refuse destructive rollback", err)
	}
	if err := db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM gfg_game_collection_tag), (SELECT count(*) FROM gfg_game_collection_exclusion)`).Scan(&rules, &exclusions); err != nil || rules != 1 || exclusions != 1 {
		t.Fatal("failed Down must preserve configuration", rules, exclusions, err)
	}
	if version, err := goose.GetDBVersionContext(ctx, db); err != nil || version != 20261008010000 {
		t.Fatal("failed Down must preserve Goose history", version, err)
	}
}
