package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/auth/authorization"
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
	selfServiceRoutes(protected.Group("/auth/self"), runtime)
	accountRoutes(protected.Group("/auth/accounts"), runtime)
	optionsRoutes(protected.Group("/options"), runtime)
	navRoutes(protected.Group("/nav"), runtime)
	gameRoutes(protected.Group("/game"), runtime)
	collectionRoutes(protected.Group("/collection"), runtime)
	metricRoutes(protected.Group("/metrics"), runtime)
	changeRoutes(protected.Group("/changes"), runtime)
	workbenchRoutes(protected.Group("/workbench"), runtime)
	dataOpsRoutes(protected.Group("/dataops"), runtime)
	auditRoutes(protected.Group("/audit"), runtime)
}

func selfServiceRoutes(root fiber.Router, runtime *bootstrap.Runtime) {
	root.Put("/username", runtime.AuthAPI.ChangeOwnUsername)
	root.Post("/password", runtime.AuthAPI.ChangeOwnPassword)
}

func workbenchRoutes(root fiber.Router, runtime *bootstrap.Runtime) {
	root.Get("/summary", authmw.Require(authorization.ContentRead), runtime.WorkbenchAPI.Summary)
}

func dataOpsRoutes(root fiber.Router, runtime *bootstrap.Runtime) {
	root.Get("/overview", authmw.Require(authorization.DataOpsRead), runtime.DataOpsAPI.Overview)
}

func auditRoutes(root fiber.Router, runtime *bootstrap.Runtime) {
	root.Get("/logs", authmw.Require(authorization.AuditRead), runtime.AuditAPI.Logs)
}

func changeRoutes(root fiber.Router, runtime *bootstrap.Runtime) {
	api := runtime.ChangeAPI
	root.Get("/overview", authmw.Require(authorization.ChangesRead), api.Overview)
	root.Get("/registry", authmw.Require(authorization.ChangesTechnical), api.Registry)
	root.Get("/checkpoints", authmw.Require(authorization.ChangesTechnical), api.Checkpoints)
	root.Get("/events", authmw.Require(authorization.ChangesRead), api.Events)
}

func metricRoutes(root fiber.Router, runtime *bootstrap.Runtime) {
	api := runtime.MetricAPI
	root.Get("/overview", authmw.Require(authorization.MetricsRead), api.Overview)
	root.Get("/registry", authmw.Require(authorization.MetricsTechnical), api.Registry)
	root.Get("/checkpoints", authmw.Require(authorization.MetricsTechnical), api.Checkpoints)
	root.Get("/daily", authmw.Require(authorization.MetricsRead), api.Daily)
	root.Get("/entities", authmw.Require(authorization.MetricsRead), api.Entities)
}

func collectionRoutes(root fiber.Router, runtime *bootstrap.Runtime) {
	api := runtime.CollectionAPI
	root.Get("/overview", authmw.Require(authorization.CollectionRead), api.Overview)
	root.Get("/instances", authmw.Require(authorization.CollectionRead), api.Instances)
	root.Get("/schedules", authmw.Require(authorization.CollectionRead), api.Schedules)
	root.Put("/schedules/:domain/:id", authmw.Require(authorization.CollectionControl), api.UpdateSchedule)
	root.Post("/schedules/:domain/:id/run", authmw.Require(authorization.CollectionExecute), api.RunSchedule)
	root.Get("/jobs", authmw.Require(authorization.CollectionRead), api.Jobs)
	root.Post("/jobs", authmw.Require(authorization.CollectionExecute), api.CreateJobs)
	root.Get("/jobs/:domain/:id", authmw.Require(authorization.CollectionRead), api.Job)
	root.Post("/jobs/:domain/:id/cancel", authmw.Require(authorization.CollectionControl), api.CancelJob)
	root.Post("/jobs/:domain/:id/retry", authmw.Require(authorization.CollectionExecute), api.RetryJob)
	root.Get("/runs", authmw.Require(authorization.CollectionRead), api.Runs)
	root.Get("/runs/:domain/:id", authmw.Require(authorization.CollectionRead), api.Run)
	root.Get("/runs/:domain/:id/results", authmw.Require(authorization.CollectionRead), api.Results)
	root.Get("/charts/outcomes", authmw.Require(authorization.CollectionRead), api.Charts)
	root.Get("/charts/coverage", authmw.Require(authorization.CollectionRead), api.Charts)
	root.Get("/charts/timing", authmw.Require(authorization.CollectionRead), api.Charts)
}

func authRoutes(root fiber.Router, runtime *bootstrap.Runtime) {
	root.Get("/state", runtime.AuthAPI.State)
	root.Post("/bootstrap", runtime.AuthAPI.Bootstrap)
	root.Post("/login", runtime.AuthAPI.Login)
	root.Post("/logout", runtime.AuthAPI.Logout)
	root.Get("/me", authmw.Required(runtime.AuthService), runtime.AuthAPI.Me)
}

