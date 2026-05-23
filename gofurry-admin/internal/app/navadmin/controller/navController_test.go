package controller

import (
	"testing"

	"github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/navadmin/models"
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
