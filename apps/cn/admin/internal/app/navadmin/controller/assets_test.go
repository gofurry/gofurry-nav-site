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
	valid := patternPayload{Name: "Test", NameEn: "Test", LightColor: "#123abc", DarkColor: "#abcdef", LightOpacity: 0, DarkOpacity: 1, DefaultSizePx: 1}
	if err := validatePattern(valid); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*patternPayload){func(p *patternPayload) { p.DarkOpacity = 1.001 }, func(p *patternPayload) { p.LightColor = "red" }, func(p *patternPayload) { p.DefaultSizePx = 0 }, func(p *patternPayload) { p.NameEn = "" }} {
		value := valid
		mutate(&value)
		if validatePattern(value) == nil {
			t.Fatal("invalid appearance accepted")
		}
	}
	for _, body := range []string{`{"name":"x","variant":"mobile"}`, `{"name":"x","object_key":"anything"}`, `{"name":"x"} {}`} {
		if decodeAssetJSON([]byte(body), &heroPayload{}) == nil {
			t.Fatal("immutable fields or multiple payloads accepted")
		}
	}
}
