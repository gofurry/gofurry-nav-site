package dao_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	v2controller "github.com/gofurry/gofurry-game-backend/apps/game/v2/controller"
	v2dao "github.com/gofurry/gofurry-game-backend/apps/game/v2/dao"
	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	v2service "github.com/gofurry/gofurry-game-backend/apps/game/v2/service"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Runs inside the existing disposable-database CI gate, never against the
// supplied config's application database and never as a separately skipped test.
func assertCollectionReadModels(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, query, args...); err != nil {
			t.Fatalf("collection fixture: %v", err)
		}
	}
	reader := v2dao.NewReadModelDAO(pool)
	svc := v2service.NewCollectionService(reader, nil)
	asOf := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	svc.Now = func() time.Time { return asOf }
	index, err := svc.List(ctx, v2models.CollectionQuery{})
	if err != nil || index.Total != 0 || index.Items == nil || len(index.Items) != 0 {
		t.Fatalf("migration seeded collections: %+v %v", index, err)
	}
	home, err := svc.Home(ctx, v2models.CollectionQuery{})
	if err != nil || home.Slots == nil || len(home.Slots) != 0 {
		t.Fatalf("empty home=%+v %v", home, err)
	}
	for id := int64(140101); id <= 140109; id++ {
		exec(`INSERT INTO gfg_game(id,name,name_en,info,info_en,appid,header,developers,publishers,weight,create_time,update_time)
VALUES($1,'游戏','Game','简介','Summary',$1,'https://example.test/site-header.jpg','[]','[]',0,now(),now())`, id)
	}
	exec(`INSERT INTO gfg_game_collection(id,code,name,name_en,info,info_en,status,published_at,archived_at) VALUES
(140001,'chronicle','中文谱系','Chronicle','中文简介','','published','2026-10-01',NULL),
(140002,'adult-chronicle','成人谱系','Adult chronicle','','','published','2026-09-01',NULL),
(140003,'draft-chronicle','草稿','Draft','','','draft',NULL,NULL),
(140004,'archived-chronicle','归档','Archived','','','archived','2026-08-01','2026-09-01'),
(140005,'empty-chronicle','空','Empty','','','published','2026-08-01',NULL),
(140006,'later-chronicle','同日较新','Later','','','published','2026-10-01',NULL)`)
	exec(`INSERT INTO gfg_game_collection_item(collection_id,game_id) SELECT 140001,id FROM gfg_game WHERE id BETWEEN 140101 AND 140109`)
	exec(`INSERT INTO gfg_game_collection_item(collection_id,game_id) VALUES (140002,140108),(140003,140101),(140004,140101),(140006,140101)`)
	exec(`INSERT INTO gfg_game_collection_home_slot(slot,collection_id) VALUES (1,140001),(2,140002),(3,140003),(4,140004),(5,140005)`)
	exec(`INSERT INTO gfg_tag(id,code,name,name_en,info,info_en,category_id,create_time,update_time)
VALUES(1014,'not-adult-140','成人名字不是语义','adult','','',1,now(),now())`)
	exec(`INSERT INTO gfg_game_tag(game_id,tag_id,role,create_time,update_time) VALUES (140109,1014,'normal',now(),now())`)
	exec(`UPDATE gfg_tag SET archived_at=now() WHERE code='adult'`)
	exec(`INSERT INTO gfg_game_tag(game_id,tag_id,role,create_time,update_time) SELECT 140108,id,'normal',now(),now() FROM gfg_tag WHERE code='adult'`)
	exec(`INSERT INTO gfg_game_release_state(game_id,availability,precision,exact_date,release_year,release_month,window_start,window_end,raw_text,source,source_region,source_locale,normalizer_version,observed_at) VALUES
(140101,'upcoming','day','2027-01-01',2027,1,'2027-01-01','2027-01-01','1 Jan 2027','steam','US','en','test',now()),
(140102,'available','month',NULL,2020,2,'2020-02-01','2020-02-29','Feb 2020','steam','US','en','test',now()),
(140103,'available','none',NULL,NULL,NULL,NULL,NULL,'available','steam','US','en','test',now()),
(140104,'upcoming','day','2026-10-05',2026,10,'2026-10-05','2026-10-05','5 Oct 2026','steam','US','en','test',now()),
(140105,'upcoming','day','2026-10-06',2026,10,'2026-10-06','2026-10-06','6 Oct 2026','steam','US','en','test',now()),
(140106,'upcoming','tba',NULL,NULL,NULL,NULL,NULL,'TBA','steam','US','en','test',now()),
(140107,'unknown','unknown',NULL,NULL,NULL,NULL,NULL,'2026-01-01','steam','US','en','test',now())`)
	exec(`INSERT INTO gfg_game_first_available(game_id,precision,release_year,window_start,window_end,source,inferred,normalizer_version)
VALUES(140101,'year',2019,'2019-01-01','2019-12-31','steam_backfill',true,'test')`)
	exec(`INSERT INTO gfg_game_localized_details(game_id,appid,lang,name,short_description,collected_at,updated_at)
VALUES(140101,140101,'en','Localized Game','Localized summary',now(),now())`)
	exec(`INSERT INTO gfg_game_assets(game_id,appid,asset_type,asset_family,source,lang,media_key,url,exists,collected_at,updated_at)
VALUES(140101,140101,'header','store','store_browse','en','header','https://example.test/asset-header.jpg',true,now(),now())`)

	t.Run("published index and paging", func(t *testing.T) {
		index, e := svc.List(ctx, v2models.CollectionQuery{Lang: "en"})
		if e != nil || index.Total != 4 || len(index.Items) != 4 || index.HasMore {
			t.Fatalf("index=%+v %v", index, e)
		}
		var codes []string
		for _, item := range index.Items {
			codes = append(codes, item.Code)
		}
		if !reflect.DeepEqual(codes, []string{"later-chronicle", "chronicle", "adult-chronicle", "empty-chronicle"}) {
			t.Fatal("publication ordering", codes)
		}
		if index.Items[1].Name != "Chronicle" || index.Items[1].Info != "中文简介" || index.Items[1].VisibleGameCount != 8 || index.Items[2].VisibleGameCount != 0 || len(index.Items[2].PreviewGames) != 0 {
			t.Fatal("localized/filter counts", index.Items)
		}
		var previews []string
		for _, p := range index.Items[1].PreviewGames {
			previews = append(previews, p.GameID)
		}
		if !reflect.DeepEqual(previews, []string{"140101", "140105", "140109"}) {
			t.Fatal("filtered chronological preview", previews)
		}
		page, e := svc.List(ctx, v2models.CollectionQuery{Page: 2, PageSize: 2})
		if e != nil || page.Total != 4 || page.HasMore || len(page.Items) != 2 || page.Items[0].Code != "adult-chronicle" {
			t.Fatalf("page=%+v %v", page, e)
		}
		page, e = svc.List(ctx, v2models.CollectionQuery{Page: 1, PageSize: 2})
		if e != nil || !page.HasMore {
			t.Fatal("has_more", page, e)
		}
		page, e = svc.List(ctx, v2models.CollectionQuery{Page: 3, PageSize: 2})
		if e != nil || page.Items == nil || len(page.Items) != 0 || page.Total != 4 || page.HasMore {
			t.Fatal("empty page", page, e)
		}
	})
	t.Run("timeline and adult semantics", func(t *testing.T) {
		for _, mode := range []string{"sfw", "nsfw"} {
			q := v2models.CollectionQuery{Lang: "en", Mode: mode}
			detail, e := svc.Detail(ctx, "chronicle", q)
			count := 8
			if mode == "nsfw" {
				count = 9
			}
			if e != nil || len(detail.Items) != count || detail.Collection.VisibleGameCount != count {
				t.Fatalf("%s detail=%+v %v", mode, detail, e)
			}
			wantPhases := []string{"released", "released", "released_unknown", "upcoming_overdue", "upcoming", "upcoming_tba", "unknown", "unknown"}
			if mode == "nsfw" {
				wantPhases = append(wantPhases, "unknown")
			}
			for i, item := range detail.Items {
				if item.Phase != wantPhases[i] {
					t.Fatalf("phase %d=%s", i, item.Phase)
				}
				if mode == "sfw" && item.GameID == "140108" {
					t.Fatal("archived adult tag bypassed")
				}
			}
			first := detail.Items[0]
			if first.GameID != "140101" || first.Name != "Game" || first.Summary != "Summary" || first.HeaderURL != "https://example.test/asset-header.jpg" || first.Chronology.Source != "first_available" || first.Chronology.Precision != "year" || !first.Chronology.Inferred || first.Chronology.WindowStart != "2019-01-01" {
				t.Fatalf("first available/media/localization=%+v", first)
			}
			if detail.Items[1].Chronology.Precision != "month" || detail.Items[1].Chronology.Source != "release" || detail.Items[2].Chronology != nil {
				t.Fatal("precision/null chronology drift")
			}
			if detail.Items[len(detail.Items)-1].GameID != "140109" {
				t.Fatal("numeric 1014 treated as adult")
			}
			adult, e := svc.Detail(ctx, "adult-chronicle", q)
			adultCount := 0
			if mode == "nsfw" {
				adultCount = 1
			}
			if e != nil || adult.Items == nil || len(adult.Items) != adultCount {
				t.Fatal("adult only detail must exist", adult, e)
			}
			home, e := svc.Home(ctx, q)
			slots := 1
			if mode == "nsfw" {
				slots = 2
			}
			if e != nil || len(home.Slots) != slots || home.Slots[0].Slot != 1 {
				t.Fatal("home published/current mode visibility", home, e)
			}
			if slots == 2 && home.Slots[1].Slot != 2 {
				t.Fatal("slot ordering")
			}
			body, _ := json.Marshal(detail)
			for _, forbidden := range []string{"hidden_adult_count", "raw_count", "steam_backfill", "legacy_manual", "observed_transition", "archived_at", "\"status\"", "\"version\""} {
				if strings.Contains(string(body), forbidden) {
					t.Fatal("private metadata exposed", forbidden)
				}
			}
		}
	})
	t.Run("public lifecycle HTTP", func(t *testing.T) {
		api := v2controller.New(reader, nil, nil, nil).WithCollections(svc)
		app := fiber.New()
		app.Get("/collections/home", api.GetCollectionHome)
		app.Get("/collections", api.GetCollections)
		app.Get("/collections/:code", api.GetCollection)
		for _, tc := range []struct {
			path   string
			status int
		}{{"/collections", 200}, {"/collections/home", 200}, {"/collections/chronicle", 200}, {"/collections/adult-chronicle", 200}, {"/collections/draft-chronicle", 404}, {"/collections/archived-chronicle", 404}, {"/collections/nonexistent", 404}, {"/collections/bad--code", 404}} {
			res, e := app.Test(httptest.NewRequest("GET", tc.path, nil))
			if e != nil {
				t.Fatal(e)
			}
			var body struct {
				Code int             `json:"code"`
				Data json.RawMessage `json:"data"`
			}
			e = json.NewDecoder(res.Body).Decode(&body)
			res.Body.Close()
			if e != nil || res.StatusCode != tc.status {
				t.Fatal(tc.path, res.StatusCode, e)
			}
			if tc.status == 404 && string(body.Data) != `"Collection not found"` {
				t.Fatal("lifecycle disclosure", string(body.Data))
			}
		}
	})
	t.Run("Home reads only placement and visible membership", func(t *testing.T) {
		tracer := &collectionQueryTrace{}
		cfg := pool.Config()
		cfg.ConnConfig.Tracer = tracer
		traced, e := pgxpool.NewWithConfig(ctx, cfg)
		if e != nil {
			t.Fatal(e)
		}
		defer traced.Close()
		homeService := v2service.NewCollectionService(v2dao.NewReadModelDAO(traced), nil)
		home, e := homeService.Home(ctx, v2models.CollectionQuery{Lang: "en"})
		if e != nil || len(home.Slots) != 1 || home.Slots[0].Collection.VisibleGameCount != 8 || home.Slots[0].Collection.Name != "Chronicle" {
			t.Fatalf("home=%+v err=%v", home, e)
		}
		if home.Slots[0].Collection.PreviewGames == nil || len(home.Slots[0].Collection.PreviewGames) != 0 {
			t.Fatal("Home preview must stay empty")
		}
		if tracer.count != 1 {
			t.Fatalf("Home must use one SQL read, got %d", tracer.count)
		}
		for _, forbidden := range []string{"gfg_game_localized_details", "gfg_game_assets", "gfg_game_media", "gfg_game_release", "gfg_game_first_available", "price", "player", "review"} {
			if strings.Contains(strings.Join(tracer.queries, "\n"), forbidden) {
				t.Fatal("heavy Home read:", forbidden)
			}
		}
	})
	t.Run("bounded batches", func(t *testing.T) {
		tracer := &collectionQueryTrace{}
		cfg := pool.Config()
		cfg.ConnConfig.Tracer = tracer
		traced, e := pgxpool.NewWithConfig(ctx, cfg)
		if e != nil {
			t.Fatal(e)
		}
		defer traced.Close()
		dao := v2dao.NewReadModelDAO(traced)
		small, e := dao.LoadCollectionProjectionGames(ctx, []int64{140002}, "en")
		if e != nil || len(small.Games) != 1 {
			t.Fatal(small, e)
		}
		firstCount := tracer.count
		tracer.count = 0
		tracer.siteIDs = nil
		large, e := dao.LoadCollectionProjectionGames(ctx, []int64{140001, 140002}, "en")
		if e != nil || len(large.Games) != 9 || len(large.Memberships) != 10 {
			t.Fatal("batch dedupe", large, e)
		}
		if tracer.count != firstCount || firstCount != 6 || len(tracer.siteIDs) != 9 {
			t.Fatalf("N+1 or duplicate IDs: small=%d large=%d ids=%v", firstCount, tracer.count, tracer.siteIDs)
		}
		for _, forbidden := range []string{"gfg_game_prices", "gfg_game_player", "gfg_game_comment", "gfg_game_requirements", "gfg_game_news", "gfg_game_recommendations"} {
			if strings.Contains(strings.Join(tracer.queries, "\n"), forbidden) {
				t.Fatal("unexpected projection read", forbidden)
			}
		}
		service := v2service.NewCollectionService(dao, nil)
		for _, endpoint := range []string{"list", "detail"} {
			tracer.count = 0
			tracer.queries = nil
			start := time.Now()
			if endpoint == "list" {
				_, e = service.List(ctx, v2models.CollectionQuery{})
			} else {
				_, e = service.Detail(ctx, "chronicle", v2models.CollectionQuery{})
			}
			want := 8
			if endpoint == "detail" {
				want = 8
			}
			if e != nil || tracer.count != want {
				t.Fatalf("%s queries=%d want=%d err=%v", endpoint, tracer.count, want, e)
			}
			forbiddenTables := []string{"gfg_game_prices", "gfg_game_requirements", "gfg_game_news", "gfg_game_player_peaks"}
			if endpoint == "list" {
				forbiddenTables = append(forbiddenTables, "gfg_game_player", "gfg_game_comment")
			}
			for _, forbidden := range forbiddenTables {
				if strings.Contains(strings.Join(tracer.queries, "\n"), forbidden) {
					t.Fatal(endpoint, forbidden)
				}
			}
			t.Logf("cold %s: %d bounded queries, %s", endpoint, tracer.count, time.Since(start))
		}
		t.Logf("one member and ten memberships both use %d SQL queries; nine unique games loaded once", firstCount)
	})
	assertCollectionDiscovery(t, ctx, pool)
	assertCollectionConstraints(t, ctx, pool)
	assertCollectionTimelineDecorations(t, ctx, pool)
	assertHybridCollectionReads(t, ctx, pool)
}

