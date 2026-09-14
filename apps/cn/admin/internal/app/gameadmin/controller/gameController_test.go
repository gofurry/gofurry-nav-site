package controller

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/gofurry/gofurry-admin/internal/app/gameadmin/models"
	"github.com/gofurry/steam-go/web/storefront"
)

func TestResolveSteamPrefillPartialSuccess(t *testing.T) {
	for _, tc := range []struct {
		name          string
		zh, en, asset bool
	}{
		{"bilingual", true, true, true},
		{"Chinese unavailable", false, true, true},
		{"English unavailable", true, false, true},
		{"assets unavailable", true, true, false},
		{"assets only", false, false, true},
		{"fully unavailable", false, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := resolveSteamPrefill(context.Background(), 550, func(_ context.Context, lang string) (storefront.AppDetailsData, error) {
				if (lang == "schinese" && tc.zh) || (lang == "english" && tc.en) {
					return storefront.AppDetailsData{Name: lang, HeaderImage: "fallback.jpg", ShortDescription: lang + " description"}, nil
				}
				return storefront.AppDetailsData{}, errors.New("locale unavailable")
			}, func(context.Context) (string, error) {
				if tc.asset {
					return "asset.jpg", nil
				}
				return "", errors.New("assets unavailable")
			})
			if (err == nil) != (tc.zh || tc.en || tc.asset) {
				t.Fatalf("unexpected result: %+v, %v", data, err)
			}
			if (data.Name != "") != tc.zh || (data.NameEn != "") != tc.en {
				t.Fatalf("locale data lost: %+v", data)
			}
			if tc.asset && data.Header != "asset.jpg" {
				t.Fatalf("asset data lost: %+v", data)
			}
			if !tc.asset && (tc.zh || tc.en) && data.Header != "fallback.jpg" {
				t.Fatalf("header fallback lost: %+v", data)
			}
		})
	}
}

func TestResolveSteamPrefillRejectsEmptySuccess(t *testing.T) {
	_, err := resolveSteamPrefill(context.Background(), 550, func(context.Context, string) (storefront.AppDetailsData, error) {
		return storefront.AppDetailsData{}, nil
	}, func(context.Context) (string, error) { return "", nil })
	if err == nil {
		t.Fatal("synthetic links and appid must not count as Steam data")
	}
}

func TestGameDTOParsesJSONCollections(t *testing.T) {
	t.Parallel()

	resources := `[{"key":"k1","value":"v1"}]`
	groups := `[{"key":"official","value":"https://example.com"}]`
	links := `[{"key":"steamdb","value":"https://steamdb.info"}]`
	dto := gameDTO(models.Game{
		Developers: `["Dev A","Dev B"]`,
		Publishers: `["Pub A"]`,
		Resources:  &resources,
		Groups:     &groups,
		Links:      &links,
	})

	if !reflect.DeepEqual(dto.Developers, []string{"Dev A", "Dev B"}) {
		t.Fatalf("unexpected developers: %#v", dto.Developers)
	}
	if len(dto.Resources) != 1 || dto.Resources[0].Key != "k1" {
		t.Fatalf("unexpected resources: %#v", dto.Resources)
	}
	if len(dto.Groups) != 1 || dto.Groups[0].Key != "official" {
		t.Fatalf("unexpected groups: %#v", dto.Groups)
	}
	if len(dto.Links) != 1 || dto.Links[0].Key != "steamdb" {
		t.Fatalf("unexpected links: %#v", dto.Links)
	}
}

func TestSteamGamePrefillMapsRequestedFields(t *testing.T) {
	t.Parallel()

	dto := steamGamePrefill(550, storefront.AppDetailsData{
		Name:             "Left 4 Dead 2 中文",
		ShortDescription: " 中文简介 ",
		Website:          " https://www.l4d.com ",
		Developers:       []string{"Valve", "Valve"},
		Publishers:       []string{"Valve"},
	}, storefront.AppDetailsData{
		Name:             "Left 4 Dead 2",
		ShortDescription: " English description ",
	}, " https://cdn.example/header.jpg ")

	if dto.AppID != 550 || dto.Name != "Left 4 Dead 2 中文" || dto.NameEn != "Left 4 Dead 2" {
		t.Fatalf("unexpected identity: %#v", dto)
	}
	if dto.Info != "中文简介" || dto.InfoEn != "English description" {
		t.Fatalf("unexpected descriptions: info=%q info_en=%q", dto.Info, dto.InfoEn)
	}
	if len(dto.Groups) != 1 || dto.Groups[0].Key != "official" || dto.Groups[0].Value != "https://www.l4d.com" {
		t.Fatalf("unexpected groups: %#v", dto.Groups)
	}
	if !reflect.DeepEqual(dto.Developers, []string{"Valve"}) || !reflect.DeepEqual(dto.Publishers, []string{"Valve"}) {
		t.Fatalf("unexpected companies: developers=%#v publishers=%#v", dto.Developers, dto.Publishers)
	}
	if dto.Header != "https://cdn.example/header.jpg" {
		t.Fatalf("unexpected header: %q", dto.Header)
	}
	if len(dto.Links) != 2 || dto.Links[0].Key != "steamdb" || dto.Links[1].Key != "gamalytic" {
		t.Fatalf("unexpected links: %#v", dto.Links)
	}
}
