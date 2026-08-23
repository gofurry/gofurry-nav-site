package runner

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gofurry/gofurry-game-collector/collector/game/models"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/report"
)

// GameCollector runs one v2 task for one game.
type GameCollector interface {
	CollectGame(context.Context, models.GameID) (report.TaskResult, error)
}

// TaskBinding connects one task type to its collector.
type TaskBinding struct {
	Task      domain.TaskType
	Collector GameCollector
}

// Options controls one unified runner execution.
type Options struct {
	RunID      string
	MaxWorkers int
}

// Runner coordinates enabled v2 collectors over a batch of games.
type Runner struct {
	options Options
	tasks   []TaskBinding
}

// New creates one unified v2 runner.
func New(options Options, tasks []TaskBinding) *Runner {
	if options.MaxWorkers <= 0 {
		options.MaxWorkers = 1
	}
	return &Runner{
		options: options,
		tasks:   append([]TaskBinding(nil), tasks...),
	}
}

// Run executes every configured task for every game. A single app failure is
// captured in the report and does not stop the rest of the batch.
func (r *Runner) Run(ctx context.Context, games []models.GameID) (report.RunSummary, error) {
	startedAt := time.Now()
	runID := r.runID(startedAt)
	summary := report.RunSummary{
		ID:        runID,
		Status:    domain.StatusSuccess,
		StartedAt: startedAt,
	}

	if ctx == nil {
		ctx = context.Background()
	}
	if r == nil {
		return failRun(summary, report.ErrorValidation, "v2 runner is nil")
	}
	if len(games) == 0 {
		summary.Status = domain.StatusSkipped
		summary.EndedAt = time.Now()
		return summary, nil
	}
	if len(r.tasks) == 0 {
		return failRun(summary, report.ErrorValidation, "v2 runner has no enabled tasks")
	}

	jobs := make(chan runJob)
	results := make(chan report.TaskResult)
	var wg sync.WaitGroup

	for workerID := 0; workerID < r.options.MaxWorkers; workerID++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				results <- executeJob(ctx, runID, job)
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, game := range games {
			for _, task := range r.tasks {
				select {
				case <-ctx.Done():
					return
				case jobs <- runJob{game: game, task: task}:
				}
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		summary.Results = append(summary.Results, result)
		applyResult(&summary, result)
	}

	summary.EndedAt = time.Now()
	if err := ctx.Err(); err != nil {
		summary.Status = domain.StatusPartial
		summary.Error = &report.ErrorInfo{Kind: report.ErrorCanceled, Message: err.Error()}
		return summary, err
	}
	if summary.FailedCount > 0 || summary.PartialCount > 0 {
		summary.Status = domain.StatusPartial
	}
	return summary, nil
}

type runJob struct {
	game models.GameID
	task TaskBinding
}

func executeJob(ctx context.Context, runID string, job runJob) report.TaskResult {
	startedAt := time.Now()
	if job.task.Collector == nil {
		return report.TaskResult{
			RunID:          runID,
			Task:           job.task.Task,
			Status:         domain.StatusFailed,
			GameID:         job.game.ID,
			AppID:          uint32(job.game.Appid),
			StartedAt:      startedAt,
			EndedAt:        time.Now(),
			DurationMillis: time.Since(startedAt).Milliseconds(),
			Error:          &report.ErrorInfo{Kind: report.ErrorValidation, Message: "v2 task collector is nil"},
		}
	}
	result, err := job.task.Collector.CollectGame(report.ContextWithRunID(ctx, runID), job.game)
	result.RunID = runID
	result.Task = job.task.Task
	if result.GameID == 0 {
		result.GameID = job.game.ID
	}
	if result.AppID == 0 {
		result.AppID = uint32(job.game.Appid)
	}
	if err != nil && result.Error == nil {
		result.Error = &report.ErrorInfo{Kind: report.ErrorUnknown, Message: err.Error()}
	}
	if result.Status == "" {
		if err != nil {
			result.Status = domain.StatusFailed
		} else {
			result.Status = domain.StatusSuccess
		}
	}
	if result.StartedAt.IsZero() {
		result.StartedAt = startedAt
	}
	if result.EndedAt.IsZero() {
		result.EndedAt = time.Now()
	}
	if result.DurationMillis == 0 {
		result.DurationMillis = result.EndedAt.Sub(result.StartedAt).Milliseconds()
	}
	return result
}

func applyResult(summary *report.RunSummary, result report.TaskResult) {
	summary.TotalCount++
	switch result.Status {
	case domain.StatusSuccess:
		summary.SuccessCount++
	case domain.StatusSkipped:
		summary.SkippedCount++
	case domain.StatusPartial:
		summary.PartialCount++
	default:
		summary.FailedCount++
	}

	idx := -1
	for i := range summary.TaskSummaries {
		if summary.TaskSummaries[i].Task == result.Task {
			idx = i
			break
		}
	}
	if idx == -1 {
		summary.TaskSummaries = append(summary.TaskSummaries, report.TaskSummary{Task: result.Task})
		idx = len(summary.TaskSummaries) - 1
	}

	task := &summary.TaskSummaries[idx]
	task.TotalCount++
	task.DurationMillis += result.DurationMillis
	switch result.Status {
	case domain.StatusSuccess:
		task.SuccessCount++
	case domain.StatusSkipped:
		task.SkippedCount++
	case domain.StatusPartial:
		task.PartialCount++
	default:
		task.FailedCount++
	}
}

func failRun(summary report.RunSummary, kind report.ErrorKind, message string) (report.RunSummary, error) {
	summary.Status = domain.StatusFailed
	summary.EndedAt = time.Now()
	summary.Error = &report.ErrorInfo{Kind: kind, Message: message}
	return summary, errors.New(message)
}

func (r *Runner) runID(startedAt time.Time) string {
	if r != nil && r.options.RunID != "" {
		return r.options.RunID
	}
	return fmt.Sprintf("game-v2-%s-%s-%s", r.taskLabel(), startedAt.Format("20060102-150405"), randomSuffix())
}

func (r *Runner) taskLabel() string {
	if r == nil || len(r.tasks) == 0 {
		return "unknown"
	}
	parts := make([]string, 0, len(r.tasks))
	for _, task := range r.tasks {
		if task.Task != "" {
			parts = append(parts, string(task.Task))
		}
	}
	if len(parts) == 0 {
		return "unknown"
	}
	return strings.Join(parts, "-")
}

func randomSuffix() string {
	var buf [4]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "nosuffix"
	}
	return hex.EncodeToString(buf[:])
}
