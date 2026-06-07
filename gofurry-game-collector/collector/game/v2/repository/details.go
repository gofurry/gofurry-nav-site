package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	cs "github.com/gofurry/gofurry-game-collector/common/service"
	database "github.com/gofurry/gofurry-game-collector/roof/db"
	"gorm.io/gorm"
)

const defaultDetailsCacheTTL = 7 * 24 * time.Hour

// DetailsRepository writes v2 game details into PostgreSQL and Redis.
type DetailsRepository struct {
	db       *gorm.DB
	cacheTTL time.Duration
}

// NewDetailsRepository creates a repository backed by the global PostgreSQL handle.
func NewDetailsRepository() *DetailsRepository {
	return NewDetailsRepositoryWithDB(database.Orm.DB())
}

// NewDetailsRepositoryWithDB creates a repository with an explicit PostgreSQL handle.
func NewDetailsRepositoryWithDB(db *gorm.DB) *DetailsRepository {
	return &DetailsRepository{db: db, cacheTTL: defaultDetailsCacheTTL}
}

// SaveDetails upserts one complete v2 details collection.
func (r *DetailsRepository) SaveDetails(ctx context.Context, data domain.DetailsCollection) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("details repository database is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := upsertDetails(ctx, tx, data.Details); err != nil {
			return err
		}
		for _, item := range data.Localized {
			if err := upsertLocalizedDetails(ctx, tx, item); err != nil {
				return err
			}
		}
		for _, item := range data.Prices {
			if err := upsertPrice(ctx, tx, item); err != nil {
				return err
			}
		}
		if err := replaceMedia(ctx, tx, data.Media); err != nil {
			return err
		}
		if err := upsertRequirements(ctx, tx, data.Requirements); err != nil {
			return err
		}
		for _, snapshot := range data.Snapshots {
			if err := insertSnapshot(ctx, tx, snapshot); err != nil {
				return err
			}
			if err := pruneSnapshots(ctx, tx, snapshot.AppID, snapshot.Language, snapshot.Region); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	r.refreshCache(data)
	return nil
}

func upsertDetails(ctx context.Context, tx *gorm.DB, item domain.GameDetails) error {
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

	return tx.WithContext(ctx).Exec(`
INSERT INTO gfg_game_v2_details (
    game_id, appid, source, type, name, is_free, website, header_url,
    developers, publishers, release_coming_soon, release_date_text,
    platforms, supported_languages, support_info, content_descriptors, ratings,
    collected_at, updated_at
) VALUES (
    ?, ?, 'steam', ?, ?, ?, ?, ?, ?::jsonb, ?::jsonb, ?, ?, ?::jsonb, ?, ?::jsonb, ?::jsonb, ?::jsonb, ?, now()
)
ON CONFLICT (game_id)
DO UPDATE SET
    appid = EXCLUDED.appid,
    source = EXCLUDED.source,
    type = EXCLUDED.type,
    name = EXCLUDED.name,
    is_free = EXCLUDED.is_free,
    website = EXCLUDED.website,
    header_url = EXCLUDED.header_url,
    developers = EXCLUDED.developers,
    publishers = EXCLUDED.publishers,
    release_coming_soon = EXCLUDED.release_coming_soon,
    release_date_text = EXCLUDED.release_date_text,
    platforms = EXCLUDED.platforms,
    supported_languages = EXCLUDED.supported_languages,
    support_info = EXCLUDED.support_info,
    content_descriptors = EXCLUDED.content_descriptors,
    ratings = EXCLUDED.ratings,
    collected_at = EXCLUDED.collected_at,
    updated_at = now()
`,
		item.GameID, item.AppID, item.Type, item.Name, item.IsFree, item.Website, item.HeaderURL,
		string(developers), string(publishers), item.Release.ComingSoon, item.Release.DateText,
		string(platforms), item.SupportedLanguages, string(supportInfo), string(contentDescriptors), string(ratings),
		item.CollectedAt,
	).Error
}

func upsertLocalizedDetails(ctx context.Context, tx *gorm.DB, item domain.GameLocalizedDetails) error {
	return tx.WithContext(ctx).Exec(`
INSERT INTO gfg_game_v2_localized_details (
    game_id, appid, lang, name, short_description, detailed_description, about_the_game, collected_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, now()
)
ON CONFLICT (game_id, lang)
DO UPDATE SET
    appid = EXCLUDED.appid,
    name = EXCLUDED.name,
    short_description = EXCLUDED.short_description,
    detailed_description = EXCLUDED.detailed_description,
    about_the_game = EXCLUDED.about_the_game,
    collected_at = EXCLUDED.collected_at,
    updated_at = now()
`,
		item.GameID, item.AppID, string(item.Language), item.Name, item.ShortDescription, item.DetailedDescription, item.AboutTheGame, item.CollectedAt,
	).Error
}

