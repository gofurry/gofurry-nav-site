package models

import (
	"time"

	cm "github.com/gofurry/gofurry-nav-backend/common/models"
)

const TableNameGfnSite = "gfn_site"

func (*GfnSite) TableName() string {
	return TableNameGfnSite
}

type GfnSite struct {
	ID         int64        `gorm:"column:id;type:bigint;primaryKey;comment:site id" json:"id"`
	Name       string       `gorm:"column:name;type:character varying(255);not null;comment:site name" json:"name"`
	NameEn     string       `gorm:"column:name_en;type:character varying(255);not null;comment:site name en" json:"nameEn"`
	Domain     string       `gorm:"column:domain;type:json;not null;comment:site domain" json:"domain"`
	Info       string       `gorm:"column:info;type:text;not null;comment:site info" json:"info"`
	InfoEn     string       `gorm:"column:info_en;type:text;not null;comment:site info en" json:"infoEn"`
	CreateTime cm.LocalTime `gorm:"column:create_time;type:int;type:unsigned;not null;autoCreateTime;comment:create time" json:"createTime"`
	UpdateTime cm.LocalTime `gorm:"column:update_time;type:int;type:unsigned;not null;autoUpdateTime;comment:update time" json:"updateTime"`
	Country    *string      `gorm:"column:country;type:character varying(20);comment:country" json:"country"`
	Nsfw       string       `gorm:"column:nsfw;type:character varying(4);default:''::character varying;comment:nsfw" json:"nsfw"`
	Welfare    string       `gorm:"column:welfare;type:character varying(4);comment:welfare" json:"welfare"`
	ViewCount  int64        `gorm:"column:view_count;type:bigint;not null;default:0;comment:view count" json:"view_count"`
	Icon       *string      `gorm:"column:icon;type:character varying(255);comment:icon" json:"icon"`
	Deleted    bool         `gorm:"column:deleted;type:boolean;comment:deleted" json:"deleted"`
}

const TableNameGfnSiteGroup = "gfn_site_group"

func (*GfnSiteGroup) TableName() string {
	return TableNameGfnSiteGroup
}

type GfnSiteGroup struct {
	ID         int64     `gorm:"column:id;type:bigint;primaryKey;comment:group id" json:"id"`
	Name       string    `gorm:"column:name;type:character varying(255);not null;comment:group name" json:"name"`
	NameEn     string    `gorm:"column:name_en;type:character varying(255);not null;comment:group name en" json:"nameEn"`
	Info       string    `gorm:"column:info;type:character varying(255);not null;comment:group info" json:"info"`
	InfoEn     string    `gorm:"column:info_en;type:character varying(255);not null;comment:group info en" json:"infoEn"`
	Priority   int64     `gorm:"column:priority;type:bigint;not null;comment:priority" json:"priority"`
	CreateTime time.Time `gorm:"column:create_time;type:int;type:unsigned;not null;autoCreateTime;comment:create time" json:"createTime"`
	UpdateTime time.Time `gorm:"column:update_time;type:int;type:unsigned;not null;autoUpdateTime;comment:update time" json:"updateTime"`
}

const TableNameGfnSiteGroupMap = "gfn_site_group_map"

func (*GfnSiteGroupMap) TableName() string {
	return TableNameGfnSiteGroupMap
}

type GfnSiteGroupMap struct {
	ID         int64     `gorm:"column:id;type:bigint;primaryKey;comment:group map id" json:"id"`
	SiteID     int64     `gorm:"column:site_id;type:bigint;not null;comment:site id" json:"siteId,string"`
	GroupID    int64     `gorm:"column:group_id;type:bigint;not null;comment:group id" json:"groupId,string"`
	CreateTime time.Time `gorm:"column:create_time;type:int;type:unsigned;not null;autoCreateTime;comment:create time" json:"createTime"`
	UpdateTime time.Time `gorm:"column:update_time;type:int;type:unsigned;not null;autoUpdateTime;comment:update time" json:"updateTime"`
}
