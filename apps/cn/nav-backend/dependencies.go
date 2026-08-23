package main

import (
	collectcontroller "github.com/gofurry/gofurry-nav-backend/apps/nav/collect/controller"
	collectdao "github.com/gofurry/gofurry-nav-backend/apps/nav/collect/dao"
	collectservice "github.com/gofurry/gofurry-nav-backend/apps/nav/collect/service"
	detailcontroller "github.com/gofurry/gofurry-nav-backend/apps/nav/detail/controller"
	detaildao "github.com/gofurry/gofurry-nav-backend/apps/nav/detail/dao"
	detailservice "github.com/gofurry/gofurry-nav-backend/apps/nav/detail/service"
	homecontroller "github.com/gofurry/gofurry-nav-backend/apps/nav/home/controller"
	homeservice "github.com/gofurry/gofurry-nav-backend/apps/nav/home/service"
	navpagecontroller "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/controller"
	navpagedao "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/dao"
	navpageservice "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/service"
	observationdao "github.com/gofurry/gofurry-nav-backend/apps/nav/readmodel/dao"
	readmodelservice "github.com/gofurry/gofurry-nav-backend/apps/nav/readmodel/service"
	searchcontroller "github.com/gofurry/gofurry-nav-backend/apps/nav/search/controller"
	searchservice "github.com/gofurry/gofurry-nav-backend/apps/nav/search/service"
	sitepagedao "github.com/gofurry/gofurry-nav-backend/apps/nav/sitePage/dao"
	sitepageservice "github.com/gofurry/gofurry-nav-backend/apps/nav/sitePage/service"
	siteindexcontroller "github.com/gofurry/gofurry-nav-backend/apps/nav/siteindex/controller"
	siteindexservice "github.com/gofurry/gofurry-nav-backend/apps/nav/siteindex/service"
	summaryservice "github.com/gofurry/gofurry-nav-backend/apps/nav/summary/service"
	updatescontroller "github.com/gofurry/gofurry-nav-backend/apps/nav/updates/controller"
	updatesservice "github.com/gofurry/gofurry-nav-backend/apps/nav/updates/service"
	"github.com/gofurry/gofurry-nav-backend/apps/schedule/task"
	navsqlc "github.com/gofurry/gofurry-nav-backend/internal/db/nav/sqlc"
	"github.com/gofurry/gofurry-nav-backend/routers"
	"github.com/jackc/pgx/v5/pgxpool"
)

type applicationDependencies struct {
	routes    routers.NavDependencies
	navStore  task.NavCacheStore
	navReader task.NavCacheReader
	home      task.HomeCacheReader
	views     task.SiteViewStore
}

func newApplicationDependencies(pool *pgxpool.Pool) applicationDependencies {
	queries := navsqlc.New(pool)
	navStore := navpagedao.New(queries)
	navService := navpageservice.New(navStore)
	homeService := homeservice.New(navService)
	readModelService := readmodelservice.New(observationdao.New(queries))
	summaryService := summaryservice.GetSummaryService()
	detailService := detailservice.New(detaildao.New(queries), summaryService, readModelService)
	sitePageStore := sitepagedao.New(queries)
	sitePageService := sitepageservice.New(sitePageStore)
	collectService := collectservice.New(collectdao.New(queries), readModelService)
	searchService := searchservice.New(navService)

	return applicationDependencies{
		routes: routers.NavDependencies{
			Home:      homecontroller.New(homeService),
			Updates:   updatescontroller.New(updatesservice.New(queries)),
			Search:    searchcontroller.New(searchService, searchservice.NewRedisSuggestionRateLimiter()),
			SiteIndex: siteindexcontroller.New(siteindexservice.New(navStore)),
			NavPage:   navpagecontroller.New(navService),
			Detail:    detailcontroller.New(detailService, sitePageService),
			Collect:   collectcontroller.New(collectService),
		},
		navStore:  navStore,
		navReader: navService,
		home:      homeService,
		views:     sitePageStore,
	}
}
