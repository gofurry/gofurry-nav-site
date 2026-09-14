package details

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	steam "github.com/gofurry/steam-go"
	"github.com/gofurry/steam-go/addons/assets"
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
	runSteam           func(context.Context, steamclient.Bucket, func(context.Context, *steam.Client) error) error
	fetchAppDetailsFn  func(context.Context, uint32, requestPlan) (storefront.AppDetailsData, []byte, error)
	fetchStoreBrowseFn func(context.Context, uint32, requestPlan) ([]assets.URLItem, error)
	repo               Repository
	mapper             steammapper.DetailsMapper

	requests []requestPlan
}

type requestPlan struct {
	region       domain.Region
	lang         domain.StoreLocale
	steamLang    string
	localized    bool
	preferAsBase bool
	canonical    bool
}

// NewCollector creates one v2 details collector.
func NewCollector(adapter *steamclient.Adapter, repo Repository) *Collector {
	collector := &Collector{
		repo:   repo,
		mapper: steammapper.NewDetailsMapper(),
		requests: []requestPlan{
			{region: domain.RegionCN, lang: domain.StoreLocaleZH, steamLang: "schinese", localized: true, preferAsBase: true},
			{region: domain.RegionUS, lang: domain.StoreLocaleEN, steamLang: "english", localized: true, canonical: true},
			{region: domain.RegionHK, lang: domain.StoreLocaleEN, steamLang: "english"},
		},
	}
	if adapter != nil {
		collector.runSteam = adapter.Run
	}
	collector.fetchAppDetailsFn = collector.fetchAppDetails
	collector.fetchStoreBrowseFn = collector.fetchStoreBrowseAssets
	return collector
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

	if c == nil || c.runSteam == nil || c.fetchAppDetailsFn == nil || c.fetchStoreBrowseFn == nil {
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
	localizedSeen := make(map[domain.StoreLocale]struct{})
	var collectionErr error
	var haveBase bool

	for _, plan := range c.requests {
		data, rawPayload, err := c.fetchAppDetailsFn(ctx, uint32(game.Appid), plan)
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
			collectionErr = errors.Join(collectionErr, withSteamBucket(steamclient.BucketStore,
				fmt.Errorf("appdetails observation appid=%d region=%s lang=%s: %w", game.Appid, plan.region, plan.lang, err)))
			continue
		}

		collection.Prices = append(collection.Prices, c.mapper.ToPrice(game.ID, uint32(game.Appid), plan.region, data, collectedAt))

		if plan.canonical {
			if data.ReleaseDate != nil {
				release := c.mapper.ToCanonicalRelease(game.ID, data.ReleaseDate, collectedAt)
				collection.CanonicalRelease = &release
			}
			if strings.TrimSpace(data.SupportedLanguages) != "" {
				parsed := storefront.ParseSupportedLanguages(data.SupportedLanguages)
				if len(parsed) > 0 {
					languages := c.mapper.ToCanonicalLanguages(parsed, collectedAt)
					collection.CanonicalLanguages = &languages
				}
			}
		}

		if plan.localized {
			if _, exists := localizedSeen[plan.lang]; !exists {
				collection.Localized = append(collection.Localized, c.mapper.ToLocalized(game.ID, uint32(game.Appid), plan.lang, data, collectedAt))
				localizedSeen[plan.lang] = struct{}{}
			}
		}

		if !haveBase || plan.preferAsBase {
			details, err := c.mapper.ToDetails(game.ID, uint32(game.Appid), data, collectedAt)
			if err != nil {
				collectionErr = errors.Join(collectionErr, err)
				continue
			}
			collection.Details = details
			collection.Media = c.mapper.ToMedia(game.ID, uint32(game.Appid), data, collectedAt)
			collection.Assets = append(collection.Assets, c.mapper.ToStorefrontAssets(game.ID, uint32(game.Appid), data, collectedAt)...)
			collection.AssetReplaceScopes = append(collection.AssetReplaceScopes, domain.AssetReplaceScope{Source: domain.AssetSourceStorefront})
			collection.Requirements = c.mapper.ToRequirements(game.ID, uint32(game.Appid), data, collectedAt)
			haveBase = true
		}
	}

	if !haveBase {
		if collectionErr != nil {
			return c.finishFailed(result, report.ErrorUpstream, collectionErr.Error())
		}
		return c.finishFailed(result, report.ErrorUpstream, "no successful appdetails payload")
	}

	storeBrowse, storeBrowseErr := c.collectStoreBrowseAssets(ctx, game)
	collection.Assets = append(collection.Assets, storeBrowse.Assets...)
	collection.AssetReplaceScopes = append(collection.AssetReplaceScopes, storeBrowse.ReplaceScopes...)
	collectionErr = errors.Join(collectionErr, storeBrowseErr)

	if err := c.repo.SaveDetails(ctx, collection); err != nil {
		return c.finishFailed(result, report.ErrorStorage, err.Error())
	}

	if collectionErr != nil {
		result.Status = domain.StatusPartial
		applyTaskErrorDiagnostics(&result, collectionErr)
	}
	result.EndedAt = time.Now()
	result.DurationMillis = result.EndedAt.Sub(startedAt).Milliseconds()
	return result, collectionErr
}

