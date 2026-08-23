package dao

import (
	"context"
	"strings"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/readmodel/models"
	"github.com/gofurry/gofurry-nav-backend/common"
	navsqlc "github.com/gofurry/gofurry-nav-backend/internal/db/nav/sqlc"
)

type ObservationDAO struct {
	queries *navsqlc.Queries
}

func New(queries *navsqlc.Queries) *ObservationDAO {
	return &ObservationDAO{queries: queries}
}

func (dao *ObservationDAO) ListObservations(siteID int64, target string, protocol string, limit int) ([]models.GfnCollectorObservation, common.GFError) {
	rows, err := dao.queries.ListTargetObservations(context.Background(), navsqlc.ListTargetObservationsParams{
		SiteID: siteID, Target: strings.TrimSpace(target), Protocol: strings.TrimSpace(protocol),
		RowLimit: int32(models.NormalizeObservationLimit(limit)),
	})
	if err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	result := make([]models.GfnCollectorObservation, 0, len(rows))
	for _, row := range rows {
		result = append(result, models.GfnCollectorObservation{
			ID: row.ID, SiteID: row.SiteID, Target: row.Target, Protocol: row.Protocol,
			Status: row.Status, ObservedAt: row.ObservedAt.Time, DurationMS: row.DurationMs,
			ErrorCode: row.ErrorCode, ErrorMessage: row.ErrorMessage, Payload: string(row.Payload),
			SchemaVersion: int(row.SchemaVersion), CreateTime: row.CreateTime.Time,
		})
	}
	return result, nil
}
