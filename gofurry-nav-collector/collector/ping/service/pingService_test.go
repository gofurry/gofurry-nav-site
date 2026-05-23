package service

import (
	"testing"

	models2 "github.com/gofurry/gofurry-nav-collector/collector/ping/models"
)

func TestBuildPingTargetsKeepsSiteID(t *testing.T) {
	www := "www."
	domains, siteIDByDomain := buildPingTargets([]models2.GfnCollectorDomain{
		{ID: 1, SiteID: 101, Name: "example.com"},
		{ID: 2, SiteID: 101, Name: "example.com", Prefix: &www},
		{ID: 3, SiteID: 202, Name: "example.net"},
		{ID: 4, SiteID: 0, Name: "missing-site-id.example"},
	})
	if len(domains) != 3 {
		t.Fatalf("domains length = %d, want 3", len(domains))
	}
	if siteIDByDomain["example.com"] != 101 {
		t.Fatalf("example.com site id = %d, want 101", siteIDByDomain["example.com"])
	}
	if siteIDByDomain["www.example.com"] != 101 {
		t.Fatalf("www.example.com site id = %d, want 101", siteIDByDomain["www.example.com"])
	}
	if siteIDByDomain["example.net"] != 202 {
		t.Fatalf("example.net site id = %d, want 202", siteIDByDomain["example.net"])
	}
}
