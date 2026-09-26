package collaboration

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/shared/adminutil"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	"github.com/gofurry/gofurry-admin/pkg/common"
)

type API struct{ service *Service }

func NewAPI(service *Service) *API { return &API{service} }
func response(c fiber.Ctx, data any, err common.Error) error {
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(data)
}

// Reject client-supplied actors/status/link fields instead of silently accepting
// them. Dedicated transitions own state changes.
func decode(c fiber.Ctx, target any) common.Error {
	decoder := json.NewDecoder(bytes.NewReader(c.Body()))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return common.NewValidationError("请求字段无效")
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return common.NewValidationError("request body must contain one JSON object")
	}
	return nil
}
func (api *API) List(c fiber.Ctx) error {
	meta := audit.MetaFromFiber(c)
	account, e := actor(meta)
	if e != nil {
		return response(c, nil, e)
	}
	page, _ := strconv.Atoi(c.Query("page_num", "1"))
	size, _ := strconv.Atoi(c.Query("page_size", "50"))
	data, e := api.service.List(c.Context(), Filters{Page: page, PageSize: size, Kind: c.Query("kind"), Status: c.Query("status"), Priority: c.Query("priority"), Researcher: c.Query("researcher"), Keyword: c.Query("keyword")}, account)
	return response(c, data, e)
}
func (api *API) Get(c fiber.Ctx) error {
	id, e := adminutil.ParseIDParam(c)
	if e != nil {
		return response(c, nil, e)
	}
	data, e := api.service.Get(c.Context(), id)
	return response(c, data, e)
}
func (api *API) Create(c fiber.Ctx) error {
	var in Input
	if e := decode(c, &in); e != nil {
		return response(c, nil, e)
	}
	data, e := api.service.Create(c.Context(), audit.MetaFromFiber(c), in)
	return response(c, data, e)
}
func (api *API) Update(c fiber.Ctx) error {
	var in Input
	if e := decode(c, &in); e != nil {
		return response(c, nil, e)
	}
	id, e := adminutil.ParseIDParam(c)
	if e != nil {
		return response(c, nil, e)
	}
	data, e := api.service.Update(c.Context(), audit.MetaFromFiber(c), id, in)
	return response(c, data, e)
}
func (api *API) Preview(c fiber.Ctx) error {
	var in BatchInput
	if e := decode(c, &in); e != nil {
		return response(c, nil, e)
	}
	data, e := api.service.Preview(c.Context(), in.Items)
	return response(c, data, e)
}

func (api *API) Delete(c fiber.Ctx) error {
	var in struct {
		Version int64 `json:"version"`
	}
	if e := decode(c, &in); e != nil {
		return response(c, nil, e)
	}
	id, e := adminutil.ParseIDParam(c)
	if e != nil {
		return response(c, nil, e)
	}
	return response(c, nil, api.service.Delete(c.Context(), audit.MetaFromFiber(c), id, in.Version))
}
func (api *API) Batch(c fiber.Ctx) error {
	var in BatchInput
	if e := decode(c, &in); e != nil {
		return response(c, nil, e)
	}
	data, e := api.service.Batch(c.Context(), audit.MetaFromFiber(c), in)
	return response(c, data, e)
}
func (api *API) Transition(action string) fiber.Handler {
	return func(c fiber.Ctx) error {
		var in Transition
		if e := decode(c, &in); e != nil {
			return response(c, nil, e)
		}
		id, e := adminutil.ParseIDParam(c)
		if e != nil {
			return response(c, nil, e)
		}
		data, e := api.service.Transition(c.Context(), audit.MetaFromFiber(c), id, action, in)
		return response(c, data, e)
	}
}
func (api *API) Board(c fiber.Ctx) error {
	data, e := api.service.Board(c.Context())
	return response(c, data, e)
}
func (api *API) CreateBoard(c fiber.Ctx) error { return api.saveBoard(c, 0) }
func (api *API) UpdateBoard(c fiber.Ctx) error {
	id, e := adminutil.ParseIDParam(c)
	if e != nil {
		return response(c, nil, e)
	}
	return api.saveBoard(c, id)
}
func (api *API) saveBoard(c fiber.Ctx, id int64) error {
	var in BoardInput
	if e := decode(c, &in); e != nil {
		return response(c, nil, e)
	}
	data, e := api.service.SaveBoard(c.Context(), audit.MetaFromFiber(c), id, in)
	return response(c, data, e)
}
func (api *API) DeleteBoard(c fiber.Ctx) error {
	var in struct {
		Version int64 `json:"version"`
	}
	if e := decode(c, &in); e != nil {
		return response(c, nil, e)
	}
	id, e := adminutil.ParseIDParam(c)
	if e != nil {
		return response(c, nil, e)
	}
	return response(c, nil, api.service.DeleteBoard(c.Context(), audit.MetaFromFiber(c), id, in.Version))
}

func (api *API) Summary(c fiber.Ctx) error {
	data, err := api.service.Summary(c.Context())
	return response(c, data, databaseError(err))
}
