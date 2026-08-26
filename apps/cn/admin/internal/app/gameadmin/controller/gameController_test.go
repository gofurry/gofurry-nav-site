package controller

import (
	"reflect"
	"testing"

	"github.com/gofurry/gofurry-admin/internal/app/gameadmin/models"
	"github.com/gofurry/steam-go/web/storefront"
)

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
