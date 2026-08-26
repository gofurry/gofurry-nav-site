package backfill

import (
	"context"
	"errors"
	"fmt"
	"time"

	gamesqlc "github.com/gofurry/gofurry-game-collector/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const legacyNormalizerVersion = "gofurry-legacy-release/v1"

// Summary is the stable machine-readable result of a legacy backfill run.
type Summary struct {
	Scanned               int64 `json:"scanned"`
	Eligible              int64 `json:"eligible"`
	Inserted              int64 `json:"inserted"`
	AlreadyExists         int64 `json:"already_exists"`
	SkippedUpcoming       int64 `json:"skipped_upcoming"`
	SkippedNoCurrentState int64 `json:"skipped_no_current_state"`
	SkippedInvalid        int64 `json:"skipped_invalid"`
	SkippedFuture         int64 `json:"skipped_future"`
}

// Runner backfills trusted legacy manual release values into write-once facts.
type Runner struct {
	queries *gamesqlc.Queries
	now     func() time.Time
}

// New creates a backfill runner for the Game database.
func New(pool *pgxpool.Pool) *Runner {
	return &Runner{queries: gamesqlc.New(pool), now: time.Now}
}

// Run scans legacy candidates. A dry run performs no writes.
func (runner *Runner) Run(ctx context.Context, dryRun bool) (Summary, error) {
	if runner == nil || runner.queries == nil {
		return Summary{}, errors.New("first available backfill database is nil")
	}
	rows, err := runner.queries.ListLegacyFirstAvailableCandidates(ctx)
	if err != nil {
		return Summary{}, fmt.Errorf("list legacy first available candidates: %w", err)
	}
	today := runner.now()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	summary := Summary{Scanned: int64(len(rows))}

	for _, row := range rows {
		if !row.HasCurrentState {
			summary.SkippedNoCurrentState++
			continue
		}
		if row.ReleaseComingSoon {
			summary.SkippedUpcoming++
			continue
		}
		parsed, ok := parseLegacyRelease(row.ReleaseDate)
		if !ok {
			summary.SkippedInvalid++
			continue
		}
		if parsed.WindowStart.After(today) {
			summary.SkippedFuture++
			continue
		}
		summary.Eligible++

		if _, err := runner.queries.GetFirstAvailable(ctx, row.GameID); err == nil {
			summary.AlreadyExists++
			continue
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return summary, fmt.Errorf("load first available game_id=%d: %w", row.GameID, err)
		}
		if dryRun {
			continue
		}

		inserted, err := runner.queries.InsertFirstAvailableIfAbsent(ctx, gamesqlc.InsertFirstAvailableIfAbsentParams{
			GameID:            row.GameID,
			Precision:         parsed.Precision,
			ExactDate:         backfillDate(parsed.ExactDate),
			ReleaseYear:       int32(parsed.Year),
			ReleaseMonth:      backfillInt32(parsed.Month),
			ReleaseQuarter:    backfillInt32(parsed.Quarter),
			WindowStart:       backfillDate(&parsed.WindowStart),
			WindowEnd:         backfillDate(&parsed.WindowEnd),
			Source:            "legacy_manual",
			Inferred:          false,
			SourceRaw:         row.ReleaseDate,
			SourceObservedAt:  row.SourceObservedAt,
			NormalizerVersion: legacyNormalizerVersion,
		})
		if err != nil {
			return summary, fmt.Errorf("insert first available game_id=%d: %w", row.GameID, err)
		}
		if inserted == 0 {
			summary.AlreadyExists++
		} else {
			summary.Inserted++
		}
	}
	return summary, nil
}

func backfillDate(value *time.Time) pgtype.Date {
	if value == nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: *value, Valid: true}
}

func backfillInt32(value *int) *int32 {
	if value == nil {
		return nil
	}
	converted := int32(*value)
	return &converted
}
