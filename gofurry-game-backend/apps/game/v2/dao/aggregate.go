package dao

import (
	"context"
	"fmt"
	"strings"
	"time"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/roof/env"
)

func (dao *ReadModelDAO) loadSiteRecord(ctx context.Context, query DetailQuery) (*v2models.GameV2SiteRecord, error) {
	where := "id = $1"
	value := query.GameID
	if query.GameID <= 0 {
		where = "appid = $1"
		value = query.AppID
	}
	row, err := queryOptional[v2models.GameV2SiteRecord](ctx, dao.pool,
		"SELECT "+siteColumns+" FROM gfg_game WHERE "+where, value)
	if err != nil {
		return nil, fmt.Errorf("查询站内游戏主档案失败: %w", err)
	}
	if row == nil {
		return nil, fmt.Errorf("查询站内游戏主档案失败: game not found")
	}
	return row, nil
}

func (dao *ReadModelDAO) loadAggregateExtras(ctx context.Context, aggregate *v2models.GameV2Aggregate, lang string, newsLimit int) error {
	gameID := aggregate.Site.ID
	details, err := queryOptional[v2models.GfgGameV2Details](ctx, dao.pool,
		"SELECT "+detailsColumns+" FROM gfg_game_v2_details WHERE game_id = $1", gameID)
	if err != nil {
		return fmt.Errorf("查询游戏 v2 详情失败: %w", err)
	}
	aggregate.Details = details

	localized, err := dao.loadLocalized(ctx, gameID, lang)
	if err != nil {
		return err
	}
	aggregate.Localized = localized

	if aggregate.Prices, err = queryMany[v2models.GfgGameV2Price](ctx, dao.pool,
		"SELECT "+priceColumns+" FROM gfg_game_v2_prices WHERE game_id = $1 ORDER BY region ASC", gameID); err != nil {
		return fmt.Errorf("查询游戏 v2 价格失败: %w", err)
	}
	if aggregate.Media, err = queryMany[v2models.GfgGameV2Media](ctx, dao.pool,
		"SELECT "+mediaColumns+" FROM gfg_game_v2_media WHERE game_id = $1 ORDER BY media_type, sort_order, id", gameID); err != nil {
		return fmt.Errorf("查询游戏 v2 媒体失败: %w", err)
	}
	if aggregate.Assets, err = queryMany[v2models.GfgGameV2Asset](ctx, dao.pool,
		"SELECT "+assetColumns+" FROM gfg_game_v2_assets WHERE game_id = $1 ORDER BY asset_family, sort_order, id", gameID); err != nil {
		return fmt.Errorf("查询游戏 v2 统一媒体资产失败: %w", err)
	}
	aggregate.Requirements, err = queryOptional[v2models.GfgGameV2Requirements](ctx, dao.pool,
		"SELECT "+requirementsColumns+" FROM gfg_game_v2_requirements WHERE game_id = $1", gameID)
	if err != nil {
		return fmt.Errorf("查询游戏 v2 配置需求失败: %w", err)
	}
	if newsLimit > 0 {
		aggregate.News, err = dao.loadNews(ctx, gameID, lang, newsLimit)
		if err != nil {
			return err
		}
	}
	aggregate.OnlineCount, err = queryOptional[v2models.GfgGameV2PlayerCount](ctx, dao.pool,
		"SELECT "+playerCountColumns+" FROM gfg_game_v2_player_counts WHERE game_id = $1 AND status = 'success' ORDER BY collected_at DESC, id DESC LIMIT 1", gameID)
	if err != nil {
		return fmt.Errorf("查询游戏 v2 在线人数失败: %w", err)
	}
	if aggregate.OnlineCount != nil {
		if err := dao.loadOnlinePeakForRow(ctx, aggregate.OnlineCount); err != nil {
			return err
		}
	}
	if err := dao.pool.QueryRow(ctx, `SELECT COALESCE(AVG(score), 0)::double precision, COUNT(*)::bigint
FROM gfg_game_comment WHERE game_id = $1`, gameID).
		Scan(&aggregate.ReviewStats.AvgScore, &aggregate.ReviewStats.CommentCount); err != nil {
		return fmt.Errorf("查询游戏 v2 评论统计失败: %w", err)
	}
	aggregate.Tags, err = dao.loadTags(ctx, gameID, lang)
	if err != nil {
		return err
	}
	return nil
}

