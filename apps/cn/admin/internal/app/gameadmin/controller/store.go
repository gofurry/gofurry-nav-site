package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofurry/gofurry-admin/internal/app/gameadmin/models"
	"github.com/gofurry/gofurry-admin/internal/app/shared/adminutil"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	gamesqlc "github.com/gofurry/gofurry-admin/internal/db/game/sqlc"
	"github.com/gofurry/gofurry-admin/pkg/common"
	pkgmodels "github.com/gofurry/gofurry-admin/pkg/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type gameStore struct {
	pool  *pgxpool.Pool
	q     *gamesqlc.Queries
	audit *audit.Logger
}

type gameMutation func(*gamesqlc.Queries) (targetID int64, before any, after any, err error)

func newGameStore(pool *pgxpool.Pool, auditLogger *audit.Logger) *gameStore {
	return &gameStore{pool: pool, q: gamesqlc.New(pool), audit: auditLogger}
}

func (store *gameStore) mutate(ctx context.Context, meta audit.Meta, action, resource string, change gameMutation) common.Error {
	tx, err := store.pool.Begin(ctx)
	if err != nil {
		return gameDAOError(err)
	}
	defer tx.Rollback(ctx)
	targetID, before, after, err := change(store.q.WithTx(tx))
	if err != nil {
		return gameDAOError(err)
	}
	// gfg and gfa are separate databases. The audit write remains independent,
	// matching the previous behavior while still rolling gfg back on audit failure.
	if err := store.audit.Log(ctx, meta, action, resource, targetID, before, after); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return gameDAOError(err)
	}
	return nil
}

func (store *gameStore) listGames(ctx context.Context, page adminutil.PageQuery) (int64, []models.Game, common.Error) {
	total, err := store.q.CountGames(ctx, page.Keyword)
	if err != nil {
		return 0, nil, gameDAOError(err)
	}
	limit, offset := pageArgsGame(page)
	rows, err := store.q.ListGames(ctx, gamesqlc.ListGamesParams{Keyword: page.Keyword, RowLimit: limit, RowOffset: offset})
	if err != nil {
		return 0, nil, gameDAOError(err)
	}
	items := make([]models.Game, 0, len(rows))
	for _, row := range rows {
		items = append(items, listGameModel(row))
	}
	return total, items, nil
}

func (store *gameStore) getGame(ctx context.Context, id int64) (models.Game, common.Error) {
	row, err := store.q.GetGame(ctx, id)
	return getGameModel(row), gameDAOError(err)
}

func (store *gameStore) createGame(ctx context.Context, meta audit.Meta, req gamesqlc.InsertGameParams) (models.Game, common.Error) {
	var result models.Game
	err := store.mutate(ctx, meta, "create", "gfg_game", func(q *gamesqlc.Queries) (int64, any, any, error) {
		if err := ensureUniqueGameAppIDQuery(ctx, q, req.Appid, 0); err != nil {
			return 0, nil, nil, err
		}
		id, err := q.NextGameID(ctx)
		if err != nil {
			return 0, nil, nil, err
		}
		req.ID = id
		row, err := q.InsertGame(ctx, req)
		if err == nil {
			gameID := id
			_, err = q.EnqueueGameEntityCollectionJob(ctx, gamesqlc.EnqueueGameEntityCollectionJobParams{
				Trigger: "entity_created", GameID: gameID, Appid: req.Appid, RequestedBy: meta.Operator,
			})
		}
		result = insertGameModel(row)
		return id, nil, row, err
	})
	return result, err
}