type collectionQueryTrace struct {
	mu              sync.Mutex
	count           int
	siteIDs         []int64
	decorationIDs   []int64
	decorationCount int
	queries         []string
}

func (q *collectionQueryTrace) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.count++
	q.queries = append(q.queries, data.SQL)
	if strings.Contains(data.SQL, "-- name: BatchCollectionTimelineDecorations") {
		q.decorationIDs = append([]int64{}, data.Args[1].([]int64)...)
		q.decorationCount++
	}
	if strings.Contains(data.SQL, "-- name: BatchCollectionProjectionGames") {
		q.siteIDs = append([]int64{}, data.Args[0].([]int64)...)
	}
	return ctx
}
func (*collectionQueryTrace) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

func assertCollectionConstraints(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	for _, tc := range []struct{ name, query, code string }{
		{"invalid code", `INSERT INTO gfg_game_collection(code,name,name_en) VALUES('Bad_code','中文','English')`, "23514"},
		{"too long code", `INSERT INTO gfg_game_collection(code,name,name_en) VALUES(repeat('a',65),'中文','English')`, "23514"},
		{"blank name", `INSERT INTO gfg_game_collection(code,name,name_en) VALUES('blank',' ','English')`, "23514"},
		{"blank English name", `INSERT INTO gfg_game_collection(code,name,name_en) VALUES('blank-en','中文',' ')`, "23514"},
		{"invalid status", `INSERT INTO gfg_game_collection(code,name,name_en,status) VALUES('status','中文','English','deleted')`, "23514"},
		{"invalid version", `INSERT INTO gfg_game_collection(code,name,name_en,version) VALUES('version','中文','English',0)`, "23514"},
		{"publish requires date", `INSERT INTO gfg_game_collection(code,name,name_en,status) VALUES('publish','中文','English','published')`, "23514"},
		{"archive requires date", `INSERT INTO gfg_game_collection(code,name,name_en,status) VALUES('archive','中文','English','archived')`, "23514"},
		{"draft cannot be archived", `UPDATE gfg_game_collection SET archived_at=now() WHERE id=140003`, "23514"},
		{"duplicate code", `INSERT INTO gfg_game_collection(code,name,name_en) VALUES('chronicle','中文','English')`, "23505"},
		{"duplicate membership", `INSERT INTO gfg_game_collection_item(collection_id,game_id) VALUES(140001,140101)`, "23505"},
		{"missing collection", `INSERT INTO gfg_game_collection_item(collection_id,game_id) VALUES(999999,140101)`, "23503"},
		{"missing game", `INSERT INTO gfg_game_collection_item(collection_id,game_id) VALUES(140001,999999)`, "23503"},
		{"slot zero", `INSERT INTO gfg_game_collection_home_slot(slot,collection_id) VALUES(0,140006)`, "23514"},
		{"slot six", `INSERT INTO gfg_game_collection_home_slot(slot,collection_id) VALUES(6,140006)`, "23514"},
		{"duplicate slot", `INSERT INTO gfg_game_collection_home_slot(slot,collection_id) VALUES(1,140006)`, "23505"},
		{"duplicate home collection", `UPDATE gfg_game_collection_home_slot SET collection_id=140001 WHERE slot=2`, "23505"},
		{"home FK", `UPDATE gfg_game_collection_home_slot SET collection_id=999999 WHERE slot=2`, "23503"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, e := pool.Exec(ctx, tc.query)
			var pgErr *pgconn.PgError
			if !errors.As(e, &pgErr) || pgErr.Code != tc.code {
				t.Fatalf("expected SQLSTATE %s, got %v", tc.code, e)
			}
		})
	}
	t.Run("cascade and historical publish time", func(t *testing.T) {
		tx, e := pool.Begin(ctx)
		if e != nil {
			t.Fatal(e)
		}
		defer tx.Rollback(ctx)
		for _, sql := range []string{
			`UPDATE gfg_game_collection SET status='draft' WHERE id=140001`,
			`UPDATE gfg_game_collection SET status='draft',archived_at=NULL WHERE id=140004`,
			`DELETE FROM gfg_game_tag WHERE game_id=140108`,
			`DELETE FROM gfg_game WHERE id=140108`,
		} {
			if _, e = tx.Exec(ctx, sql); e != nil {
				t.Fatal(e)
			}
		}
		var count int
		if e = tx.QueryRow(ctx, `SELECT count(*) FROM gfg_game_collection_item WHERE game_id=140108`).Scan(&count); e != nil || count != 0 {
			t.Fatal("game cascade", count, e)
		}
		if e = tx.QueryRow(ctx, `SELECT count(*) FROM gfg_game_collection WHERE id=140001 AND status='draft' AND published_at IS NOT NULL`).Scan(&count); e != nil || count != 1 {
			t.Fatal("draft publication history", count, e)
		}
		if _, e = tx.Exec(ctx, `DELETE FROM gfg_game_collection WHERE id=140001`); e != nil {
			t.Fatal(e)
		}
		for _, table := range []string{"gfg_game_collection_item", "gfg_game_collection_home_slot"} {
			if e = tx.QueryRow(ctx, fmt.Sprintf("SELECT count(*) FROM %s WHERE collection_id=140001", table)).Scan(&count); e != nil || count != 0 {
				t.Fatal("collection cascade", table, count, e)
			}
		}
		if e = tx.QueryRow(ctx, `SELECT count(*) FROM gfg_game WHERE id=140101`).Scan(&count); e != nil || count != 1 {
			t.Fatal("collection deletion removed game", count, e)
		}
	})
}

