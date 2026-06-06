package router

import (
	"github.com/gofiber/fiber/v3"
	auth "github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/auth/controller"
	authmw "github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/auth/middleware"
	gameadmin "github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/gameadmin/controller"
	navadmin "github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/navadmin/controller"
	options "github.com/gofurry/awesome-fiber-template/v3/medium/internal/app/shared/options/controller"
)

func api(root fiber.Router) {
	v1(root.Group("/v1"))
}

func v1(root fiber.Router) {
	authRoutes(root.Group("/auth"))

	protected := root.Group("")
	protected.Use(authmw.Required())
	optionsRoutes(protected.Group("/options"))
	navRoutes(protected.Group("/nav"))
	gameRoutes(protected.Group("/game"))
}

func authRoutes(root fiber.Router) {
	root.Get("/state", auth.AuthAPI.State)
	root.Post("/bootstrap", auth.AuthAPI.Bootstrap)
	root.Post("/login", auth.AuthAPI.Login)
	root.Post("/logout", auth.AuthAPI.Logout)
	root.Get("/me", authmw.Required(), auth.AuthAPI.Me)
}

func optionsRoutes(root fiber.Router) {
	root.Get("/sites", options.OptionsAPI.SiteOptions)
	root.Get("/site-groups", options.OptionsAPI.SiteGroupOptions)
	root.Get("/games", options.OptionsAPI.GameOptions)
	root.Get("/tags", options.OptionsAPI.TagOptions)
}

func navRoutes(root fiber.Router) {
	root.Get("/sayings", navadmin.NavAPI.ListSayings)
	root.Post("/sayings", navadmin.NavAPI.CreateSaying)
	root.Get("/sayings/:id", navadmin.NavAPI.GetSaying)
	root.Put("/sayings/:id", navadmin.NavAPI.UpdateSaying)
	root.Delete("/sayings/:id", navadmin.NavAPI.DeleteSaying)

	root.Get("/update-notices", navadmin.NavAPI.ListUpdateNotices)
	root.Post("/update-notices", navadmin.NavAPI.CreateUpdateNotice)
	root.Get("/update-notices/:id", navadmin.NavAPI.GetUpdateNotice)
	root.Put("/update-notices/:id", navadmin.NavAPI.UpdateUpdateNotice)
	root.Delete("/update-notices/:id", navadmin.NavAPI.DeleteUpdateNotice)

	root.Get("/collector-domains", navadmin.NavAPI.ListCollectorDomains)
	root.Post("/collector-domains", navadmin.NavAPI.CreateCollectorDomain)
	root.Get("/collector-domains/:id", navadmin.NavAPI.GetCollectorDomain)
	root.Put("/collector-domains/:id", navadmin.NavAPI.UpdateCollectorDomain)
	root.Delete("/collector-domains/:id", navadmin.NavAPI.DeleteCollectorDomain)

	root.Get("/sites", navadmin.NavAPI.ListSites)
	root.Post("/sites", navadmin.NavAPI.CreateSite)
	root.Get("/sites/:id", navadmin.NavAPI.GetSite)
	root.Put("/sites/:id", navadmin.NavAPI.UpdateSite)
	root.Delete("/sites/:id", navadmin.NavAPI.DeleteSite)

	root.Get("/site-groups", navadmin.NavAPI.ListSiteGroups)
	root.Post("/site-groups", navadmin.NavAPI.CreateSiteGroup)
	root.Get("/site-groups/:id", navadmin.NavAPI.GetSiteGroup)
	root.Put("/site-groups/:id", navadmin.NavAPI.UpdateSiteGroup)
	root.Delete("/site-groups/:id", navadmin.NavAPI.DeleteSiteGroup)

	root.Get("/site-group-maps", navadmin.NavAPI.ListSiteGroupMaps)
	root.Post("/site-group-maps", navadmin.NavAPI.CreateSiteGroupMap)
	root.Put("/site-group-maps/bulk-replace", navadmin.NavAPI.BulkReplaceSiteGroupMaps)
	root.Get("/site-group-maps/:id", navadmin.NavAPI.GetSiteGroupMap)
	root.Put("/site-group-maps/:id", navadmin.NavAPI.UpdateSiteGroupMap)
	root.Delete("/site-group-maps/:id", navadmin.NavAPI.DeleteSiteGroupMap)

	root.Get("/featured-sites", navadmin.NavAPI.ListFeaturedSites)
	root.Post("/featured-sites", navadmin.NavAPI.CreateFeaturedSite)
	root.Get("/featured-sites/:id", navadmin.NavAPI.GetFeaturedSite)
	root.Put("/featured-sites/:id", navadmin.NavAPI.UpdateFeaturedSite)
	root.Delete("/featured-sites/:id", navadmin.NavAPI.DeleteFeaturedSite)
}

func gameRoutes(root fiber.Router) {
	root.Get("/games", gameadmin.GameAPI.ListGames)
	root.Post("/games", gameadmin.GameAPI.CreateGame)
	root.Get("/games/:id", gameadmin.GameAPI.GetGame)
	root.Put("/games/:id", gameadmin.GameAPI.UpdateGame)
	root.Delete("/games/:id", gameadmin.GameAPI.DeleteGame)

	root.Get("/comments", gameadmin.GameAPI.ListComments)
	root.Post("/comments", gameadmin.GameAPI.CreateComment)
	root.Get("/comments/:id", gameadmin.GameAPI.GetComment)
	root.Put("/comments/:id", gameadmin.GameAPI.UpdateComment)
	root.Delete("/comments/:id", gameadmin.GameAPI.DeleteComment)

	root.Get("/creators", gameadmin.GameAPI.ListCreators)
	root.Post("/creators", gameadmin.GameAPI.CreateCreator)
	root.Get("/creators/:id", gameadmin.GameAPI.GetCreator)
	root.Put("/creators/:id", gameadmin.GameAPI.UpdateCreator)
	root.Delete("/creators/:id", gameadmin.GameAPI.DeleteCreator)

	root.Get("/prizes", gameadmin.GameAPI.ListPrizes)
	root.Post("/prizes", gameadmin.GameAPI.CreatePrize)
	root.Get("/prizes/:id", gameadmin.GameAPI.GetPrize)
	root.Put("/prizes/:id", gameadmin.GameAPI.UpdatePrize)
	root.Delete("/prizes/:id", gameadmin.GameAPI.DeletePrize)

	root.Get("/tags", gameadmin.GameAPI.ListTags)
	root.Post("/tags", gameadmin.GameAPI.CreateTag)
	root.Get("/tags/:id", gameadmin.GameAPI.GetTag)
	root.Put("/tags/:id", gameadmin.GameAPI.UpdateTag)
	root.Delete("/tags/:id", gameadmin.GameAPI.DeleteTag)

	root.Get("/tag-maps", gameadmin.GameAPI.ListTagMaps)
	root.Post("/tag-maps", gameadmin.GameAPI.CreateTagMap)
	root.Put("/tag-maps/bulk-replace", gameadmin.GameAPI.BulkReplaceTagMaps)
	root.Get("/tag-maps/:id", gameadmin.GameAPI.GetTagMap)
	root.Put("/tag-maps/:id", gameadmin.GameAPI.UpdateTagMap)
	root.Delete("/tag-maps/:id", gameadmin.GameAPI.DeleteTagMap)
}
