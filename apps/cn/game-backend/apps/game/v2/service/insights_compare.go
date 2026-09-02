package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
)

func (s *InsightsService) GetGameCompare(ctx context.Context, rawIDs, region string) (v2models.GameCompare, error) {
	result := v2models.GameCompare{Status: "insufficient_data", Region: region, Games: []v2models.GameCompareItem{}}
	if !validInsightRegion(region) {
		return result, ErrInvalidInsightRegion
	}
	gameIDs, err := parseGameCompareIDs(rawIDs)
	if err != nil {
		return result, err
	}
	for _, gameID := range gameIDs {
		game, queryErr := s.requireGame(ctx, gameID)
		if queryErr != nil {
			return result, queryErr
		}
		result.Games = append(result.Games, v2models.GameCompareItem{
			Game:      gameCompareEntity(*game),
			Price:     v2models.InsightRegionalPrice{Region: region},
			Languages: v2models.GameCompareLanguages{Supported: []string{}, ExplicitFullAudio: []string{}, UnknownNames: []string{}},
		})
	}

	horizon, err := s.store.GetInsightGameCompareFactHorizon(ctx, gameIDs)
	if err != nil {
		return result, err
	}
	if horizon == nil {
		return result, nil
	}
	asOf := insightFormatDate(*horizon)
	result.StateAsOf = &asOf

	facts, err := s.store.ListInsightGameCompareFacts(ctx, gameIDs, *horizon, region)
	if err != nil {
		return result, err
	}
	factByGame := make(map[int64]v2models.InsightGameCompareFactRecord, len(facts))
	for _, fact := range facts {
		factByGame[fact.GameID] = fact
	}
	if len(factByGame) != len(gameIDs) {
		return result, fmt.Errorf("incomplete game comparison snapshot")
	}

	currentPlayers, err := s.store.ListInsightGameCompareCurrentPlayers(ctx, gameIDs)
	if err != nil {
		return result, err
	}
	currentByGame := make(map[int64]v2models.InsightGameCompareCurrentPlayerRecord, len(currentPlayers))
	for _, player := range currentPlayers {
		currentByGame[player.GameID] = player
		if result.PlayerSnapshotScheduledFor == nil && player.SnapshotScheduledFor != nil {
			result.PlayerSnapshotScheduledFor = player.SnapshotScheduledFor
		}
	}

	player30D, err := s.store.ListInsightGameComparePlayer30D(ctx, gameIDs)
	if err != nil {
		return result, err
	}
	player30DByGame := make(map[int64]v2models.InsightGameComparePlayer30DRecord, len(player30D))
	for _, player := range player30D {
		player30DByGame[player.GameID] = player
		if result.PlayerFactThrough == nil {
			result.PlayerFactThrough = insightDateStringPointer(player.FactThrough)
		}
	}

	for index := range result.Games {
		item := &result.Games[index]
		fact := factByGame[item.Game.ID]
		item.State = v2models.GameCompareState{
			Free: fact.Free, Windows: fact.Windows, Mac: fact.Mac, Linux: fact.Linux, Release: fact.Release,
		}
		item.Languages = v2models.GameCompareLanguages{
			Evidence: fact.LanguageEvidence, Supported: fact.LanguageCodes,
			ExplicitFullAudio: fact.FullAudioLanguageCodes, UnknownNames: fact.UnknownLanguageNames,
		}
		item.Price = v2models.InsightRegionalPrice{
			Region: region, Available: fact.PriceAvailable, State: fact.PriceState, Currency: fact.Currency,
			InitialAmount: fact.InitialAmount, FinalAmount: fact.FinalAmount, DiscountPercent: fact.DiscountPercent,
		}
		if fact.PriceAvailable && fact.PriceState != nil && *fact.PriceState == "priced" && fact.Currency != nil {
			low, queryErr := s.store.GetInsightObservedLow(ctx, fact.TrackingPeriodID, region, *horizon)
			if queryErr != nil {
				return result, queryErr
			}
			item.Price.ObservedLow = publicObservedLow(low)
		}

		if current, ok := currentByGame[item.Game.ID]; ok {
			item.Players.CurrentAvailable = current.Available
			item.Players.ObservedAt = current.CollectedAt
			if current.Available {
				value := current.PlayerCount
				item.Players.Current = &value
			}
		}
		if summary, ok := player30DByGame[item.Game.ID]; ok {
			item.Players.Peak30D = summary.Peak30D
			item.Players.Average30D = summary.Average30D
			item.Players.EligibleFrom30D = insightDateStringPointer(summary.EligibleFrom)
			item.Players.ObservedDays30D = summary.ObservedDays
			item.Players.SuccessfulSamples30D = summary.SuccessfulSamples
			item.Players.SampleCoverage30D = summary.SampleCoverage
		}
	}

	result.Status = "ready"
	return result, nil
}

func parseGameCompareIDs(raw string) ([]int64, error) {
	parts := strings.Split(raw, ",")
	result := make([]int64, 0, len(parts))
	seen := map[int64]struct{}{}
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			return nil, ErrInvalidInsightCompare
		}
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil || id <= 0 {
			return nil, ErrInvalidInsightCompare
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	if len(result) < 2 || len(result) > 4 {
		return nil, ErrInvalidInsightCompare
	}
	return result, nil
}

func gameCompareEntity(game v2models.InsightGameRecord) v2models.InsightEntityRef {
	return v2models.InsightEntityRef{ID: game.ID, Name: game.Name}
}
