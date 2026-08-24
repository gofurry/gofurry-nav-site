package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	gamesqlc "github.com/gofurry/gofurry-game-collector/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func saveCanonicalRelease(ctx context.Context, queries *gamesqlc.Queries, current domain.GameReleaseState) error {
	previous, err := queries.GetReleaseState(ctx, current.GameID)
	firstObservation := errors.Is(err, pgx.ErrNoRows)
	if err != nil && !firstObservation {
		return fmt.Errorf("load canonical release state: %w", err)
	}

	if firstObservation {
		if err := queries.InsertReleaseState(ctx, insertReleaseStateParams(current)); err != nil {
			return fmt.Errorf("insert canonical release state: %w", err)
		}
	} else if err := queries.UpdateReleaseState(ctx, updateReleaseStateParams(current)); err != nil {
		return fmt.Errorf("update canonical release state: %w", err)
	}

	if firstObservation || !sameReleaseSemantics(previous, current) {
		if err := queries.InsertReleaseHistory(ctx, insertReleaseHistoryParams(current)); err != nil {
			return fmt.Errorf("insert canonical release history: %w", err)
		}
	}

	if current.Availability != domain.ReleaseAvailable || !calculablePrecision(current.Precision) || current.Year == nil || current.WindowStart == nil || current.WindowEnd == nil {
		return nil
	}
	source := "steam_backfill"
	inferred := true
	if !firstObservation && previous.Availability == string(domain.ReleaseUpcoming) {
		source = "observed_transition"
		inferred = false
	}
	if _, err := queries.InsertFirstAvailableIfAbsent(ctx, gamesqlc.InsertFirstAvailableIfAbsentParams{
		GameID:            current.GameID,
		Precision:         string(current.Precision),
		ExactDate:         dateValue(current.ExactDate),
		ReleaseYear:       int32(*current.Year),
		ReleaseMonth:      int32Pointer(current.Month),
		ReleaseQuarter:    int32Pointer(current.Quarter),
		WindowStart:       dateValue(current.WindowStart),
		WindowEnd:         dateValue(current.WindowEnd),
		Source:            source,
		Inferred:          inferred,
		SourceRaw:         current.RawText,
		SourceObservedAt:  timestamptz(current.ObservedAt),
		NormalizerVersion: current.Normalizer,
	}); err != nil {
		return fmt.Errorf("establish first available: %w", err)
	}
	return nil
}

func replaceCanonicalLanguages(ctx context.Context, queries *gamesqlc.Queries, gameID int64, items []domain.GameLanguage) error {
	if err := queries.DeleteGameLanguages(ctx, gameID); err != nil {
		return fmt.Errorf("delete canonical languages: %w", err)
	}
	for _, item := range items {
		if err := queries.InsertGameLanguage(ctx, gamesqlc.InsertGameLanguageParams{
			GameID:             gameID,
			LanguageCode:       item.Code,
			SteamName:          item.SteamName,
			SteamApiCode:       item.SteamAPICode,
			SteamWebCode:       item.SteamWebCode,
			Tier:               item.Tier,
			InterfaceSupported: item.InterfaceSupported,
			SubtitlesSupported: item.SubtitlesSupported,
			FullAudioSupported: item.FullAudioSupported,
			SortOrder:          int32(item.SortOrder),
			Source:             string(item.Source),
			SourceRegion:       string(item.SourceRegion),
			SourceLocale:       string(item.SourceLocale),
			NormalizerVersion:  item.Normalizer,
			ObservedAt:         timestamptz(item.ObservedAt),
		}); err != nil {
			return fmt.Errorf("insert canonical language %q: %w", item.SteamName, err)
		}
	}
	return nil
}