func accountRoutes(root fiber.Router, runtime *bootstrap.Runtime) {
	api := runtime.AuthAPI
	root.Get("/", authmw.Require(authorization.AccountManage), api.ListAccounts)
	root.Post("/", authmw.Require(authorization.AccountManage), api.CreateAccount)
	root.Put("/:id/display-name", authmw.Require(authorization.AccountManage), api.UpdateDisplayName)
	root.Put("/:id/role", authmw.Require(authorization.AccountManage), api.ChangeRole)
	root.Put("/:id/status", authmw.Require(authorization.AccountManage), api.ChangeStatus)
	root.Post("/:id/password", authmw.Require(authorization.AccountManage), api.ResetAccountPassword)
	root.Post("/:id/revoke-sessions", authmw.Require(authorization.AccountManage), api.RevokeSessions)
}

func optionsRoutes(root fiber.Router, runtime *bootstrap.Runtime) {
	root.Get("/sites", authmw.Require(authorization.ContentRead), runtime.OptionsAPI.SiteOptions)
	root.Get("/site-targets", authmw.Require(authorization.ContentRead), runtime.OptionsAPI.SiteTargetOptions)
	root.Get("/site-groups", authmw.Require(authorization.ContentRead), runtime.OptionsAPI.SiteGroupOptions)
	root.Get("/games", authmw.Require(authorization.ContentRead), runtime.OptionsAPI.GameOptions)
	root.Get("/tags", authmw.Require(authorization.ContentRead), runtime.OptionsAPI.TagOptions)
}

func navRoutes(root fiber.Router, runtime *bootstrap.Runtime) {
	api := runtime.NavAPI
	root.Get("/sayings", authmw.Require(authorization.ContentRead), api.ListSayings)
	root.Post("/sayings", authmw.Require(authorization.ContentWrite), api.CreateSaying)
	root.Get("/sayings/:id", authmw.Require(authorization.ContentRead), api.GetSaying)
	root.Put("/sayings/:id", authmw.Require(authorization.ContentWrite), api.UpdateSaying)
	root.Delete("/sayings/:id", authmw.Require(authorization.ContentWrite), api.DeleteSaying)

	root.Get("/update-notices", authmw.Require(authorization.ContentRead), api.ListUpdateNotices)
	root.Post("/update-notices", authmw.Require(authorization.ContentWrite), api.CreateUpdateNotice)
	root.Get("/update-notices/:id", authmw.Require(authorization.ContentRead), api.GetUpdateNotice)
	root.Put("/update-notices/:id", authmw.Require(authorization.ContentWrite), api.UpdateUpdateNotice)
	root.Delete("/update-notices/:id", authmw.Require(authorization.ContentWrite), api.DeleteUpdateNotice)

	root.Get("/collector-domains", authmw.Require(authorization.ContentRead), api.ListCollectorDomains)
	root.Post("/collector-domains", authmw.Require(authorization.ContentWrite), api.CreateCollectorDomain)
	root.Get("/collector-domains/:id", authmw.Require(authorization.ContentRead), api.GetCollectorDomain)
	root.Put("/collector-domains/:id", authmw.Require(authorization.ContentWrite), api.UpdateCollectorDomain)
	root.Delete("/collector-domains/:id", authmw.Require(authorization.ContentWrite), api.DeleteCollectorDomain)
	root.Post("/collector-domains/:id/primary", authmw.Require(authorization.ContentWrite), api.SetPrimaryCollectorDomain)

	root.Get("/sites", authmw.Require(authorization.ContentRead), api.ListSites)
	root.Get("/site-summaries", authmw.Require(authorization.ContentRead), api.ListSiteWorkspaceSummaries)
	root.Post("/sites", authmw.Require(authorization.ContentWrite), api.CreateSite)
	root.Get("/sites/:id/workspace", authmw.Require(authorization.ContentRead), api.GetSiteWorkspace)
	root.Get("/sites/:id", authmw.Require(authorization.ContentRead), api.GetSite)
	root.Put("/sites/:id", authmw.Require(authorization.ContentWrite), api.UpdateSite)
	root.Delete("/sites/:id", authmw.Require(authorization.ContentWrite), api.DeleteSite)

	root.Get("/site-groups", authmw.Require(authorization.ContentRead), api.ListSiteGroups)
	root.Post("/site-groups", authmw.Require(authorization.ContentWrite), api.CreateSiteGroup)
	root.Get("/site-groups/:id", authmw.Require(authorization.ContentRead), api.GetSiteGroup)
	root.Put("/site-groups/:id", authmw.Require(authorization.ContentWrite), api.UpdateSiteGroup)
	root.Delete("/site-groups/:id", authmw.Require(authorization.ContentWrite), api.DeleteSiteGroup)

	root.Get("/site-group-maps", authmw.Require(authorization.ContentRead), api.ListSiteGroupMaps)
	root.Post("/site-group-maps", authmw.Require(authorization.ContentWrite), api.CreateSiteGroupMap)
	root.Put("/site-group-maps/bulk-replace", authmw.Require(authorization.ContentWrite), api.BulkReplaceSiteGroupMaps)
	root.Get("/site-group-maps/:id", authmw.Require(authorization.ContentRead), api.GetSiteGroupMap)
	root.Put("/site-group-maps/:id", authmw.Require(authorization.ContentWrite), api.UpdateSiteGroupMap)
	root.Delete("/site-group-maps/:id", authmw.Require(authorization.ContentWrite), api.DeleteSiteGroupMap)

	root.Get("/featured-sites", authmw.Require(authorization.ContentRead), api.ListFeaturedSites)
	root.Post("/featured-sites", authmw.Require(authorization.ContentWrite), api.CreateFeaturedSite)
	root.Get("/featured-sites/:id", authmw.Require(authorization.ContentRead), api.GetFeaturedSite)
	root.Put("/featured-sites/:id", authmw.Require(authorization.ContentWrite), api.UpdateFeaturedSite)
	root.Delete("/featured-sites/:id", authmw.Require(authorization.ContentWrite), api.DeleteFeaturedSite)
}

