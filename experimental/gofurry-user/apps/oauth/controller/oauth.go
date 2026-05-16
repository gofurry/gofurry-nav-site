package controller

import (
	"net/http"
	"time"

	"github.com/gofurry/gofurry-user/apps/oauth/service"
	"github.com/gofurry/gofurry-user/common"
	"github.com/gofiber/fiber/v2"
)

type oauthApi struct{}

var OauthApi *oauthApi

func init() {
	OauthApi = &oauthApi{}
}

// @Summary Github 三方登录
// @Schemes
// @Description Github 三方登录
// @Tags Oauth
// @Accept json
// @Produce json
// @Param code query string true "code"
// @Success 200 {object} common.ResultData
// @Router /oauth/callback/github [Get]
func (api *oauthApi) GithubCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	token, err := service.GetOauthService().GithubLogin(c, code)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}

	c.Cookie(&fiber.Cookie{
		Name:     "Authorization",
		Value:    token,
		Expires:  time.Now().Add(2 * time.Hour), // 2小时过期
		Path:     "/",                           // 全站有效
		Domain:   "127.0.0.1",                   // 前端域名
		Secure:   false,                         // 开发环境 false 生产环境true
		HTTPOnly: true,
		SameSite: "Lax",
	})
	return c.Redirect("https://127.0.0.1:8888/", http.StatusFound)
}

// @Summary Gitee 三方登录
// @Schemes
// @Description Gitee 三方登录
// @Tags Oauth
// @Accept json
// @Produce json
// @Param code query string true "code"
// @Success 200 {object} common.ResultData
// @Router /oauth/callback/gitee [Get]
func (api *oauthApi) GiteeCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	token, err := service.GetOauthService().GiteeLogin(c, code)
	if err != nil {
		return common.NewResponse(c).Error(err)
	}

	c.Cookie(&fiber.Cookie{
		Name:     "Authorization",
		Value:    token,
		Expires:  time.Now().Add(2 * time.Hour), // 2小时过期
		Path:     "/",                           // 全站有效
		Domain:   "127.0.0.1",                   // 前端域名
		Secure:   false,                         // 开发环境 false 生产环境true
		HTTPOnly: true,
		SameSite: "Lax",
	})
	return c.Redirect("https://127.0.0.1:8888/", http.StatusFound)
}