func (store *gameStore) updateGame(ctx context.Context, meta audit.Meta, req gamesqlc.UpdateGameParams) (models.Game, bool, common.Error) {
	var result models.Game
	var appIDChanged bool
	err := store.mutate(ctx, meta, "update", "gfg_game", func(q *gamesqlc.Queries) (int64, any, any, error) {
		before, err := q.LockGameForUpdate(ctx, req.ID)
		if err != nil {
			return req.ID, nil, nil, err
		}
		if err := ensureUniqueGameAppIDQuery(ctx, q, req.Appid, req.ID); err != nil {
			return req.ID, before, nil, err
		}
		after, err := q.UpdateGame(ctx, req)
		if err != nil {
			return req.ID, before, nil, err
		}
		appIDChanged = before.Appid != after.Appid
		if appIDChanged {
			if err := q.ResetSteamDerivedGameState(ctx, req.ID); err != nil {
				return req.ID, before, after, err
			}
			gameID := req.ID
			if _, err := q.EnqueueGameEntityCollectionJob(ctx, gamesqlc.EnqueueGameEntityCollectionJobParams{
				Trigger: "entity_changed", GameID: gameID, Appid: after.Appid, RequestedBy: meta.Operator,
			}); err != nil {
				return req.ID, before, after, err
			}
		}
		result = updateGameModel(after)
		return req.ID, before, after, nil
	})
	return result, appIDChanged, err
}

func (store *gameStore) deleteGame(ctx context.Context, meta audit.Meta, id int64) common.Error {
	return store.delete(ctx, meta, id, "gfg_game", func(q *gamesqlc.Queries) (any, error) { return q.GetGame(ctx, id) }, func(q *gamesqlc.Queries) error { _, err := q.DeleteGame(ctx, id); return err })
}

func ensureUniqueGameAppIDQuery(ctx context.Context, q *gamesqlc.Queries, appID, excludeID int64) error {
	if appID <= 0 {
		return nil
	}
	existing, err := q.FindGameByAppIDExcluding(ctx, gamesqlc.FindGameByAppIDExcludingParams{Appid: appID, ExcludeID: excludeID})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	return common.NewValidationError(fmt.Sprintf("appid already exists (game id=%d, name=%s)", existing.ID, existing.Name))
}

func (store *gameStore) listComments(ctx context.Context, page adminutil.PageQuery) (int64, []models.GameComment, common.Error) {
	total, err := store.q.CountGameComments(ctx, page.Keyword)
	if err != nil {
		return 0, nil, gameDAOError(err)
	}
	limit, offset := pageArgsGame(page)
	rows, err := store.q.ListGameComments(ctx, gamesqlc.ListGameCommentsParams{Keyword: page.Keyword, RowLimit: limit, RowOffset: offset})
	if err != nil {
		return 0, nil, gameDAOError(err)
	}
	items := make([]models.GameComment, 0, len(rows))
	for _, row := range rows {
		items = append(items, commentModel(row))
	}
	return total, items, nil
}

func (store *gameStore) getComment(ctx context.Context, id int64) (models.GameComment, common.Error) {
	row, err := store.q.GetGameComment(ctx, id)
	return commentModel(row), gameDAOError(err)
}

func (store *gameStore) createComment(ctx context.Context, meta audit.Meta, req gamesqlc.InsertGameCommentParams) (models.GameComment, common.Error) {
	var result models.GameComment
	err := store.mutate(ctx, meta, "create", "gfg_game_comment", func(q *gamesqlc.Queries) (int64, any, any, error) {
		row, err := q.InsertGameComment(ctx, req)
		result = commentModel(row)
		return req.ID, nil, row, err
	})
	return result, err
}

func (store *gameStore) updateComment(ctx context.Context, meta audit.Meta, req gamesqlc.UpdateGameCommentParams) common.Error {
	return store.mutate(ctx, meta, "update", "gfg_game_comment", func(q *gamesqlc.Queries) (int64, any, any, error) {
		before, err := q.GetGameComment(ctx, req.ID)
		if err != nil {
			return req.ID, nil, nil, err
		}
		after, err := q.UpdateGameComment(ctx, req)
		return req.ID, before, after, err
	})
}

func (store *gameStore) deleteComment(ctx context.Context, meta audit.Meta, id int64) common.Error {
	return store.delete(ctx, meta, id, "gfg_game_comment", func(q *gamesqlc.Queries) (any, error) { return q.GetGameComment(ctx, id) }, func(q *gamesqlc.Queries) error { _, err := q.DeleteGameComment(ctx, id); return err })
}

