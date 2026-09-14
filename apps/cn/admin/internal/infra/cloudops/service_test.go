package cloudops

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	env "github.com/gofurry/gofurry-admin/config"
	teo "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/teo/v20220901"
)

func TestPurgeScopeAndProviderContracts(t *testing.T) {
	for _, req := range []PurgeRequest{{Type: "all"}, {Type: "purge_all"}, {Type: "host", Targets: []string{"example.com.evil.test"}}, {Type: "url", Targets: []string{"https://example.com@evil.test/a"}}, {Type: "url", Targets: []string{"https://example.com:8443/a"}}, {Type: "prefix", Targets: []string{"https://example.com/a?b=c"}}, {Type: "host", Targets: []string{"*.example.com"}}, {Type: "url", Targets: []string{"https://example.com/a#x"}}} {
		if _, err := ValidatePurge(req, "example.com"); err == nil {
			t.Fatalf("accepted out-of-contract purge: %+v", req)
		}
	}
	for _, kind := range []string{"url", "prefix", "host"} {
		target := "https://example.com/nav/"
		if kind == "host" {
			target = "example.com"
		}
		req, err := ValidatePurge(PurgeRequest{Type: kind, Targets: []string{target, target}}, "example.com")
		if err != nil || len(req.Targets) != 1 {
			t.Fatal("valid target rejected or not deduplicated")
		}
		payload := cloudflarePayload(req)
		field := map[string]string{"url": "files", "prefix": "prefixes", "host": "hosts"}[kind]
		if kind == "prefix" {
			target = "example.com/nav/"
		}
		if !reflect.DeepEqual(payload, map[string][]string{field: {target}}) {
			t.Fatal("wrong Cloudflare mapping")
		}
	}
}

type fakeTEO struct {
	request *teo.CreatePurgeTaskRequest
	query   *teo.DescribePurgeTasksRequest
}

func (f *fakeTEO) CreatePurgeTaskWithContext(_ context.Context, r *teo.CreatePurgeTaskRequest) (*teo.CreatePurgeTaskResponse, error) {
	f.request = r
	return &teo.CreatePurgeTaskResponse{Response: &teo.CreatePurgeTaskResponseParams{JobId: pointer("job")}}, nil
}
func (f *fakeTEO) DescribePurgeTasksWithContext(_ context.Context, r *teo.DescribePurgeTasksRequest) (*teo.DescribePurgeTasksResponse, error) {
	f.query = r
	return &teo.DescribePurgeTasksResponse{Response: &teo.DescribePurgeTasksResponseParams{TotalCount: pointer(uint64(1)), Tasks: []*teo.Task{{JobId: pointer("job"), Status: pointer("success"), Type: pointer("purge_host")}}}}, nil
}
func TestEdgeOneScopeAndTaskQuery(t *testing.T) {
	fake := &fakeTEO{}
	edge := &edgeOne{client: fake, zoneID: "configured-zone"}
	for _, kind := range []string{"url", "prefix", "host", "all"} {
		if _, err := edge.purge(context.Background(), PurgeRequest{Type: kind, Targets: []string{"example.com"}}); err != nil {
			t.Fatal(err)
		}
		if value(fake.request.Type) != "purge_"+kind || value(fake.request.ZoneId) != "configured-zone" {
			t.Fatal("wrong EdgeOne contract")
		}
		if kind != "url" && value(fake.request.Method) != "delete" {
			t.Fatal("purge did not clear cached bytes")
		}
	}
	tasks, err := edge.tasks(context.Background(), "job", 0)
	if err != nil || tasks.List[0].Type != "host" || value(fake.query.Filters[0].Name) != "job-id" {
		t.Fatal("task lookup or enum mapping failed")
	}
}
func TestCloudflareHeadersMappingAndErrorRedaction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.Header.Get("Authorization") != "Bearer secret-fixture" {
			t.Error("missing authenticated purge")
		}
		var payload map[string][]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(payload, map[string][]string{"prefixes": {"example.com/nav/"}}) {
			t.Error("wrong prefix representation")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(403)
		_, _ = w.Write([]byte(`{"success":false,"errors":[{"code":10000,"message":"secret-fixture"}]}`))
	}))
	defer server.Close()
	c := &cloudflare{client: server.Client(), endpoint: server.URL, token: "secret-fixture"}
	_, err := c.purge(context.Background(), PurgeRequest{Type: "prefix", Targets: []string{"https://example.com/nav/"}})
	if err == nil || strings.Contains(err.Error(), "secret-fixture") {
		t.Fatal("provider detail exposed credentials")
	}
}
func TestOverviewNeverSerializesSecretsOrInventsHealth(t *testing.T) {
	s, err := New(nil, env.AssetStorageConfig{Primary: env.ObjectStoreConfig{SecretAccessKey: "never-serialize"}}, env.CloudOpsConfig{})
	if err != nil {
		t.Fatal(err)
	}
	result := s.Overview(context.Background())
	data, _ := json.Marshal(result)
	if strings.Contains(string(data), "never-serialize") || result.Primary.Configured || result.Primary.Reachable || result.EdgeOne.Configured {
		t.Fatal("overview lied or exposed configuration secrets")
	}
}
