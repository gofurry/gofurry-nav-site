package bootstrap_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/shared/adminutil"
	"github.com/jackc/pgx/v5/pgxpool"
)

func testGameEditingRegression(t *testing.T, ctx context.Context, app *fiber.App, cookie *http.Cookie, pool *pgxpool.Pool, gameID int64, payload string) {
	t.Helper()
	// Test the real HTTP -> sqlc -> PostgreSQL path with more than one result page.
	for i := 0; i < 23; i++ {
		_, err := pool.Exec(ctx, `INSERT INTO gfg_tag (id,name,name_en,info,info_en,prefix,create_time,update_time) VALUES ($3,$1,$2,'','',-1,NOW(),NOW())`, fmt.Sprintf("分页回归-%02d", i), fmt.Sprintf("PaginationRegression-%02d", i), 900000+i)
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, query := range []string{"", "?keyword=PaginationRegression"} {
		response := requestJSON(t, app, http.MethodGet, "/options/tags"+query, "", cookie, http.StatusOK)
		var envelope integrationEnvelope
		if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		var page struct {
			List []adminutil.OptionItem `json:"list"`
		}
		if err := json.Unmarshal(envelope.Data, &page); err != nil {
			t.Fatal(err)
		}
		if len(page.List) != 10 {
			t.Fatalf("default tag page returned %d items", len(page.List))
		}
	}
	seen := map[int64]bool{}
	for page := 1; page <= 4; page++ {
		result := requestJSON(t, app, http.MethodGet, fmt.Sprintf("/options/tags?page_num=%d&page_size=10&keyword=PaginationRegression", page), "", cookie, http.StatusOK)
		var data struct {
			Total int64                  `json:"total"`
			List  []adminutil.OptionItem `json:"list"`
		}
		var envelope integrationEnvelope
		if err := json.NewDecoder(result.Body).Decode(&envelope); err != nil {
			t.Fatal(err)
		}
		result.Body.Close()
		if err := json.Unmarshal(envelope.Data, &data); err != nil {
			t.Fatal(err)
		}
		want := 10
		if page == 3 {
			want = 3
		}
		if page == 4 {
			want = 0
		}
		if data.Total != 23 || len(data.List) != want {
			t.Fatalf("page %d: total=%d count=%d", page, data.Total, len(data.List))
		}
		for _, item := range data.List {
			if seen[item.ID] {
				t.Fatalf("repeated tag %d across pages", item.ID)
			}
			seen[item.ID] = true
		}
	}
	assertOptionPage(t, requestJSON(t, app, http.MethodGet, "/options/tags?page_num=1&page_size=10&keyword=分页回归-22", "", cookie, http.StatusOK), 1)
	var body map[string]any
	if err := json.Unmarshal([]byte(payload), &body); err != nil {
		t.Fatal(err)
	}
	body["info"] = strings.Repeat("中", 400)
	body["info_en"] = strings.Repeat("🐺", 400)
	encoded, _ := json.Marshal(body)
	requestJSON(t, app, http.MethodPut, fmt.Sprintf("/game/games/%d", gameID), string(encoded), cookie, http.StatusOK)
	if n := queryInt64(t, ctx, pool, `SELECT char_length(info) + char_length(info_en) FROM gfg_game WHERE id=$1`, gameID); n != 800 {
		t.Fatalf("stored length=%d", n)
	}
	for _, field := range []string{"info", "info_en"} {
		original := body[field]
		body[field] = strings.Repeat("中", 401)
		encoded, _ = json.Marshal(body)
		for _, method := range []string{http.MethodPost, http.MethodPut} {
			path := "/game/games"
			if method == http.MethodPut {
				path += fmt.Sprintf("/%d", gameID)
			}
			result := requestJSON(t, app, method, path, string(encoded), cookie, http.StatusBadRequest)
			var envelope integrationEnvelope
			if err := json.NewDecoder(result.Body).Decode(&envelope); err != nil {
				t.Fatal(err)
			}
			result.Body.Close()
			if !strings.Contains(envelope.Message, "400") || strings.Contains(envelope.Message, "SQLSTATE") {
				t.Fatalf("not a business validation error: %s", envelope.Message)
			}
		}
		body[field] = original
	}
}
