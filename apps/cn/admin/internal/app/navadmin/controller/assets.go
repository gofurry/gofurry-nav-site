package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/navadmin/models"
	"github.com/gofurry/gofurry-admin/internal/app/shared/adminutil"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	navsqlc "github.com/gofurry/gofurry-admin/internal/db/nav/sqlc"
	"github.com/gofurry/gofurry-admin/internal/infra/assets"
	"github.com/gofurry/gofurry-admin/pkg/common"
)

// WithAssets wires explicit runtime configuration; no provider credentials enter DTOs.
func (api *navAPI) WithAssets(storage *assets.Service, primaryBase, mirrorBase string) *NavAPI {
	api.assetStorage, api.assetPrimaryBase, api.assetMirrorBase = storage, primaryBase, mirrorBase
	return api
}

type assetLinks struct {
	PrimaryURL string `json:"primary_url"`
	MirrorURL  string `json:"mirror_url"`
}

func (api *navAPI) siteIconLinks(site *models.SiteDTO) {
	if site.Icon != nil {
		links := api.links(*site.Icon)
		site.IconPrimaryURL, site.IconMirrorURL = links.PrimaryURL, links.MirrorURL
	}
}

type heroPayload struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type heroDTO struct {
	ID      int64  `json:"id,string"`
	Variant string `json:"variant"`
	heroPayload
	ObjectKey  string    `json:"object_key"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
	assetLinks
}

type publicationDTO struct {
	assets.Publication
	Item any `json:"item"`
}

func (api *navAPI) links(key string) assetLinks {
	return assetLinks{PrimaryURL: assets.URL(api.assetPrimaryBase, key), MirrorURL: assets.URL(api.assetMirrorBase, key)}
}

func (api *navAPI) hero(row navsqlc.GfnHomeHeroAsset) heroDTO {
	return heroDTO{ID: row.ID, Variant: row.Variant, heroPayload: heroPayload{row.Name, row.Enabled}, ObjectKey: row.ObjectKey, CreateTime: row.CreateTime.Time, UpdateTime: row.UpdateTime.Time, assetLinks: api.links(row.ObjectKey)}
}

func decodeAssetJSON(data []byte, target any) common.Error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return common.NewValidationError("invalid asset metadata or unsupported field")
	}
	if decoder.Decode(new(any)) != io.EOF {
		return common.NewValidationError("metadata must contain one JSON object")
	}
	return nil
}

func validateAssetName(name string) bool {
	return strings.TrimSpace(name) != "" && utf8.RuneCountInString(name) <= 120
}

func uploadObject(c fiber.Ctx, kind string, siteID int64) (assets.Object, common.Error) {
	file, err := c.FormFile("file")
	if err != nil {
		return assets.Object{}, common.NewValidationError("file is required")
	}
	if file.Size > assets.MaxSize {
		return assets.Object{}, common.NewValidationError("asset exceeds upload size limit")
	}
	reader, err := file.Open()
	if err != nil {
		return assets.Object{}, common.NewValidationError("cannot read uploaded file")
	}
	defer reader.Close()
	data, err := io.ReadAll(io.LimitReader(reader, assets.MaxSize+1))
	if err != nil {
		return assets.Object{}, common.NewValidationError("cannot read uploaded file")
	}
	object, err := assets.NewObject(kind, siteID, file.Filename, data)
	if err != nil {
		return assets.Object{}, common.NewValidationError(err.Error())
	}
	return object, nil
}

func (api *navAPI) publish(c fiber.Ctx, object assets.Object) (assets.Publication, common.Error) {
	if api.assetStorage == nil {
		return assets.Publication{}, common.NewError(common.RETURN_FAILED, 503, "asset storage is not configured")
	}
	result, err := api.assetStorage.Publish(c.Context(), object)
	if err != nil {
		return result, common.NewError(common.RETURN_FAILED, 502, "COS publication failed; no database change was made")
	}
	return result, nil
}

func (api *navAPI) ReplaceSiteIcon(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	if _, err := api.store.getSite(c.Context(), id); err != nil {
		return common.NewResponse(c).Error(err)
	}
	object, err := uploadObject(c, "site-icon", id)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	publication, err := api.publish(c, object)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	err = api.store.mutate(c.Context(), audit.MetaFromFiber(c), "site.icon.replace", "gfn_site", func(q *navsqlc.Queries) (int64, any, any, error) {
		before, e := q.LockSiteIcon(c.Context(), id)
		if e != nil {
			return id, nil, nil, e
		}
		_, e = q.ReplaceSiteIcon(c.Context(), navsqlc.ReplaceSiteIconParams{ID: id, Icon: &object.Key})
		return id, before, map[string]any{"icon": object.Key}, e
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	invalidateNavSiteListCache()
	return common.NewResponse(c).SuccessWithData(publicationDTO{publication, api.links(object.Key)})
}

func (api *navAPI) ClearSiteIcon(c fiber.Ctx) error {
	id, err := adminutil.ParseIDParam(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	err = api.store.mutate(c.Context(), audit.MetaFromFiber(c), "site.icon.clear", "gfn_site", func(q *navsqlc.Queries) (int64, any, any, error) {
		before, e := q.LockSiteIcon(c.Context(), id)
		if e != nil {
			return id, nil, nil, e
		}
		_, e = q.ReplaceSiteIcon(c.Context(), navsqlc.ReplaceSiteIconParams{ID: id})
		return id, before, map[string]any{"icon": nil}, e
	})
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	invalidateNavSiteListCache()
	return common.NewResponse(c).SuccessWithData(map[string]any{"icon": nil})
}

func assetUpdateAction(kind string, before, after bool) string {
	if before == after {
		return kind + ".update"
	}
	if after {
		return kind + ".enable"
	}
	return kind + ".disable"
}

func (store *navStore) mutateAsset(ctx context.Context, meta audit.Meta, action, resource string, change func(*navsqlc.Queries) (int64, any, any, string, error)) common.Error {
	tx, err := store.pool.Begin(ctx)
	if err != nil {
		return navDAOError(err)
	}
	defer tx.Rollback(ctx)
	id, before, after, verb, err := change(store.q.WithTx(tx))
	if err != nil {
		return navDAOError(err)
	}
	if verb != "" {
		action = verb
	}
	if err := store.audit.Log(ctx, meta, action, resource, id, before, after); err != nil {
		return err
	}
	return navDAOError(tx.Commit(ctx))
}