func (store *gameStore) listPrizes(ctx context.Context, page adminutil.PageQuery) (int64, []models.Prize, common.Error) {
	total, err := store.q.CountPrizes(ctx, page.Keyword)
	if err != nil {
		return 0, nil, gameDAOError(err)
	}
	limit, offset := pageArgsGame(page)
	rows, err := store.q.ListPrizes(ctx, gamesqlc.ListPrizesParams{Keyword: page.Keyword, RowLimit: limit, RowOffset: offset})
	if err != nil {
		return 0, nil, gameDAOError(err)
	}
	items := make([]models.Prize, 0, len(rows))
	for _, row := range rows {
		items = append(items, prizeModel(row))
	}
	return total, items, nil
}

func (store *gameStore) getPrize(ctx context.Context, id int64) (models.Prize, common.Error) {
	row, err := store.q.GetPrize(ctx, id)
	return prizeModel(row), gameDAOError(err)
}

func (store *gameStore) createPrize(ctx context.Context, meta audit.Meta, req gamesqlc.InsertPrizeParams) (models.Prize, common.Error) {
	var result models.Prize
	err := store.mutate(ctx, meta, "create", "gfg_prize", func(q *gamesqlc.Queries) (int64, any, any, error) {
		id, err := q.NextPrizeID(ctx)
		if err != nil {
			return 0, nil, nil, err
		}
		req.ID = id
		row, err := q.InsertPrize(ctx, req)
		result = prizeModel(row)
		return id, nil, row, err
	})
	return result, err
}

func (store *gameStore) updatePrize(ctx context.Context, meta audit.Meta, req gamesqlc.UpdatePrizeParams) common.Error {
	return store.mutate(ctx, meta, "update", "gfg_prize", func(q *gamesqlc.Queries) (int64, any, any, error) {
		before, err := q.GetPrize(ctx, req.ID)
		if err != nil {
			return req.ID, nil, nil, err
		}
		after, err := q.UpdatePrize(ctx, req)
		return req.ID, before, after, err
	})
}

func (store *gameStore) deletePrize(ctx context.Context, meta audit.Meta, id int64) common.Error {
	return store.delete(ctx, meta, id, "gfg_prize", func(q *gamesqlc.Queries) (any, error) { return q.GetPrize(ctx, id) }, func(q *gamesqlc.Queries) error { _, err := q.DeletePrize(ctx, id); return err })
}

func (store *gameStore) listTags(ctx context.Context, page adminutil.PageQuery) (int64, []models.Tag, common.Error) {
	total, err := store.q.CountTags(ctx, page.Keyword)
	if err != nil {
		return 0, nil, gameDAOError(err)
	}
	limit, offset := pageArgsGame(page)
	rows, err := store.q.ListTags(ctx, gamesqlc.ListTagsParams{Keyword: page.Keyword, RowLimit: limit, RowOffset: offset})
	if err != nil {
		return 0, nil, gameDAOError(err)
	}
	items := make([]models.Tag, 0, len(rows))
	for _, row := range rows {
		items = append(items, tagModel(row))
	}
	return total, items, nil
}

func (store *gameStore) getTag(ctx context.Context, id int64) (models.Tag, common.Error) {
	row, err := store.q.GetTag(ctx, id)
	return tagModel(row), gameDAOError(err)
}

func (store *gameStore) createTag(ctx context.Context, meta audit.Meta, req gamesqlc.InsertTagParams) (models.Tag, common.Error) {
	var result models.Tag
	err := store.mutate(ctx, meta, "create", "gfg_tag", func(q *gamesqlc.Queries) (int64, any, any, error) {
		row, err := q.InsertTag(ctx, req)
		result = tagModel(row)
		return req.ID, nil, row, err
	})
	return result, err
}

