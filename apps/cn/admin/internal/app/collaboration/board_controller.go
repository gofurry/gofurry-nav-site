package collaboration

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/shared/adminutil"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
)

func (api *API) Board(c fiber.Ctx) error {
	data, e := api.service.Board(c.Context())
	return response(c, data, e)
}
func (api *API) CreateBoardNode(c fiber.Ctx) error { return api.saveBoardNode(c, 0) }
func (api *API) UpdateBoardNode(c fiber.Ctx) error {
	id, e := adminutil.ParseIDParam(c)
	if e != nil {
		return response(c, nil, e)
	}
	return api.saveBoardNode(c, id)
}
func (api *API) saveBoardNode(c fiber.Ctx, id int64) error {
	var in BoardNodeInput
	if e := decode(c, &in); e != nil {
		return response(c, nil, e)
	}
	data, e := api.service.SaveBoardNode(c.Context(), audit.MetaFromFiber(c), id, in)
	return response(c, data, e)
}
func (api *API) MoveBoardNodes(c fiber.Ctx) error {
	var in struct {
		Nodes []BoardLayout `json:"nodes"`
	}
	if e := decode(c, &in); e != nil {
		return response(c, nil, e)
	}
	data, e := api.service.MoveBoardNodes(c.Context(), audit.MetaFromFiber(c), in.Nodes)
	return response(c, data, e)
}
func (api *API) DeleteBoardNode(c fiber.Ctx) error {
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
	return response(c, nil, api.service.DeleteBoardNode(c.Context(), audit.MetaFromFiber(c), id, in.Version))
}
func (api *API) CreateBoardEdge(c fiber.Ctx) error { return api.saveBoardEdge(c, 0) }
func (api *API) UpdateBoardEdge(c fiber.Ctx) error {
	id, e := adminutil.ParseIDParam(c)
	if e != nil {
		return response(c, nil, e)
	}
	return api.saveBoardEdge(c, id)
}
func (api *API) saveBoardEdge(c fiber.Ctx, id int64) error {
	var in BoardEdgeInput
	if e := decode(c, &in); e != nil {
		return response(c, nil, e)
	}
	data, e := api.service.SaveBoardEdge(c.Context(), audit.MetaFromFiber(c), id, in)
	return response(c, data, e)
}
func (api *API) DeleteBoardEdge(c fiber.Ctx) error {
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
	return response(c, nil, api.service.DeleteBoardEdge(c.Context(), audit.MetaFromFiber(c), id, in.Version))
}
