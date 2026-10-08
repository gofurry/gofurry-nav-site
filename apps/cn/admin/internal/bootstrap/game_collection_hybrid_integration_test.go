package bootstrap_test

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/gofurry/gofurry-admin/internal/app/gameadmin/models"
	gamesqlc "github.com/gofurry/gofurry-admin/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Uses the existing disposable three-database gate and its real HTTP handlers.
func assertHybridCollectionOperations(t *testing.T, ctx context.Context, pools map[string]*pgxpool.Pool, call func(string, string, any, int) json.RawMessage, race func([][3]string) []int) {
	t.Helper()
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := pools["game"].Exec(ctx, sql, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO gfg_game(id,name,name_en,info,info_en,create_time,update_time,developers,publishers,appid,header,weight)
 SELECT id,'混合游戏'||id,'Hybrid '||id,'简介','Summary',now(),now(),'[]','[]',id,'',0 FROM generate_series(144101,144106) id;
 INSERT INTO gfg_tag_category(id,code,name,name_en,info,info_en,sort_order,create_time,update_time) VALUES(144001,'hybrid-category','混合','Hybrid','','',0,now(),now());
 INSERT INTO gfg_tag(id,code,name,name_en,info,info_en,category_id,create_time,update_time,archived_at) VALUES
 (144001,'hybrid-a','规则甲','Rule A','','',144001,now(),now(),NULL),(144002,'hybrid-b','规则乙','Rule B','','',144001,now(),now(),NULL),(144003,'hybrid-paused','暂停','Paused','','',144001,now(),now(),now());
 INSERT INTO gfg_game_tag(game_id,tag_id,role,create_time,update_time) VALUES
 (144101,144001,'normal',now(),now()),(144102,144001,'primary',now(),now()),(144103,144001,'secondary',now(),now()),
 (144103,144002,'normal',now(),now()),(144104,144002,'normal',now(),now()),(144103,140,'normal',now(),now()),(144104,140,'normal',now(),now());`)
	content := models.CollectionContent{Name: "混合分区", NameEn: "Hybrid", Info: "简介", InfoEn: "Summary"}
	decode := func(raw json.RawMessage) models.GameCollection {
		t.Helper()
		var c models.GameCollection
		if err := json.Unmarshal(raw, &c); err != nil {
			t.Fatal(err)
		}
		return c
	}
	create := func(code string) models.GameCollection {
		return decode(call("POST", "/collections", models.CreateCollection{Code: code, CollectionContent: content}, 200))
	}
	path := func(id int64) string { return fmt.Sprintf("/collections/%d", id) }
	get := func(id int64) models.GameCollection { return decode(call("GET", path(id), nil, 200)) }
	composition := func(method string, id int64, in any, status int) models.CollectionComposition {
		t.Helper()
		raw := call(method, path(id)+"/composition", in, status)
		var c models.CollectionComposition
		if status == 200 {
			if err := json.Unmarshal(raw, &c); err != nil {
				t.Fatal(err)
			}
		}
		return c
	}
	config := func(version int64, tags, manual, excluded []int64) models.ReplaceCollectionComposition {
		return models.ReplaceCollectionComposition{Version: version, TagIDs: tags, ManualGameIDs: manual, ExcludedGameIDs: excluded}
	}
	auditCount := func() int64 {
		t.Helper()
		var n int64
		if err := pools["admin"].QueryRow(ctx, `SELECT count(*) FROM gfa_admin_audit_log WHERE resource='gfg_game_collection'`).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	home := func() models.CollectionHome {
		t.Helper()
		var h models.CollectionHome
		if err := json.Unmarshal(call("GET", "/collections/home-curation", nil, 200), &h); err != nil {
			t.Fatal(err)
		}
		return h
	}
	placement := func(h models.CollectionHome, chosen map[int]int64) models.ReplaceCollectionHome {
		in := models.ReplaceCollectionHome{Revision: h.Revision, Slots: make([]models.CollectionPlacement, 5)}
		for i := range in.Slots {
			in.Slots[i].Slot = int16(i + 1)
			if id, ok := chosen[i+1]; ok {
				in.Slots[i].CollectionID = &id
			}
		}
		return in
	}
	t.Run("hybrid composition OR sources compatibility paused rules and validation", func(t *testing.T) {
		c := create("hybrid-config")
		initial := composition("GET", c.ID, nil, 200)
		if initial.RuleTags == nil || initial.EffectiveMembers == nil || initial.ManualMembers == nil || initial.ExcludedMembers == nil {
			t.Fatal("empty arrays", initial)
		}
		in := config(c.Version, []int64{144002, 144001, 144002}, []int64{144105, 144102, 144102}, []int64{144106, 144104})
		result := composition("PUT", c.ID, in, 200)
		if result.Counts != (models.CollectionCompositionCounts{AutoMatched: 4, ManualPinned: 2, Excluded: 2, Effective: 4, SFWVisible: 3}) {
			t.Fatal(result.Counts)
		}
		sources := map[int64]string{}
		for _, m := range result.EffectiveMembers {
			sources[m.GameID] = m.Source
		}
		if !reflect.DeepEqual(sources, map[int64]string{144101: "automatic", 144102: "both", 144103: "automatic", 144105: "manual"}) {
			t.Fatal(sources)
		}
		current := get(c.ID)
		if current.MemberCount != 4 || current.SFWMemberCount != 3 {
			t.Fatal(current)
		}
		var legacy models.CollectionMembers
		if err := json.Unmarshal(call("GET", path(c.ID)+"/members", nil, 200), &legacy); err != nil {
			t.Fatal(err)
		}
		if len(legacy.Members) != 2 || legacy.Members[0].GameID != 144102 || legacy.Members[1].GameID != 144105 {
			t.Fatal("legacy materialized automatic IDs", legacy)
		}
		in.Version = result.Version
		n := auditCount()
		same := composition("PUT", c.ID, in, 200)
		if same.Version != result.Version || auditCount() != n {
			t.Fatal("no-op bumped")
		}
		in.Version--
		composition("PUT", c.ID, in, 409)
		in.Version = result.Version
		for _, bad := range []any{
			map[string]any{"version": in.Version, "tag_ids": []int64{}, "manual_game_ids": []int64{}},
			config(in.Version, []int64{144003}, []int64{}, []int64{}), config(in.Version, []int64{999999}, []int64{}, []int64{}),
			config(in.Version, []int64{}, []int64{999999}, []int64{}), config(in.Version, []int64{}, []int64{}, []int64{999999}),
			config(in.Version, []int64{}, []int64{144101}, []int64{144101}),
		} {
			composition("PUT", c.ID, bad, 400)
		}
		call("PUT", path(c.ID)+"/members", models.ReplaceCollectionMembers{Version: in.Version, GameIDs: []int64{144104}}, 400)
		if get(c.ID).Version != in.Version || auditCount() != n {
			t.Fatal("invalid write mutated state")
		}
		exec(`UPDATE gfg_tag SET archived_at=now() WHERE id=144001`)
		paused := composition("GET", c.ID, nil, 200)
		if paused.RuleTags[0].Active || paused.Counts.AutoMatched != 2 || paused.Counts.Effective != 3 || paused.Version != in.Version {
			t.Fatal(paused)
		}
		// Retain the now-inactive binding while editing another explicit override.
		in.ExcludedGameIDs = []int64{144104}
		result = composition("PUT", c.ID, in, 200)
		in.Version = result.Version
		exec(`UPDATE gfg_tag_category SET archived_at=now() WHERE id=144001`)
		paused = composition("GET", c.ID, nil, 200)
		if paused.Counts.AutoMatched != 0 || paused.Counts.Effective != 2 || paused.RuleTags[1].Active {
			t.Fatal(paused)
		}
		in.TagIDs = []int64{144002}
		result = composition("PUT", c.ID, in, 200)
		in.Version = result.Version
		in.TagIDs = []int64{144001, 144002}
		composition("PUT", c.ID, in, 400)
		exec(`UPDATE gfg_tag SET archived_at=NULL WHERE id=144001; UPDATE gfg_tag_category SET archived_at=NULL WHERE id=144001`)
		result = composition("PUT", c.ID, in, 200)
		// Exclusions survive non-matching periods and can be explicitly removed.
		result = composition("PUT", c.ID, config(result.Version, []int64{144001, 144002}, []int64{144102, 144105}, []int64{}), 200)
		if result.Counts.Effective != 5 {
			t.Fatal(result.Counts)
		}
		exec(`DELETE FROM gfg_game_tag WHERE game_id=144102 AND tag_id=144001`)
		pinned := composition("GET", c.ID, nil, 200)
		if pinned.Counts.Effective != 5 {
			t.Fatal("manual pin lost after tag removal", pinned.Counts)
		}
		exec(`INSERT INTO gfg_game_tag(game_id,tag_id,role,create_time,update_time) VALUES(144102,144001,'primary',now(),now())`)
	})
	t.Run("passive lifecycle and home placement versus explicit writes", func(t *testing.T) {
		c := create("hybrid-lifecycle")
		result := composition("PUT", c.ID, config(c.Version, []int64{144001}, []int64{}, []int64{}), 200)
		c = decode(call("POST", path(c.ID)+"/publish", models.CollectionVersion{Version: result.Version}, 200))
		call("PUT", "/collections/home-curation", placement(home(), map[int]int64{1: c.ID}), 200)
		assertEligible := func(want int64) {
			t.Helper()
			var list struct {
				Total int64                   `json:"total"`
				List  []models.GameCollection `json:"list"`
			}
			if err := json.Unmarshal(call("GET", "/collections?keyword=hybrid-lifecycle&home_eligible=true", nil, 200), &list); err != nil {
				t.Fatal(err)
			}
			if list.Total != want || int64(len(list.List)) != want {
				t.Fatal("Admin eligibility disagrees with Effective SFW", list)
			}
		}
		assertEligible(1)
		beforeHome := home()
		before := get(c.ID)
		n := auditCount()
		exec(`UPDATE gfg_tag SET archived_at=now() WHERE id=144001`)
		after := get(c.ID)
		assertEligible(0)
		if after.MemberCount != 0 || after.Status != "published" || after.Version != before.Version || after.HomeSlot == nil || !after.PublishedAt.Equal(*before.PublishedAt) || home().Revision != beforeHome.Revision || auditCount() != n {
			t.Fatal("passive mutation changed configuration", after)
		}
		// An unchanged invalid mapping may survive while another slot is edited.
		other := create("hybrid-home-other")
		other = decode(call("PUT", path(other.ID)+"/members", models.ReplaceCollectionMembers{Version: other.Version, GameIDs: []int64{1, 4}}, 200))
		other = decode(call("POST", path(other.ID)+"/publish", models.CollectionVersion{Version: other.Version}, 200))
		call("PUT", "/collections/home-curation", placement(home(), map[int]int64{1: c.ID, 2: other.ID}), 200)
		call("PUT", "/collections/home-curation", placement(home(), map[int]int64{3: c.ID, 2: other.ID}), 400)
		// Pure content may be edited even if automatic members disappeared.
		changed := content
		changed.Name = "被动变化后仍可改文案"
		after = decode(call("PUT", path(c.ID), models.UpdateCollection{Version: after.Version, CollectionContent: changed}, 200))
		composition("PUT", c.ID, config(after.Version, []int64{}, []int64{}, []int64{}), 400)
		call("PUT", path(c.ID)+"/members", models.ReplaceCollectionMembers{Version: after.Version, GameIDs: []int64{1}}, 400)
		exec(`UPDATE gfg_tag SET archived_at=NULL WHERE id=144001`)
		restored := get(c.ID)
		assertEligible(1)
		if restored.MemberCount != 3 || restored.Version != after.Version || restored.HomeSlot == nil {
			t.Fatal(restored)
		}
		// Legacy full-set writes validate Effective, not the number of manual pins.
		c = decode(call("PUT", path(c.ID)+"/members", models.ReplaceCollectionMembers{Version: restored.Version, GameIDs: []int64{144105}}, 200))
		if c.MemberCount != 4 {
			t.Fatal(c)
		}
		result = composition("PUT", c.ID, config(c.Version, []int64{}, []int64{144103, 144104}, []int64{}), 200)
		if result.HomeSlot != nil || result.Counts.Effective != 2 || result.Counts.SFWVisible != 0 {
			t.Fatal(result)
		}
		assertEligible(0)
		var beforeSlot, afterSlot *int16
		if err := pools["admin"].QueryRow(ctx, `SELECT (before_data::jsonb->>'home_slot')::smallint, (after_data::jsonb->>'home_slot')::smallint FROM gfa_admin_audit_log WHERE action='composition_update' AND target_id=$1 ORDER BY id DESC LIMIT 1`, fmt.Sprint(c.ID)).Scan(&beforeSlot, &afterSlot); err != nil || beforeSlot == nil || *beforeSlot != 1 || afterSlot != nil {
			t.Fatal("audit must record automatic Home removal", beforeSlot, afterSlot, err)
		}
		c = decode(call("POST", path(c.ID)+"/archive", models.CollectionVersion{Version: result.Version}, 200))
		composition("GET", c.ID, nil, 200)
		composition("PUT", c.ID, config(c.Version, []int64{}, []int64{}, []int64{}), 409)
		c = decode(call("POST", path(c.ID)+"/restore", models.CollectionVersion{Version: c.Version}, 200))
		if c.Status != "draft" || c.HomeSlot != nil {
			t.Fatal(c)
		}
	})
	t.Run("composition audit rollback and concurrent versions", func(t *testing.T) {
		c := create("hybrid-race")
		// Block GFA Audit without changing GFG data; no distributed-transaction claim.
		if _, err := pools["admin"].Exec(ctx, `ALTER TABLE gfa_admin_audit_log RENAME TO hybrid_unavailable_audit`); err != nil {
			t.Fatal(err)
		}
		composition("PUT", c.ID, config(c.Version, []int64{144001}, []int64{144105}, []int64{144103}), 500)
		if _, err := pools["admin"].Exec(ctx, `ALTER TABLE hybrid_unavailable_audit RENAME TO gfa_admin_audit_log`); err != nil {
			t.Fatal(err)
		}
		if latest := composition("GET", c.ID, nil, 200); latest.Version != c.Version || len(latest.RuleTags) != 0 || latest.Counts.Effective != 0 {
			t.Fatal("failed audit did not roll back", latest)
		}
		// Two real writes contend on the same Collection domain lock and row version.
		body := func(in models.ReplaceCollectionComposition) string { b, _ := json.Marshal(in); return string(b) }
		statuses := race([][3]string{{"PUT", path(c.ID) + "/composition", body(config(c.Version, []int64{144001}, []int64{}, []int64{}))}, {"PUT", path(c.ID) + "/composition", body(config(c.Version, []int64{144002}, []int64{}, []int64{}))}})
		if !reflect.DeepEqual(statuses, []int{200, 409}) {
			t.Fatal(statuses)
		}
		if get(c.ID).Version != c.Version+1 {
			t.Fatal("two versions committed")
		}
		var before, after string
		if err := pools["admin"].QueryRow(ctx, `SELECT before_data,after_data FROM gfa_admin_audit_log WHERE resource='gfg_game_collection' AND action='composition_update' AND target_id=$1 ORDER BY id DESC LIMIT 1`, fmt.Sprint(c.ID)).Scan(&before, &after); err != nil {
			t.Fatal(err)
		}
		var a map[string]any
		if err := json.Unmarshal([]byte(after), &a); err != nil {
			t.Fatal(err)
		}
		if a["effective_members"] != nil || a["configuration_sha256"] == nil || a["counts"] == nil || a["tag_ids"] == nil {
			t.Fatal(after)
		}
	})
	t.Run("Tag changes proceed during Collection transaction without changing its version", func(t *testing.T) {
		c := create("hybrid-independent-domains")
		result := composition("PUT", c.ID, config(c.Version, []int64{144001}, []int64{}, []int64{}), 200)
		tx, err := pools["game"].Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(ctx)
		q := gamesqlc.New(tx)
		if err = q.LockGameCollectionDomain(ctx); err != nil {
			t.Fatal(err)
		}
		if _, err = q.LockCollectionForUpdate(ctx, c.ID); err != nil {
			t.Fatal(err)
		}
		// Holding the Collection domain/row must not prevent an independent Tag update.
		tagCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		tagTx, err := pools["game"].Begin(tagCtx)
		if err != nil {
			t.Fatal(err)
		}
		defer tagTx.Rollback(ctx)
		if err = gamesqlc.New(tagTx).LockTagDomain(tagCtx); err != nil {
			t.Fatal(err)
		}
		if _, err = tagTx.Exec(tagCtx, `UPDATE gfg_tag SET archived_at=now() WHERE id=144001`); err != nil {
			t.Fatal(err)
		}
		if err = tagTx.Commit(tagCtx); err != nil {
			t.Fatal(err)
		}
		defer exec(`UPDATE gfg_tag SET archived_at=NULL WHERE id=144001`)
		rows, err := q.GetCuratedCollectionSnapshots(ctx, []int64{c.ID})
		if err != nil || len(rows) != 1 || rows[0].MemberCount != 0 || rows[0].GfgGameCollection.Version != result.Version {
			t.Fatal("derived facts must not be frozen by Collection version", rows, err)
		}
	})

}