func (store *gameStore) updateTag(ctx context.Context, meta audit.Meta, req gamesqlc.UpdateTagParams) common.Error {
	return store.mutate(ctx, meta, "update", "gfg_tag", func(q *gamesqlc.Queries) (int64, any, any, error) {
		before, err := q.GetTag(ctx, req.ID)
		if err != nil {
			return req.ID, nil, nil, err
		}
		after, err := q.UpdateTag(ctx, req)
		return req.ID, before, after, err
	})
}

func (store *gameStore) deleteTag(ctx context.Context, meta audit.Meta, id int64) common.Error {
	return store.delete(ctx, meta, id, "gfg_tag", func(q *gamesqlc.Queries) (any, error) { return q.GetTag(ctx, id) }, func(q *gamesqlc.Queries) error { _, err := q.DeleteTag(ctx, id); return err })
}

func (store *gameStore) listTagMaps(ctx context.Context, page adminutil.PageQuery) (int64, []models.TagMapDTO, common.Error) {
	total, err := store.q.CountTagMaps(ctx, page.Keyword)
	if err != nil {
		return 0, nil, gameDAOError(err)
	}
	limit, offset := pageArgsGame(page)
	rows, err := store.q.ListTagMaps(ctx, gamesqlc.ListTagMapsParams{Keyword: page.Keyword, RowLimit: limit, RowOffset: offset})
	if err != nil {
		return 0, nil, gameDAOError(err)
	}
	items := make([]models.TagMapDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, models.TagMapDTO{ID: row.ID, GameID: row.GameID, TagID: row.TagID, GameName: row.GameName, TagName: row.TagName, CreateTime: localTime(row.CreateTime), UpdateTime: localTime(row.UpdateTime)})
	}
	return total, items, nil
}

func (store *gameStore) getTagMap(ctx context.Context, id int64) (models.TagMap, common.Error) {
	row, err := store.q.GetTagMap(ctx, id)
	return tagMapModel(row), gameDAOError(err)
}

func (store *gameStore) createTagMap(ctx context.Context, meta audit.Meta, gameID, tagID int64) (models.TagMap, common.Error) {
	var result models.TagMap
	err := store.mutate(ctx, meta, "create", "gfg_tag_map", func(q *gamesqlc.Queries) (int64, any, any, error) {
		id, err := q.NextTagMapID(ctx)
		if err != nil {
			return 0, nil, nil, err
		}
		row, err := q.InsertTagMap(ctx, gamesqlc.InsertTagMapParams{ID: id, GameID: gameID, TagID: tagID})
		result = tagMapModel(row)
		return id, nil, row, err
	})
	return result, err
}

func (store *gameStore) updateTagMap(ctx context.Context, meta audit.Meta, req gamesqlc.UpdateTagMapParams) common.Error {
	return store.mutate(ctx, meta, "update", "gfg_tag_map", func(q *gamesqlc.Queries) (int64, any, any, error) {
		before, err := q.GetTagMap(ctx, req.ID)
		if err != nil {
			return req.ID, nil, nil, err
		}
		after, err := q.UpdateTagMap(ctx, req)
		return req.ID, before, after, err
	})
}

func (store *gameStore) deleteTagMap(ctx context.Context, meta audit.Meta, id int64) common.Error {
	return store.delete(ctx, meta, id, "gfg_tag_map", func(q *gamesqlc.Queries) (any, error) { return q.GetTagMap(ctx, id) }, func(q *gamesqlc.Queries) error { _, err := q.DeleteTagMap(ctx, id); return err })
}

func (store *gameStore) replaceGameTags(ctx context.Context, meta audit.Meta, gameID int64, tagIDs []int64) common.Error {
	return store.mutate(ctx, meta, "bulk_replace", "gfg_tag_map", func(q *gamesqlc.Queries) (int64, any, any, error) {
		before, err := q.ListTagMapsByGame(ctx, gameID)
		if err != nil {
			return gameID, nil, nil, err
		}
		if _, err = q.DeleteTagMapsByGame(ctx, gameID); err != nil {
			return gameID, before, nil, err
		}
		if len(tagIDs) > 0 {
			firstID, err := q.NextTagMapID(ctx)
			if err != nil {
				return gameID, before, nil, err
			}
			for index, tagID := range tagIDs {
				if _, err = q.InsertTagMap(ctx, gamesqlc.InsertTagMapParams{ID: firstID + int64(index), GameID: gameID, TagID: tagID}); err != nil {
					return gameID, before, nil, err
				}
			}
		}
		after, err := q.ListTagMapsByGame(ctx, gameID)
		return gameID, before, after, err
	})
}

