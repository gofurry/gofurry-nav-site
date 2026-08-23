package models

import "time"

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

const TableNameGfnCollectorLogDn = "gfn_collector_log_dns"

// GfnCollectorLogDn mapped from table <gfn_collector_log_dns>
type GfnCollectorLogDn struct {
	ID         int64     `db:"id" json:"id"`                  // DNS日志表 id
	Name       string    `db:"name" json:"name"`              // 域名
	A          *string   `db:"a" json:"a"`                    // A记录
	Aaaa       *string   `db:"aaaa" json:"aaaa"`              // AAAA记录
	Mx         *string   `db:"mx" json:"mx"`                  // MX记录
	Ns         *string   `db:"ns" json:"ns"`                  // NS记录
	Soa        *string   `db:"soa" json:"soa"`                // SOA记录
	Txt        *string   `db:"txt" json:"txt"`                // TXT记录
	Caa        *string   `db:"caa" json:"caa"`                // CAA记录
	Cname      *string   `db:"cname" json:"cname"`            // CNAME记录
	Status     string    `db:"status" json:"status"`          // 采集状态 success failure
	CreateTime time.Time `db:"create_time" json:"createTime"` // 采集时间
}

// TableName GfnCollectorLogDn's table name
func (*GfnCollectorLogDn) TableName() string {
	return TableNameGfnCollectorLogDn
}
