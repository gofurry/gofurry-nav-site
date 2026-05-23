package models

import "time"

const TableNameGfnCollectorDomain = "gfn_collector_domain"
const TableNameGfnSite = "gfn_site"

// GfnCollectorDomain mapped from table <gfn_collector_domain>
type GfnCollectorDomain struct {
	ID      int64   `gorm:"column:id;type:bigint;primaryKey;comment:域名请求表id" json:"id"`                        // 域名请求表id
	SiteID  int64   `gorm:"column:site_id;type:bigint;comment:关联站点 id" json:"site_id"`                         // 关联站点 id
	Name    string  `gorm:"column:name;type:character varying(255);not null;comment:域名" json:"name"`           // 域名
	Proxy   string  `gorm:"column:proxy;type:character varying(4);not null;comment:是否需要代理加速 1 0" json:"proxy"` // 是否需要代理加速 1 0
	Prefix  *string `gorm:"column:prefix;type:character varying(255);comment:是否有前缀" json:"prefix"`             // 是否有前缀
	TLS     string  `gorm:"column:tls;type:character varying(4);not null;comment:是否 https 1 0" json:"tls"`     // 是否 https 1 0
	Deleted bool    `gorm:"column:deleted;type:boolean;comment:软删除" json:"deleted"`                            // 软删除
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
	ID         int64     `gorm:"column:id;type:bigint;primaryKey;comment:DNS日志表 id" json:"id"`                                     // DNS日志表 id
	Name       string    `gorm:"column:name;type:character varying(255);not null;comment:域名" json:"name"`                          // 域名
	A          *string   `gorm:"column:a;type:json;comment:A记录" json:"a"`                                                          // A记录
	Aaaa       *string   `gorm:"column:aaaa;type:json;comment:AAAA记录" json:"aaaa"`                                                 // AAAA记录
	Mx         *string   `gorm:"column:mx;type:json;comment:MX记录" json:"mx"`                                                       // MX记录
	Ns         *string   `gorm:"column:ns;type:json;comment:NS记录" json:"ns"`                                                       // NS记录
	Soa        *string   `gorm:"column:soa;type:json;comment:SOA记录" json:"soa"`                                                    // SOA记录
	Txt        *string   `gorm:"column:txt;type:json;comment:TXT记录" json:"txt"`                                                    // TXT记录
	Caa        *string   `gorm:"column:caa;type:json;comment:CAA记录" json:"caa"`                                                    // CAA记录
	Cname      *string   `gorm:"column:cname;type:json;comment:CNAME记录" json:"cname"`                                              // CNAME记录
	Status     string    `gorm:"column:status;type:character varying(20);not null;comment:采集状态 success failure" json:"status"`     // 采集状态 success failure
	CreateTime time.Time `gorm:"column:create_time;type:int;type:unsigned;not null;autoCreateTime;comment:采集时间" json:"createTime"` // 采集时间
}

// TableName GfnCollectorLogDn's table name
func (*GfnCollectorLogDn) TableName() string {
	return TableNameGfnCollectorLogDn
}
