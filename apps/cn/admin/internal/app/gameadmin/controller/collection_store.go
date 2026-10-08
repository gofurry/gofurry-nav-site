package controller

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/gofurry/gofurry-admin/internal/app/gameadmin/models"
	"github.com/gofurry/gofurry-admin/internal/app/shared/adminutil"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	gamesqlc "github.com/gofurry/gofurry-admin/internal/db/game/sqlc"
	"github.com/gofurry/gofurry-admin/pkg/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

const collectionConflict = "此游戏分区已被其他操作修改，请重新加载后重试。"

func collectionError(err error) common.Error {
	if err == nil {
		return nil
	}
	var app common.Error
	if errors.As(err, &app) {
		return app
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return common.NewError(common.RETURN_FAILED, 404, "游戏分区不存在")
	}
	var pg *pgconn.PgError
	if errors.As(err, &pg) && pg.Code == "23505" {
		return common.NewConflictError("游戏分区 Code 已存在")
	}
	return common.NewDaoError("游戏分区服务暂不可用")
}

// Audit is an independent GFA write, not a cross-database ACID transaction.
// A failed Audit prevents GFG commit; no-op callbacks return nil snapshots.
func (s *gameStore) mutateCollection(ctx context.Context, meta audit.Meta, action, resource string, change gameMutation) common.Error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return collectionError(err)
	}
	defer tx.Rollback(ctx)
	q := s.q.WithTx(tx)
	if err = q.LockGameCollectionDomain(ctx); err != nil {
		return collectionError(err)
	}
	id, before, after, err := change(q)
	if err != nil {
		return collectionError(err)
	}
	if before != nil || after != nil {
		var target any = id
		if resource == "gfg_game_collection_home_slot" {
			target = nil // Global placement change, not an individual Collection.
		}
		if err := s.audit.Log(ctx, meta, action, resource, target, before, after); err != nil {
			return common.NewDaoError("操作审计保存失败，游戏分区未保存")
		}
	}
	return collectionError(tx.Commit(ctx))
}

func collectionTime(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	v := t.Time.UTC()
	return &v
}
func collectionDTO(r gamesqlc.GetCuratedCollectionSnapshotsRow) models.GameCollection {
	c := r.GfgGameCollection
	return models.GameCollection{ID: c.ID, Code: c.Code, Name: c.Name, NameEn: c.NameEn, Info: c.Info, InfoEn: c.InfoEn, Status: c.Status, Version: c.Version, PublishedAt: collectionTime(c.PublishedAt), ArchivedAt: collectionTime(c.ArchivedAt), CreatedAt: c.CreatedAt.Time.UTC(), UpdatedAt: c.UpdatedAt.Time.UTC(), MemberCount: r.MemberCount, SFWMemberCount: r.SfwMemberCount, HomeSlot: r.HomeSlot}
}

