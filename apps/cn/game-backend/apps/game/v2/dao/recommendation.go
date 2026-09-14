package dao

import (
	"context"
	"fmt"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	"github.com/gofurry/gofurry-game-backend/common"
)

func (dao *ReadModelDAO) ListSimilarRecommendations(ctx context.Context, query v2models.GameV2SimilarRecommendationQuery) ([]v2models.GameV2RecommendationRow, common.GFError) {
	if err := dao.ready(); err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	if query.GameID <= 0 {
		return nil, common.NewDaoError("game_id is required")
	}
	if query.Limit <= 0 {
		query.Limit = 8
	}
	rows, err := queryMany[v2models.GameV2RecommendationRow](ctx, dao.pool, recommendationRowsSQL,
		normalizeDAOLang(query.Lang), normalizeDAORegion(query.Region), query.GameID, query.AlgorithmVersion, query.Limit)
	if err != nil {
		return nil, common.NewDaoError(fmt.Sprintf("查询游戏 v2 相似推荐失败: %v", err))
	}
	return rows, nil
}

func (dao *ReadModelDAO) SaveSimilarRecommendations(ctx context.Context, sourceGameID int64, rows []v2models.GfgGameV2Recommendation) common.GFError {
	if err := dao.ready(); err != nil {
		return common.NewDaoError(err.Error())
	}
	if sourceGameID <= 0 {
		return common.NewDaoError("source_game_id is required")
	}
	tx, err := dao.pool.Begin(ctx)
	if err != nil {
		return common.NewDaoError(fmt.Sprintf("保存游戏 v2 相似推荐失败: %v", err))
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `DELETE FROM gfg_game_recommendations WHERE source_game_id = $1`, sourceGameID); err != nil {
		return common.NewDaoError(fmt.Sprintf("保存游戏 v2 相似推荐失败: %v", err))
	}
	for _, row := range rows {
		_, err = tx.Exec(ctx, `INSERT INTO gfg_game_recommendations
(source_game_id, target_game_id, score, display_score, rank, reason_json, algorithm_version, computed_at)
VALUES ($1,$2,$3,$4,$5,$6::jsonb,$7,$8)
ON CONFLICT (source_game_id, target_game_id) DO UPDATE SET
score=EXCLUDED.score, display_score=EXCLUDED.display_score, rank=EXCLUDED.rank,
reason_json=EXCLUDED.reason_json, algorithm_version=EXCLUDED.algorithm_version,
computed_at=EXCLUDED.computed_at`, row.SourceGameID, row.TargetGameID, row.Score,
			row.DisplayScore, row.Rank, row.ReasonJSON, row.AlgorithmVersion, row.ComputedAt)
		if err != nil {
			return common.NewDaoError(fmt.Sprintf("保存游戏 v2 相似推荐失败: %v", err))
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return common.NewDaoError(fmt.Sprintf("保存游戏 v2 相似推荐失败: %v", err))
	}
	return nil
}

func (dao *ReadModelDAO) ListRecommendationFeatures(ctx context.Context, lang string, region string) ([]v2models.GameV2RecommendationFeature, common.GFError) {
	if err := dao.ready(); err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	rows, err := queryMany[v2models.GameV2RecommendationFeature](ctx, dao.pool, recommendationFeaturesSQL,
		normalizeDAOLang(lang), normalizeDAORegion(region))
	if err != nil {
		return nil, common.NewDaoError(fmt.Sprintf("查询游戏 v2 推荐特征失败: %v", err))
	}
	return rows, nil
}

const recommendationAssetColumnsSQL = `
COALESCE((SELECT a.url FROM gfg_game_assets a
  WHERE a.game_id=g.id AND a.exists IS DISTINCT FROM false AND a.source='store_browse'
    AND a.asset_type IN ('header_2x','header') AND COALESCE(a.url,'')<>''
  ORDER BY CASE a.asset_type WHEN 'header' THEN 0 WHEN 'header_2x' THEN 1 ELSE 2 END,
    a.sort_order,a.id LIMIT 1), NULLIF(d.header_url,''), NULLIF(g.header,''), '') AS header_url,
COALESCE((SELECT a.url FROM gfg_game_assets a
  WHERE a.game_id=g.id AND a.exists IS DISTINCT FROM false AND a.source='store_browse'
    AND a.asset_type IN ('capsule_main_2x','capsule_main','hero_capsule_2x','hero_capsule') AND COALESCE(a.url,'')<>''
  ORDER BY CASE a.asset_type WHEN 'capsule_main' THEN 0 WHEN 'hero_capsule' THEN 1
    WHEN 'capsule_main_2x' THEN 2 WHEN 'hero_capsule_2x' THEN 3 ELSE 4 END,
    a.sort_order,a.id LIMIT 1), '') AS capsule_url,
COALESCE(
  (SELECT a.url FROM gfg_game_assets a WHERE a.game_id=g.id
    AND a.exists IS DISTINCT FROM false AND a.source='store_browse'
    AND a.asset_type='library_capsule' AND COALESCE(a.url,'')<>'' ORDER BY a.sort_order,a.id LIMIT 1),
  (SELECT a.url FROM gfg_game_assets a WHERE a.game_id=g.id
    AND a.exists IS DISTINCT FROM false AND a.source='store_browse'
    AND a.asset_type='library_capsule_2x' AND COALESCE(a.url,'')<>'' ORDER BY a.sort_order,a.id LIMIT 1), '') AS library_cover_url,
COALESCE(
  (SELECT a.url FROM gfg_game_assets a WHERE a.game_id=g.id
    AND a.exists IS DISTINCT FROM false AND a.source='store_browse'
    AND a.asset_type='library_capsule_2x' AND COALESCE(a.url,'')<>'' ORDER BY a.sort_order,a.id LIMIT 1),
  (SELECT a.url FROM gfg_game_assets a WHERE a.game_id=g.id
    AND a.exists IS DISTINCT FROM false AND a.source='store_browse'
    AND a.asset_type='library_capsule' AND COALESCE(a.url,'')<>'' ORDER BY a.sort_order,a.id LIMIT 1), '') AS library_cover_2x_url`

const recommendationProjectionSQL = `
g.appid,
CASE WHEN $1='en' THEN COALESCE(NULLIF(g.name_en,''),NULLIF(g.name,''),NULLIF(ld.name,''),NULLIF(d.name,''),'')
ELSE COALESCE(NULLIF(g.name,''),NULLIF(g.name_en,''),NULLIF(ld.name,''),NULLIF(d.name,''),'') END AS name,
CASE WHEN $1='en' THEN COALESCE(NULLIF(g.info_en,''),NULLIF(g.info,''),NULLIF(ld.short_description,''),'')
ELSE COALESCE(NULLIF(g.info,''),NULLIF(g.info_en,''),NULLIF(ld.short_description,''),'') END AS summary,` + recommendationAssetColumnsSQL + `,
COALESCE(tags.tags,'[]'::jsonb)::text AS tags,
COALESCE(p.region,$2) AS price_region,
CASE WHEN p.game_id IS NULL THEN false WHEN p.is_free THEN true
 WHEN COALESCE(p.currency,'')<>'' AND (p.final_amount>0 OR COALESCE(p.final_formatted,'')<>'') THEN true
 ELSE false END AS price_available,
COALESCE(p.is_free,false) AS is_free, COALESCE(p.currency,'') AS currency,
COALESCE(p.initial_amount,0) AS initial_amount, COALESCE(p.final_amount,0) AS final_amount,
COALESCE(p.discount_percent,0) AS discount_percent,
COALESCE(p.initial_formatted,'') AS initial_formatted,
COALESCE(p.final_formatted,'') AS final_formatted, p.updated_at AS price_updated_at,
COALESCE(player.count,0) AS online_count, COALESCE(player.status,'unknown') AS online_status,
player.collected_at AS online_collected_at`

const recommendationJoinsSQL = `
JOIN gfg_game_details d ON d.game_id=g.id
LEFT JOIN gfg_game_localized_details ld ON ld.game_id=g.id AND ld.lang=$1
LEFT JOIN gfg_game_prices p ON p.game_id=g.id AND p.region=$2
LEFT JOIN LATERAL (
 SELECT pc.count,pc.status,pc.collected_at FROM gfg_game_player_counts pc
 WHERE pc.game_id=g.id AND pc.status='success' ORDER BY pc.collected_at DESC,pc.id DESC LIMIT 1
) player ON true
LEFT JOIN LATERAL (
 SELECT jsonb_agg(jsonb_build_object(
   'id',t.id::text,
   'name',CASE WHEN $1='en' THEN COALESCE(NULLIF(t.name_en,''),t.name) ELSE COALESCE(NULLIF(t.name,''),t.name_en) END,
   'desc',CASE WHEN $1='en' THEN COALESCE(NULLIF(t.info_en,''),t.info) ELSE COALESCE(NULLIF(t.info,''),t.info_en) END,
   'prefix',t.prefix::text) ORDER BY t.id) AS tags
 FROM gfg_tag_map tm JOIN gfg_tag t ON t.id=tm.tag_id WHERE tm.game_id=g.id
) tags ON true`

const recommendationRowsSQL = `SELECT
r.source_game_id,r.target_game_id,r.score,r.display_score,r.rank,
r.reason_json::text AS reason_json,r.algorithm_version,r.computed_at,` + recommendationProjectionSQL + `
FROM gfg_game_recommendations r
JOIN gfg_game g ON g.id=r.target_game_id
` + recommendationJoinsSQL + `
WHERE r.source_game_id=$3 AND r.algorithm_version=$4
ORDER BY r.rank,r.score DESC,r.target_game_id LIMIT $5`

const recommendationFeaturesSQL = `SELECT g.id AS game_id,` + recommendationProjectionSQL + `,
d.developers::text AS developers,d.publishers::text AS publishers,d.platforms::text AS platforms,
COALESCE(g.primary_tag,0) AS primary_tag_id,COALESCE(g.secondary_tag,0) AS secondary_tag_id,
GREATEST(g.update_time,COALESCE(d.updated_at,g.update_time),COALESCE(ld.updated_at,g.update_time)) AS updated_at
FROM gfg_game g
` + recommendationJoinsSQL + `
ORDER BY g.weight,g.id`
