package router

import (
	"github.com/gofiber/fiber/v3"
	authmw "github.com/gofurry/gofurry-admin/internal/app/auth/middleware"
	"github.com/gofurry/gofurry-admin/internal/bootstrap"
)

func api(root fiber.Router, runtime *bootstrap.Runtime) {
	v1(root.Group("/v1"), runtime)
}

func v1(root fiber.Router, runtime *bootstrap.Runtime) {
	authRoutes(root.Group("/auth"), runtime)

	protected := root.Group("")
	protected.Use(authmw.Required(runtime.AuthService))
	optionsRoutes(protected.Group("/options"), runtime)
	navRoutes(protected.Group("/nav"), runtime)
	gameRoutes(protected.Group("/game"), runtime)
	collectionRoutes(protected.Group("/collection"), runtime)
	metricRoutes(protected.Group("/metrics"), runtime)
	changeRoutes(protected.Group("/changes"), runtime)
}

func changeRoutes(root fiber.Router, runtime *bootstrap.Runtime) {
	api := runtime.ChangeAPI
	root.Get("/overview", api.Overview)
	root.Get("/registry", api.Registry)
	root.Get("/checkpoints", api.Checkpoints)
	root.Get("/events", api.Events)
}

func metricRoutes(root fiber.Router, runtime *bootstrap.Runtime) {
	api := runtime.MetricAPI
	root.Get("/overview", api.Overview)
	root.Get("/registry", api.Registry)
	root.Get("/checkpoints", api.Checkpoints)
	root.Get("/daily", api.Daily)
	root.Get("/entities", api.Entities)
}

func collectionRoutes(root fiber.Router, runtime *bootstrap.Runtime) {
	api := runtime.CollectionAPI
	root.Get("/overview", api.Overview)
	root.Get("/instances", api.Instances)
	root.Get("/schedules", api.Schedules)
	root.Put("/schedules/:domain/:id", api.UpdateSchedule)
	root.Post("/schedules/:domain/:id/run", api.RunSchedule)
	root.Get("/jobs", api.Jobs)
	root.Post("/jobs", api.CreateJobs)
	root.Get("/jobs/:domain/:id", api.Job)
	root.Post("/jobs/:domain/:id/cancel", api.CancelJob)
	root.Post("/jobs/:domain/:id/retry", api.RetryJob)
	root.Get("/runs", api.Runs)
	root.Get("/runs/:domain/:id", api.Run)
	root.Get("/runs/:domain/:id/results", api.Results)
	root.Get("/charts/outcomes", api.Charts)
	root.Get("/charts/coverage", api.Charts)
	root.Get("/charts/timing", api.Charts)
}

func authRoutes(root fiber.Router, runtime *bootstrap.Runtime) {
	root.Get("/state", runtime.AuthAPI.State)
	root.Post("/bootstrap", runtime.AuthAPI.Bootstrap)
	root.Post("/login", runtime.AuthAPI.Login)
	root.Post("/logout", runtime.AuthAPI.Logout)
	root.Get("/me", authmw.Required(runtime.AuthService), runtime.AuthAPI.Me)
}

func optionsRoutes(root fiber.Router, runtime *bootstrap.Runtime) {
	root.Get("/sites", runtime.OptionsAPI.SiteOptions)
	root.Get("/site-targets", runtime.OptionsAPI.SiteTargetOptions)
	root.Get("/site-groups", runtime.OptionsAPI.SiteGroupOptions)
	root.Get("/games", runtime.OptionsAPI.GameOptions)
	root.Get("/tags", runtime.OptionsAPI.TagOptions)
}

