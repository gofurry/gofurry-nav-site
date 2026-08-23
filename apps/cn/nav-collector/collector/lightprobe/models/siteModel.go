package models

const TableNameGfnCollectorDomain = "gfn_collector_domain"
const TableNameGfnSite = "gfn_site"

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

func (d GfnCollectorDomain) TargetName() string {
	if d.Prefix == nil {
		return d.Name
	}
	return *d.Prefix + d.Name
}
