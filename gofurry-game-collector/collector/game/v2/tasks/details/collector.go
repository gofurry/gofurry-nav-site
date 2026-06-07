package details

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	steam "github.com/gofurry/steam-go"
	"github.com/gofurry/steam-go/web/storefront"

	"github.com/gofurry/gofurry-game-collector/collector/game/models"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	steammapper "github.com/gofurry/gofurry-game-collector/collector/game/v2/mapper/steam"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/report"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/steamclient"
)

// Repository persists one complete v2 details collection.
type Repository interface {
	SaveDetails(context.Context, domain.DetailsCollection) error
}

// Collector collects Store appdetails into the v2 storage contract.
type Collector struct {
	adapter *steamclient.Adapter
	repo    Repository
	mapper  steammapper.DetailsMapper

	requests []requestPlan
}

type requestPlan struct {
	region       domain.Region
	lang         domain.Language
	steamLang    string
	localized    bool
	preferAsBase bool
}

// NewCollector creates one v2 details collector.
func NewCollector(adapter *steamclient.Adapter, repo Repository) *Collector {
	return &Collector{
		adapter: adapter,
		repo:    repo,
		mapper:  steammapper.NewDetailsMapper(),
		requests: []requestPlan{
			{region: domain.RegionCN, lang: domain.LanguageZH, steamLang: "schinese", localized: true, preferAsBase: true},
			{region: domain.RegionUS, lang: domain.LanguageEN, steamLang: "english", localized: true},
			{region: domain.RegionHK, lang: domain.LanguageEN, steamLang: "english"},
		},
	}
}

// CollectGame collects details, localized copy, prices, media, requirements, and snapshots.
func (c *Collector) CollectGame(ctx context.Context, game models.GameID) (report.TaskResult, error) {
	startedAt := time.Now()
	result := report.TaskResult{
		Task:      domain.TaskDetails,
		Status:    domain.StatusSuccess,
		GameID:    game.ID,
		AppID:     uint32(game.Appid),
		StartedAt: startedAt,
	}

	if c == nil || c.adapter == nil {
		return c.finishFailed(result, report.ErrorValidation, "v2 steam adapter is nil")
	}
	if c.repo == nil {
		return c.finishFailed(result, report.ErrorValidation, "v2 details repository is nil")
	}
	if game.ID <= 0 || game.Appid <= 0 {
		return c.finishFailed(result, report.ErrorValidation, "game id and appid must be greater than zero")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	collection := domain.DetailsCollection{}
	localizedSeen := make(map[domain.Language]struct{})
	var firstErr error
	var haveBase bool

	for _, plan := range c.requests {
		data, rawPayload, err := c.fetchAppDetails(ctx, uint32(game.Appid), plan)
		collectedAt := time.Now()
		if len(rawPayload) > 0 && json.Valid(rawPayload) {
			collection.Snapshots = append(collection.Snapshots, domain.RawSnapshot{
				GameID:      game.ID,
				AppID:       uint32(game.Appid),
				Kind:        domain.SnapshotDetails,
				Language:    plan.lang,
				Region:      plan.region,
				Source:      domain.SourceSteam,
				PayloadHash: hashPayload(rawPayload),
				RawPayload:  rawPayload,
				CollectedAt: collectedAt,
			})
		}
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}

		collection.Prices = append(collection.Prices, c.mapper.ToPrice(game.ID, uint32(game.Appid), plan.region, data, collectedAt))

		if plan.localized {
			if _, exists := localizedSeen[plan.lang]; !exists {
				collection.Localized = append(collection.Localized, c.mapper.ToLocalized(game.ID, uint32(game.Appid), plan.lang, data, collectedAt))
				localizedSeen[plan.lang] = struct{}{}
			}
		}

		if !haveBase || plan.preferAsBase {
			details, err := c.mapper.ToDetails(game.ID, uint32(game.Appid), data, collectedAt)
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			collection.Details = details
			collection.Media = c.mapper.ToMedia(game.ID, uint32(game.Appid), data, collectedAt)
			collection.Requirements = c.mapper.ToRequirements(game.ID, uint32(game.Appid), data, collectedAt)
			haveBase = true
		}
	}

	if !haveBase {
		if firstErr != nil {
			return c.finishFailed(result, report.ErrorUpstream, firstErr.Error())
		}
		return c.finishFailed(result, report.ErrorUpstream, "no successful appdetails payload")
	}

	if err := c.repo.SaveDetails(ctx, collection); err != nil {
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

func (c *Collector) fetchAppDetails(ctx context.Context, appID uint32, plan requestPlan) (storefront.AppDetailsData, []byte, error) {
	var raw []byte
	err := c.adapter.Run(ctx, steamclient.BucketStore, func(runCtx context.Context, sdk *steam.Client) error {
		if sdk == nil || sdk.Web == nil || sdk.Web.Storefront == nil {
			return fmt.Errorf("steam storefront client is nil")
		}
		var err error
		raw, err = sdk.Web.Storefront.GetAppDetailsRaw(runCtx, appID, &storefront.GetAppDetailsOptions{
			CountryCode: string(plan.region),
			Language:    plan.steamLang,
		})
		return err
	})
	if err != nil {
		return storefront.AppDetailsData{}, raw, fmt.Errorf("get appdetails appid=%d region=%s lang=%s: %w", appID, plan.region, plan.lang, err)
	}

	var envelope storefront.AppDetailsEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return storefront.AppDetailsData{}, raw, fmt.Errorf("decode appdetails appid=%d region=%s lang=%s: %w", appID, plan.region, plan.lang, err)
	}
	result, ok := envelope[strconv.FormatUint(uint64(appID), 10)]
	if !ok {
		return storefront.AppDetailsData{}, raw, fmt.Errorf("appdetails appid=%d missing envelope entry", appID)
	}
	if !result.Success {
		return storefront.AppDetailsData{}, raw, fmt.Errorf("appdetails appid=%d region=%s lang=%s success=false", appID, plan.region, plan.lang)
	}
	return result.Data, raw, nil
}

func (c *Collector) finishFailed(result report.TaskResult, kind report.ErrorKind, message string) (report.TaskResult, error) {
	result.Status = domain.StatusFailed
	result.Error = &report.ErrorInfo{Kind: kind, Message: message}
	result.EndedAt = time.Now()
	result.DurationMillis = result.EndedAt.Sub(result.StartedAt).Milliseconds()
	return result, errors.New(message)
}

func hashPayload(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}
