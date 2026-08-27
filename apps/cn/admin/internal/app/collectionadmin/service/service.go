package service

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"

	collectionmodels "github.com/gofurry/gofurry-admin/internal/app/collectionadmin/models"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	gamesqlc "github.com/gofurry/gofurry-admin/internal/db/game/sqlc"
	navsqlc "github.com/gofurry/gofurry-admin/internal/db/nav/sqlc"
	"github.com/gofurry/gofurry-admin/internal/infra/cache"
	"github.com/gofurry/gofurry-admin/pkg/common"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	gamePool *pgxpool.Pool
	navPool  *pgxpool.Pool
	game     *gamesqlc.Queries
	nav      *navsqlc.Queries
	audit    *audit.Logger
}

func New(gamePool, navPool *pgxpool.Pool, auditLogger *audit.Logger) *Service {
	return &Service{
		gamePool: gamePool, navPool: navPool,
		game: gamesqlc.New(gamePool), nav: navsqlc.New(navPool),
		audit: auditLogger,
	}
}

type Filters struct {
	Domain  string
	Status  string
	JobKey  string
	Trigger string
	Since   *time.Time
	Until   *time.Time
	Limit   int32
	Offset  int32
}

type ResultFilters struct {
	GameID   *int64
	AppID    *int64
	SiteID   *int64
	Target   *string
	Protocol string
	Limit    int32
	Offset   int32
}

func (s *Service) Overview(ctx context.Context) (collectionmodels.Overview, common.Error) {
	game, err := s.game.AdminGameCollectionCounts(ctx)
	if err != nil {
		return collectionmodels.Overview{}, daoError(err)
	}
	nav, err := s.nav.AdminNavCollectionCounts(ctx)
	if err != nil {
		return collectionmodels.Overview{}, daoError(err)
	}
	return collectionmodels.Overview{
		RunningCount: game.RunningCount + nav.RunningCount,
		QueuedCount:  game.QueuedCount + nav.QueuedCount,
		Failed24h:    game.Failed24h + nav.Failed24h,
		Missed24h:    game.Missed24h + nav.Missed24h,
	}, nil
}

func (s *Service) Instances(ctx context.Context) ([]collectionmodels.Instance, common.Error) {
	gameRows, err := s.game.AdminListGameCollectorInstances(ctx, 50)
	if err != nil {
		return nil, daoError(err)
	}
	navRows, err := s.nav.AdminListNavCollectorInstances(ctx, 50)
	if err != nil {
		return nil, daoError(err)
	}
	rows := make([]collectionmodels.Instance, 0, len(gameRows)+len(navRows))
	for _, row := range gameRows {
		rows = append(rows, instanceDTO("game", row.InstanceID, row.CollectorID, row.Hostname, row.Version, row.CommitSha, row.Capabilities, row.StartedAt, row.LastHeartbeatAt, row.StoppedAt, row.HeartbeatAgeSeconds))
	}
	for _, row := range navRows {
		rows = append(rows, instanceDTO("nav", row.InstanceID, row.CollectorID, row.Hostname, row.Version, row.CommitSha, row.Capabilities, row.StartedAt, row.LastHeartbeatAt, row.StoppedAt, row.HeartbeatAgeSeconds))
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].LastHeartbeatAt.After(rows[j].LastHeartbeatAt) })
	return rows, nil
}