func gameRoutes(root fiber.Router, runtime *bootstrap.Runtime) {
	api := runtime.GameAPI
	root.Get("/games", authmw.Require(authorization.ContentRead), api.ListGames)
	root.Post("/games", authmw.Require(authorization.ContentWrite), api.CreateGame)
	root.Get("/games/steam-asset", authmw.Require(authorization.ContentRead), api.ResolveSteamGameAsset)
	root.Get("/games/steam-prefill", authmw.Require(authorization.ContentRead), api.ResolveSteamGamePrefill)
	root.Get("/games/:id/workspace", authmw.Require(authorization.ContentRead), api.GetGameWorkspace)
	root.Get("/games/:id", authmw.Require(authorization.ContentRead), api.GetGame)
	root.Put("/games/:id", authmw.Require(authorization.ContentWrite), api.UpdateGame)
	root.Delete("/games/:id", authmw.Require(authorization.ContentWrite), api.DeleteGame)

	root.Get("/comments", authmw.Require(authorization.ContentRead), api.ListComments)
	root.Post("/comments", authmw.Require(authorization.ContentWrite), api.CreateComment)
	root.Get("/comments/:id", authmw.Require(authorization.ContentRead), api.GetComment)
	root.Put("/comments/:id", authmw.Require(authorization.ContentWrite), api.UpdateComment)
	root.Delete("/comments/:id", authmw.Require(authorization.ContentWrite), api.DeleteComment)

	root.Get("/prizes", authmw.Require(authorization.ContentRead), api.ListPrizes)
	root.Post("/prizes", authmw.Require(authorization.ContentWrite), api.CreatePrize)
	root.Get("/prizes/:id", authmw.Require(authorization.ContentRead), api.GetPrize)
	root.Put("/prizes/:id", authmw.Require(authorization.ContentWrite), api.UpdatePrize)
	root.Delete("/prizes/:id", authmw.Require(authorization.ContentWrite), api.DeletePrize)

	root.Get("/tags", authmw.Require(authorization.ContentRead), api.ListTags)
	root.Post("/tags", authmw.Require(authorization.ContentWrite), api.CreateTag)
	root.Get("/tags/:id", authmw.Require(authorization.ContentRead), api.GetTag)
	root.Put("/tags/:id", authmw.Require(authorization.ContentWrite), api.UpdateTag)
	root.Delete("/tags/:id", authmw.Require(authorization.ContentWrite), api.DeleteTag)

	root.Get("/tag-maps", authmw.Require(authorization.ContentRead), api.ListTagMaps)
	root.Post("/tag-maps", authmw.Require(authorization.ContentWrite), api.CreateTagMap)
	root.Put("/tag-maps/bulk-replace", authmw.Require(authorization.ContentWrite), api.BulkReplaceTagMaps)
	root.Get("/tag-maps/by-tag/:id", authmw.Require(authorization.ContentRead), api.ListTagMapGameIDs)
	root.Put("/tag-maps/bulk-replace-by-tag", authmw.Require(authorization.ContentWrite), api.BulkReplaceTagGameMaps)
	root.Get("/tag-maps/:id", authmw.Require(authorization.ContentRead), api.GetTagMap)
	root.Put("/tag-maps/:id", authmw.Require(authorization.ContentWrite), api.UpdateTagMap)
	root.Delete("/tag-maps/:id", authmw.Require(authorization.ContentWrite), api.DeleteTagMap)
}
