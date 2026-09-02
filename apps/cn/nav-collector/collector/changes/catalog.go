package changes

import (
	"context"
	"fmt"
	"slices"

	navsqlc "github.com/gofurry/gofurry-nav-collector/internal/db/nav/sqlc"
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
	{Key: "ipv6_transition", Version: 1, SourceKind: "metric", SourceContracts: []string{"ipv6_adoption/1", "gfn_site_daily"}, DetectionPolicy: "metric_semantic_transition_v1", WatermarkPolicy: "metric_checkpoint_v1", EventCodes: []string{"ipv6_enabled", "ipv6_disabled"}, ProcessingGrain: "day"},
	{Key: "ipv6_transition", Version: 2, SourceKind: "metric", SourceContracts: []string{"ipv6_adoption/2", "gfn_site_daily"}, DetectionPolicy: "metric_semantic_transition_v2", WatermarkPolicy: "metric_checkpoint_v1", EventCodes: []string{"ipv6_enabled", "ipv6_disabled"}, ProcessingGrain: "day"},
	{Key: "tls13_transition", Version: 1, SourceKind: "metric", SourceContracts: []string{"tls13_adoption/1", "gfn_site_daily"}, DetectionPolicy: "metric_semantic_transition_v1", WatermarkPolicy: "metric_checkpoint_v1", EventCodes: []string{"tls13_enabled", "tls13_disabled"}, ProcessingGrain: "day"},
	{Key: "security_txt_transition", Version: 1, SourceKind: "metric", SourceContracts: []string{"security_txt_adoption/1", "gfn_site_daily"}, DetectionPolicy: "metric_semantic_transition_v1", WatermarkPolicy: "metric_checkpoint_v1", EventCodes: []string{"security_txt_added", "security_txt_removed"}, ProcessingGrain: "day"},
	{Key: "security_txt_transition", Version: 2, SourceKind: "metric", SourceContracts: []string{"security_txt_adoption/2", "gfn_site_daily"}, DetectionPolicy: "metric_semantic_transition_v2", WatermarkPolicy: "metric_checkpoint_v1", EventCodes: []string{"security_txt_added", "security_txt_removed"}, ProcessingGrain: "day"},
	{Key: "primary_target_transition", Version: 1, SourceKind: "effective_period", SourceContracts: []string{"gfn_site_primary_target_periods", "gfn_target_tracking_periods"}, DetectionPolicy: "primary_continuous_replacement_v1", WatermarkPolicy: "closed_day_v1", EventCodes: []string{"primary_target_changed"}, ProcessingGrain: "day"},
	{Key: "tls_certificate_transition", Version: 1, SourceKind: "fact", SourceContracts: []string{"gfn_site_target_daily"}, DetectionPolicy: "tls_certificate_semantic_memory_v1", WatermarkPolicy: "fact_checkpoint_v1", EventCodes: []string{"tls_certificate_changed"}, ProcessingGrain: "day"},
	{Key: "http2_transition", Version: 1, SourceKind: "metric", SourceContracts: []string{"http2_adoption/1", "gfn_site_daily"}, DetectionPolicy: "metric_semantic_transition_v1", WatermarkPolicy: "metric_checkpoint_v1", EventCodes: []string{"http2_enabled", "http2_disabled"}, ProcessingGrain: "day"},
	{Key: "hsts_transition", Version: 1, SourceKind: "metric", SourceContracts: []string{"hsts_adoption/1", "gfn_site_daily"}, DetectionPolicy: "metric_semantic_transition_v1", WatermarkPolicy: "metric_checkpoint_v1", EventCodes: []string{"hsts_added", "hsts_removed"}, ProcessingGrain: "day"},
	{Key: "csp_transition", Version: 1, SourceKind: "metric", SourceContracts: []string{"csp_adoption/1", "gfn_site_daily"}, DetectionPolicy: "metric_semantic_transition_v1", WatermarkPolicy: "metric_checkpoint_v1", EventCodes: []string{"csp_added", "csp_removed"}, ProcessingGrain: "day"},
	{Key: "tls_certificate_verification_transition", Version: 1, SourceKind: "metric", SourceContracts: []string{"tls_certificate_verification/1", "gfn_site_daily"}, DetectionPolicy: "metric_semantic_transition_v1", WatermarkPolicy: "metric_checkpoint_v1", EventCodes: []string{"tls_certificate_verification_failed", "tls_certificate_verification_restored"}, ProcessingGrain: "day"},
}

func (engine *Engine) ValidateCatalog(ctx context.Context) error {
	if engine == nil || engine.pool == nil {
		return fmt.Errorf("Nav change engine PostgreSQL pool is nil")
	}
	rows, err := navsqlc.New(engine.pool).ListNavChangeRegistry(ctx)
	if err != nil {
		return fmt.Errorf("load Nav change registry: %w", err)
	}
	return validateCatalogRows(rows)
}

func validateCatalogRows(rows []navsqlc.GfnChangeRegistry) error {
	compiled := make(map[string]Contract, len(detectorCatalog))
	for _, contract := range detectorCatalog {
		compiled[catalogID(contract.Key, contract.Version)] = contract
	}
	registered := make(map[string]navsqlc.GfnChangeRegistry, len(rows))
	for _, row := range rows {
		id := catalogID(row.DetectorKey, row.DetectorVersion)
		registered[id] = row
		if _, ok := compiled[id]; !ok {
			return fmt.Errorf("Nav change registry %s has no compiled detector", id)
		}
	}
	for id, contract := range compiled {
		row, ok := registered[id]
		if !ok {
			return fmt.Errorf("compiled Nav change detector %s has no registry contract", id)
		}
		if row.SourceKind != contract.SourceKind || row.DetectionPolicy != contract.DetectionPolicy || row.WatermarkPolicy != contract.WatermarkPolicy || row.ProcessingGrain != contract.ProcessingGrain {
			return fmt.Errorf("Nav change registry/detector drift for %s: scalar contract mismatch", id)
		}
		if !slices.Equal(row.SourceContracts, contract.SourceContracts) || !slices.Equal(row.EventCodes, contract.EventCodes) {
			return fmt.Errorf("Nav change registry/detector drift for %s: array contract mismatch", id)
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
