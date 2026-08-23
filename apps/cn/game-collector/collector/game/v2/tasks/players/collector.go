package players

import (
	"context"
	"errors"
	"fmt"
	"time"

	steam "github.com/gofurry/steam-go"

	"github.com/gofurry/gofurry-game-collector/collector/game/models"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/report"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/steamclient"
)

// Repository persists one v2 player-count result.
type Repository interface {
	SavePlayerCount(context.Context, domain.PlayerCount) error
}

// Collector collects current player counts through steam-go official API.
type Collector struct {
	adapter *steamclient.Adapter
	repo    Repository
}

// NewCollector creates one v2 player-count collector.
func NewCollector(adapter *steamclient.Adapter, repo Repository) *Collector {
	return &Collector{
		adapter: adapter,
		repo:    repo,
	}
}

// CollectGame collects current players for one game.
func (c *Collector) CollectGame(ctx context.Context, game models.GameID) (report.TaskResult, error) {
	startedAt := time.Now()
	result := report.TaskResult{
		Task:      domain.TaskPlayers,
		Status:    domain.StatusSuccess,
		GameID:    game.ID,
		AppID:     uint32(game.Appid),
		StartedAt: startedAt,
	}

	if c == nil || c.adapter == nil {
		return c.finishFailed(ctx, result, game, report.ErrorValidation, "v2 steam adapter is nil")
	}
	if c.repo == nil {
		return c.finishFailed(ctx, result, game, report.ErrorValidation, "v2 player repository is nil")
	}
	if game.ID <= 0 || game.Appid <= 0 {
		return c.finishFailed(ctx, result, game, report.ErrorValidation, "game id and appid must be greater than zero")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	var playerCount int64
	err := c.adapter.Run(ctx, steamclient.BucketOfficialAPI, func(runCtx context.Context, sdk *steam.Client) error {
		if sdk == nil || sdk.API == nil || sdk.API.SteamUserStats == nil {
			return fmt.Errorf("steam user stats client is nil")
		}
		resp, err := sdk.API.SteamUserStats.GetNumberOfCurrentPlayers(runCtx, uint32(game.Appid))
		if err != nil {
			return err
		}
		if resp.Response.Result != 1 {
			return fmt.Errorf("steam player-count result is %d", resp.Response.Result)
		}
		playerCount = int64(resp.Response.PlayerCount)
		return nil
	})
	if err != nil {
		return c.finishFailed(ctx, result, game, report.ErrorUpstream, err.Error())
	}

	item := domain.PlayerCount{
		RunID:       report.RunIDFromContext(ctx),
		GameID:      game.ID,
		AppID:       uint32(game.Appid),
		Count:       playerCount,
		Status:      domain.StatusSuccess,
		CollectedAt: time.Now(),
	}
	if err := c.repo.SavePlayerCount(ctx, item); err != nil {
		return c.finishFailed(ctx, result, game, report.ErrorStorage, err.Error())
	}

	result.EndedAt = time.Now()
	result.DurationMillis = result.EndedAt.Sub(startedAt).Milliseconds()
	return result, nil
}

func (c *Collector) finishFailed(ctx context.Context, result report.TaskResult, game models.GameID, kind report.ErrorKind, message string) (report.TaskResult, error) {
	result.Status = domain.StatusFailed
	result.Error = &report.ErrorInfo{Kind: kind, Message: message}
	result.EndedAt = time.Now()
	result.DurationMillis = result.EndedAt.Sub(result.StartedAt).Milliseconds()

	if c != nil && c.repo != nil && game.ID > 0 && game.Appid > 0 {
		_ = c.repo.SavePlayerCount(ctx, domain.PlayerCount{
			RunID:        report.RunIDFromContext(ctx),
			GameID:       game.ID,
			AppID:        uint32(game.Appid),
			Status:       domain.StatusFailed,
			ErrorKind:    string(kind),
			ErrorMessage: message,
			CollectedAt:  time.Now(),
		})
	}

	return result, errors.New(message)
}