func (dao *ReadModelDAO) loadLocalized(ctx context.Context, gameID int64, lang string) (*v2models.GfgGameV2LocalizedDetails, error) {
	requestedLang := normalizeDAOLang(lang)
	primary, err := queryOptional[v2models.GfgGameV2LocalizedDetails](ctx, dao.pool,
		"SELECT "+localizedColumns+" FROM gfg_game_v2_localized_details WHERE game_id = $1 AND lang = $2", gameID, requestedLang)
	if err != nil {
		return nil, fmt.Errorf("查询游戏 v2 本地化详情失败: %w", err)
	}
	fallbackLang := localizedFallbackLang(requestedLang)
	if fallbackLang == "" {
		return primary, nil
	}
	fallback, err := queryOptional[v2models.GfgGameV2LocalizedDetails](ctx, dao.pool,
		"SELECT "+localizedColumns+" FROM gfg_game_v2_localized_details WHERE game_id = $1 AND lang = $2", gameID, fallbackLang)
	if err != nil {
		return nil, fmt.Errorf("查询游戏 v2 回退详情失败: %w", err)
	}
	return mergeLocalizedDetails(primary, fallback), nil
}

func localizedFallbackLang(lang string) string {
	if normalizeDAOLang(lang) == "en" {
		return "zh"
	}
	return "en"
}

func mergeLocalizedDetails(primary *v2models.GfgGameV2LocalizedDetails, fallback *v2models.GfgGameV2LocalizedDetails) *v2models.GfgGameV2LocalizedDetails {
	if primary == nil {
		return fallback
	}
	if fallback == nil {
		return primary
	}
	merged := *primary
	if strings.TrimSpace(merged.Name) == "" {
		merged.Name = fallback.Name
	}
	if merged.GameID == 0 {
		merged.GameID = fallback.GameID
	}
	if merged.AppID == 0 {
		merged.AppID = fallback.AppID
	}
	if strings.TrimSpace(merged.Lang) == "" {
		merged.Lang = fallback.Lang
	}
	if merged.CollectedAt.IsZero() {
		merged.CollectedAt = fallback.CollectedAt
	}
	if merged.UpdatedAt.IsZero() {
		merged.UpdatedAt = fallback.UpdatedAt
	}
	merged.ShortDescription = chooseLocalizedText(primary.ShortDescription, fallback.ShortDescription)
	merged.DetailedDescription = chooseLocalizedText(primary.DetailedDescription, fallback.DetailedDescription)
	merged.AboutTheGame = chooseLocalizedText(primary.AboutTheGame, fallback.AboutTheGame)
	return &merged
}

func chooseLocalizedText(primary *string, fallback *string) *string {
	if strings.TrimSpace(strPtrValue(primary)) != "" {
		return primary
	}
	if strings.TrimSpace(strPtrValue(fallback)) != "" {
		return fallback
	}
	if primary != nil {
		return primary
	}
	return fallback
}

func (dao *ReadModelDAO) loadNews(ctx context.Context, gameID int64, lang string, limit int) ([]v2models.GfgGameV2News, error) {
	lang = normalizeDAOLang(lang)
	rows, err := queryMany[v2models.GfgGameV2News](ctx, dao.pool,
		"SELECT "+newsColumns+" FROM gfg_game_v2_news WHERE game_id = $1 AND lang = $2 ORDER BY published_at DESC NULLS LAST, id DESC LIMIT $3",
		gameID, lang, limit)
	if err != nil {
		return nil, fmt.Errorf("查询游戏 v2 新闻失败: %w", err)
	}
	if len(rows) > 0 || lang == "zh" {
		return rows, nil
	}
	rows, err = queryMany[v2models.GfgGameV2News](ctx, dao.pool,
		"SELECT "+newsColumns+" FROM gfg_game_v2_news WHERE game_id = $1 AND lang = 'zh' ORDER BY published_at DESC NULLS LAST, id DESC LIMIT $2",
		gameID, limit)
	if err != nil {
		return nil, fmt.Errorf("查询游戏 v2 中文回退新闻失败: %w", err)
	}
	return rows, nil
}

