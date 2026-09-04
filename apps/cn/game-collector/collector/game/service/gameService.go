package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofurry/gofurry-game-collector/collector/game/models"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/report"
	v2repo "github.com/gofurry/gofurry-game-collector/collector/game/v2/repository"
	v2runner "github.com/gofurry/gofurry-game-collector/collector/game/v2/runner"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/steamclient"
	v2details "github.com/gofurry/gofurry-game-collector/collector/game/v2/tasks/details"
	v2news "github.com/gofurry/gofurry-game-collector/collector/game/v2/tasks/news"
	v2players "github.com/gofurry/gofurry-game-collector/collector/game/v2/tasks/players"
	"github.com/gofurry/gofurry-game-collector/common"
	"github.com/gofurry/gofurry-game-collector/common/log"
	gamesqlc "github.com/gofurry/gofurry-game-collector/internal/db/game/sqlc"
	"github.com/gofurry/gofurry-game-collector/roof/env"
	"github.com/jackc/pgx/v5/pgxpool"
)

type gameService struct {
	pool    *pgxpool.Pool
	queries *gamesqlc.Queries
}

var gameSingleton = new(gameService)

func GetGameService() *gameService { return gameSingleton }

// InitPersistence injects the process-owned PostgreSQL pool during bootstrap.
func InitPersistence(pool *pgxpool.Pool) {
	gameSingleton.pool = pool
	if pool == nil {
		gameSingleton.queries = nil
		return
	}
	gameSingleton.queries = gamesqlc.New(pool)
}

var v2SteamAdapter *steamclient.Adapter

// InitLimiter keeps the public initialization hook stable while v2 owns all Steam limits.
func InitLimiter() {
	InitV2SteamAdapter()
}

// InitV2SteamAdapter initializes the collector v2 Steam client.
func InitV2SteamAdapter() {
	v2Cfg := env.GetServerConfig().Collector.V2
	adapter, err := steamclient.New(steamclient.Config{
		Proxy:                 env.GetServerConfig().Collector.Proxy,
		APIRequestsPer5Min:    v2Cfg.Steam.APIRequestsPer5Minutes,
		StoreRequestsPer5Min:  v2Cfg.Steam.StoreRequestsPer5Minutes,
		Burst:                 v2Cfg.Steam.Burst,
		MaxWorkers:            v2Cfg.Steam.MaxWorkers,
		RequestTimeoutSeconds: v2Cfg.Steam.RequestTimeoutSeconds,
		Retry: steamclient.RetryConfig{
			MaxAttempts:          v2Cfg.Steam.Retry.MaxAttempts,
			BaseDelaySeconds:     v2Cfg.Steam.Retry.BaseDelaySeconds,
			CooldownOn429Seconds: v2Cfg.Steam.Retry.CooldownOn429Seconds,
		},
	})
	if err != nil {
		log.Error("init game collector v2 steam adapter failed: ", err)
		return
	}
	adapter.SetObserver(steamclient.ObserverFunc(observeV2SteamRequest))

	if v2SteamAdapter != nil {
		v2SteamAdapter.Close()
	}
	v2SteamAdapter = adapter
	log.Info("game collector v2 steam adapter initialized")
}

func observeV2SteamRequest(event steamclient.Event) {
	if !steamRequestNeedsDiagnostic(event) {
		return
	}
	log.WarnFields(steamRequestDiagnosticFields(event), "Steam request degraded")
}

func steamRequestNeedsDiagnostic(event steamclient.Event) bool {
	return event.ErrorKind != "" || event.BlockDetected || !event.CooldownUntil.IsZero() ||
		event.StatusCode == 403 || event.StatusCode == 429 || event.StatusCode >= 500
}

