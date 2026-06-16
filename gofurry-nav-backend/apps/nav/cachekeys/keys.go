package cachekeys

import "strings"

const (
	SiteListV2       = "site:list:v2"
	GroupList        = "group:list"
	GroupSiteMap     = "group:site:map"
	FeaturedSiteList = "featured-sites:list"

	HomePrefix          = "nav:home:v3:"
	SiteDirectoryPrefix = "nav:site-directory:v1:"
	SiteGroupPrefix     = "nav:site-group:v1:"
)

func normalizeLang(lang string) string {
	if strings.EqualFold(strings.TrimSpace(lang), "en") {
		return "en"
	}
	return "zh"
}

func Home(lang string) string {
	return HomePrefix + normalizeLang(lang)
}

func SiteDirectory(lang string) string {
	return SiteDirectoryPrefix + normalizeLang(lang)
}

func SiteGroup(lang, groupID string) string {
	return SiteGroupPrefix + normalizeLang(lang) + ":" + strings.TrimSpace(groupID)
}