func (s *Service) Schedules(ctx context.Context) ([]collectionmodels.Schedule, common.Error) {
	gameRows, err := s.game.AdminListGameCollectionSchedules(ctx)
	if err != nil {
		return nil, daoError(err)
	}
	navRows, err := s.nav.AdminListNavCollectionSchedules(ctx)
	if err != nil {
		return nil, daoError(err)
	}
	rows := make([]collectionmodels.Schedule, 0, len(gameRows)+len(navRows))
	for _, row := range gameRows {
		rows = append(rows, scheduleDTO("game", row.ID, row.JobKey, row.Name, row.Enabled, row.ScheduleKind,
			row.CronExpression, row.IntervalSeconds, row.AnchorAt, row.Timezone, row.MisfirePolicy,
			row.MisfireGraceSeconds, row.OverlapPolicy, row.Priority, row.ConcurrencyKey, row.Version,
			row.LastMaterializedFor, row.NextScheduledFor, row.LastStatus, row.LastSuccessCount, row.LastExpectedCount))
	}
	for _, row := range navRows {
		rows = append(rows, scheduleDTO("nav", row.ID, row.JobKey, row.Name, row.Enabled, row.ScheduleKind,
			row.CronExpression, row.IntervalSeconds, row.AnchorAt, row.Timezone, row.MisfirePolicy,
			row.MisfireGraceSeconds, row.OverlapPolicy, row.Priority, row.ConcurrencyKey, row.Version,
			row.LastMaterializedFor, row.NextScheduledFor, row.LastStatus, row.LastSuccessCount, row.LastExpectedCount))
	}
	return rows, nil
}

func (s *Service) Jobs(ctx context.Context, filter Filters) ([]collectionmodels.Job, common.Error) {
	filter = normalizeFilters(filter)
	if filter.Domain != "" && filter.Domain != "game" && filter.Domain != "nav" {
		return nil, common.NewValidationError("domain must be game or nav")
	}
	queryLimit, queryOffset := perDomainPage(filter)
	rows := make([]collectionmodels.Job, 0)
	if filter.Domain == "" || filter.Domain == "game" {
		gameRows, err := s.game.AdminListGameCollectionJobs(ctx, gamesqlc.AdminListGameCollectionJobsParams{
			Status: filter.Status, JobKey: filter.JobKey, Trigger: filter.Trigger, RowLimit: queryLimit, RowOffset: queryOffset,
		})
		if err != nil {
			return nil, daoError(err)
		}
		for _, row := range gameRows {
			item := jobDTO("game", row.ID, row.ScheduleID, row.JobKey, row.Trigger, row.ScopeType, row.ScopeID,
				row.Target, row.Tasks, row.Priority, row.ScheduledFor, row.Status, row.RequestedBy,
				row.ClaimedBy, row.LeaseUntil, row.CancelRequestedAt, row.CreatedAt, row.CompletedAt, row.RunID)
			item.Progress = realtimeProgress("game", row.RunID)
			rows = append(rows, item)
		}
	}
	if filter.Domain == "" || filter.Domain == "nav" {
		navRows, err := s.nav.AdminListNavCollectionJobs(ctx, navsqlc.AdminListNavCollectionJobsParams{
			Status: filter.Status, JobKey: filter.JobKey, Trigger: filter.Trigger, RowLimit: queryLimit, RowOffset: queryOffset,
		})
		if err != nil {
			return nil, daoError(err)
		}
		for _, row := range navRows {
			item := jobDTO("nav", row.ID, row.ScheduleID, row.JobKey, row.Trigger, row.ScopeType, row.ScopeID,
				row.Target, row.Tasks, row.Priority, row.ScheduledFor, row.Status, row.RequestedBy,
				row.ClaimedBy, row.LeaseUntil, row.CancelRequestedAt, row.CreatedAt, row.CompletedAt, row.RunID)
			item.Progress = realtimeProgress("nav", row.RunID)
			rows = append(rows, item)
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].CreatedAt.After(rows[j].CreatedAt) })
	if filter.Domain == "" {
		rows = page(rows, filter.Offset, filter.Limit)
	}
	return rows, nil
}

func (s *Service) Job(ctx context.Context, domain string, id int64) (collectionmodels.Job, common.Error) {
	switch domain {
	case "game":
		row, err := s.game.AdminGetGameCollectionJob(ctx, id)
		if err != nil {
			return collectionmodels.Job{}, daoError(err)
		}
		return gameJobDTO(row), nil
	case "nav":
		row, err := s.nav.AdminGetNavCollectionJob(ctx, id)
		if err != nil {
			return collectionmodels.Job{}, daoError(err)
		}
		return navJobDTO(row), nil
	default:
		return collectionmodels.Job{}, common.NewValidationError("domain must be game or nav")
	}
}