func steamRequestDiagnosticFields(event steamclient.Event) map[string]interface{} {
	cooldownUntil := ""
	if !event.CooldownUntil.IsZero() {
		cooldownUntil = event.CooldownUntil.UTC().Format(time.RFC3339)
	}
	return map[string]interface{}{
		"bucket":         string(event.Bucket),
		"traffic_class":  string(event.TrafficClass),
		"method":         event.Method,
		"host":           sanitizeSteamObserverHost(event.Host),
		"path":           sanitizeSteamObserverPath(event.Path),
		"status_code":    event.StatusCode,
		"error_kind":     event.ErrorKind,
		"attempts":       event.Attempts,
		"block_detected": event.BlockDetected,
		"duration":       event.Duration,
		"cooldown_until": cooldownUntil,
	}
}

func sanitizeSteamObserverHost(host string) string {
	host = strings.TrimSpace(host)
	if index := strings.LastIndex(host, "@"); index >= 0 {
		host = host[index+1:]
	}
	if index := strings.IndexAny(host, "/?#"); index >= 0 {
		host = host[:index]
	}
	return host
}

func sanitizeSteamObserverPath(path string) string {
	path = strings.TrimSpace(path)
	if index := strings.IndexAny(path, "?#"); index >= 0 {
		path = path[:index]
	}
	if path == "" {
		return "/"
	}
	return path
}

// GetV2SteamAdapter returns the initialized collector v2 Steam adapter.
func GetV2SteamAdapter() *steamclient.Adapter {
	return v2SteamAdapter
}

func (s *gameService) runV2Tasks(
	ctx context.Context,
	gameList []models.GameID,
	tasks []domain.TaskType,
	runID string,
	onResult func(report.TaskResult),
) (report.RunSummary, error) {
	bindings := make([]v2runner.TaskBinding, 0, len(tasks))
	for _, task := range tasks {
		switch task {
		case domain.TaskDetails:
			bindings = append(bindings, v2runner.TaskBinding{
				Task:      domain.TaskDetails,
				Collector: v2details.NewCollector(GetV2SteamAdapter(), v2repo.NewDetailsRepository(s.pool)),
			})
		case domain.TaskNews:
			bindings = append(bindings, v2runner.TaskBinding{
				Task:      domain.TaskNews,
				Collector: v2news.NewCollector(GetV2SteamAdapter(), v2repo.NewNewsRepository(s.pool)),
			})
		case domain.TaskPlayers:
			bindings = append(bindings, v2runner.TaskBinding{
				Task:      domain.TaskPlayers,
				Collector: v2players.NewCollector(GetV2SteamAdapter(), v2repo.NewPlayerRepository(s.pool)),
			})
		}
	}

	maxWorkers := env.GetServerConfig().Collector.V2.Steam.MaxWorkers
	r := v2runner.New(v2runner.Options{RunID: runID, MaxWorkers: maxWorkers, OnResult: onResult}, bindings)
	return r.Run(ctx, gameList)
}

// ResolveControlTargets freezes the target set for one durable Job.
func (s *gameService) ResolveControlTargets(ctx context.Context, gameID *int64) ([]models.GameID, error) {
	if gameID == nil {
		rows, err := s.addAllGameToList(ctx)
		if err != nil {
			return nil, fmt.Errorf("list Game targets: %s", err.GetMsg())
		}
		return rows, nil
	}
	if s == nil || s.queries == nil {
		return nil, fmt.Errorf("game persistence is not initialized")
	}
	row, err := s.queries.GetGameTarget(ctx, *gameID)
	if err != nil {
		return nil, fmt.Errorf("get Game target %d: %w", *gameID, err)
	}
	return []models.GameID{{ID: row.ID, Appid: row.Appid}}, nil
}

