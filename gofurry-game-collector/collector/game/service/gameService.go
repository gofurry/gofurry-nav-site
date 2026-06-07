package service

import (
	"context"

	"github.com/gofurry/gofurry-game-collector/collector/game/dao"
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
	"github.com/gofurry/gofurry-game-collector/roof/env"
)

type gameService struct{}

var gameSingleton = new(gameService)

func GetGameService() *gameService { return gameSingleton }

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

	if v2SteamAdapter != nil {
		v2SteamAdapter.Close()
	}
	v2SteamAdapter = adapter
	log.Info("game collector v2 steam adapter initialized")
}

// GetV2SteamAdapter returns the initialized collector v2 Steam adapter.
func GetV2SteamAdapter() *steamclient.Adapter {
	return v2SteamAdapter
}

func runV2Tasks(ctx context.Context, gameList []models.GameID, tasks []domain.TaskType) (report.RunSummary, error) {
	bindings := make([]v2runner.TaskBinding, 0, len(tasks))
	for _, task := range tasks {
		switch task {
		case domain.TaskDetails:
			bindings = append(bindings, v2runner.TaskBinding{
				Task:      domain.TaskDetails,
				Collector: v2details.NewCollector(GetV2SteamAdapter(), v2repo.NewDetailsRepository()),
			})
		case domain.TaskNews:
			bindings = append(bindings, v2runner.TaskBinding{
				Task:      domain.TaskNews,
				Collector: v2news.NewCollector(GetV2SteamAdapter(), v2repo.NewNewsRepository()),
			})
		case domain.TaskPlayers:
			bindings = append(bindings, v2runner.TaskBinding{
				Task:      domain.TaskPlayers,
				Collector: v2players.NewCollector(GetV2SteamAdapter(), v2repo.NewPlayerRepository()),
			})
		}
	}

	maxWorkers := env.GetServerConfig().Collector.V2.Steam.MaxWorkers
	r := v2runner.New(v2runner.Options{MaxWorkers: maxWorkers}, bindings)
	return r.Run(ctx, gameList)
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

func persistV2RunSummary(ctx context.Context, prefix string, summary report.RunSummary) {
	if summary.ID == "" {
		return
	}
	if err := v2repo.NewRunRepository().SaveRunSummary(ctx, summary); err != nil {
		log.Error(prefix, " observation persist failed, run_id=", summary.ID, " err=", err)
		return
	}
	retention := env.GetServerConfig().Collector.V2.Retention
	if err := v2repo.NewRetentionRepository().Prune(ctx, v2repo.RetentionConfig{
		PlayerCountsDays:       retention.PlayerCountsDays,
		CollectRunsDays:        retention.CollectRunsDays,
		CollectTaskResultsDays: retention.CollectTaskResultsDays,
	}); err != nil {
		log.Error(prefix, " retention prune failed, run_id=", summary.ID, " err=", err)
	}
}

// Collect runs the stable v2 details and news collectors.
func (s gameService) Collect() {
	ctx := context.Background()
	gameList, err := addAllGameToList()
	if err != nil {
		log.Error("receive InitGameCollection recover: ", err)
	}

	log.Info("Game Collect v2 采集开始")
	summary, runErr := runV2Tasks(ctx, gameList, []domain.TaskType{domain.TaskDetails, domain.TaskNews})
	logV2RunSummary("Game Collect v2", summary, runErr)
	persistV2RunSummary(ctx, "Game Collect v2", summary)
	log.Info("Game Collect v2 采集结束")
}

// CollectCurrentPlayers runs the stable v2 current-player collector.
func (s gameService) CollectCurrentPlayers() {
	ctx := context.Background()
	gameList, err := addAllGameToList()
	if err != nil {
		log.Error("receive InitGameCollection recover: ", err)
	}

	log.Info("CollectCurrentPlayers v2 采集开始")
	summary, runErr := runV2Tasks(ctx, gameList, []domain.TaskType{domain.TaskPlayers})
	logV2RunSummary("CollectCurrentPlayers v2", summary, runErr)
	persistV2RunSummary(ctx, "CollectCurrentPlayers v2", summary)
	log.Info("CollectCurrentPlayers v2 采集结束")
}

func addAllGameToList() (gameList []models.GameID, err common.GFError) {
	gameList, err = dao.GetGameList()
	if err != nil {
		log.Error("receive addAllGameToList recover: ", err)
	}
	return
}
