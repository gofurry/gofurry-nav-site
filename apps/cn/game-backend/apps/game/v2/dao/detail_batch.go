package dao

import (
	"context"
	"fmt"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	"github.com/jackc/pgx/v5"
)

// Queue independent detail reads on one connection, without increasing the
// connection budget per request. Callbacks are consumed in queue order by Close.
func (dao *ReadModelDAO) loadDetailBatch(ctx context.Context, aggregate *v2models.GameV2Aggregate, lang string) error {
	gameID := aggregate.Site.ID
	requested := normalizeDAOLang(lang)
	fallback := localizedFallbackLang(requested)
	var batch pgx.Batch
	batch.Queue("SELECT "+detailsColumns+" FROM gfg_game_details WHERE game_id = $1", gameID).Query(func(rows pgx.Rows) error {
		var err error
		aggregate.Details, err = collectOptional[v2models.GfgGameV2Details](rows)
		return err
	})
	batch.Queue("SELECT "+localizedColumns+" FROM gfg_game_localized_details WHERE game_id = $1 AND lang = ANY($2::text[])", gameID, []string{requested, fallback}).Query(func(rows pgx.Rows) error {
		items, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[v2models.GfgGameV2LocalizedDetails])
		if err != nil {
			return err
		}
		var primary, secondary *v2models.GfgGameV2LocalizedDetails
		for i := range items {
			if normalizeDAOLang(items[i].Lang) == requested {
				primary = &items[i]
			} else {
				secondary = &items[i]
			}
		}
		aggregate.Localized = mergeLocalizedDetails(primary, secondary)
		return nil
	})
	batch.Queue("SELECT "+priceColumns+" FROM gfg_game_prices WHERE game_id = $1 ORDER BY region ASC", gameID).Query(func(rows pgx.Rows) error {
		var err error
		aggregate.Prices, err = pgx.CollectRows(rows, pgx.RowToStructByNameLax[v2models.GfgGameV2Price])
		return err
	})
	batch.Queue("SELECT "+mediaColumns+" FROM gfg_game_media WHERE game_id = $1 ORDER BY media_type, sort_order, id", gameID).Query(func(rows pgx.Rows) error {
		var err error
		aggregate.Media, err = pgx.CollectRows(rows, pgx.RowToStructByNameLax[v2models.GfgGameV2Media])
		return err
	})
	batch.Queue("SELECT "+assetColumns+" FROM gfg_game_assets WHERE game_id = $1 ORDER BY asset_family, sort_order, id", gameID).Query(func(rows pgx.Rows) error {
		var err error
		aggregate.Assets, err = pgx.CollectRows(rows, pgx.RowToStructByNameLax[v2models.GfgGameV2Asset])
		return err
	})
	batch.Queue("SELECT "+requirementsColumns+" FROM gfg_game_requirements WHERE game_id = $1", gameID).Query(func(rows pgx.Rows) error {
		var err error
		aggregate.Requirements, err = collectOptional[v2models.GfgGameV2Requirements](rows)
		return err
	})
	batch.Queue(`SELECT COALESCE(AVG(score), 0)::double precision AS avg_score, COUNT(*)::bigint AS comment_count
FROM gfg_game_comment WHERE game_id = $1`, gameID).Query(func(rows pgx.Rows) error {
		var err error
		aggregate.ReviewStats, err = pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[v2models.GameV2ReviewStats])
		return err
	})
	nameColumns := "t.name AS name, t.info AS desc"
	if requested == "en" {
		nameColumns = "t.name_en AS name, t.info_en AS desc"
	}
	batch.Queue(fmt.Sprintf(`SELECT t.id::text AS id, %s FROM gfg_tag_map tm
JOIN gfg_tag t ON tm.tag_id=t.id WHERE tm.game_id=$1 ORDER BY t.id`, nameColumns), gameID).Query(func(rows pgx.Rows) error {
		var err error
		aggregate.Tags, err = pgx.CollectRows(rows, pgx.RowToStructByNameLax[v2models.GameV2Tag])
		return err
	})
	return dao.pool.SendBatch(ctx, &batch).Close()
}
