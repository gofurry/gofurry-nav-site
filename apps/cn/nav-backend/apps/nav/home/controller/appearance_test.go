package controller

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/home/models"
	"github.com/gofurry/gofurry-nav-backend/common"
)

type appearanceStub struct {
	homeReader
	query     models.HeroCatalogQuery
	selection models.HeroSelection
}

func (s *appearanceStub) GetHeroes(_ context.Context, q models.HeroCatalogQuery) (models.HeroCatalog, common.GFError) {
	s.query = q
	return models.HeroCatalog{SchemaVersion: 1, Items: []models.HeroCatalogItem{{ID: 9007199254740993}}}, nil
}
func (s *appearanceStub) GetHomeHero(_ context.Context, selection models.HeroSelection) models.HomeHeroResponse {
	s.selection = selection
	return models.HomeHeroResponse{}
}

func TestHeroCatalogValidation(t *testing.T) {
	stub := &appearanceStub{}
	app := fiber.New()
	app.Get("/heroes", New(stub).GetHeroes)
	for _, query := range []string{"", "?variant=tablet", "?variant=mobile&page_size=25", "?variant=desktop&page_num=0", "?variant=desktop&page_num=1000001", "?variant=desktop&selected_id=abc", "?variant=desktop&selected_id=9223372036854775808"} {
		res, err := app.Test(httptest.NewRequest("GET", "/heroes"+query, nil))
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != 400 {
			t.Fatalf("accepted invalid query %s", query)
		}
	}
	res, err := app.Test(httptest.NewRequest("GET", "/heroes?variant=desktop&page_num=2&page_size=12&selected_id=9007199254740993", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var body struct {
		Data models.HeroCatalog `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if stub.query.SelectedID != 9007199254740993 || stub.query.PageNum != 2 || len(body.Data.Items) != 1 {
		t.Fatal("catalog query or string ID contract lost")
	}
}

func TestHeroRequestSelection(t *testing.T) {
	stub := &appearanceStub{}
	app := fiber.New()
	app.Get("/hero", New(stub).GetHomeHero)
	for _, tc := range []struct {
		query string
		want  models.HeroSelection
	}{
		{"?hero_desktop_id=9007199254740993&hero_mobile_id=20", models.HeroSelection{DesktopID: 9007199254740993, MobileID: 20}},
		{"?hero_desktop_id=invalid&hero_mobile_id=20", models.HeroSelection{MobileID: 20}},
		{"?hero_desktop_id=9223372036854775808", models.HeroSelection{}},
		{"?hero_mode=local", models.HeroSelection{Local: true}},
	} {
		res, err := app.Test(httptest.NewRequest("GET", "/hero"+tc.query, nil))
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if stub.selection != tc.want {
			t.Fatalf("unexpected selection %+v", stub.selection)
		}
	}
}
