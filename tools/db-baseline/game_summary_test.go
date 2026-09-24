package main

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pressly/goose/v3"
)

func TestGameSummaryLengthMigration(t *testing.T) {
	dsn := integrationAdminDSN(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	admin := openDatabase(t, dsn, "postgres")
	defer admin.Close()
	name := temporaryDatabaseName("gfg", "summary")
	createDatabase(t, ctx, admin, name)
	defer dropDatabase(t, admin, name)
	db := openDatabase(t, dsn, name)
	defer db.Close()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join("..", "..", "db", "game", "migrations")
	if err := goose.UpToContext(ctx, db, dir, 20260902020000); err != nil {
		t.Fatal(err)
	}
	assertLengths := func(want int) {
		t.Helper()
		rows, err := db.QueryContext(ctx, `SELECT character_maximum_length FROM information_schema.columns WHERE table_schema='public' AND table_name='gfg_game' AND column_name IN ('info','info_en')`)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		count := 0
		for rows.Next() {
			var n int
			if err := rows.Scan(&n); err != nil {
				t.Fatal(err)
			}
			if n != want {
				t.Fatalf("length=%d want=%d", n, want)
			}
			count++
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		if count != 2 {
			t.Fatalf("columns=%d", count)
		}
	}
	assertLengths(300)
	_, err := db.ExecContext(ctx, `INSERT INTO gfg_game (id,name,name_en,info,info_en,create_time,update_time,developers,publishers,appid,header,weight,primary_tag,secondary_tag) VALUES (1,'Game','Game',$1,$1,NOW(),NOW(),'[]','[]',1,'',0,0,0)`, strings.Repeat("中", 300))
	if err != nil {
		t.Fatal(err)
	}
	if err := goose.UpToContext(ctx, db, dir, 20260916010000); err != nil {
		t.Fatal(err)
	}
	assertLengths(400)
	var preserved string
	if err := db.QueryRowContext(ctx, `SELECT info FROM gfg_game WHERE id=1`).Scan(&preserved); err != nil {
		t.Fatal(err)
	}
	if preserved != strings.Repeat("中", 300) {
		t.Fatal("upgrade changed existing content")
	}
	for _, value := range []string{strings.Repeat("中", 400), strings.Repeat("🐺", 400)} {
		if _, err := db.ExecContext(ctx, `UPDATE gfg_game SET info=$1,info_en=$1 WHERE id=1`, value); err != nil {
			t.Fatal(err)
		}
	}
	for _, field := range []string{"info", "info_en"} {
		if _, err := db.ExecContext(ctx, `UPDATE gfg_game SET `+field+`=$1 WHERE id=1`, strings.Repeat("a", 401)); err == nil {
			t.Fatalf("%s accepted 401 characters", field)
		}
	}
	if err := goose.DownContext(ctx, db, dir); err == nil {
		t.Fatal("rollback must refuse to truncate 400-character content")
	}
	assertLengths(400)
	if _, err := db.ExecContext(ctx, `UPDATE gfg_game SET info=$1,info_en=$1 WHERE id=1`, strings.Repeat("中", 300)); err != nil {
		t.Fatal(err)
	}
	if err := goose.DownContext(ctx, db, dir); err != nil {
		t.Fatal(err)
	}
	assertLengths(300)
	if err := goose.UpToContext(ctx, db, dir, 20260916010000); err != nil {
		t.Fatal(err)
	}
	assertLengths(400)
}
