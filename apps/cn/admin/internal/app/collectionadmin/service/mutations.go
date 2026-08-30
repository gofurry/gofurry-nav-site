package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	collectionmodels "github.com/gofurry/gofurry-admin/internal/app/collectionadmin/models"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	gamesqlc "github.com/gofurry/gofurry-admin/internal/db/game/sqlc"
	navsqlc "github.com/gofurry/gofurry-admin/internal/db/nav/sqlc"
	"github.com/gofurry/gofurry-admin/pkg/common"
	"github.com/robfig/cron/v3"
)

var gameTasks = map[string]struct{}{"details": {}, "news": {}, "players": {}}
var navTasks = map[string]struct{}{
	"ping": {}, "http": {}, "dns": {}, "rdap": {}, "robots": {},
	"security_txt": {}, "llms_txt": {}, "page_assets": {}, "port_check": {}, "waf_canary": {},
}

func (s *Service) UpdateSchedule(ctx context.Context, meta audit.Meta, domain string, id int64, request collectionmodels.ScheduleUpdate) (collectionmodels.Schedule, common.Error) {
	if err := validateScheduleUpdate(request); err != nil {
		return collectionmodels.Schedule{}, err
	}
	if request.ScheduleKind == "cron" {
		request.IntervalSeconds, request.AnchorAt = nil, nil
	} else {
		request.CronExpression = nil
	}
	anchor := nullableTime(request.AnchorAt)
	switch domain {
	case "game":
		tx, err := s.gamePool.Begin(ctx)
		if err != nil {
			return collectionmodels.Schedule{}, daoError(err)
		}
		defer tx.Rollback(ctx)
		queries := s.game.WithTx(tx)
		before, err := queries.AdminGetGameCollectionSchedule(ctx, id)
		if err != nil {
			return collectionmodels.Schedule{}, daoError(err)
		}
		after, err := queries.AdminUpdateGameCollectionSchedule(ctx, gamesqlc.AdminUpdateGameCollectionScheduleParams{
			Enabled: request.Enabled, ScheduleKind: request.ScheduleKind, CronExpression: request.CronExpression,
			IntervalSeconds: request.IntervalSeconds, AnchorAt: anchor, Timezone: request.Timezone,
			MisfirePolicy: request.MisfirePolicy, MisfireGraceSeconds: request.MisfireGraceSeconds, ID: id,
		})
		if err != nil {
			return collectionmodels.Schedule{}, daoError(err)
		}
		if auditErr := s.audit.Log(ctx, meta, "collection.schedule.update", "game.collection_schedule", id, before, after); auditErr != nil {
			return collectionmodels.Schedule{}, auditErr
		}
		if err := tx.Commit(ctx); err != nil {
			return collectionmodels.Schedule{}, daoError(err)
		}
		return gameScheduleModel(after), nil
	case "nav":
		tx, err := s.navPool.Begin(ctx)
		if err != nil {
			return collectionmodels.Schedule{}, daoError(err)
		}
		defer tx.Rollback(ctx)
		queries := s.nav.WithTx(tx)
		before, err := queries.AdminGetNavCollectionSchedule(ctx, id)
		if err != nil {
			return collectionmodels.Schedule{}, daoError(err)
		}
		after, err := queries.AdminUpdateNavCollectionSchedule(ctx, navsqlc.AdminUpdateNavCollectionScheduleParams{
			Enabled: request.Enabled, ScheduleKind: request.ScheduleKind, CronExpression: request.CronExpression,
			IntervalSeconds: request.IntervalSeconds, AnchorAt: anchor, Timezone: request.Timezone,
			MisfirePolicy: request.MisfirePolicy, MisfireGraceSeconds: request.MisfireGraceSeconds, ID: id,
		})
		if err != nil {
			return collectionmodels.Schedule{}, daoError(err)
		}
		if auditErr := s.audit.Log(ctx, meta, "collection.schedule.update", "nav.collection_schedule", id, before, after); auditErr != nil {
			return collectionmodels.Schedule{}, auditErr
		}
		if err := tx.Commit(ctx); err != nil {
			return collectionmodels.Schedule{}, daoError(err)
		}
		return navScheduleModel(after), nil
	default:
		return collectionmodels.Schedule{}, common.NewValidationError("domain must be game or nav")
	}
}

