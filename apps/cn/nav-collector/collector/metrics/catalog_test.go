package metrics

import (
	"strings"
	"testing"

	navsqlc "github.com/gofurry/gofurry-nav-collector/internal/db/nav/sqlc"
)

func TestValidateCatalogRows(t *testing.T) {
	rows := make([]navsqlc.GfnMetricRegistry, 0, len(evaluatorCatalog))
	for _, contract := range evaluatorCatalog {
		freshness := contract.FreshnessSeconds
		rows = append(rows, navsqlc.GfnMetricRegistry{
			MetricKey: contract.Key, MetricVersion: contract.Version,
			MetricKind: contract.Kind, EntityLevel: contract.EntityLevel, TimeGrain: contract.TimeGrain,
			SourceFacts: contract.SourceFacts, EligibilityPolicy: contract.EligibilityPolicy,
			StatePolicy: contract.StatePolicy, CoveragePolicy: contract.CoveragePolicy,
			FreshnessSeconds: &freshness, AllowedDimensions: contract.AllowedDimensions, Status: "active",
		})
	}
	if err := validateCatalogRows(rows); err != nil {
		t.Fatalf("valid catalog rejected: %v", err)
	}
	for i := range rows {
		if rows[i].MetricKey == "http2_adoption" && rows[i].MetricVersion == 1 {
			rows[i].Status = "retired"
		}
	}
	if err := validateCatalogRows(rows); err != nil {
		t.Fatalf("HTTP/2 retired v1 and active v2 rejected: %v", err)
	}
	for _, version := range []int32{1, 2} {
		if _, ok := contractFor("http2_adoption", version); !ok {
			t.Fatalf("HTTP/2 evaluator version %d missing", version)
		}
	}

	rows[0].Status = "retired"
	rows[0].AllowedDimensions = []string{"site_country"}
	if err := validateCatalogRows(rows); err == nil || !strings.Contains(err.Error(), "allowed_dimensions") {
		t.Fatalf("retired dimension drift was not rejected: %v", err)
	}
}

func TestValidateCatalogRowsRejectsUnknownMetricInEveryStatus(t *testing.T) {
	for _, status := range []string{"active", "retired"} {
		t.Run(status, func(t *testing.T) {
			if err := validateCatalogRows([]navsqlc.GfnMetricRegistry{{
				MetricKey: "uncompiled", MetricVersion: 1, Status: status,
			}}); err == nil || !strings.Contains(err.Error(), "no compiled evaluator") {
				t.Fatalf("%s unknown metric was not rejected: %v", status, err)
			}
		})
	}
}
