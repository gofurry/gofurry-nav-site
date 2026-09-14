package service

import (
	"context"
	"os"
	"testing"
	"time"

	navsqlc "github.com/gofurry/gofurry-nav-backend/internal/db/nav/sqlc"
	"github.com/jackc/pgx/v5"
)

func TestRealDevAppearanceQueries(t *testing.T) {
	dsn := os.Getenv("GOFURRY_NAV_ASSET_TEST_URL")
	if dsn == "" {
		t.Skip("set GOFURRY_NAV_ASSET_TEST_URL for transactional temporary-table acceptance")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal("development PostgreSQL connection failed")
	}
	defer conn.Close(context.Background())
	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(context.Background())
	// Clone Goose-owned table definitions into this connection's temporary schema.
	// The production tables and records are never changed by query acceptance.
	_, err = tx.Exec(ctx, `CREATE TEMP TABLE gfn_home_hero_asset (LIKE public.gfn_home_hero_asset INCLUDING ALL) ON COMMIT DROP;
CREATE TEMP TABLE gfn_background_pattern (LIKE public.gfn_background_pattern INCLUDING ALL) ON COMMIT DROP;
INSERT INTO pg_temp.gfn_home_hero_asset(id,variant,name,object_key,enabled,deleted) VALUES
(1,'desktop','visible','nav/hero/desktop/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.avif',true,false),
(2,'desktop','disabled','nav/hero/desktop/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb.avif',false,false),
(3,'desktop','deleted','nav/hero/desktop/cccccccccccccccccccccccccccccccc.avif',true,true);
INSERT INTO pg_temp.gfn_background_pattern(id,name,name_en,object_key,light_color,dark_color,light_opacity,dark_opacity,default_size_px,enabled,sort_order,deleted) VALUES
(1,'later','later','nav/patterns/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.svg','#123456','#abcdef',0.1,0.2,96,true,20,false),
(2,'first','first','nav/patterns/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb.svg','#123456','#abcdef',0.25,0.5,128,true,10,false),
(3,'disabled','disabled','nav/patterns/cccccccccccccccccccccccccccccccc.svg','#123456','#abcdef',0.1,0.2,96,false,0,false),
(4,'deleted','deleted','nav/patterns/dddddddddddddddddddddddddddddddd.svg','#123456','#abcdef',0.1,0.2,96,true,0,true);`)
	if err != nil {
		t.Fatal(err)
	}
	svc := newHomeService(nil, time.Now).WithAppearance(navsqlc.New(tx))
	// A transaction is one connection, so exercise sqlc directly for both pools;
	// concurrent request-level reads are covered by TestIndependentHeroPools.
	q := navsqlc.New(tx)
	hero, err := q.RandomHeroAsset(ctx, "desktop")
	if err != nil || hero.ID != 1 {
		t.Fatal("desktop selection included disabled/deleted asset")
	}
	if _, err := q.RandomHeroAsset(ctx, "mobile"); err != pgx.ErrNoRows {
		t.Fatal("empty mobile pool crossed to desktop")
	}
	patterns, queryErr := svc.GetPatterns(ctx)
	if queryErr != nil || len(patterns.Patterns) != 2 || patterns.Patterns[0].ID != 2 || patterns.Patterns[0].LightOpacity != 0.25 {
		t.Fatal("catalog filtering, sorting or numeric conversion failed")
	}
	if _, err := tx.Exec(ctx, `UPDATE pg_temp.gfn_home_hero_asset SET deleted=true WHERE id=1`); err != nil {
		t.Fatal(err)
	}
	if _, err := q.RandomHeroAsset(ctx, "desktop"); err != pgx.ErrNoRows {
		t.Fatal("soft-deleted Hero remained visible")
	}
	t.Log("real PostgreSQL appearance query acceptance passed; only temporary fixtures were used")
}
