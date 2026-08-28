package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	cs "github.com/gofurry/gofurry-game-collector/common/service"
	gamesqlc "github.com/gofurry/gofurry-game-collector/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultDetailsCacheTTL = 7 * 24 * time.Hour

// ErrStaleCollection means the parent game's AppID changed after Steam data was fetched.
var ErrStaleCollection = errors.New("stale Steam details collection")

// DetailsRepository writes v2 game details into PostgreSQL and Redis.
type DetailsRepository struct {
	pool     *pgxpool.Pool
	cacheTTL time.Duration
}

// NewDetailsRepository creates a repository with an explicit PostgreSQL pool.
func NewDetailsRepository(pool *pgxpool.Pool) *DetailsRepository {
	return &DetailsRepository{pool: pool, cacheTTL: defaultDetailsCacheTTL}
}

// SaveDetails upserts one complete v2 details collection.
func (r *DetailsRepository) SaveDetails(ctx context.Context, data domain.DetailsCollection) error {
	if r == nil || r.pool == nil {
		return fmt.Errorf("details repository database is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	queries := gamesqlc.New(tx)
	target, err := queries.LockGameTarget(ctx, data.Details.GameID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w: game_id=%d no longer exists", ErrStaleCollection, data.Details.GameID)
		}
		return fmt.Errorf("lock game target: %w", err)
	}
	if target.Appid != int64(data.Details.AppID) {
		return fmt.Errorf("%w: game_id=%d collected_appid=%d current_appid=%d", ErrStaleCollection, data.Details.GameID, data.Details.AppID, target.Appid)
	}
	if err := upsertDetails(ctx, queries, data.Details); err != nil {
		return err
	}
	for _, item := range data.Localized {
		if err := upsertLocalizedDetails(ctx, queries, item); err != nil {
			return err
		}
	}
	for _, item := range data.Prices {
		if err := upsertPrice(ctx, queries, item); err != nil {
			return err
		}
	}
	if err := replaceMedia(ctx, queries, data.Media); err != nil {
		return err
	}
	if err := replaceAssets(ctx, queries, data.Details.GameID, data.Assets); err != nil {
		return err
	}
	if err := upsertRequirements(ctx, queries, data.Requirements); err != nil {
		return err
	}
	if data.CanonicalRelease != nil {
		if err := saveCanonicalRelease(ctx, queries, *data.CanonicalRelease); err != nil {
			return err
		}
	}
	if data.CanonicalLanguages != nil {
		if err := replaceCanonicalLanguages(ctx, queries, data.Details.GameID, data.CanonicalLanguages.Items); err != nil {
			return err
		}
	}
	for _, snapshot := range data.Snapshots {
		if err := insertSnapshot(ctx, queries, snapshot); err != nil {
			return err
		}
		if err := pruneSnapshots(ctx, queries, snapshot.AppID, snapshot.Language, snapshot.Region); err != nil {
			return err
		}
	}
	if err := queries.RefreshCollectedGameDaily(ctx, gamesqlc.RefreshCollectedGameDailyParams{
		GameID: data.Details.GameID, MaterializationSource: "observed",
	}); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	r.refreshCache(data)
	return nil
}

func upsertDetails(ctx context.Context, queries *gamesqlc.Queries, item domain.GameDetails) error {
	developers, err := marshalJSON(item.Developers)
	if err != nil {
		return fmt.Errorf("marshal developers: %w", err)
	}
	publishers, err := marshalJSON(item.Publishers)
	if err != nil {
		return fmt.Errorf("marshal publishers: %w", err)
	}
	platforms, err := marshalJSON(item.Platforms)
	if err != nil {
		return fmt.Errorf("marshal platforms: %w", err)
	}
	supportInfo, err := marshalJSON(item.SupportInfo)
	if err != nil {
		return fmt.Errorf("marshal support info: %w", err)
	}
	contentDescriptors, err := marshalJSON(item.ContentDescriptors)
	if err != nil {
		return fmt.Errorf("marshal content descriptors: %w", err)
	}
	ratings, err := marshalJSON(item.Ratings)
	if err != nil {
		return fmt.Errorf("marshal ratings: %w", err)
	}

	return queries.UpsertDetails(ctx, gamesqlc.UpsertDetailsParams{
		GameID:             item.GameID,
		Appid:              int64(item.AppID),
		Type:               item.Type,
		Name:               item.Name,
		IsFree:             item.IsFree,
		Website:            item.Website,
		HeaderUrl:          item.HeaderURL,
		Developers:         developers,
		Publishers:         publishers,
		ReleaseComingSoon:  item.ReleaseRaw.ComingSoon,
		ReleaseDateText:    item.ReleaseRaw.DateText,
		Platforms:          platforms,
		SupportedLanguages: item.SupportedLanguagesRaw,
		SupportInfo:        supportInfo,
		ContentDescriptors: contentDescriptors,
		Ratings:            ratings,
		CollectedAt:        timestamptz(item.CollectedAt),
	})
}

func upsertLocalizedDetails(ctx context.Context, queries *gamesqlc.Queries, item domain.GameLocalizedDetails) error {
	return queries.UpsertLocalizedDetails(ctx, gamesqlc.UpsertLocalizedDetailsParams{
		GameID:              item.GameID,
		Appid:               int64(item.AppID),
		Lang:                string(item.Language),
		Name:                item.Name,
		ShortDescription:    item.ShortDescription,
		DetailedDescription: item.DetailedDescription,
		AboutTheGame:        item.AboutTheGame,
		CollectedAt:         timestamptz(item.CollectedAt),
	})
}

func upsertPrice(ctx context.Context, queries *gamesqlc.Queries, item domain.GamePrice) error {
	state := item.PriceState
	if state == "" {
		switch {
		case item.IsFree:
			state = domain.PriceStateFree
		case item.Currency != "":
			state = domain.PriceStatePriced
		default:
			state = domain.PriceStateUnknown
		}
	}
	if err := queries.UpsertPriceDailyObserved(ctx, gamesqlc.UpsertPriceDailyObservedParams{
		GameID: item.GameID, Appid: int64(item.AppID), Region: string(item.Region),
		PriceState: string(state), Currency: item.Currency,
		InitialAmount: item.Initial, FinalAmount: item.Final,
		DiscountPercent: int32(item.DiscountPercent), ObservedAt: timestamptz(item.CollectedAt),
	}); err != nil {
		return err
	}
	return queries.UpsertPrice(ctx, gamesqlc.UpsertPriceParams{
		GameID:           item.GameID,
		Appid:            int64(item.AppID),
		Region:           string(item.Region),
		IsFree:           item.IsFree,
		PriceState:       string(state),
		Currency:         item.Currency,
		InitialAmount:    item.Initial,
		FinalAmount:      item.Final,
		DiscountPercent:  item.DiscountPercent,
		InitialFormatted: item.InitialFormatted,
		FinalFormatted:   item.FinalFormatted,
		CollectedAt:      timestamptz(item.CollectedAt),
	})
}

func replaceMedia(ctx context.Context, queries *gamesqlc.Queries, media domain.GameMedia) error {
	if err := queries.DeleteMediaByGame(ctx, media.GameID); err != nil {
		return err
	}
	items, err := mediaItems(media)
	if err != nil {
		return err
	}
	for _, item := range items {
		if err := insertMedia(ctx, queries, media, item); err != nil {
			return err
		}
	}
	return nil
}

func replaceAssets(ctx context.Context, queries *gamesqlc.Queries, gameID int64, assets []domain.GameMediaAsset) error {
	if err := queries.DeleteAssetsByGame(ctx, gameID); err != nil {
		return err
	}
	for _, item := range assets {
		if err := insertAsset(ctx, queries, item); err != nil {
			return err
		}
	}
	return nil
}

func insertAsset(ctx context.Context, queries *gamesqlc.Queries, item domain.GameMediaAsset) error {
	if item.GameID <= 0 || item.AppID == 0 || item.AssetType == "" || item.URL == "" {
		return nil
	}
	if item.Extra == nil {
		item.Extra = map[string]any{}
	}
	extra, err := marshalJSON(item.Extra)
	if err != nil {
		return fmt.Errorf("marshal asset extra: %w", err)
	}
	return queries.UpsertAsset(ctx, gamesqlc.UpsertAssetParams{
		GameID:        item.GameID,
		Appid:         int64(item.AppID),
		AssetType:     item.AssetType,
		AssetFamily:   item.AssetFamily,
		Source:        item.Source,
		Lang:          item.Language,
		MediaKey:      item.MediaKey,
		Title:         item.Title,
		Url:           item.URL,
		ThumbnailUrl:  item.ThumbnailURL,
		Format:        item.Format,
		Exists:        item.Exists,
		StatusCode:    int32(item.StatusCode),
		ContentType:   item.ContentType,
		ContentLength: item.ContentLength,
		Extra:         extra,
		SortOrder:     int32(item.SortOrder),
		CheckedAt:     nullableTimestamptz(item.CheckedAt),
		CollectedAt:   timestamptz(item.CollectedAt),
	})
}

type mediaItem struct {
	typ          string
	key          string
	title        string
	url          string
	thumbnailURL string
	extra        any
	sortOrder    int
}

func mediaItems(media domain.GameMedia) ([]mediaItem, error) {
	items := []mediaItem{
		{typ: "header", key: "header", url: media.HeaderURL},
		{typ: "capsule", key: "capsule", url: media.CapsuleURL},
		{typ: "capsule_v5", key: "capsule_v5", url: media.CapsuleV5URL},
		{typ: "background", key: "background", url: media.BackgroundURL},
		{typ: "background_raw", key: "background_raw", url: media.BackgroundRawURL},
	}
	for idx, screenshot := range media.Screenshots {
		items = append(items, mediaItem{
			typ:          "screenshot",
			key:          strconv.Itoa(screenshot.ID),
			url:          screenshot.FullURL,
			thumbnailURL: screenshot.ThumbnailURL,
			sortOrder:    idx,
		})
	}
	for idx, movie := range media.Movies {
		items = append(items, mediaItem{
			typ:          "movie",
			key:          strconv.Itoa(movie.ID),
			title:        movie.Name,
			url:          movie.DASHH264URL,
			thumbnailURL: movie.ThumbnailURL,
			extra:        movie,
			sortOrder:    idx,
		})
	}
	return items, nil
}

func insertMedia(ctx context.Context, queries *gamesqlc.Queries, media domain.GameMedia, item mediaItem) error {
	if item.url == "" && item.thumbnailURL == "" {
		return nil
	}
	extra, err := marshalJSON(item.extra)
	if err != nil {
		return fmt.Errorf("marshal media extra: %w", err)
	}
	return queries.UpsertMedia(ctx, gamesqlc.UpsertMediaParams{
		GameID:       media.GameID,
		Appid:        int64(media.AppID),
		MediaType:    item.typ,
		MediaKey:     item.key,
		Title:        item.title,
		Url:          item.url,
		ThumbnailUrl: item.thumbnailURL,
		Extra:        extra,
		SortOrder:    int32(item.sortOrder),
		CollectedAt:  timestamptz(media.CollectedAt),
	})
}

func upsertRequirements(ctx context.Context, queries *gamesqlc.Queries, item domain.SystemRequirements) error {
	pc, err := marshalJSON(item.PC)
	if err != nil {
		return fmt.Errorf("marshal pc requirements: %w", err)
	}
	mac, err := marshalJSON(item.Mac)
	if err != nil {
		return fmt.Errorf("marshal mac requirements: %w", err)
	}
	linux, err := marshalJSON(item.Linux)
	if err != nil {
		return fmt.Errorf("marshal linux requirements: %w", err)
	}
	return queries.UpsertRequirements(ctx, gamesqlc.UpsertRequirementsParams{
		GameID:      item.GameID,
		Appid:       int64(item.AppID),
		Pc:          pc,
		Mac:         mac,
		Linux:       linux,
		CollectedAt: timestamptz(item.CollectedAt),
	})
}

func insertSnapshot(ctx context.Context, queries *gamesqlc.Queries, item domain.RawSnapshot) error {
	payloadHash := item.PayloadHash
	if payloadHash == "" {
		payloadHash = hashPayload(item.RawPayload)
	}
	return queries.InsertDetailSnapshot(ctx, gamesqlc.InsertDetailSnapshotParams{
		GameID:      item.GameID,
		Appid:       int64(item.AppID),
		Lang:        string(item.Language),
		Region:      string(item.Region),
		Source:      string(item.Source),
		PayloadHash: payloadHash,
		RawPayload:  item.RawPayload,
		CollectedAt: timestamptz(item.CollectedAt),
	})
}

func pruneSnapshots(ctx context.Context, queries *gamesqlc.Queries, appID uint32, lang domain.StoreLocale, region domain.Region) error {
	_, err := queries.PruneDetailSnapshots(ctx, gamesqlc.PruneDetailSnapshotsParams{Appid: int64(appID), Lang: string(lang), Region: string(region)})
	return err
}

func (r *DetailsRepository) refreshCache(data domain.DetailsCollection) {
	if cs.GetRedisService() == nil {
		return
	}
	for _, localized := range data.Localized {
		payload, err := marshalJSON(struct {
			Details      domain.GameDetails          `json:"details"`
			Localized    domain.GameLocalizedDetails `json:"localized"`
			Prices       []domain.GamePrice          `json:"prices"`
			Media        domain.GameMedia            `json:"media"`
			Assets       []domain.GameMediaAsset     `json:"assets"`
			Requirements domain.SystemRequirements   `json:"requirements"`
		}{
			Details:      data.Details,
			Localized:    localized,
			Prices:       data.Prices,
			Media:        data.Media,
			Assets:       data.Assets,
			Requirements: data.Requirements,
		})
		if err == nil {
			_ = cs.SetExpire(detailsCacheKey(data.Details.GameID, localized.Language), string(payload), r.cacheTTL)
		}
	}
	if payload, err := marshalJSON(data.Prices); err == nil {
		_ = cs.SetExpire(pricesCacheKey(data.Details.GameID), string(payload), r.cacheTTL)
	}
	if payload, err := marshalJSON(data.Media); err == nil {
		_ = cs.SetExpire(mediaCacheKey(data.Details.GameID), string(payload), r.cacheTTL)
	}
	if payload, err := marshalJSON(data.Assets); err == nil {
		_ = cs.SetExpire(assetsCacheKey(data.Details.GameID), string(payload), r.cacheTTL)
	}
}

func detailsCacheKey(gameID int64, lang domain.StoreLocale) string {
	return fmt.Sprintf("game:v2:details:%d:%s", gameID, lang)
}

func pricesCacheKey(gameID int64) string {
	return fmt.Sprintf("game:v2:prices:%d", gameID)
}

func mediaCacheKey(gameID int64) string {
	return fmt.Sprintf("game:v2:media:%d", gameID)
}

func assetsCacheKey(gameID int64) string {
	return fmt.Sprintf("game:v2:assets:%d", gameID)
}

func hashPayload(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}
