package dao

import (
	"context"
	"errors"
	"fmt"
	"strings"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	"github.com/gofurry/gofurry-game-backend/common"
	gamesqlc "github.com/gofurry/gofurry-game-backend/internal/db/game/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableNameGfgGame = "gfg_game"

const (
	siteColumns = `id, name, name_en, info, info_en,
resources::text AS resources, groups::text AS groups, links::text AS links,
appid, header, view_count, weight, create_time, update_time`

	siteColumnsG = `g.id, g.name, g.name_en, g.info, g.info_en,
g.resources::text AS resources, g.groups::text AS groups, g.links::text AS links,
g.appid, g.header, g.view_count, g.weight, g.create_time, g.update_time`

	detailsColumns = `game_id, appid, source, type, name, is_free, website, header_url,
developers::text AS developers, publishers::text AS publishers,
release_coming_soon, release_date_text, platforms::text AS platforms,
supported_languages::text AS supported_languages, support_info::text AS support_info,
content_descriptors::text AS content_descriptors, ratings::text AS ratings,
collected_at, updated_at`

	localizedColumns = `game_id, appid, lang, name, short_description,
detailed_description, about_the_game, collected_at, updated_at`

	priceColumns = `game_id, appid, region, is_free, currency, initial_amount,
final_amount, discount_percent, initial_formatted, final_formatted, collected_at, updated_at`

	mediaColumns = `id, game_id, appid, media_type, media_key, title, url,
thumbnail_url, extra::text AS extra, sort_order, collected_at, updated_at`

	assetColumns = `id, game_id, appid, asset_type, asset_family, source, lang,
media_key, title, url, thumbnail_url, format, exists, status_code, content_type,
content_length, extra::text AS extra, sort_order, checked_at, collected_at, updated_at`

	requirementsColumns = `game_id, appid, pc::text AS pc, mac::text AS mac,
linux::text AS linux, collected_at, updated_at`

	newsColumns = `id, game_id, appid, lang, event_gid, announcement_gid,
forum_topic_id, headline, raw_body, html, plain_text, summary, url,
tags::text AS tags, vote_up_count, vote_down_count, comment_count,
raw_event::text AS raw_event, COALESCE(published_at, collected_at) AS published_at,
COALESCE(updated_at, collected_at) AS updated_at, collected_at`

	newsColumnsN = `n.id, n.game_id, n.appid, n.lang, n.event_gid, n.announcement_gid,
n.forum_topic_id, n.headline, n.raw_body, n.html, n.plain_text, n.summary, n.url,
n.tags::text AS tags, n.vote_up_count, n.vote_down_count, n.comment_count,
n.raw_event::text AS raw_event, COALESCE(n.published_at, n.collected_at) AS published_at,
COALESCE(n.updated_at, n.collected_at) AS updated_at, n.collected_at`

	playerCountColumns = `id, run_id, game_id, appid, count, status,
upstream_status_code, error_kind, error_message, collected_at`

	collectRunColumns = `id, task_type, status, total_count, success_count,
failed_count, skipped_count, partial_count, task_summary::text AS task_summary,
duration_millis, error_kind, error_message, started_at, ended_at`

	collectTaskResultColumns = `id, run_id, task_type, status, game_id, appid,
upstream_status_code, traffic_bucket, retry_count, duration_millis,
error_kind, error_message, started_at, ended_at`
)

type DetailQuery = v2models.GameV2DetailQuery

// ReadModelDAO owns no process-global database state. The pool is created by
// bootstrap and constructor-injected here.
type ReadModelDAO struct {
	pool *pgxpool.Pool
	q    *gamesqlc.Queries
}

func NewReadModelDAO(pool *pgxpool.Pool) *ReadModelDAO {
	return &ReadModelDAO{pool: pool, q: gamesqlc.New(pool)}
}

func (dao *ReadModelDAO) ready() error {
	if dao == nil || dao.pool == nil {
		return errors.New("game v2 read model database is not initialized")
	}
	return nil
}

// The read model has bounded dynamic filtering and ordering that sqlc cannot
// reasonably express without multiplying near-identical queries. These local
// helpers are the documented direct-pgx exception: callers supply fixed SQL,
// every value remains parameterized, and no CRUD/repository abstraction is
// exposed outside this package.
func queryMany[T any](ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) ([]T, error) {
	rows, err := pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByNameLax[T])
}

