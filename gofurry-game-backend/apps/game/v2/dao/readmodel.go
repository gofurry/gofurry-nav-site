package dao

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	"github.com/gofurry/gofurry-game-backend/common"
	cm "github.com/gofurry/gofurry-game-backend/common/models"
	database "github.com/gofurry/gofurry-game-backend/roof/db"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	q := db.Table(tableNameGfgGame + " AS g").
		Select("g.id, g.name, g.name_en, g.info, g.info_en, g.resources, g.groups, g.links, g.appid, g.header, g.view_count, g.weight, g.create_time, g.update_time").
		Order(listOrder(query.Sort)).
		Limit(query.Limit).
		Offset(query.Offset)
	if query.Sort == "release_date" {
		q = q.Joins("LEFT JOIN " + v2models.TableNameGfgGameV2Details + " d ON d.game_id = g.id")
	}
	if !query.UpdatedSince.IsZero() {
		q = q.Where(`
g.update_time >= ?
OR EXISTS (SELECT 1 FROM gfg_game_v2_details d WHERE d.game_id = g.id AND d.updated_at >= ?)
OR EXISTS (SELECT 1 FROM gfg_game_v2_localized_details ld WHERE ld.game_id = g.id AND ld.updated_at >= ?)
OR EXISTS (SELECT 1 FROM gfg_game_v2_prices p WHERE p.game_id = g.id AND p.updated_at >= ?)
OR EXISTS (SELECT 1 FROM gfg_game_v2_media m WHERE m.game_id = g.id AND m.updated_at >= ?)
OR EXISTS (SELECT 1 FROM gfg_game_v2_assets a WHERE a.game_id = g.id AND a.updated_at >= ?)
OR EXISTS (SELECT 1 FROM gfg_game_v2_requirements r WHERE r.game_id = g.id AND r.updated_at >= ?)
OR EXISTS (SELECT 1 FROM gfg_game_v2_news n WHERE n.game_id = g.id AND n.updated_at >= ?)
`, query.UpdatedSince, query.UpdatedSince, query.UpdatedSince, query.UpdatedSince, query.UpdatedSince, query.UpdatedSince, query.UpdatedSince, query.UpdatedSince)
	}
	if err := q.Find(&sites).Error; err != nil {
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

func (dao *ReadModelDAO) SearchGames(ctx context.Context, query v2models.GameV2SearchPageQuery) (cm.PageResponse, common.GFError) {
	res := cm.PageResponse{}
	if dao == nil || dao.db == nil {
		return res, common.NewDaoError("game v2 read model database is not initialized")
	}
	if query.Lang == "" {
		query.Lang = "zh"
	}
	if query.PageNum <= 0 {
		query.PageNum = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	db := dao.db.WithContext(ctx)
	q := dao.buildSearchQuery(db, query)
	if err := q.Count(&res.Total).Error; err != nil {
		return res, common.NewDaoError(fmt.Sprintf("统计游戏 v2 搜索结果失败: %v", err))
	}

	q = dao.buildSearchQuery(db, query).
		Select(searchSelectSQL(query.Lang)).
		Offset((query.PageNum - 1) * query.PageSize).
		Limit(query.PageSize)
	q = applySearchSort(q, query)

	items := []v2models.GameV2SearchPageItem{}
	if err := q.Find(&items).Error; err != nil {
		return res, common.NewDaoError(fmt.Sprintf("查询游戏 v2 搜索结果失败: %v", err))
	}
	res.Data = items
	return res, nil
}

func (dao *ReadModelDAO) ListTags(ctx context.Context, lang string) ([]v2models.GameV2TagRecord, common.GFError) {
	if dao == nil || dao.db == nil {
		return nil, common.NewDaoError("game v2 read model database is not initialized")
	}
	countSubQuery := dao.db.WithContext(ctx).
		Table("gfg_tag_map tm").
		Select("tm.tag_id, COUNT(DISTINCT tm.game_id) as game_count").
		Joins("JOIN " + v2models.TableNameGfgGameV2Details + " d ON d.game_id = tm.game_id").
		Group("tm.tag_id")

	nameField := "COALESCE(NULLIF(gfg_tag.name, ''), gfg_tag.name_en) AS name"
	if normalizeDAOLang(lang) == "en" {
		nameField = "COALESCE(NULLIF(gfg_tag.name_en, ''), gfg_tag.name) AS name"
	}

	rows := []v2models.GameV2TagRecord{}
	if err := dao.db.WithContext(ctx).Table("gfg_tag").
		Joins("LEFT JOIN (?) AS tag_count ON gfg_tag.id = tag_count.tag_id", countSubQuery).
		Select(
			"CAST(gfg_tag.id AS VARCHAR) AS id",
			nameField,
			"CAST(gfg_tag.prefix AS VARCHAR) AS prefix",
			"COALESCE(tag_count.game_count, 0) AS game_count",
		).
		Order("game_count DESC, gfg_tag.id ASC").
		Find(&rows).Error; err != nil {
		return nil, common.NewDaoError(fmt.Sprintf("查询游戏 v2 标签失败: %v", err))
	}
	return rows, nil
}

func (dao *ReadModelDAO) GetGameReviews(ctx context.Context, query v2models.GameV2ReviewQuery) (v2models.GameV2ReviewList, common.GFError) {
	var res v2models.GameV2ReviewList
	if dao == nil || dao.db == nil {
		return res, common.NewDaoError("game v2 read model database is not initialized")
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

	db := dao.db.WithContext(ctx)
	var stats struct {
		Total    int64
		AvgScore float64
	}
	if err := db.Table("gfg_game_comment").
		Select("COUNT(*) AS total, COALESCE(AVG(score), 0) AS avg_score").
		Where("game_id = ?", query.GameID).
		Take(&stats).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return res, common.NewDaoError(fmt.Sprintf("统计游戏 v2 评论失败: %v", err))
	}
	res.Total = int(stats.Total)
	res.AvgScore = stats.AvgScore
	res.PageNum = query.PageNum
	res.PageSize = query.PageSize
	res.Remarks = []v2models.GameV2ReviewItem{}
	if stats.Total == 0 {
		return res, nil
	}
	if err := db.Table("gfg_game_comment").
		Select("region, content, score, create_time, ip, name").
		Where("game_id = ?", query.GameID).
		Order("create_time DESC, id DESC").
		Offset((query.PageNum - 1) * query.PageSize).
		Limit(query.PageSize).
		Find(&res.Remarks).Error; err != nil {
		return res, common.NewDaoError(fmt.Sprintf("查询游戏 v2 评论失败: %v", err))
	}
	return res, nil
}

func (dao *ReadModelDAO) ListLatestReviews(ctx context.Context, lang string, limit int) ([]v2models.GameV2LatestReview, common.GFError) {
	if dao == nil || dao.db == nil {
		return nil, common.NewDaoError("game v2 read model database is not initialized")
	}
	if limit <= 0 {
		limit = 5
	}
	nameField := "COALESCE(NULLIF(ld.name, ''), NULLIF(g.name, ''), NULLIF(d.name, ''), g.name_en) AS game_name"
	if normalizeDAOLang(lang) == "en" {
		nameField = "COALESCE(NULLIF(ld.name, ''), NULLIF(d.name, ''), NULLIF(g.name_en, ''), g.name) AS game_name"
	}
	rows := []v2models.GameV2LatestReview{}
	if err := dao.db.WithContext(ctx).
		Table("gfg_game_comment c").
		Joins("JOIN "+tableNameGfgGame+" g ON c.game_id = g.id").
		Joins("LEFT JOIN "+v2models.TableNameGfgGameV2Details+" d ON d.game_id = g.id").
		Joins("LEFT JOIN "+v2models.TableNameGfgGameV2LocalizedDetails+" ld ON ld.game_id = g.id AND ld.lang = ?", normalizeDAOLang(lang)).
		Select(
			"c.region",
			"c.score",
			"c.content",
			"c.ip",
			"c.create_time AS time",
			"COALESCE(NULLIF(d.header_url, ''), NULLIF(g.header, '')) AS game_cover",
			nameField,
		).
		Order("c.create_time DESC, c.id DESC").
		Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, common.NewDaoError(fmt.Sprintf("查询游戏 v2 最新评论失败: %v", err))
	}
	return rows, nil
}

func (dao *ReadModelDAO) GetRandomGameID(ctx context.Context) (string, common.GFError) {
	if dao == nil || dao.db == nil {
		return "", common.NewDaoError("game v2 read model database is not initialized")
	}
	var row struct {
		ID string `gorm:"column:id"`
	}
	if err := dao.db.WithContext(ctx).
		Table(tableNameGfgGame + " g").
		Select("CAST(g.id AS VARCHAR) AS id").
		Joins("JOIN " + v2models.TableNameGfgGameV2Details + " d ON d.game_id = g.id").
		Order("RANDOM()").
		Limit(1).
		Take(&row).Error; err != nil {
		return "", common.NewDaoError(fmt.Sprintf("随机查询游戏 v2 ID 失败: %v", err))
	}
	return row.ID, nil
}

func (dao *ReadModelDAO) ListSimilarRecommendations(ctx context.Context, query v2models.GameV2SimilarRecommendationQuery) ([]v2models.GameV2RecommendationRow, common.GFError) {
	if dao == nil || dao.db == nil {
		return nil, common.NewDaoError("game v2 read model database is not initialized")
	}
	if query.GameID <= 0 {
		return nil, common.NewDaoError("game_id is required")
	}
	if query.Limit <= 0 {
		query.Limit = 8
	}
	lang := normalizeDAOLang(query.Lang)
	region := normalizeDAORegion(query.Region)
	rows := []v2models.GameV2RecommendationRow{}
	if err := dao.db.WithContext(ctx).Raw(recommendationRowsSQL(), lang, lang, region, lang, region, query.GameID, query.AlgorithmVersion, query.Limit).Scan(&rows).Error; err != nil {
		return nil, common.NewDaoError(fmt.Sprintf("查询游戏 v2 相似推荐失败: %v", err))
	}
	return rows, nil
}

func (dao *ReadModelDAO) SaveSimilarRecommendations(ctx context.Context, sourceGameID int64, rows []v2models.GfgGameV2Recommendation) common.GFError {
	if dao == nil || dao.db == nil {
		return common.NewDaoError("game v2 read model database is not initialized")
	}
	if sourceGameID <= 0 {
		return common.NewDaoError("source_game_id is required")
	}
	if err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Table(v2models.TableNameGfgGameV2Recommendations).Where("source_game_id = ?", sourceGameID).Delete(&v2models.GfgGameV2Recommendation{}).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		return tx.Table(v2models.TableNameGfgGameV2Recommendations).
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "source_game_id"}, {Name: "target_game_id"}},
				UpdateAll: true,
			}).
			Create(&rows).Error
	}); err != nil {
		return common.NewDaoError(fmt.Sprintf("保存游戏 v2 相似推荐失败: %v", err))
	}
	return nil
}