func (store *gameStore) listGameIDsByTag(ctx context.Context, tagID int64) ([]int64, common.Error) {
	rows, err := store.q.ListGameIDsByTag(ctx, tagID)
	return rows, gameDAOError(err)
}

func (store *gameStore) replaceTagGames(ctx context.Context, meta audit.Meta, tagID int64, gameIDs []int64) common.Error {
	return store.mutate(ctx, meta, "bulk_replace_by_tag", "gfg_tag_map", func(q *gamesqlc.Queries) (int64, any, any, error) {
		before, err := q.ListTagMapsByTag(ctx, tagID)
		if err != nil {
			return tagID, nil, nil, err
		}
		existing := make(map[int64]struct{}, len(before))
		for _, row := range before {
			existing[row.GameID] = struct{}{}
		}
		if _, err = q.DeleteTagMapsByTagExceptGames(ctx, gamesqlc.DeleteTagMapsByTagExceptGamesParams{TagID: tagID, GameIds: gameIDs}); err != nil {
			return tagID, before, nil, err
		}
		missing := make([]int64, 0, len(gameIDs))
		for _, gameID := range gameIDs {
			if _, ok := existing[gameID]; !ok {
				missing = append(missing, gameID)
			}
		}
		if len(missing) > 0 {
			firstID, err := q.NextTagMapID(ctx)
			if err != nil {
				return tagID, before, nil, err
			}
			for index, gameID := range missing {
				if _, err = q.InsertTagMap(ctx, gamesqlc.InsertTagMapParams{ID: firstID + int64(index), GameID: gameID, TagID: tagID}); err != nil {
					return tagID, before, nil, err
				}
			}
		}
		after, err := q.ListTagMapsByTag(ctx, tagID)
		return tagID, before, after, err
	})
}

func (store *gameStore) delete(ctx context.Context, meta audit.Meta, id int64, resource string, beforeFn func(*gamesqlc.Queries) (any, error), deleteFn func(*gamesqlc.Queries) error) common.Error {
	return store.mutate(ctx, meta, "delete", resource, func(q *gamesqlc.Queries) (int64, any, any, error) {
		before, err := beforeFn(q)
		if err == nil {
			err = deleteFn(q)
		}
		return id, before, nil, err
	})
}

func pageArgsGame(page adminutil.PageQuery) (int32, int32) {
	return int32(page.PageSize), int32((page.PageNum - 1) * page.PageSize)
}

func gameDAOError(err error) common.Error {
	if err == nil {
		return nil
	}
	if appErr, ok := err.(common.Error); ok {
		return appErr
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return common.NewDaoError("record not found")
	}
	return common.NewDaoError(err.Error())
}

func localTime(value pgtype.Timestamp) pkgmodels.LocalTime { return pkgmodels.LocalTime(value.Time) }
func gameTimestamp(value time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{Time: value, Valid: !value.IsZero()}
}

func bytesPointer(value []byte) *string {
	if value == nil {
		return nil
	}
	text := string(value)
	return &text
}

type gameFields struct {
	ID, Appid, Weight, PrimaryTag, SecondaryTag      int64
	Name, NameEn, Info, InfoEn, Header               string
	CreateTime, UpdateTime                           pgtype.Timestamp
	Resources, Groups, Developers, Publishers, Links []byte
}

func gameModelFromFields(row gameFields) models.Game {
	return models.Game{ID: row.ID, Name: row.Name, NameEn: row.NameEn, Info: row.Info, InfoEn: row.InfoEn,
		CreateTime: localTime(row.CreateTime), UpdateTime: localTime(row.UpdateTime), Resources: bytesPointer(row.Resources), Groups: bytesPointer(row.Groups),
		Developers: string(row.Developers), Publishers: string(row.Publishers), Appid: row.Appid, Header: row.Header,
		Links: bytesPointer(row.Links), Weight: row.Weight, PrimaryTag: row.PrimaryTag, SecondaryTag: row.SecondaryTag}
}

