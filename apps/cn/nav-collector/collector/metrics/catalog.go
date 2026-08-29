package metrics

import (
	"context"
	"fmt"
	"slices"

	navsqlc "github.com/gofurry/gofurry-nav-collector/internal/db/nav/sqlc"
)

type Contract struct {
	Key               string
	Version           int32
	Kind              string
	EntityLevel       string
	TimeGrain         string
	SourceFacts       []string
	EligibilityPolicy string
	StatePolicy       string
	CoveragePolicy    string
	FreshnessSeconds  int64
	AllowedDimensions []string
}

var evaluatorCatalog = []Contract{
	{
		Key: "ipv6_adoption", Version: 1, Kind: "state_ratio", EntityLevel: "site", TimeGrain: "day",
		SourceFacts:       []string{"gfn_site_daily", "gfn_site_target_daily", "gfn_site_target_protocol_daily"},
		EligibilityPolicy: "active_site_primary_target_v1", StatePolicy: "ipv6_adoption_state_v1",
		CoveragePolicy: "known_over_eligible_v1", FreshnessSeconds: 259200,
		AllowedDimensions: []string{"group_id", "nsfw", "site_country", "welfare"},
	},
	{
		Key: "security_txt_adoption", Version: 1, Kind: "state_ratio", EntityLevel: "site", TimeGrain: "day",
		SourceFacts:       []string{"gfn_site_daily", "gfn_site_target_daily", "gfn_site_target_protocol_daily"},
		EligibilityPolicy: "active_site_primary_target_v1", StatePolicy: "security_txt_adoption_state_v1",
		CoveragePolicy: "known_over_eligible_v1", FreshnessSeconds: 1814400,
		AllowedDimensions: []string{"group_id", "nsfw", "site_country", "welfare"},
	},
	{
		Key: "tls13_adoption", Version: 1, Kind: "state_ratio", EntityLevel: "site", TimeGrain: "day",
		SourceFacts:       []string{"gfn_site_daily", "gfn_site_target_daily", "gfn_site_target_protocol_daily"},
		EligibilityPolicy: "active_site_primary_target_v1", StatePolicy: "tls13_adoption_state_v1",
		CoveragePolicy: "known_over_eligible_v1", FreshnessSeconds: 172800,
		AllowedDimensions: []string{"group_id", "nsfw", "site_country", "welfare"},
	},
}

func (engine *Engine) ValidateCatalog(ctx context.Context) error {
	if engine == nil || engine.pool == nil {
		return fmt.Errorf("Nav metric engine PostgreSQL pool is nil")
	}
	rows, err := navsqlc.New(engine.pool).ListNavMetricRegistry(ctx)
	if err != nil {
		return fmt.Errorf("load Nav metric registry: %w", err)
	}
	return validateCatalogRows(rows)
}

func validateCatalogRows(rows []navsqlc.GfnMetricRegistry) error {
	compiled := make(map[string]Contract, len(evaluatorCatalog))
	for _, contract := range evaluatorCatalog {
		compiled[catalogID(contract.Key, contract.Version)] = contract
	}
	registered := make(map[string]navsqlc.GfnMetricRegistry, len(rows))
	for _, row := range rows {
		id := catalogID(row.MetricKey, row.MetricVersion)
		registered[id] = row
		if row.Status == "active" {
			if _, ok := compiled[id]; !ok {
				return fmt.Errorf("active Nav metric registry %s has no compiled evaluator", id)
			}
		}
	}
	for id, contract := range compiled {
		row, ok := registered[id]
		if !ok {
			return fmt.Errorf("compiled Nav metric evaluator %s has no registry contract", id)
		}
		if err := compareContract(contract, row); err != nil {
			return fmt.Errorf("Nav metric registry/evaluator drift for %s: %w", id, err)
		}
	}
	return nil
}

func compareContract(contract Contract, row navsqlc.GfnMetricRegistry) error {
	if row.MetricKind != contract.Kind || row.EntityLevel != contract.EntityLevel || row.TimeGrain != contract.TimeGrain {
		return fmt.Errorf("kind/entity/grain mismatch")
	}
	if row.EligibilityPolicy != contract.EligibilityPolicy || row.StatePolicy != contract.StatePolicy || row.CoveragePolicy != contract.CoveragePolicy {
		return fmt.Errorf("policy identifier mismatch")
	}
	if row.FreshnessSeconds == nil || *row.FreshnessSeconds != contract.FreshnessSeconds {
		return fmt.Errorf("freshness_seconds mismatch")
	}
	if !slices.Equal(row.SourceFacts, contract.SourceFacts) {
		return fmt.Errorf("source_facts mismatch")
	}
	if !slices.Equal(row.AllowedDimensions, contract.AllowedDimensions) {
		return fmt.Errorf("allowed_dimensions mismatch")
	}
	return nil
}

func contractFor(key string, version int32) (Contract, bool) {
	for _, contract := range evaluatorCatalog {
		if contract.Key == key && contract.Version == version {
			return contract, true
		}
	}
	return Contract{}, false
}

func catalogID(key string, version int32) string { return fmt.Sprintf("%s/%d", key, version) }