func (dao *ReadModelDAO) ListRecommendationFeatures(ctx context.Context, lang string, region string) ([]v2models.GameV2RecommendationFeature, common.GFError) {
	if dao == nil || dao.db == nil {
		return nil, common.NewDaoError("game v2 read model database is not initialized")
	}
	lang = normalizeDAOLang(lang)
	region = normalizeDAORegion(region)
	rows := []v2models.GameV2RecommendationFeature{}
	if err := dao.db.WithContext(ctx).Raw(recommendationFeaturesSQL(), lang, lang, region, lang, region).Scan(&rows).Error; err != nil {
		return nil, common.NewDaoError(fmt.Sprintf("查询游戏 v2 推荐特征失败: %v", err))
	}
	return rows, nil
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
ORDER BY random(), p.game_id ASC
LIMIT ?
`, normalizeDAORegion(query.Region), query.Limit)
}

func (dao *ReadModelDAO) ListPopularGameAggregates(ctx context.Context, query v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError) {
	return dao.listPanelAggregatesBySQL(ctx, query, `
SELECT id
FROM gfg_game
ORDER BY view_count DESC, update_time DESC, id DESC
LIMIT ?
`, query.Limit)
}

func (dao *ReadModelDAO) ListHighestPriceAggregates(ctx context.Context, query v2models.GameV2PanelQuery) ([]v2models.GameV2Aggregate, common.GFError) {
	return dao.listPanelAggregatesBySQL(ctx, query, `
SELECT p.game_id
FROM gfg_game_v2_prices p
JOIN gfg_game g ON p.game_id = g.id
WHERE p.region = ?
  AND p.is_free = false
  AND p.final_amount > 0
  AND COALESCE(p.currency, '') <> ''
  AND COALESCE(p.final_formatted, '') <> ''
ORDER BY p.final_amount DESC, p.discount_percent DESC, g.weight ASC, p.game_id ASC
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
	region := normalizeDAORegion(query.Region)
	return dao.listPanelAggregatesBySQL(ctx, query, `
SELECT game_id
FROM (
    (
        SELECT p.game_id, p.final_amount, p.discount_percent, g.weight
        FROM gfg_game_v2_prices p
        JOIN gfg_game g ON p.game_id = g.id
        WHERE p.region = ?
          AND p.is_free = false
          AND p.final_amount > 0
          AND p.final_amount <= 1000
          AND COALESCE(p.currency, '') <> ''
          AND COALESCE(p.final_formatted, '') <> ''
        ORDER BY p.final_amount DESC, p.discount_percent DESC, g.weight ASC, p.game_id ASC
        LIMIT 15
    )
    UNION ALL
    (
        SELECT p.game_id, p.final_amount, p.discount_percent, g.weight
        FROM gfg_game_v2_prices p
        JOIN gfg_game g ON p.game_id = g.id
        WHERE p.region = ?
          AND p.is_free = false
          AND p.final_amount > 0
          AND p.final_amount <= 1500
          AND COALESCE(p.currency, '') <> ''
          AND COALESCE(p.final_formatted, '') <> ''
        ORDER BY p.final_amount DESC, p.discount_percent DESC, g.weight ASC, p.game_id ASC
        LIMIT 15
    )
    UNION ALL
    (
        SELECT p.game_id, p.final_amount, p.discount_percent, g.weight
        FROM gfg_game_v2_prices p
        JOIN gfg_game g ON p.game_id = g.id
        WHERE p.region = ?
          AND p.is_free = false
          AND p.final_amount > 0
          AND p.final_amount <= 2000
          AND COALESCE(p.currency, '') <> ''
          AND COALESCE(p.final_formatted, '') <> ''
        ORDER BY p.final_amount DESC, p.discount_percent DESC, g.weight ASC, p.game_id ASC
        LIMIT 15
    )
    UNION ALL
    (
        SELECT p.game_id, p.final_amount, p.discount_percent, g.weight
        FROM gfg_game_v2_prices p
        JOIN gfg_game g ON p.game_id = g.id
        WHERE p.region = ?
          AND p.is_free = false
          AND p.final_amount > 0
          AND p.final_amount <= 2500
          AND COALESCE(p.currency, '') <> ''
          AND COALESCE(p.final_formatted, '') <> ''
        ORDER BY p.final_amount DESC, p.discount_percent DESC, g.weight ASC, p.game_id ASC
        LIMIT 15
    )
) bucketed
GROUP BY game_id
ORDER BY MAX(final_amount) DESC, MAX(discount_percent) DESC, MIN(weight) ASC, game_id ASC
`, region, region, region, region)
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

	if err := db.Table(v2models.TableNameGfgGameV2Assets).
		Where("game_id = ?", gameID).
		Order("asset_family ASC, sort_order ASC, id ASC").
		Find(&aggregate.Assets).Error; err != nil {
		return fmt.Errorf("查询游戏 v2 统一媒体资产失败: %v", err)
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

	if err := db.Table("gfg_game_comment").
		Select("COALESCE(AVG(score), 0) AS avg_score, COUNT(*) AS comment_count").
		Where("game_id = ?", gameID).
		Scan(&aggregate.ReviewStats).Error; err != nil {
		return fmt.Errorf("查询游戏 v2 评论统计失败: %v", err)
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
	requestedLang := normalizeDAOLang(lang)
	localized, err := dao.takeLocalizedDetail(db, gameID, requestedLang)
	if err != nil {
		return nil, fmt.Errorf("查询游戏 v2 本地化详情失败: %w", err)
	}

	fallbackLang := localizedFallbackLang(requestedLang)
	if fallbackLang == "" {
		return localized, nil
	}

	fallback, err := dao.takeLocalizedDetail(db, gameID, fallbackLang)
	if err != nil {
		return nil, fmt.Errorf("查询游戏 v2 回退详情失败: %w", err)
	}

	return mergeLocalizedDetails(localized, fallback), nil
}

func (dao *ReadModelDAO) takeLocalizedDetail(db *gorm.DB, gameID int64, lang string) (*v2models.GfgGameV2LocalizedDetails, error) {
	return takeOptional[v2models.GfgGameV2LocalizedDetails](db.Table(v2models.TableNameGfgGameV2LocalizedDetails).
		Where("game_id = ? AND lang = ?", gameID, lang))
}

func localizedFallbackLang(lang string) string {
	switch normalizeDAOLang(lang) {
	case "en":
		return "zh"
	default:
		return "en"
	}
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
	uniqueGameIDs := uniqueInt64s(gameIDs)

	var sites []v2models.GameV2SiteRecord
	if err := db.Table(tableNameGfgGame).
		Select("id, name, name_en, info, info_en, resources, groups, links, appid, header, view_count, weight, create_time, update_time").
		Where("id IN ?", uniqueGameIDs).
		Find(&sites).Error; err != nil {
		return nil, fmt.Errorf("查询游戏 v2 面板站内档案失败: %w", err)
	}

	aggregateMap := make(map[int64]*v2models.GameV2Aggregate, len(sites))
	for _, site := range sites {
		siteCopy := site
		aggregateMap[site.ID] = &v2models.GameV2Aggregate{Site: siteCopy}
	}

	if err := dao.loadAggregateExtrasBatch(db, aggregateMap, uniqueGameIDs, lang); err != nil {
		return nil, err
	}

	if newsLimit > 0 {
		for _, gameID := range uniqueGameIDs {
			aggregate := aggregateMap[gameID]
			if aggregate == nil {
				continue
			}
			if err := dao.loadNews(db, gameID, lang, newsLimit, &aggregate.News); err != nil {
				return nil, err
			}
		}
	}

	aggregates := make([]v2models.GameV2Aggregate, 0, len(gameIDs))
	for _, gameID := range gameIDs {
		aggregate := aggregateMap[gameID]
		if aggregate == nil {
			continue
		}
		aggregates = append(aggregates, *aggregate)
	}
	return aggregates, nil
}

func (dao *ReadModelDAO) loadAggregateExtrasBatch(db *gorm.DB, aggregateMap map[int64]*v2models.GameV2Aggregate, gameIDs []int64, lang string) error {
	if len(gameIDs) == 0 || len(aggregateMap) == 0 {
		return nil
	}

	if err := dao.loadDetailsBatch(db, aggregateMap, gameIDs); err != nil {
		return err
	}
	if err := dao.loadLocalizedBatch(db, aggregateMap, gameIDs, lang); err != nil {
		return err
	}
	if err := dao.loadPricesBatch(db, aggregateMap, gameIDs); err != nil {
		return err
	}
	if err := dao.loadMediaBatch(db, aggregateMap, gameIDs); err != nil {
		return err
	}
	if err := dao.loadAssetsBatch(db, aggregateMap, gameIDs); err != nil {
		return err
	}
	if err := dao.loadRequirementsBatch(db, aggregateMap, gameIDs); err != nil {
		return err
	}
	if err := dao.loadOnlineCountBatch(db, aggregateMap, gameIDs); err != nil {
		return err
	}
	if err := dao.loadReviewStatsBatch(db, aggregateMap, gameIDs); err != nil {
		return err
	}
	if err := dao.loadTagsBatch(db, aggregateMap, gameIDs, lang); err != nil {
		return err
	}

	return nil
}

func (dao *ReadModelDAO) loadDetailsBatch(db *gorm.DB, aggregateMap map[int64]*v2models.GameV2Aggregate, gameIDs []int64) error {
	rows := make([]v2models.GfgGameV2Details, 0, len(gameIDs))
	if err := db.Table(v2models.TableNameGfgGameV2Details).
		Where("game_id IN ?", gameIDs).
		Find(&rows).Error; err != nil {
		return fmt.Errorf("批量查询游戏 v2 详情失败: %w", err)
	}

	for i := range rows {
		row := rows[i]
		aggregate := aggregateMap[row.GameID]
		if aggregate == nil {
			continue
		}
		aggregate.Details = &rows[i]
	}
	return nil
}

func (dao *ReadModelDAO) loadLocalizedBatch(db *gorm.DB, aggregateMap map[int64]*v2models.GameV2Aggregate, gameIDs []int64, lang string) error {
	requestedLang := normalizeDAOLang(lang)
	fallbackLang := localizedFallbackLang(requestedLang)
	langs := []string{requestedLang}
	if fallbackLang != "" && fallbackLang != requestedLang {
		langs = append(langs, fallbackLang)
	}

	rows := make([]v2models.GfgGameV2LocalizedDetails, 0, len(gameIDs))
	if err := db.Table(v2models.TableNameGfgGameV2LocalizedDetails).
		Where("game_id IN ? AND lang IN ?", gameIDs, langs).
		Find(&rows).Error; err != nil {
		return fmt.Errorf("批量查询游戏 v2 本地化详情失败: %w", err)
	}

	primaryRows := make(map[int64]*v2models.GfgGameV2LocalizedDetails, len(rows))
	fallbackRows := make(map[int64]*v2models.GfgGameV2LocalizedDetails, len(rows))
	for i := range rows {
		row := &rows[i]
		switch normalizeDAOLang(row.Lang) {
		case requestedLang:
			primaryRows[row.GameID] = row
		case fallbackLang:
			fallbackRows[row.GameID] = row
		}
	}

	for _, gameID := range gameIDs {
		aggregate := aggregateMap[gameID]
		if aggregate == nil {
			continue
		}
		aggregate.Localized = mergeLocalizedDetails(primaryRows[gameID], fallbackRows[gameID])
	}
	return nil
}

func (dao *ReadModelDAO) loadPricesBatch(db *gorm.DB, aggregateMap map[int64]*v2models.GameV2Aggregate, gameIDs []int64) error {
	rows := make([]v2models.GfgGameV2Price, 0)
	if err := db.Table(v2models.TableNameGfgGameV2Prices).
		Where("game_id IN ?", gameIDs).
		Order("game_id ASC, region ASC").
		Find(&rows).Error; err != nil {
		return fmt.Errorf("批量查询游戏 v2 价格失败: %w", err)
	}

	for _, row := range rows {
		aggregate := aggregateMap[row.GameID]
		if aggregate == nil {
			continue
		}
		aggregate.Prices = append(aggregate.Prices, row)
	}
	return nil
}

func (dao *ReadModelDAO) loadMediaBatch(db *gorm.DB, aggregateMap map[int64]*v2models.GameV2Aggregate, gameIDs []int64) error {
	rows := make([]v2models.GfgGameV2Media, 0)
	if err := db.Table(v2models.TableNameGfgGameV2Media).
		Where("game_id IN ?", gameIDs).
		Order("game_id ASC, media_type ASC, sort_order ASC, id ASC").
		Find(&rows).Error; err != nil {
		return fmt.Errorf("批量查询游戏 v2 媒体失败: %w", err)
	}

	for _, row := range rows {
		aggregate := aggregateMap[row.GameID]
		if aggregate == nil {
			continue
		}
		aggregate.Media = append(aggregate.Media, row)
	}
	return nil
}

func (dao *ReadModelDAO) loadAssetsBatch(db *gorm.DB, aggregateMap map[int64]*v2models.GameV2Aggregate, gameIDs []int64) error {
	rows := make([]v2models.GfgGameV2Asset, 0)
	if err := db.Table(v2models.TableNameGfgGameV2Assets).
		Where("game_id IN ?", gameIDs).
		Order("game_id ASC, asset_family ASC, sort_order ASC, id ASC").
		Find(&rows).Error; err != nil {
		return fmt.Errorf("批量查询游戏 v2 统一媒体资产失败: %w", err)
	}

	for _, row := range rows {
		aggregate := aggregateMap[row.GameID]
		if aggregate == nil {
			continue
		}
		aggregate.Assets = append(aggregate.Assets, row)
	}
	return nil
}

func (dao *ReadModelDAO) loadRequirementsBatch(db *gorm.DB, aggregateMap map[int64]*v2models.GameV2Aggregate, gameIDs []int64) error {
	rows := make([]v2models.GfgGameV2Requirements, 0, len(gameIDs))
	if err := db.Table(v2models.TableNameGfgGameV2Requirements).
		Where("game_id IN ?", gameIDs).
		Find(&rows).Error; err != nil {
		return fmt.Errorf("批量查询游戏 v2 配置需求失败: %w", err)
	}

	for i := range rows {
		row := rows[i]
		aggregate := aggregateMap[row.GameID]
		if aggregate == nil {
			continue
		}
		aggregate.Requirements = &rows[i]
	}
	return nil
}

func (dao *ReadModelDAO) loadOnlineCountBatch(db *gorm.DB, aggregateMap map[int64]*v2models.GameV2Aggregate, gameIDs []int64) error {
	rows := make([]v2models.GfgGameV2PlayerCount, 0, len(gameIDs))
	if err := db.Raw(`
SELECT DISTINCT ON (game_id) *
FROM `+v2models.TableNameGfgGameV2PlayerCounts+`
WHERE game_id IN ? AND status = 'success'
ORDER BY game_id, collected_at DESC, id DESC
`, gameIDs).Scan(&rows).Error; err != nil {
		return fmt.Errorf("批量查询游戏 v2 在线人数失败: %w", err)
	}

	for i := range rows {
		row := rows[i]
		aggregate := aggregateMap[row.GameID]
		if aggregate == nil {
			continue
		}
		aggregate.OnlineCount = &rows[i]
	}
	return nil
}

func (dao *ReadModelDAO) loadReviewStatsBatch(db *gorm.DB, aggregateMap map[int64]*v2models.GameV2Aggregate, gameIDs []int64) error {
	type reviewStatsRow struct {
		GameID       int64   `gorm:"column:game_id"`
		AvgScore     float64 `gorm:"column:avg_score"`
		CommentCount int64   `gorm:"column:comment_count"`
	}

	rows := make([]reviewStatsRow, 0, len(gameIDs))
	if err := db.Table("gfg_game_comment").
		Select("game_id, COALESCE(AVG(score), 0) AS avg_score, COUNT(*) AS comment_count").
		Where("game_id IN ?", gameIDs).
		Group("game_id").
		Find(&rows).Error; err != nil {
		return fmt.Errorf("批量查询游戏 v2 评论统计失败: %w", err)
	}

	for _, row := range rows {
		aggregate := aggregateMap[row.GameID]
		if aggregate == nil {
			continue
		}
		aggregate.ReviewStats = v2models.GameV2ReviewStats{
			AvgScore:     row.AvgScore,
			CommentCount: row.CommentCount,
		}
	}
	return nil
}

func (dao *ReadModelDAO) loadTagsBatch(db *gorm.DB, aggregateMap map[int64]*v2models.GameV2Aggregate, gameIDs []int64, lang string) error {
	type tagRow struct {
		GameID int64  `gorm:"column:game_id"`
		ID     string `gorm:"column:id"`
		Name   string `gorm:"column:name"`
		Desc   string `gorm:"column:desc"`
	}

	selectSQL := "tm.game_id, t.id::varchar as id, t.name as name, t.info as desc"
	if normalizeDAOLang(lang) == "en" {
		selectSQL = "tm.game_id, t.id::varchar as id, t.name_en as name, t.info_en as desc"
	}

	rows := make([]tagRow, 0)
	if err := db.Table("gfg_tag_map tm").
		Select(selectSQL).
		Joins("JOIN gfg_tag t ON tm.tag_id = t.id").
		Where("tm.game_id IN ?", gameIDs).
		Order("tm.game_id ASC, t.id ASC").
		Find(&rows).Error; err != nil {
		return fmt.Errorf("批量查询游戏 v2 标签失败: %w", err)
	}

	for _, row := range rows {
		aggregate := aggregateMap[row.GameID]
		if aggregate == nil {
			continue
		}
		aggregate.Tags = append(aggregate.Tags, v2models.GameV2Tag{
			ID:   row.ID,
			Name: row.Name,
			Desc: row.Desc,
		})
	}
	return nil
}

func uniqueInt64s(values []int64) []int64 {
	if len(values) == 0 {
		return nil
	}

	res := make([]int64, 0, len(values))
	seen := make(map[int64]struct{}, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		res = append(res, value)
	}
	return res
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
	if !query.UpdatedSince.IsZero() {
		q = q.Where("n.updated_at >= ?", query.UpdatedSince)
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

func (dao *ReadModelDAO) buildSearchQuery(db *gorm.DB, query v2models.GameV2SearchPageQuery) *gorm.DB {
	q := db.Table(tableNameGfgGame+" g").
		Joins("JOIN "+v2models.TableNameGfgGameV2Details+" d ON d.game_id = g.id").
		Joins("LEFT JOIN "+v2models.TableNameGfgGameV2LocalizedDetails+" ld ON ld.game_id = g.id AND ld.lang = ?", normalizeDAOLang(query.Lang)).
		Joins(`LEFT JOIN (
			SELECT DISTINCT ON (game_id) game_id, url
			FROM ` + v2models.TableNameGfgGameV2Assets + `
			WHERE exists IS DISTINCT FROM false
			  AND source = 'store_browse'
			  AND asset_type IN ('header_2x', 'header')
			ORDER BY game_id,
				CASE asset_type
					WHEN 'header_2x' THEN 0
					WHEN 'header' THEN 1
					ELSE 2
				END,
				sort_order ASC,
				id ASC
		) asset_media ON asset_media.game_id = g.id`).
		Joins(`LEFT JOIN (
			SELECT game_id, COUNT(*) AS remark_count, AVG(score) AS avg_score
			FROM gfg_game_comment
			GROUP BY game_id
		) comment_stats ON comment_stats.game_id = g.id`).
		Joins("LEFT JOIN gfg_tag primary_tag ON g.primary_tag = primary_tag.id").
		Joins("LEFT JOIN gfg_tag secondary_tag ON g.secondary_tag = secondary_tag.id")

	if query.Content != "" {
		keyword := "%" + strings.ReplaceAll(query.Content, " ", "%") + "%"
		q = q.Where(`
			(
				g.name ILIKE ?
				OR g.name_en ILIKE ?
				OR g.info ILIKE ?
				OR g.info_en ILIKE ?
				OR d.name ILIKE ?
				OR d.developers::TEXT ILIKE ?
				OR d.publishers::TEXT ILIKE ?
				OR ld.name ILIKE ?
				OR ld.short_description ILIKE ?
				OR EXISTS (
					SELECT 1
					FROM gfg_tag_map tm
					JOIN gfg_tag t ON t.id = tm.tag_id
					WHERE tm.game_id = g.id
					  AND (t.name ILIKE ? OR t.name_en ILIKE ?)
				)
			)
		`, keyword, keyword, keyword, keyword, keyword, keyword, keyword, keyword, keyword, keyword, keyword)
	}

	if !query.UpdateStartTime.IsZero() && !query.UpdateEndTime.IsZero() {
		q = q.Where(searchUpdatedAtExpr()+" BETWEEN ? AND ?", query.UpdateStartTime, query.UpdateEndTime)
	}

	if !query.PubStartTime.IsZero() && !query.PubEndTime.IsZero() {
		q = q.Where(searchReleaseDateExpr()+" BETWEEN ? AND ?", query.PubStartTime, query.PubEndTime)
	}

	if len(query.TagList) > 0 {
		tagSubQuery := db.Table("gfg_tag_map").
			Select("game_id").
			Where("tag_id IN ?", query.TagList).
			Group("game_id").
			Having("COUNT(DISTINCT tag_id) = ?", len(query.TagList))
		q = q.Where("g.id IN (?)", tagSubQuery)
	}
	return q
}

func searchSelectSQL(lang string) string {
	nameExpr := "COALESCE(NULLIF(ld.name, ''), NULLIF(g.name, ''), NULLIF(d.name, ''), g.name_en)"
	infoExpr := "COALESCE(NULLIF(ld.short_description, ''), NULLIF(g.info, ''), g.info_en)"
	primaryTagExpr := "COALESCE(NULLIF(primary_tag.name, ''), primary_tag.name_en, '')"
	secondaryTagExpr := "COALESCE(NULLIF(secondary_tag.name, ''), secondary_tag.name_en, '')"
	if normalizeDAOLang(lang) == "en" {
		nameExpr = "COALESCE(NULLIF(ld.name, ''), NULLIF(d.name, ''), NULLIF(g.name_en, ''), g.name)"
		infoExpr = "COALESCE(NULLIF(ld.short_description, ''), NULLIF(g.info_en, ''), g.info)"
		primaryTagExpr = "COALESCE(NULLIF(primary_tag.name_en, ''), primary_tag.name, '')"
		secondaryTagExpr = "COALESCE(NULLIF(secondary_tag.name_en, ''), secondary_tag.name, '')"
	}
	return fmt.Sprintf(`
		CAST(g.id AS VARCHAR) AS id,
		%s AS name,
		%s AS info,
		COALESCE(NULLIF(asset_media.url, ''), NULLIF(d.header_url, ''), NULLIF(g.header, ''), '') AS cover,
		g.appid AS appid,
		%s AS update_time,
		COALESCE(NULLIF(d.release_date_text, ''), NULLIF(g.release_date, '')) AS release_date,
		COALESCE(comment_stats.remark_count, 0) AS remark_count,
		COALESCE(comment_stats.avg_score, 0) AS avg_score,
		%s AS primary_tag,
		%s AS secondary_tag
	`, nameExpr, infoExpr, searchUpdatedAtExpr(), primaryTagExpr, secondaryTagExpr)
}

func applySearchSort(db *gorm.DB, query v2models.GameV2SearchPageQuery) *gorm.DB {
	orderClauses := []string{}
	if query.TimeOrder {
		orderClauses = append(orderClauses, "g.create_time DESC")
	}
	if query.RemarkOrder {
		orderClauses = append(orderClauses, "comment_stats.remark_count DESC NULLS LAST")
	}
	if query.ScoreOrder {
		orderClauses = append(orderClauses, "comment_stats.avg_score DESC NULLS LAST")
	}
	if query.TimeOrder {
		orderClauses = append(orderClauses, "g.weight DESC", "g.id ASC")
	} else {
		orderClauses = append(orderClauses, "g.weight ASC", "g.id ASC")
	}
	return db.Order(strings.Join(orderClauses, ", "))
}

func searchUpdatedAtExpr() string {
	return "GREATEST(g.update_time, COALESCE(d.updated_at, g.update_time), COALESCE(ld.updated_at, g.update_time))"
}

func searchReleaseDateExpr() string {
	return `
		COALESCE(
			CASE
				WHEN regexp_replace(COALESCE(g.release_date, ''), '[[:space:]]+', '', 'g') ~ '^[0-9]{4}[.-][0-9]{2}[.-][0-9]{2}$'
				THEN to_date(REPLACE(regexp_replace(COALESCE(g.release_date, ''), '[[:space:]]+', '', 'g'), '.', '-'), 'YYYY-MM-DD')
				ELSE NULL
			END,
			CASE
				WHEN regexp_replace(COALESCE(d.release_date_text, ''), '[[:space:]]+', '', 'g') ~ '^[0-9]{4}[.-][0-9]{2}[.-][0-9]{2}$'
				THEN to_date(REPLACE(regexp_replace(COALESCE(d.release_date_text, ''), '[[:space:]]+', '', 'g'), '.', '-'), 'YYYY-MM-DD')
				ELSE NULL
			END
		)
	`
}

func listOrder(sort string) string {
	switch sort {
	case "release_date":
		return releaseDateOrderExpr() + " DESC NULLS LAST, g.id DESC"
	case "newest":
		return "g.create_time DESC, g.id DESC"
	case "updated":
		return "g.update_time DESC, g.id DESC"
	case "weight":
		fallthrough
	default:
		return "g.weight ASC, g.id ASC"
	}
}

func releaseDateOrderExpr() string {
	return `
		COALESCE(
			CASE
				WHEN regexp_replace(COALESCE(g.release_date, ''), '[[:space:]]+', '', 'g') ~ '^[0-9]{4}[.-][0-9]{2}[.-][0-9]{2}$'
				THEN to_date(REPLACE(regexp_replace(COALESCE(g.release_date, ''), '[[:space:]]+', '', 'g'), '.', '-'), 'YYYY-MM-DD')
				ELSE NULL
			END,
			CASE
				WHEN regexp_replace(COALESCE(d.release_date_text, ''), '[[:space:]]+', '', 'g') ~ '^[0-9]{4}[.-][0-9]{2}[.-][0-9]{2}$'
				THEN to_date(REPLACE(regexp_replace(COALESCE(d.release_date_text, ''), '[[:space:]]+', '', 'g'), '.', '-'), 'YYYY-MM-DD')
				ELSE NULL
			END
		)
	`
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

func recommendationRowsSQL() string {
	return `
WITH latest_player_count AS (
    SELECT DISTINCT ON (game_id) game_id, count, status, collected_at
    FROM gfg_game_v2_player_counts
    WHERE status = 'success'
    ORDER BY game_id, collected_at DESC, id DESC
),
target_tags AS (
    SELECT tm.game_id,
           COALESCE(
               jsonb_agg(
                   jsonb_build_object(
                       'id', t.id::text,
                       'name', CASE WHEN ? = 'en' THEN COALESCE(NULLIF(t.name_en, ''), t.name) ELSE COALESCE(NULLIF(t.name, ''), t.name_en) END,
                       'desc', CASE WHEN ? = 'en' THEN COALESCE(NULLIF(t.info_en, ''), t.info) ELSE COALESCE(NULLIF(t.info, ''), t.info_en) END,
                       'prefix', t.prefix::text
                   )
                   ORDER BY t.id ASC
               ),
               '[]'::jsonb
           ) AS tags
    FROM gfg_tag_map tm
    JOIN gfg_tag t ON t.id = tm.tag_id
    GROUP BY tm.game_id
),
header_media AS (
    SELECT DISTINCT ON (game_id) game_id, url
    FROM (
        SELECT game_id, url,
               CASE asset_type
                   WHEN 'header_2x' THEN 0
                   WHEN 'header' THEN 1
                   ELSE 2
               END AS priority,
               sort_order,
               id
        FROM gfg_game_v2_assets
        WHERE exists IS DISTINCT FROM false
          AND source = 'store_browse'
          AND asset_type IN ('header_2x', 'header')
    ) candidates
    WHERE COALESCE(url, '') <> ''
    ORDER BY game_id, priority ASC, sort_order ASC, id ASC
),
capsule_media AS (
    SELECT DISTINCT ON (game_id) game_id, url
    FROM (
        SELECT game_id, url,
               CASE asset_type
                   WHEN 'capsule_main_2x' THEN 0
                   WHEN 'capsule_main' THEN 1
                   WHEN 'hero_capsule_2x' THEN 2
                   WHEN 'hero_capsule' THEN 3
                   ELSE 4
               END AS priority,
               sort_order,
               id
        FROM gfg_game_v2_assets
        WHERE exists IS DISTINCT FROM false
          AND source = 'store_browse'
          AND asset_type IN ('capsule_main_2x', 'capsule_main', 'hero_capsule_2x', 'hero_capsule')
    ) candidates
    WHERE COALESCE(url, '') <> ''
    ORDER BY game_id, priority ASC, sort_order ASC, id ASC
),
library_cover_media AS (
    SELECT DISTINCT ON (game_id) game_id, url
    FROM gfg_game_v2_assets
    WHERE exists IS DISTINCT FROM false
      AND source = 'store_browse'
      AND asset_type = 'library_capsule'
      AND COALESCE(url, '') <> ''
    ORDER BY game_id, sort_order ASC, id ASC
),
library_cover_2x_media AS (
    SELECT DISTINCT ON (game_id) game_id, url
    FROM gfg_game_v2_assets
    WHERE exists IS DISTINCT FROM false
      AND source = 'store_browse'
      AND asset_type = 'library_capsule_2x'
      AND COALESCE(url, '') <> ''
    ORDER BY game_id, sort_order ASC, id ASC
)
SELECT
    r.source_game_id,
    r.target_game_id,
    r.score,
    r.display_score,
    r.rank,
    r.reason_json::text AS reason_json,
    r.algorithm_version,
    r.computed_at,
    g.appid,
    COALESCE(NULLIF(ld.name, ''), NULLIF(d.name, ''), NULLIF(g.name, ''), g.name_en) AS name,
    COALESCE(NULLIF(ld.short_description, ''), NULLIF(g.info, ''), g.info_en) AS summary,
    COALESCE(NULLIF(header_media.url, ''), NULLIF(d.header_url, ''), NULLIF(g.header, ''), '') AS header_url,
    COALESCE(NULLIF(capsule_media.url, ''), '') AS capsule_url,
    COALESCE(NULLIF(library_cover_media.url, ''), '') AS library_cover_url,
    COALESCE(NULLIF(library_cover_2x_media.url, ''), '') AS library_cover_2x_url,
    COALESCE(target_tags.tags, '[]'::jsonb)::text AS tags,
    COALESCE(p.region, ?) AS price_region,
    CASE
        WHEN p.game_id IS NULL THEN false
        WHEN p.is_free THEN true
        WHEN COALESCE(p.currency, '') <> '' AND (p.final_amount > 0 OR COALESCE(p.final_formatted, '') <> '') THEN true
        ELSE false
    END AS price_available,
    COALESCE(p.is_free, false) AS is_free,
    COALESCE(p.currency, '') AS currency,
    COALESCE(p.initial_amount, 0) AS initial_amount,
    COALESCE(p.final_amount, 0) AS final_amount,
    COALESCE(p.discount_percent, 0) AS discount_percent,
    COALESCE(p.initial_formatted, '') AS initial_formatted,
    COALESCE(p.final_formatted, '') AS final_formatted,
    p.updated_at AS price_updated_at,
    COALESCE(latest_player_count.count, 0) AS online_count,
    COALESCE(latest_player_count.status, 'unknown') AS online_status,
    latest_player_count.collected_at AS online_collected_at
FROM gfg_game_v2_recommendations r
JOIN gfg_game g ON g.id = r.target_game_id
JOIN gfg_game_v2_details d ON d.game_id = g.id
LEFT JOIN gfg_game_v2_localized_details ld ON ld.game_id = g.id AND ld.lang = ?
LEFT JOIN gfg_game_v2_prices p ON p.game_id = g.id AND p.region = ?
LEFT JOIN latest_player_count ON latest_player_count.game_id = g.id
LEFT JOIN target_tags ON target_tags.game_id = g.id
LEFT JOIN header_media ON header_media.game_id = g.id
LEFT JOIN capsule_media ON capsule_media.game_id = g.id
LEFT JOIN library_cover_media ON library_cover_media.game_id = g.id
LEFT JOIN library_cover_2x_media ON library_cover_2x_media.game_id = g.id
WHERE r.source_game_id = ?
  AND r.algorithm_version = ?
ORDER BY r.rank ASC, r.score DESC, r.target_game_id ASC
LIMIT ?
`
}

func recommendationFeaturesSQL() string {
	return `
WITH latest_player_count AS (
    SELECT DISTINCT ON (game_id) game_id, count, status, collected_at
    FROM gfg_game_v2_player_counts
    WHERE status = 'success'
    ORDER BY game_id, collected_at DESC, id DESC
),
target_tags AS (
    SELECT tm.game_id,
           COALESCE(
               jsonb_agg(
                   jsonb_build_object(
                       'id', t.id::text,
                       'name', CASE WHEN ? = 'en' THEN COALESCE(NULLIF(t.name_en, ''), t.name) ELSE COALESCE(NULLIF(t.name, ''), t.name_en) END,
                       'desc', CASE WHEN ? = 'en' THEN COALESCE(NULLIF(t.info_en, ''), t.info) ELSE COALESCE(NULLIF(t.info, ''), t.info_en) END,
                       'prefix', t.prefix::text
                   )
                   ORDER BY t.id ASC
               ),
               '[]'::jsonb
           ) AS tags
    FROM gfg_tag_map tm
    JOIN gfg_tag t ON t.id = tm.tag_id
    GROUP BY tm.game_id
),
header_media AS (
    SELECT DISTINCT ON (game_id) game_id, url
    FROM (
        SELECT game_id, url,
               CASE asset_type
                   WHEN 'header_2x' THEN 0
                   WHEN 'header' THEN 1
                   ELSE 2
               END AS priority,
               sort_order,
               id
        FROM gfg_game_v2_assets
        WHERE exists IS DISTINCT FROM false
          AND source = 'store_browse'
          AND asset_type IN ('header_2x', 'header')
    ) candidates
    WHERE COALESCE(url, '') <> ''
    ORDER BY game_id, priority ASC, sort_order ASC, id ASC
),
capsule_media AS (
    SELECT DISTINCT ON (game_id) game_id, url
    FROM (
        SELECT game_id, url,
               CASE asset_type
                   WHEN 'capsule_main_2x' THEN 0
                   WHEN 'capsule_main' THEN 1
                   WHEN 'hero_capsule_2x' THEN 2
                   WHEN 'hero_capsule' THEN 3
                   ELSE 4
               END AS priority,
               sort_order,
               id
        FROM gfg_game_v2_assets
        WHERE exists IS DISTINCT FROM false
          AND source = 'store_browse'
          AND asset_type IN ('capsule_main_2x', 'capsule_main', 'hero_capsule_2x', 'hero_capsule')
    ) candidates
    WHERE COALESCE(url, '') <> ''
    ORDER BY game_id, priority ASC, sort_order ASC, id ASC
),
library_cover_media AS (
    SELECT DISTINCT ON (game_id) game_id, url
    FROM gfg_game_v2_assets
    WHERE exists IS DISTINCT FROM false
      AND source = 'store_browse'
      AND asset_type = 'library_capsule'
      AND COALESCE(url, '') <> ''
    ORDER BY game_id, sort_order ASC, id ASC
),
library_cover_2x_media AS (
    SELECT DISTINCT ON (game_id) game_id, url
    FROM gfg_game_v2_assets
    WHERE exists IS DISTINCT FROM false
      AND source = 'store_browse'
      AND asset_type = 'library_capsule_2x'
      AND COALESCE(url, '') <> ''
    ORDER BY game_id, sort_order ASC, id ASC
)
SELECT
    g.id AS game_id,
    g.appid,
    COALESCE(NULLIF(ld.name, ''), NULLIF(d.name, ''), NULLIF(g.name, ''), g.name_en) AS name,
    COALESCE(NULLIF(ld.short_description, ''), NULLIF(g.info, ''), g.info_en) AS summary,
    COALESCE(NULLIF(header_media.url, ''), NULLIF(d.header_url, ''), NULLIF(g.header, ''), '') AS header_url,
    COALESCE(NULLIF(capsule_media.url, ''), '') AS capsule_url,
    COALESCE(NULLIF(library_cover_media.url, ''), '') AS library_cover_url,
    COALESCE(NULLIF(library_cover_2x_media.url, ''), '') AS library_cover_2x_url,
    d.developers::text AS developers,
    d.publishers::text AS publishers,
    d.platforms::text AS platforms,
    COALESCE(g.primary_tag, 0) AS primary_tag_id,
    COALESCE(g.secondary_tag, 0) AS secondary_tag_id,
    COALESCE(target_tags.tags, '[]'::jsonb)::text AS tags,
    COALESCE(p.region, ?) AS price_region,
    CASE
        WHEN p.game_id IS NULL THEN false
        WHEN p.is_free THEN true
        WHEN COALESCE(p.currency, '') <> '' AND (p.final_amount > 0 OR COALESCE(p.final_formatted, '') <> '') THEN true
        ELSE false
    END AS price_available,
    COALESCE(p.is_free, false) AS is_free,
    COALESCE(p.currency, '') AS currency,
    COALESCE(p.initial_amount, 0) AS initial_amount,
    COALESCE(p.final_amount, 0) AS final_amount,
    COALESCE(p.discount_percent, 0) AS discount_percent,
    COALESCE(p.initial_formatted, '') AS initial_formatted,
    COALESCE(p.final_formatted, '') AS final_formatted,
    p.updated_at AS price_updated_at,
    COALESCE(latest_player_count.count, 0) AS online_count,
    COALESCE(latest_player_count.status, 'unknown') AS online_status,
    latest_player_count.collected_at AS online_collected_at,
    GREATEST(g.update_time, COALESCE(d.updated_at, g.update_time), COALESCE(ld.updated_at, g.update_time)) AS updated_at
FROM gfg_game g
JOIN gfg_game_v2_details d ON d.game_id = g.id
LEFT JOIN gfg_game_v2_localized_details ld ON ld.game_id = g.id AND ld.lang = ?
LEFT JOIN gfg_game_v2_prices p ON p.game_id = g.id AND p.region = ?
LEFT JOIN latest_player_count ON latest_player_count.game_id = g.id
LEFT JOIN target_tags ON target_tags.game_id = g.id
LEFT JOIN header_media ON header_media.game_id = g.id
LEFT JOIN capsule_media ON capsule_media.game_id = g.id
LEFT JOIN library_cover_media ON library_cover_media.game_id = g.id
LEFT JOIN library_cover_2x_media ON library_cover_2x_media.game_id = g.id
ORDER BY g.weight ASC, g.id ASC
`
}
