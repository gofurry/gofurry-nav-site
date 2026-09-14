package bootstrap_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Runs in TestAdminThreeDatabasePersistence's disposable databases.
func testHomepageGroupCuration(t *testing.T, ctx context.Context, app *fiber.App, cookie *http.Cookie, navPool, adminPool *pgxpool.Pool) {
	t.Helper()
	groupID := responseID(t, requestJSON(t, app, http.MethodPost, "/nav/site-groups", `{"name":"Curation","name_en":"Curation","priority":1}`, cookie, http.StatusOK))
	ids := make([]string, 0, 10)
	for i := range 10 {
		siteID := responseID(t, requestJSON(t, app, http.MethodPost, "/nav/sites", fmt.Sprintf(`{"name":"Curation %d","name_en":"Curation","country":"CN","nsfw":"0","welfare":"0"}`, i), cookie, http.StatusOK))
		ids = append(ids, strconv.FormatInt(siteID, 10))
		closeResponse(requestJSON(t, app, http.MethodPut, "/nav/site-group-maps/bulk-replace", fmt.Sprintf(`{"owner_id":%d,"ids":[%d]}`, siteID, groupID), cookie, http.StatusOK))
	}
	endpoint := fmt.Sprintf("/nav/site-groups/%d/curation", groupID)
	read := func() (string, []string) {
		resp := requestJSON(t, app, http.MethodGet, endpoint, "", cookie, http.StatusOK)
		defer resp.Body.Close()
		var envelope integrationEnvelope
		if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
			t.Fatal(err)
		}
		var data struct {
			Revision string `json:"revision"`
			Sites    []struct {
				SiteID string `json:"site_id"`
			} `json:"sites"`
		}
		if err := json.Unmarshal(envelope.Data, &data); err != nil {
			t.Fatal(err)
		}
		ordered := make([]string, 0, len(data.Sites))
		for _, site := range data.Sites {
			ordered = append(ordered, site.SiteID)
		}
		return data.Revision, ordered
	}
	post := func(revision string, order []string, status int) {
		payload, err := json.Marshal(map[string]any{"revision": revision, "site_ids": order})
		if err != nil {
			t.Fatal(err)
		}
		closeResponse(requestJSON(t, app, http.MethodPut, endpoint, string(payload), cookie, status))
	}
	revision, initial := read()
	wantInitial := slices.Clone(ids)
	slices.Reverse(wantInitial) // All weights/timestamps tie; public homepage uses descending site ID.
	if !reflect.DeepEqual(initial, wantInitial) {
		t.Fatalf("initial public ordering mismatch: %v", initial)
	}
	post(revision, ids, http.StatusOK)
	_, saved := read()
	if !reflect.DeepEqual(saved, ids) {
		t.Fatalf("curated order=%v want=%v", saved, ids)
	}
	if count := queryInt64(t, ctx, navPool, `SELECT count(*) FROM gfn_site_group_map WHERE group_id=$1`, groupID); count != 10 {
		t.Fatalf("lost non-preview membership: %d", count)
	}
	for i, id := range ids {
		if weight := queryInt64(t, ctx, navPool, `SELECT weight FROM gfn_site_group_map WHERE group_id=$1 AND site_id=$2`, groupID, id); weight != int64(10-i) {
			t.Fatalf("incorrect weight: %d", weight)
		}
	}
	// Stale and incomplete writes must not partially reorder or discard rows.
	post(revision, wantInitial, http.StatusBadRequest)
	revision, _ = read()
	post(revision, ids[:8], http.StatusBadRequest)
	_, unchanged := read()
	if !reflect.DeepEqual(unchanged, ids) {
		t.Fatal("rejected save changed order")
	}
	closeResponse(requestJSON(t, app, http.MethodPut, "/nav/site-group-maps/bulk-replace", fmt.Sprintf(`{"owner_id":%s,"ids":[%d]}`, ids[0], groupID), cookie, http.StatusOK))
	if weight := queryInt64(t, ctx, navPool, `SELECT weight FROM gfn_site_group_map WHERE group_id=$1 AND site_id=$2`, groupID, ids[0]); weight != 10 {
		t.Fatalf("bulk replace lost weight: %d", weight)
	}
	post(revision, wantInitial, http.StatusBadRequest)
	if count := queryInt64(t, ctx, adminPool, `SELECT count(*) FROM gfa_admin_audit_log WHERE action='reorder' AND target_id=$1`, strconv.FormatInt(groupID, 10)); count != 1 {
		t.Fatalf("unexpected reorder audit count: %d", count)
	}
}
