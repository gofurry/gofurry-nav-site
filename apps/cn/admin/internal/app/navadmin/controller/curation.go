package controller

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/shared/adminutil"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	navsqlc "github.com/gofurry/gofurry-admin/internal/db/nav/sqlc"
	"github.com/gofurry/gofurry-admin/pkg/common"
)

type curationSite struct {
	SiteID  int64  `json:"site_id,string"`
	Name    string `json:"name"`
	Deleted bool   `json:"deleted"`
}

type groupCuration struct {
	Name     string         `json:"name"`
	Revision string         `json:"revision"`
	Sites    []curationSite `json:"sites"`
}

type curationOrder struct {
	Revision string   `json:"revision"`
	SiteIDs  []string `json:"site_ids"`
}

func curationRevision(rows []navsqlc.ListGroupCurationRow) string {
	data, _ := json.Marshal(rows)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func validateCurationOrder(rows []navsqlc.ListGroupCurationRow, order curationOrder) ([]int64, error) {
	if order.Revision == "" || order.Revision != curationRevision(rows) {
		return nil, common.NewValidationError("分组内容或排序已变化，请重新加载后编排")
	}
	if len(order.SiteIDs) != len(rows) {
		return nil, common.NewValidationError("排序必须包含该分组的所有站点")
	}
	members := make(map[int64]bool, len(rows))
	for _, row := range rows {
		members[row.SiteID] = true
	}
	ids := make([]int64, 0, len(rows))
	for _, value := range order.SiteIDs {
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil || !members[id] {
			return nil, common.NewValidationError("排序包含重复或不属于该分组的站点")
		}
		delete(members, id)
		ids = append(ids, id)
	}
	return ids, nil
}

func (api *navAPI) GetGroupCuration(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	group, err := api.store.getSiteGroup(c.Context(), id)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	rows, queryErr := api.store.q.ListGroupCuration(c.Context(), id)
	if queryErr != nil {
		return common.NewResponse(c).Error(navDAOError(queryErr))
	}
	result := groupCuration{Name: group.Name, Revision: curationRevision(rows), Sites: make([]curationSite, 0, len(rows))}
	for _, row := range rows {
		result.Sites = append(result.Sites, curationSite{SiteID: row.SiteID, Name: row.Name, Deleted: row.Deleted})
	}
	return common.NewResponse(c).SuccessWithData(result)
}

func (store *navStore) reorderGroup(ctx context.Context, meta audit.Meta, id int64, order curationOrder) common.Error {
	return store.mutate(ctx, meta, "reorder", "gfn_site_group_map", func(q *navsqlc.Queries) (int64, any, any, error) {
		if err := q.LockSiteGroupCuration(ctx); err != nil {
			return id, nil, nil, err
		}
		if _, err := q.GetSiteGroup(ctx, id); err != nil {
			return id, nil, nil, err
		}
		before, err := q.ListGroupCuration(ctx, id)
		if err != nil {
			return id, nil, nil, err
		}
		ids, err := validateCurationOrder(before, order)
		if err != nil {
			return id, nil, nil, err
		}
		for index, siteID := range ids {
			count, err := q.UpdateGroupCurationWeight(ctx, navsqlc.UpdateGroupCurationWeightParams{GroupID: id, SiteID: siteID, Weight: int64(len(ids) - index)})
			if err != nil {
				return id, nil, nil, err
			}
			if count != 1 {
				return id, nil, nil, common.NewValidationError("分组关系已变化，请重新加载")
			}
		}
		after, err := q.ListGroupCuration(ctx, id)
		return id, before, after, err
	})
}

func (api *navAPI) ReorderGroupCuration(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	var order curationOrder
	if err := adminutil.DecodeBody(c, &order); err != nil {
		return common.NewResponse(c).Error(err)
	}
	if err := api.store.reorderGroup(c.Context(), audit.MetaFromFiber(c), id, order); err != nil {
		return common.NewResponse(c).Error(err)
	}
	invalidateNavGroupMapCache()
	return api.GetGroupCuration(c)
}
