package bootstrap_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
)

func testTagDomainRegression(t *testing.T, ctx context.Context, app *fiber.App, cookie *http.Cookie, pool *pgxpool.Pool, gameID, otherGameID int64) {
	t.Helper()
	catPayload := `{"code":"integration-category","name":"类别","name_en":"Category","info":"","info_en":"","sort_order":50}`
	category := responseID(t, requestJSON(t, app, http.MethodPost, "/game/tag-categories", catPayload, cookie, http.StatusOK))
	tagPayload := fmt.Sprintf(`{"code":"integration-tag","name":"标签","name_en":"Tag","info":"","info_en":"","category_id":%d}`, category)
	tagID := responseID(t, requestJSON(t, app, http.MethodPost, "/game/tags", tagPayload, cookie, http.StatusOK))
	tagPath := fmt.Sprintf("/game/tags/%d", tagID)
	catPath := fmt.Sprintf("/game/tag-categories/%d", category)
	classificationPath := fmt.Sprintf("/game/games/%d/classification", gameID)
	bad := fmt.Sprintf(`{"code":"changed-code","name":"标签","name_en":"Tag","category_id":%d}`, category)
	requestJSON(t, app, http.MethodPut, tagPath, bad, cookie, http.StatusBadRequest).Body.Close()
	requestJSON(t, app, http.MethodPost, "/game/tags", `{"code":"Bad_Code","name":"x","name_en":"x","category_id":1}`, cookie, http.StatusBadRequest).Body.Close()
	requestJSON(t, app, http.MethodDelete, catPath, "", cookie, http.StatusBadRequest).Body.Close()
	classify := fmt.Sprintf(`{"weight":42,"primary_tag_id":%d,"secondary_tag_id":null,"tag_ids":[]}`, tagID)
	requestJSON(t, app, http.MethodPut, classificationPath, classify, cookie, http.StatusOK).Body.Close()
	var count int
	var created, updated time.Time
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM gfg_game_tag WHERE game_id=$1 AND tag_id=$2 AND role='primary'`, gameID, tagID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("role union count=%d err=%v", count, err)
	}
	if _, err := pool.Exec(ctx, `UPDATE gfg_game_tag SET create_time='2020-01-01',update_time='2020-02-01' WHERE game_id=$1`, gameID); err != nil {
		t.Fatal(err)
	}
	requestJSON(t, app, http.MethodPut, classificationPath, classify, cookie, http.StatusOK).Body.Close()
	if err := pool.QueryRow(ctx, `SELECT create_time,update_time FROM gfg_game_tag WHERE game_id=$1 AND tag_id=$2`, gameID, tagID).Scan(&created, &updated); err != nil {
		t.Fatal(err)
	}
	if created.Year() != 2020 || updated.Month() != time.February {
		t.Fatal("no-op classification changed relation timestamps")
	}
	requestJSON(t, app, http.MethodDelete, tagPath, "", cookie, http.StatusBadRequest).Body.Close()
	requestJSON(t, app, http.MethodPut, classificationPath, fmt.Sprintf(`{"weight":999,"primary_tag_id":%d,"tag_ids":[999999999]}`, tagID), cookie, http.StatusBadRequest).Body.Close()
	var weight int64
	if err := pool.QueryRow(ctx, `SELECT weight FROM gfg_game WHERE id=$1`, gameID).Scan(&weight); err != nil || weight != 42 {
		t.Fatalf("invalid classification was not rolled back: weight=%d %v", weight, err)
	}
	requestJSON(t, app, http.MethodPut, classificationPath, fmt.Sprintf(`{"primary_tag_id":%d,"secondary_tag_id":%d,"tag_ids":[]}`, tagID, tagID), cookie, http.StatusBadRequest).Body.Close()
	// An unrelated source cache must also be invalidated by membership/role changes.
	seedCache := func() {
		t.Helper()
		if _, err := pool.Exec(ctx, `INSERT INTO gfg_game_recommendations(source_game_id,target_game_id,score,display_score,rank,reason_json,algorithm_version,computed_at) VALUES($1,$2,1,1,1,'[]','similar-v2.4.0-hybrid-cbf',NOW())`, otherGameID, gameID); err != nil {
			t.Fatal(err)
		}
	}
	assertEmpty := func() {
		t.Helper()
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM gfg_game_recommendations`).Scan(&count); err != nil || count != 0 {
			t.Fatalf("global cache not invalidated: %d %v", count, err)
		}
	}
	seedCache()
	requestJSON(t, app, http.MethodPut, classificationPath, fmt.Sprintf(`{"weight":42,"secondary_tag_id":%d,"tag_ids":[]}`, tagID), cookie, http.StatusOK).Body.Close()
	assertEmpty()
	if err := pool.QueryRow(ctx, `SELECT create_time FROM gfg_game_tag WHERE game_id=$1 AND tag_id=$2`, gameID, tagID).Scan(&created); err != nil || created.Year() != 2020 {
		t.Fatal("role change lost original create_time")
	}
	seedCache()
	requestJSON(t, app, http.MethodPut, tagPath, fmt.Sprintf(`{"code":"integration-tag","name":"标签","name_en":"Tag","category_id":1}`), cookie, http.StatusOK).Body.Close()
	assertEmpty()
	seedCache()
	requestJSON(t, app, http.MethodPut, tagPath, fmt.Sprintf(`{"code":"integration-tag","name":"改名","name_en":"Renamed","category_id":1}`), cookie, http.StatusOK).Body.Close()
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM gfg_game_recommendations`).Scan(&count); err != nil || count != 1 {
		t.Fatal("label-only edit invalidated scores")
	}
	requestJSON(t, app, http.MethodPut, classificationPath, `{"weight":42,"tag_ids":[]}`, cookie, http.StatusOK).Body.Close()
	assertEmpty()
	requestJSON(t, app, http.MethodPut, tagPath, tagPayload, cookie, http.StatusOK).Body.Close()
	requestJSON(t, app, http.MethodDelete, tagPath, "", cookie, http.StatusOK).Body.Close()
	result := requestJSON(t, app, http.MethodGet, "/options/tags?keyword=integration-tag", "", cookie, http.StatusOK)
	var envelope integrationEnvelope
	if err := json.NewDecoder(result.Body).Decode(&envelope); err != nil {
		t.Fatal(err)
	}
	result.Body.Close()
	var page struct {
		Total int `json:"total"`
	}
	if err := json.Unmarshal(envelope.Data, &page); err != nil || page.Total != 0 {
		t.Fatalf("archived option exposed: %s", envelope.Data)
	}
	requestJSON(t, app, http.MethodPut, classificationPath, classify, cookie, http.StatusBadRequest).Body.Close()
	requestJSON(t, app, http.MethodDelete, catPath, "", cookie, http.StatusOK).Body.Close()
	requestJSON(t, app, http.MethodPost, tagPath+"/restore", "", cookie, http.StatusBadRequest).Body.Close()
	requestJSON(t, app, http.MethodPost, catPath+"/restore", "", cookie, http.StatusOK).Body.Close()
	requestJSON(t, app, http.MethodPost, tagPath+"/restore", "", cookie, http.StatusOK).Body.Close()
}