type storeBrowseCollection struct {
	Assets        []domain.GameMediaAsset
	ReplaceScopes []domain.AssetReplaceScope
}

func (c *Collector) collectStoreBrowseAssets(ctx context.Context, game models.GameID) (storeBrowseCollection, error) {
	collectedAt := time.Now()
	appID := uint32(game.Appid)
	result := storeBrowseCollection{}
	var collectionErr error

	for _, plan := range c.requests {
		if !plan.localized {
			continue
		}
		items, err := c.fetchStoreBrowseFn(ctx, appID, plan)
		if err != nil {
			collectionErr = errors.Join(collectionErr, withSteamBucket(steamclient.BucketOfficialAPI,
				fmt.Errorf("store browse observation appid=%d region=%s lang=%s: %w", appID, plan.region, plan.lang, err)))
			continue
		}
		mapped := c.mapper.ToStoreBrowseAssets(game.ID, appID, plan.lang, items, collectedAt)
		if !hasVerticalCover(mapped) {
			collectionErr = errors.Join(collectionErr, fmt.Errorf("store browse assets incomplete appid=%d region=%s lang=%s: missing library_capsule/library_capsule_2x", appID, plan.region, plan.lang))
			continue
		}
		result.Assets = append(result.Assets, mapped...)
		result.ReplaceScopes = append(result.ReplaceScopes, domain.AssetReplaceScope{Source: domain.AssetSourceStoreBrowse, Language: string(plan.lang)})
	}

	return result, collectionErr
}

func hasVerticalCover(items []domain.GameMediaAsset) bool {
	for _, item := range items {
		if (item.AssetType == string(assets.KindLibraryCapsule) || item.AssetType == string(assets.KindLibraryCapsule2x)) && strings.TrimSpace(item.URL) != "" && (item.Exists == nil || *item.Exists) {
			return true
		}
	}
	return false
}

func (c *Collector) fetchStoreBrowseAssets(ctx context.Context, appID uint32, plan requestPlan) ([]assets.URLItem, error) {
	var items []assets.URLItem
	err := c.runSteam(ctx, steamclient.BucketOfficialAPI, func(runCtx context.Context, sdk *steam.Client) error {
		if sdk == nil || sdk.API == nil || sdk.API.StoreBrowseService == nil {
			return fmt.Errorf("steam store browse service is nil")
		}
		fetched, err := assets.FetchStoreItemAssetURLs(runCtx, sdk.API.StoreBrowseService, assets.StoreItemAssetOptions{
			CountryCode: string(plan.region),
			Language:    plan.steamLang,
			Kinds:       assets.StoreItemAssetKinds(),
			StripQuery:  true,
		}, appID)
		if err != nil {
			return fmt.Errorf("fetch store browse assets appid=%d region=%s lang=%s: %w", appID, plan.region, plan.lang, err)
		}
		items = fetched
		return nil
	})
	return items, err
}

func (c *Collector) fetchAppDetails(ctx context.Context, appID uint32, plan requestPlan) (storefront.AppDetailsData, []byte, error) {
	var raw []byte
	err := c.runSteam(ctx, steamclient.BucketStore, func(runCtx context.Context, sdk *steam.Client) error {
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

type steamRequestError struct {
	bucket steamclient.Bucket
	err    error
}

func (err *steamRequestError) Error() string { return err.err.Error() }
func (err *steamRequestError) Unwrap() error { return err.err }

func withSteamBucket(bucket steamclient.Bucket, err error) error {
	if err == nil {
		return nil
	}
	var annotated *steamRequestError
	if errors.As(err, &annotated) {
		return err
	}
	return &steamRequestError{bucket: bucket, err: err}
}

func applyTaskErrorDiagnostics(result *report.TaskResult, err error) {
	if result == nil || err == nil {
		return
	}
	kind := report.ErrorUpstream
	var apiErr *steam.APIError
	if errors.As(err, &apiErr) && apiErr != nil {
		result.UpstreamStatusCode = apiErr.StatusCode
		switch {
		case apiErr.StatusCode == 429:
			kind = report.ErrorRateLimited
		case apiErr.StatusCode == 403:
			kind = report.ErrorBlocked
		case apiErr.Kind == steam.ErrorKindDecode:
			kind = report.ErrorDecode
		}
	}
	var requestErr *steamRequestError
	if errors.As(err, &requestErr) && requestErr != nil {
		result.TrafficBucket = string(requestErr.bucket)
	}
	result.Error = &report.ErrorInfo{Kind: kind, Message: err.Error()}
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