func (s *gameStore) listCollections(ctx context.Context, page adminutil.PageQuery, status string, eligible bool) (int64, []models.GameCollection, common.Error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return 0, nil, collectionError(err)
	}
	defer tx.Rollback(ctx)
	q := s.q.WithTx(tx)
	total, err := q.CountCuratedCollections(ctx, gamesqlc.CountCuratedCollectionsParams{Keyword: page.Keyword, Status: status, HomeEligible: eligible})
	if err != nil {
		return 0, nil, collectionError(err)
	}
	limit, offset := pageArgsGame(page)
	ids, err := q.ListCuratedCollectionIDs(ctx, gamesqlc.ListCuratedCollectionIDsParams{Keyword: page.Keyword, Status: status, HomeEligible: eligible, RowLimit: limit, RowOffset: offset})
	if err != nil {
		return 0, nil, collectionError(err)
	}
	rows, err := q.GetCuratedCollectionSnapshots(ctx, ids)
	if err != nil {
		return 0, nil, collectionError(err)
	}
	items := make([]models.GameCollection, 0, len(rows))
	for _, row := range rows {
		items = append(items, collectionDTO(row))
	}
	return total, items, collectionError(tx.Commit(ctx))
}
func readCollection(ctx context.Context, q *gamesqlc.Queries, id int64) (models.GameCollection, error) {
	rows, err := q.GetCuratedCollectionSnapshots(ctx, []int64{id})
	if err != nil {
		return models.GameCollection{}, err
	}
	if len(rows) == 0 {
		return models.GameCollection{}, pgx.ErrNoRows
	}
	return collectionDTO(rows[0]), nil
}
func lockCollection(ctx context.Context, q *gamesqlc.Queries, id, version int64) (models.GameCollection, error) {
	if version < 1 {
		return models.GameCollection{}, common.NewValidationError("version 必须为正整数")
	}
	row, err := q.LockCollectionForUpdate(ctx, id)
	if err != nil {
		return models.GameCollection{}, err
	}
	if row.Version != version {
		return models.GameCollection{}, common.NewConflictError(collectionConflict)
	}
	return readCollection(ctx, q, id)
}
func collectionContent(c models.GameCollection) models.CollectionContent {
	return models.CollectionContent{Name: c.Name, NameEn: c.NameEn, Info: c.Info, InfoEn: c.InfoEn}
}
func requireCollectionContent(c models.CollectionContent) error {
	if strings.TrimSpace(c.Name) == "" || strings.TrimSpace(c.NameEn) == "" || strings.TrimSpace(c.Info) == "" || strings.TrimSpace(c.InfoEn) == "" {
		return common.NewValidationError("已发布游戏分区须有完整双语名称与简介")
	}
	return nil
}
func requirePublishable(c models.CollectionContent, count int64) error {
	if err := requireCollectionContent(c); err != nil {
		return err
	}
	if count < 2 {
		return common.NewValidationError("已发布游戏分区须至少有效收录两个游戏")
	}
	return nil
}
func (s *gameStore) createCollection(ctx context.Context, meta audit.Meta, in models.CreateCollection) (models.GameCollection, common.Error) {
	var result models.GameCollection
	err := s.mutateCollection(ctx, meta, "create", "gfg_game_collection", func(q *gamesqlc.Queries) (int64, any, any, error) {
		id, err := q.InsertCuratedCollection(ctx, gamesqlc.InsertCuratedCollectionParams{Code: in.Code, Name: in.Name, NameEn: in.NameEn, Info: in.Info, InfoEn: in.InfoEn})
		if err != nil {
			return 0, nil, nil, err
		}
		result, err = readCollection(ctx, q, id)
		return id, nil, result, err
	})
	return result, err
}
func (s *gameStore) updateCollection(ctx context.Context, meta audit.Meta, id int64, in models.UpdateCollection) (models.GameCollection, common.Error) {
	var result models.GameCollection
	err := s.mutateCollection(ctx, meta, "update", "gfg_game_collection", func(q *gamesqlc.Queries) (int64, any, any, error) {
		before, err := lockCollection(ctx, q, id, in.Version)
		if err != nil {
			return id, nil, nil, err
		}
		if before.Status == "archived" {
			return id, nil, nil, common.NewConflictError("已归档游戏分区只允许恢复")
		}
		if before.Status == "published" {
			if err = requireCollectionContent(in.CollectionContent); err != nil {
				return id, nil, nil, err
			}
		}
		if collectionContent(before) == in.CollectionContent {
			result = before
			return id, nil, nil, nil
		}
		err = q.UpdateCuratedCollectionContent(ctx, gamesqlc.UpdateCuratedCollectionContentParams{ID: id, Name: in.Name, NameEn: in.NameEn, Info: in.Info, InfoEn: in.InfoEn})
		if err != nil {
			return id, nil, nil, err
		}
		result, err = readCollection(ctx, q, id)
		return id, before, result, err
	})
	return result, err
}
func readCollectionMembers(ctx context.Context, q *gamesqlc.Queries, id int64) ([]models.CollectionMember, error) {
	rows, err := q.GetCuratedCollectionMembers(ctx, id)
	if err != nil {
		return nil, err
	}
	members := make([]models.CollectionMember, 0, len(rows))
	for _, r := range rows {
		members = append(members, models.CollectionMember{GameID: r.GameID, Name: r.Name, NameEn: r.NameEn, AppID: r.Appid, Adult: r.Adult})
	}
	return members, nil
}
func (s *gameStore) collectionMembers(ctx context.Context, id int64) (models.CollectionMembers, common.Error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return models.CollectionMembers{}, collectionError(err)
	}
	defer tx.Rollback(ctx)
	q := s.q.WithTx(tx)
	c, err := readCollection(ctx, q, id)
	if err != nil {
		return models.CollectionMembers{}, collectionError(err)
	}
	members, err := readCollectionMembers(ctx, q, id)
	if err != nil {
		return models.CollectionMembers{}, collectionError(err)
	}
	return models.CollectionMembers{CollectionID: id, Version: c.Version, Members: members}, collectionError(tx.Commit(ctx))
}

