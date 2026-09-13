package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	env "github.com/gofurry/gofurry-admin/config"
	"github.com/gofurry/gofurry-admin/internal/app/navadmin/models"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	"github.com/gofurry/gofurry-admin/internal/infra/assets"
	"github.com/gofurry/gofurry-admin/internal/infra/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type unavailableAssetStore struct{}

func (unavailableAssetStore) Put(context.Context, assets.Object) error {
	return errors.New("test outage")
}
func (unavailableAssetStore) Head(context.Context, string) (assets.Info, error) {
	return assets.Info{}, errors.New("test outage")
}
func (unavailableAssetStore) Get(context.Context, string) ([]byte, error) {
	return nil, errors.New("test outage")
}

// Real clouds + shared development databases. Only this test's records are
// changed; records are soft-deleted afterward, with audit and immutable files retained.
func TestRealDevManagedAssetAPI(t *testing.T) {
	path := os.Getenv("GOFURRY_ASSET_DEV_CONFIG")
	if path == "" {
		t.Skip("set GOFURRY_ASSET_DEV_CONFIG to ignored Admin dev YAML")
	}
	if err := env.MustInitServerConfig("gofurry-admin", path); err != nil {
		t.Fatal("dev configuration failed")
	}
	cfg := env.GetServerConfig()
	storageCfg := cfg.ExternalServices.AssetStorage
	if !strings.Contains(storageCfg.Primary.Bucket, "-dev-") || !strings.HasSuffix(storageCfg.Mirror.Bucket, "-dev") {
		t.Fatal("requires development cloud configuration")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pools, err := db.Open(ctx, cfg)
	if err != nil {
		t.Fatal("cannot open dev database pools")
	}
	defer pools.Close()
	storage, err := assets.New(storageCfg)
	if err != nil {
		t.Fatal(err)
	}
	api := New(pools.Nav, audit.New(pools.Admin)).WithAssets(storage, storageCfg.Primary.PublicBaseURL, storageCfg.Mirror.PublicBaseURL)
	app := fiber.New(fiber.Config{BodyLimit: 6 * 1024 * 1024})
	app.Post("/hero", api.CreateHeroAsset)
	app.Get("/hero/:id", api.GetHeroAsset)
	app.Put("/hero/:id", api.UpdateHeroAsset)
	app.Post("/hero/:id/file", api.ReplaceHeroAssetFile)
	app.Delete("/hero/:id", api.DeleteHeroAsset)
	app.Post("/pattern", api.CreateBackgroundPattern)
	app.Get("/pattern/:id", api.GetBackgroundPattern)
	app.Put("/pattern/:id", api.UpdateBackgroundPattern)
	app.Post("/pattern/:id/file", api.ReplaceBackgroundPatternFile)
	app.Delete("/pattern/:id", api.DeleteBackgroundPattern)
	app.Post("/site/:id/icon", api.ReplaceSiteIcon)
	app.Delete("/site/:id/icon", api.ClearSiteIcon)
	app.Put("/site/:id", api.UpdateSite)
	call := func(method, path, contentType string, body io.Reader, wantSuccess bool) json.RawMessage {
		t.Helper()
		req := httptest.NewRequest(method, path, body)
		req.Header.Set("Content-Type", contentType)
		response, err := app.Test(req, fiber.TestConfig{Timeout: 40 * time.Second})
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		var result struct {
			Code    int
			Message string
			Data    json.RawMessage
		}
		if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
			t.Fatal(err)
		}
		if (result.Code == 1) != wantSuccess {
			t.Fatalf("%s %s: %s", method, path, result.Message)
		}
		return result.Data
	}
	expectedMirror := "ready"
	upload := func(path, name string, data []byte, fields map[string]string, success bool) publicationDTO {
		t.Helper()
		body := new(bytes.Buffer)
		form := multipart.NewWriter(body)
		for key, value := range fields {
			if err := form.WriteField(key, value); err != nil {
				t.Fatal(err)
			}
		}
		file, err := form.CreateFormFile("file", name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := file.Write(data); err != nil {
			t.Fatal(err)
		}
		if err := form.Close(); err != nil {
			t.Fatal(err)
		}
		result := call("POST", path, form.FormDataContentType(), body, success)
		var publication publicationDTO
		if success {
			if err := json.Unmarshal(result, &publication); err != nil {
				t.Fatal(err)
			}
			if publication.Primary != "ready" || publication.Mirror != expectedMirror {
				t.Fatalf("real cloud publication incomplete: %+v", publication)
			}
		}
		return publication
	}
	jsonCall := func(method, path string, body any, success bool) json.RawMessage {
		data, _ := json.Marshal(body)
		return call(method, path, "application/json", bytes.NewReader(data), success)
	}
	avif, err := os.ReadFile("../../../infra/assets/testdata/hero.avif")
	if err != nil {
		t.Fatal(err)
	}
	for _, variant := range []string{"desktop", "mobile"} {
		publication := upload("/hero", "test.avif", avif, map[string]string{"name": "[dev acceptance] " + variant, "variant": variant, "enabled": "false"}, true)
		item := publication.Item.(map[string]any)
		id := item["id"].(string)
		defer jsonCall("DELETE", "/hero/"+id, nil, true)
		if item["variant"] != variant || !strings.HasPrefix(publication.ObjectKey, "nav/hero/"+variant+"/") {
			t.Fatal("hero pools crossed")
		}
		jsonCall("PUT", "/hero/"+id, map[string]any{"name": "[dev acceptance] " + variant, "enabled": true}, true)
		jsonCall("PUT", "/hero/"+id, map[string]any{"name": "test", "variant": "other", "enabled": true}, false)
		upload("/hero/"+id+"/file", "replace.avif", avif, nil, true)
		jsonCall("PUT", "/hero/"+id, map[string]any{"name": "[dev acceptance] " + variant, "enabled": false}, true)
		var count int
		if err := pools.Admin.QueryRow(ctx, `SELECT COUNT(DISTINCT action) FROM gfa_admin_audit_log WHERE resource='gfn_home_hero_asset' AND target_id=$1 AND action IN ('hero.create','hero.enable','hero.disable','hero.file.replace')`, id).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 4 {
			t.Fatal("missing hero audits")
		}
	}
	pattern := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24"><path d="M1 1h3v3H1z"/></svg>`)
	publication := upload("/pattern", "pattern.svg", pattern, map[string]string{"name": "[dev acceptance] Pattern", "name_en": "Dev acceptance", "light_color": "#123456", "dark_color": "#abcdef", "light_opacity": "0.1", "dark_opacity": "0.2", "default_size_px": "96", "enabled": "false", "sort_order": "999999"}, true)
	patternID := publication.Item.(map[string]any)["id"].(string)
	defer jsonCall("DELETE", "/pattern/"+patternID, nil, true)
	upload("/pattern/"+patternID+"/file", "unsafe.svg", []byte(`<svg><script/></svg>`), nil, false)
	upload("/pattern/"+patternID+"/file", "safe.svg", pattern, nil, true)
	patternUpdate := patternPayload{Name: "[dev acceptance] Pattern", NameEn: "Dev acceptance", LightColor: "#111111", DarkColor: "#eeeeee", LightOpacity: 0.25, DarkOpacity: 0.5, DefaultSizePx: 128, Enabled: false, SortOrder: 999999}
	jsonCall("PUT", "/pattern/"+patternID, patternUpdate, true)
	var got patternDTO
	if err := json.Unmarshal(jsonCall("GET", "/pattern/"+patternID, nil, true), &got); err != nil {
		t.Fatal(err)
	}
	if got.DefaultSizePx != 128 || got.LightOpacity != 0.25 {
		t.Fatal("pattern metadata lost")
	}
	site, storeErr := api.store.createSite(ctx, audit.SystemMeta("managed-assets-dev-acceptance"), models.SitePayload{Name: "[dev acceptance] Icon", NameEn: "Dev acceptance", Nsfw: "0", Welfare: "0"})
	if storeErr != nil {
		t.Fatal(storeErr)
	}
	defer func() {
		if err := api.store.deleteSite(context.Background(), audit.SystemMeta("managed-assets-dev-acceptance"), site.ID); err != nil {
			t.Error(err)
		}
	}()
	sitePath := "/site/" + strconv.FormatInt(site.ID, 10)
	icon := upload(sitePath+"/icon", "Icon.SVG", pattern, nil, true)
	jsonCall("PUT", sitePath, map[string]any{"name": "test", "icon": nil}, false)
	jsonCall("PUT", sitePath, models.SitePayload{Name: "[dev acceptance] Icon updated", NameEn: "Dev acceptance", Nsfw: "0", Welfare: "0"}, true)
	saved, e := api.store.q.GetSite(ctx, site.ID)
	if e != nil || saved.Icon == nil || *saved.Icon != icon.ObjectKey {
		t.Fatal("ordinary update damaged icon")
	}
	jsonCall("DELETE", sitePath+"/icon", nil, true)
	saved, e = api.store.q.GetSite(ctx, site.ID)
	if e != nil || saved.Icon != nil {
		t.Fatal("clear did not use SQL NULL")
	}
	for _, store := range []assets.ObjectStore{storage.Primary, storage.Mirror} {
		info, e := store.Head(ctx, icon.ObjectKey)
		if e != nil || !info.Exists {
			t.Fatal("clear removed immutable object")
		}
	}
	// A failed audit must leave the business row unchanged, even after cloud publication.
	closed, err := pgxpool.New(ctx, cfg.DataBase.Postgres.ConnectionString())
	if err != nil {
		t.Fatal("test audit pool configuration failed")
	}
	closed.Close()
	api.store.audit = audit.New(closed)
	jsonCall("PUT", "/pattern/"+patternID, patternPayload{Name: "should roll back", NameEn: "unchanged", LightColor: "#111111", DarkColor: "#eeeeee", LightOpacity: 0.1, DarkOpacity: 0.1, DefaultSizePx: 64}, false)
	api.store.audit = audit.New(pools.Admin)
	current, e := api.store.q.GetBackgroundPattern(ctx, got.ID)
	if e != nil || current.Name != patternUpdate.Name {
		t.Fatal("audit failure did not roll back")
	}
	api.assetStorage = &assets.Service{Primary: storage.Primary, Mirror: unavailableAssetStore{}}
	expectedMirror = "failed"
	partial := upload("/hero", "test.avif", avif, map[string]string{"name": "[dev acceptance] Mirror outage", "variant": "desktop", "enabled": "false"}, true)
	partialID := partial.Item.(map[string]any)["id"].(string)
	defer jsonCall("DELETE", "/hero/"+partialID, nil, true)
	if len(partial.Warnings) == 0 {
		t.Fatal("mirror outage warning missing")
	}
	beforeCount, e := api.store.q.CountHeroAssets(ctx, "desktop")
	if e != nil {
		t.Fatal(e)
	}
	api.assetStorage = &assets.Service{Primary: unavailableAssetStore{}, Mirror: storage.Mirror}
	upload("/hero", "test.avif", avif, map[string]string{"name": "must not publish", "variant": "desktop"}, false)
	afterCount, e := api.store.q.CountHeroAssets(ctx, "desktop")
	if e != nil || beforeCount != afterCount {
		t.Fatal("primary failure published a database row")
	}
	api.assetStorage = storage
	t.Log(fmt.Sprintf("real API acceptance passed: two hero pools, pattern metadata, site icon replace/clear, immutable objects, audit rollback and provider outage behavior (site %d)", site.ID))
}
