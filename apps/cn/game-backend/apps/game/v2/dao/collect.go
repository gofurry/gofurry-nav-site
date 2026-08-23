package dao

import (
	"context"
	"fmt"
	"strings"
	"time"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	"github.com/gofurry/gofurry-game-backend/common"
)

func (dao *ReadModelDAO) GetCollectStatus(ctx context.Context) (v2models.GameV2CollectStatus, common.GFError) {
	res := v2models.GameV2CollectStatus{}
	if err := dao.ready(); err != nil {
		return res, common.NewDaoError(err.Error())
	}
	latest, err := queryOptional[v2models.GfgGameV2CollectRun](ctx, dao.pool,
		"SELECT "+collectRunColumns+" FROM gfg_game_v2_collect_runs ORDER BY started_at DESC,id DESC LIMIT 1")
	if err != nil {
		return res, common.NewDaoError(fmt.Sprintf("查询游戏 v2 最新采集批次失败: %v", err))
	}
	res.LatestRun = latest
	res.LatestTaskRuns, err = queryMany[v2models.GfgGameV2CollectRun](ctx, dao.pool,
		"SELECT "+collectRunColumns+` FROM gfg_game_v2_collect_runs
WHERE id IN (SELECT DISTINCT ON (task_type) id FROM gfg_game_v2_collect_runs
ORDER BY task_type,started_at DESC,id DESC) ORDER BY task_type`)
	if err != nil {
		return res, common.NewDaoError(fmt.Sprintf("查询游戏 v2 任务最新批次失败: %v", err))
	}
	res.Summary, err = queryMany[v2models.GameV2CollectTaskStatusSummary](ctx, dao.pool, `
SELECT task_type,status,COUNT(*)::bigint AS count FROM gfg_game_v2_collect_task_results
WHERE started_at >= NOW()-INTERVAL '7 days' GROUP BY task_type,status ORDER BY task_type,status`)
	if err != nil {
		return res, common.NewDaoError(fmt.Sprintf("查询游戏 v2 任务结果摘要失败: %v", err))
	}
	res.GeneratedAt = time.Now()
	return res, nil
}

func (dao *ReadModelDAO) ListCollectRuns(ctx context.Context, query v2models.GameV2CollectRunQuery) ([]v2models.GfgGameV2CollectRun, common.GFError) {
	if err := dao.ready(); err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.Offset < 0 {
		query.Offset = 0
	}
	where, args := optionalTextFilters([]filter{{"task_type", query.TaskType}, {"status", query.Status}})
	args = append(args, query.Limit, query.Offset)
	sql := fmt.Sprintf("SELECT %s FROM gfg_game_v2_collect_runs%s ORDER BY started_at DESC,id DESC LIMIT $%d OFFSET $%d",
		collectRunColumns, where, len(args)-1, len(args))
	rows, err := queryMany[v2models.GfgGameV2CollectRun](ctx, dao.pool, sql, args...)
	if err != nil {
		return nil, common.NewDaoError(fmt.Sprintf("查询游戏 v2 采集批次失败: %v", err))
	}
	return rows, nil
}

func (dao *ReadModelDAO) GetCollectRun(ctx context.Context, runID string) (*v2models.GfgGameV2CollectRun, common.GFError) {
	if err := dao.ready(); err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	if strings.TrimSpace(runID) == "" {
		return nil, common.NewDaoError("run_id is required")
	}
	row, err := queryOptional[v2models.GfgGameV2CollectRun](ctx, dao.pool,
		"SELECT "+collectRunColumns+" FROM gfg_game_v2_collect_runs WHERE id=$1", runID)
	if err != nil {
		return nil, common.NewDaoError(fmt.Sprintf("查询游戏 v2 采集批次失败: %v", err))
	}
	if row == nil {
		return nil, common.NewDaoError("collect run not found")
	}
	return row, nil
}

func (dao *ReadModelDAO) ListCollectTaskResults(ctx context.Context, query v2models.GameV2CollectTaskResultQuery) ([]v2models.GfgGameV2CollectTaskResult, common.GFError) {
	if err := dao.ready(); err != nil {
		return nil, common.NewDaoError(err.Error())
	}
	if query.Limit <= 0 {
		query.Limit = 50
	}
	if query.Offset < 0 {
		query.Offset = 0
	}
	clauses := []string{}
	args := []any{}
	for _, item := range []filter{{"run_id", query.RunID}, {"task_type", query.TaskType}, {"status", query.Status}} {
		if item.value != "" {
			args = append(args, item.value)
			clauses = append(clauses, fmt.Sprintf("%s=$%d", item.column, len(args)))
		}
	}
	if query.GameID > 0 {
		args = append(args, query.GameID)
		clauses = append(clauses, fmt.Sprintf("game_id=$%d", len(args)))
	}
	if query.AppID > 0 {
		args = append(args, query.AppID)
		clauses = append(clauses, fmt.Sprintf("appid=$%d", len(args)))
	}
	where := ""
	if len(clauses) > 0 {
		where = " WHERE " + strings.Join(clauses, " AND ")
	}
	args = append(args, query.Limit, query.Offset)
	sql := fmt.Sprintf("SELECT %s FROM gfg_game_v2_collect_task_results%s ORDER BY started_at DESC,id DESC LIMIT $%d OFFSET $%d",
		collectTaskResultColumns, where, len(args)-1, len(args))
	rows, err := queryMany[v2models.GfgGameV2CollectTaskResult](ctx, dao.pool, sql, args...)
	if err != nil {
		return nil, common.NewDaoError(fmt.Sprintf("查询游戏 v2 任务结果失败: %v", err))
	}
	return rows, nil
}

