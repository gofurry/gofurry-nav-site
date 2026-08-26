package dao

import (
	"context"
	"fmt"
	"strings"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	"github.com/gofurry/gofurry-game-backend/common"
	cm "github.com/gofurry/gofurry-game-backend/common/models"
)

const searchJoins = `
LEFT JOIN gfg_game_v2_details d ON d.game_id = g.id
LEFT JOIN gfg_game_first_available fa ON fa.game_id = g.id
LEFT JOIN gfg_game_release_state rs ON rs.game_id = g.id
LEFT JOIN gfg_game_v2_localized_details ld ON ld.game_id = g.id AND ld.lang = $1
LEFT JOIN (
    SELECT DISTINCT ON (game_id) game_id, url
    FROM gfg_game_v2_assets
    WHERE exists IS DISTINCT FROM false AND source = 'store_browse'
      AND asset_type IN ('header_2x', 'header')
    ORDER BY game_id,
      CASE asset_type WHEN 'header' THEN 0 WHEN 'header_2x' THEN 1 ELSE 2 END,
      sort_order, id
) asset_media ON asset_media.game_id = g.id
LEFT JOIN (
    SELECT game_id, COUNT(*) AS remark_count, AVG(score) AS avg_score
    FROM gfg_game_comment GROUP BY game_id
) comment_stats ON comment_stats.game_id = g.id
LEFT JOIN gfg_tag primary_tag ON g.primary_tag = primary_tag.id
LEFT JOIN gfg_tag secondary_tag ON g.secondary_tag = secondary_tag.id`

