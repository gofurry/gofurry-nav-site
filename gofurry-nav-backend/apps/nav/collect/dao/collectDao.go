package dao

import (
	"context"
	"strings"

	collectmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/collect/models"
	"github.com/gofurry/gofurry-nav-backend/common"
	navsqlc "github.com/gofurry/gofurry-nav-backend/internal/db/nav/sqlc"
)

type CollectDAO struct {
	queries *navsqlc.Queries
}

func New(queries *navsqlc.Queries) *CollectDAO {
	return &CollectDAO{queries: queries}
}

func (dao *CollectDAO) ListObservationSummary() ([]collectmodels.ObservationStatusSummary, common.GFError) {
	rows, err := dao.queries.ListObservationStatusSummary(context.Background())
	if err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	result := make([]collectmodels.ObservationStatusSummary, 0, len(rows))
	for _, row := range rows {
		result = append(result, collectmodels.ObservationStatusSummary{
			Protocol: row.Protocol, Status: row.Status, Count: row.Count,
		})
	}
	return result, nil
}

func (dao *CollectDAO) ListObservations(query collectmodels.ObservationQuery) ([]collectmodels.ObservationItem, common.GFError) {
	limit := normalizeLimit(query.Limit, 50, 200)
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}
	rows, err := dao.queries.ListCollectorObservations(context.Background(), navsqlc.ListCollectorObservationsParams{
		SiteID: query.SiteID, Target: strings.TrimSpace(query.Target),
		Protocol: strings.TrimSpace(query.Protocol), Status: strings.TrimSpace(query.Status),
		RowLimit: int32(limit), RowOffset: int32(offset),
	})
	if err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	result := make([]collectmodels.ObservationItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, collectmodels.ObservationItem{
			ID: row.ID, SiteID: row.SiteID, Target: row.Target, Protocol: row.Protocol,
			Status: row.Status, ObservedAt: row.ObservedAt.Time, DurationMS: row.DurationMs,
			ErrorCode: row.ErrorCode, ErrorMessage: row.ErrorMessage,
			CollectorID: row.CollectorID, JobID: row.JobID,
		})
	}
	return result, nil
}

func normalizeLimit(value int, fallback int, max int) int {
	if value <= 0 {
		return fallback
	}
	if value > max {
		return max
	}
	return value
}