func upsertPrice(ctx context.Context, tx *gorm.DB, item domain.GamePrice) error {
	return tx.WithContext(ctx).Exec(`
INSERT INTO gfg_game_v2_prices (
    game_id, appid, region, is_free, currency, initial_amount, final_amount,
    discount_percent, initial_formatted, final_formatted, collected_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now()
)
ON CONFLICT (game_id, region)
DO UPDATE SET
    appid = EXCLUDED.appid,
    is_free = EXCLUDED.is_free,
    currency = EXCLUDED.currency,
    initial_amount = EXCLUDED.initial_amount,
    final_amount = EXCLUDED.final_amount,
    discount_percent = EXCLUDED.discount_percent,
    initial_formatted = EXCLUDED.initial_formatted,
    final_formatted = EXCLUDED.final_formatted,
    collected_at = EXCLUDED.collected_at,
    updated_at = now()
`,
		item.GameID, item.AppID, string(item.Region), item.IsFree, item.Currency, item.Initial, item.Final,
		item.DiscountPercent, item.InitialFormatted, item.FinalFormatted, item.CollectedAt,
	).Error
}

func replaceMedia(ctx context.Context, tx *gorm.DB, media domain.GameMedia) error {
	if err := tx.WithContext(ctx).Exec("DELETE FROM gfg_game_v2_media WHERE game_id = ?", media.GameID).Error; err != nil {
		return err
	}
	items, err := mediaItems(media)
	if err != nil {
		return err
	}
	for _, item := range items {
		if err := insertMedia(ctx, tx, media, item); err != nil {
			return err
		}
	}
	return nil
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

func insertMedia(ctx context.Context, tx *gorm.DB, media domain.GameMedia, item mediaItem) error {
	if item.url == "" && item.thumbnailURL == "" {
		return nil
	}
	extra, err := marshalJSON(item.extra)
	if err != nil {
		return fmt.Errorf("marshal media extra: %w", err)
	}
	return tx.WithContext(ctx).Exec(`
INSERT INTO gfg_game_v2_media (
    game_id, appid, media_type, media_key, title, url, thumbnail_url, extra, sort_order, collected_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?::jsonb, ?, ?, now()
)
ON CONFLICT (game_id, media_type, media_key)
DO UPDATE SET
    appid = EXCLUDED.appid,
    title = EXCLUDED.title,
    url = EXCLUDED.url,
    thumbnail_url = EXCLUDED.thumbnail_url,
    extra = EXCLUDED.extra,
    sort_order = EXCLUDED.sort_order,
    collected_at = EXCLUDED.collected_at,
    updated_at = now()
`,
		media.GameID, media.AppID, item.typ, item.key, item.title, item.url, item.thumbnailURL, string(extra), item.sortOrder, media.CollectedAt,
	).Error
}

func upsertRequirements(ctx context.Context, tx *gorm.DB, item domain.SystemRequirements) error {
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
	return tx.WithContext(ctx).Exec(`
INSERT INTO gfg_game_v2_requirements (
    game_id, appid, pc, mac, linux, collected_at, updated_at
) VALUES (
    ?, ?, ?::jsonb, ?::jsonb, ?::jsonb, ?, now()
)
ON CONFLICT (game_id)
DO UPDATE SET
    appid = EXCLUDED.appid,
    pc = EXCLUDED.pc,
    mac = EXCLUDED.mac,
    linux = EXCLUDED.linux,
    collected_at = EXCLUDED.collected_at,
    updated_at = now()
`,
		item.GameID, item.AppID, string(pc), string(mac), string(linux), item.CollectedAt,
	).Error
}

func insertSnapshot(ctx context.Context, tx *gorm.DB, item domain.RawSnapshot) error {
	payloadHash := item.PayloadHash
	if payloadHash == "" {
		payloadHash = hashPayload(item.RawPayload)
	}
	return tx.WithContext(ctx).Exec(`
INSERT INTO gfg_game_v2_detail_snapshots (
    game_id, appid, lang, region, source, payload_hash, raw_payload, collected_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?::jsonb, ?
)
`,
		item.GameID, item.AppID, string(item.Language), string(item.Region), string(item.Source), payloadHash, string(item.RawPayload), item.CollectedAt,
	).Error
}

func pruneSnapshots(ctx context.Context, tx *gorm.DB, appID uint32, lang domain.Language, region domain.Region) error {
	return tx.WithContext(ctx).Exec("SELECT gfg_game_v2_prune_detail_snapshots(?, ?, ?, 5)", appID, string(lang), string(region)).Error
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
			Requirements domain.SystemRequirements   `json:"requirements"`
		}{
			Details:      data.Details,
			Localized:    localized,
			Prices:       data.Prices,
			Media:        data.Media,
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
}

func detailsCacheKey(gameID int64, lang domain.Language) string {
	return fmt.Sprintf("game:v2:details:%d:%s", gameID, lang)
}

func pricesCacheKey(gameID int64) string {
	return fmt.Sprintf("game:v2:prices:%d", gameID)
}

func mediaCacheKey(gameID int64) string {
	return fmt.Sprintf("game:v2:media:%d", gameID)
}

func hashPayload(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}
