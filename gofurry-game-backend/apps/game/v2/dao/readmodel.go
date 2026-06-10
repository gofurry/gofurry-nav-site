package dao

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	"github.com/gofurry/gofurry-game-backend/common"
	database "github.com/gofurry/gofurry-game-backend/roof/db"
	"gorm.io/gorm"
)

const tableNameGfgGame = "gfg_game"

type DetailQuery = v2models.GameV2DetailQuery

type ReadModelDAO struct {
	db *gorm.DB
}

func NewReadModelDAO() *ReadModelDAO {
	return NewReadModelDAOWithDB(database.Orm.DB())
}

func NewReadModelDAOWithDB(db *gorm.DB) *ReadModelDAO {
	return &ReadModelDAO{db: db}
}

func (dao *ReadModelDAO) GetGameDetailAggregate(ctx context.Context, query DetailQuery) (v2models.GameV2Aggregate, common.GFError) {
	var aggregate v2models.GameV2Aggregate
	if dao == nil || dao.db == nil {
		return aggregate, common.NewDaoError("game v2 read model database is not initialized")
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

	db := dao.db.WithContext(ctx)
	if err := dao.loadSiteRecord(db, query, &aggregate.Site); err != nil {
		return aggregate, common.NewDaoError(err.Error())
	}

	if query.AppID <= 0 {
		query.AppID = aggregate.Site.AppID
	}
	if err := dao.loadAggregateExtras(db, &aggregate, query.Lang, query.NewsLimit); err != nil {
		return aggregate, common.NewDaoError(err.Error())
	}

	return aggregate, nil
}

func (dao *ReadModelDAO) ListGameAggregates(ctx context.Context, query v2models.GameV2ListQuery) ([]v2models.GameV2Aggregate, common.GFError) {
	if dao == nil || dao.db == nil {
		return nil, common.NewDaoError("game v2 read model database is not initialized")
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

	var sites []v2models.GameV2SiteRecord
	db := dao.db.WithContext(ctx)
	if err := db.Table(tableNameGfgGame).
		Select("id, name, name_en, info, info_en, resources, groups, links, appid, header, view_count, weight, create_time, update_time").
		Order(listOrder(query.Sort)).
		Limit(query.Limit).
		Offset(query.Offset).
		Find(&sites).Error; err != nil {
		return nil, common.NewDaoError(fmt.Sprintf("查询游戏 v2 列表失败: %v", err))
	}

	aggregates := make([]v2models.GameV2Aggregate, 0, len(sites))
	for _, site := range sites {
		aggregate := v2models.GameV2Aggregate{Site: site}
		if err := dao.loadAggregateExtras(db, &aggregate, query.Lang, 0); err != nil {
			return nil, common.NewDaoError(err.Error())
		}
		aggregates = append(aggregates, aggregate)
	}
	return aggregates, nil
}

func (dao *ReadModelDAO) GetGameNews(ctx context.Context, query v2models.GameV2NewsQuery) ([]v2models.GameV2NewsRow, common.GFError) {
	if dao == nil || dao.db == nil {
		return nil, common.NewDaoError("game v2 read model database is not initialized")
	}
	rows, err := dao.queryNewsRows(dao.db.WithContext(ctx), query, true)
	if err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	return rows, nil
}

func (dao *ReadModelDAO) GetLatestGameNews(ctx context.Context, query v2models.GameV2NewsQuery) ([]v2models.GameV2NewsRow, common.GFError) {
	if dao == nil || dao.db == nil {
		return nil, common.NewDaoError("game v2 read model database is not initialized")
	}
	rows, err := dao.queryNewsRows(dao.db.WithContext(ctx), query, false)
	if err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	return rows, nil
}

func (dao *ReadModelDAO) ListTopOnlineAggregates(ctx context.Context, query v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError) {
	return dao.listPanelAggregatesBySQL(ctx, query, `
SELECT game_id
FROM (
    SELECT DISTINCT ON (game_id) game_id, count, collected_at, id
    FROM gfg_game_v2_player_counts
    WHERE status = 'success'
    ORDER BY game_id, collected_at DESC, id DESC
) latest
ORDER BY count DESC, collected_at DESC
LIMIT ?
`, query.Limit)
}

func (dao *ReadModelDAO) ListFreeGameAggregates(ctx context.Context, query v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError) {
	return dao.listPanelAggregatesBySQL(ctx, query, `
SELECT p.game_id
FROM gfg_game_v2_prices p
JOIN gfg_game g ON p.game_id = g.id
WHERE p.region = ? AND p.is_free = true
ORDER BY p.updated_at DESC, g.weight ASC, p.game_id ASC
LIMIT ?
`, normalizeDAORegion(query.Region), query.Limit)
}

func (dao *ReadModelDAO) ListHighestDiscountAggregates(ctx context.Context, query v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError) {
	return dao.listPanelAggregatesBySQL(ctx, query, `
SELECT p.game_id
FROM gfg_game_v2_prices p
JOIN gfg_game g ON p.game_id = g.id
WHERE p.region = ?
  AND p.is_free = false
  AND p.discount_percent > 0
  AND COALESCE(p.currency, '') <> ''
  AND COALESCE(p.final_formatted, '') <> ''
ORDER BY p.discount_percent DESC, p.final_amount ASC, g.weight ASC, p.game_id ASC
LIMIT ?
`, normalizeDAORegion(query.Region), query.Limit)
}

func (dao *ReadModelDAO) ListLowPriceAggregates(ctx context.Context, query v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError) {
	return dao.listPanelAggregatesBySQL(ctx, query, `
SELECT p.game_id
FROM gfg_game_v2_prices p
JOIN gfg_game g ON p.game_id = g.id
WHERE p.region = ?
  AND p.is_free = false
  AND p.final_amount > 0
  AND COALESCE(p.currency, '') <> ''
  AND COALESCE(p.final_formatted, '') <> ''
ORDER BY p.final_amount ASC, p.discount_percent DESC, g.weight ASC, p.game_id ASC
LIMIT ?
`, normalizeDAORegion(query.Region), query.Limit)
}

func (dao *ReadModelDAO) GetCollectStatus(ctx context.Context) (v2models.GameV2CollectStatus, common.GFError) {
	var res v2models.GameV2CollectStatus
	if dao == nil || dao.db == nil {
		return res, common.NewDaoError("game v2 read model database is not initialized")
	}
	db := dao.db.WithContext(ctx)
	latest, err := takeOptional[v2models.GfgGameV2CollectRun](db.Table(v2models.TableNameGfgGameV2CollectRuns).Order("started_at DESC, id DESC"))
	if err != nil {
		return res, common.NewDaoError(fmt.Sprintf("查询游戏 v2 最新采集批次失败: %v", err))
	}
	res.LatestRun = latest

	if err := db.Raw(`
SELECT DISTINCT ON (task_type) *
FROM gfg_game_v2_collect_runs
ORDER BY task_type, started_at DESC, id DESC
`).Scan(&res.LatestTaskRuns).Error; err != nil {
		return res, common.NewDaoError(fmt.Sprintf("查询游戏 v2 任务最新批次失败: %v", err))
	}

	if err := db.Raw(`
SELECT task_type, status, COUNT(*) AS count
FROM gfg_game_v2_collect_task_results
WHERE started_at >= NOW() - INTERVAL '7 days'
GROUP BY task_type, status
ORDER BY task_type ASC, status ASC
`).Scan(&res.Summary).Error; err != nil {
		return res, common.NewDaoError(fmt.Sprintf("查询游戏 v2 任务结果摘要失败: %v", err))
	}
	res.GeneratedAt = time.Now()
	return res, nil
}

func (dao *ReadModelDAO) ListCollectRuns(ctx context.Context, query v2models.GameV2CollectRunQuery) ([]v2models.GfgGameV2CollectRun, common.GFError) {
	if dao == nil || dao.db == nil {
		return nil, common.NewDaoError("game v2 read model database is not initialized")
	}
	var rows []v2models.GfgGameV2CollectRun
	q := dao.db.WithContext(ctx).Table(v2models.TableNameGfgGameV2CollectRuns)
	if query.TaskType != "" {
		q = q.Where("task_type = ?", query.TaskType)
	}
	if query.Status != "" {
		q = q.Where("status = ?", query.Status)
	}
	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.Offset < 0 {
		query.Offset = 0
	}
	if err := q.Order("started_at DESC, id DESC").Limit(query.Limit).Offset(query.Offset).Find(&rows).Error; err != nil {
		return nil, common.NewDaoError(fmt.Sprintf("查询游戏 v2 采集批次失败: %v", err))
	}
	return rows, nil
}

func (dao *ReadModelDAO) GetCollectRun(ctx context.Context, runID string) (*v2models.GfgGameV2CollectRun, common.GFError) {
	if dao == nil || dao.db == nil {
		return nil, common.NewDaoError("game v2 read model database is not initialized")
	}
	if strings.TrimSpace(runID) == "" {
		return nil, common.NewDaoError("run_id is required")
	}
	row, err := takeOptional[v2models.GfgGameV2CollectRun](dao.db.WithContext(ctx).Table(v2models.TableNameGfgGameV2CollectRuns).Where("id = ?", runID))
	if err != nil {
		return nil, common.NewDaoError(fmt.Sprintf("查询游戏 v2 采集批次失败: %v", err))
	}
	if row == nil {
		return nil, common.NewDaoError("collect run not found")
	}
	return row, nil
}

func (dao *ReadModelDAO) ListCollectTaskResults(ctx context.Context, query v2models.GameV2CollectTaskResultQuery) ([]v2models.GfgGameV2CollectTaskResult, common.GFError) {
	if dao == nil || dao.db == nil {
		return nil, common.NewDaoError("game v2 read model database is not initialized")
	}
	var rows []v2models.GfgGameV2CollectTaskResult
	q := dao.db.WithContext(ctx).Table(v2models.TableNameGfgGameV2CollectTaskResults)
	if query.RunID != "" {
		q = q.Where("run_id = ?", query.RunID)
	}
	if query.TaskType != "" {
		q = q.Where("task_type = ?", query.TaskType)
	}
	if query.Status != "" {
		q = q.Where("status = ?", query.Status)
	}
	if query.GameID > 0 {
		q = q.Where("game_id = ?", query.GameID)
	}
	if query.AppID > 0 {
		q = q.Where("appid = ?", query.AppID)
	}
	if query.Limit <= 0 {
		query.Limit = 50
	}
	if query.Offset < 0 {
		query.Offset = 0
	}
	if err := q.Order("started_at DESC, id DESC").Limit(query.Limit).Offset(query.Offset).Find(&rows).Error; err != nil {
		return nil, common.NewDaoError(fmt.Sprintf("查询游戏 v2 任务结果失败: %v", err))
	}
	return rows, nil
}

func (dao *ReadModelDAO) GetGameCollectStatus(ctx context.Context, gameID int64, appID int64) (v2models.GameV2CollectGameStatus, common.GFError) {
	var res v2models.GameV2CollectGameStatus
	if dao == nil || dao.db == nil {
		return res, common.NewDaoError("game v2 read model database is not initialized")
	}
	if gameID <= 0 && appID <= 0 {
		return res, common.NewDaoError("game_id or appid is required")
	}
	db := dao.db.WithContext(ctx)
	var site v2models.GameV2SiteRecord
	if err := dao.loadSiteRecord(db, DetailQuery{GameID: gameID, AppID: appID}, &site); err != nil {
		return res, common.NewDaoError(err.Error())
	}
	res.GameID = site.ID
	res.AppID = site.AppID
	res.Name = site.Name

	if details, err := takeOptional[v2models.GfgGameV2Details](db.Table(v2models.TableNameGfgGameV2Details).Where("game_id = ?", site.ID)); err != nil {
		return res, common.NewDaoError(fmt.Sprintf("查询游戏 v2 详情新鲜度失败: %v", err))
	} else if details != nil {
		updatedAt := details.UpdatedAt
		res.DetailsUpdatedAt = &updatedAt
	}

	var localized []v2models.GfgGameV2LocalizedDetails
	if err := db.Table(v2models.TableNameGfgGameV2LocalizedDetails).Where("game_id = ?", site.ID).Order("lang ASC").Find(&localized).Error; err != nil {
		return res, common.NewDaoError(fmt.Sprintf("查询游戏 v2 本地化新鲜度失败: %v", err))
	}
	res.Localized = make([]v2models.GameV2CollectLocalizedStatus, 0, len(localized))
	for _, item := range localized {
		res.Localized = append(res.Localized, v2models.GameV2CollectLocalizedStatus{
			Lang:        item.Lang,
			Name:        item.Name,
			CollectedAt: item.CollectedAt,
			UpdatedAt:   item.UpdatedAt,
		})
	}

	var prices []v2models.GfgGameV2Price
	if err := db.Table(v2models.TableNameGfgGameV2Prices).Where("game_id = ?", site.ID).Order("region ASC").Find(&prices).Error; err != nil {
		return res, common.NewDaoError(fmt.Sprintf("查询游戏 v2 价格新鲜度失败: %v", err))
	}
	res.Prices = make([]v2models.GameV2CollectRegionFreshness, 0, len(prices))
	for _, price := range prices {
		available := price.IsFree || (strings.TrimSpace(strPtrValue(price.Currency)) != "" && (price.FinalAmount > 0 || strings.TrimSpace(strPtrValue(price.FinalFormatted)) != ""))
		res.Prices = append(res.Prices, v2models.GameV2CollectRegionFreshness{
			Region:      price.Region,
			Available:   available,
			Currency:    strPtrValue(price.Currency),
			FinalAmount: price.FinalAmount,
			CollectedAt: price.CollectedAt,
			UpdatedAt:   price.UpdatedAt,
		})
	}

	if err := db.Table(v2models.TableNameGfgGameV2Media).Where("game_id = ?", site.ID).Count(&res.MediaCount).Error; err != nil {
		return res, common.NewDaoError(fmt.Sprintf("查询游戏 v2 媒体数量失败: %v", err))
	}
	if err := db.Table(v2models.TableNameGfgGameV2News).Where("game_id = ?", site.ID).Count(&res.NewsCount).Error; err != nil {
		return res, common.NewDaoError(fmt.Sprintf("查询游戏 v2 新闻数量失败: %v", err))
	}
	type latestNewsAtRow struct {
		LatestNewsAt *time.Time `gorm:"column:latest_news_at"`
	}
	var newsAt latestNewsAtRow
	if err := db.Table(v2models.TableNameGfgGameV2News).Select("MAX(published_at) AS latest_news_at").Where("game_id = ?", site.ID).Scan(&newsAt).Error; err != nil {
		return res, common.NewDaoError(fmt.Sprintf("查询游戏 v2 最新新闻时间失败: %v", err))
	}
	res.LatestNewsAt = newsAt.LatestNewsAt

	if online, err := takeOptional[v2models.GfgGameV2PlayerCount](db.Table(v2models.TableNameGfgGameV2PlayerCounts).Where("game_id = ?", site.ID).Order("collected_at DESC, id DESC")); err != nil {
		return res, common.NewDaoError(fmt.Sprintf("查询游戏 v2 在线人数新鲜度失败: %v", err))
	} else {
		res.LatestPlayerCount = online
	}

	if err := db.Raw(`
SELECT DISTINCT ON (task_type) *
FROM gfg_game_v2_collect_task_results
WHERE game_id = ?
ORDER BY task_type, started_at DESC, id DESC
`, site.ID).Scan(&res.LatestTaskResults).Error; err != nil {
		return res, common.NewDaoError(fmt.Sprintf("查询游戏 v2 游戏最新任务结果失败: %v", err))
	}
	return res, nil
}

func (dao *ReadModelDAO) loadAggregateExtras(db *gorm.DB, aggregate *v2models.GameV2Aggregate, lang string, newsLimit int) error {
	gameID := aggregate.Site.ID
	if details, err := takeOptional[v2models.GfgGameV2Details](db.Table(v2models.TableNameGfgGameV2Details).Where("game_id = ?", gameID)); err != nil {
		return fmt.Errorf("查询游戏 v2 详情失败: %v", err)
	} else {
		aggregate.Details = details
	}

	localized, err := dao.loadLocalized(db, gameID, lang)
	if err != nil {
		return err
	}
	aggregate.Localized = localized

	if err := db.Table(v2models.TableNameGfgGameV2Prices).
		Where("game_id = ?", gameID).
		Order("region ASC").
		Find(&aggregate.Prices).Error; err != nil {
		return fmt.Errorf("查询游戏 v2 价格失败: %v", err)
	}

	if err := db.Table(v2models.TableNameGfgGameV2Media).
		Where("game_id = ?", gameID).
		Order("media_type ASC, sort_order ASC, id ASC").
		Find(&aggregate.Media).Error; err != nil {
		return fmt.Errorf("查询游戏 v2 媒体失败: %v", err)
	}

	if requirements, err := takeOptional[v2models.GfgGameV2Requirements](db.Table(v2models.TableNameGfgGameV2Requirements).Where("game_id = ?", gameID)); err != nil {
		return fmt.Errorf("查询游戏 v2 配置需求失败: %v", err)
	} else {
		aggregate.Requirements = requirements
	}

	if newsLimit > 0 {
		if err := dao.loadNews(db, gameID, lang, newsLimit, &aggregate.News); err != nil {
			return err
		}
	}

	if online, err := takeOptional[v2models.GfgGameV2PlayerCount](db.Table(v2models.TableNameGfgGameV2PlayerCounts).
		Where("game_id = ? AND status = ?", gameID, "success").
		Order("collected_at DESC, id DESC")); err != nil {
		return fmt.Errorf("查询游戏 v2 在线人数失败: %v", err)
	} else {
		aggregate.OnlineCount = online
	}

	if err := dao.loadTags(db, gameID, lang, &aggregate.Tags); err != nil {
		return err
	}

	return nil
}

func (dao *ReadModelDAO) listPanelAggregatesBySQL(ctx context.Context, query v2models.GameV2PanelQuery, sql string, args ...any) ([]v2models.GameV2Aggregate, common.GFError) {
	if dao == nil || dao.db == nil {
		return nil, common.NewDaoError("game v2 read model database is not initialized")
	}
	if query.Lang == "" {
		query.Lang = "zh"
	}
	if query.Limit <= 0 {
		query.Limit = 8
	}
	gameIDs := make([]int64, 0)
	if err := dao.db.WithContext(ctx).Raw(sql, args...).Scan(&gameIDs).Error; err != nil {
		return nil, common.NewDaoError(fmt.Sprintf("查询游戏 v2 面板候选失败: %v", err))
	}
	aggregates, err := dao.loadAggregatesByGameIDs(dao.db.WithContext(ctx), gameIDs, query.Lang, 0)
	if err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	return aggregates, nil
}

func (dao *ReadModelDAO) loadSiteRecord(db *gorm.DB, query DetailQuery, dest *v2models.GameV2SiteRecord) error {
	q := db.Table(tableNameGfgGame).
		Select("id, name, name_en, info, info_en, resources, groups, links, appid, header, view_count, weight, create_time, update_time")
	if query.GameID > 0 {
		q = q.Where("id = ?", query.GameID)
	} else {
		q = q.Where("appid = ?", query.AppID)
	}
	if err := q.Take(dest).Error; err != nil {
		return fmt.Errorf("查询站内游戏主档案失败: %w", err)
	}
	return nil
}

func (dao *ReadModelDAO) loadLocalized(db *gorm.DB, gameID int64, lang string) (*v2models.GfgGameV2LocalizedDetails, error) {
	localized, err := takeOptional[v2models.GfgGameV2LocalizedDetails](db.Table(v2models.TableNameGfgGameV2LocalizedDetails).
		Where("game_id = ? AND lang = ?", gameID, lang))
	if err != nil {
		return nil, fmt.Errorf("查询游戏 v2 本地化详情失败: %w", err)
	}
	if localized != nil || lang == "zh" {
		return localized, nil
	}
	fallback, err := takeOptional[v2models.GfgGameV2LocalizedDetails](db.Table(v2models.TableNameGfgGameV2LocalizedDetails).
		Where("game_id = ? AND lang = ?", gameID, "zh"))
	if err != nil {
		return nil, fmt.Errorf("查询游戏 v2 中文回退详情失败: %w", err)
	}
	return fallback, nil
}

func (dao *ReadModelDAO) loadNews(db *gorm.DB, gameID int64, lang string, limit int, dest *[]v2models.GfgGameV2News) error {
	if err := db.Table(v2models.TableNameGfgGameV2News).
		Where("game_id = ? AND lang = ?", gameID, lang).
		Order("published_at DESC, id DESC").
		Limit(limit).
		Find(dest).Error; err != nil {
		return fmt.Errorf("查询游戏 v2 新闻失败: %w", err)
	}
	if len(*dest) > 0 || lang == "zh" {
		return nil
	}
	if err := db.Table(v2models.TableNameGfgGameV2News).
		Where("game_id = ? AND lang = ?", gameID, "zh").
		Order("published_at DESC, id DESC").
		Limit(limit).
		Find(dest).Error; err != nil {
		return fmt.Errorf("查询游戏 v2 中文回退新闻失败: %w", err)
	}
	return nil
}

func (dao *ReadModelDAO) loadTags(db *gorm.DB, gameID int64, lang string, dest *[]v2models.GameV2Tag) error {
	q := db.Table("gfg_tag_map tm")
	if lang == "en" {
		q = q.Select("t.id::varchar as id, t.name_en as name, t.info_en as desc")
	} else {
		q = q.Select("t.id::varchar as id, t.name as name, t.info as desc")
	}
	if err := q.Joins("JOIN gfg_tag t ON tm.tag_id = t.id").
		Where("tm.game_id = ?", gameID).
		Order("t.id ASC").
		Find(dest).Error; err != nil {
		return fmt.Errorf("查询游戏 v2 标签失败: %w", err)
	}
	return nil
}

func (dao *ReadModelDAO) loadAggregatesByGameIDs(db *gorm.DB, gameIDs []int64, lang string, newsLimit int) ([]v2models.GameV2Aggregate, error) {
	if len(gameIDs) == 0 {
		return []v2models.GameV2Aggregate{}, nil
	}
	var sites []v2models.GameV2SiteRecord
	if err := db.Table(tableNameGfgGame).
		Select("id, name, name_en, info, info_en, resources, groups, links, appid, header, view_count, weight, create_time, update_time").
		Where("id IN ?", gameIDs).
		Find(&sites).Error; err != nil {
		return nil, fmt.Errorf("查询游戏 v2 面板站内档案失败: %w", err)
	}

	siteMap := make(map[int64]v2models.GameV2SiteRecord, len(sites))
	for _, site := range sites {
		siteMap[site.ID] = site
	}

	aggregates := make([]v2models.GameV2Aggregate, 0, len(gameIDs))
	for _, gameID := range gameIDs {
		site, ok := siteMap[gameID]
		if !ok {
			continue
		}
		aggregate := v2models.GameV2Aggregate{Site: site}
		if err := dao.loadAggregateExtras(db, &aggregate, lang, newsLimit); err != nil {
			return nil, err
		}
		aggregates = append(aggregates, aggregate)
	}
	return aggregates, nil
}

func (dao *ReadModelDAO) queryNewsRows(db *gorm.DB, query v2models.GameV2NewsQuery, requireGame bool) ([]v2models.GameV2NewsRow, error) {
	if query.Lang == "" {
		query.Lang = "zh"
	}
	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.Offset < 0 {
		query.Offset = 0
	}
	rows := make([]v2models.GameV2NewsRow, 0)
	q := db.Table(v2models.TableNameGfgGameV2News+" n").
		Select("n.*, g.name as game_name, g.name_en as game_name_en, g.header as header_url").
		Joins("LEFT JOIN "+tableNameGfgGame+" g ON n.game_id = g.id").
		Where("n.lang = ?", query.Lang)
	if requireGame {
		if query.GameID > 0 {
			q = q.Where("n.game_id = ?", query.GameID)
		} else if query.AppID > 0 {
			q = q.Where("n.appid = ?", query.AppID)
		} else {
			return rows, fmt.Errorf("game_id or appid is required")
		}
	}
	if err := q.Order("n.published_at DESC NULLS LAST, n.collected_at DESC, n.id DESC").
		Limit(query.Limit).
		Offset(query.Offset).
		Find(&rows).Error; err != nil {
		return rows, fmt.Errorf("查询游戏 v2 新闻失败: %w", err)
	}
	if len(rows) > 0 || query.Lang == "zh" {
		return rows, nil
	}
	query.Lang = "zh"
	return dao.queryNewsRows(db, query, requireGame)
}

func listOrder(sort string) string {
	switch sort {
	case "newest":
		return "create_time DESC, id DESC"
	case "updated":
		return "update_time DESC, id DESC"
	case "weight":
		fallthrough
	default:
		return "weight ASC, id ASC"
	}
}

func normalizeDAORegion(region string) string {
	region = strings.ToUpper(strings.TrimSpace(region))
	if region == "" {
		return "CN"
	}
	return region
}

func strPtrValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func takeOptional[T any](db *gorm.DB) (*T, error) {
	var value T
	if err := db.Take(&value).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &value, nil
}