func (dao *ReadModelDAO) loadTags(ctx context.Context, gameID int64, lang string) ([]v2models.GameV2Tag, error) {
	nameColumns := "t.name AS name, t.info AS desc"
	if normalizeDAOLang(lang) == "en" {
		nameColumns = "t.name_en AS name, t.info_en AS desc"
	}
	rows, err := queryMany[v2models.GameV2Tag](ctx, dao.pool, fmt.Sprintf(`
SELECT t.id::text AS id, %s FROM gfg_tag_map tm
JOIN gfg_tag t ON tm.tag_id = t.id WHERE tm.game_id = $1 ORDER BY t.id`, nameColumns), gameID)
	if err != nil {
		return nil, fmt.Errorf("查询游戏 v2 标签失败: %w", err)
	}
	return rows, nil
}

func (dao *ReadModelDAO) loadAggregatesBySites(ctx context.Context, sites []v2models.GameV2SiteRecord, lang string, newsLimit int) ([]v2models.GameV2Aggregate, error) {
	if len(sites) == 0 {
		return []v2models.GameV2Aggregate{}, nil
	}
	gameIDs := make([]int64, 0, len(sites))
	aggregateMap := make(map[int64]*v2models.GameV2Aggregate, len(sites))
	for i := range sites {
		gameIDs = append(gameIDs, sites[i].ID)
		aggregateMap[sites[i].ID] = &v2models.GameV2Aggregate{Site: sites[i]}
	}
	if err := dao.loadAggregateExtrasBatch(ctx, aggregateMap, gameIDs, lang); err != nil {
		return nil, err
	}
	if newsLimit > 0 {
		for _, gameID := range gameIDs {
			rows, err := dao.loadNews(ctx, gameID, lang, newsLimit)
			if err != nil {
				return nil, err
			}
			aggregateMap[gameID].News = rows
		}
	}
	result := make([]v2models.GameV2Aggregate, 0, len(sites))
	for _, site := range sites {
		result = append(result, *aggregateMap[site.ID])
	}
	return result, nil
}

func (dao *ReadModelDAO) loadAggregatesByGameIDs(ctx context.Context, gameIDs []int64, lang string, newsLimit int) ([]v2models.GameV2Aggregate, error) {
	if len(gameIDs) == 0 {
		return []v2models.GameV2Aggregate{}, nil
	}
	unique := uniqueInt64s(gameIDs)
	sites, err := queryMany[v2models.GameV2SiteRecord](ctx, dao.pool,
		"SELECT "+siteColumns+" FROM gfg_game WHERE id = ANY($1::bigint[])", unique)
	if err != nil {
		return nil, fmt.Errorf("查询游戏 v2 面板站内档案失败: %w", err)
	}
	byID := make(map[int64]v2models.GameV2SiteRecord, len(sites))
	for _, site := range sites {
		byID[site.ID] = site
	}
	ordered := make([]v2models.GameV2SiteRecord, 0, len(gameIDs))
	for _, gameID := range gameIDs {
		if site, ok := byID[gameID]; ok {
			ordered = append(ordered, site)
		}
	}
	return dao.loadAggregatesBySites(ctx, ordered, lang, newsLimit)
}