func (s *Service) Runs(ctx context.Context, filter Filters) ([]collectionmodels.Run, common.Error) {
	filter = normalizeFilters(filter)
	if filter.Domain != "" && filter.Domain != "game" && filter.Domain != "nav" {
		return nil, common.NewValidationError("domain must be game or nav")
	}
	queryLimit, queryOffset := perDomainPage(filter)
	since, until := nullableTime(filter.Since), nullableTime(filter.Until)
	rows := make([]collectionmodels.Run, 0)
	if filter.Domain == "" || filter.Domain == "game" {
		gameRows, err := s.game.AdminListGameCollectionRuns(ctx, gamesqlc.AdminListGameCollectionRunsParams{
			Status: filter.Status, JobKey: filter.JobKey, Trigger: filter.Trigger,
			Since: since, Until: until, RowLimit: queryLimit, RowOffset: queryOffset,
		})
		if err != nil {
			return nil, daoError(err)
		}
		for _, row := range gameRows {
			rows = append(rows, runDTO("game", row.ID, row.JobID, row.JobKey, row.Trigger, row.ScopeType, row.ScopeID, row.Target,
				row.AttemptNo, row.CollectorInstanceID, row.Status, row.ScheduledFor, row.StartedAt, row.EndedAt,
				row.ExpectedCount, row.AttemptedCount, row.SuccessCount, row.PartialCount, row.FailureCount,
				row.SkippedCount, row.ScheduleDelayMs, row.DurationMs, row.ErrorKind, row.ErrorMessage))
		}
	}
	if filter.Domain == "" || filter.Domain == "nav" {
		navRows, err := s.nav.AdminListNavCollectionRuns(ctx, navsqlc.AdminListNavCollectionRunsParams{
			Status: filter.Status, JobKey: filter.JobKey, Trigger: filter.Trigger,
			Since: since, Until: until, RowLimit: queryLimit, RowOffset: queryOffset,
		})
		if err != nil {
			return nil, daoError(err)
		}
		for _, row := range navRows {
			rows = append(rows, runDTO("nav", row.ID, row.JobID, row.JobKey, row.Trigger, row.ScopeType, row.ScopeID, row.Target,
				row.AttemptNo, row.CollectorInstanceID, row.Status, row.ScheduledFor, row.StartedAt, row.EndedAt,
				row.ExpectedCount, row.AttemptedCount, row.SuccessCount, row.PartialCount, row.FailureCount,
				row.SkippedCount, row.ScheduleDelayMs, row.DurationMs, row.ErrorKind, row.ErrorMessage))
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].StartedAt.After(rows[j].StartedAt) })
	if filter.Domain == "" {
		rows = page(rows, filter.Offset, filter.Limit)
	}
	return rows, nil
}

func (s *Service) Run(ctx context.Context, domain, id string) (collectionmodels.Run, common.Error) {
	switch domain {
	case "game":
		row, err := s.game.AdminGetGameCollectionRun(ctx, id)
		if err != nil {
			return collectionmodels.Run{}, daoError(err)
		}
		return runDTO("game", row.ID, row.JobID, row.JobKey, row.Trigger, row.ScopeType, row.ScopeID, row.Target,
			row.AttemptNo, row.CollectorInstanceID, row.Status, row.ScheduledFor, row.StartedAt, row.EndedAt,
			row.ExpectedCount, row.AttemptedCount, row.SuccessCount, row.PartialCount, row.FailureCount,
			row.SkippedCount, row.ScheduleDelayMs, row.DurationMs, row.ErrorKind, row.ErrorMessage), nil
	case "nav":
		row, err := s.nav.AdminGetNavCollectionRun(ctx, id)
		if err != nil {
			return collectionmodels.Run{}, daoError(err)
		}
		return runDTO("nav", row.ID, row.JobID, row.JobKey, row.Trigger, row.ScopeType, row.ScopeID, row.Target,
			row.AttemptNo, row.CollectorInstanceID, row.Status, row.ScheduledFor, row.StartedAt, row.EndedAt,
			row.ExpectedCount, row.AttemptedCount, row.SuccessCount, row.PartialCount, row.FailureCount,
			row.SkippedCount, row.ScheduleDelayMs, row.DurationMs, row.ErrorKind, row.ErrorMessage), nil
	default:
		return collectionmodels.Run{}, common.NewValidationError("domain must be game or nav")
	}
}