type collectionMemberAudit struct {
	models.GameCollection
	GameIDs []int64 `json:"game_ids"`
}

func (s *gameStore) replaceCollectionMembers(ctx context.Context, meta audit.Meta, id int64, in models.ReplaceCollectionMembers) (models.GameCollection, common.Error) {
	var result models.GameCollection
	err := s.mutateCollection(ctx, meta, "members_update", "gfg_game_collection", func(q *gamesqlc.Queries) (int64, any, any, error) {
		before, err := lockCollection(ctx, q, id, in.Version)
		if err != nil {
			return id, nil, nil, err
		}
		if before.Status == "archived" {
			return id, nil, nil, common.NewConflictError("已归档游戏分区只允许恢复")
		}
		excluded, err := q.GetCollectionExcludedIDs(ctx, id)
		if err != nil {
			return id, nil, nil, err
		}
		if err = disjointCollectionOverrides(in.GameIDs, excluded); err != nil {
			return id, nil, nil, err
		}
		existing, err := q.ExistingCuratedCollectionGameIDs(ctx, in.GameIDs)
		if err != nil {
			return id, nil, nil, err
		}
		if !slices.Equal(existing, in.GameIDs) {
			return id, nil, nil, common.NewValidationError("部分游戏不存在，请重新选择")
		}
		old, err := readCollectionMembers(ctx, q, id)
		if err != nil {
			return id, nil, nil, err
		}
		oldIDs := make([]int64, 0, len(old))
		for _, m := range old {
			oldIDs = append(oldIDs, m.GameID)
		}
		if slices.Equal(oldIDs, in.GameIDs) {
			result = before
			return id, nil, nil, nil
		}
		if err = q.DeleteRemovedCollectionMembers(ctx, gamesqlc.DeleteRemovedCollectionMembersParams{CollectionID: id, GameIds: in.GameIDs}); err != nil {
			return id, nil, nil, err
		}
		if err = q.InsertCollectionMembers(ctx, gamesqlc.InsertCollectionMembersParams{CollectionID: id, GameIds: in.GameIDs}); err != nil {
			return id, nil, nil, err
		}
		if err = q.BumpCuratedCollectionVersion(ctx, id); err != nil {
			return id, nil, nil, err
		}
		result, err = readCollection(ctx, q, id)
		if err != nil {
			return id, nil, nil, err
		}
		if result.Status == "published" {
			if err = requirePublishable(collectionContent(result), result.MemberCount); err != nil {
				return id, nil, nil, err
			}
		}
		if result.SFWMemberCount == 0 && result.HomeSlot != nil {
			if err = q.RemoveCollectionHomeSlot(ctx, id); err != nil {
				return id, nil, nil, err
			}
			result.HomeSlot = nil
		}
		return id, collectionMemberAudit{before, oldIDs}, collectionMemberAudit{result, in.GameIDs}, nil
	})
	return result, err
}
func collectionTransition(status, action string) (string, error) {
	switch {
	case status == "draft" && action == "publish":
		return "published", nil
	case status == "published" && action == "unpublish":
		return "draft", nil
	case (status == "draft" || status == "published") && action == "archive":
		return "archived", nil
	case status == "archived" && action == "restore":
		return "draft", nil
	default:
		return "", common.NewConflictError("当前游戏分区状态不支持此操作")
	}
}
func (s *gameStore) transitionCollection(ctx context.Context, meta audit.Meta, id, version int64, action string) (models.GameCollection, common.Error) {
	var result models.GameCollection
	err := s.mutateCollection(ctx, meta, action, "gfg_game_collection", func(q *gamesqlc.Queries) (int64, any, any, error) {
		before, err := lockCollection(ctx, q, id, version)
		if err != nil {
			return id, nil, nil, err
		}
		status, err := collectionTransition(before.Status, action)
		if err != nil {
			return id, nil, nil, err
		}
		if action == "publish" {
			if err = requirePublishable(collectionContent(before), before.MemberCount); err != nil {
				return id, nil, nil, err
			}
		}
		if err = q.TransitionCuratedCollection(ctx, gamesqlc.TransitionCuratedCollectionParams{ID: id, Status: status}); err != nil {
			return id, nil, nil, err
		}
		if action == "archive" || action == "unpublish" {
			if err = q.RemoveCollectionHomeSlot(ctx, id); err != nil {
				return id, nil, nil, err
			}
		}
		result, err = readCollection(ctx, q, id)
		return id, before, result, err
	})
	return result, err
}

