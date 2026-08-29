package metricadmin

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"

	gamesqlc "github.com/gofurry/gofurry-admin/internal/db/game/sqlc"
	navsqlc "github.com/gofurry/gofurry-admin/internal/db/nav/sqlc"
	"github.com/gofurry/gofurry-admin/pkg/common"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	game *gamesqlc.Queries
	nav  *navsqlc.Queries
}

func New(gamePool, navPool *pgxpool.Pool) *Service {
	return &Service{game: gamesqlc.New(gamePool), nav: navsqlc.New(navPool)}
}

func (service *Service) Overview(ctx context.Context, domain string) ([]Overview, common.Error) {
	if err := validateDomain(domain); err != nil {
		return nil, err
	}
	result := make([]Overview, 0, 6)
	if domain == "" || domain == "game" {
		rows, err := service.game.AdminGameMetricOverview(ctx)
		if err != nil {
			return nil, daoError(err)
		}
		for _, row := range rows {
			result = append(result, overviewDTO("game", row.MetricKey, row.MetricVersion, row.Description,
				row.SourceStartDate, row.ProcessedThrough, row.UpstreamProcessedThrough, row.LatestFactDate,
				row.PopulationCount, row.EligibleCount, row.NotApplicableCount, row.PositiveCount,
				row.NegativeCount, row.StaleCount, row.NotProbedCount, row.ProbeFailedCount, row.UnknownCount))
		}
	}
	if domain == "" || domain == "nav" {
		rows, err := service.nav.AdminNavMetricOverview(ctx)
		if err != nil {
			return nil, daoError(err)
		}
		for _, row := range rows {
			upstream := minDate(row.TargetProcessedThrough, row.SiteProcessedThrough)
			result = append(result, overviewDTO("nav", row.MetricKey, row.MetricVersion, row.Description,
				row.SourceStartDate, row.ProcessedThrough, upstream, row.LatestFactDate,
				row.PopulationCount, row.EligibleCount, row.NotApplicableCount, row.PositiveCount,
				row.NegativeCount, row.StaleCount, row.NotProbedCount, row.ProbeFailedCount, row.UnknownCount))
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Domain == result[j].Domain {
			return result[i].MetricKey < result[j].MetricKey
		}
		return result[i].Domain < result[j].Domain
	})
	return result, nil
}

func (service *Service) Registry(ctx context.Context, filter Filters) ([]Registry, common.Error) {
	if err := validateDomain(filter.Domain); err != nil {
		return nil, err
	}
	result := make([]Registry, 0, 6)
	if filter.Domain == "" || filter.Domain == "game" {
		rows, err := service.game.AdminListGameMetricRegistry(ctx)
		if err != nil {
			return nil, daoError(err)
		}
		for _, row := range rows {
			item := Registry{Domain: "game", MetricKey: row.MetricKey, MetricVersion: row.MetricVersion,
				MetricKind: row.MetricKind, EntityLevel: row.EntityLevel, TimeGrain: row.TimeGrain,
				SourceFacts: row.SourceFacts, EligibilityPolicy: row.EligibilityPolicy, StatePolicy: row.StatePolicy,
				CoveragePolicy: row.CoveragePolicy, FreshnessSeconds: row.FreshnessSeconds,
				AllowedDimensions: row.AllowedDimensions, Status: row.Status, Description: row.Description,
				CreatedAt: timestampText(row.CreatedAt), RetiredAt: timestampTextPtr(row.RetiredAt)}
			if registryMatches(item, filter) {
				result = append(result, item)
			}
		}
	}
	if filter.Domain == "" || filter.Domain == "nav" {
		rows, err := service.nav.AdminListNavMetricRegistry(ctx)
		if err != nil {
			return nil, daoError(err)
		}
		for _, row := range rows {
			item := Registry{Domain: "nav", MetricKey: row.MetricKey, MetricVersion: row.MetricVersion,
				MetricKind: row.MetricKind, EntityLevel: row.EntityLevel, TimeGrain: row.TimeGrain,
				SourceFacts: row.SourceFacts, EligibilityPolicy: row.EligibilityPolicy, StatePolicy: row.StatePolicy,
				CoveragePolicy: row.CoveragePolicy, FreshnessSeconds: row.FreshnessSeconds,
				AllowedDimensions: row.AllowedDimensions, Status: row.Status, Description: row.Description,
				CreatedAt: timestampText(row.CreatedAt), RetiredAt: timestampTextPtr(row.RetiredAt)}
			if registryMatches(item, filter) {
				result = append(result, item)
			}
		}
	}
	return result, nil
}