func (s *Service) RunScheduleNow(ctx context.Context, meta audit.Meta, domain string, id int64) ([]collectionmodels.Job, common.Error) {
	switch domain {
	case "game":
		schedule, err := s.game.AdminGetGameCollectionSchedule(ctx, id)
		if err != nil {
			return nil, daoError(err)
		}
		tasks := []string{"details", "news"}
		if schedule.JobKey == "game.players" {
			tasks = []string{"players"}
		} else if schedule.JobKey != "game.metadata" {
			return nil, common.NewValidationError("unknown Game schedule capability")
		}
		return s.createGameJobs(ctx, meta, collectionmodels.ManualJobRequest{Domain: "game", ScopeType: "all", Tasks: tasks}, "collection.schedule.run", id, &schedule.ID, &schedule.Version)
	case "nav":
		schedule, err := s.nav.AdminGetNavCollectionSchedule(ctx, id)
		if err != nil {
			return nil, daoError(err)
		}
		protocol := strings.TrimPrefix(schedule.JobKey, "nav.")
		if _, ok := navTasks[protocol]; !ok {
			return nil, common.NewValidationError("unknown Nav schedule capability")
		}
		return s.createNavJobs(ctx, meta, collectionmodels.ManualJobRequest{Domain: "nav", ScopeType: "all", Tasks: []string{protocol}}, "collection.schedule.run", id, &schedule.ID, &schedule.Version)
	default:
		return nil, common.NewValidationError("domain must be game or nav")
	}
}

func (s *Service) CreateManualJobs(ctx context.Context, meta audit.Meta, request collectionmodels.ManualJobRequest) ([]collectionmodels.Job, common.Error) {
	switch request.Domain {
	case "game":
		return s.createGameJobs(ctx, meta, request, "collection.job.create", nil, nil, nil)
	case "nav":
		return s.createNavJobs(ctx, meta, request, "collection.job.create", nil, nil, nil)
	default:
		return nil, common.NewValidationError("domain must be game or nav")
	}
}

func (s *Service) createGameJobs(ctx context.Context, meta audit.Meta, request collectionmodels.ManualJobRequest, action string, targetID any, scheduleID, scheduleVersion *int64) ([]collectionmodels.Job, common.Error) {
	tasks, validationErr := normalizeKnownTasks(request.Tasks, gameTasks)
	if validationErr != nil {
		return nil, validationErr
	}
	if request.ScopeType != "all" && request.ScopeType != "game" {
		return nil, common.NewValidationError("Game scope_type must be all or game")
	}
	if request.ScopeType == "all" {
		request.ScopeID = nil
	} else if request.ScopeID == nil || *request.ScopeID <= 0 {
		return nil, common.NewValidationError("Game scope_id is required")
	} else {
		exists, err := s.game.AdminGameCollectionTargetExists(ctx, *request.ScopeID)
		if err != nil {
			return nil, daoError(err)
		}
		if !exists {
			return nil, common.NewValidationError("Game target does not exist")
		}
	}
	jobKey := "game.metadata"
	if len(tasks) == 1 && tasks[0] == "players" {
		jobKey = "game.players"
	}
	dedupe := manualDedupe("game", request.ScopeType, request.ScopeID, nil, tasks)
	if scheduleID != nil && scheduleVersion != nil {
		dedupe = fmt.Sprintf("game:schedule:%d:v%d:run-now", *scheduleID, *scheduleVersion)
	}
	tx, err := s.gamePool.Begin(ctx)
	if err != nil {
		return nil, daoError(err)
	}
	defer tx.Rollback(ctx)
	row, err := s.game.WithTx(tx).AdminInsertGameManualJob(ctx, gamesqlc.AdminInsertGameManualJobParams{
		ScheduleID: scheduleID, ScheduleVersion: scheduleVersion,
		JobKey: jobKey, ScopeType: request.ScopeType, ScopeID: request.ScopeID,
		Tasks: tasks, RequestedBy: meta.Operator, DedupeKey: &dedupe,
	})
	if err != nil {
		return nil, daoError(err)
	}
	if targetID == nil {
		targetID = row.ID
	}
	if auditErr := s.audit.Log(ctx, meta, action, "game.collection_job", targetID, nil, row); auditErr != nil {
		return nil, auditErr
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, daoError(err)
	}
	return []collectionmodels.Job{gameJobDTO(row)}, nil
}