func listGameModel(row gamesqlc.ListGamesRow) models.Game {
	return gameModelFromFields(gameFields{ID: row.ID, Appid: row.Appid, Weight: row.Weight, PrimaryTag: row.PrimaryTag, SecondaryTag: row.SecondaryTag, Name: row.Name, NameEn: row.NameEn, Info: row.Info, InfoEn: row.InfoEn, Header: row.Header, CreateTime: row.CreateTime, UpdateTime: row.UpdateTime, Resources: row.Resources, Groups: row.Groups, Developers: row.Developers, Publishers: row.Publishers, Links: row.Links})
}

func getGameModel(row gamesqlc.GetGameRow) models.Game {
	return gameModelFromFields(gameFields{ID: row.ID, Appid: row.Appid, Weight: row.Weight, PrimaryTag: row.PrimaryTag, SecondaryTag: row.SecondaryTag, Name: row.Name, NameEn: row.NameEn, Info: row.Info, InfoEn: row.InfoEn, Header: row.Header, CreateTime: row.CreateTime, UpdateTime: row.UpdateTime, Resources: row.Resources, Groups: row.Groups, Developers: row.Developers, Publishers: row.Publishers, Links: row.Links})
}

func insertGameModel(row gamesqlc.InsertGameRow) models.Game {
	return gameModelFromFields(gameFields{ID: row.ID, Appid: row.Appid, Weight: row.Weight, PrimaryTag: row.PrimaryTag, SecondaryTag: row.SecondaryTag, Name: row.Name, NameEn: row.NameEn, Info: row.Info, InfoEn: row.InfoEn, Header: row.Header, CreateTime: row.CreateTime, UpdateTime: row.UpdateTime, Resources: row.Resources, Groups: row.Groups, Developers: row.Developers, Publishers: row.Publishers, Links: row.Links})
}

func updateGameModel(row gamesqlc.UpdateGameRow) models.Game {
	return gameModelFromFields(gameFields{ID: row.ID, Appid: row.Appid, Weight: row.Weight, PrimaryTag: row.PrimaryTag, SecondaryTag: row.SecondaryTag, Name: row.Name, NameEn: row.NameEn, Info: row.Info, InfoEn: row.InfoEn, Header: row.Header, CreateTime: row.CreateTime, UpdateTime: row.UpdateTime, Resources: row.Resources, Groups: row.Groups, Developers: row.Developers, Publishers: row.Publishers, Links: row.Links})
}

func commentModel(row gamesqlc.GfgGameComment) models.GameComment {
	return models.GameComment{ID: row.ID, Region: row.Region, Content: row.Content, Score: row.Score, CreateTime: localTime(row.CreateTime), GameID: row.GameID, IP: row.Ip, Name: row.Name}
}

func prizeModel(row gamesqlc.GfgPrize) models.Prize {
	return models.Prize{ID: row.ID, Title: row.Title, Desc: row.Desc, Prize: string(row.Prize), Key: row.Key, StartTime: localTime(row.StartTime), EndTime: localTime(row.EndTime), CreateTime: localTime(row.CreateTime), Status: row.Status}
}

func tagModel(row gamesqlc.GfgTag) models.Tag {
	return models.Tag{ID: row.ID, Name: row.Name, NameEn: row.NameEn, Info: row.Info, InfoEn: row.InfoEn, Prefix: row.Prefix, CreateTime: localTime(row.CreateTime), UpdateTime: localTime(row.UpdateTime)}
}

func tagMapModel(row gamesqlc.GfgTagMap) models.TagMap {
	return models.TagMap{ID: row.ID, GameID: row.GameID, TagID: row.TagID, CreateTime: localTime(row.CreateTime), UpdateTime: localTime(row.UpdateTime)}
}
