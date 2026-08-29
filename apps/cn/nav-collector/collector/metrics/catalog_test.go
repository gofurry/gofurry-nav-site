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
