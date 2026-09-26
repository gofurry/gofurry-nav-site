package details

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	"github.com/gofurry/gofurry-game-collector/collector/game/models"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/steamclient"
	steam "github.com/gofurry/steam-go"
	"github.com/gofurry/steam-go/addons/assets"
	"github.com/gofurry/steam-go/web/storefront"
)

func appDetailsTestClient(t *testing.T, id uint32, body string) *steam.Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/appdetails" || r.URL.Query().Get("appids") != strconv.FormatUint(uint64(id), 10) {
			t.Errorf("unexpected AppDetails request: %s", r.URL)
		}
		country, language := r.URL.Query().Get("cc"), r.URL.Query().Get("l")
		if (country != "CN" || language != "schinese") && (country != "US" && country != "HK" || language != "english") {
			t.Errorf("unexpected region/language: %s", r.URL)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, body)
	}))
	t.Cleanup(server.Close)
	client, err := steam.NewClient(steam.WithStorefrontBaseURL(server.URL), steam.WithRetry(0))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(client.Close)
	return client
}

func appDetailsRunner(t *testing.T, client *steam.Client) func(context.Context, steamclient.Bucket, func(context.Context, *steam.Client) error) error {
	t.Helper()
	return func(ctx context.Context, bucket steamclient.Bucket, run func(context.Context, *steam.Client) error) error {
		if bucket != steamclient.BucketStore {
			t.Fatalf("AppDetails used traffic bucket %s", bucket)
		}
		return run(ctx, client)
	}
}

func TestFetchAppDetailsResolvesIdentityAndPreservesRaw(t *testing.T) {
	for _, tc := range []struct {
		id  uint32
		key string
	}{{10, "10"}, {550, "322070"}, {620, "323180"}, {1206410, "4602510"}} {
		t.Run(strconv.Itoa(int(tc.id)), func(t *testing.T) {
			// Include whitespace and an unknown field to catch accidental re-encoding.
			body := fmt.Sprintf("\n"+`{"%s": {"success":true,"data":{"steam_appid":%d,"name":"Requested game","short_description":"Description","future_field":{"keep":true}}}}`+"\n", tc.key, tc.id)
			client := appDetailsTestClient(t, tc.id, body)
			collector := NewCollector(nil, nil)
			collector.runSteam = appDetailsRunner(t, client)
			for _, plan := range collector.requests {
				data, raw, err := collector.fetchAppDetails(context.Background(), tc.id, plan)
				if err != nil || data.SteamAppID != tc.id || data.Name != "Requested game" || data.ShortDescription != "Description" {
					t.Fatalf("requested details lost: %+v, %v", data, err)
				}
				if !bytes.Equal(raw, []byte(body)) {
					t.Fatal("original snapshot payload was modified")
				}
			}
		})
	}
}

func TestFetchAppDetailsRejectsUnverifiedIdentityAndKeepsRaw(t *testing.T) {
	for name, body := range map[string]string{
		"different game":          `{"620":{"success":true,"data":{"steam_appid":620}}}`,
		"missing identity":        `{"550":{"success":true,"data":{"name":"Unverified"}}}`,
		"conflict plus alternate": `{"550":{"success":true,"data":{"steam_appid":620}},"322070":{"success":true,"data":{"steam_appid":550}}}`,
		"duplicate identity":      `{"550":{"success":true,"data":{"steam_appid":550}},"322070":{"success":true,"data":{"steam_appid":550}}}`,
		"unsuccessful":            `{"322070":{"success":false,"data":{"steam_appid":550}}}`,
	} {
		t.Run(name, func(t *testing.T) {
			client := appDetailsTestClient(t, 550, body)
			collector := NewCollector(nil, nil)
			collector.runSteam = appDetailsRunner(t, client)
			data, raw, err := collector.fetchAppDetails(context.Background(), 550, collector.requests[0])
			var apiErr *steam.APIError
			if !errors.As(err, &apiErr) || apiErr.Kind != steam.ErrorKindAPIResponse || !reflect.DeepEqual(data, storefront.AppDetailsData{}) {
				t.Fatalf("unverified content accepted or typed error lost: %+v, %v", data, err)
			}
			if !bytes.Equal(raw, []byte(body)) {
				t.Fatal("failure discarded or modified the original payload")
			}
		})
	}
}

func TestCollectGameIdentityControlsPersistence(t *testing.T) {
	for _, tc := range []struct {
		name string
		id   uint32
	}{{"key drift", 550}, {"identity conflict", 620}} {
		t.Run(tc.name, func(t *testing.T) {
			body := fmt.Sprintf(`{"322070":{"success":true,"data":{"steam_appid":%d,"type":"game","name":"Requested game","short_description":"Description"}}}`, tc.id)
			client := appDetailsTestClient(t, 550, body)
			repo := &recordingRepository{}
			collector := NewCollector(nil, repo)
			collector.runSteam = appDetailsRunner(t, client)
			collector.fetchStoreBrowseFn = func(context.Context, uint32, requestPlan) ([]assets.URLItem, error) {
				return []assets.URLItem{{Kind: assets.KindLibraryCapsule, URL: "https://example.test/library_capsule.jpg"}}, nil
			}
			result, err := collector.CollectGame(context.Background(), models.GameID{ID: 1, Appid: 550})
			if tc.id != 550 {
				if err == nil || result.Status != domain.StatusFailed || repo.calls != 0 {
					t.Fatalf("wrong-game content reached persistence: %+v, %v, calls=%d", result, err, repo.calls)
				}
				return
			}
			if err != nil || result.Status != domain.StatusSuccess || repo.calls != 1 || repo.data.Details.AppID != 550 || len(repo.data.Snapshots) != 3 || len(repo.data.Localized) != 2 || len(repo.data.Prices) != 3 {
				t.Fatalf("resolved observation not saved correctly: %+v, %v, data=%+v", result, err, repo.data)
			}
			for _, snapshot := range repo.data.Snapshots {
				if snapshot.AppID != 550 || !bytes.Equal(snapshot.RawPayload, []byte(body)) || snapshot.PayloadHash != hashPayload([]byte(body)) {
					t.Fatalf("snapshot identity or provenance changed: %+v", snapshot)
				}
			}
		})
	}
}