func queryOptional[T any](ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) (*T, error) {
	rows, err := pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	value, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[T])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func (dao *ReadModelDAO) GetGameDetailAggregate(ctx context.Context, query DetailQuery) (v2models.GameV2Aggregate, common.GFError) {
	var aggregate v2models.GameV2Aggregate
	if err := dao.ready(); err != nil {
		return aggregate, common.NewDaoError(err.Error())
	}
	if query.GameID <= 0 && query.AppID <= 0 {
		return aggregate, common.NewDaoError("game_id or appid is required")
	}
	if query.Lang == "" {
		query.Lang = "zh"
	}
	if query.NewsLimit <= 0 {
		query.NewsLimit = 5
	}

	site, err := dao.loadSiteRecord(ctx, query)
	if err != nil {
		return aggregate, common.NewDaoError(err.Error())
	}
	aggregate.Site = *site
	if err := dao.loadAggregateExtras(ctx, &aggregate, query.Lang, query.NewsLimit); err != nil {
		return aggregate, common.NewDaoError(err.Error())
	}
	return aggregate, nil
}

func (dao *ReadModelDAO) ListGameAggregates(ctx context.Context, query v2models.GameV2ListQuery) ([]v2models.GameV2Aggregate, common.GFError) {
	if err := dao.ready(); err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	if query.Lang == "" {
		query.Lang = "zh"
	}
	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.Offset < 0 {
		query.Offset = 0
	}

	join := ""
	if query.Sort == "release_date" {
		join = " JOIN gfg_game_first_available fa ON fa.game_id = g.id"
	}
	args := []any{}
	where := ""
	if !query.UpdatedSince.IsZero() {
		args = append(args, query.UpdatedSince)
		where = ` WHERE g.update_time >= $1
OR EXISTS (SELECT 1 FROM gfg_game_v2_details d2 WHERE d2.game_id = g.id AND d2.updated_at >= $1)
OR EXISTS (SELECT 1 FROM gfg_game_v2_localized_details ld WHERE ld.game_id = g.id AND ld.updated_at >= $1)
OR EXISTS (SELECT 1 FROM gfg_game_v2_prices p WHERE p.game_id = g.id AND p.updated_at >= $1)
OR EXISTS (SELECT 1 FROM gfg_game_v2_media m WHERE m.game_id = g.id AND m.updated_at >= $1)
OR EXISTS (SELECT 1 FROM gfg_game_v2_assets a WHERE a.game_id = g.id AND a.updated_at >= $1)
OR EXISTS (SELECT 1 FROM gfg_game_v2_requirements r WHERE r.game_id = g.id AND r.updated_at >= $1)
OR EXISTS (SELECT 1 FROM gfg_game_v2_news n WHERE n.game_id = g.id AND n.updated_at >= $1)`
	}
	args = append(args, query.Limit, query.Offset)
	limitPos := len(args) - 1
	sql := fmt.Sprintf(`SELECT %s FROM gfg_game g%s%s ORDER BY %s LIMIT $%d OFFSET $%d`,
		siteColumnsG, join, where, listOrder(query.Sort), limitPos, limitPos+1)
	sites, err := queryMany[v2models.GameV2SiteRecord](ctx, dao.pool, sql, args...)
	if err != nil {
		return nil, common.NewDaoError(fmt.Sprintf("查询游戏 v2 列表失败: %v", err))
	}

	aggregates, err := dao.loadAggregatesBySites(ctx, sites, query.Lang, 0)
	if err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	return aggregates, nil
}

