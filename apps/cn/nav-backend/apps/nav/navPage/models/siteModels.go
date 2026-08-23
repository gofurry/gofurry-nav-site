package models

import (
	"time"

	cm "github.com/gofurry/gofurry-nav-backend/common/models"
)

const TableNameGfnSite = "gfn_site"
const TableNameGfnCollectorDomain = "gfn_collector_domain"

func (*GfnSite) TableName() string {
	return TableNameGfnSite
}

type GfnSite struct {
	ID         int64        `json:"id"`
	Name       string       `json:"name"`
	NameEn     string       `json:"nameEn"`
	Domain     string       `json:"domain"`
	Info       string       `json:"info"`
	InfoEn     string       `json:"infoEn"`
	CreateTime cm.LocalTime `json:"createTime"`
	UpdateTime cm.LocalTime `json:"updateTime"`
	Country    *string      `json:"country"`
	Nsfw       string       `json:"nsfw"`
	Welfare    string       `json:"welfare"`
	ViewCount  int64        `json:"view_count"`
	Icon       *string      `json:"icon"`
	Deleted    bool         `json:"deleted"`
}

type GfnSiteIndex struct {
	ID         int64        `json:"id"`
	Domain     string       `json:"domain"`
	UpdateTime cm.LocalTime `json:"update_time"`
}

const TableNameGfnSiteGroup = "gfn_site_group"

func (*GfnSiteGroup) TableName() string {
	return TableNameGfnSiteGroup
}

type GfnSiteGroup struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	NameEn     string    `json:"nameEn"`
	Info       string    `json:"info"`
	InfoEn     string    `json:"infoEn"`
	Priority   int64     `json:"priority"`
	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `json:"updateTime"`
}

const TableNameGfnSiteGroupMap = "gfn_site_group_map"

func (*GfnSiteGroupMap) TableName() string {
	return TableNameGfnSiteGroupMap
}

type GfnSiteGroupMap struct {
	ID         int64     `json:"id"`
	SiteID     int64     `json:"siteId,string"`
	GroupID    int64     `json:"groupId,string"`
	Weight     int64     `json:"weight"`
	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `json:"updateTime"`
}

const TableNameGfnFeaturedSite = "gfn_featured_site"

func (*GfnFeaturedSite) TableName() string {
	return TableNameGfnFeaturedSite
}

type GfnFeaturedSite struct {
	ID         int64     `json:"id"`
	SiteID     int64     `json:"siteId,string"`
	Weight     int64     `json:"weight"`
	CreateTime time.Time `json:"createTime"`
	UpdateTime time.Time `json:"updateTime"`
}
