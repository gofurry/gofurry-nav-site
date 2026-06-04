package routers

import (
	"github.com/gofiber/fiber/v3"
	nav "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/controller"
)

/*
 * @Desc: 接口层
 * @author: 福狼
 * @version: v1.0.0
 */

// 导航相关
func navApi(g fiber.Router) {
	// 导航页
	// 导航部分
	g.Get("/page/site/list", nav.NavPageApi.GetSiteList)   // 获取所有导航站点信息
	g.Get("/page/group/list", nav.NavPageApi.GetGroupList) // 获取所有导航站点分组信息
}
