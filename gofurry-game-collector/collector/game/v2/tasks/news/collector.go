package news

import (
	"context"
	"errors"
	"fmt"
	"time"

	steam "github.com/gofurry/steam-go"
	"github.com/gofurry/steam-go/web/storefront"

	"github.com/gofurry/gofurry-game-collector/collector/game/models"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	steammapper "github.com/gofurry/gofurry-game-collector/collector/game/v2/mapper/steam"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/report"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/steamclient"
)

const defaultNewsCount = 10

// Repository persists one batch of v2 news.
type Repository interface {
	SaveNews(context.Context, []domain.GameNews) error
}

// Collector collects Steam Store events news into the v2 storage contract.
type Collector struct {
	adapter *steamclient.Adapter
	repo    Repository
	mapper  steammapper.NewsMapper

	count     int
	languages []languageQuery
}

type languageQuery struct {
	lang         domain.Language
	languageList string
}

// NewCollector creates one v2 news collector.
func NewCollector(adapter *steamclient.Adapter, repo Repository) *Collector {
	return &Collector{
		adapter: adapter,
		repo:    repo,
		mapper:  steammapper.NewNewsMapper(),
		count:   defaultNewsCount,
		languages: []languageQuery{
			{lang: domain.LanguageZH, languageList: "6_0"},
			{lang: domain.LanguageEN, languageList: "0"},
		},
	}
}

// CollectGame collects zh/en Store events news for one game.
func (c *Collector) CollectGame(ctx context.Context, game models.GameID) (report.TaskResult, error) {
	startedAt := time.Now()
	result := report.TaskResult{
		Task:      domain.TaskNews,
		Status:    domain.StatusSuccess,
		GameID:    game.ID,
		AppID:     uint32(game.Appid),
		StartedAt: startedAt,
	}

	if c == nil || c.adapter == nil {
		return c.finishFailed(result, report.ErrorValidation, "v2 steam adapter is nil")
	}
	if c.repo == nil {
		return c.finishFailed(result, report.ErrorValidation, "v2 news repository is nil")
	}
	if game.ID <= 0 || game.Appid <= 0 {
		return c.finishFailed(result, report.ErrorValidation, "game id and appid must be greater than zero")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	var allNews []domain.GameNews
	var firstErr error
	for _, query := range c.languages {
		items, err := c.collectLanguage(ctx, game, query)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		allNews = append(allNews, items...)
	}

	if len(allNews) == 0 {
		if firstErr != nil {
			return c.finishFailed(result, report.ErrorUpstream, firstErr.Error())
		}
		result.Status = domain.StatusSkipped
		result.EndedAt = time.Now()
		result.DurationMillis = result.EndedAt.Sub(startedAt).Milliseconds()
		return result, nil
	}

	if err := c.repo.SaveNews(ctx, allNews); err != nil {
		return c.finishFailed(result, report.ErrorStorage, err.Error())
	}

	if firstErr != nil {
		result.Status = domain.StatusPartial
		result.Error = &report.ErrorInfo{Kind: report.ErrorUpstream, Message: firstErr.Error()}
	}
	result.EndedAt = time.Now()
	result.DurationMillis = result.EndedAt.Sub(startedAt).Milliseconds()
	return result, firstErr
}

func (c *Collector) collectLanguage(ctx context.Context, game models.GameID, query languageQuery) ([]domain.GameNews, error) {
	var resp storefront.AdjacentPartnerEventsResponse
	err := c.adapter.Run(ctx, steamclient.BucketStore, func(runCtx context.Context, sdk *steam.Client) error {
		if sdk == nil || sdk.Web == nil || sdk.Web.Storefront == nil {
			return fmt.Errorf("steam storefront client is nil")
		}
		var err error
		resp, err = sdk.Web.Storefront.GetAdjacentPartnerEvents(runCtx, uint32(game.Appid), &storefront.GetAdjacentPartnerEventsOptions{
			CountBefore:  1,
			CountAfter:   c.count,
			LanguageList: query.languageList,
		})
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("get adjacent partner events appid=%d lang=%s: %w", game.Appid, query.lang, err)
	}

	items := make([]domain.GameNews, 0, len(resp.Events))
	for _, event := range resp.Events {
		item, err := c.mapper.FromPartnerEvent(game.ID, uint32(game.Appid), query.lang, event)
		if err != nil {
			return nil, fmt.Errorf("map partner event appid=%d lang=%s gid=%s: %w", game.Appid, query.lang, event.GID, err)
		}
		if item.Headline == "" && item.RawBody == "" {
			continue
		}
		if item.EventGID == "" && item.AnnouncementGID == "" {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func (c *Collector) finishFailed(result report.TaskResult, kind report.ErrorKind, message string) (report.TaskResult, error) {
	result.Status = domain.StatusFailed
	result.Error = &report.ErrorInfo{Kind: kind, Message: message}
	result.EndedAt = time.Now()
	result.DurationMillis = result.EndedAt.Sub(result.StartedAt).Milliseconds()
	return result, errors.New(message)
}