func (dao *ReadModelDAO) ListTags(ctx context.Context, lang string) ([]v2models.GameV2TagRecord, common.GFError) {
	if err := dao.ready(); err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	nameColumn := "COALESCE(NULLIF(t.name, ''), t.name_en)"
	if normalizeDAOLang(lang) == "en" {
		nameColumn = "COALESCE(NULLIF(t.name_en, ''), t.name)"
	}
	sql := fmt.Sprintf(`SELECT t.id::text AS id, %s AS name, t.prefix::text AS prefix,
COALESCE(tc.game_count, 0)::integer AS game_count
FROM gfg_tag t
LEFT JOIN (
    SELECT tm.tag_id, COUNT(DISTINCT tm.game_id) AS game_count
    FROM gfg_tag_map tm
    JOIN gfg_game_v2_details d ON d.game_id = tm.game_id
    GROUP BY tm.tag_id
) tc ON t.id = tc.tag_id
ORDER BY game_count DESC, t.id ASC`, nameColumn)
	rows, err := queryMany[v2models.GameV2TagRecord](ctx, dao.pool, sql)
	if err != nil {
		return nil, common.NewDaoError(fmt.Sprintf("查询游戏 v2 标签失败: %v", err))
	}
	return rows, nil
}

func (dao *ReadModelDAO) GetGameReviews(ctx context.Context, query v2models.GameV2ReviewQuery) (v2models.GameV2ReviewList, common.GFError) {
	res := v2models.GameV2ReviewList{}
	if err := dao.ready(); err != nil {
		return res, common.NewDaoError(err.Error())
	}
	if query.GameID <= 0 {
		return res, common.NewDaoError("game_id is required")
	}
	if query.PageNum <= 0 {
		query.PageNum = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 5
	}
	if err := dao.pool.QueryRow(ctx, `SELECT COUNT(*)::integer, COALESCE(AVG(score), 0)::double precision
FROM gfg_game_comment WHERE game_id = $1`, query.GameID).Scan(&res.Total, &res.AvgScore); err != nil {
		return res, common.NewDaoError(fmt.Sprintf("统计游戏 v2 评论失败: %v", err))
	}
	res.PageNum = query.PageNum
	res.PageSize = query.PageSize
	res.Remarks = []v2models.GameV2ReviewItem{}
	if res.Total == 0 {
		return res, nil
	}
	rows, err := queryMany[v2models.GameV2ReviewItem](ctx, dao.pool, `
SELECT region, content, score, create_time, ip, name
FROM gfg_game_comment WHERE game_id = $1
ORDER BY create_time DESC, id DESC LIMIT $2 OFFSET $3`,
		query.GameID, query.PageSize, (query.PageNum-1)*query.PageSize)
	if err != nil {
		return res, common.NewDaoError(fmt.Sprintf("查询游戏 v2 评论失败: %v", err))
	}
	res.Remarks = rows
	return res, nil
}

func (dao *ReadModelDAO) ListLatestReviews(ctx context.Context, lang string, limit int) ([]v2models.GameV2LatestReview, common.GFError) {
	if err := dao.ready(); err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	if limit <= 0 {
		limit = 5
	}
	nameColumn := "COALESCE(NULLIF(ld.name, ''), NULLIF(g.name, ''), NULLIF(d.name, ''), g.name_en)"
	if normalizeDAOLang(lang) == "en" {
		nameColumn = "COALESCE(NULLIF(ld.name, ''), NULLIF(d.name, ''), NULLIF(g.name_en, ''), g.name)"
	}
	sql := fmt.Sprintf(`SELECT c.region, c.score, c.content, c.ip, c.create_time AS time,
COALESCE(NULLIF(d.header_url, ''), NULLIF(g.header, ''), '') AS game_cover,
%s AS game_name
FROM gfg_game_comment c
JOIN gfg_game g ON c.game_id = g.id
LEFT JOIN gfg_game_v2_details d ON d.game_id = g.id
LEFT JOIN gfg_game_v2_localized_details ld ON ld.game_id = g.id AND ld.lang = $1
ORDER BY c.create_time DESC, c.id DESC LIMIT $2`, nameColumn)
	rows, err := queryMany[v2models.GameV2LatestReview](ctx, dao.pool, sql, normalizeDAOLang(lang), limit)
	if err != nil {
		return nil, common.NewDaoError(fmt.Sprintf("查询游戏 v2 最新评论失败: %v", err))
	}
	return rows, nil
}

