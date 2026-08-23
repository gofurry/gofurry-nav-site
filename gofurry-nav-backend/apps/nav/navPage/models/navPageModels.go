package models

import "time"

const TableNameGfnSaying = "gfn_saying"

// GfnSaying mapped from table <gfn_saying>
type GfnSaying struct {
	ID         int64     `json:"id"`         // 金句表ID
	Author     *string   `json:"author"`     // 金句提供者
	Language   string    `json:"language"`   // 语言
	Saying     string    `json:"saying"`     // 金句
	CreateTime time.Time `json:"createTime"` // 创建时间
	UpdateTime time.Time `json:"updateTime"` // 修改时间
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
