package dataops

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/pkg/common"
)

type API struct{ service *Service }

func NewAPI(service *Service) *API { return &API{service: service} }

func (api *API) Overview(c fiber.Ctx) error {
	return common.NewResponse(c).SuccessWithData(api.service.Overview(c.Context()))
}
