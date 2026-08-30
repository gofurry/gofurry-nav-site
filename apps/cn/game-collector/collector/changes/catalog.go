package changes

import (
	"context"
	"fmt"
	"slices"

	gamesqlc "github.com/gofurry/gofurry-game-collector/internal/db/game/sqlc"
)

type Contract struct {
	Key             string
	Version         int32
	SourceKind      string
	SourceContracts []string
	DetectionPolicy string
	WatermarkPolicy string
	EventCodes      []string
	ProcessingGrain string
}

var detectorCatalog = []Contract{
	{Key: "free_game_transition", Version: 1, SourceKind: "metric", SourceContracts: []string{"free_game_share/1", "gfg_game_daily"}, DetectionPolicy: "metric_semantic_transition_v1", WatermarkPolicy: "metric_checkpoint_v1", EventCodes: []string{"game_became_free", "game_became_paid"}, ProcessingGrain: "day"},
	{Key: "windows_support_transition", Version: 1, SourceKind: "metric", SourceContracts: []string{"windows_support/1", "gfg_game_daily"}, DetectionPolicy: "metric_semantic_transition_v1", WatermarkPolicy: "metric_checkpoint_v1", EventCodes: []string{"windows_support_added", "windows_support_removed"}, ProcessingGrain: "day"},
	{Key: "linux_support_transition", Version: 1, SourceKind: "metric", SourceContracts: []string{"linux_support/1", "gfg_game_daily"}, DetectionPolicy: "metric_semantic_transition_v1", WatermarkPolicy: "metric_checkpoint_v1", EventCodes: []string{"linux_support_added", "linux_support_removed"}, ProcessingGrain: "day"},
	{Key: "game_release_transition", Version: 1, SourceKind: "domain_history", SourceContracts: []string{"gfg_game_release_history", "gfg_game_tracking_periods"}, DetectionPolicy: "release_adjacent_history_v1", WatermarkPolicy: "closed_day_v1", EventCodes: []string{"game_became_available", "game_availability_withdrawn", "game_release_plan_changed"}, ProcessingGrain: "day"},
	{Key: "game_price_transition", Version: 1, SourceKind: "fact", SourceContracts: []string{"gfg_game_price_daily"}, DetectionPolicy: "price_semantic_memory_v1", WatermarkPolicy: "fact_checkpoint_v1", EventCodes: []string{"game_price_state_changed", "game_price_currency_changed", "game_price_decreased", "game_price_increased", "game_discount_started", "game_discount_ended", "game_discount_changed"}, ProcessingGrain: "day"},
}

func (engine *Engine) ValidateCatalog(ctx context.Context) error {
	if engine == nil || engine.pool == nil {
		return fmt.Errorf("Game change engine PostgreSQL pool is nil")
	}
	rows, err := gamesqlc.New(engine.pool).ListGameChangeRegistry(ctx)
	if err != nil {
		return fmt.Errorf("load Game change registry: %w", err)
	}
	return validateCatalogRows(rows)
}

func validateCatalogRows(rows []gamesqlc.GfgChangeRegistry) error {
	compiled := make(map[string]Contract, len(detectorCatalog))
	for _, contract := range detectorCatalog {
		compiled[catalogID(contract.Key, contract.Version)] = contract
	}
	registered := make(map[string]gamesqlc.GfgChangeRegistry, len(rows))
	for _, row := range rows {
		id := catalogID(row.DetectorKey, row.DetectorVersion)
		registered[id] = row
		if _, ok := compiled[id]; !ok {
			return fmt.Errorf("Game change registry %s has no compiled detector", id)
		}
	}
	for id, contract := range compiled {
		row, ok := registered[id]
		if !ok {
			return fmt.Errorf("compiled Game change detector %s has no registry contract", id)
		}
		if row.SourceKind != contract.SourceKind || row.DetectionPolicy != contract.DetectionPolicy || row.WatermarkPolicy != contract.WatermarkPolicy || row.ProcessingGrain != contract.ProcessingGrain {
			return fmt.Errorf("Game change registry/detector drift for %s: scalar contract mismatch", id)
		}
		if !slices.Equal(row.SourceContracts, contract.SourceContracts) || !slices.Equal(row.EventCodes, contract.EventCodes) {
			return fmt.Errorf("Game change registry/detector drift for %s: array contract mismatch", id)
		}
	}
	return nil
}

func contractFor(key string, version int32) (Contract, bool) {
	for _, contract := range detectorCatalog {
		if contract.Key == key && contract.Version == version {
			return contract, true
		}
	}
	return Contract{}, false
}

func catalogID(key string, version int32) string { return fmt.Sprintf("%s/%d", key, version) }
