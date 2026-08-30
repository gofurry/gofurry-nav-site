package models

import (
	"time"

	pkgmodels "github.com/gofurry/gofurry-admin/pkg/models"
)

type Saying struct {
	ID         int64     `json:"id"`
	Author     *string   `json:"author"`
	Language   string    `json:"language"`
	Saying     string    `json:"saying"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

func (*Saying) TableName() string { return "gfn_saying" }

type UpdateNotice struct {
	ID          int64               `json:"id"`
	Title       string              `json:"title"`
	TitleEn     string              `json:"title_en"`
	Body        string              `json:"body"`
	BodyEn      string              `json:"body_en"`
	PublishedAt pkgmodels.LocalTime `json:"published_at"`
	CreateTime  pkgmodels.LocalTime `json:"create_time"`
	UpdateTime  pkgmodels.LocalTime `json:"update_time"`
	Deleted     bool                `json:"deleted"`
}

func (*UpdateNotice) TableName() string { return "gfn_nav_update_notice" }

type CollectorDomain struct {
	ID      int64   `json:"id"`
	SiteID  int64   `json:"site_id"`
	Name    string  `json:"name"`
	Proxy   string  `json:"proxy"`
	Prefix  *string `json:"prefix"`
	TLS     string  `json:"tls"`
	Deleted bool    `json:"deleted"`
}

func (*CollectorDomain) TableName() string { return "gfn_collector_domain" }

type CollectorDomainDTO struct {
	ID       int64   `json:"id"`
	SiteID   int64   `json:"site_id"`
	SiteName string  `json:"site_name"`
	Name     string  `json:"name"`
	Proxy    string  `json:"proxy"`
	Prefix   *string `json:"prefix"`
	TLS      string  `json:"tls"`
	Deleted  bool    `json:"deleted"`
}

type Site struct {
	ID         int64               `json:"id"`
	Name       string              `json:"name"`
	NameEn     string              `json:"name_en"`
	Info       string              `json:"info"`
	InfoEn     string              `json:"info_en"`
	CreateTime pkgmodels.LocalTime `json:"create_time"`
	UpdateTime pkgmodels.LocalTime `json:"update_time"`
	Country    *string             `json:"country"`
	Nsfw       string              `json:"nsfw"`
	Welfare    string              `json:"welfare"`
	Icon       *string             `json:"icon"`
	Deleted    bool                `json:"deleted"`
}

func (*Site) TableName() string { return "gfn_site" }

type SiteGroup struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	NameEn     string    `json:"name_en"`
	Info       string    `json:"info"`
	InfoEn     string    `json:"info_en"`
	Priority   int64     `json:"priority"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

func (*SiteGroup) TableName() string { return "gfn_site_group" }

type SiteGroupMap struct {
	ID         int64     `json:"id"`
	SiteID     int64     `json:"site_id"`
	GroupID    int64     `json:"group_id"`
	Weight     int64     `json:"weight"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

func (*SiteGroupMap) TableName() string { return "gfn_site_group_map" }

type FeaturedSite struct {
	ID         int64     `json:"id"`
	SiteID     int64     `json:"site_id"`
	Weight     int64     `json:"weight"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

func (*FeaturedSite) TableName() string { return "gfn_featured_site" }

type SayingPayload struct {
	Author   *string `json:"author"`
	Language string  `json:"language"`
	Saying   string  `json:"saying"`
}

type UpdateNoticePayload struct {
	Title       string `json:"title"`
	TitleEn     string `json:"title_en"`
	Body        string `json:"body"`
	BodyEn      string `json:"body_en"`
	PublishedAt string `json:"published_at"`
}

type CollectorDomainPayload struct {
	SiteID int64   `json:"site_id"`
	Name   string  `json:"name"`
	Proxy  string  `json:"proxy"`
	Prefix *string `json:"prefix"`
	TLS    string  `json:"tls"`
}

type SitePayload struct {
	Name    string  `json:"name"`
	NameEn  string  `json:"name_en"`
	Info    string  `json:"info"`
	InfoEn  string  `json:"info_en"`
	Country *string `json:"country"`
	Nsfw    string  `json:"nsfw"`
	Welfare string  `json:"welfare"`
	Icon    *string `json:"icon"`
}

type SiteDTO struct {
	ID         int64               `json:"id"`
	Name       string              `json:"name"`
	NameEn     string              `json:"name_en"`
	Info       string              `json:"info"`
	InfoEn     string              `json:"info_en"`
	CreateTime pkgmodels.LocalTime `json:"create_time"`
	UpdateTime pkgmodels.LocalTime `json:"update_time"`
	Country    *string             `json:"country"`
	Nsfw       string              `json:"nsfw"`
	Welfare    string              `json:"welfare"`
	Icon       *string             `json:"icon"`
	Deleted    bool                `json:"deleted"`
}

type SiteWorkspaceTarget struct {
	ID      int64   `json:"id"`
	SiteID  int64   `json:"site_id"`
	Name    string  `json:"name"`
	Proxy   string  `json:"proxy"`
	Prefix  *string `json:"prefix"`
	TLS     string  `json:"tls"`
	Primary bool    `json:"primary"`
}

type SiteWorkspaceSummary struct {
	ID            int64               `json:"id"`
	Name          string              `json:"name"`
	NameEn        string              `json:"name_en"`
	UpdateTime    pkgmodels.LocalTime `json:"update_time"`
	PrimaryTarget string              `json:"primary_target"`
	GroupNames    []string            `json:"group_names"`
	Featured      bool                `json:"featured"`
}

type SiteWorkspaceGroup struct {
	ID        int64  `json:"id"`
	SiteID    int64  `json:"site_id"`
	GroupID   int64  `json:"group_id"`
	GroupName string `json:"group_name"`
	Weight    int64  `json:"weight"`
}

type SiteWorkspaceFeatured struct {
	ID     int64 `json:"id"`
	SiteID int64 `json:"site_id"`
	Weight int64 `json:"weight"`
}

type SiteWorkspace struct {
	Site     SiteDTO                `json:"site"`
	Targets  []SiteWorkspaceTarget  `json:"targets"`
	Groups   []SiteWorkspaceGroup   `json:"groups"`
	Featured *SiteWorkspaceFeatured `json:"featured"`
}

type SiteGroupPayload struct {
	Name     string `json:"name"`
	NameEn   string `json:"name_en"`
	Info     string `json:"info"`
	InfoEn   string `json:"info_en"`
	Priority int64  `json:"priority"`
}

type SiteGroupMapPayload struct {
	SiteID  int64 `json:"site_id"`
	GroupID int64 `json:"group_id"`
	Weight  int64 `json:"weight"`
}

type FeaturedSitePayload struct {
	SiteID int64 `json:"site_id"`
	Weight int64 `json:"weight"`
}

type SiteGroupMapDTO struct {
	ID         int64     `json:"id"`
	SiteID     int64     `json:"site_id"`
	GroupID    int64     `json:"group_id"`
	SiteName   string    `json:"site_name"`
	GroupName  string    `json:"group_name"`
	Weight     int64     `json:"weight"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

type FeaturedSiteDTO struct {
	ID         int64     `json:"id"`
	SiteID     int64     `json:"site_id"`
	SiteName   string    `json:"site_name"`
	Weight     int64     `json:"weight"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}