func (dao *ReadModelDAO) GetGameCollectStatus(ctx context.Context, gameID int64, appID int64) (v2models.GameV2CollectGameStatus, common.GFError) {
	res := v2models.GameV2CollectGameStatus{}
	if err := dao.ready(); err != nil {
		return res, common.NewDaoError(err.Error())
	}
	if gameID <= 0 && appID <= 0 {
		return res, common.NewDaoError("game_id or appid is required")
	}
	site, err := dao.loadSiteRecord(ctx, DetailQuery{GameID: gameID, AppID: appID})
	if err != nil {
		return res, common.NewDaoError(err.Error())
	}
	res.GameID, res.AppID, res.Name = site.ID, site.AppID, site.Name

	details, err := queryOptional[v2models.GfgGameV2Details](ctx, dao.pool,
		"SELECT "+detailsColumns+" FROM gfg_game_v2_details WHERE game_id=$1", site.ID)
	if err != nil {
		return res, common.NewDaoError(fmt.Sprintf("查询游戏 v2 详情新鲜度失败: %v", err))
	}
	if details != nil {
		updatedAt := details.UpdatedAt
		res.DetailsUpdatedAt = &updatedAt
	}
	localized, err := queryMany[v2models.GfgGameV2LocalizedDetails](ctx, dao.pool,
		"SELECT "+localizedColumns+" FROM gfg_game_v2_localized_details WHERE game_id=$1 ORDER BY lang", site.ID)
	if err != nil {
		return res, common.NewDaoError(fmt.Sprintf("查询游戏 v2 本地化新鲜度失败: %v", err))
	}
	res.Localized = make([]v2models.GameV2CollectLocalizedStatus, 0, len(localized))
	for _, item := range localized {
		res.Localized = append(res.Localized, v2models.GameV2CollectLocalizedStatus{
			Lang: item.Lang, Name: item.Name, CollectedAt: item.CollectedAt, UpdatedAt: item.UpdatedAt})
	}
	prices, err := queryMany[v2models.GfgGameV2Price](ctx, dao.pool,
		"SELECT "+priceColumns+" FROM gfg_game_v2_prices WHERE game_id=$1 ORDER BY region", site.ID)
	if err != nil {
		return res, common.NewDaoError(fmt.Sprintf("查询游戏 v2 价格新鲜度失败: %v", err))
	}
	res.Prices = make([]v2models.GameV2CollectRegionFreshness, 0, len(prices))
	for _, price := range prices {
		available := price.IsFree || (strings.TrimSpace(strPtrValue(price.Currency)) != "" &&
			(price.FinalAmount > 0 || strings.TrimSpace(strPtrValue(price.FinalFormatted)) != ""))
		res.Prices = append(res.Prices, v2models.GameV2CollectRegionFreshness{
			Region: price.Region, Available: available, Currency: strPtrValue(price.Currency),
			FinalAmount: price.FinalAmount, CollectedAt: price.CollectedAt, UpdatedAt: price.UpdatedAt})
	}
	if err := dao.pool.QueryRow(ctx, `SELECT COUNT(*)::bigint FROM gfg_game_v2_media WHERE game_id=$1`, site.ID).Scan(&res.MediaCount); err != nil {
		return res, common.NewDaoError(fmt.Sprintf("查询游戏 v2 媒体数量失败: %v", err))
	}
	if err := dao.pool.QueryRow(ctx, `SELECT COUNT(*)::bigint,MAX(published_at) FROM gfg_game_v2_news WHERE game_id=$1`, site.ID).
		Scan(&res.NewsCount, &res.LatestNewsAt); err != nil {
		return res, common.NewDaoError(fmt.Sprintf("查询游戏 v2 新闻状态失败: %v", err))
	}
	res.LatestPlayerCount, err = queryOptional[v2models.GfgGameV2PlayerCount](ctx, dao.pool,
		"SELECT "+playerCountColumns+" FROM gfg_game_v2_player_counts WHERE game_id=$1 ORDER BY collected_at DESC,id DESC LIMIT 1", site.ID)
	if err != nil {
		return res, common.NewDaoError(fmt.Sprintf("查询游戏 v2 在线人数新鲜度失败: %v", err))
	}
	res.LatestTaskResults, err = queryMany[v2models.GfgGameV2CollectTaskResult](ctx, dao.pool,
		"SELECT "+collectTaskResultColumns+` FROM gfg_game_v2_collect_task_results
WHERE id IN (SELECT DISTINCT ON (task_type) id FROM gfg_game_v2_collect_task_results
WHERE game_id=$1 ORDER BY task_type,started_at DESC,id DESC) ORDER BY task_type`, site.ID)
	if err != nil {
		return res, common.NewDaoError(fmt.Sprintf("查询游戏 v2 游戏最新任务结果失败: %v", err))
	}
	return res, nil
}

type filter struct {
	column string
	value  string
}

func optionalTextFilters(filters []filter) (string, []any) {
	clauses := []string{}
	args := []any{}
	for _, item := range filters {
		if item.value == "" {
			continue
		}
		args = append(args, item.value)
		clauses = append(clauses, fmt.Sprintf("%s=$%d", item.column, len(args)))
	}
	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}
