package models

import "time"

const TableNameGfnSaying = "gfn_saying"

// GfnSaying mapped from table <gfn_saying>
type GfnSaying struct {
	ID         int64     `gorm:"column:id;type:bigint;primaryKey;comment:金句表ID" json:"id"`                                         // 金句表ID
	Author     *string   `gorm:"column:author;type:character varying(255);comment:金句提供者" json:"author"`                            // 金句提供者
	Language   string    `gorm:"column:language;type:character varying(8);not null;default:zh;comment:语言" json:"language"`         // 语言
	Saying     string    `gorm:"column:saying;type:character varying(255);not null;comment:金句" json:"saying"`                      // 金句
	CreateTime time.Time `gorm:"column:create_time;type:int;type:unsigned;not null;autoCreateTime;comment:创建时间" json:"createTime"` // 创建时间
	UpdateTime time.Time `gorm:"column:update_time;type:int;type:unsigned;not null;autoUpdateTime;comment:修改时间" json:"updateTime"` // 修改时间
}

// TableName GfnSaying's table name
func (*GfnSaying) TableName() string {
	return TableNameGfnSaying
}

type SiteVo struct {
	ID          string  `form:"id" json:"id"`
	Name        string  `form:"name" json:"name"`
	Domain      string  `form:"domain" json:"domain"`
	Info        string  `form:"info" json:"info"`
	Country     *string `form:"country" json:"country"`
	Nsfw        string  `form:"nsfw" json:"nsfw"`
	Welfare     string  `form:"welfare" json:"welfare"`
	Icon        *string `form:"icon" json:"icon"`
	GroupWeight int64   `form:"-" json:"-"`
	ViewCount   int64   `form:"view_count" json:"view_count"`
	CreateTime  string  `form:"create_time" json:"create_time"`
	UpdateTime  string  `form:"update_time" json:"update_time"`
}

type GroupVo struct {
	ID          string           `form:"id" json:"id"`
	Name        string           `form:"name" json:"name"`
	Info        string           `form:"info" json:"info"`
	Priority    int64            `form:"priority" json:"priority"`
	Sites       []string         `form:"sites" json:"sites"`
	SiteWeights map[string]int64 `form:"-" json:"-"`
}

type SayingModel struct {
	Author   *string `json:"author"`
	Content  string  `json:"content"`
	Language string  `json:"language"`
}

type FeaturedSiteVo struct {
	ID     string `json:"id"`
	SiteID string `json:"site_id"`
	Weight int64  `json:"weight"`
}
