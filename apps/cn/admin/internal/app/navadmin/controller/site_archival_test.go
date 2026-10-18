package controller

import (
	"strings"
	"testing"

	"github.com/gofurry/gofurry-admin/internal/app/navadmin/models"
)

func TestNormalizeSiteArchival(t *testing.T) {
	for _, code := range []string{"unreachable", "closed", "migrated", "duplicate", "owner_request", "other"} {
		note := "  已确认说明  "
		got, err := normalizeSiteArchival(models.SiteArchivalPayload{DeletionReasonCode: code, DeletionReason: &note})
		if err != nil || got.DeletionReason == nil || *got.DeletionReason != "已确认说明" {
			t.Fatalf("%s: %+v %v", code, got, err)
		}
	}
	for _, code := range []string{"", " ", "closed ", "Closed", "invented"} {
		if _, err := normalizeSiteArchival(models.SiteArchivalPayload{DeletionReasonCode: code}); err == nil {
			t.Fatalf("accepted %q", code)
		}
	}
	for _, size := range []int{1000, 1001} {
		note := strings.Repeat("🦊", size)
		_, err := normalizeSiteArchival(models.SiteArchivalPayload{DeletionReasonCode: "other", DeletionReason: &note})
		if (err == nil) != (size == 1000) {
			t.Fatalf("Unicode length %d: %v", size, err)
		}
	}
	blank := " \t "
	got, err := normalizeSiteArchival(models.SiteArchivalPayload{DeletionReasonCode: "other", DeletionReason: &blank})
	if err != nil || got.DeletionReason != nil {
		t.Fatalf("optional blank note: %+v %v", got, err)
	}
	invalid := "bad\x00text"
	if _, err := normalizeSiteArchival(models.SiteArchivalPayload{DeletionReasonCode: "other", DeletionReason: &invalid}); err == nil {
		t.Fatal("accepted NUL")
	}
}
