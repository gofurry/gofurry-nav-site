package routers

import (
	"github.com/gofiber/fiber/v3"
	nav "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/controller"
	site "github.com/gofurry/gofurry-nav-backend/apps/nav/sitePage/controller"
	siteCommon "github.com/gofurry/gofurry-nav-backend/apps/system/site/controller"
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
	g.Get("/page/ping/list", nav.NavPageApi.GetPingList)   // 获取所有导航站点延迟信息
	// 导航页头部组件
	g.Get("/page/search/baidu", nav.NavPageApi.GetBaiduSearchSuggestion)       // 解析百度搜索建议框
	g.Get("/page/search/bing", nav.NavPageApi.GetBingSearchSuggestion)         // 解析必应搜索建议框
	g.Get("/page/search/google", nav.NavPageApi.GetGoogleSearchSuggestion)     // 解析谷歌搜索建议框
	g.Get("/page/search/bilibili", nav.NavPageApi.GetBiliBiliSearchSuggestion) // 解析B站搜索建议框
	g.Get("/page/header/getSaying", nav.NavPageApi.GetSaying)                  // 提供随机金句
	g.Get("/page/header/image/url", nav.NavPageApi.GetImageUrl)                // 获取 CDN 中随机壁纸

	// 详情页
	g.Get("/site/getSiteDetail", site.SitePageApi.GetSiteDetail)         // 获取单个站点的信息
	g.Get("/site/getSitePingRecord", site.SitePageApi.GetSitePingRecord) // 获取单个站点的 Ping 记录
	g.Get("/site/getSiteHttpRecord", site.SitePageApi.GetSiteHttpRecord) // 获取单个站点的 HTTP 记录
	g.Get("/site/getSiteDnsRecord", site.SitePageApi.GetSiteDnsRecord)   // 获取单个站点的 DNS 记录
}

func siteApi(g fiber.Router) {
	g.Get("/changelog", siteCommon.SiteApi.GetSiteChangeLog) // 更新公告
}
