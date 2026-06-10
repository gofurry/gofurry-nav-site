package models

import (
	cm "github.com/gofurry/gofurry-game-backend/common/models"
)

const TableNameGfgTag = "gfg_tag"

// GfgTag mapped from table <gfg_tag>
type GfgTag struct {
	ID         int64        `gorm:"column:id;type:bigint;primaryKey;comment:标签表id" json:"id"`                                         // 标签表id
	Name       string       `gorm:"column:name;type:character varying(255);not null;comment:标签名称" json:"name"`                        // 标签名称
	NameEn     string       `gorm:"column:name_en;type:character varying(255);not null;comment:标签英文名称" json:"nameEn"`                 // 标签英文名称
	Info       string       `gorm:"column:info;type:character varying(255);not null;comment:标签简介" json:"info"`                        // 标签简介
	InfoEn     string       `gorm:"column:info_en;type:character varying(255);not null;comment:标签英文简介" json:"infoEn"`                 // 标签英文简介
	Prefix     int64        `gorm:"column:prefix;type:bigint;not null;comment:父标签 没有为-1" json:"prefix"`                               // 父标签 没有为-1
	CreateTime cm.LocalTime `gorm:"column:create_time;type:int;type:unsigned;not null;autoCreateTime;comment:创建时间" json:"createTime"` // 创建时间
	UpdateTime cm.LocalTime `gorm:"column:update_time;type:int;type:unsigned;not null;autoUpdateTime;comment:修改时间" json:"updateTime"` // 修改时间
}

// TableName GfgTag's table name
func (*GfgTag) TableName() string {
	return TableNameGfgTag
}

const TableNameGfgTagMap = "gfg_tag_map"

// GfgTagMap mapped from table <gfg_tag_map>
type GfgTagMap struct {
	ID         int64        `gorm:"column:id;type:bigint;primaryKey;comment:游戏标签映射表id" json:"id"`                                     // 游戏标签映射表id
	GameID     int64        `gorm:"column:game_id;type:bigint;not null;comment:游戏id" json:"gameId,string"`                            // 游戏id
	TagID      int64        `gorm:"column:tag_id;type:bigint;not null;comment:标签id" json:"tagId,string"`                              // 标签id
	CreateTime cm.LocalTime `gorm:"column:create_time;type:int;type:unsigned;not null;autoCreateTime;comment:创建时间" json:"createTime"` // 创建时间
	UpdateTime cm.LocalTime `gorm:"column:update_time;type:int;type:unsigned;not null;autoUpdateTime;comment:修改时间" json:"updateTime"` // 修改时间
}

// TableName GfgTagMap's table name
func (*GfgTagMap) TableName() string {
	return TableNameGfgTagMap
}