func assertCollectionDiscovery(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	// Changes are confined to this disposable fixture and restored afterwards.
	exec := func(sql string) {
		t.Helper()
		if _, err := pool.Exec(ctx, sql); err != nil {
			t.Fatal(err)
		}
	}
	exec(`UPDATE gfg_game SET name='独特可见作品',name_en='VisibleUnique' WHERE id=140109`)
	exec(`UPDATE gfg_game SET name='隐藏秘密作品',name_en='HiddenUnique' WHERE id=140108`)
	exec(`INSERT INTO gfg_game_release_state(game_id,availability,precision,raw_text,source,source_region,source_locale,normalizer_version,observed_at) VALUES(140108,'upcoming','tba','TBA','steam','US','en','test',now())`)
	defer exec(`DELETE FROM gfg_game_release_state WHERE game_id=140108`)
	exec(`INSERT INTO gfg_game_collection(id,code,name,name_en,info,info_en,status,published_at) VALUES
        (140007,'unknown-only','未知','Unknown','','English description','published','2026-08-01'),
        (140008,'visible-with-adult','可见与隐藏','Visible hidden','','','published','2026-08-01')`)
	defer exec(`DELETE FROM gfg_game_collection WHERE id IN(140007,140008)`)
	exec(`INSERT INTO gfg_game_collection_item(collection_id,game_id) VALUES(140007,140107),(140008,140101),(140008,140108)`)
	svc := v2service.NewCollectionService(v2dao.NewReadModelDAO(pool), nil)
	codes := func(q v2models.CollectionQuery) ([]string, v2models.CollectionIndex) {
		t.Helper()
		result, err := svc.List(ctx, q)
		if err != nil {
			t.Fatal(err)
		}
		out := []string{}
		for _, item := range result.Items {
			out = append(out, item.Code)
		}
		return out, result
	}
	for _, tc := range []struct {
		q    v2models.CollectionQuery
		want []string
	}{
		{v2models.CollectionQuery{Q: "中文谱系"}, []string{"chronicle"}},
		{v2models.CollectionQuery{Q: "CHRONICLE"}, []string{"later-chronicle", "chronicle", "adult-chronicle", "empty-chronicle"}},
		{v2models.CollectionQuery{Q: "Visible hidden"}, []string{"visible-with-adult"}},
		{v2models.CollectionQuery{Q: "later-chronicle"}, []string{"later-chronicle"}},
		{v2models.CollectionQuery{Q: "中文简介"}, []string{"chronicle"}},
		{v2models.CollectionQuery{Q: "English description"}, []string{"unknown-only"}},
		{v2models.CollectionQuery{Q: "独特可见作品"}, []string{"chronicle"}},
		{v2models.CollectionQuery{Q: "VisibleUnique"}, []string{"chronicle"}},
		{v2models.CollectionQuery{Q: "HiddenUnique"}, []string{}},
		{v2models.CollectionQuery{Q: "HiddenUnique", Mode: "nsfw"}, []string{"chronicle", "adult-chronicle", "visible-with-adult"}},
		{v2models.CollectionQuery{Q: "no match"}, []string{}},
		{v2models.CollectionQuery{Q: "%' OR 1=1 --"}, []string{}},
		{v2models.CollectionQuery{Phase: "released"}, []string{"later-chronicle", "chronicle", "visible-with-adult"}},
		{v2models.CollectionQuery{Phase: "upcoming"}, []string{"chronicle"}},
		{v2models.CollectionQuery{Phase: "mixed"}, []string{"chronicle"}},
		{v2models.CollectionQuery{Phase: "upcoming", Mode: "nsfw"}, []string{"chronicle", "adult-chronicle", "visible-with-adult"}},
		{v2models.CollectionQuery{Phase: "mixed", Mode: "nsfw"}, []string{"chronicle", "visible-with-adult"}},
		{v2models.CollectionQuery{Sort: "count_desc"}, []string{"chronicle", "later-chronicle", "visible-with-adult", "unknown-only", "adult-chronicle", "empty-chronicle"}},
		{v2models.CollectionQuery{Sort: "count_asc"}, []string{"adult-chronicle", "empty-chronicle", "later-chronicle", "visible-with-adult", "unknown-only", "chronicle"}},
		{v2models.CollectionQuery{Sort: "count_desc", Mode: "nsfw"}, []string{"chronicle", "visible-with-adult", "later-chronicle", "adult-chronicle", "unknown-only", "empty-chronicle"}},
		{v2models.CollectionQuery{Lang: "en", Sort: "name_asc"}, []string{"adult-chronicle", "chronicle", "empty-chronicle", "later-chronicle", "unknown-only", "visible-with-adult"}},
		{v2models.CollectionQuery{Lang: "en", Sort: "name_desc"}, []string{"visible-with-adult", "unknown-only", "later-chronicle", "empty-chronicle", "chronicle", "adult-chronicle"}},
	} {
		got, result := codes(tc.q)
		if !reflect.DeepEqual(got, tc.want) || result.Total != int64(len(tc.want)) {
			t.Fatalf("criteria %+v: %v total=%d want=%v", tc.q, got, result.Total, tc.want)
		}
	}
	for page := int64(1); page <= 3; page++ {
		got, result := codes(v2models.CollectionQuery{Phase: "released", Page: page, PageSize: 2})
		if result.Total != 3 || result.HasMore != (page == 1) || (page == 2 && !reflect.DeepEqual(got, []string{"visible-with-adult"})) || (page == 3 && len(got) != 0) {
			t.Fatal("filtered pagination", result)
		}
	}
	// Tie-breaking for localized names is numeric ID ascending, independent of publication.
	exec(`UPDATE gfg_game_collection SET name_en='Same' WHERE id IN(140007,140008)`)
	got, _ := codes(v2models.CollectionQuery{Q: "Same", Lang: "en", Sort: "name_desc"})
	if !reflect.DeepEqual(got, []string{"unknown-only", "visible-with-adult"}) {
		t.Fatal("name ties", got)
	}
	exec(`UPDATE gfg_game_collection SET name=CASE WHEN id=140007 THEN 'A' ELSE 'Z' END WHERE id IN(140007,140008)`)
	got, _ = codes(v2models.CollectionQuery{Q: "Same", Lang: "zh", Sort: "name_desc"})
	if !reflect.DeepEqual(got, []string{"visible-with-adult", "unknown-only"}) {
		t.Fatal("sort must use projected locale name", got)
	}
	exec(`UPDATE gfg_game SET name='',name_en='',info='',info_en='' WHERE id=140101`)
	for _, lang := range []string{"zh", "en"} {
		detail, err := svc.Detail(ctx, "later-chronicle", v2models.CollectionQuery{Lang: lang})
		if err != nil || detail.Items[0].Name != "Localized Game" || detail.Items[0].Summary != "Localized summary" || detail.Items[0].HeaderURL != "https://example.test/asset-header.jpg" {
			t.Fatalf("localized projection fallback: %+v %v", detail, err)
		}
	}

}
