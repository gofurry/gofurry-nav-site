package controller

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"slices"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/gameadmin/models"
	"github.com/gofurry/gofurry-admin/internal/app/shared/adminutil"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	gamesqlc "github.com/gofurry/gofurry-admin/internal/db/game/sqlc"
	"github.com/gofurry/gofurry-admin/pkg/common"
	"github.com/jackc/pgx/v5"
)

func normalizeCollectionComposition(in *models.ReplaceCollectionComposition) common.Error {
	if in.TagIDs == nil || in.ManualGameIDs == nil || in.ExcludedGameIDs == nil {
		return common.NewValidationError("tag_ids、manual_game_ids、excluded_game_ids 须为完整数组")
	}
	for _, ids := range []*[]int64{&in.TagIDs, &in.ManualGameIDs, &in.ExcludedGameIDs} {
		canonical, err := canonicalCollectionIDs(*ids)
		if err != nil {
			return err
		}
		*ids = canonical
	}
	return disjointCollectionOverrides(in.ManualGameIDs, in.ExcludedGameIDs)
}
func disjointCollectionOverrides(manual, excluded []int64) common.Error {
	for _, id := range manual {
		if _, found := slices.BinarySearch(excluded, id); found {
			return common.NewValidationError("人工固定与排除不可包含同一游戏")
		}
	}
	return nil
}
func (api *GameAPI) GetGameCollectionComposition(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	result, err := api.store.collectionComposition(c.Context(), id)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(result)
}
func (api *GameAPI) ReplaceGameCollectionComposition(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var in models.ReplaceCollectionComposition
	if err = decodeCollectionBody(c, &in); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if err = normalizeCollectionComposition(&in); err != nil {
		return common.NewResponse(c).Error(err)
	}
	result, err := api.store.replaceCollectionComposition(c.Context(), audit.MetaFromFiber(c), id, in)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(result)
}
func readCollectionComposition(ctx context.Context, q *gamesqlc.Queries, c models.GameCollection) (models.CollectionComposition, error) {
	result := models.CollectionComposition{CollectionID: c.ID, Version: c.Version, HomeSlot: c.HomeSlot,
		RuleTags: []models.CollectionRuleTag{}, ManualMembers: []models.CollectionMember{}, ExcludedMembers: []models.CollectionMember{}, EffectiveMembers: []models.EffectiveCollectionMember{}}
	tags, err := q.GetCollectionRuleTags(ctx, c.ID)
	if err != nil {
		return result, err
	}
	for _, tag := range tags {
		result.RuleTags = append(result.RuleTags, models.CollectionRuleTag{TagID: tag.TagID, Code: tag.Code, Name: tag.Name, NameEn: tag.NameEn, Active: tag.Active})
	}
	rows, err := q.GetCollectionCompositionMembers(ctx, c.ID)
	if err != nil {
		return result, err
	}
	for _, r := range rows {
		member := models.CollectionMember{GameID: r.GameID, Name: r.Name, NameEn: r.NameEn, AppID: r.Appid, Adult: r.Adult}
		if r.Automatic {
			result.Counts.AutoMatched++
		}
		if r.Manual {
			result.ManualMembers = append(result.ManualMembers, member)
			result.Counts.ManualPinned++
		}
		if r.Excluded {
			result.ExcludedMembers = append(result.ExcludedMembers, member)
			result.Counts.Excluded++
			continue
		}
		if !r.Automatic && !r.Manual {
			continue
		}
		source := "automatic"
		if r.Manual {
			source = "manual"
			if r.Automatic {
				source = "both"
			}
		}
		result.EffectiveMembers = append(result.EffectiveMembers, models.EffectiveCollectionMember{CollectionMember: member, Source: source})
		result.Counts.Effective++
		if !r.Adult {
			result.Counts.SFWVisible++
		}
	}
	return result, nil
}
func (s *gameStore) collectionComposition(ctx context.Context, id int64) (models.CollectionComposition, common.Error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return models.CollectionComposition{}, collectionError(err)
	}
	defer tx.Rollback(ctx)
	q := s.q.WithTx(tx)
	c, err := readCollection(ctx, q, id)
	if err != nil {
		return models.CollectionComposition{}, collectionError(err)
	}
	result, err := readCollectionComposition(ctx, q, c)
	if err != nil {
		return result, collectionError(err)
	}
	return result, collectionError(tx.Commit(ctx))
}
func compositionConfig(c models.CollectionComposition) models.ReplaceCollectionComposition {
	result := models.ReplaceCollectionComposition{Version: c.Version, TagIDs: []int64{}, ManualGameIDs: []int64{}, ExcludedGameIDs: []int64{}}
	for _, tag := range c.RuleTags {
		result.TagIDs = append(result.TagIDs, tag.TagID)
	}
	for _, member := range c.ManualMembers {
		result.ManualGameIDs = append(result.ManualGameIDs, member.GameID)
	}
	for _, member := range c.ExcludedMembers {
		result.ExcludedGameIDs = append(result.ExcludedGameIDs, member.GameID)
	}
	return result
}

// Audit size follows the explicit configuration, never the derived member list.
// Large configurations retain canonical prefixes plus a digest of all IDs.
type collectionCompositionAudit struct {
	Version             int64                              `json:"version"`
	TagIDs              []int64                            `json:"tag_ids"`
	ManualGameIDs       []int64                            `json:"manual_game_ids"`
	ExcludedGameIDs     []int64                            `json:"excluded_game_ids"`
	ConfigurationSHA256 string                             `json:"configuration_sha256"`
	IDsTruncated        bool                               `json:"ids_truncated"`
	Counts              models.CollectionCompositionCounts `json:"counts"`
	HomeSlot            *int16                             `json:"home_slot"`
}

