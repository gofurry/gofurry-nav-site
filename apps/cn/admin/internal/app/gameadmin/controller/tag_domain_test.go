package controller

import (
	"strings"
	"testing"

	"github.com/gofurry/gofurry-admin/internal/app/gameadmin/models"
)

func TestTagCodeAndClassificationValidation(t *testing.T) {
	for _, code := range []string{"adult", "2-5d", "future-tag"} {
		if e := validateTagText(code, "名", "Name", "", ""); e != nil {
			t.Fatal(e)
		}
	}
	for _, code := range []string{"", "UPPER", "bad_code", "-bad", "bad--code", strings.Repeat("a", 65)} {
		if e := validateTagText(code, "名", "Name", "", ""); e == nil {
			t.Fatalf("accepted %q", code)
		}
	}
	p, s := int64(17), int64(99)
	roles, e := classificationRoles(models.GameClassificationPayload{PrimaryTagID: &p, SecondaryTagID: &s, TagIDs: []int64{4, 4}})
	if e != nil || len(roles) != 3 || roles[p] != "primary" || roles[s] != "secondary" {
		t.Fatalf("union=%v err=%v", roles, e)
	}
	if _, e := classificationRoles(models.GameClassificationPayload{PrimaryTagID: &p, SecondaryTagID: &p}); e == nil {
		t.Fatal("accepted duplicate roles")
	}
}
