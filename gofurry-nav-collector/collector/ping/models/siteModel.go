package models

import (
	cm "github.com/gofurry/gofurry-nav-collector/common/models"
)

const TableNameGfnSite = "gfn_site"

// TableName GfnSite's table name
func (*GfnSite) TableName() string {
	return TableNameGfnSite
}

// GfnSite mapped from table <gfn_site>
type GfnSite struct {
	ID         int64        `gorm:"column:id;type:bigint;primaryKey;comment:站点表id" json:"id"`                                           // 站点表id
	Name       string       `gorm:"column:name;type:character varying(255);not null;comment:站点名称" json:"name"`                          // 站点名称
	NameEn     string       `gorm:"column:name_en;type:character varying(255);not null;comment:站点名称-英文" json:"nameEn"`                  // 站点名称-英文
	Domain     string       `gorm:"column:domain;type:json;not null;comment:站点域名" json:"domain"`                                        // 站点域名
	Info       string       `gorm:"column:info;type:text;not null;comment:站点描述" json:"info"`                                            // 站点描述
	InfoEn     string       `gorm:"column:info_en;type:text;not null;comment:站点描述-英文" json:"infoEn"`                                    // 站点描述-英文
	CreateTime cm.LocalTime `gorm:"column:create_time;type:int;type:unsigned;not null;autoCreateTime;comment:创建时间" json:"createTime"`   // 创建时间
	UpdateTime cm.LocalTime `gorm:"column:update_time;type:int;type:unsigned;not null;autoUpdateTime;comment:修改时间" json:"updateTime"`   // 修改时间
	Country    *string      `gorm:"column:country;type:character varying(20);comment:站点所属国家" json:"country"`                            // 站点所属国家
	Nsfw       *string      `gorm:"column:nsfw;type:character varying(4);default:''::character varying;comment:是否NSFW 1 0" json:"nsfw"` // 是否NSFW 1 0
	Welfare    *string      `gorm:"column:welfare;type:character varying(4);comment:是否公益项目 1 0" json:"welfare"`                         // 是否公益项目 1 0
	Icon       *string      `gorm:"column:icon;type:character varying(255);comment:站点图标" json:"icon"`                                   // 站点图标
	Deleted    bool         `gorm:"column:deleted;type:boolean;comment:软删除" json:"deleted"`
}

type Domain struct {
	ID     int64  `gorm:"column:id;type:bigint;primaryKey;comment:站点表id" json:"id"`
	Domain string `gorm:"column:domain;type:json;not null;comment:站点域名" json:"domain"`
}

type Domains struct {
	Domain []string `json:"domain"`
}

type PingTarget struct {
	SiteID int64
	Domain string
}

const TableNameGfnCollectorLogPing = "gfn_collector_log_ping"

// GfnCollectorLogPing mapped from table <gfn_collector_log_ping>
type GfnCollectorLogPing struct {
	ID         int64        `gorm:"column:id;type:bigint;primaryKey;comment:ping记录表id" json:"id"`                                     // ping记录表id
	Name       string       `gorm:"column:name;type:character varying(255);not null;comment:域名" json:"name"`                          // 域名
	Delay      string       `gorm:"column:delay;type:character varying(20);not null;comment:延迟" json:"delay"`                         // 延迟
	Loss       string       `gorm:"column:loss;type:character varying(20);not null;comment:丢包" json:"loss"`                           // 丢包
	Status     string       `gorm:"column:status;type:character varying(20);not null;comment:可达性 up down" json:"status"`              // 可达性 up down
	CreateTime cm.LocalTime `gorm:"column:create_time;type:int;type:unsigned;not null;autoCreateTime;comment:日志时间" json:"createTime"` // 日志时间
}

// TableName GfnCollectorLogPing's table name
func (*GfnCollectorLogPing) TableName() string {
	return TableNameGfnCollectorLogPing
}
