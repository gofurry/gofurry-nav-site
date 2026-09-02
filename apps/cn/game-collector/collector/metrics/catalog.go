package metrics

import (
	"context"
	"fmt"
	"slices"

	gamesqlc "github.com/gofurry/gofurry-game-collector/internal/db/game/sqlc"
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
		Key: "free_game_share", Version: 1, Kind: "state_ratio", EntityLevel: "game", TimeGrain: "day",
		SourceFacts: []string{"gfg_game_daily"}, EligibilityPolicy: "tracked_game_availability_v1",
		StatePolicy: "free_game_share_state_v1", CoveragePolicy: "known_over_eligible_v1",
		FreshnessSeconds: 259200, AllowedDimensions: []string{"primary_tag_id", "tag_id"},
	},
	{
		Key: "linux_support", Version: 1, Kind: "state_ratio", EntityLevel: "game", TimeGrain: "day",
		SourceFacts: []string{"gfg_game_daily"}, EligibilityPolicy: "tracked_game_v1",
		StatePolicy: "linux_support_state_v1", CoveragePolicy: "known_over_eligible_v1",
		FreshnessSeconds: 259200, AllowedDimensions: []string{"primary_tag_id", "tag_id"},
	},
	{
		Key: "mac_support", Version: 1, Kind: "state_ratio", EntityLevel: "game", TimeGrain: "day",
		SourceFacts: []string{"gfg_game_daily"}, EligibilityPolicy: "tracked_game_v1",
		StatePolicy: "mac_support_state_v1", CoveragePolicy: "known_over_eligible_v1",
		FreshnessSeconds: 259200, AllowedDimensions: []string{"primary_tag_id", "tag_id"},
	},
	{
		Key: "windows_support", Version: 1, Kind: "state_ratio", EntityLevel: "game", TimeGrain: "day",
		SourceFacts: []string{"gfg_game_daily"}, EligibilityPolicy: "tracked_game_v1",
		StatePolicy: "windows_support_state_v1", CoveragePolicy: "known_over_eligible_v1",
		FreshnessSeconds: 259200, AllowedDimensions: []string{"primary_tag_id", "tag_id"},
	},
}

func (engine *Engine) ValidateCatalog(ctx context.Context) error {
	if engine == nil || engine.pool == nil {
		return fmt.Errorf("Game metric engine PostgreSQL pool is nil")
	}
	rows, err := gamesqlc.New(engine.pool).ListGameMetricRegistry(ctx)
	if err != nil {
		return fmt.Errorf("load Game metric registry: %w", err)
	}
	return validateCatalogRows(rows)
}

func validateCatalogRows(rows []gamesqlc.GfgMetricRegistry) error {
	compiled := make(map[string]Contract, len(evaluatorCatalog))
	for _, contract := range evaluatorCatalog {
		compiled[catalogID(contract.Key, contract.Version)] = contract
	}
	registered := make(map[string]gamesqlc.GfgMetricRegistry, len(rows))
	for _, row := range rows {
		id := catalogID(row.MetricKey, row.MetricVersion)
		registered[id] = row
		if _, ok := compiled[id]; !ok {
			return fmt.Errorf("Game metric registry %s has no compiled evaluator", id)
		}
	}
	for id, contract := range compiled {
		row, ok := registered[id]
		if !ok {
			return fmt.Errorf("compiled Game metric evaluator %s has no registry contract", id)
		}
		if err := compareContract(contract, row); err != nil {
			return fmt.Errorf("Game metric registry/evaluator drift for %s: %w", id, err)
		}
	}
	return nil
}

func compareContract(contract Contract, row gamesqlc.GfgMetricRegistry) error {
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
