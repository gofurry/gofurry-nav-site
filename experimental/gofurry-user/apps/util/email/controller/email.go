package controller

import (
	"github.com/gofurry/gofurry-user/apps/util/email/service"
	"github.com/gofurry/gofurry-user/common"
	"github.com/gofiber/fiber/v2"
)

/*
 * @Desc: 邮箱服务
 * @author: bsyz
 * @version: v1.0.0
 */

type emailApi struct{}

var EmailApi *emailApi

func init() {
	EmailApi = &emailApi{}
}

// @Summary 邮箱验证码
// @Schemes
// @Description 邮箱验证码
// @Tags Util-email
// @Accept json
// @Produce json
// @Param email query string true "邮箱"
// @Success 200 {object} common.ResultData
// @Router /api/util/email/send [Get]
func (api *emailApi) Send(c *fiber.Ctx) error {
	email := c.Query("email")
	err := service.GetEmailService().SendEmail(email)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}
	return common.NewResponse(c).Success()
}