func (dao *ReadModelDAO) SearchGames(ctx context.Context, query v2models.GameV2SearchPageQuery) (cm.PageResponse, common.GFError) {
	res := cm.PageResponse{}
	if err := dao.ready(); err != nil {
		return res, common.NewDaoError(err.Error())
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
	where, args := buildSearchWhere(query)
	countSQL := "SELECT COUNT(*)::bigint FROM gfg_game g " + searchJoins + where
	if err := dao.pool.QueryRow(ctx, countSQL, args...).Scan(&res.Total); err != nil {
		return res, common.NewDaoError(fmt.Sprintf("统计游戏 v2 搜索结果失败: %v", err))
	}
	args = append(args, query.PageSize, (query.PageNum-1)*query.PageSize)
	itemsSQL := `SELECT ` + searchSelectSQL(query.Lang) + ` FROM gfg_game g ` + searchJoins + where +
		` ORDER BY ` + searchOrder(query) + fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	items, err := queryMany[v2models.GameV2SearchPageItem](ctx, dao.pool, itemsSQL, args...)
	if err != nil {
		return res, common.NewDaoError(fmt.Sprintf("查询游戏 v2 搜索结果失败: %v", err))
	}
	attachSearchCanonicalDomain(items)
	res.Data = items
	return res, nil
}

func buildSearchWhere(query v2models.GameV2SearchPageQuery) (string, []any) {
	args := []any{normalizeDAOLang(query.Lang)}
	clauses := []string{"TRUE"}
	if query.Content != "" {
		args = append(args, "%"+strings.ReplaceAll(query.Content, " ", "%")+"%")
		p := fmt.Sprintf("$%d", len(args))
		clauses = append(clauses, `(g.name ILIKE `+p+` OR g.name_en ILIKE `+p+`
OR g.info ILIKE `+p+` OR g.info_en ILIKE `+p+` OR d.name ILIKE `+p+`
OR d.developers::text ILIKE `+p+` OR d.publishers::text ILIKE `+p+`
OR ld.name ILIKE `+p+` OR ld.short_description ILIKE `+p+`
OR EXISTS (SELECT 1 FROM gfg_tag_map tm JOIN gfg_tag t ON t.id=tm.tag_id
WHERE tm.game_id=g.id AND (t.name ILIKE `+p+` OR t.name_en ILIKE `+p+`)))`)
	}
	if query.Availability != "" {
		args = append(args, query.Availability)
		clauses = append(clauses, fmt.Sprintf("rs.availability = $%d", len(args)))
	}
	if !query.UpdateStartTime.IsZero() && !query.UpdateEndTime.IsZero() {
		args = append(args, query.UpdateStartTime, query.UpdateEndTime)
		clauses = append(clauses, fmt.Sprintf("%s BETWEEN $%d AND $%d", searchUpdatedAtExpr(), len(args)-1, len(args)))
	}
	if query.Availability != "upcoming" && !query.PubStartTime.IsZero() && !query.PubEndTime.IsZero() {
		args = append(args, query.PubStartTime, query.PubEndTime)
		clauses = append(clauses, fmt.Sprintf("fa.window_start <= $%d::date AND fa.window_end >= $%d::date", len(args), len(args)-1))
	}
	if query.Availability == "upcoming" && !query.PlannedStartTime.IsZero() && !query.PlannedEndTime.IsZero() {
		args = append(args, query.PlannedStartTime, query.PlannedEndTime)
		clauses = append(clauses, fmt.Sprintf("rs.window_start <= $%d::date AND rs.window_end >= $%d::date", len(args), len(args)-1))
	}
	if len(query.TagList) > 0 {
		args = append(args, query.TagList, len(query.TagList))
		clauses = append(clauses, fmt.Sprintf(`g.id IN (SELECT game_id FROM gfg_tag_map
WHERE tag_id = ANY($%d::bigint[]) GROUP BY game_id HAVING COUNT(DISTINCT tag_id) = $%d)`, len(args)-1, len(args)))
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func searchSelectSQL(lang string) string {
	nameExpr := "COALESCE(NULLIF(g.name, ''), NULLIF(g.name_en, ''), NULLIF(ld.name, ''), NULLIF(d.name, ''), '')"
	infoExpr := "COALESCE(NULLIF(g.info, ''), NULLIF(g.info_en, ''), NULLIF(ld.short_description, ''), '')"
	primaryTagExpr := "COALESCE(NULLIF(primary_tag.name, ''), primary_tag.name_en, '')"
	secondaryTagExpr := "COALESCE(NULLIF(secondary_tag.name, ''), secondary_tag.name_en, '')"
	if normalizeDAOLang(lang) == "en" {
		nameExpr = "COALESCE(NULLIF(g.name_en, ''), NULLIF(g.name, ''), NULLIF(ld.name, ''), NULLIF(d.name, ''), '')"
		infoExpr = "COALESCE(NULLIF(g.info_en, ''), NULLIF(g.info, ''), NULLIF(ld.short_description, ''), '')"
		primaryTagExpr = "COALESCE(NULLIF(primary_tag.name_en, ''), primary_tag.name, '')"
		secondaryTagExpr = "COALESCE(NULLIF(secondary_tag.name_en, ''), secondary_tag.name, '')"
	}
	return fmt.Sprintf(`g.id::text AS id, %s AS name, %s AS info,
COALESCE(NULLIF(asset_media.url, ''), NULLIF(d.header_url, ''), NULLIF(g.header, ''), '') AS cover,
g.appid, %s AS update_time,
COALESCE(NULLIF(d.release_date_text, ''), '') AS release_date,
COALESCE(comment_stats.remark_count, 0)::integer AS remark_count,
COALESCE(comment_stats.avg_score, 0)::double precision AS avg_score,
%s AS primary_tag, %s AS secondary_tag,
fa.precision AS fa_precision, fa.exact_date AS fa_exact_date,
fa.release_year AS fa_release_year, fa.release_month AS fa_release_month,
fa.release_quarter AS fa_release_quarter, fa.window_start AS fa_window_start,
fa.window_end AS fa_window_end, fa.source AS fa_source, fa.inferred AS fa_inferred,
rs.availability AS rs_availability, rs.precision AS rs_precision,
rs.exact_date AS rs_exact_date, rs.release_year AS rs_release_year,
rs.release_month AS rs_release_month, rs.release_quarter AS rs_release_quarter,
rs.window_start AS rs_window_start, rs.window_end AS rs_window_end,
rs.raw_text AS rs_raw_text, rs.observed_at AS rs_observed_at`,
		nameExpr, infoExpr, searchUpdatedAtExpr(), primaryTagExpr, secondaryTagExpr)
}

func searchOrder(query v2models.GameV2SearchPageQuery) string {
	orders := []string{}
	if query.TimeOrder {
		if query.Availability == "upcoming" {
			orders = append(orders,
				"CASE WHEN rs.window_end >= CURRENT_DATE THEN 0 WHEN rs.window_end IS NOT NULL THEN 1 ELSE 2 END ASC",
				"rs.window_end ASC NULLS LAST",
				"rs.window_start ASC NULLS LAST",
				"CASE rs.precision WHEN 'day' THEN 1 WHEN 'month' THEN 2 WHEN 'quarter' THEN 3 WHEN 'year' THEN 4 WHEN 'tba' THEN 5 ELSE 6 END ASC",
			)
		} else {
			orders = append(orders, "g.create_time DESC")
		}
	}
	if query.RemarkOrder {
		orders = append(orders, "comment_stats.remark_count DESC NULLS LAST")
	}
	if query.ScoreOrder {
		orders = append(orders, "comment_stats.avg_score DESC NULLS LAST")
	}
	if query.TimeOrder {
		orders = append(orders, "g.weight DESC", "g.id ASC")
	} else {
		orders = append(orders, "g.weight ASC", "g.id ASC")
	}
	return strings.Join(orders, ", ")
}

func searchUpdatedAtExpr() string {
	return "GREATEST(g.update_time, COALESCE(d.updated_at, g.update_time), COALESCE(ld.updated_at, g.update_time))"
}