func (dao *ReadModelDAO) loadAggregateExtrasBatch(ctx context.Context, aggregateMap map[int64]*v2models.GameV2Aggregate, gameIDs []int64, lang string) error {
	details, err := queryMany[v2models.GfgGameV2Details](ctx, dao.pool,
		"SELECT "+detailsColumns+" FROM gfg_game_v2_details WHERE game_id = ANY($1::bigint[])", gameIDs)
	if err != nil {
		return fmt.Errorf("批量查询游戏 v2 详情失败: %w", err)
	}
	for i := range details {
		aggregateMap[details[i].GameID].Details = &details[i]
	}

	requested := normalizeDAOLang(lang)
	fallback := localizedFallbackLang(requested)
	localized, err := queryMany[v2models.GfgGameV2LocalizedDetails](ctx, dao.pool,
		"SELECT "+localizedColumns+" FROM gfg_game_v2_localized_details WHERE game_id = ANY($1::bigint[]) AND lang = ANY($2::text[])", gameIDs, []string{requested, fallback})
	if err != nil {
		return fmt.Errorf("批量查询游戏 v2 本地化详情失败: %w", err)
	}
	primaryRows := map[int64]*v2models.GfgGameV2LocalizedDetails{}
	fallbackRows := map[int64]*v2models.GfgGameV2LocalizedDetails{}
	for i := range localized {
		if normalizeDAOLang(localized[i].Lang) == requested {
			primaryRows[localized[i].GameID] = &localized[i]
		} else {
			fallbackRows[localized[i].GameID] = &localized[i]
		}
	}
	for _, gameID := range gameIDs {
		aggregateMap[gameID].Localized = mergeLocalizedDetails(primaryRows[gameID], fallbackRows[gameID])
	}

	prices, err := queryMany[v2models.GfgGameV2Price](ctx, dao.pool,
		"SELECT "+priceColumns+" FROM gfg_game_v2_prices WHERE game_id = ANY($1::bigint[]) ORDER BY game_id, region", gameIDs)
	if err != nil {
		return fmt.Errorf("批量查询游戏 v2 价格失败: %w", err)
	}
	for _, row := range prices {
		aggregateMap[row.GameID].Prices = append(aggregateMap[row.GameID].Prices, row)
	}

	media, err := queryMany[v2models.GfgGameV2Media](ctx, dao.pool,
		"SELECT "+mediaColumns+" FROM gfg_game_v2_media WHERE game_id = ANY($1::bigint[]) ORDER BY game_id, media_type, sort_order, id", gameIDs)
	if err != nil {
		return fmt.Errorf("批量查询游戏 v2 媒体失败: %w", err)
	}
	for _, row := range media {
		aggregateMap[row.GameID].Media = append(aggregateMap[row.GameID].Media, row)
	}

	assets, err := queryMany[v2models.GfgGameV2Asset](ctx, dao.pool,
		"SELECT "+assetColumns+" FROM gfg_game_v2_assets WHERE game_id = ANY($1::bigint[]) ORDER BY game_id, asset_family, sort_order, id", gameIDs)
	if err != nil {
		return fmt.Errorf("批量查询游戏 v2 统一媒体资产失败: %w", err)
	}
	for _, row := range assets {
		aggregateMap[row.GameID].Assets = append(aggregateMap[row.GameID].Assets, row)
	}

	requirements, err := queryMany[v2models.GfgGameV2Requirements](ctx, dao.pool,
		"SELECT "+requirementsColumns+" FROM gfg_game_v2_requirements WHERE game_id = ANY($1::bigint[])", gameIDs)
	if err != nil {
		return fmt.Errorf("批量查询游戏 v2 配置需求失败: %w", err)
	}
	for i := range requirements {
		aggregateMap[requirements[i].GameID].Requirements = &requirements[i]
	}

	players, err := queryMany[v2models.GfgGameV2PlayerCount](ctx, dao.pool,
		"SELECT "+playerCountColumns+` FROM gfg_game_v2_player_counts
WHERE id IN (SELECT DISTINCT ON (game_id) id FROM gfg_game_v2_player_counts
WHERE game_id = ANY($1::bigint[]) AND status = 'success' ORDER BY game_id, collected_at DESC, id DESC)`, gameIDs)
	if err != nil {
		return fmt.Errorf("批量查询游戏 v2 在线人数失败: %w", err)
	}
	peaks, peakWindow, err := dao.queryOnlinePeakCounts(ctx, gameIDs)
	if err != nil {
		return err
	}
	for i := range players {
		applyOnlinePeak(&players[i], peaks, peakWindow)
		aggregateMap[players[i].GameID].OnlineCount = &players[i]
	}

	type reviewStatsRow struct {
		GameID       int64   `db:"game_id"`
		AvgScore     float64 `db:"avg_score"`
		CommentCount int64   `db:"comment_count"`
	}
	stats, err := queryMany[reviewStatsRow](ctx, dao.pool, `SELECT game_id,
COALESCE(AVG(score), 0)::double precision AS avg_score, COUNT(*)::bigint AS comment_count
FROM gfg_game_comment WHERE game_id = ANY($1::bigint[]) GROUP BY game_id`, gameIDs)
	if err != nil {
		return fmt.Errorf("批量查询游戏 v2 评论统计失败: %w", err)
	}
	for _, row := range stats {
		aggregateMap[row.GameID].ReviewStats = v2models.GameV2ReviewStats{AvgScore: row.AvgScore, CommentCount: row.CommentCount}
	}

	type tagRow struct {
		GameID int64  `db:"game_id"`
		ID     string `db:"id"`
		Name   string `db:"name"`
		Desc   string `db:"desc"`
	}
	tagNames := "t.name AS name, t.info AS desc"
	if requested == "en" {
		tagNames = "t.name_en AS name, t.info_en AS desc"
	}
	tags, err := queryMany[tagRow](ctx, dao.pool, fmt.Sprintf(`SELECT tm.game_id, t.id::text AS id, %s
FROM gfg_tag_map tm JOIN gfg_tag t ON tm.tag_id = t.id
WHERE tm.game_id = ANY($1::bigint[]) ORDER BY tm.game_id, t.id`, tagNames), gameIDs)
	if err != nil {
		return fmt.Errorf("批量查询游戏 v2 标签失败: %w", err)
	}
	for _, row := range tags {
		aggregateMap[row.GameID].Tags = append(aggregateMap[row.GameID].Tags, v2models.GameV2Tag{ID: row.ID, Name: row.Name, Desc: row.Desc})
	}
	return nil
}