func insertReleaseStateParams(value domain.GameReleaseState) gamesqlc.InsertReleaseStateParams {
	return gamesqlc.InsertReleaseStateParams{
		GameID:            value.GameID,
		Availability:      string(value.Availability),
		Precision:         string(value.Precision),
		ExactDate:         dateValue(value.ExactDate),
		ReleaseYear:       int32Pointer(value.Year),
		ReleaseMonth:      int32Pointer(value.Month),
		ReleaseQuarter:    int32Pointer(value.Quarter),
		WindowStart:       dateValue(value.WindowStart),
		WindowEnd:         dateValue(value.WindowEnd),
		RawText:           value.RawText,
		Source:            string(value.Source),
		SourceRegion:      string(value.SourceRegion),
		SourceLocale:      string(value.SourceLocale),
		NormalizerVersion: value.Normalizer,
		ObservedAt:        timestamptz(value.ObservedAt),
	}
}

func updateReleaseStateParams(value domain.GameReleaseState) gamesqlc.UpdateReleaseStateParams {
	return gamesqlc.UpdateReleaseStateParams{
		GameID:            value.GameID,
		Availability:      string(value.Availability),
		Precision:         string(value.Precision),
		ExactDate:         dateValue(value.ExactDate),
		ReleaseYear:       int32Pointer(value.Year),
		ReleaseMonth:      int32Pointer(value.Month),
		ReleaseQuarter:    int32Pointer(value.Quarter),
		WindowStart:       dateValue(value.WindowStart),
		WindowEnd:         dateValue(value.WindowEnd),
		RawText:           value.RawText,
		Source:            string(value.Source),
		SourceRegion:      string(value.SourceRegion),
		SourceLocale:      string(value.SourceLocale),
		NormalizerVersion: value.Normalizer,
		ObservedAt:        timestamptz(value.ObservedAt),
	}
}

func insertReleaseHistoryParams(value domain.GameReleaseState) gamesqlc.InsertReleaseHistoryParams {
	return gamesqlc.InsertReleaseHistoryParams{
		GameID:            value.GameID,
		Availability:      string(value.Availability),
		Precision:         string(value.Precision),
		ExactDate:         dateValue(value.ExactDate),
		ReleaseYear:       int32Pointer(value.Year),
		ReleaseMonth:      int32Pointer(value.Month),
		ReleaseQuarter:    int32Pointer(value.Quarter),
		WindowStart:       dateValue(value.WindowStart),
		WindowEnd:         dateValue(value.WindowEnd),
		RawText:           value.RawText,
		Source:            string(value.Source),
		SourceRegion:      string(value.SourceRegion),
		SourceLocale:      string(value.SourceLocale),
		NormalizerVersion: value.Normalizer,
		ObservedAt:        timestamptz(value.ObservedAt),
	}
}

func sameReleaseSemantics(previous gamesqlc.GfgGameReleaseState, current domain.GameReleaseState) bool {
	return previous.Availability == string(current.Availability) &&
		previous.Precision == string(current.Precision) &&
		sameDate(previous.ExactDate, current.ExactDate) &&
		sameInt32(previous.ReleaseYear, current.Year) &&
		sameInt32(previous.ReleaseMonth, current.Month) &&
		sameInt32(previous.ReleaseQuarter, current.Quarter) &&
		sameDate(previous.WindowStart, current.WindowStart) &&
		sameDate(previous.WindowEnd, current.WindowEnd)
}

func calculablePrecision(value domain.ReleasePrecision) bool {
	switch value {
	case domain.ReleasePrecisionDay, domain.ReleasePrecisionMonth, domain.ReleasePrecisionQuarter, domain.ReleasePrecisionYear:
		return true
	default:
		return false
	}
}

func dateValue(value *time.Time) pgtype.Date {
	if value == nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC), Valid: true}
}

func int32Pointer(value *int) *int32 {
	if value == nil {
		return nil
	}
	converted := int32(*value)
	return &converted
}

func sameInt32(previous *int32, current *int) bool {
	if previous == nil || current == nil {
		return previous == nil && current == nil
	}
	return *previous == int32(*current)
}

func sameDate(previous pgtype.Date, current *time.Time) bool {
	if !previous.Valid || current == nil {
		return !previous.Valid && current == nil
	}
	return previous.Time.Year() == current.Year() && previous.Time.Month() == current.Month() && previous.Time.Day() == current.Day()
}
