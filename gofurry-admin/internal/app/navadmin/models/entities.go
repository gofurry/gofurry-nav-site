package models

import (
	"time"

	pkgmodels "github.com/gofurry/awesome-fiber-template/v3/medium/pkg/models"
)

type Saying struct {
	ID         int64     `gorm:"column:id;primaryKey" json:"id"`
	Author     *string   `gorm:"column:author" json:"author"`
	Saying     string    `gorm:"column:saying;not null" json:"saying"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime" json:"update_time"`
}

func (*Saying) TableName() string { return "gfn_saying" }

type LogUpdate struct {
	ID         int64               `gorm:"column:id;primaryKey" json:"id"`
	Title      string              `gorm:"column:title;not null" json:"title"`
	URL        string              `gorm:"column:url;not null" json:"url"`
	CreateTime pkgmodels.LocalTime `gorm:"column:create_time;autoCreateTime" json:"create_time"`
	UpdateTime pkgmodels.LocalTime `gorm:"column:update_time;autoUpdateTime" json:"update_time"`
	Deleted    bool                `gorm:"column:deleted" json:"deleted"`
}

func (*LogUpdate) TableName() string { return "gfn_log_update" }

type CollectorDomain struct {
	ID      int64   `gorm:"column:id;primaryKey" json:"id"`
	SiteID  int64   `gorm:"column:site_id" json:"site_id"`
	Name    string  `gorm:"column:name;not null" json:"name"`
	Proxy   string  `gorm:"column:proxy;not null" json:"proxy"`
	Prefix  *string `gorm:"column:prefix" json:"prefix"`
	TLS     string  `gorm:"column:tls;not null" json:"tls"`
	Deleted bool    `gorm:"column:deleted" json:"deleted"`
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
	ID         int64               `gorm:"column:id;primaryKey" json:"id"`
	Name       string              `gorm:"column:name;not null" json:"name"`
	NameEn     string              `gorm:"column:name_en;not null" json:"name_en"`
	Info       string              `gorm:"column:info;not null" json:"info"`
	InfoEn     string              `gorm:"column:info_en;not null" json:"info_en"`
	CreateTime pkgmodels.LocalTime `gorm:"column:create_time;autoCreateTime" json:"create_time"`
	UpdateTime pkgmodels.LocalTime `gorm:"column:update_time;autoUpdateTime" json:"update_time"`
	Country    *string             `gorm:"column:country" json:"country"`
	Nsfw       string              `gorm:"column:nsfw" json:"nsfw"`
	Welfare    string              `gorm:"column:welfare" json:"welfare"`
	Icon       *string             `gorm:"column:icon" json:"icon"`
	Deleted    bool                `gorm:"column:deleted" json:"deleted"`
}

func (*Site) TableName() string { return "gfn_site" }

type SiteGroup struct {
	ID         int64     `gorm:"column:id;primaryKey" json:"id"`
	Name       string    `gorm:"column:name;not null" json:"name"`
	NameEn     string    `gorm:"column:name_en;not null" json:"name_en"`
	Info       string    `gorm:"column:info;not null" json:"info"`
	InfoEn     string    `gorm:"column:info_en;not null" json:"info_en"`
	Priority   int64     `gorm:"column:priority;not null" json:"priority"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime" json:"update_time"`
}

func (*SiteGroup) TableName() string { return "gfn_site_group" }

type SiteGroupMap struct {
	ID         int64     `gorm:"column:id;primaryKey" json:"id"`
	SiteID     int64     `gorm:"column:site_id;not null" json:"site_id"`
	GroupID    int64     `gorm:"column:group_id;not null" json:"group_id"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime" json:"update_time"`
}

func (*SiteGroupMap) TableName() string { return "gfn_site_group_map" }

type SayingPayload struct {
	Author *string `json:"author"`
	Saying string  `json:"saying"`
}

type LogUpdatePayload struct {
	Title string `json:"title"`
	URL   string `json:"url"`
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
}

type SiteGroupMapDTO struct {
	ID         int64     `json:"id"`
	SiteID     int64     `json:"site_id"`
	GroupID    int64     `json:"group_id"`
	SiteName   string    `json:"site_name"`
	GroupName  string    `json:"group_name"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}