func compositionAudit(c models.CollectionComposition) collectionCompositionAudit {
	config := compositionConfig(c)
	body, _ := json.Marshal([][]int64{config.TagIDs, config.ManualGameIDs, config.ExcludedGameIDs})
	hash := sha256.Sum256(body)
	const limit = 256
	return collectionCompositionAudit{Version: c.Version, TagIDs: config.TagIDs[:min(len(config.TagIDs), limit)], ManualGameIDs: config.ManualGameIDs[:min(len(config.ManualGameIDs), limit)], ExcludedGameIDs: config.ExcludedGameIDs[:min(len(config.ExcludedGameIDs), limit)], ConfigurationSHA256: hex.EncodeToString(hash[:]), IDsTruncated: len(config.TagIDs) > limit || len(config.ManualGameIDs) > limit || len(config.ExcludedGameIDs) > limit, Counts: c.Counts, HomeSlot: c.HomeSlot}
}
func (s *gameStore) replaceCollectionComposition(ctx context.Context, meta audit.Meta, id int64, in models.ReplaceCollectionComposition) (models.CollectionComposition, common.Error) {
	var result models.CollectionComposition
	err := s.mutateCollection(ctx, meta, "composition_update", "gfg_game_collection", func(q *gamesqlc.Queries) (int64, any, any, error) {
		c, err := lockCollection(ctx, q, id, in.Version)
		if err != nil {
			return id, nil, nil, err
		}
		if c.Status == "archived" {
			return id, nil, nil, common.NewConflictError("已归档游戏分区只允许恢复")
		}
		before, err := readCollectionComposition(ctx, q, c)
		if err != nil {
			return id, nil, nil, err
		}
		old := compositionConfig(before)
		if slices.Equal(old.TagIDs, in.TagIDs) && slices.Equal(old.ManualGameIDs, in.ManualGameIDs) && slices.Equal(old.ExcludedGameIDs, in.ExcludedGameIDs) {
			result = before
			return id, nil, nil, nil
		}
		tags, err := q.ValidateCollectionRuleTags(ctx, gamesqlc.ValidateCollectionRuleTagsParams{CollectionID: id, TagIds: in.TagIDs})
		if err != nil {
			return id, nil, nil, err
		}
		if len(tags) != len(in.TagIDs) {
			return id, nil, nil, common.NewValidationError("部分标签不存在，请重新选择")
		}
		for _, tag := range tags {
			if !tag.Active && !tag.AlreadyBound {
				return id, nil, nil, common.NewValidationError("新增自动规则仅可绑定有效标签与类别")
			}
		}
		ids, err := canonicalCollectionIDs(append(append([]int64{}, in.ManualGameIDs...), in.ExcludedGameIDs...))
		if err != nil {
			return id, nil, nil, err
		}
		existing, err := q.ExistingCuratedCollectionGameIDs(ctx, ids)
		if err != nil {
			return id, nil, nil, err
		}
		if !slices.Equal(existing, ids) {
			return id, nil, nil, common.NewValidationError("部分游戏不存在，请重新选择")
		}
		if err = q.DeleteRemovedCollectionRules(ctx, gamesqlc.DeleteRemovedCollectionRulesParams{CollectionID: id, TagIds: in.TagIDs}); err != nil {
			return id, nil, nil, err
		}
		if err = q.InsertCollectionRules(ctx, gamesqlc.InsertCollectionRulesParams{CollectionID: id, TagIds: in.TagIDs}); err != nil {
			return id, nil, nil, err
		}
		if err = q.DeleteRemovedCollectionMembers(ctx, gamesqlc.DeleteRemovedCollectionMembersParams{CollectionID: id, GameIds: in.ManualGameIDs}); err != nil {
			return id, nil, nil, err
		}
		if err = q.InsertCollectionMembers(ctx, gamesqlc.InsertCollectionMembersParams{CollectionID: id, GameIds: in.ManualGameIDs}); err != nil {
			return id, nil, nil, err
		}
		if err = q.DeleteRemovedCollectionExclusions(ctx, gamesqlc.DeleteRemovedCollectionExclusionsParams{CollectionID: id, GameIds: in.ExcludedGameIDs}); err != nil {
			return id, nil, nil, err
		}
		if err = q.InsertCollectionExclusions(ctx, gamesqlc.InsertCollectionExclusionsParams{CollectionID: id, GameIds: in.ExcludedGameIDs}); err != nil {
			return id, nil, nil, err
		}
		result, err = readCollectionComposition(ctx, q, c)
		if err != nil {
			return id, nil, nil, err
		}
		if c.Status == "published" {
			if err = requirePublishable(collectionContent(c), result.Counts.Effective); err != nil {
				return id, nil, nil, err
			}
		}
		if result.Counts.SFWVisible == 0 && c.HomeSlot != nil {
			if err = q.RemoveCollectionHomeSlot(ctx, id); err != nil {
				return id, nil, nil, err
			}
			result.HomeSlot = nil
		}
		if err = q.BumpCuratedCollectionVersion(ctx, id); err != nil {
			return id, nil, nil, err
		}
		result.Version++
		return id, compositionAudit(before), compositionAudit(result), nil
	})
	return result, err
}
