package dao_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	v2dao "github.com/gofurry/gofurry-game-backend/apps/game/v2/dao"
	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	v2service "github.com/gofurry/gofurry-game-backend/apps/game/v2/service"
	gamesqlc "github.com/gofurry/gofurry-game-backend/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func assertHybridCollectionReads(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO gfg_game(id,name,name_en,info,info_en,appid,header,developers,publishers,weight,create_time,update_time)
 SELECT id,'混合作品'||id,'HybridGame'||id,'简介','Summary',id,'','[]','[]',0,now(),now() FROM generate_series(147101,147106) id;
 UPDATE gfg_game SET name='自动安全独特',name_en='AutoSafeUnique' WHERE id=147101;
 UPDATE gfg_game SET name='自动成人独特',name_en='AutoAdultUnique' WHERE id=147103;
 INSERT INTO gfg_tag_category(id,code,name,name_en,info,info_en,sort_order,create_time,update_time) VALUES(147001,'hybrid-read','混合读取','Hybrid read','','',0,now(),now());
 INSERT INTO gfg_tag(id,code,name,name_en,info,info_en,category_id,create_time,update_time) VALUES
 (147201,'hybrid-read-a','甲','A','','',147001,now(),now()),(147202,'hybrid-read-b','乙','B','','',147001,now(),now()),(147203,'hybrid-read-adult','单独成人规则','Adult rule','','',147001,now(),now());
 INSERT INTO gfg_game_tag(game_id,tag_id,role,create_time,update_time) VALUES
 (147101,147201,'normal',now(),now()),(147102,147201,'primary',now(),now()),(147103,147201,'secondary',now(),now()),
 (147103,147202,'normal',now(),now()),(147104,147202,'normal',now(),now()),(147103,147203,'normal',now(),now());
 INSERT INTO gfg_game_tag(game_id,tag_id,role,create_time,update_time) SELECT 147103,id,'normal',now(),now() FROM gfg_tag WHERE code='adult';
 INSERT INTO gfg_game_release_state(game_id,availability,precision,raw_text,source,source_region,source_locale,normalizer_version,observed_at) VALUES
 (147101,'available','none','available','steam','US','en','test',now()),(147103,'upcoming','tba','TBA','steam','US','en','test',now());
 INSERT INTO gfg_game_collection(id,code,name,name_en,info,info_en,status,published_at) VALUES
 (147001,'hybrid-public','混合甲','Hybrid A','','','published','2026-10-08'),(147002,'hybrid-hidden','混合乙','Hybrid B','','','published','2026-10-08'),(147003,'hybrid-manual','混合丙','Hybrid C','','','published','2026-10-08');
 INSERT INTO gfg_game_collection_tag(collection_id,tag_id) VALUES(147001,147201),(147001,147202),(147002,147203);
 INSERT INTO gfg_game_collection_item(collection_id,game_id) VALUES(147001,147102),(147001,147105),(147003,147106);
 INSERT INTO gfg_game_collection_exclusion(collection_id,game_id) VALUES(147001,147104),(147001,147106);
 DELETE FROM gfg_game_collection_home_slot;
 INSERT INTO gfg_game_collection_home_slot(slot,collection_id) VALUES(1,147001),(2,147002);`)
	reader := v2dao.NewReadModelDAO(pool)
	svc := v2service.NewCollectionService(reader, nil)
	svc.Now = func() time.Time { return time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC) }
	memberIDs := func() []int64 {
		t.Helper()
		rows, err := gamesqlc.New(pool).BatchCollectionMemberships(ctx, []int64{147001})
		if err != nil {
			t.Fatal(err)
		}
		ids := []int64{}
		for _, r := range rows {
			ids = append(ids, r.GameID)
		}
		return ids
	}
	t.Run("hybrid public OR roles manual exclusion and archived rules", func(t *testing.T) {
		if !reflect.DeepEqual(memberIDs(), []int64{147101, 147102, 147103, 147105}) {
			t.Fatal(memberIDs())
		}
		for _, mode := range []string{"sfw", "nsfw"} {
			result, err := svc.Detail(ctx, "hybrid-public", v2models.CollectionQuery{Mode: mode})
			want := 3
			if mode == "nsfw" {
				want = 4
			}
			if err != nil || len(result.Items) != want || result.Collection.VisibleGameCount != want {
				t.Fatalf("%s %+v %v", mode, result, err)
			}
			for _, item := range result.Items {
				if mode == "sfw" && item.GameID == "147103" {
					t.Fatal("adult leakage")
				}
			}
			index, err := svc.List(ctx, v2models.CollectionQuery{Q: "hybrid-public", Mode: mode})
			if err != nil || index.Total != 1 || index.Items[0].VisibleGameCount != want {
				t.Fatal(index, err)
			}
			preview := index.Items[0].PreviewGames
			wantIDs := []string{result.Items[0].GameID, result.Items[want/2].GameID, result.Items[want-1].GameID}
			got := []string{}
			for _, g := range preview {
				got = append(got, g.GameID)
			}
			if !reflect.DeepEqual(got, wantIDs) {
				t.Fatal("preview not authoritative", got, wantIDs)
			}
		}
		empty, err := svc.Detail(ctx, "hybrid-hidden", v2models.CollectionQuery{})
		if err != nil || len(empty.Items) != 0 {
			t.Fatal(empty, err)
		}
		home, err := svc.Home(ctx, v2models.CollectionQuery{})
		if err != nil || len(home.Slots) != 1 || home.Slots[0].Slot != 1 {
			t.Fatal(home, err)
		}
		exec(`UPDATE gfg_tag SET archived_at=now() WHERE id=147201`)
		if !reflect.DeepEqual(memberIDs(), []int64{147102, 147103, 147105}) {
			t.Fatal("paused tag/manual pin", memberIDs())
		}
		exec(`UPDATE gfg_tag_category SET archived_at=now() WHERE id=147001`)
		if !reflect.DeepEqual(memberIDs(), []int64{147102, 147105}) {
			t.Fatal("paused category", memberIDs())
		}
		exec(`UPDATE gfg_tag SET archived_at=NULL WHERE id=147201;UPDATE gfg_tag_category SET archived_at=NULL WHERE id=147001`)
		if !reflect.DeepEqual(memberIDs(), []int64{147101, 147102, 147103, 147105}) {
			t.Fatal("restored rules", memberIDs())
		}
		exec(`DELETE FROM gfg_game_collection_exclusion WHERE collection_id=147001 AND game_id=147104`)
		if len(memberIDs()) != 5 {
			t.Fatal("exclusion removal", memberIDs())
		}
		exec(`INSERT INTO gfg_game_collection_exclusion(collection_id,game_id) VALUES(147001,147104)`)
		// Defensive read semantics still give exclusion priority over both sources,
		// even though write APIs reject overlapping configuration.
		exec(`INSERT INTO gfg_game_collection_exclusion(collection_id,game_id) VALUES(147001,147102)`)
		if !reflect.DeepEqual(memberIDs(), []int64{147101, 147103, 147105}) {
			t.Fatal("exclusion must override both sources", memberIDs())
		}
		exec(`DELETE FROM gfg_game_collection_exclusion WHERE collection_id=147001 AND game_id=147102`)
	})
	t.Run("hybrid discovery never leaks hidden members into matches phases counts or previews", func(t *testing.T) {
		for _, tc := range []struct {
			q    v2models.CollectionQuery
			want []string
		}{
			{v2models.CollectionQuery{Q: "AutoSafeUnique"}, []string{"hybrid-public"}},
			{v2models.CollectionQuery{Q: "自动安全独特"}, []string{"hybrid-public"}},
			{v2models.CollectionQuery{Q: "AutoAdultUnique"}, []string{}},
			{v2models.CollectionQuery{Q: "AutoAdultUnique", Mode: "nsfw"}, []string{"hybrid-hidden", "hybrid-public"}},
			{v2models.CollectionQuery{Q: "hybrid-", Phase: "released"}, []string{"hybrid-public"}},
			{v2models.CollectionQuery{Q: "hybrid-", Phase: "upcoming"}, []string{}},
			{v2models.CollectionQuery{Q: "hybrid-", Phase: "mixed"}, []string{}},
			{v2models.CollectionQuery{Q: "hybrid-", Phase: "mixed", Mode: "nsfw"}, []string{"hybrid-public"}},
			{v2models.CollectionQuery{Q: "hybrid-", Sort: "count_asc"}, []string{"hybrid-hidden", "hybrid-manual", "hybrid-public"}},
			{v2models.CollectionQuery{Q: "hybrid-", Sort: "count_asc", Mode: "nsfw"}, []string{"hybrid-manual", "hybrid-hidden", "hybrid-public"}},
			{v2models.CollectionQuery{Q: "hybrid-", Sort: "count_desc"}, []string{"hybrid-public", "hybrid-manual", "hybrid-hidden"}},
			{v2models.CollectionQuery{Q: "hybrid-", Sort: "name_asc", Lang: "en"}, []string{"hybrid-public", "hybrid-hidden", "hybrid-manual"}},
		} {
			result, err := svc.List(ctx, tc.q)
			if err != nil {
				t.Fatal(err)
			}
			codes := []string{}
			for _, c := range result.Items {
				codes = append(codes, c.Code)
			}
			if !reflect.DeepEqual(codes, tc.want) || result.Total != int64(len(tc.want)) {
				t.Fatalf("%+v: %v total %d want %v", tc.q, codes, result.Total, tc.want)
			}
		}
		result, err := svc.List(ctx, v2models.CollectionQuery{Q: "AutoAdultUnique", Mode: "nsfw", PageSize: 1})
		if err != nil || result.Total != 2 || !result.HasMore {
			t.Fatal(result, err)
		}
	})
	t.Run("passive rule changes hide Home without deleting its placement", func(t *testing.T) {
		exec(`INSERT INTO gfg_game_collection_home_slot(slot,collection_id) VALUES(3,147003);
  INSERT INTO gfg_game_collection_tag(collection_id,tag_id) VALUES(147003,147201);
  DELETE FROM gfg_game_collection_item WHERE collection_id=147003`)
		exec(`UPDATE gfg_tag SET archived_at=now() WHERE id=147201`)
		home, err := svc.Home(ctx, v2models.CollectionQuery{})
		if err != nil {
			t.Fatal(err)
		}
		for _, slot := range home.Slots {
			if slot.Slot == 3 {
				t.Fatal("empty slot shown")
			}
		}
		var n int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM gfg_game_collection_home_slot WHERE slot=3`).Scan(&n); err != nil || n != 1 {
			t.Fatal("passive read removed placement", err, n)
		}
		exec(`UPDATE gfg_tag SET archived_at=NULL WHERE id=147201`)
		home, err = svc.Home(ctx, v2models.CollectionQuery{})
		if err != nil || len(home.Slots) != 2 || home.Slots[1].Slot != 3 {
			t.Fatal(home, err)
		}
	})
	t.Run("hybrid schema constraints", func(t *testing.T) {
		for _, sql := range []string{
			`INSERT INTO gfg_game_collection_tag(collection_id,tag_id) VALUES(147001,147201)`,
			`INSERT INTO gfg_game_collection_tag(collection_id,tag_id) VALUES(147001,999999)`,
			`INSERT INTO gfg_game_collection_exclusion(collection_id,game_id) VALUES(147001,147104)`,
			`INSERT INTO gfg_game_collection_exclusion(collection_id,game_id) VALUES(999999,147104)`,
			`INSERT INTO gfg_game_collection_exclusion(collection_id,game_id) VALUES(147001,999999)`,
			`DELETE FROM gfg_tag WHERE id=147201`,
		} {
			_, err := pool.Exec(ctx, sql)
			var pg *pgconn.PgError
			if !errors.As(err, &pg) || (pg.Code != "23503" && pg.Code != "23505" && pg.Code != "23001") {
				t.Fatal(sql, err)
			}
		}
		exec(`INSERT INTO gfg_game_collection(id,code,name,name_en) VALUES(147010,'hybrid-cascade','级联','Cascade');
  INSERT INTO gfg_tag(id,code,name,name_en,info,info_en,category_id,create_time,update_time) VALUES(147210,'bound-only','仅规则','Rule only','','',147001,now(),now());
  INSERT INTO gfg_game_collection_tag(collection_id,tag_id) VALUES(147010,147210);
  INSERT INTO gfg_game_collection_exclusion(collection_id,game_id) VALUES(147010,147106)`)
		_, err := pool.Exec(ctx, `DELETE FROM gfg_tag WHERE id=147210`)
		var pg *pgconn.PgError
		if !errors.As(err, &pg) || pg.ConstraintName != "gfg_game_collection_tag_tag_id_fkey" {
			t.Fatal("bound Tag must be explicitly detached before deletion", err)
		}
		exec(`DELETE FROM gfg_game_collection WHERE id=147010`)
		exec(`DELETE FROM gfg_tag WHERE id=147210`)
		var n int
		if err := pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM gfg_game_collection_tag WHERE collection_id=147010)+(SELECT count(*) FROM gfg_game_collection_exclusion WHERE collection_id=147010)`).Scan(&n); err != nil || n != 0 {
			t.Fatal("cascade left rules", n, err)
		}
	})
	t.Run("3 30 100 300 effective members retain fixed bounded batches", func(t *testing.T) {
		exec(`INSERT INTO gfg_game(id,name,name_en,info,info_en,appid,header,developers,publishers,weight,create_time,update_time) SELECT id,'批量','Batch','','',id,'','[]','[]',0,now(),now() FROM generate_series(148001,148300) id;
  INSERT INTO gfg_tag(id,code,name,name_en,info,info_en,category_id,create_time,update_time) VALUES(148001,'hybrid-batch','批量','Batch','','',147001,now(),now());
  INSERT INTO gfg_game_tag(game_id,tag_id,role,create_time,update_time) SELECT id,148001,'normal',now(),now() FROM generate_series(148001,148300) id;
  INSERT INTO gfg_game_collection(id,code,name,name_en,status,published_at) VALUES(148001,'hybrid-batch','批量','Batch','published',now());
  INSERT INTO gfg_game_collection_tag(collection_id,tag_id) VALUES(148001,148001)`)
		tracer := &collectionQueryTrace{}
		cfg := pool.Config()
		cfg.ConnConfig.Tracer = tracer
		traced, err := pgxpool.NewWithConfig(ctx, cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer traced.Close()
		service := v2service.NewCollectionService(v2dao.NewReadModelDAO(traced), nil)
		for _, size := range []int{3, 30, 100, 300} {
			exec(`DELETE FROM gfg_game_collection_exclusion WHERE collection_id=148001;`)
			exec(`INSERT INTO gfg_game_collection_exclusion(collection_id,game_id) SELECT 148001,id FROM generate_series(148001+$1::bigint,148300) id`, size)
			tracer.count = 0
			tracer.decorationCount = 0
			tracer.queries = nil
			start := time.Now()
			detail, err := service.Detail(ctx, "hybrid-batch", v2models.CollectionQuery{})
			if err != nil || len(detail.Items) != size || tracer.count != 8 || tracer.decorationCount != 1 {
				t.Fatalf("size=%d count=%d queries=%d err=%v", size, len(detail.Items), tracer.count, err)
			}
			for _, forbidden := range []string{"gfg_game_prices", "gfg_game_requirements", "gfg_game_news", "queryOnlinePeakCounts"} {
				if strings.Contains(strings.Join(tracer.queries, "\n"), forbidden) {
					t.Fatal(forbidden)
				}
			}
			t.Logf("%d effective members: %d SQL, %s (diagnostic)", size, tracer.count, time.Since(start))
			tracer.count = 0
			_, err = service.List(ctx, v2models.CollectionQuery{Q: "hybrid-batch"})
			if err != nil || tracer.count != 8 {
				t.Fatal("Index batch drift", tracer.count, err)
			}
		}
		rows, err := pool.Query(ctx, `EXPLAIN (ANALYZE,BUFFERS,FORMAT TEXT) SELECT gt.game_id FROM gfg_game_collection_tag rule JOIN gfg_tag t ON t.id=rule.tag_id AND t.archived_at IS NULL JOIN gfg_tag_category c ON c.id=t.category_id AND c.archived_at IS NULL JOIN gfg_game_tag gt ON gt.tag_id=t.id WHERE rule.collection_id=148001`)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			var line string
			if err := rows.Scan(&line); err != nil {
				t.Fatal(err)
			}
			t.Log(line)
		}
		if err := rows.Err(); err != nil {
			t.Fatal(fmt.Sprint(err))
		}
	})
}
