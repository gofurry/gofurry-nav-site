package controller

import (
	"testing"

	"github.com/gofurry/gofurry-admin/internal/app/navadmin/models"
)

func TestValidateSitePayloadDoesNotRequireLegacyDomains(t *testing.T) {
	t.Parallel()

	err := validateSitePayload(models.SitePayload{
		Name:   "测试站点",
		NameEn: "Test Site",
	})
	if err != nil {
		t.Fatalf("validateSitePayload() error = %v", err)
	}
}

func TestSiteGroupMapListOrderNewestFirst(t *testing.T) {
	t.Parallel()

	if siteGroupMapListOrder != "m.id DESC" {
		t.Fatalf("site group map order = %q, want newest ID first", siteGroupMapListOrder)
	}
}
