package service

import (
	"github.com/bytedance/sonic"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/cachekeys"
	navmodels "github.com/gofurry/gofurry-nav-backend/apps/nav/navPage/models"
	cs "github.com/gofurry/gofurry-nav-backend/common/service"
)

func GetCachedSiteDirectory(lang string) []navmodels.SiteVo {
	raw, err := cs.GetString(cachekeys.SiteDirectory(lang))
	if err != nil || raw == "" {
		return []navmodels.SiteVo{}
	}

	var items []navmodels.SiteVo
	if unmarshalErr := sonic.Unmarshal([]byte(raw), &items); unmarshalErr != nil {
		return []navmodels.SiteVo{}
	}
	return items
}
