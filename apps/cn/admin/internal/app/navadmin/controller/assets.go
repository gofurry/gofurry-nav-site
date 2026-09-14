package controller

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"regexp"
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

type patternPayload struct {
	Name          string  `json:"name"`
	NameEn        string  `json:"name_en"`
	LightColor    string  `json:"light_color"`
	DarkColor     string  `json:"dark_color"`
	LightOpacity  float64 `json:"light_opacity"`
	DarkOpacity   float64 `json:"dark_opacity"`
	DefaultSizePx int32   `json:"default_size_px"`
	Enabled       bool    `json:"enabled"`
	SortOrder     int64   `json:"sort_order"`
}

type patternDTO struct {
	ID int64 `json:"id,string"`
	patternPayload
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

func (api *navAPI) pattern(row navsqlc.GetBackgroundPatternRow) patternDTO {
	return patternDTO{ID: row.ID, patternPayload: patternPayload{row.Name, row.NameEn, row.LightColor, row.DarkColor, row.LightOpacity, row.DarkOpacity, row.DefaultSizePx, row.Enabled, row.SortOrder}, ObjectKey: row.ObjectKey, CreateTime: row.CreateTime.Time, UpdateTime: row.UpdateTime.Time, assetLinks: api.links(row.ObjectKey)}
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

var assetColor = regexp.MustCompile(`^#[a-fA-F0-9]{6}$`)

func validatePattern(req patternPayload) common.Error {
	if !validateAssetName(req.Name) || !validateAssetName(req.NameEn) {
		return common.NewValidationError("name and name_en are required (at most 120 characters)")
	}
	if !assetColor.MatchString(req.LightColor) || !assetColor.MatchString(req.DarkColor) {
		return common.NewValidationError("colors must use #RRGGBB")
	}
	for _, value := range []float64{req.LightOpacity, req.DarkOpacity} {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 1 {
			return common.NewValidationError("opacity must be between 0 and 1")
		}
	}
	if req.DefaultSizePx <= 0 {
		return common.NewValidationError("default_size_px must be positive")
	}
	return nil
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
