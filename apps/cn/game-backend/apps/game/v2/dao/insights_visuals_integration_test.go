package dao_test

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	v2dao "github.com/gofurry/gofurry-game-backend/apps/game/v2/dao"
	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	v2service "github.com/gofurry/gofurry-game-backend/apps/game/v2/service"
	gamesqlc "github.com/gofurry/gofurry-game-backend/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Runs only inside the existing disposable database, never a configured GFG.
func assertInsightVisuals(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := tx.Exec(ctx, sql, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`UPDATE gfg_game SET showcase_eligible=false`)
	dao := v2dao.NewInsightsDAO(gamesqlc.New(tx))
	read := func(day string) []v2models.InsightVisualCandidateRecord {
		t.Helper()
		rows, err := dao.ListInsightVisualCandidates(ctx, day)
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) > 3 {
			t.Fatal("unbounded result")
		}
		return rows
	}
	const day = "2026-10-09"
	if len(read(day)) != 0 {
		t.Fatal("no permission must mean no candidates")
	}
	seed := func(id, app int64) {
		t.Helper()
		exec(`INSERT INTO gfg_game(id,name,name_en,info,info_en,appid,header,developers,publishers,weight,create_time,update_time,showcase_eligible) VALUES($1,' 标题 ',' Title ','','',$2,'','[]','[]',0,now(),now(),true)`, id, app)
		exec(`INSERT INTO gfg_game_assets(game_id,appid,asset_type,asset_family,source,lang,media_key,url,exists,collected_at,updated_at) VALUES($1,$2,'header','store','store_browse','zh','insight-header',$3,true,now(),now())`, id, app, fmt.Sprintf("https://shared.steamstatic.com/steam/apps/%d/header.jpg", app))
	}
	seed(1082200, 12345)
	data, err := os.ReadFile(filepath.Join(integrationRepositoryRoot(t), "apps/cn/game-backend/apps/game/v2/service/testdata/insight-header-urls.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name, URL string
		Valid     bool
	}
	if err = json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			exec(`UPDATE gfg_game_assets SET url=$1 WHERE game_id=1082200`, tc.URL)
			rows := read(day)
			if (len(rows) == 1) != tc.Valid {
				t.Fatalf("SQL/parser admission mismatch: %+v valid=%v", rows, tc.Valid)
			}
			got, err := v2service.NewInsightsService(dao).GetInsightsOverview(ctx)
			if err != nil || (len(got.FeaturedVisuals) == 1) != tc.Valid {
				t.Fatalf("mapped admission mismatch: %+v %v", got.FeaturedVisuals, err)
			}
		})
	}
	url := "https://shared.steamstatic.com/steam/apps/12345/header.jpg"
	exec(`UPDATE gfg_game_assets SET url=$1 WHERE game_id=1082200`, url)
	for _, tc := range []struct{ Name, Mutation, Restore string }{
		{"permission withdrawal", `UPDATE gfg_game SET showcase_eligible=false WHERE id=1082200`, `UPDATE gfg_game SET showcase_eligible=true WHERE id=1082200`},
		{"adult including archived", `INSERT INTO gfg_game_tag(game_id,tag_id,role,create_time,update_time) SELECT 1082200,id,'normal',now(),now() FROM gfg_tag WHERE code='adult'`, `DELETE FROM gfg_game_tag WHERE game_id=1082200`},
		{"asset app identity", `UPDATE gfg_game_assets SET appid=999 WHERE game_id=1082200`, `UPDATE gfg_game_assets SET appid=12345 WHERE game_id=1082200`},
		{"asset absent", `UPDATE gfg_game_assets SET exists=false WHERE game_id=1082200`, `UPDATE gfg_game_assets SET exists=true WHERE game_id=1082200`},
		{"blank unicode names", `UPDATE gfg_game SET name=U&'\0085\00A0\3000',name_en=E'\t\n' WHERE id=1082200`, `UPDATE gfg_game SET name='标题',name_en='' WHERE id=1082200`},
		{"no header", `UPDATE gfg_game_assets SET asset_type='capsule' WHERE game_id=1082200`, `UPDATE gfg_game_assets SET asset_type='header' WHERE game_id=1082200`},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			exec(`UPDATE gfg_tag SET archived_at=now() WHERE code='adult'`)
			exec(tc.Mutation)
			if len(read(day)) != 0 {
				t.Fatal("revocation not effective on next read")
			}
			exec(tc.Restore)
			if len(read(day)) != 1 {
				t.Fatal("valid row lost")
			}
		})
	}
	// A replacement URL takes effect immediately; no additional positive cache.
	exec(`UPDATE gfg_game_assets SET url=$1 WHERE game_id=1082200`, url+"?t=99")
	if read(day)[0].Asset != url+"?t=99" {
		t.Fatal("stale cached URL")
	}
	exec(`INSERT INTO gfg_game_assets(game_id,appid,asset_type,asset_family,source,lang,media_key,url,exists,collected_at,updated_at) VALUES(1082200,12345,'header_2x','store','store_browse','zh','header-2x',$1,true,now(),now()),(1082200,12345,'header','store','store_browse','en','header-en',$2,true,now(),now())`, strings.Replace(url, "header.jpg", "header_2x.jpg", 1), url+"?t=2")
	if rows := read(day); len(rows) != 1 || rows[0].Asset != url+"?t=99" {
		t.Fatalf("header/zh precedence or duplicate: %+v", rows)
	}
	ids := []int64{1082200}
	for n := 1; n <= 9; n++ {
		id := int64(1082200 + n)
		seed(id, id+10000000)
		ids = append(ids, id)
		if len(read(day)) != min(n+1, 3) {
			t.Fatal("0/1/2/3 contract")
		}
	}
	want := slices.Clone(ids)
	slices.SortFunc(want, func(a, b int64) int {
		return strings.Compare(fmt.Sprintf("%x", md5.Sum(fmt.Appendf(nil, "%s:%d", day, a))), fmt.Sprintf("%x", md5.Sum(fmt.Appendf(nil, "%s:%d", day, b))))
	})
	first := read(day)
	var gotIDs []int64
	for _, r := range first {
		gotIDs = append(gotIDs, r.GameID)
	}
	if !reflect.DeepEqual(gotIDs, want[:3]) || !reflect.DeepEqual(first, read(day)) {
		t.Fatalf("daily deterministic order: %v want %v", gotIDs, want[:3])
	}
	// Invalid records exceed any plausible small oversampling window. SQL must
	// filter before LIMIT, so none can displace the three valid games.
	exec(`INSERT INTO gfg_game(id,name,name_en,info,info_en,appid,header,developers,publishers,weight,create_time,update_time,showcase_eligible) SELECT n,'Invalid','','','',n,'','[]','[]',0,now(),now(),true FROM generate_series(1082300,1082599) n`)
	exec(`INSERT INTO gfg_game_assets(game_id,appid,asset_type,asset_family,source,lang,media_key,url,exists,collected_at,updated_at) SELECT id,appid,'header','store','store_browse','zh','invalid','https://shared.steamstatic.com/header.jpg',true,now(),now() FROM gfg_game WHERE id BETWEEN 1082300 AND 1082599`)
	if !reflect.DeepEqual(first, read(day)) {
		t.Fatal("invalid candidate window starved valid results")
	}
	query, err := os.ReadFile(filepath.Join(integrationRepositoryRoot(t), "apps/cn/game-backend/internal/db/game/queries/insights.sql"))
	if err != nil {
		t.Fatal(err)
	}
	_, tail, found := strings.Cut(strings.ReplaceAll(string(query), "\r\n", "\n"), "-- name: ListGameInsightVisualCandidates :many\n")
	if !found {
		t.Fatal("candidate query not found")
	}
	selected, _, _ := strings.Cut(tail, "-- name: GetGameInsightGame")
	plan, err := tx.Query(ctx, "EXPLAIN (ANALYZE, BUFFERS) "+strings.ReplaceAll(selected, "sqlc.arg(utc_day)", "$1"), day)
	if err != nil {
		t.Fatal(err)
	}
	for plan.Next() {
		var line string
		if err := plan.Scan(&line); err != nil {
			t.Fatal(err)
		}
		t.Log(line)
	}
	if err := plan.Err(); err != nil {
		t.Fatal(err)
	}
	plan.Close()
	short, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := dao.ListInsightVisualCandidates(short, day); err == nil {
		t.Fatal("cancellation was ignored")
	}
	// Rollback releases the transaction's connection, even after cancellation.
	_ = tx.Rollback(context.Background())
	ping, cancelPing := context.WithTimeout(ctx, time.Second)
	defer cancelPing()
	if err := pool.Ping(ping); err != nil {
		t.Fatal("pool unavailable after optional read cancellation", err)
	}
	// An in-flight blocked query must time out and return its pool connection.
	baselineConnections := pool.Stat().AcquiredConns()
	locker, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer locker.Rollback(context.Background())
	if _, err := locker.Exec(ctx, `LOCK TABLE gfg_game_assets IN ACCESS EXCLUSIVE MODE`); err != nil {
		t.Fatal(err)
	}
	blocked, cancelBlocked := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancelBlocked()
	_, err = v2dao.NewInsightsDAO(gamesqlc.New(pool)).ListInsightVisualCandidates(blocked, day)
	if err == nil || blocked.Err() != context.DeadlineExceeded {
		t.Fatalf("blocked query did not honor deadline: %v", err)
	}
	// pgx destroys a canceled/busy connection asynchronously. Assert bounded
	// reclamation instead of assuming the pool statistic changes synchronously.
	reclaimed, cancelReclaimed := context.WithTimeout(ctx, time.Second)
	defer cancelReclaimed()
	tick := time.NewTicker(5 * time.Millisecond)
	defer tick.Stop()
	for pool.Stat().AcquiredConns() != baselineConnections+1 {
		select {
		case <-reclaimed.Done():
			t.Fatalf("query connection retained: acquired=%d baseline=%d", pool.Stat().AcquiredConns(), baselineConnections)
		case <-tick.C:
		}
	}
	_ = locker.Rollback(context.Background())
}
