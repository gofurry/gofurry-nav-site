package workbench

import (
	"github.com/gofiber/fiber/v3"
	authmw "github.com/gofurry/gofurry-admin/internal/app/auth/middleware"
	"github.com/gofurry/gofurry-admin/pkg/common"
)

type API struct{ service *Service }

func NewAPI(service *Service) *API { return &API{service: service} }

func (api *API) Summary(c fiber.Ctx) error {
	principal, err := authmw.CurrentPrincipal(c)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).SuccessWithData(api.service.Summary(c.Context(), principal))
}
