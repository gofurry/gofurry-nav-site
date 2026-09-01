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

	"github.com/gofurry/gofurry-nav-backend/apps/nav/insights/models"
)

const explorerCursorVersion = 1

type explorerCursor struct {
	Version      int                    `json:"version"`
	Range        string                 `json:"range"`
	Category     string                 `json:"category"`
	Type         string                 `json:"type"`
	RangeThrough string                 `json:"range_through"`
	Position     explorerCursorPosition `json:"position"`
}

type explorerCursorPosition struct {
	Date          string `json:"date"`
	PrecisionRank int32  `json:"precision_rank"`
	SortAt        string `json:"sort_at"`
	Tie           string `json:"tie"`
}

func (s *InsightsService) GetChanges(ctx context.Context, query models.ChangeExplorerQuery) (models.ChangeExplorerPage, error) {
	result := models.ChangeExplorerPage{Items: []models.ExplorerChange{}}
	if query.Range == "" {
		query.Range = "30d"
	}
	rangeDays, ok := parseExplorerRange(query.Range)
	if !ok {
		return result, ErrInvalidChanges
	}
	if query.Limit == 0 {
		query.Limit = 20
	}
	if query.Limit < 1 || query.Limit > 50 {
		return result, ErrInvalidChanges
	}
	contracts, ok := explorerContracts(query.Category, query.Type)
	if !ok {
		return result, ErrInvalidChanges
	}
	rangeThrough := utcDate(s.now())
	var position *models.ChangeExplorerPosition
	if query.Cursor != "" {
		cursor, err := decodeExplorerCursor(query.Cursor)
		if err != nil || cursor.Version != explorerCursorVersion || cursor.Range != query.Range ||
			cursor.Category != query.Category || cursor.Type != query.Type {
			return result, ErrInvalidCursor
		}
		through, err := time.Parse("2006-01-02", cursor.RangeThrough)
		if err != nil {
			return result, ErrInvalidCursor
		}
		date, err := time.Parse("2006-01-02", cursor.Position.Date)
		if err != nil || cursor.Position.PrecisionRank < 0 || cursor.Position.PrecisionRank > 1 || !validOpaqueTie(cursor.Position.Tie) {
			return result, ErrInvalidCursor
		}
		sortAt, err := time.Parse(time.RFC3339Nano, cursor.Position.SortAt)
		if err != nil || date.After(through) {
			return result, ErrInvalidCursor
		}
		rangeThrough = through
		position = &models.ChangeExplorerPosition{
			ProjectionDate: date, PrecisionRank: cursor.Position.PrecisionRank,
			EventSortAt: sortAt, OpaqueTie: cursor.Position.Tie,
		}
	}
	detectors, contractIDs := explorerQueryContracts(contracts)
	rows, err := s.store.ListExplorerChanges(ctx, models.ChangeExplorerConditions{
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
		contract, found := explorerContractForRecord(row)
		if !found {
			return result, fmt.Errorf("map explorer change contract")
		}
		var occurredAt *time.Time
		if row.TimeBasis != "day" {
			occurredAt = row.EventAt
		}
		result.Items = append(result.Items, models.ExplorerChange{
			Domain: "site", Category: contract.category, Type: contract.public,
			Date: formatDate(row.ProjectionDate), OccurredAt: occurredAt,
			Entity: models.EntityRef{ID: row.EntityID, Name: row.EntityName}, Detail: nil,
		})
	}
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		cursor, err := encodeExplorerCursor(explorerCursor{
			Version: explorerCursorVersion, Range: query.Range, Category: query.Category, Type: query.Type,
			RangeThrough: formatDate(rangeThrough),
			Position: explorerCursorPosition{
				Date: formatDate(last.ProjectionDate), PrecisionRank: last.PrecisionRank,
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

func parseExplorerRange(value string) (int32, bool) {
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

func explorerContracts(category, publicType string) ([]changeContract, bool) {
	validCategory := category == ""
	validType := publicType == ""
	result := []changeContract{}
	for _, contract := range changeContracts {
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

func explorerQueryContracts(contracts []changeContract) ([]string, []string) {
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

func explorerContractForRecord(record models.ChangeRecord) (changeContract, bool) {
	for _, contract := range changeContracts {
		if contract.detector == record.DetectorKey && contract.version == record.DetectorVersion && contract.code == record.EventCode {
			return contract, true
		}
	}
	return changeContract{}, false
}

func encodeExplorerCursor(cursor explorerCursor) (string, error) {
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeExplorerCursor(value string) (explorerCursor, error) {
	var result explorerCursor
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
		return result, ErrInvalidCursor
	}
	return result, nil
}

func utcDate(value time.Time) time.Time {
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func validOpaqueTie(value string) bool {
	if len(value) != 32 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