// RunControlJob executes a target/task snapshot owned by the durable control plane.
func (s *gameService) RunControlJob(
	ctx context.Context,
	targets []models.GameID,
	taskNames []string,
	runID string,
	onResult func(report.TaskResult),
) (report.RunSummary, error) {
	tasks := make([]domain.TaskType, 0, len(taskNames))
	for _, task := range taskNames {
		switch domain.TaskType(task) {
		case domain.TaskDetails, domain.TaskNews, domain.TaskPlayers:
			tasks = append(tasks, domain.TaskType(task))
		default:
			return report.RunSummary{ID: runID}, fmt.Errorf("unsupported Game collection task %q", task)
		}
	}
	return s.runV2Tasks(ctx, targets, tasks, runID, onResult)
}

func logV2RunSummary(prefix string, summary report.RunSummary, err error) {
	if err != nil {
		log.Error(prefix, " failed, run_id=", summary.ID, " status=", summary.Status, " total=", summary.TotalCount, " success=", summary.SuccessCount, " partial=", summary.PartialCount, " failed=", summary.FailedCount, " skipped=", summary.SkippedCount, " err=", err)
	} else {
		log.Info(prefix, " finished, run_id=", summary.ID, " status=", summary.Status, " total=", summary.TotalCount, " success=", summary.SuccessCount, " partial=", summary.PartialCount, " failed=", summary.FailedCount, " skipped=", summary.SkippedCount)
	}
	for _, task := range summary.TaskSummaries {
		log.Info(prefix, " task summary, run_id=", summary.ID, " task=", task.Task, " total=", task.TotalCount, " success=", task.SuccessCount, " partial=", task.PartialCount, " failed=", task.FailedCount, " skipped=", task.SkippedCount, " duration_ms=", task.DurationMillis)
	}
}

// Collect runs the stable v2 details and news collectors.
func (s gameService) Collect() {
	ctx := context.Background()
	gameList, err := s.addAllGameToList(ctx)
	if err != nil {
		log.Error("receive InitGameCollection recover: ", err)
	}

	log.Info("Game Collect v2 采集开始")
	summary, runErr := s.runV2Tasks(ctx, gameList, []domain.TaskType{domain.TaskDetails, domain.TaskNews}, "", nil)
	logV2RunSummary("Game Collect v2", summary, runErr)
	log.Info("Game Collect v2 采集结束")
}

// CollectCurrentPlayers runs the stable v2 current-player collector.
func (s gameService) CollectCurrentPlayers() {
	ctx := context.Background()
	gameList, err := s.addAllGameToList(ctx)
	if err != nil {
		log.Error("receive InitGameCollection recover: ", err)
	}

	log.Info("CollectCurrentPlayers v2 采集开始")
	summary, runErr := s.runV2Tasks(ctx, gameList, []domain.TaskType{domain.TaskPlayers}, "", nil)
	logV2RunSummary("CollectCurrentPlayers v2", summary, runErr)
	log.Info("CollectCurrentPlayers v2 采集结束")
}

func singleGameTaskTypes() []domain.TaskType {
	return []domain.TaskType{domain.TaskDetails, domain.TaskNews}
}

func (s gameService) CollectSingleGame(game models.GameID) (report.RunSummary, error) {
	ctx := context.Background()
	log.Info("SingleGame Collect v2 采集开始, game_id=", game.ID, " appid=", game.Appid)
	summary, runErr := s.runV2Tasks(ctx, []models.GameID{game}, singleGameTaskTypes(), "", nil)
	logV2RunSummary("SingleGame Collect v2", summary, runErr)
	log.Info("SingleGame Collect v2 采集结束, game_id=", game.ID, " appid=", game.Appid)
	return summary, runErr
}

func (s *gameService) addAllGameToList(ctx context.Context) ([]models.GameID, common.GFError) {
	if s == nil || s.queries == nil {
		return nil, common.NewDaoError("game persistence is not initialized")
	}
	rows, err := s.queries.ListGameTargets(ctx)
	if err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	games := make([]models.GameID, 0, len(rows))
	for _, row := range rows {
		games = append(games, models.GameID{ID: row.ID, Appid: row.Appid})
	}
	return games, nil
}