func (s *Service) Results(ctx context.Context, domain, runID string, filter ResultFilters) ([]collectionmodels.Result, common.Error) {
	if filter.Limit <= 0 || filter.Limit > 500 {
		filter.Limit = 100
	}
	results := make([]collectionmodels.Result, 0)
	switch domain {
	case "game":
		rows, err := s.game.AdminListGameCollectionResults(ctx, gamesqlc.AdminListGameCollectionResultsParams{
			RunID: runID, GameID: filter.GameID, Appid: filter.AppID, RowLimit: filter.Limit, RowOffset: filter.Offset,
		})
		if err != nil {
			return nil, daoError(err)
		}
		for _, row := range rows {
			results = append(results, collectionmodels.Result{Domain: "game", ID: row.ID, RunID: row.RunID,
				Task: row.TaskType, EntityID: row.GameID, AppID: row.Appid, Status: row.Status,
				DurationMS: row.DurationMs, ErrorKind: row.ErrorKind, ErrorMessage: row.ErrorMessage,
				StartedAt: row.StartedAt.Time, EndedAt: timePointer(row.EndedAt)})
		}
	case "nav":
		rows, err := s.nav.AdminListNavCollectionResults(ctx, navsqlc.AdminListNavCollectionResultsParams{
			RunID: runID, SiteID: filter.SiteID, Target: filter.Target, Protocol: filter.Protocol,
			RowLimit: filter.Limit, RowOffset: filter.Offset,
		})
		if err != nil {
			return nil, daoError(err)
		}
		for _, row := range rows {
			results = append(results, collectionmodels.Result{Domain: "nav", ID: row.ID, RunID: row.RunID,
				Task: row.Protocol, EntityID: row.SiteID, Target: row.Target, Status: row.Status,
				ObservationID: row.ObservationID, DurationMS: row.DurationMs,
				ErrorKind: row.ErrorKind, ErrorMessage: row.ErrorMessage,
				StartedAt: row.StartedAt.Time, EndedAt: timePointer(row.EndedAt)})
		}
	default:
		return nil, common.NewValidationError("domain must be game or nav")
	}
	return results, nil
}

func (s *Service) Charts(ctx context.Context, domain, jobKey string, window time.Duration) ([]collectionmodels.ChartPoint, common.Error) {
	if window <= 0 || window > 31*24*time.Hour {
		window = 24 * time.Hour
	}
	points := make([]collectionmodels.ChartPoint, 0)
	if domain == "" || domain == "game" {
		clock, err := s.game.AdminGameCollectionClock(ctx)
		if err != nil {
			return nil, daoError(err)
		}
		if !clock.Valid {
			return nil, common.NewDaoError("Game control-plane clock is unavailable")
		}
		rows, err := s.game.AdminListGameCollectionChartRows(ctx, gamesqlc.AdminListGameCollectionChartRowsParams{Since: timestamp(clock.Time.Add(-window)), JobKey: jobKey})
		if err != nil {
			return nil, daoError(err)
		}
		for _, row := range rows {
			points = append(points, chartPoint("game", row.JobID, row.JobKey, row.JobStatus, row.RunID, row.RunStatus,
				row.ExpectedCount, row.AttemptedCount, row.SuccessCount, row.PartialCount, row.FailureCount,
				row.SkippedCount, row.ScheduleDelayMs, row.DurationMs, row.CreatedAt, row.StartedAt))
		}
	}
	if domain == "" || domain == "nav" {
		clock, err := s.nav.AdminNavCollectionClock(ctx)
		if err != nil {
			return nil, daoError(err)
		}
		if !clock.Valid {
			return nil, common.NewDaoError("Nav control-plane clock is unavailable")
		}
		rows, err := s.nav.AdminListNavCollectionChartRows(ctx, navsqlc.AdminListNavCollectionChartRowsParams{Since: timestamp(clock.Time.Add(-window)), JobKey: jobKey})
		if err != nil {
			return nil, daoError(err)
		}
		for _, row := range rows {
			points = append(points, chartPoint("nav", row.JobID, row.JobKey, row.JobStatus, row.RunID, row.RunStatus,
				row.ExpectedCount, row.AttemptedCount, row.SuccessCount, row.PartialCount, row.FailureCount,
				row.SkippedCount, row.ScheduleDelayMs, row.DurationMs, row.CreatedAt, row.StartedAt))
		}
	}
	sort.Slice(points, func(i, j int) bool { return points[i].CreatedAt.Before(points[j].CreatedAt) })
	return points, nil
}

