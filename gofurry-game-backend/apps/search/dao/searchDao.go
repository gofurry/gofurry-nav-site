package dao

import (
	"strings"

	gm "github.com/gofurry/gofurry-game-backend/apps/game/models"
	rm "github.com/gofurry/gofurry-game-backend/apps/review/models"
	"github.com/gofurry/gofurry-game-backend/apps/search/models"
	"github.com/gofurry/gofurry-game-backend/common"
	"github.com/gofurry/gofurry-game-backend/common/abstract"
	cm "github.com/gofurry/gofurry-game-backend/common/models"
	"gorm.io/gorm"
)

var newSearchDao = new(searchDao)

func init() {
	newSearchDao.Init()
}

type searchDao struct{ abstract.Dao }

func GetSearchDao() *searchDao { return newSearchDao }

func (dao searchDao) GetGameListByText(text string, limit int) (res []gm.GfgGame, err common.GFError) {
	db := dao.Gm.Table(gm.TableNameGfgGame)
	if db.Error != nil {
		return res, common.NewDaoError(db.Error.Error())
	}

	// 不区分大小写模糊匹配
	searchText := "%" + strings.TrimSpace(text) + "%"
	db.Where("name ILIKE ? OR name_en ILIKE ? OR info ILIKE ? OR info_en ILIKE ?",
		searchText, searchText, searchText, searchText)
	db.Order("weight ASC, update_time DESC").Limit(limit)

	if errDb := db.Find(&res).Error; errDb != nil {
		return res, common.NewDaoError(errDb.Error())
	}
	return res, nil
}

// Paginate 游戏搜索分页查询
func (dao searchDao) Paginate(req models.SearchPageQueryRequest) (cm.PageResponse, common.GFError) {
	pageData := cm.PageResponse{}

	// 构建评论统计子查询
	commentSubQuery := dao.Gm.Table(rm.TableNameGfgGameComment).
		Select(`
			game_id, 
			COUNT(*) AS remark_count, 
        	AVG(score) AS avg_score
		`).
		Group("game_id")

	// 定义标签名字段
	tagNameField := "name"
	if req.Lang == "en" {
		tagNameField = "name_en"
	}

	// 主查询
	mainDB := dao.Gm.Table(gm.TableNameGfgGame).
		// 关联评论统计
		Joins("LEFT JOIN (?) AS comment_stats ON gfg_game.id = comment_stats.game_id", commentSubQuery).
		// 关联主标签表
		Joins("LEFT JOIN gfg_tag AS primary_tag ON gfg_game.primary_tag = primary_tag.id").
		// 关联次标签表
		Joins("LEFT JOIN gfg_tag AS secondary_tag ON gfg_game.secondary_tag = secondary_tag.id").
		Select(`
			CAST(gfg_game.id AS VARCHAR) AS id,
			` + getGameNameField(req.Lang) + ` AS name,
			` + getGameInfoField(req.Lang) + ` AS info,
			gfg_game.header AS cover,
			gfg_game.update_time,
			gfg_game.appid,
			to_char(to_date(gfg_game.release_date, 'YYYY.MM.DD'), 'YYYY-MM-DD') AS release_date,
			CAST(gfg_game.appid AS VARCHAR) AS appid,
			COALESCE(comment_stats.remark_count, 0) AS remark_count,
			COALESCE(comment_stats.avg_score, 0) AS avg_score,
			COALESCE(primary_tag.` + tagNameField + `, '') AS primary_tag,
			COALESCE(secondary_tag.` + tagNameField + `, '') AS secondary_tag
		`)

	// 构建查询条件
	buildSearchPageCondition(mainDB, &req, dao.Gm)

	// 统计总数
	var total int64
	if err := mainDB.Count(&total).Error; err != nil {
		return pageData, common.NewDaoError("统计游戏总数失败: " + err.Error())
	}

	// 分页 + 权重排序
	mainDB = mainDB.
		Offset((req.PageNum - 1) * req.PageSize).
		Limit(req.PageSize)

	// 应用自定义排序
	applyCustomSort(mainDB, req)

	mainDB.Order("weight ASC") // 最后才权重排序

	// 查询到列表
	var list []models.GamePageQueryVo
	if err := mainDB.Find(&list).Error; err != nil {
		return pageData, common.NewDaoError("查询游戏列表失败: " + err.Error())
	}

	// 组装分页结果
	pageData.Data = list
	pageData.Total = total

	return pageData, nil
}

// buildSearchPageCondition 构建搜索分页查询条件
func buildSearchPageCondition(db *gorm.DB, req *models.SearchPageQueryRequest, rootDB *gorm.DB) {
	// 更新时间范围筛选
	if !req.UpdateStartTime.IsZero() && !req.UpdateEndTime.IsZero() {
		db.Where("gfg_game.update_time BETWEEN ? AND ?", req.UpdateStartTime, req.UpdateEndTime)
	}

	// 发售时间范围筛选
	if !req.PubStartTime.IsZero() && !req.PubEndTime.IsZero() {
		pubStart := formatLocalTime(req.PubStartTime)
		pubEnd := formatLocalTime(req.PubEndTime)
		db.Where("gfg_game.release_date IS NOT NULL AND to_date(gfg_game.release_date, 'YYYY.MM.DD') BETWEEN ? AND ?", pubStart, pubEnd)
	}

	// 关键词模糊搜索
	if req.Content != nil {
		keyword := "%" + strings.ReplaceAll(*req.Content, " ", "%") + "%"
		db.Where(`
			(
				gfg_game.name ILIKE ? 
				OR gfg_game.name_en ILIKE ? 
				OR gfg_game.info ILIKE ? 
				OR gfg_game.info_en ILIKE ? 
				OR gfg_game.developers::TEXT ILIKE ? 
				OR gfg_game.publishers::TEXT ILIKE ?
			)
		`, keyword, keyword, keyword, keyword, keyword, keyword)
	}

	// 标签筛选
	if len(req.TagList) > 0 {
		tagSubQuery := rootDB.Table("gfg_tag_map").
			Select("game_id").
			Where("tag_id IN (?)", req.TagList).
			Group("game_id").
			Having("COUNT(DISTINCT tag_id) = ?", len(req.TagList))

		db.Where("gfg_game.id IN (?)", tagSubQuery)
	}
}

// applyCustomSort 应用自定义排序规则
func applyCustomSort(db *gorm.DB, req models.SearchPageQueryRequest) {
	var orderClauses []string

	if req.TimeOrder {
		orderClauses = append(orderClauses, "gfg_game.update_time DESC")
	}

	if req.RemarkOrder {
		// 如果没有评论排到最后(NULL)
		orderClauses = append(orderClauses, "comment_stats.remark_count DESC NULLS LAST")
	}

	if req.ScoreOrder {
		// 没有评分排到最后(NULL)
		orderClauses = append(orderClauses, "comment_stats.avg_score DESC NULLS LAST")
	}

	if len(orderClauses) > 0 {
		db.Order(strings.Join(orderClauses, ", "))
	}
}

// getGameNameField 根据语言获取游戏名字段
func getGameNameField(lang string) string {
	if lang == "en" {
		return "gfg_game.name_en"
	}
	return "gfg_game.name"
}

// getGameInfoField 根据语言获取游戏详情字段
func getGameInfoField(lang string) string {
	if lang == "en" {
		return "gfg_game.info_en"
	}
	return "gfg_game.info"
}

// formatLocalTime 格式化本地时间为 YYYY.MM.DD 字符串
func formatLocalTime(t cm.LocalTime) string {
	return t.Time().Format("2006.01.02")
}
