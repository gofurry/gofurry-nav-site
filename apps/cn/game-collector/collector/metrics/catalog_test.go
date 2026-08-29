package metrics

import (
	"strings"
	"testing"

	gamesqlc "github.com/gofurry/gofurry-game-collector/internal/db/game/sqlc"
)

func TestValidateCatalogRows(t *testing.T) {
	rows := make([]gamesqlc.GfgMetricRegistry, 0, len(evaluatorCatalog))
	for _, contract := range evaluatorCatalog {
		freshness := contract.FreshnessSeconds
		rows = append(rows, gamesqlc.GfgMetricRegistry{
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

	rows[0].FreshnessSeconds = ptrInt64(1)
	if err := validateCatalogRows(rows); err == nil || !strings.Contains(err.Error(), "freshness") {
		t.Fatalf("freshness drift was not rejected: %v", err)
	}
}

func TestValidateCatalogRowsRejectsActiveUnknownMetric(t *testing.T) {
	if err := validateCatalogRows([]gamesqlc.GfgMetricRegistry{{
		MetricKey: "uncompiled", MetricVersion: 1, Status: "active",
	}}); err == nil || !strings.Contains(err.Error(), "no compiled evaluator") {
		t.Fatalf("active unknown metric was not rejected: %v", err)
	}
}

func ptrInt64(value int64) *int64 { return &value }
