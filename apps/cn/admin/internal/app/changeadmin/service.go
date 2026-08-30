package changeadmin

import (
	"context"
	"encoding/json"
	"sort"
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
	result := make([]Overview, 0, 10)
	if domain == "" || domain == "game" {
		rows, err := service.game.AdminGameChangeOverview(ctx)
		if err != nil {
			return nil, daoError(err)
		}
		for _, row := range rows {
			result = append(result, overviewDTO("game", row.DetectorKey, row.DetectorVersion, row.Status, row.Description, row.WatermarkPolicy, row.SourceStartDate, row.ProcessedThrough, row.UpstreamProcessedThrough, row.LatestProjectionDate, row.LatestEventCount, row.TotalEventCount))
		}
	}
	if domain == "" || domain == "nav" {
		rows, err := service.nav.AdminNavChangeOverview(ctx)
		if err != nil {
			return nil, daoError(err)
		}
		for _, row := range rows {
			result = append(result, overviewDTO("nav", row.DetectorKey, row.DetectorVersion, row.Status, row.Description, row.WatermarkPolicy, row.SourceStartDate, row.ProcessedThrough, row.UpstreamProcessedThrough, row.LatestProjectionDate, row.LatestEventCount, row.TotalEventCount))
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Domain == result[j].Domain {
			return result[i].DetectorKey < result[j].DetectorKey
		}
		return result[i].Domain < result[j].Domain
	})
	return result, nil
}
func (service *Service) Registry(ctx context.Context, filter Filters) ([]Registry, common.Error) {
	if err := validateDomain(filter.Domain); err != nil {
		return nil, err
	}
	result := make([]Registry, 0, 10)
	if filter.Domain == "" || filter.Domain == "game" {
		rows, err := service.game.AdminListGameChangeRegistry(ctx)
		if err != nil {
			return nil, daoError(err)
		}
		for _, row := range rows {
			item := Registry{Domain: "game", DetectorKey: row.DetectorKey, DetectorVersion: row.DetectorVersion, SourceKind: row.SourceKind, SourceContracts: row.SourceContracts, DetectionPolicy: row.DetectionPolicy, WatermarkPolicy: row.WatermarkPolicy, EventCodes: row.EventCodes, ProcessingGrain: row.ProcessingGrain, Status: row.Status, Description: row.Description, CreatedAt: timestampText(row.CreatedAt), RetiredAt: timestampTextPtr(row.RetiredAt)}
			if matches(item.Domain, item.DetectorKey, item.DetectorVersion, item.Status, filter) {
				result = append(result, item)
			}
		}
	}
	if filter.Domain == "" || filter.Domain == "nav" {
		rows, err := service.nav.AdminListNavChangeRegistry(ctx)
		if err != nil {
			return nil, daoError(err)
		}
		for _, row := range rows {
			item := Registry{Domain: "nav", DetectorKey: row.DetectorKey, DetectorVersion: row.DetectorVersion, SourceKind: row.SourceKind, SourceContracts: row.SourceContracts, DetectionPolicy: row.DetectionPolicy, WatermarkPolicy: row.WatermarkPolicy, EventCodes: row.EventCodes, ProcessingGrain: row.ProcessingGrain, Status: row.Status, Description: row.Description, CreatedAt: timestampText(row.CreatedAt), RetiredAt: timestampTextPtr(row.RetiredAt)}
			if matches(item.Domain, item.DetectorKey, item.DetectorVersion, item.Status, filter) {
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
	result := make([]Checkpoint, 0, 10)
	if filter.Domain == "" || filter.Domain == "game" {
		rows, err := service.game.AdminListGameChangeCheckpoints(ctx)
		if err != nil {
			return nil, daoError(err)
		}
		for _, row := range rows {
			item := checkpointDTO("game", row.DetectorKey, row.DetectorVersion, row.Status, row.WatermarkPolicy, row.SourceStartDate, row.ProcessedThrough, row.UpstreamProcessedThrough, row.CreatedAt, row.UpdatedAt)
			if matches(item.Domain, item.DetectorKey, item.DetectorVersion, item.Status, filter) {
				result = append(result, item)
			}
		}
	}
	if filter.Domain == "" || filter.Domain == "nav" {
		rows, err := service.nav.AdminListNavChangeCheckpoints(ctx)
		if err != nil {
			return nil, daoError(err)
		}
		for _, row := range rows {
			item := checkpointDTO("nav", row.DetectorKey, row.DetectorVersion, row.Status, row.WatermarkPolicy, row.SourceStartDate, row.ProcessedThrough, row.UpstreamProcessedThrough, row.CreatedAt, row.UpdatedAt)
			if matches(item.Domain, item.DetectorKey, item.DetectorVersion, item.Status, filter) {
				result = append(result, item)
			}
		}
	}
	return result, nil
}
func (service *Service) Events(ctx context.Context, filter Filters) (EventPage, common.Error) {
	if filter.Domain != "game" && filter.Domain != "nav" {
		return EventPage{}, common.NewValidationError("domain must be game or nav")
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 200 {
		filter.PageSize = 50
	}
	offset := (filter.Page - 1) * filter.PageSize
	from, through := dateParam(filter.From), dateParam(filter.Through)
	page := EventPage{List: make([]Event, 0), Page: filter.Page, PageSize: filter.PageSize}
	if filter.Domain == "game" {
		params := gamesqlc.AdminCountGameChangeEventsParams{DetectorKey: filter.DetectorKey, DetectorVersion: filter.DetectorVersion, FromDate: from, ThroughDate: through, EventCode: filter.EventCode, ScopeKind: filter.ScopeKind, ScopeKey: filter.ScopeKey, EntityID: filter.EntityID}
		total, err := service.game.AdminCountGameChangeEvents(ctx, params)
		if err != nil {
			return EventPage{}, daoError(err)
		}
		rows, err := service.game.AdminListGameChangeEvents(ctx, gamesqlc.AdminListGameChangeEventsParams{DetectorKey: filter.DetectorKey, DetectorVersion: filter.DetectorVersion, FromDate: from, ThroughDate: through, EventCode: filter.EventCode, ScopeKind: filter.ScopeKind, ScopeKey: filter.ScopeKey, EntityID: filter.EntityID, RowOffset: offset, RowLimit: filter.PageSize})
		if err != nil {
			return EventPage{}, daoError(err)
		}
		page.Total = total
		for _, row := range rows {
			page.List = append(page.List, eventDTO("game", row.EventKey, row.DetectorKey, row.DetectorVersion, row.EntityID, row.HistoricalName, row.ProjectionDate, row.EventAt, row.TimeBasis, row.EventCode, row.ScopeKind, row.ScopeKey, row.OldValue, row.NewValue, row.SourceEventKey, row.SourceBeforeKey, row.SourceAfterKey, row.SourceBeforeAt, row.SourceAfterAt, row.SourceVersions, row.MaterializedAt))
		}
		return page, nil
	}
	params := navsqlc.AdminCountNavChangeEventsParams{DetectorKey: filter.DetectorKey, DetectorVersion: filter.DetectorVersion, FromDate: from, ThroughDate: through, EventCode: filter.EventCode, ScopeKind: filter.ScopeKind, ScopeKey: filter.ScopeKey, EntityID: filter.EntityID}
	total, err := service.nav.AdminCountNavChangeEvents(ctx, params)
	if err != nil {
		return EventPage{}, daoError(err)
	}
	rows, err := service.nav.AdminListNavChangeEvents(ctx, navsqlc.AdminListNavChangeEventsParams{DetectorKey: filter.DetectorKey, DetectorVersion: filter.DetectorVersion, FromDate: from, ThroughDate: through, EventCode: filter.EventCode, ScopeKind: filter.ScopeKind, ScopeKey: filter.ScopeKey, EntityID: filter.EntityID, RowOffset: offset, RowLimit: filter.PageSize})
	if err != nil {
		return EventPage{}, daoError(err)
	}
	page.Total = total
	for _, row := range rows {
		page.List = append(page.List, eventDTO("nav", row.EventKey, row.DetectorKey, row.DetectorVersion, row.EntityID, row.HistoricalName, row.ProjectionDate, row.EventAt, row.TimeBasis, row.EventCode, row.ScopeKind, row.ScopeKey, row.OldValue, row.NewValue, row.SourceEventKey, row.SourceBeforeKey, row.SourceAfterKey, row.SourceBeforeAt, row.SourceAfterAt, row.SourceVersions, row.MaterializedAt))
	}
	return page, nil
}

func overviewDTO(domain, key string, version int32, status, description, watermark string, source, processed, upstream, latest pgtype.Date, latestCount, total int64) Overview {
	return Overview{Domain: domain, DetectorKey: key, DetectorVersion: version, Status: status, Description: description, WatermarkPolicy: watermark, SourceStartDate: dateTextPtr(source), ProcessedThrough: dateTextPtr(processed), UpstreamProcessedThrough: dateTextPtr(upstream), LagDays: lagDays(source, processed, upstream), LatestProjectionDate: dateTextPtr(latest), LatestEventCount: latestCount, TotalEventCount: total}
}
func checkpointDTO(domain, key string, version int32, status, watermark string, source, processed, upstream pgtype.Date, created, updated pgtype.Timestamptz) Checkpoint {
	return Checkpoint{Domain: domain, DetectorKey: key, DetectorVersion: version, Status: status, WatermarkPolicy: watermark, SourceStartDate: dateTextPtr(source), ProcessedThrough: dateTextPtr(processed), UpstreamProcessedThrough: dateTextPtr(upstream), LagDays: lagDays(source, processed, upstream), CreatedAt: timestampText(created), UpdatedAt: timestampText(updated)}
}
func eventDTO(domain, eventKey, detector string, version int32, entity int64, name string, date pgtype.Date, eventAt pgtype.Timestamptz, timeBasis, code, scopeKind, scopeKey string, oldValue, newValue []byte, sourceEvent, sourceBefore, sourceAfter string, beforeAt, afterAt pgtype.Timestamptz, versions []byte, materialized pgtype.Timestamptz) Event {
	if len(oldValue) == 0 {
		oldValue = []byte(`{}`)
	}
	if len(newValue) == 0 {
		newValue = []byte(`{}`)
	}
	if len(versions) == 0 {
		versions = []byte(`{}`)
	}
	return Event{Domain: domain, EventKey: eventKey, DetectorKey: detector, DetectorVersion: version, EntityID: entity, HistoricalName: name, ProjectionDate: dateText(date), EventAt: timestampTextPtr(eventAt), TimeBasis: timeBasis, EventCode: code, ScopeKind: scopeKind, ScopeKey: scopeKey, OldValue: json.RawMessage(oldValue), NewValue: json.RawMessage(newValue), SourceEventKey: sourceEvent, SourceBeforeKey: sourceBefore, SourceAfterKey: sourceAfter, SourceBeforeAt: timestampTextPtr(beforeAt), SourceAfterAt: timestampTextPtr(afterAt), SourceVersions: json.RawMessage(versions), MaterializedAt: timestampText(materialized)}
}
func validateDomain(domain string) common.Error {
	if domain != "" && domain != "game" && domain != "nav" {
		return common.NewValidationError("domain must be game or nav")
	}
	return nil
}
func matches(domain, key string, version int32, status string, filter Filters) bool {
	return (filter.Domain == "" || filter.Domain == domain) && (filter.DetectorKey == "" || filter.DetectorKey == key) && (filter.DetectorVersion == 0 || filter.DetectorVersion == version) && (filter.Status == "" || filter.Status == status)
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
func daoError(err error) common.Error {
	if err == nil {
		return nil
	}
	return common.NewDaoError(err.Error())
}