func normalizeFilters(filter Filters) Filters {
	filter.Domain = strings.TrimSpace(filter.Domain)
	if filter.Limit <= 0 || filter.Limit > 500 {
		filter.Limit = 100
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return filter
}

func perDomainPage(filter Filters) (limit, offset int32) {
	if filter.Domain != "" {
		return filter.Limit, filter.Offset
	}
	total := int64(filter.Limit) + int64(filter.Offset)
	if total > 2147483647 {
		total = 2147483647
	}
	return int32(total), 0
}

func page[T any](rows []T, offset, limit int32) []T {
	start := min(int(offset), len(rows))
	end := min(start+int(limit), len(rows))
	return rows[start:end]
}

func realtimeProgress(domain, runID string) any {
	if runID == "" || !cache.RedisReady() {
		return nil
	}
	raw, err := cache.GetString("collection:" + domain + ":run:" + runID + ":progress")
	if err != nil || raw == "" {
		return nil
	}
	var value map[string]any
	if json.Unmarshal([]byte(raw), &value) != nil {
		return nil
	}
	return value
}

func daoError(err error) common.Error {
	if err == nil {
		return nil
	}
	return common.NewDaoError(err.Error())
}

func timestamp(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: !value.IsZero()}
}

func nullableTime(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return timestamp(*value)
}

func timePointer(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}

func int32Value(value *int32) int32 {
	if value == nil {
		return 0
	}
	return *value
}

func instanceDTO(domain, instanceID, collectorID, hostname, version, commitSHA string, capabilities []string,
	startedAt, heartbeatAt, stoppedAt pgtype.Timestamptz, heartbeatAgeSeconds int64) collectionmodels.Instance {
	health := "online"
	if stoppedAt.Valid || heartbeatAgeSeconds > 120 {
		health = "offline"
	} else if heartbeatAgeSeconds >= 60 {
		health = "degraded"
	}
	return collectionmodels.Instance{Domain: domain, InstanceID: instanceID, CollectorID: collectorID,
		Hostname: hostname, Version: version, CommitSHA: commitSHA, Capabilities: capabilities,
		StartedAt: startedAt.Time, LastHeartbeatAt: heartbeatAt.Time, StoppedAt: timePointer(stoppedAt),
		Health: health, HeartbeatAgeSec: maxInt64(0, heartbeatAgeSeconds)}
}

func scheduleDTO(domain string, id int64, jobKey, name string, enabled bool, kind string, cronExpression *string,
	intervalSeconds *int64, anchorAt pgtype.Timestamptz, timezone, misfire string, grace int32,
	overlap string, priority int32, concurrency string, version int64, last, next pgtype.Timestamptz,
	lastStatus string, lastSuccess, lastExpected *int32) collectionmodels.Schedule {
	expected := int32Value(lastExpected)
	coverage := float64(0)
	if expected > 0 {
		coverage = float64(int32Value(lastSuccess)) / float64(expected)
	}
	return collectionmodels.Schedule{Domain: domain, ID: id, JobKey: jobKey, Name: name, Enabled: enabled,
		ScheduleKind: kind, CronExpression: cronExpression, IntervalSeconds: intervalSeconds,
		AnchorAt: timePointer(anchorAt), Timezone: timezone, MisfirePolicy: misfire,
		MisfireGraceSeconds: grace, OverlapPolicy: overlap, Priority: priority,
		ConcurrencyKey: concurrency, Version: version, LastMaterializedFor: timePointer(last),
		NextScheduledFor: timePointer(next), LastStatus: lastStatus,
		LastSuccessCount: int32Value(lastSuccess), LastExpectedCount: expected, LastSuccessCoverage: coverage}
}

