package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	gamesqlc "github.com/gofurry/gofurry-game-backend/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (dao *ReadModelDAO) loadCanonicalGameDomain(ctx context.Context, aggregate *v2models.GameV2Aggregate) error {
	gameID := aggregate.Site.ID
	release, err := dao.q.GetReleaseStateByGame(ctx, gameID)
	if err == nil {
		mapped := releaseStateView(release)
		aggregate.ReleaseState = &mapped
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("查询 canonical release state 失败: %w", err)
	}

	firstAvailable, err := dao.q.GetFirstAvailableByGame(ctx, gameID)
	if err == nil {
		mapped := firstAvailableView(firstAvailable)
		aggregate.FirstAvailable = &mapped
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("查询 canonical first available 失败: %w", err)
	}

	languages, err := dao.q.ListLanguagesByGame(ctx, gameID)
	if err != nil {
		return fmt.Errorf("查询 canonical languages 失败: %w", err)
	}
	aggregate.Languages = make([]v2models.GameV2Language, 0, len(languages))
	for _, language := range languages {
		aggregate.Languages = append(aggregate.Languages, languageView(language))
	}
	return nil
}

func (dao *ReadModelDAO) loadCanonicalGameDomainBatch(ctx context.Context, aggregateMap map[int64]*v2models.GameV2Aggregate, gameIDs []int64) error {
	releases, err := dao.q.BatchReleaseStatesByGames(ctx, gameIDs)
	if err != nil {
		return fmt.Errorf("批量查询 canonical release state 失败: %w", err)
	}
	for _, release := range releases {
		mapped := releaseStateView(release)
		aggregateMap[release.GameID].ReleaseState = &mapped
	}
	firstAvailable, err := dao.q.BatchFirstAvailableByGames(ctx, gameIDs)
	if err != nil {
		return fmt.Errorf("批量查询 canonical first available 失败: %w", err)
	}
	for _, fact := range firstAvailable {
		mapped := firstAvailableView(fact)
		aggregateMap[fact.GameID].FirstAvailable = &mapped
	}
	return nil
}

func releaseStateView(row gamesqlc.GfgGameReleaseState) v2models.GameV2ReleaseState {
	return v2models.GameV2ReleaseState{
		Availability: row.Availability,
		Precision:    row.Precision,
		ExactDate:    calendarDate(row.ExactDate),
		Year:         row.ReleaseYear,
		Month:        row.ReleaseMonth,
		Quarter:      row.ReleaseQuarter,
		WindowStart:  calendarDate(row.WindowStart),
		WindowEnd:    calendarDate(row.WindowEnd),
		RawText:      row.RawText,
		ObservedAt:   timestampPointer(row.ObservedAt),
		UpdatedAt:    timestampValue(row.UpdatedAt),
	}
}

func firstAvailableView(row gamesqlc.GfgGameFirstAvailable) v2models.GameV2FirstAvailable {
	return v2models.GameV2FirstAvailable{
		Precision:   row.Precision,
		ExactDate:   calendarDate(row.ExactDate),
		Year:        row.ReleaseYear,
		Month:       row.ReleaseMonth,
		Quarter:     row.ReleaseQuarter,
		WindowStart: calendarDateValue(row.WindowStart),
		WindowEnd:   calendarDateValue(row.WindowEnd),
		Source:      row.Source,
		Inferred:    row.Inferred,
		UpdatedAt:   timestampValue(row.UpdatedAt),
	}
}

func languageView(row gamesqlc.GfgGameLanguage) v2models.GameV2Language {
	return v2models.GameV2Language{
		Code:               row.LanguageCode,
		SteamName:          row.SteamName,
		SteamAPICode:       row.SteamApiCode,
		SteamWebCode:       row.SteamWebCode,
		Tier:               row.Tier,
		InterfaceSupported: row.InterfaceSupported,
		SubtitlesSupported: row.SubtitlesSupported,
		FullAudioSupported: row.FullAudioSupported,
	}
}

func attachSearchFirstAvailable(items []v2models.GameV2SearchPageItem) {
	for index := range items {
		item := &items[index]
		if item.FAPrecision == nil || item.FAReleaseYear == nil || item.FAWindowStart == nil || item.FAWindowEnd == nil || item.FASource == nil || item.FAInferred == nil {
			continue
		}
		item.FirstAvailable = &v2models.GameV2FirstAvailable{
			Precision:   *item.FAPrecision,
			ExactDate:   calendarTime(item.FAExactDate),
			Year:        *item.FAReleaseYear,
			Month:       item.FAReleaseMonth,
			Quarter:     item.FAReleaseQuarter,
			WindowStart: calendarTimeValue(item.FAWindowStart),
			WindowEnd:   calendarTimeValue(item.FAWindowEnd),
			Source:      *item.FASource,
			Inferred:    *item.FAInferred,
		}
	}
}

func calendarDate(value pgtype.Date) *string {
	if !value.Valid {
		return nil
	}
	formatted := value.Time.Format(time.DateOnly)
	return &formatted
}

func calendarDateValue(value pgtype.Date) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format(time.DateOnly)
}

func calendarTime(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format(time.DateOnly)
	return &formatted
}

func calendarTimeValue(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format(time.DateOnly)
}

func timestampPointer(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}

func timestampValue(value pgtype.Timestamptz) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time
}