func (dao *ReadModelDAO) GetRandomGameID(ctx context.Context) (string, common.GFError) {
	if err := dao.ready(); err != nil {
		return "", common.NewDaoError(err.Error())
	}
	var id string
	err := dao.pool.QueryRow(ctx, `SELECT g.id::text FROM gfg_game g
JOIN gfg_game_v2_details d ON d.game_id = g.id ORDER BY RANDOM() LIMIT 1`).Scan(&id)
	if err != nil {
		return "", common.NewDaoError(fmt.Sprintf("随机查询游戏 v2 ID 失败: %v", err))
	}
	return id, nil
}

func (dao *ReadModelDAO) GetGameNews(ctx context.Context, query v2models.GameV2NewsQuery) ([]v2models.GameV2NewsRow, common.GFError) {
	if err := dao.ready(); err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	rows, err := dao.queryNewsRows(ctx, query, true)
	if err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	return rows, nil
}

func (dao *ReadModelDAO) GetLatestGameNews(ctx context.Context, query v2models.GameV2NewsQuery) ([]v2models.GameV2NewsRow, common.GFError) {
	if err := dao.ready(); err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	rows, err := dao.queryNewsRows(ctx, query, false)
	if err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	return rows, nil
}

func (dao *ReadModelDAO) queryNewsRows(ctx context.Context, query v2models.GameV2NewsQuery, requireGame bool) ([]v2models.GameV2NewsRow, error) {
	if query.Lang == "" {
		query.Lang = "zh"
	}
	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.Offset < 0 {
		query.Offset = 0
	}
	args := []any{query.Lang}
	where := "n.lang = $1"
	if requireGame {
		if query.GameID > 0 {
			args = append(args, query.GameID)
			where += fmt.Sprintf(" AND n.game_id = $%d", len(args))
		} else if query.AppID > 0 {
			args = append(args, query.AppID)
			where += fmt.Sprintf(" AND n.appid = $%d", len(args))
		} else {
			return []v2models.GameV2NewsRow{}, errors.New("game_id or appid is required")
		}
	}
	if !query.UpdatedSince.IsZero() {
		args = append(args, query.UpdatedSince)
		where += fmt.Sprintf(" AND n.updated_at >= $%d", len(args))
	}
	args = append(args, query.Limit, query.Offset)
	sql := fmt.Sprintf(`SELECT %s, COALESCE(g.name, '') AS game_name,
COALESCE(g.name_en, '') AS game_name_en, COALESCE(g.header, '') AS header_url
FROM gfg_game_v2_news n LEFT JOIN gfg_game g ON n.game_id = g.id
WHERE %s ORDER BY n.published_at DESC NULLS LAST, n.collected_at DESC, n.id DESC
LIMIT $%d OFFSET $%d`, newsColumnsN, where, len(args)-1, len(args))
	rows, err := queryMany[v2models.GameV2NewsRow](ctx, dao.pool, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("查询游戏 v2 新闻失败: %w", err)
	}
	if len(rows) > 0 || query.Lang == "zh" {
		return rows, nil
	}
	query.Lang = "zh"
	return dao.queryNewsRows(ctx, query, requireGame)
}

func normalizeDAORegion(region string) string {
	region = strings.ToUpper(strings.TrimSpace(region))
	if region == "" {
		return "CN"
	}
	return region
}

func normalizeDAOLang(lang string) string {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "en", "en-us", "en_us":
		return "en"
	default:
		return "zh"
	}
}

func strPtrValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func listOrder(sort string) string {
	switch sort {
	case "release_date":
		return "fa.window_end DESC, g.id DESC"
	case "newest":
		return "g.create_time DESC, g.id DESC"
	case "updated":
		return "g.update_time DESC, g.id DESC"
	default:
		return "g.weight ASC, g.id ASC"
	}
}
