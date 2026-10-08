package bootstrap_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/auth/authorization"
	authmw "github.com/gofurry/gofurry-admin/internal/app/auth/middleware"
	gameadmin "github.com/gofurry/gofurry-admin/internal/app/gameadmin/controller"
	"github.com/gofurry/gofurry-admin/internal/app/gameadmin/models"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestAdminGameCollectionThreeDatabase(t *testing.T) {
	base := adminIntegrationDSN(t, requireAdminIntegrationConfig(t))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	root := adminIntegrationSQLDB(t, base, "postgres")
	defer root.Close()
	pools := map[string]*pgxpool.Pool{}
	for _, domain := range []string{"admin", "game", "nav"} {
		name := integrationDatabaseName("game_collection_" + domain)
		createIntegrationDatabase(t, ctx, root, name)
		defer dropIntegrationDatabase(t, root, name)
		dsn := integrationDatabaseDSN(base, name)
		applyIntegrationBaseline(t, domain, dsn)
		pools[domain] = integrationPool(t, ctx, dsn)
		defer pools[domain].Close()
	}
	exec := func(domain, sql string, args ...any) {
		t.Helper()
		if _, err := pools[domain].Exec(ctx, sql, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec("admin", `INSERT INTO gfa_admin_account(id,username,display_name,password_hash,role,status,session_version,created_at,updated_at) VALUES(140,'collection-tester','Collection Tester','test','operator','active',1,now(),now())`)
	exec("game", `INSERT INTO gfg_game(id,name,name_en,info,info_en,create_time,update_time,developers,publishers,appid,header,weight) SELECT id,'游戏'||id,'Game '||id,'简介','Description',now(),now(),'[]','[]',id,'',100 FROM generate_series(1,4) id`)
	exec("game", `INSERT INTO gfg_tag_category(id,code,name,name_en,info,info_en,sort_order,create_time,update_time) VALUES(140,'test-category','测试','Test','','',0,now(),now())`)
	exec("game", `INSERT INTO gfg_tag(id,code,name,name_en,info,info_en,category_id,create_time,update_time,archived_at) VALUES(140,'adult','成人','Adult','','',140,now(),now(),now()),(1014,'innocent','普通','Ordinary','','',140,now(),now(),NULL)`)
	exec("game", `INSERT INTO gfg_game_tag(game_id,tag_id,role,create_time,update_time) VALUES(1,1014,'normal',now(),now()),(2,140,'normal',now(),now()),(3,140,'normal',now(),now())`)
	principal := &authorization.Principal{AccountID: 140, Username: "collection-tester", DisplayName: "Collection Tester", Role: authorization.RoleOperator, Capabilities: authorization.CapabilitiesFor(authorization.RoleOperator)}
	api := gameadmin.New(pools["game"], audit.New(pools["admin"]))
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error { c.Locals(authorization.PrincipalContextKey, principal); return c.Next() })
	read, write := authmw.Require(authorization.ContentRead), authmw.Require(authorization.ContentWrite)
	app.Get("/collections", read, api.ListGameCollections)
	app.Post("/collections", write, api.CreateGameCollection)
	app.Get("/collections/home-curation", read, api.GetGameCollectionHome)
	app.Put("/collections/home-curation", write, api.ReplaceGameCollectionHome)
	app.Get("/collections/:id", read, api.GetGameCollection)
	app.Put("/collections/:id", write, api.UpdateGameCollection)
	app.Get("/collections/:id/composition", read, api.GetGameCollectionComposition)
	app.Put("/collections/:id/composition", write, api.ReplaceGameCollectionComposition)
	app.Get("/collections/:id/members", read, api.GetGameCollectionMembers)
	app.Put("/collections/:id/members", write, api.ReplaceGameCollectionMembers)
	for _, action := range []string{"publish", "unpublish", "archive", "restore"} {
		app.Post("/collections/:id/"+action, write, api.TransitionGameCollection(action))
	}
	call := func(method, path string, body any, status int) json.RawMessage {
		t.Helper()
		payload := ""
		if body != nil {
			b, err := json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
			payload = string(b)
		}
		res := requestJSON(t, app, method, path, payload, nil, status)
		defer res.Body.Close()
		var envelope integrationEnvelope
		if err := json.NewDecoder(res.Body).Decode(&envelope); err != nil {
			t.Fatal(err)
		}
		return envelope.Data
	}
	decode := func(raw json.RawMessage) models.GameCollection {
		t.Helper()
		var c models.GameCollection
		if e := json.Unmarshal(raw, &c); e != nil {
			t.Fatal(e)
		}
		return c
	}
	path := func(c models.GameCollection) string { return fmt.Sprintf("/collections/%d", c.ID) }
	content := models.CollectionContent{Name: "游戏分区", NameEn: "Game Collection", Info: "简介", InfoEn: "Description"}
	create := func(code string) models.GameCollection {
		return decode(call("POST", "/collections", models.CreateCollection{Code: code, CollectionContent: content}, 200))
	}
	get := func(c models.GameCollection) models.GameCollection { return decode(call("GET", path(c), nil, 200)) }
	members := func(c models.GameCollection, ids []int64) models.GameCollection {
		return decode(call("PUT", path(c)+"/members", models.ReplaceCollectionMembers{Version: c.Version, GameIDs: ids}, 200))
	}
	transition := func(c models.GameCollection, action string) models.GameCollection {
		return decode(call("POST", path(c)+"/"+action, models.CollectionVersion{Version: c.Version}, 200))
	}
	home := func() models.CollectionHome {
		t.Helper()
		var h models.CollectionHome
		if err := json.Unmarshal(call("GET", "/collections/home-curation", nil, 200), &h); err != nil {
			t.Fatal(err)
		}
		return h
	}
	placement := func(h models.CollectionHome, selected map[int]int64) models.ReplaceCollectionHome {
		in := models.ReplaceCollectionHome{Revision: h.Revision, Slots: make([]models.CollectionPlacement, 5)}
		for i := range in.Slots {
			in.Slots[i].Slot = int16(i + 1)
			if id, ok := selected[i+1]; ok {
				in.Slots[i].CollectionID = &id
			}
		}
		return in
	}
	auditCount := func() int {
		t.Helper()
		var n int
		if e := pools["admin"].QueryRow(ctx, `SELECT count(*) FROM gfa_admin_audit_log WHERE resource LIKE 'gfg_game_collection%'`).Scan(&n); e != nil {
			t.Fatal(e)
		}
		return n
	}

	t.Run("validation immutable code and draft", func(t *testing.T) {
		call("POST", "/collections", map[string]any{"code": "Bad", "name": "名称", "name_en": "Name"}, 400)
		call("POST", "/collections", map[string]any{"code": "valid", "name": " ", "name_en": "Name"}, 400)
		c := decode(call("POST", "/collections", map[string]any{"code": "draft", "name": " 名称 ", "name_en": " Name "}, 200))
		if c.Status != "draft" || c.Version != 1 || c.Info != "" || c.Name != "名称" {
			t.Fatal(c)
		}
		call("POST", "/collections", models.CreateCollection{Code: "draft", CollectionContent: content}, 409)
		call("PUT", path(c), map[string]any{"version": 1, "code": "changed", "name": "名称", "name_en": "Name"}, 400)
		call("PUT", path(c), models.UpdateCollection{Version: 0, CollectionContent: content}, 400)
		call("GET", "/collections/999999", nil, 404)
		call("PUT", "/collections/999999", models.UpdateCollection{Version: 1, CollectionContent: content}, 404)
		c = members(c, []int64{})
		if c.Version != 1 {
			t.Fatal("empty no-op", c)
		}
		c = members(c, []int64{1})
		if c.MemberCount != 1 {
			t.Fatal(c)
		}
		call("POST", path(c)+"/publish", models.CollectionVersion{Version: c.Version}, 400)
		c = members(c, []int64{1, 2})
		call("POST", path(c)+"/publish", models.CollectionVersion{Version: c.Version}, 400)
	})

	c := create("main")
	c = members(c, []int64{2, 1, 2})
	t.Run("canonical membership adult code and rollback", func(t *testing.T) {
		if c.MemberCount != 2 || c.SFWMemberCount != 1 {
			t.Fatal(c)
		}
		var m models.CollectionMembers
		json.Unmarshal(call("GET", path(c)+"/members", nil, 200), &m)
		if len(m.Members) != 2 || m.Members[0].GameID != 1 || m.Members[0].Adult || !m.Members[1].Adult || m.Version != c.Version {
			t.Fatal(m)
		}
		n := auditCount()
		same := members(c, []int64{2, 1, 1})
		if same.Version != c.Version || auditCount() != n {
			t.Fatal("same set must be no-op")
		}
		call("PUT", path(c)+"/members", models.ReplaceCollectionMembers{Version: c.Version, GameIDs: []int64{999, 1}}, 400)
		call("PUT", path(c)+"/members", models.ReplaceCollectionMembers{Version: c.Version, GameIDs: []int64{-1}}, 400)
		call("PUT", path(c)+"/members", map[string]any{"version": c.Version}, 400)
		if latest := get(c); latest.Version != c.Version || latest.MemberCount != 2 {
			t.Fatal("failed write was not rolled back", latest)
		}
	})
	c = transition(c, "publish")
	publishedAt := *c.PublishedAt
	t.Run("published readiness no-op stale content", func(t *testing.T) {
		if c.Version != 3 || c.ArchivedAt != nil || c.Status != "published" {
			t.Fatal(c)
		}
		call("PUT", path(c)+"/members", models.ReplaceCollectionMembers{Version: c.Version, GameIDs: []int64{1}}, 400)
		bad := content
		bad.Info = ""
		call("PUT", path(c), models.UpdateCollection{Version: c.Version, CollectionContent: bad}, 400)
		call("PUT", path(c), models.UpdateCollection{Version: c.Version - 1, CollectionContent: content}, 409)
		n := auditCount()
		same := decode(call("PUT", path(c), models.UpdateCollection{Version: c.Version, CollectionContent: content}, 200))
		if same.Version != c.Version || auditCount() != n {
			t.Fatal("content no-op")
		}
	})
	t.Run("fixed slots eligibility revision replacement", func(t *testing.T) {
		h := home()
		if len(h.Slots) != 5 {
			t.Fatal(h)
		}
		for i, s := range h.Slots {
			if int(s.Slot) != i+1 || s.Collection != nil {
				t.Fatal(h)
			}
		}
		call("PUT", "/collections/home-curation", placement(h, map[int]int64{1: c.ID}), 200)
		c = get(c)
		if c.HomeSlot == nil || *c.HomeSlot != 1 || c.Version != 3 {
			t.Fatal(c)
		}
		call("PUT", "/collections/home-curation", placement(h, map[int]int64{}), 409)
		h = home()
		call("PUT", "/collections/home-curation", map[string]any{"revision": h.Revision, "slots": []map[string]int{{"slot": 1}, {"slot": 2}, {"slot": 3}, {"slot": 4}, {"slot": 5}}}, 400)
		changed := content
		changed.Name = "更新名称"
		c = decode(call("PUT", path(c), models.UpdateCollection{Version: c.Version, CollectionContent: changed}, 200))
		if home().Revision != h.Revision {
			t.Fatal("metadata changed placement revision")
		}
		call("PUT", "/collections/home-curation", placement(h, map[int]int64{5: c.ID}), 200)
		c = get(c)
		if *c.HomeSlot != 5 || c.Version != 4 {
			t.Fatal(c)
		}
		h = home()
		call("PUT", "/collections/home-curation", placement(h, map[int]int64{1: c.ID, 2: c.ID}), 400)
		in := placement(h, nil)
		in.Slots = in.Slots[:4]
		call("PUT", "/collections/home-curation", in, 400)
		call("PUT", "/collections/home-curation", placement(h, map[int]int64{1: 99999}), 400)
		d := create("ineligible")
		call("PUT", "/collections/home-curation", placement(h, map[int]int64{1: d.ID}), 400)
		if home().Revision != h.Revision {
			t.Fatal("invalid curation changed slots")
		}
	})
	t.Run("last SFW removal auto clears with bounded audit", func(t *testing.T) {
		beforeVersion := c.Version
		c = members(c, []int64{2, 3})
		if c.Status != "published" || c.SFWMemberCount != 0 || c.HomeSlot != nil || c.Version != beforeVersion+1 {
			t.Fatal(c)
		}
		call("PUT", "/collections/home-curation", placement(home(), map[int]int64{1: c.ID}), 400)
		var before, after string
		if e := pools["admin"].QueryRow(ctx, `SELECT before_data,after_data FROM gfa_admin_audit_log WHERE resource='gfg_game_collection' AND action='members_update' AND target_id=$1 ORDER BY id DESC LIMIT 1`, fmt.Sprint(c.ID)).Scan(&before, &after); e != nil {
			t.Fatal(e)
		}
		if !strings.Contains(before, `"home_slot":5`) || !strings.Contains(after, `"home_slot":null`) || !strings.Contains(after, `"game_ids":[2,3]`) {
			t.Fatal(before, after)
		}
	})
	t.Run("lifecycle preserves historical publication and members", func(t *testing.T) {
		c = transition(c, "unpublish")
		if c.Status != "draft" || !c.PublishedAt.Equal(publishedAt) {
			t.Fatal(c)
		}
		call("POST", path(c)+"/unpublish", models.CollectionVersion{Version: c.Version}, 409)
		c = transition(c, "publish")
		c = members(c, []int64{1, 2})
		call("PUT", "/collections/home-curation", placement(home(), map[int]int64{1: c.ID}), 200)
		c = get(c)
		c = transition(c, "archive")
		if c.ArchivedAt == nil || c.HomeSlot != nil || c.PublishedAt == nil {
			t.Fatal(c)
		}
		lastPublished := *c.PublishedAt
		call("PUT", path(c), models.UpdateCollection{Version: c.Version, CollectionContent: content}, 409)
		call("PUT", path(c)+"/members", models.ReplaceCollectionMembers{Version: c.Version, GameIDs: []int64{1, 4}}, 409)
		call("POST", path(c)+"/publish", models.CollectionVersion{Version: c.Version}, 409)
		c = transition(c, "restore")
		if c.Status != "draft" || c.ArchivedAt != nil || c.HomeSlot != nil || c.MemberCount != 2 || !c.PublishedAt.Equal(lastPublished) {
			t.Fatal(c)
		}
		call("POST", path(c)+"/restore", models.CollectionVersion{Version: c.Version}, 409)
		d := create("archive-draft")
		d = transition(d, "archive")
		if d.PublishedAt != nil {
			t.Fatal(d)
		}
		d = transition(d, "restore")
		if d.PublishedAt != nil {
			t.Fatal(d)
		}
	})
	t.Run("list filter order pagination and reader", func(t *testing.T) {
		c = transition(c, "publish")
		var list struct {
			List  []models.GameCollection `json:"list"`
			Total int64                   `json:"total"`
		}
		json.Unmarshal(call("GET", "/collections?status=published&home_eligible=true&keyword=main&page_num=1&page_size=1", nil, 200), &list)
		if list.Total != 1 || len(list.List) != 1 || list.List[0].ID != c.ID {
			t.Fatal(list)
		}
		json.Unmarshal(call("GET", "/collections?page_num=1&page_size=1", nil, 200), &list)
		if len(list.List) != 1 || list.List[0].ID != c.ID {
			t.Fatal("updated order", list)
		}
		call("GET", "/collections?status=invalid", nil, 400)
		call("GET", "/collections?home_eligible=invalid", nil, 400)
		principal.Capabilities = []authorization.Capability{authorization.ContentRead}
		call("GET", path(c), nil, 200)
		call("GET", path(c)+"/members", nil, 200)
		call("GET", "/collections/home-curation", nil, 200)
		call("PUT", path(c), models.UpdateCollection{Version: c.Version, CollectionContent: content}, 403)
		principal.Capabilities = authorization.CapabilitiesFor(authorization.RoleOperator)
	})

	// Concurrent requests use real PostgreSQL locks, not a mock transaction.
	race := func(requests [][3]string) []int {
		t.Helper()
		start := make(chan struct{})
		statuses := make(chan int, len(requests))
		var wg sync.WaitGroup
		for _, r := range requests {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				req := httptest.NewRequest(r[0], r[1], strings.NewReader(r[2]))
				req.Header.Set("Content-Type", "application/json")
				res, e := app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second})
				if e != nil {
					statuses <- 0
					return
				}
				defer res.Body.Close()
				statuses <- res.StatusCode
			}()
		}
		close(start)
		wg.Wait()
		close(statuses)
		out := []int{}
		for s := range statuses {
			out = append(out, s)
		}
		sort.Ints(out)
		return out
	}
	body := func(v any) string { b, _ := json.Marshal(v); return string(b) }
	t.Run("advisory lock version race and publish members race", func(t *testing.T) {
		in := models.UpdateCollection{Version: c.Version, CollectionContent: content}
		in.Name = "race-a"
		other := in
		other.Name = "race-b"
		if statuses := race([][3]string{{"PUT", path(c), body(in)}, {"PUT", path(c), body(other)}}); !reflect.DeepEqual(statuses, []int{200, 409}) {
			t.Fatal(statuses)
		}
		c = get(c)
		d := members(create("publish-race"), []int64{1, 2})
		statuses := race([][3]string{{"POST", path(d) + "/publish", body(models.CollectionVersion{Version: d.Version})}, {"PUT", path(d) + "/members", body(models.ReplaceCollectionMembers{Version: d.Version, GameIDs: []int64{1}})}})
		if !reflect.DeepEqual(statuses, []int{200, 409}) {
			t.Fatal(statuses)
		}
		d = get(d)
		if d.Status == "published" && d.MemberCount < 2 {
			t.Fatal(d)
		}
	})
	t.Run("advisory lock prevents curation racing last SFW removal", func(t *testing.T) {
		c = members(c, []int64{1, 2})
		h := home()
		statuses := race([][3]string{{"PUT", "/collections/home-curation", body(placement(h, map[int]int64{1: c.ID}))}, {"PUT", path(c) + "/members", body(models.ReplaceCollectionMembers{Version: c.Version, GameIDs: []int64{2, 3}})}})
		if !reflect.DeepEqual(statuses, []int{200, 200}) && !reflect.DeepEqual(statuses, []int{200, 400}) {
			t.Fatal(statuses)
		}
		c = get(c)
		if c.HomeSlot != nil || c.SFWMemberCount != 0 {
			t.Fatal(c)
		}
	})
	t.Run("audit actions and failed audit rollback", func(t *testing.T) {
		rows, e := pools["admin"].Query(ctx, `SELECT DISTINCT action FROM gfa_admin_audit_log WHERE resource LIKE 'gfg_game_collection%'`)
		if e != nil {
			t.Fatal(e)
		}
		got := map[string]bool{}
		for rows.Next() {
			var action string
			rows.Scan(&action)
			got[action] = true
		}
		rows.Close()
		for _, action := range []string{"create", "update", "members_update", "publish", "unpublish", "archive", "restore", "home_curation_update"} {
			if !got[action] {
				t.Fatal("missing audit", action)
			}
		}
		exec("admin", `ALTER TABLE gfa_admin_audit_log RENAME TO test_unavailable_audit`)
		in := models.UpdateCollection{Version: c.Version, CollectionContent: content}
		in.Name = "must rollback"
		call("PUT", path(c), in, 500)
		exec("admin", `ALTER TABLE test_unavailable_audit RENAME TO gfa_admin_audit_log`)
		if latest := get(c); latest.Version != c.Version || latest.Name == in.Name {
			t.Fatal(latest)
		}
	})

	assertHybridCollectionOperations(t, ctx, pools, call, race)
}