func jobDTO(domain string, id int64, scheduleID *int64, jobKey, trigger, scopeType string, scopeID *int64,
	target *string, tasks []string, priority int32, scheduled pgtype.Timestamptz, status, requestedBy string,
	claimedBy *string, lease, cancel, created, completed pgtype.Timestamptz, runID string) collectionmodels.Job {
	return collectionmodels.Job{Domain: domain, ID: id, ScheduleID: scheduleID, JobKey: jobKey, Trigger: trigger,
		ScopeType: scopeType, ScopeID: scopeID, Target: target, Tasks: tasks, Priority: priority,
		ScheduledFor: timePointer(scheduled), Status: status, RequestedBy: requestedBy, ClaimedBy: claimedBy,
		LeaseUntil: timePointer(lease), CancelRequestedAt: timePointer(cancel), CreatedAt: created.Time,
		CompletedAt: timePointer(completed), RunID: runID}
}

func gameJobDTO(row gamesqlc.GfgCollectionJob) collectionmodels.Job {
	return jobDTO("game", row.ID, row.ScheduleID, row.JobKey, row.Trigger, row.ScopeType, row.ScopeID, row.Target,
		row.Tasks, row.Priority, row.ScheduledFor, row.Status, row.RequestedBy, row.ClaimedBy,
		row.LeaseUntil, row.CancelRequestedAt, row.CreatedAt, row.CompletedAt, "")
}

func navJobDTO(row navsqlc.GfnCollectionJob) collectionmodels.Job {
	return jobDTO("nav", row.ID, row.ScheduleID, row.JobKey, row.Trigger, row.ScopeType, row.ScopeID, row.Target,
		row.Tasks, row.Priority, row.ScheduledFor, row.Status, row.RequestedBy, row.ClaimedBy,
		row.LeaseUntil, row.CancelRequestedAt, row.CreatedAt, row.CompletedAt, "")
}

func runDTO(domain, id string, jobID int64, jobKey, trigger, scopeType string, scopeID *int64, target *string,
	attempt int32, instanceID, status string, scheduled, started, ended pgtype.Timestamptz,
	expected, attempted, success, partial, failed, skipped int32, delay, duration int64,
	errorKind, errorMessage string) collectionmodels.Run {
	return collectionmodels.Run{Domain: domain, ID: id, JobID: jobID, JobKey: jobKey, Trigger: trigger,
		ScopeType: scopeType, ScopeID: scopeID, Target: target, AttemptNo: attempt,
		CollectorInstanceID: instanceID, Status: status, ScheduledFor: timePointer(scheduled),
		StartedAt: started.Time, EndedAt: timePointer(ended), ExpectedCount: expected,
		AttemptedCount: attempted, SuccessCount: success, PartialCount: partial, FailureCount: failed,
		SkippedCount: skipped, ScheduleDelayMS: delay, DurationMS: duration,
		ErrorKind: errorKind, ErrorMessage: errorMessage}
}

func chartPoint(domain string, jobID int64, jobKey, jobStatus, runID, runStatus string,
	expected, attempted, success, partial, failed, skipped int32, delay, duration int64,
	created, started pgtype.Timestamptz) collectionmodels.ChartPoint {
	coverage := float64(0)
	if expected > 0 {
		coverage = float64(success) / float64(expected)
	}
	return collectionmodels.ChartPoint{Domain: domain, JobID: jobID, JobKey: jobKey, JobStatus: jobStatus,
		RunID: runID, RunStatus: runStatus, Expected: expected, Attempted: attempted, Success: success,
		Partial: partial, Failed: failed, Skipped: skipped, Coverage: coverage,
		ScheduleDelayMS: delay, DurationMS: duration, CreatedAt: created.Time, StartedAt: timePointer(started)}
}

func maxInt64(left, right int64) int64 {
	if left > right {
		return left
	}
	return right
}
