package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	cs "github.com/gofurry/gofurry-game-collector/common/service"
	gamesqlc "github.com/gofurry/gofurry-game-collector/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultNewsCacheTTL = 7 * 24 * time.Hour

// NewsRepository writes v2 news into PostgreSQL and refreshes Redis hot cache.
type NewsRepository struct {
	pool     *pgxpool.Pool
	cacheTTL time.Duration
}

// NewNewsRepository creates a repository with an explicit PostgreSQL pool.
func NewNewsRepository(pool *pgxpool.Pool) *NewsRepository {
	return &NewsRepository{
		pool:     pool,
		cacheTTL: defaultNewsCacheTTL,
	}
}

// SaveNews upserts one batch of v2 news and refreshes per-language Redis cache.
func (r *NewsRepository) SaveNews(ctx context.Context, items []domain.GameNews) error {
	if len(items) == 0 {
		return nil
	}
	if r == nil || r.pool == nil {
		return fmt.Errorf("news repository database is nil")
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
	for _, item := range items {
		if err := upsertNews(ctx, queries, item); err != nil {
			return err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	r.refreshCache(items)
	return nil
}

func upsertNews(ctx context.Context, queries *gamesqlc.Queries, item domain.GameNews) error {
	tags, err := marshalJSON(item.Tags)
	if err != nil {
		return fmt.Errorf("marshal news tags: %w", err)
	}
	rawEvent := item.RawEvent
	if len(rawEvent) == 0 {
		rawEvent = []byte("{}")
	}

	err = queries.UpsertNews(ctx, gamesqlc.UpsertNewsParams{
		GameID:          item.GameID,
		Appid:           int64(item.AppID),
		Lang:            string(item.Language),
		EventGid:        item.EventGID,
		AnnouncementGid: item.AnnouncementGID,
		ForumTopicID:    item.ForumTopicID,
		Headline:        item.Headline,
		RawBody:         item.RawBody,
		Html:            item.HTML,
		PlainText:       item.PlainText,
		Summary:         item.Summary,
		Url:             item.URL,
		Tags:            tags,
		VoteUpCount:     int32(item.VoteUpCount),
		VoteDownCount:   int32(item.VoteDownCount),
		CommentCount:    int32(item.CommentCount),
		RawEvent:        rawEvent,
		PublishedAt:     timestamptz(item.PublishedAt),
		UpdatedAt:       timestamptz(item.UpdatedAt),
		CollectedAt:     timestamptz(item.CollectedAt),
	})
	if err != nil {
		return fmt.Errorf("upsert v2 news appid=%d lang=%s event=%s announcement=%s: %w", item.AppID, item.Language, item.EventGID, item.AnnouncementGID, err)
	}
	return nil
}

func (r *NewsRepository) refreshCache(items []domain.GameNews) {
	if cs.GetRedisService() == nil {
		return
	}

	grouped := make(map[string][]domain.GameNews)
	for _, item := range items {
		key := newsCacheKey(item.GameID, item.Language)
		grouped[key] = append(grouped[key], item)
	}

	for key, newsItems := range grouped {
		payload, err := marshalJSON(newsItems)
		if err != nil {
			continue
		}
		_ = cs.SetExpire(key, string(payload), r.cacheTTL)
	}
}

func newsCacheKey(gameID int64, lang domain.StoreLocale) string {
	return fmt.Sprintf("game:v2:news:%d:%s", gameID, lang)
}

func marshalJSON(value any) ([]byte, error) {
	return json.Marshal(value)
}