func (dao *ReadModelDAO) loadOnlinePeakForRow(ctx context.Context, row *v2models.GfgGameV2PlayerCount) error {
	peaks, window, err := dao.queryOnlinePeakCounts(ctx, []int64{row.GameID})
	if err != nil {
		return err
	}
	applyOnlinePeak(row, peaks, window)
	return nil
}

func (dao *ReadModelDAO) queryOnlinePeakCounts(ctx context.Context, gameIDs []int64) (map[int64]int64, int, error) {
	window := env.GetServerConfig().OnlinePeakCacheDays()
	result := make(map[int64]int64, len(gameIDs))
	if len(gameIDs) == 0 {
		return result, window, nil
	}
	type peakRow struct {
		GameID    int64 `db:"game_id"`
		PeakCount int64 `db:"peak_count"`
	}
	rows, err := queryMany[peakRow](ctx, dao.pool, `SELECT game_id,
COALESCE(MAX(count), 0)::bigint AS peak_count FROM gfg_game_v2_player_counts
WHERE game_id = ANY($1::bigint[]) AND status = 'success' AND collected_at >= $2
GROUP BY game_id`, gameIDs, time.Now().AddDate(0, 0, -window))
	if err != nil {
		return nil, window, fmt.Errorf("批量查询游戏 v2 在线峰值失败: %w", err)
	}
	for _, row := range rows {
		result[row.GameID] = row.PeakCount
	}
	return result, window, nil
}

func applyOnlinePeak(row *v2models.GfgGameV2PlayerCount, peaks map[int64]int64, window int) {
	if row == nil {
		return
	}
	row.PeakWindowDays = window
	row.PeakCount = peaks[row.GameID]
	if row.PeakCount < row.Count {
		row.PeakCount = row.Count
	}
}

func uniqueInt64s(values []int64) []int64 {
	seen := make(map[int64]struct{}, len(values))
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func (dao *ReadModelDAO) listPanelAggregatesBySQL(ctx context.Context, query v2models.GameV2PanelQuery, sql string, args ...any) ([]v2models.GameV2Aggregate, common.GFError) {
	if err := dao.ready(); err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	if query.Lang == "" {
		query.Lang = "zh"
	}
	if query.Limit <= 0 {
		query.Limit = 8
	}
	rows, err := dao.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, common.NewDaoError(fmt.Sprintf("查询游戏 v2 面板候选失败: %v", err))
	}
	gameIDs, err := pgxCollectInt64(rows)
	if err != nil {
		return nil, common.NewDaoError(fmt.Sprintf("查询游戏 v2 面板候选失败: %v", err))
	}
	aggregates, err := dao.loadAggregatesByGameIDs(ctx, gameIDs, query.Lang, 0)
	if err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	return aggregates, nil
}