func (service *Service) Checkpoints(ctx context.Context, filter Filters) ([]Checkpoint, common.Error) {
	if err := validateDomain(filter.Domain); err != nil {
		return nil, err
	}
	result := make([]Checkpoint, 0, 6)
	if filter.Domain == "" || filter.Domain == "game" {
		rows, err := service.game.AdminListGameMetricCheckpoints(ctx)
		if err != nil {
			return nil, daoError(err)
		}
		for _, row := range rows {
			item := checkpointDTO("game", row.MetricKey, row.MetricVersion, row.Status, row.SourceStartDate,
				row.ProcessedThrough, row.UpstreamProcessedThrough, row.CreatedAt, row.UpdatedAt)
			if checkpointMatches(item, filter) {
				result = append(result, item)
			}
		}
	}
	if filter.Domain == "" || filter.Domain == "nav" {
		rows, err := service.nav.AdminListNavMetricCheckpoints(ctx)
		if err != nil {
			return nil, daoError(err)
		}
		for _, row := range rows {
			item := checkpointDTO("nav", row.MetricKey, row.MetricVersion, row.Status, row.SourceStartDate,
				row.ProcessedThrough, minDate(row.TargetProcessedThrough, row.SiteProcessedThrough), row.CreatedAt, row.UpdatedAt)
			if checkpointMatches(item, filter) {
				result = append(result, item)
			}
		}
	}
	return result, nil
}

func (service *Service) Daily(ctx context.Context, filter Filters) ([]Daily, common.Error) {
	if err := validateDomain(filter.Domain); err != nil {
		return nil, err
	}
	if filter.DimensionKey == "" {
		filter.DimensionKey = "global"
	}
	if filter.DimensionValue == "" {
		filter.DimensionValue = "all"
	}
	from, through := dateParam(filter.From), dateParam(filter.Through)
	result := make([]Daily, 0)
	if filter.Domain == "" || filter.Domain == "game" {
		rows, err := service.game.AdminListGameMetricDaily(ctx, gamesqlc.AdminListGameMetricDailyParams{
			MetricKey: filter.MetricKey, MetricVersion: filter.MetricVersion, FromDate: from, ThroughDate: through,
			DimensionKey: filter.DimensionKey, DimensionValue: filter.DimensionValue, RowLimit: 1000,
		})
		if err != nil {
			return nil, daoError(err)
		}
		for _, row := range rows {
			result = append(result, dailyDTO("game", row.MetricKey, row.MetricVersion, row.FactDate,
				row.DimensionKey, row.DimensionValue, row.PopulationCount, row.EligibleCount,
				row.NotApplicableCount, row.PositiveCount, row.NegativeCount, row.StaleCount,
				row.NotProbedCount, row.ProbeFailedCount, row.UnknownCount, row.ComputedAt))
		}
	}
	if filter.Domain == "" || filter.Domain == "nav" {
		rows, err := service.nav.AdminListNavMetricDaily(ctx, navsqlc.AdminListNavMetricDailyParams{
			MetricKey: filter.MetricKey, MetricVersion: filter.MetricVersion, FromDate: from, ThroughDate: through,
			DimensionKey: filter.DimensionKey, DimensionValue: filter.DimensionValue, RowLimit: 1000,
		})
		if err != nil {
			return nil, daoError(err)
		}
		for _, row := range rows {
			result = append(result, dailyDTO("nav", row.MetricKey, row.MetricVersion, row.FactDate,
				row.DimensionKey, row.DimensionValue, row.PopulationCount, row.EligibleCount,
				row.NotApplicableCount, row.PositiveCount, row.NegativeCount, row.StaleCount,
				row.NotProbedCount, row.ProbeFailedCount, row.UnknownCount, row.ComputedAt))
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].FactDate == result[j].FactDate {
			return result[i].MetricKey < result[j].MetricKey
		}
		return result[i].FactDate > result[j].FactDate
	})
	return result, nil
}