func (s *Service) createNavJobs(ctx context.Context, meta audit.Meta, request collectionmodels.ManualJobRequest, action string, targetID any, scheduleID, scheduleVersion *int64) ([]collectionmodels.Job, common.Error) {
	tasks, validationErr := normalizeKnownTasks(request.Tasks, navTasks)
	if validationErr != nil {
		return nil, validationErr
	}
	if request.ScopeType != "all" && request.ScopeType != "site" && request.ScopeType != "target" {
		return nil, common.NewValidationError("Nav scope_type must be all, site, or target")
	}
	if request.ScopeType == "all" {
		request.ScopeID, request.Target = nil, nil
	} else if request.ScopeID == nil || *request.ScopeID <= 0 {
		return nil, common.NewValidationError("Nav scope_id is required")
	} else {
		exists, err := s.nav.AdminNavCollectionSiteExists(ctx, *request.ScopeID)
		if err != nil {
			return nil, daoError(err)
		}
		if !exists {
			return nil, common.NewValidationError("Nav site does not exist")
		}
		if request.ScopeType == "target" {
			if request.Target == nil || strings.TrimSpace(*request.Target) == "" {
				return nil, common.NewValidationError("Nav target is required")
			}
			targetExists, err := s.nav.AdminNavCollectionTargetExists(ctx, navsqlc.AdminNavCollectionTargetExistsParams{SiteID: request.ScopeID, Target: *request.Target})
			if err != nil {
				return nil, daoError(err)
			}
			if !targetExists {
				return nil, common.NewValidationError("Nav target does not exist")
			}
		} else {
			request.Target = nil
		}
	}
	tx, err := s.navPool.Begin(ctx)
	if err != nil {
		return nil, daoError(err)
	}
	defer tx.Rollback(ctx)
	queries := s.nav.WithTx(tx)
	created := make([]navsqlc.GfnCollectionJob, 0, len(tasks))
	for _, protocol := range tasks {
		dedupe := manualDedupe("nav", request.ScopeType, request.ScopeID, request.Target, []string{protocol})
		if scheduleID != nil && scheduleVersion != nil {
			dedupe = fmt.Sprintf("nav:schedule:%d:v%d:run-now:%s", *scheduleID, *scheduleVersion, protocol)
		}
		row, err := queries.AdminInsertNavManualJob(ctx, navsqlc.AdminInsertNavManualJobParams{
			ScheduleID: scheduleID, ScheduleVersion: scheduleVersion,
			JobKey: "nav." + protocol, ScopeType: request.ScopeType, ScopeID: request.ScopeID,
			Target: request.Target, Protocol: protocol, RequestedBy: meta.Operator, DedupeKey: &dedupe,
		})
		if err != nil {
			return nil, daoError(err)
		}
		created = append(created, row)
	}
	ids := make([]int64, 0, len(created))
	for _, row := range created {
		ids = append(ids, row.ID)
	}
	if targetID == nil {
		targetID = fmt.Sprint(ids)
	}
	if auditErr := s.audit.Log(ctx, meta, action, "nav.collection_job", targetID, nil, created); auditErr != nil {
		return nil, auditErr
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, daoError(err)
	}
	result := make([]collectionmodels.Job, 0, len(created))
	for _, row := range created {
		result = append(result, navJobDTO(row))
	}
	return result, nil
}

