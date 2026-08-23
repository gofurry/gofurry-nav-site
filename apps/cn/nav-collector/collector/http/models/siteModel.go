package models

import (
	cm "github.com/gofurry/gofurry-nav-collector/common/models"
)

const TableNameGfnCollectorDomain = "gfn_collector_domain"
const TableNameGfnSite = "gfn_site"

// GfnCollectorDomain mapped from table <gfn_collector_domain>
type GfnCollectorDomain struct {
	ID      int64   `db:"id" json:"id"`           // 域名请求表id
	SiteID  int64   `db:"site_id" json:"site_id"` // 关联站点 id
	Name    string  `db:"name" json:"name"`       // 域名
	Proxy   string  `db:"proxy" json:"proxy"`     // 是否需要代理加速 1 0
	Prefix  *string `db:"prefix" json:"prefix"`   // 是否有前缀
	TLS     string  `db:"tls" json:"tls"`         // 是否 https 1 0
	Deleted bool    `db:"deleted" json:"deleted"` // 软删除
}

// TableName GfnCollectorDomain's table name
func (*GfnCollectorDomain) TableName() string {
	return TableNameGfnCollectorDomain
}

func (d GfnCollectorDomain) TargetName() string {
	if d.Prefix == nil {
		return d.Name
	}
	return *d.Prefix + d.Name
}

const TableNameGfnCollectorLogHTTP = "gfn_collector_log_http"

// GfnCollectorLogHTTP mapped from table <gfn_collector_log_http>
type GfnCollectorLogHTTP struct {
	ID         int64        `db:"id" json:"id"`                  // http请求日志表
	Name       string       `db:"name" json:"name"`              // 域名
	Info       string       `db:"info" json:"info"`              // 日志内容
	Status     string       `db:"status" json:"status"`          // 请求状态 success failure
	CreateTime cm.LocalTime `db:"create_time" json:"createTime"` // 请求时间
}

// TableName GfnCollectorLogHTTP's table name
func (*GfnCollectorLogHTTP) TableName() string {
	return TableNameGfnCollectorLogHTTP
}
