package models

import (
	cm "github.com/gofurry/gofurry-nav-collector/common/models"
)

const TableNameGfnSite = "gfn_site"
const TableNameGfnCollectorDomain = "gfn_collector_domain"

// TableName GfnSite's table name
func (*GfnSite) TableName() string {
	return TableNameGfnSite
}

// GfnSite mapped from table <gfn_site>
type GfnSite struct {
	ID         int64        `db:"id" json:"id"`                  // 站点表id
	Name       string       `db:"name" json:"name"`              // 站点名称
	NameEn     string       `db:"name_en" json:"nameEn"`         // 站点名称-英文
	Info       string       `db:"info" json:"info"`              // 站点描述
	InfoEn     string       `db:"info_en" json:"infoEn"`         // 站点描述-英文
	CreateTime cm.LocalTime `db:"create_time" json:"createTime"` // 创建时间
	UpdateTime cm.LocalTime `db:"update_time" json:"updateTime"` // 修改时间
	Country    *string      `db:"country" json:"country"`        // 站点所属国家
	Nsfw       *string      `db:"nsfw" json:"nsfw"`              // 是否NSFW 1 0
	Welfare    *string      `db:"welfare" json:"welfare"`        // 是否公益项目 1 0
	Icon       *string      `db:"icon" json:"icon"`              // 站点图标
	Deleted    bool         `db:"deleted" json:"deleted"`
}

// GfnCollectorDomain mapped from table <gfn_collector_domain>.
type GfnCollectorDomain struct {
	ID      int64   `db:"id" json:"id"`
	SiteID  int64   `db:"site_id" json:"site_id"`
	Name    string  `db:"name" json:"name"`
	Proxy   string  `db:"proxy" json:"proxy"`
	Prefix  *string `db:"prefix" json:"prefix"`
	TLS     string  `db:"tls" json:"tls"`
	Deleted bool    `db:"deleted" json:"deleted"`
}

func (*GfnCollectorDomain) TableName() string {
	return TableNameGfnCollectorDomain
}

type PingTarget struct {
	SiteID int64
	Domain string
}

const TableNameGfnCollectorLogPing = "gfn_collector_log_ping"

// GfnCollectorLogPing mapped from table <gfn_collector_log_ping>
type GfnCollectorLogPing struct {
	ID         int64        `db:"id" json:"id"`                  // ping记录表id
	Name       string       `db:"name" json:"name"`              // 域名
	Delay      string       `db:"delay" json:"delay"`            // 延迟
	Loss       string       `db:"loss" json:"loss"`              // 丢包
	Status     string       `db:"status" json:"status"`          // 可达性 up down
	CreateTime cm.LocalTime `db:"create_time" json:"createTime"` // 日志时间
}

// TableName GfnCollectorLogPing's table name
func (*GfnCollectorLogPing) TableName() string {
	return TableNameGfnCollectorLogPing
}
