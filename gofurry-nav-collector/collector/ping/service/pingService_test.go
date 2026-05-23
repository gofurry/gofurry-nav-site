package service

import (
	"testing"

	models2 "github.com/gofurry/gofurry-nav-collector/collector/ping/models"
)

func TestBuildPingTargetsKeepsSiteID(t *testing.T) {
	domains, siteIDByDomain, err := buildPingTargets([]models2.Domain{
		{ID: 101, Domain: `{"domain":["example.com","www.example.com"]}`},
		{ID: 202, Domain: `{"domain":["example.net"]}`},
	})
	if err != nil {
		t.Fatalf("buildPingTargets() error = %v", err)
	}
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
