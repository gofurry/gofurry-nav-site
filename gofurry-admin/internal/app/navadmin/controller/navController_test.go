package controller

import (
	"reflect"
	"testing"

	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/navadmin/models"
)

func TestSiteDTOParsesDomainJSON(t *testing.T) {
	t.Parallel()

	dto := siteDTO(models.Site{
		ID:     1,
		Name:   "Test",
		Domain: `{"domain":["a.com","b.com","a.com"]}`,
	})

	want := []string{"a.com", "b.com"}
	if !reflect.DeepEqual(dto.Domains, want) {
		t.Fatalf("expected %v, got %v", want, dto.Domains)
	}
}