func (s *Service) CancelJob(ctx context.Context, meta audit.Meta, domain string, id int64) (collectionmodels.Job, common.Error) {
	switch domain {
	case "game":
		tx, err := s.gamePool.Begin(ctx)
		if err != nil {
			return collectionmodels.Job{}, daoError(err)
		}
		defer tx.Rollback(ctx)
		queries := s.game.WithTx(tx)
		before, err := queries.AdminGetGameCollectionJob(ctx, id)
		if err != nil {
			return collectionmodels.Job{}, daoError(err)
		}
		after, err := queries.AdminCancelGameCollectionJob(ctx, id)
		if err != nil {
			return collectionmodels.Job{}, daoError(err)
		}
		if auditErr := s.audit.Log(ctx, meta, "collection.job.cancel", "game.collection_job", id, before, after); auditErr != nil {
			return collectionmodels.Job{}, auditErr
		}
		if err := tx.Commit(ctx); err != nil {
			return collectionmodels.Job{}, daoError(err)
		}
		return gameJobDTO(after), nil
	case "nav":
		tx, err := s.navPool.Begin(ctx)
		if err != nil {
			return collectionmodels.Job{}, daoError(err)
		}
		defer tx.Rollback(ctx)
		queries := s.nav.WithTx(tx)
		before, err := queries.AdminGetNavCollectionJob(ctx, id)
		if err != nil {
			return collectionmodels.Job{}, daoError(err)
		}
		after, err := queries.AdminCancelNavCollectionJob(ctx, id)
		if err != nil {
			return collectionmodels.Job{}, daoError(err)
		}
		if auditErr := s.audit.Log(ctx, meta, "collection.job.cancel", "nav.collection_job", id, before, after); auditErr != nil {
			return collectionmodels.Job{}, auditErr
		}
		if err := tx.Commit(ctx); err != nil {
			return collectionmodels.Job{}, daoError(err)
		}
		return navJobDTO(after), nil
	default:
		return collectionmodels.Job{}, common.NewValidationError("domain must be game or nav")
	}
}

func (s *Service) RetryJob(ctx context.Context, meta audit.Meta, domain string, id int64) (collectionmodels.Job, common.Error) {
	switch domain {
	case "game":
		tx, err := s.gamePool.Begin(ctx)
		if err != nil {
			return collectionmodels.Job{}, daoError(err)
		}
		defer tx.Rollback(ctx)
		queries := s.game.WithTx(tx)
		before, err := queries.AdminGetGameCollectionJob(ctx, id)
		if err != nil {
			return collectionmodels.Job{}, daoError(err)
		}
		if containsTask(before.Tasks, "players") {
			return collectionmodels.Job{}, common.NewValidationError("point-in-time Game players jobs cannot be retried; use Run Now")
		}
		after, err := queries.AdminRetryGameCollectionJob(ctx, gamesqlc.AdminRetryGameCollectionJobParams{RequestedBy: meta.Operator, ID: id})
		if err != nil {
			return collectionmodels.Job{}, daoError(err)
		}
		if auditErr := s.audit.Log(ctx, meta, "collection.job.retry", "game.collection_job", id, before, after); auditErr != nil {
			return collectionmodels.Job{}, auditErr
		}
		if err := tx.Commit(ctx); err != nil {
			return collectionmodels.Job{}, daoError(err)
		}
		return gameJobDTO(after), nil
	case "nav":
		tx, err := s.navPool.Begin(ctx)
		if err != nil {
			return collectionmodels.Job{}, daoError(err)
		}
		defer tx.Rollback(ctx)
		queries := s.nav.WithTx(tx)
		before, err := queries.AdminGetNavCollectionJob(ctx, id)
		if err != nil {
			return collectionmodels.Job{}, daoError(err)
		}
		for _, task := range []string{"ping", "http", "dns", "port_check"} {
			if containsTask(before.Tasks, task) {
				return collectionmodels.Job{}, common.NewValidationError("point-in-time Nav jobs cannot be retried; use Run Now")
			}
		}
		after, err := queries.AdminRetryNavCollectionJob(ctx, navsqlc.AdminRetryNavCollectionJobParams{RequestedBy: meta.Operator, ID: id})
		if err != nil {
			return collectionmodels.Job{}, daoError(err)
		}
		if auditErr := s.audit.Log(ctx, meta, "collection.job.retry", "nav.collection_job", id, before, after); auditErr != nil {
			return collectionmodels.Job{}, auditErr
		}
		if err := tx.Commit(ctx); err != nil {
			return collectionmodels.Job{}, daoError(err)
		}
		return navJobDTO(after), nil
	default:
		return collectionmodels.Job{}, common.NewValidationError("domain must be game or nav")
	}
}

