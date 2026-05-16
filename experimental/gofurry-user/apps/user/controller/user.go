package controller

import (
	"github.com/gofurry/gofurry-user/apps/user/models"
	"github.com/gofurry/gofurry-user/apps/user/service"
	"github.com/gofurry/gofurry-user/common"
	"github.com/gofiber/fiber/v2"
)

type userApi struct{}

var UserApi *userApi

func init() {
	UserApi = &userApi{}
}

// @Summary 登录
// @Schemes
// @Description 登录
// @Tags System-user
// @Accept json
// @Produce json
// @Param body body models.UserLoginRequest true "请求body"
// @Success 200 {object} common.ResultData
// @Router /api/user/login [Post]
func (api *userApi) Login(c *fiber.Ctx) error {
	var req models.UserLoginRequest

	if err := c.BodyParser(&req); err != nil {
		return common.NewResponse(c).Error("参数错误: " + err.Error())
	}

	token, err := service.GetUserService().Login(c, req)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	var data = map[string]interface{}{
		"token": token,
	}
	return common.NewResponse(c).SuccessWithData(data)
}

// @Summary 注册
// @Schemes
// @Description 注册
// @Tags System-user
// @Accept json
// @Produce json
// @Param body body models.UserRegisterRequest true "请求body"
// @Success 200 {object} common.ResultData
// @Router /api/user/register [Post]
func (api *userApi) Register(c *fiber.Ctx) error {
	var req models.UserRegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return common.NewResponse(c).Error("参数错误: " + err.Error())
	}
	err := service.GetUserService().Register(req)
	if err != nil {
		return common.NewResponse(c).Error(err.GetMsg())
	}
	return common.NewResponse(c).Success()
}
