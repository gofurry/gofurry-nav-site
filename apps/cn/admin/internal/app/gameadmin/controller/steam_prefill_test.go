package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v3"
	steam "github.com/gofurry/steam-go"
	"github.com/gofurry/steam-go/web/storefront"
)

func TestFetchSteamAppDetailsResolvesUpstreamIdentity(t *testing.T) {
	// The alternate keys reproduce observed Steam responses; only the inner ID is trusted.
	for _, tc := range []struct {
		id  int64
		key string
	}{{10, "10"}, {550, "322070"}, {620, "323180"}, {1206410, "4602510"}} {
		for _, language := range []string{"schinese", "english"} {
			t.Run(fmt.Sprintf("%d/%s", tc.id, language), func(t *testing.T) {
				country := "CN"
				if language == "english" {
					country = "US"
				}
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != "/api/appdetails" || r.URL.Query().Get("appids") != strconv.FormatInt(tc.id, 10) || r.URL.Query().Get("cc") != country || r.URL.Query().Get("l") != language {
						t.Errorf("unexpected request: %s", r.URL)
					}
					w.Header().Set("Content-Type", "application/json")
					fmt.Fprintf(w, `{"%s":{"success":true,"data":{"steam_appid":%d,"name":"Requested game","short_description":"Description","developers":["Developer"],"publishers":["Publisher"]}}}`, tc.key, tc.id)
				}))
				defer server.Close()
				client, err := steam.NewClient(steam.WithStorefrontBaseURL(server.URL), steam.WithRetry(0))
				if err != nil {
					t.Fatal(err)
				}
				defer client.Close()
				data, err := fetchSteamAppDetails(context.Background(), client, tc.id, language)
				if err != nil || data.SteamAppID != uint32(tc.id) || data.Name != "Requested game" || data.ShortDescription != "Description" || !reflect.DeepEqual(data.Developers, []string{"Developer"}) || !reflect.DeepEqual(data.Publishers, []string{"Publisher"}) {
					t.Fatalf("prefill lost requested content: %+v, %v", data, err)
				}
			})
		}
	}
}

func TestFetchSteamAppDetailsRejectsUnverifiedIdentity(t *testing.T) {
	for name, body := range map[string]string{
		"different game":          `{"620":{"success":true,"data":{"steam_appid":620}}}`,
		"missing identity":        `{"550":{"success":true,"data":{"name":"Unverified"}}}`,
		"conflict plus alternate": `{"550":{"success":true,"data":{"steam_appid":620}},"322070":{"success":true,"data":{"steam_appid":550}}}`,
		"duplicate identity":      `{"550":{"success":true,"data":{"steam_appid":550}},"322070":{"success":true,"data":{"steam_appid":550}}}`,
		"unsuccessful":            `{"322070":{"success":false,"data":{"steam_appid":550}}}`,
	} {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, body)
			}))
			defer server.Close()
			client, err := steam.NewClient(steam.WithStorefrontBaseURL(server.URL), steam.WithRetry(0))
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			data, err := fetchSteamAppDetails(context.Background(), client, 550, "english")
			var apiErr *steam.APIError
			if !errors.As(err, &apiErr) || apiErr.Kind != steam.ErrorKindAPIResponse || !reflect.DeepEqual(data, storefront.AppDetailsData{}) {
				t.Fatalf("unverified content accepted or error lost: %+v, %v", data, err)
			}
		})
	}
}

func TestSteamPrefillRejectsInvalidAppIDBeforeRequest(t *testing.T) {
	app := fiber.New()
	api := New(nil, nil)
	app.Get("/prefill", api.ResolveSteamGamePrefill)
	for _, id := range []string{"", "0", "-1", "abc", "4294967296", "4294967846"} {
		response, err := app.Test(httptest.NewRequest(http.MethodGet, "/prefill?appid="+id, nil))
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != http.StatusBadRequest {
			t.Errorf("appid=%q returned %d, want 400", id, response.StatusCode)
		}
	}
}

func TestResolveSteamPrefillWarnsOnMissingFields(t *testing.T) {
	data, err := resolveSteamPrefill(context.Background(), 550, func(_ context.Context, language string) (storefront.AppDetailsData, error) {
		return storefront.AppDetailsData{Name: language}, nil
	}, func(context.Context) (string, error) { return "", nil })
	if err != nil || len(data.Warnings) != 3 || data.Name == "" || data.NameEn == "" {
		t.Fatalf("partial content should survive with locale/cover warnings: %+v, %v", data, err)
	}
}