func validateScheduleUpdate(request collectionmodels.ScheduleUpdate) common.Error {
	request.ScheduleKind = strings.TrimSpace(request.ScheduleKind)
	request.Timezone = strings.TrimSpace(request.Timezone)
	if request.Timezone == "" {
		return common.NewValidationError("timezone is required")
	}
	if _, err := time.LoadLocation(request.Timezone); err != nil {
		return common.NewValidationError("invalid timezone")
	}
	if request.MisfirePolicy != "skip" && request.MisfirePolicy != "catch_up_once" {
		return common.NewValidationError("misfire_policy must be skip or catch_up_once")
	}
	if request.MisfireGraceSeconds < 0 {
		return common.NewValidationError("misfire_grace_seconds must not be negative")
	}
	switch request.ScheduleKind {
	case "cron":
		if request.CronExpression == nil || strings.TrimSpace(*request.CronExpression) == "" {
			return common.NewValidationError("cron_expression is required")
		}
		if _, err := cron.ParseStandard(*request.CronExpression); err != nil {
			return common.NewValidationError("invalid cron_expression")
		}
		request.IntervalSeconds, request.AnchorAt = nil, nil
	case "interval":
		if request.IntervalSeconds == nil || *request.IntervalSeconds <= 0 || request.AnchorAt == nil {
			return common.NewValidationError("interval_seconds and anchor_at are required")
		}
		request.CronExpression = nil
	default:
		return common.NewValidationError("schedule_kind must be cron or interval")
	}
	return nil
}

func normalizeKnownTasks(tasks []string, known map[string]struct{}) ([]string, common.Error) {
	unique := make(map[string]struct{}, len(tasks))
	for _, task := range tasks {
		task = strings.TrimSpace(task)
		if _, ok := known[task]; !ok {
			return nil, common.NewValidationError("unknown collection task: " + task)
		}
		unique[task] = struct{}{}
	}
	if len(unique) == 0 {
		return nil, common.NewValidationError("at least one task is required")
	}
	result := make([]string, 0, len(unique))
	for task := range unique {
		result = append(result, task)
	}
	sort.Strings(result)
	return result, nil
}

func manualDedupe(domain, scope string, scopeID *int64, target *string, tasks []string) string {
	entity := "all"
	if scopeID != nil {
		entity = fmt.Sprint(*scopeID)
	}
	if target != nil {
		entity += ":" + strings.ToLower(strings.TrimSpace(*target))
	}
	return strings.Join([]string{domain, scope, entity, strings.Join(tasks, ",")}, ":")
}

func containsTask(tasks []string, expected string) bool {
	for _, task := range tasks {
		if task == expected {
			return true
		}
	}
	return false
}

func gameScheduleModel(row gamesqlc.GfgCollectionSchedule) collectionmodels.Schedule {
	return scheduleDTO("game", row.ID, row.JobKey, row.Name, row.Enabled, row.ScheduleKind,
		row.CronExpression, row.IntervalSeconds, row.AnchorAt, row.Timezone, row.MisfirePolicy,
		row.MisfireGraceSeconds, row.OverlapPolicy, row.Priority, row.ConcurrencyKey, row.Version,
		row.LastMaterializedFor, row.NextScheduledFor, "", nil, nil)
}

func navScheduleModel(row navsqlc.GfnCollectionSchedule) collectionmodels.Schedule {
	return scheduleDTO("nav", row.ID, row.JobKey, row.Name, row.Enabled, row.ScheduleKind,
		row.CronExpression, row.IntervalSeconds, row.AnchorAt, row.Timezone, row.MisfirePolicy,
		row.MisfireGraceSeconds, row.OverlapPolicy, row.Priority, row.ConcurrencyKey, row.Version,
		row.LastMaterializedFor, row.NextScheduledFor, "", nil, nil)
}
