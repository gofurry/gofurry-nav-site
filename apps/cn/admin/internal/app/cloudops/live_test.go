package cloudops

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	env "github.com/gofurry/gofurry-admin/config"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	"github.com/gofurry/gofurry-admin/internal/infra/assets"
	infra "github.com/gofurry/gofurry-admin/internal/infra/cloudops"
	"github.com/gofurry/gofurry-admin/internal/infra/db"
)

func TestRealDevCloudOps(t *testing.T) {
	path := os.Getenv("GOFURRY_ASSET_DEV_CONFIG")
	if path == "" {
		t.Skip("set GOFURRY_ASSET_DEV_CONFIG to ignored Admin dev YAML")
	}
	if err := env.MustInitServerConfig("gofurry-admin", path); err != nil {
		t.Fatal("dev config failed")
	}
	cfg := env.GetServerConfig()
	storageCfg := cfg.ExternalServices.AssetStorage
	cloudCfg := cfg.ExternalServices.CloudOps
	if !strings.Contains(storageCfg.Primary.Bucket, "-dev-") || !strings.HasPrefix(cloudCfg.EdgeOne.AssetHost, "assets-dev.") || !strings.HasPrefix(cloudCfg.Cloudflare.AssetHost, "assets-dev.") {
		t.Fatal("real acceptance requires development storage and CDN hosts")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	pools, err := db.Open(ctx, cfg)
	if err != nil {
		t.Fatal("dev pools failed")
	}
	defer pools.Close()
	storage, err := assets.New(storageCfg)
	if err != nil {
		t.Fatal(err)
	}
	cloud, err := infra.New(storage, storageCfg, cloudCfg)
	if err != nil {
		t.Fatal(err)
	}
	api := New(cloud, audit.New(pools.Admin))
	app := fiber.New()
	app.Get("/overview", api.Overview)
	app.Get("/object", api.Object)
	app.Post("/repair", api.RepairMirror)
	app.Post("/edgeone", api.EdgeOnePurge)
	app.Post("/cloudflare", api.CloudflarePurge)
	app.Get("/tasks", api.EdgeOnePurgeTasks)
	// Full-zone purge is intentionally not part of dev cloud acceptance: the
	// configured zone also serves the production main host. Its permission and
	// request mapping are verified by unit tests.
	call := func(method, path string, body any) json.RawMessage {
		t.Helper()
		data, _ := json.Marshal(body)
		req := httptest.NewRequest(method, path, bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "managed-assets-dev-cloud-acceptance")
		response, err := app.Test(req, fiber.TestConfig{Timeout: 40 * time.Second})
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		var envelope struct {
			Code    int
			Message string
			Data    json.RawMessage
		}
		if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
			t.Fatal(err)
		}
		if envelope.Code != 1 {
			t.Fatalf("%s %s: %s", method, path, envelope.Message)
		}
		return envelope.Data
	}
	var overview infra.Overview
	if err := json.Unmarshal(call("GET", "/overview", nil), &overview); err != nil {
		t.Fatal(err)
	}
	if !overview.Primary.Reachable || !overview.Mirror.Reachable || !overview.EdgeOne.Configured || !overview.Cloudflare.Configured {
		t.Fatal("real cloud overview incomplete")
	}
	object, err := assets.NewObject("pattern", 0, "pattern.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24"><path d="M1 1h3v3H1z"/></svg>`))
	if err != nil {
		t.Fatal(err)
	}
	publication, err := storage.Publish(ctx, object)
	if err != nil || publication.Mirror != "ready" {
		t.Fatal("real acceptance fixture publication failed")
	}
	var inspection infra.Inspection
	if err := json.Unmarshal(call("GET", "/object?key="+object.Key, nil), &inspection); err != nil {
		t.Fatal(err)
	}
	if inspection.Comparison != "matching" {
		t.Fatal("real object HEAD comparison failed")
	}
	call("POST", "/repair", map[string]string{"key": object.Key})
	for _, provider := range []string{"cloudflare", "edgeone"} {
		target := assets.URL(storageCfg.Mirror.PublicBaseURL, object.Key)
		if provider == "edgeone" {
			target = assets.URL(storageCfg.Primary.PublicBaseURL, object.Key)
		}
		var result struct {
			Result   infra.PurgeResult `json:"result"`
			Warnings []string          `json:"warnings"`
		}
		if err := json.Unmarshal(call("POST", "/"+provider, infra.PurgeRequest{Type: "url", Targets: []string{target}}), &result); err != nil {
			t.Fatal(err)
		}
		if result.Result.Status != "submitted" || len(result.Warnings) > 0 {
			t.Fatalf("%s purge incomplete: %+v", provider, result)
		}
		t.Logf("%s real URL purge submitted: %s", provider, result.Result.JobID)
		if provider == "edgeone" {
			awaitPurge(t, call, result.Result.JobID)
		}
	}
	var host struct {
		Result infra.PurgeResult `json:"result"`
	}
	if err := json.Unmarshal(call("POST", "/edgeone", infra.PurgeRequest{Type: "host", Targets: []string{cloudCfg.EdgeOne.AssetHost}}), &host); err != nil {
		t.Fatal(err)
	}
	if host.Result.Status != "submitted" {
		t.Fatal("dev asset host purge failed")
	}
	t.Logf("EdgeOne real dev asset Host purge submitted: %s", host.Result.JobID)
	awaitPurge(t, call, host.Result.JobID)
	client := &http.Client{Timeout: 15 * time.Second}
	for _, baseURL := range []string{storageCfg.Primary.PublicBaseURL, storageCfg.Mirror.PublicBaseURL} {
		response, err := client.Get(assets.URL(baseURL, object.Key))
		if err != nil {
			t.Fatal("CDN retrieval after purge failed")
		}
		data, readErr := io.ReadAll(io.LimitReader(response.Body, assets.MaxSize+1))
		response.Body.Close()
		if readErr != nil || response.StatusCode != 200 || !bytes.Equal(data, object.Data) {
			t.Fatal("CDN bytes changed after purge")
		}
	}
	var actions int
	if err := pools.Admin.QueryRow(ctx, `SELECT COUNT(DISTINCT action) FROM gfa_admin_audit_log WHERE user_agent='managed-assets-dev-cloud-acceptance' AND action IN ('cloud.mirror.repair','cloud.cloudflare.purge.url','cloud.edgeone.purge.url','cloud.edgeone.purge.host') AND after_data::jsonb->>'status'='completed'`).Scan(&actions); err != nil {
		t.Fatal(err)
	}
	if actions != 4 {
		t.Fatal("real cloud completion audits missing")
	}
	t.Log("real CloudOps acceptance passed: storage overview, HEAD inspector, repair, Cloudflare URL purge, EdgeOne URL/Host purge, completed task queries and audits")
}

func awaitPurge(t *testing.T, call func(string, string, any) json.RawMessage, id string) {
	t.Helper()
	deadline := time.Now().Add(45 * time.Second)
	for time.Now().Before(deadline) {
		var tasks infra.Tasks
		if err := json.Unmarshal(call("GET", "/tasks?job_id="+id, nil), &tasks); err != nil {
			t.Fatal(err)
		}
		if len(tasks.List) > 0 {
			done := true
			for _, task := range tasks.List {
				switch task.Status {
				case "success":
				case "processing":
					done = false
				default:
					t.Fatalf("EdgeOne task failed: %s (%s)", task.Status, task.Failure)
				}
			}
			if done {
				t.Logf("EdgeOne task completed: %s", id)
				return
			}
		}
		time.Sleep(3 * time.Second)
	}
	t.Fatalf("EdgeOne task did not finish within acceptance window: %s", id)
}