func (service *Service) Entities(ctx context.Context, filter Filters) (EntityPage, common.Error) {
	if filter.Domain != "game" && filter.Domain != "nav" {
		return EntityPage{}, common.NewValidationError("domain must be game or nav")
	}
	if strings.TrimSpace(filter.MetricKey) == "" || filter.MetricVersion <= 0 || filter.FactDate.IsZero() {
		return EntityPage{}, common.NewValidationError("metric, version, and fact_date are required")
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 200 {
		filter.PageSize = 50
	}
	offset := (filter.Page - 1) * filter.PageSize
	date := pgtype.Date{Time: filter.FactDate, Valid: true}
	page := EntityPage{List: make([]Entity, 0)}
	if filter.Domain == "game" {
		params := gamesqlc.AdminCountGameMetricEntitiesParams{MetricKey: filter.MetricKey, MetricVersion: filter.MetricVersion,
			FactDate: date, State: filter.State, ReasonCode: filter.ReasonCode}
		total, err := service.game.AdminCountGameMetricEntities(ctx, params)
		if err != nil {
			return EntityPage{}, daoError(err)
		}
		rows, err := service.game.AdminListGameMetricEntities(ctx, gamesqlc.AdminListGameMetricEntitiesParams{
			MetricKey: filter.MetricKey, MetricVersion: filter.MetricVersion, FactDate: date,
			State: filter.State, ReasonCode: filter.ReasonCode, RowLimit: filter.PageSize, RowOffset: offset,
		})
		if err != nil {
			return EntityPage{}, daoError(err)
		}
		page.Total = total
		for _, row := range rows {
			page.List = append(page.List, entityDTO("game", row.EntityID, row.HistoricalName, row.State, row.ReasonCode,
				row.SourceObservedAt, row.DimensionValues, row.SourceProjectionVersions, row.EvaluatedAt))
		}
		return page, nil
	}
	params := navsqlc.AdminCountNavMetricEntitiesParams{MetricKey: filter.MetricKey, MetricVersion: filter.MetricVersion,
		FactDate: date, State: filter.State, ReasonCode: filter.ReasonCode}
	total, err := service.nav.AdminCountNavMetricEntities(ctx, params)
	if err != nil {
		return EntityPage{}, daoError(err)
	}
	rows, err := service.nav.AdminListNavMetricEntities(ctx, navsqlc.AdminListNavMetricEntitiesParams{
		MetricKey: filter.MetricKey, MetricVersion: filter.MetricVersion, FactDate: date,
		State: filter.State, ReasonCode: filter.ReasonCode, RowLimit: filter.PageSize, RowOffset: offset,
	})
	if err != nil {
		return EntityPage{}, daoError(err)
	}
	page.Total = total
	for _, row := range rows {
		page.List = append(page.List, entityDTO("nav", row.EntityID, row.HistoricalName, row.State, row.ReasonCode,
			row.SourceObservedAt, row.DimensionValues, row.SourceProjectionVersions, row.EvaluatedAt))
	}
	return page, nil
}

func overviewDTO(domain, key string, version int32, description string, source, processed, upstream, latest pgtype.Date,
	population, eligible, notApplicable, positive, negative, stale, notProbed, probeFailed, unknown int64) Overview {
	adoption, coverage := rates(positive, negative, eligible)
	return Overview{Domain: domain, MetricKey: key, MetricVersion: version, Description: description,
		ProcessedThrough: dateTextPtr(processed), UpstreamProcessedThrough: dateTextPtr(upstream),
		LagDays: lagDays(source, processed, upstream), LatestFactDate: dateTextPtr(latest),
		PopulationCount: population, EligibleCount: eligible, NotApplicableCount: notApplicable,
		PositiveCount: positive, NegativeCount: negative, StaleCount: stale, NotProbedCount: notProbed,
		ProbeFailedCount: probeFailed, UnknownCount: unknown, AdoptionRate: adoption, CoverageRate: coverage}
}

func checkpointDTO(domain, key string, version int32, status string, source, processed, upstream pgtype.Date,
	created, updated pgtype.Timestamptz) Checkpoint {
	return Checkpoint{Domain: domain, MetricKey: key, MetricVersion: version, Status: status,
		SourceStartDate: dateTextPtr(source), ProcessedThrough: dateTextPtr(processed),
		UpstreamProcessedThrough: dateTextPtr(upstream), LagDays: lagDays(source, processed, upstream),
		CreatedAt: timestampText(created), UpdatedAt: timestampText(updated)}
}

func dailyDTO(domain, key string, version int32, factDate pgtype.Date, dimensionKey, dimensionValue string,
	population, eligible, notApplicable, positive, negative, stale, notProbed, probeFailed, unknown int64,
	computed pgtype.Timestamptz) Daily {
	adoption, coverage := rates(positive, negative, eligible)
	return Daily{Domain: domain, MetricKey: key, MetricVersion: version, FactDate: dateText(factDate),
		DimensionKey: dimensionKey, DimensionValue: dimensionValue, PopulationCount: population,
		EligibleCount: eligible, NotApplicableCount: notApplicable, PositiveCount: positive,
		NegativeCount: negative, StaleCount: stale, NotProbedCount: notProbed,
		ProbeFailedCount: probeFailed, UnknownCount: unknown, AdoptionRate: adoption,
		CoverageRate: coverage, ComputedAt: timestampText(computed)}
}

func entityDTO(domain string, id int64, name, state, reason string, observed pgtype.Timestamptz,
	dimensions, projections []byte, evaluated pgtype.Timestamptz) Entity {
	if len(dimensions) == 0 {
		dimensions = []byte(`{}`)
	}
	if len(projections) == 0 {
		projections = []byte(`{}`)
	}
	return Entity{Domain: domain, EntityID: id, HistoricalName: name, State: state, ReasonCode: reason,
		SourceObservedAt: timestampTextPtr(observed), DimensionValues: json.RawMessage(dimensions),
		SourceProjectionVersions: json.RawMessage(projections), EvaluatedAt: timestampText(evaluated)}
}

func rates(positive, negative, eligible int64) (*float64, *float64) {
	known := positive + negative
	var adoption, coverage *float64
	if known > 0 {
		value := float64(positive) / float64(known)
		adoption = &value
	}
	if eligible > 0 {
		value := float64(known) / float64(eligible)
		coverage = &value
	}
	return adoption, coverage
}

func validateDomain(domain string) common.Error {
	if domain != "" && domain != "game" && domain != "nav" {
		return common.NewValidationError("domain must be game or nav")
	}
	return nil
}

func registryMatches(item Registry, filter Filters) bool {
	return (filter.MetricKey == "" || item.MetricKey == filter.MetricKey) &&
		(filter.MetricVersion == 0 || item.MetricVersion == filter.MetricVersion) &&
		(filter.Status == "" || item.Status == filter.Status)
}

func checkpointMatches(item Checkpoint, filter Filters) bool {
	return (filter.MetricKey == "" || item.MetricKey == filter.MetricKey) &&
		(filter.MetricVersion == 0 || item.MetricVersion == filter.MetricVersion) &&
		(filter.Status == "" || item.Status == filter.Status)
}

func minDate(first, second pgtype.Date) pgtype.Date {
	if !first.Valid || !second.Valid {
		return pgtype.Date{}
	}
	if first.Time.Before(second.Time) {
		return first
	}
	return second
}

func lagDays(source, processed, upstream pgtype.Date) *int {
	if !source.Valid || !upstream.Valid {
		return nil
	}
	next := source.Time.UTC()
	if processed.Valid {
		next = processed.Time.UTC().AddDate(0, 0, 1)
	}
	lag := int(upstream.Time.UTC().Sub(next).Hours()/24) + 1
	if lag < 0 {
		lag = 0
	}
	return &lag
}

func dateParam(value *time.Time) pgtype.Date {
	if value == nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: value.UTC(), Valid: true}
}

func dateText(value pgtype.Date) string {
	if !value.Valid {
		return ""
	}
	return value.Time.UTC().Format(time.DateOnly)
}

func dateTextPtr(value pgtype.Date) *string {
	if !value.Valid {
		return nil
	}
	text := dateText(value)
	return &text
}

func timestampText(value pgtype.Timestamptz) string {
	if !value.Valid {
		return ""
	}
	return value.Time.UTC().Format(time.RFC3339)
}

func timestampTextPtr(value pgtype.Timestamptz) *string {
	if !value.Valid {
		return nil
	}
	text := timestampText(value)
	return &text
}

func daoError(err error) common.Error {
	if err == nil {
		return nil
	}
	return common.NewDaoError(err.Error())
}
