package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"time"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
)

const insightExplorerCursorVersion = 1

type insightExplorerCursor struct {
	Version      int                           `json:"version"`
	Range        string                        `json:"range"`
	Category     string                        `json:"category"`
	Type         string                        `json:"type"`
	RangeThrough string                        `json:"range_through"`
	Position     insightExplorerCursorPosition `json:"position"`
}

type insightExplorerCursorPosition struct {
	Date          string `json:"date"`
	PrecisionRank int32  `json:"precision_rank"`
	SortAt        string `json:"sort_at"`
	Tie           string `json:"tie"`
}

func (s *InsightsService) GetInsightsChanges(ctx context.Context, query v2models.InsightChangeExplorerQuery) (v2models.InsightChangeExplorerPage, error) {
	result := v2models.InsightChangeExplorerPage{Items: []v2models.InsightExplorerChange{}}
	if query.Range == "" {
		query.Range = "30d"
	}
	rangeDays, ok := parseInsightExplorerRange(query.Range)
	if !ok {
		return result, ErrInvalidInsightChanges
	}
	if query.Limit == 0 {
		query.Limit = 20
	}
	if query.Limit < 1 || query.Limit > 50 {
		return result, ErrInvalidInsightChanges
	}
	contracts, ok := insightExplorerContracts(query.Category, query.Type)
	if !ok {
		return result, ErrInvalidInsightChanges
	}
	rangeThrough := insightUTCDate(s.now())
	var position *v2models.InsightChangeExplorerPosition
	if query.Cursor != "" {
		cursor, err := decodeInsightExplorerCursor(query.Cursor)
		if err != nil || cursor.Version != insightExplorerCursorVersion || cursor.Range != query.Range ||
			cursor.Category != query.Category || cursor.Type != query.Type {
			return result, ErrInvalidInsightCursor
		}
		through, err := time.Parse("2006-01-02", cursor.RangeThrough)
		if err != nil {
			return result, ErrInvalidInsightCursor
		}
		date, err := time.Parse("2006-01-02", cursor.Position.Date)
		if err != nil || cursor.Position.PrecisionRank < 0 || cursor.Position.PrecisionRank > 1 || !validInsightOpaqueTie(cursor.Position.Tie) {
			return result, ErrInvalidInsightCursor
		}
		sortAt, err := time.Parse(time.RFC3339Nano, cursor.Position.SortAt)
		if err != nil || date.After(through) {
			return result, ErrInvalidInsightCursor
		}
		rangeThrough = through
		position = &v2models.InsightChangeExplorerPosition{
			ProjectionDate: date, PrecisionRank: cursor.Position.PrecisionRank,
			EventSortAt: sortAt, OpaqueTie: cursor.Position.Tie,
		}
	}
	detectors, contractIDs := insightExplorerQueryContracts(contracts)
	rows, err := s.store.ListInsightExplorerChanges(ctx, v2models.InsightChangeExplorerConditions{
		DetectorKeys: detectors, ContractIDs: contractIDs, RangeThrough: rangeThrough,
		RangeDays: rangeDays, Position: position, Limit: query.Limit + 1,
	})
	if err != nil {
		return result, err
	}
	hasMore := len(rows) > int(query.Limit)
	if hasMore {
		rows = rows[:query.Limit]
	}
	for _, row := range rows {
		contract, found := insightExplorerContractForRecord(row)
		if !found {
			return result, fmt.Errorf("map explorer change contract")
		}
		var occurredAt *time.Time
		if row.TimeBasis != "day" {
			occurredAt = row.EventAt
		}
		result.Items = append(result.Items, v2models.InsightExplorerChange{
			Domain: "game", Category: contract.category, Type: contract.public,
			Date: insightFormatDate(row.ProjectionDate), OccurredAt: occurredAt,
			Entity: v2models.InsightEntityRef{ID: row.EntityID, Name: row.EntityName}, Detail: nil,
		})
	}
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		cursor, err := encodeInsightExplorerCursor(insightExplorerCursor{
			Version: insightExplorerCursorVersion, Range: query.Range, Category: query.Category, Type: query.Type,
			RangeThrough: insightFormatDate(rangeThrough),
			Position: insightExplorerCursorPosition{
				Date: insightFormatDate(last.ProjectionDate), PrecisionRank: last.PrecisionRank,
				SortAt: last.EventSortAt.UTC().Format(time.RFC3339Nano), Tie: last.OpaqueTie,
			},
		})
		if err != nil {
			return result, err
		}
		result.NextCursor = &cursor
	}
	return result, nil
}

func parseInsightExplorerRange(value string) (int32, bool) {
	switch value {
	case "7d":
		return 7, true
	case "30d":
		return 30, true
	case "90d":
		return 90, true
	case "all":
		return 0, true
	default:
		return 0, false
	}
}

func insightExplorerContracts(category, publicType string) ([]insightChangeContract, bool) {
	validCategory := category == ""
	validType := publicType == ""
	result := []insightChangeContract{}
	for _, contract := range insightChangeContracts {
		if contract.category == category {
			validCategory = true
		}
		if contract.public == publicType {
			validType = true
		}
		if (category == "" || contract.category == category) && (publicType == "" || contract.public == publicType) {
			result = append(result, contract)
		}
	}
	return result, validCategory && validType && len(result) > 0
}

func insightExplorerQueryContracts(contracts []insightChangeContract) ([]string, []string) {
	detectorSet := map[string]struct{}{}
	detectors := []string{}
	ids := make([]string, 0, len(contracts))
	for _, contract := range contracts {
		if _, found := detectorSet[contract.detector]; !found {
			detectorSet[contract.detector] = struct{}{}
			detectors = append(detectors, contract.detector)
		}
		ids = append(ids, fmt.Sprintf("%s/%d/%s", contract.detector, contract.version, contract.code))
	}
	return detectors, ids
}

func insightExplorerContractForRecord(record v2models.InsightChangeRecord) (insightChangeContract, bool) {
	for _, contract := range insightChangeContracts {
		if contract.detector == record.DetectorKey && contract.version == record.DetectorVersion && contract.code == record.EventCode {
			return contract, true
		}
	}
	return insightChangeContract{}, false
}

func encodeInsightExplorerCursor(cursor insightExplorerCursor) (string, error) {
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeInsightExplorerCursor(value string) (insightExplorerCursor, error) {
	var result insightExplorerCursor
	payload, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return result, err
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		return result, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return result, ErrInvalidInsightCursor
	}
	return result, nil
}

func insightUTCDate(value time.Time) time.Time {
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func validInsightOpaqueTie(value string) bool {
	if len(value) != 32 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
