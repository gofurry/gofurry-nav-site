package changes

import (
	"errors"
	"strings"
	"testing"
	"time"

	gamesqlc "github.com/gofurry/gofurry-game-collector/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestCatalogRejectsUncompiledRetiredDetector(t *testing.T) {
	rows := make([]gamesqlc.GfgChangeRegistry, 0, len(detectorCatalog)+1)
	for _, contract := range detectorCatalog {
		rows = append(rows, gamesqlc.GfgChangeRegistry{DetectorKey: contract.Key, DetectorVersion: contract.Version,
			SourceKind: contract.SourceKind, SourceContracts: contract.SourceContracts, DetectionPolicy: contract.DetectionPolicy,
			WatermarkPolicy: contract.WatermarkPolicy, EventCodes: contract.EventCodes, ProcessingGrain: contract.ProcessingGrain, Status: "active"})
	}
	rows = append(rows, gamesqlc.GfgChangeRegistry{DetectorKey: "retired_without_code", DetectorVersion: 1, Status: "retired"})
	if err := validateCatalogRows(rows); err == nil || !strings.Contains(err.Error(), "no compiled detector") {
		t.Fatalf("expected retired detector drift, got %v", err)
	}
}

func TestReconcileContinuesAfterDetectorFailure(t *testing.T) {
	detectors := []registeredDetector{{Contract: detectorCatalog[0]}, {Contract: detectorCatalog[1]}, {Contract: detectorCatalog[2]}}
	called := make([]string, 0, len(detectors))
	err := reconcileDetectors(detectors, func(contract Contract) (DayResult, error) {
		called = append(called, contract.Key)
		if len(called) == 1 {
			return DayResult{}, errors.New("fixture failure")
		}
		return DayResult{DetectorKey: contract.Key, DetectorVersion: contract.Version}, nil
	})
	if err == nil || len(called) != len(detectors) {
		t.Fatalf("err=%v called=%v", err, called)
	}
}

func TestRebuildBoundsRequireForwardPropagation(t *testing.T) {
	start := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	processed := start.AddDate(0, 0, 2)
	sourceDate := pgtype.Date{Time: start, Valid: true}
	processedDate := pgtype.Date{Time: processed, Valid: true}
	if end, err := validateRebuildBounds(sourceDate, processedDate, start, nil, 0); err != nil || !end.Equal(processed) {
		t.Fatalf("full forward range end=%s err=%v", end, err)
	}
	truncated := start.AddDate(0, 0, 1)
	if _, err := validateRebuildBounds(sourceDate, processedDate, start, &truncated, 0); err == nil || !strings.Contains(err.Error(), "must equal processed_through") {
		t.Fatalf("expected truncated --through rejection, got %v", err)
	}
	if _, err := validateRebuildBounds(sourceDate, processedDate, start, nil, 2); err == nil || !strings.Contains(err.Error(), "cannot truncate") {
		t.Fatalf("expected max-days rejection, got %v", err)
	}
}

func TestCatalogRejectsContractDrift(t *testing.T) {
	rows := make([]gamesqlc.GfgChangeRegistry, 0, len(detectorCatalog))
	for _, contract := range detectorCatalog {
		rows = append(rows, gamesqlc.GfgChangeRegistry{DetectorKey: contract.Key, DetectorVersion: contract.Version,
			SourceKind: contract.SourceKind, SourceContracts: contract.SourceContracts, DetectionPolicy: contract.DetectionPolicy,
			WatermarkPolicy: contract.WatermarkPolicy, EventCodes: contract.EventCodes, ProcessingGrain: contract.ProcessingGrain, Status: "active"})
	}
	rows[0].EventCodes = []string{"wrong"}
	if err := validateCatalogRows(rows); err == nil || !strings.Contains(err.Error(), "array contract mismatch") {
		t.Fatalf("expected contract drift, got %v", err)
	}
}
