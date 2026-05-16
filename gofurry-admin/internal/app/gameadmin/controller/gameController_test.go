package controller

import (
	"reflect"
	"testing"

	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/gameadmin/models"
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
