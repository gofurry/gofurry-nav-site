package observation

import (
	"context"
	"time"

	"github.com/gofurry/gofurry-nav-collector/common"
	navsqlc "github.com/gofurry/gofurry-nav-collector/internal/db/nav/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ObservationDAO struct {
	pool    *pgxpool.Pool
	queries *navsqlc.Queries
}

func NewDAO(pool *pgxpool.Pool) *ObservationDAO {
	return &ObservationDAO{pool: pool, queries: navsqlc.New(pool)}
}

func (dao *ObservationDAO) AddObservation(record *GfnCollectorObservation) common.GFError {
	duration := record.DurationMS
	err := dao.queries.InsertObservation(context.Background(), navsqlc.InsertObservationParams{
		ID: record.ID, SiteID: record.SiteID, Target: record.Target, Protocol: record.Protocol,
		Status: record.Status, ObservedAt: timestamptz(record.ObservedAt), DurationMs: &duration,
		ErrorCode: record.ErrorCode, ErrorMessage: record.ErrorMessage, Payload: []byte(record.Payload),
		SchemaVersion: int32(record.SchemaVersion), CreateTime: timestamptz(record.CreateTime),
		JobID: record.JobID, RunID: record.RunID, CollectorInstanceID: record.InstanceID,
	})
	if err != nil {
		return common.NewDaoError(err.Error())
	}
	return nil
}

func (dao *ObservationDAO) DeleteByProtocolLimit(protocol string, count string) (int64, common.GFError) {
	// Destructive gfn_collector_observation pruning is frozen in P0.1.1.
	// Keep the hook so legacy log retention can continue without silently
	// re-enabling fact deletion before the P0.2 retention design.
	return 0, nil
}

type ObservationTrendRow struct {
	Protocol   string
	Status     string
	ObservedAt time.Time
	DurationMS int64
	ErrorCode  *string
	Payload    string
}

func (dao *ObservationDAO) ListTrendRows(ctx context.Context, siteID int64, target string, since time.Time, limit int) ([]ObservationTrendRow, common.GFError) {
	if siteID <= 0 || target == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = defaultTrendMaxRows
	}
	rows, err := dao.queries.ListObservationTrendRows(ctx, navsqlc.ListObservationTrendRowsParams{
		SiteID: siteID, Target: target, Since: timestamptz(since), LimitCount: int32(limit),
	})
	if err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	result := make([]ObservationTrendRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, ObservationTrendRow{Protocol: row.Protocol, Status: row.Status,
			ObservedAt: row.ObservedAt.Time, DurationMS: row.DurationMs, ErrorCode: row.ErrorCode, Payload: row.Payload})
	}
	return result, nil
}

func (dao *ObservationDAO) ListChangeRows(ctx context.Context, siteID int64, target string, since time.Time, limit int) ([]ObservationTrendRow, common.GFError) {
	if siteID <= 0 || target == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = defaultChangeMaxRows
	}
	rows, err := dao.queries.ListObservationChangeRows(ctx, navsqlc.ListObservationChangeRowsParams{
		SiteID: siteID, Target: target, Since: timestamptz(since), LimitCount: int32(limit),
	})
	if err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	result := make([]ObservationTrendRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, ObservationTrendRow{Protocol: row.Protocol, Status: row.Status,
			ObservedAt: row.ObservedAt.Time, DurationMS: row.DurationMs, ErrorCode: row.ErrorCode, Payload: row.Payload})
	}
	return result, nil
}

func timestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: !value.IsZero()}
}
