package models

const TableNameGfnCollectorDomain = "gfn_collector_domain"
const TableNameGfnSite = "gfn_site"

type GfnCollectorDomain struct {
	ID      int64   `gorm:"column:id;type:bigint;primaryKey" json:"id"`
	SiteID  int64   `gorm:"column:site_id;type:bigint" json:"site_id"`
	Name    string  `gorm:"column:name;type:character varying(255);not null" json:"name"`
	Proxy   string  `gorm:"column:proxy;type:character varying(4);not null" json:"proxy"`
	Prefix  *string `gorm:"column:prefix;type:character varying(255)" json:"prefix"`
	TLS     string  `gorm:"column:tls;type:character varying(4);not null" json:"tls"`
	Deleted bool    `gorm:"column:deleted;type:boolean" json:"deleted"`
}

func (*GfnCollectorDomain) TableName() string {
	return TableNameGfnCollectorDomain
}

func (d GfnCollectorDomain) TargetName() string {
	if d.Prefix == nil {
		return d.Name
	}
	return *d.Prefix + d.Name
}
