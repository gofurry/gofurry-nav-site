package controller

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/pkg/common"
)

func TestSiteIconCannotBeWrittenThroughOrdinaryPayload(t *testing.T) {
	app := fiber.New()
	app.Post("/", func(c fiber.Ctx) error {
		_, err := decodeSite(c)
		if err != nil {
			return common.NewResponse(c).Error(err)
		}
		return common.NewResponse(c).Success()
	})
	for _, value := range []string{`"https://example.com/icon.png"`, `"nav/sites/1/icon/aaaa.svg"`, `null`} {
		response, err := app.Test(httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"test","icon":`+value+`}`)))
		if err != nil {
			t.Fatal(err)
		}
		var body struct{ Code int }
		if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if body.Code == 1 {
			t.Fatal("ordinary site mutation accepted icon")
		}
	}
}

func TestAssetMetadataBoundaries(t *testing.T) {
	for _, body := range []string{`{"name":"x","variant":"mobile"}`, `{"name":"x","object_key":"anything"}`, `{"name":"x"} {}`} {
		if decodeAssetJSON([]byte(body), &heroPayload{}) == nil {
			t.Fatal("immutable fields or multiple payloads accepted")
		}
	}
}
