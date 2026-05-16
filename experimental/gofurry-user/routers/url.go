package routers

import (
	oauth "github.com/gofurry/gofurry-user/apps/oauth/controller"
	user "github.com/gofurry/gofurry-user/apps/user/controller"
	email "github.com/gofurry/gofurry-user/apps/util/email/controller"
	"github.com/gofurry/gofurry-user/middleware"
	"github.com/gofiber/fiber/v2"
)

/*
 * @Desc: 接口层
 * @author: 福狼
 * @version: v1.0.0
 */

func userApi(g fiber.Router) {
	g.Post("/login", user.UserApi.Login)       // 登录
	g.Post("/register", user.UserApi.Register) // 注册
	//g.POST("/retrieve", user.UserApi.Retrieve) // 邮箱找回密码
	//g.GET("/logout", user.UserApi.Logout)      // 登出账户
	//
	g.Use(middleware.JWTMiddleWare())
	{
		//	g.POST("/updateInfo", user.UserApi.UpdateInfo)         // 修改个人信息
		//	g.POST("/updateEmail", user.UserApi.UpdateEmail)       // 修改邮箱
		//	g.POST("/updatePassword", user.UserApi.UpdatePassword) // 修改密码
		//	g.GET("/info", user.UserApi.GetInfo)                   // 展示个人信息
		//	// 登录记录
		//	g.GET("/login/log", user.LoginLogApi.GetLoginLog)
	}
}

func oauthApi(g fiber.Router) {
	g.Get("/callback/github", oauth.OauthApi.GithubCallback) // github 三方登录
	g.Get("/callback/gitee", oauth.OauthApi.GiteeCallback)   // gitee 三方登录
	//g.Get("/callback/google", oauth.OauthApi.GoogleCallback)
}

func utilApi(g fiber.Router) {
	// 邮箱接口
	g.Get("/email/send", email.EmailApi.Send) // 邮箱验证码
}
