package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	cs "github.com/gofurry/gofurry-game-collector/common/service"
	database "github.com/gofurry/gofurry-game-collector/roof/db"
	"gorm.io/gorm"
)

const defaultNewsCacheTTL = 7 * 24 * time.Hour

// NewsRepository writes v2 news into PostgreSQL and refreshes Redis hot cache.
type NewsRepository struct {
	db       *gorm.DB
	cacheTTL time.Duration
}

// NewNewsRepository creates a repository backed by the global PostgreSQL handle.
func NewNewsRepository() *NewsRepository {
	return NewNewsRepositoryWithDB(database.Orm.DB())
}

// NewNewsRepositoryWithDB creates a repository with an explicit PostgreSQL handle.
func NewNewsRepositoryWithDB(db *gorm.DB) *NewsRepository {
	return &NewsRepository{
		db:       db,
		cacheTTL: defaultNewsCacheTTL,
	}
}

// SaveNews upserts one batch of v2 news and refreshes per-language Redis cache.
func (r *NewsRepository) SaveNews(ctx context.Context, items []domain.GameNews) error {
	if len(items) == 0 {
		return nil
	}
	if r == nil || r.db == nil {
		return fmt.Errorf("news repository database is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			if err := upsertNews(ctx, tx, item); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	r.refreshCache(items)
	return nil
}

func upsertNews(ctx context.Context, tx *gorm.DB, item domain.GameNews) error {
	tags, err := marshalJSON(item.Tags)
	if err != nil {
		return fmt.Errorf("marshal news tags: %w", err)
	}
	rawEvent := item.RawEvent
	if len(rawEvent) == 0 {
		rawEvent = []byte("{}")
	}

	err = tx.WithContext(ctx).Exec(`
INSERT INTO gfg_game_v2_news (
    game_id,
    appid,
    lang,
    event_gid,
    announcement_gid,
    forum_topic_id,
    headline,
    raw_body,
    html,
    plain_text,
    summary,
    url,
    tags,
    vote_up_count,
    vote_down_count,
    comment_count,
    raw_event,
    published_at,
    updated_at,
    collected_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?::jsonb, ?, ?, ?, ?::jsonb, ?, ?, ?
)
ON CONFLICT (appid, lang, event_gid, announcement_gid)
DO UPDATE SET
    game_id = EXCLUDED.game_id,
    forum_topic_id = EXCLUDED.forum_topic_id,
    headline = EXCLUDED.headline,
    raw_body = EXCLUDED.raw_body,
    html = EXCLUDED.html,
    plain_text = EXCLUDED.plain_text,
    summary = EXCLUDED.summary,
    url = EXCLUDED.url,
    tags = EXCLUDED.tags,
    vote_up_count = EXCLUDED.vote_up_count,
    vote_down_count = EXCLUDED.vote_down_count,
    comment_count = EXCLUDED.comment_count,
    raw_event = EXCLUDED.raw_event,
    published_at = EXCLUDED.published_at,
    updated_at = EXCLUDED.updated_at,
    collected_at = EXCLUDED.collected_at
`,
		item.GameID,
		item.AppID,
		string(item.Language),
		item.EventGID,
		item.AnnouncementGID,
		item.ForumTopicID,
		item.Headline,
		item.RawBody,
		item.HTML,
		item.PlainText,
		item.Summary,
		item.URL,
		string(tags),
		item.VoteUpCount,
		item.VoteDownCount,
		item.CommentCount,
		string(rawEvent),
		nullableTime(item.PublishedAt),
		nullableTime(item.UpdatedAt),
		item.CollectedAt,
	).Error
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

func newsCacheKey(gameID int64, lang domain.Language) string {
	return fmt.Sprintf("game:v2:news:%d:%s", gameID, lang)
}

func marshalJSON(value any) ([]byte, error) {
	return json.Marshal(value)
}

func nullableTime(value time.Time) any {
	if value.IsZero() {
		return sql.NullTime{}
	}
	return value
}