func collectionRevision(placements [5]int64) string {
	b, _ := json.Marshal(placements)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
func readCollectionHome(ctx context.Context, q *gamesqlc.Queries) (models.CollectionHome, error) {
	rows, err := q.GetCollectionHomePlacements(ctx)
	if err != nil {
		return models.CollectionHome{}, err
	}
	var placements [5]int64
	ids := make([]int64, 0, len(rows))
	for _, r := range rows {
		placements[r.Slot-1] = r.CollectionID
		ids = append(ids, r.CollectionID)
	}
	collections, err := q.GetCuratedCollectionSnapshots(ctx, ids)
	if err != nil {
		return models.CollectionHome{}, err
	}
	byID := map[int64]models.GameCollection{}
	for _, r := range collections {
		byID[r.GfgGameCollection.ID] = collectionDTO(r)
	}
	home := models.CollectionHome{Revision: collectionRevision(placements), Slots: make([]models.CollectionHomeSlot, 5)}
	for i, id := range placements {
		home.Slots[i].Slot = int16(i + 1)
		if c, ok := byID[id]; ok {
			home.Slots[i].Collection = &c
		}
	}
	return home, nil
}
func (s *gameStore) collectionHome(ctx context.Context) (models.CollectionHome, common.Error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return models.CollectionHome{}, collectionError(err)
	}
	defer tx.Rollback(ctx)
	home, err := readCollectionHome(ctx, s.q.WithTx(tx))
	if err != nil {
		return home, collectionError(err)
	}
	return home, collectionError(tx.Commit(ctx))
}
func homeAudit(home models.CollectionHome) []models.CollectionPlacement {
	slots := make([]models.CollectionPlacement, 5)
	for i, s := range home.Slots {
		slots[i].Slot = s.Slot
		if s.Collection != nil {
			id := s.Collection.ID
			slots[i].CollectionID = &id
		}
	}
	return slots
}
func (s *gameStore) replaceCollectionHome(ctx context.Context, meta audit.Meta, in models.ReplaceCollectionHome) (models.CollectionHome, common.Error) {
	var result models.CollectionHome
	err := s.mutateCollection(ctx, meta, "home_curation_update", "gfg_game_collection_home_slot", func(q *gamesqlc.Queries) (int64, any, any, error) {
		before, err := readCollectionHome(ctx, q)
		if err != nil {
			return 0, nil, nil, err
		}
		if before.Revision != in.Revision {
			return 0, nil, nil, common.NewConflictError("首页入口编排已被其他操作修改，请重新加载后重试。")
		}
		var placements [5]int64
		ids := []int64{}
		slots := []int16{}
		for _, s := range in.Slots {
			if s.CollectionID != nil {
				placements[s.Slot-1] = *s.CollectionID
				ids = append(ids, *s.CollectionID)
				slots = append(slots, s.Slot)
			}
		}
		rows, err := q.GetCuratedCollectionSnapshots(ctx, ids)
		if err != nil {
			return 0, nil, nil, err
		}
		if len(rows) != len(ids) {
			return 0, nil, nil, common.NewValidationError("所选游戏分区不存在")
		}
		unchanged := map[int64]bool{}
		for i, slot := range before.Slots {
			if slot.Collection != nil && placements[i] == slot.Collection.ID {
				unchanged[slot.Collection.ID] = true
			}
		}
		for _, r := range rows {
			if !unchanged[r.GfgGameCollection.ID] && (r.GfgGameCollection.Status != "published" || r.SfwMemberCount == 0) {
				return 0, nil, nil, common.NewValidationError("首页仅可选择已发布且有 SFW 可见游戏的分区")
			}
		}
		if before.Revision == collectionRevision(placements) {
			result = before
			return 0, nil, nil, nil
		}
		if err = q.ClearCollectionHomePlacements(ctx); err != nil {
			return 0, nil, nil, err
		}
		if err = q.InsertCollectionHomePlacements(ctx, gamesqlc.InsertCollectionHomePlacementsParams{Slots: slots, CollectionIds: ids}); err != nil {
			return 0, nil, nil, err
		}
		result, err = readCollectionHome(ctx, q)
		return 0, homeAudit(before), homeAudit(result), err
	})
	return result, err
}
