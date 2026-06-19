package models

import cm "github.com/gofurry/gofurry-game-backend/common/models"

const TableNameGfgGame = "gfg_game"

// GfgGame is the site-maintained game profile table.
type GfgGame struct {
	ID           int64        `gorm:"column:id;type:bigint;primaryKey;comment:游戏表ID" json:"id"`
	Name         string       `gorm:"column:name;type:character varying(255);not null;comment:游戏名称" json:"name"`
	NameEn       string       `gorm:"column:name_en;type:character varying(255);not null;comment:游戏英文名称" json:"nameEn"`
	Info         string       `gorm:"column:info;type:character varying(300);not null;comment:游戏简介" json:"info"`
	InfoEn       string       `gorm:"column:info_en;type:character varying(300);not null;comment:游戏英文简介" json:"infoEn"`
	CreateTime   cm.LocalTime `gorm:"column:create_time;type:int;type:unsigned;not null;autoCreateTime;comment:创建时间" json:"createTime"`
	UpdateTime   cm.LocalTime `gorm:"column:update_time;type:int;type:unsigned;not null;autoUpdateTime;comment:更新时间" json:"updateTime"`
	Resources    *string      `gorm:"column:resources;type:json;comment:游戏相关资源" json:"resources"`
	Groups       *string      `gorm:"column:groups;type:json;comment:游戏相关社群" json:"groups"`
	ReleaseDate  string       `gorm:"column:release_date;type:character varying(255);not null;comment:发行日期" json:"releaseDate"`
	Developers   string       `gorm:"column:developers;type:json;not null;comment:开发商" json:"developers"`
	Publishers   string       `gorm:"column:publishers;type:json;not null;comment:发行商" json:"publishers"`
	Appid        int64        `gorm:"column:appid;type:bigint;not null;comment:SteamAPI appid" json:"appid"`
	Header       string       `gorm:"column:header;type:character varying(255);not null;comment:游戏封面图" json:"header"`
	Links        *string      `gorm:"column:links;type:json;comment:三方网站链接" json:"links"`
	Weight       int64        `gorm:"column:weight;type:bigint;not null;comment:权重" json:"weight"`
	PrimaryTag   int64        `gorm:"column:primary_tag;type:bigint;not null;comment:主标签" json:"primaryTag"`
	SecondaryTag int64        `gorm:"column:secondary_tag;type:bigint;not null;comment:次标签" json:"secondaryTag"`
}

func (*GfgGame) TableName() string {
	return TableNameGfgGame
}