func navRoutes(root fiber.Router, runtime *bootstrap.Runtime) {
	api := runtime.NavAPI
	root.Get("/sayings", api.ListSayings)
	root.Post("/sayings", api.CreateSaying)
	root.Get("/sayings/:id", api.GetSaying)
	root.Put("/sayings/:id", api.UpdateSaying)
	root.Delete("/sayings/:id", api.DeleteSaying)

	root.Get("/update-notices", api.ListUpdateNotices)
	root.Post("/update-notices", api.CreateUpdateNotice)
	root.Get("/update-notices/:id", api.GetUpdateNotice)
	root.Put("/update-notices/:id", api.UpdateUpdateNotice)
	root.Delete("/update-notices/:id", api.DeleteUpdateNotice)

	root.Get("/collector-domains", api.ListCollectorDomains)
	root.Post("/collector-domains", api.CreateCollectorDomain)
	root.Get("/collector-domains/:id", api.GetCollectorDomain)
	root.Put("/collector-domains/:id", api.UpdateCollectorDomain)
	root.Delete("/collector-domains/:id", api.DeleteCollectorDomain)
	root.Post("/collector-domains/:id/primary", api.SetPrimaryCollectorDomain)

	root.Get("/sites", api.ListSites)
	root.Post("/sites", api.CreateSite)
	root.Get("/sites/:id", api.GetSite)
	root.Put("/sites/:id", api.UpdateSite)
	root.Delete("/sites/:id", api.DeleteSite)

	root.Get("/site-groups", api.ListSiteGroups)
	root.Post("/site-groups", api.CreateSiteGroup)
	root.Get("/site-groups/:id", api.GetSiteGroup)
	root.Put("/site-groups/:id", api.UpdateSiteGroup)
	root.Delete("/site-groups/:id", api.DeleteSiteGroup)

	root.Get("/site-group-maps", api.ListSiteGroupMaps)
	root.Post("/site-group-maps", api.CreateSiteGroupMap)
	root.Put("/site-group-maps/bulk-replace", api.BulkReplaceSiteGroupMaps)
	root.Get("/site-group-maps/:id", api.GetSiteGroupMap)
	root.Put("/site-group-maps/:id", api.UpdateSiteGroupMap)
	root.Delete("/site-group-maps/:id", api.DeleteSiteGroupMap)

	root.Get("/featured-sites", api.ListFeaturedSites)
	root.Post("/featured-sites", api.CreateFeaturedSite)
	root.Get("/featured-sites/:id", api.GetFeaturedSite)
	root.Put("/featured-sites/:id", api.UpdateFeaturedSite)
	root.Delete("/featured-sites/:id", api.DeleteFeaturedSite)
}

func gameRoutes(root fiber.Router, runtime *bootstrap.Runtime) {
	api := runtime.GameAPI
	root.Get("/games", api.ListGames)
	root.Post("/games", api.CreateGame)
	root.Get("/games/steam-asset", api.ResolveSteamGameAsset)
	root.Get("/games/steam-prefill", api.ResolveSteamGamePrefill)
	root.Get("/games/:id", api.GetGame)
	root.Put("/games/:id", api.UpdateGame)
	root.Delete("/games/:id", api.DeleteGame)

	root.Get("/comments", api.ListComments)
	root.Post("/comments", api.CreateComment)
	root.Get("/comments/:id", api.GetComment)
	root.Put("/comments/:id", api.UpdateComment)
	root.Delete("/comments/:id", api.DeleteComment)

	root.Get("/prizes", api.ListPrizes)
	root.Post("/prizes", api.CreatePrize)
	root.Get("/prizes/:id", api.GetPrize)
	root.Put("/prizes/:id", api.UpdatePrize)
	root.Delete("/prizes/:id", api.DeletePrize)

	root.Get("/tags", api.ListTags)
	root.Post("/tags", api.CreateTag)
	root.Get("/tags/:id", api.GetTag)
	root.Put("/tags/:id", api.UpdateTag)
	root.Delete("/tags/:id", api.DeleteTag)

	root.Get("/tag-maps", api.ListTagMaps)
	root.Post("/tag-maps", api.CreateTagMap)
	root.Put("/tag-maps/bulk-replace", api.BulkReplaceTagMaps)
	root.Get("/tag-maps/by-tag/:id", api.ListTagMapGameIDs)
	root.Put("/tag-maps/bulk-replace-by-tag", api.BulkReplaceTagGameMaps)
	root.Get("/tag-maps/:id", api.GetTagMap)
	root.Put("/tag-maps/:id", api.UpdateTagMap)
	root.Delete("/tag-maps/:id", api.DeleteTagMap)
}