func pgxCollectInt64(rows interface {
	Next() bool
	Scan(...any) error
	Err() error
	Close()
}) ([]int64, error) {
	defer rows.Close()
	result := []int64{}
	for rows.Next() {
		var value int64
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}

func (dao *ReadModelDAO) ListTopOnlineAggregates(ctx context.Context, query v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError) {
	if query.Limit <= 0 {
		query.Limit = 8
	}
	return dao.listPanelAggregatesBySQL(ctx, query, `SELECT game_id FROM (
SELECT DISTINCT ON (game_id) game_id, count, collected_at, id
FROM gfg_game_v2_player_counts WHERE status = 'success'
ORDER BY game_id, collected_at DESC, id DESC) latest
ORDER BY count DESC, collected_at DESC LIMIT $1`, query.Limit)
}

func (dao *ReadModelDAO) ListFreeGameAggregates(ctx context.Context, query v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError) {
	if query.Limit <= 0 {
		query.Limit = 8
	}
	return dao.listPanelAggregatesBySQL(ctx, query, `SELECT p.game_id FROM gfg_game_v2_prices p
JOIN gfg_game g ON p.game_id = g.id WHERE p.region = $1 AND p.is_free = true
ORDER BY random(), p.game_id LIMIT $2`, normalizeDAORegion(query.Region), query.Limit)
}

func (dao *ReadModelDAO) ListPopularGameAggregates(ctx context.Context, query v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError) {
	if query.Limit <= 0 {
		query.Limit = 8
	}
	return dao.listPanelAggregatesBySQL(ctx, query, `SELECT id FROM gfg_game
ORDER BY view_count DESC, update_time DESC, id DESC LIMIT $1`, query.Limit)
}

func (dao *ReadModelDAO) ListHighestPriceAggregates(ctx context.Context, query v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError) {
	if query.Limit <= 0 {
		query.Limit = 8
	}
	return dao.listPanelAggregatesBySQL(ctx, query, pricePanelSQL("p.final_amount DESC, p.discount_percent DESC"), normalizeDAORegion(query.Region), query.Limit)
}

func (dao *ReadModelDAO) ListHighestDiscountAggregates(ctx context.Context, query v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError) {
	if query.Limit <= 0 {
		query.Limit = 8
	}
	return dao.listPanelAggregatesBySQL(ctx, query, pricePanelSQL("p.discount_percent DESC, p.final_amount ASC"), normalizeDAORegion(query.Region), query.Limit)
}

func pricePanelSQL(order string) string {
	extra := ""
	if strings.HasPrefix(order, "p.discount_percent") {
		extra = " AND p.discount_percent > 0"
	}
	return `SELECT p.game_id FROM gfg_game_v2_prices p JOIN gfg_game g ON p.game_id = g.id
WHERE p.region = $1 AND p.is_free = false AND p.final_amount > 0` + extra + `
AND COALESCE(p.currency, '') <> '' AND COALESCE(p.final_formatted, '') <> ''
ORDER BY ` + order + `, g.weight ASC, p.game_id ASC LIMIT $2`
}

func (dao *ReadModelDAO) ListLowPriceAggregates(ctx context.Context, query v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError) {
	region := normalizeDAORegion(query.Region)
	return dao.listPanelAggregatesBySQL(ctx, query, `SELECT game_id FROM (
(SELECT p.game_id, p.final_amount, p.discount_percent, g.weight FROM gfg_game_v2_prices p JOIN gfg_game g ON p.game_id=g.id WHERE p.region=$1 AND p.is_free=false AND p.final_amount>0 AND p.final_amount<=1000 AND COALESCE(p.currency,'')<>'' AND COALESCE(p.final_formatted,'')<>'' ORDER BY p.final_amount DESC,p.discount_percent DESC,g.weight,p.game_id LIMIT 15)
UNION ALL
(SELECT p.game_id, p.final_amount, p.discount_percent, g.weight FROM gfg_game_v2_prices p JOIN gfg_game g ON p.game_id=g.id WHERE p.region=$1 AND p.is_free=false AND p.final_amount>0 AND p.final_amount<=1500 AND COALESCE(p.currency,'')<>'' AND COALESCE(p.final_formatted,'')<>'' ORDER BY p.final_amount DESC,p.discount_percent DESC,g.weight,p.game_id LIMIT 15)
UNION ALL
(SELECT p.game_id, p.final_amount, p.discount_percent, g.weight FROM gfg_game_v2_prices p JOIN gfg_game g ON p.game_id=g.id WHERE p.region=$1 AND p.is_free=false AND p.final_amount>0 AND p.final_amount<=2000 AND COALESCE(p.currency,'')<>'' AND COALESCE(p.final_formatted,'')<>'' ORDER BY p.final_amount DESC,p.discount_percent DESC,g.weight,p.game_id LIMIT 15)
UNION ALL
(SELECT p.game_id, p.final_amount, p.discount_percent, g.weight FROM gfg_game_v2_prices p JOIN gfg_game g ON p.game_id=g.id WHERE p.region=$1 AND p.is_free=false AND p.final_amount>0 AND p.final_amount<=2500 AND COALESCE(p.currency,'')<>'' AND COALESCE(p.final_formatted,'')<>'' ORDER BY p.final_amount DESC,p.discount_percent DESC,g.weight,p.game_id LIMIT 15)
) bucketed GROUP BY game_id ORDER BY MAX(final_amount) DESC,MAX(discount_percent) DESC,MIN(weight),game_id`, region)
}
