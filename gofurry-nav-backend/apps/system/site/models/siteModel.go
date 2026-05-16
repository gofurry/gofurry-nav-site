package models

import (
	cm "github.com/gofurry/gofurry-nav-backend/common/models"
)

const TableNameGfnLogUpdate = "gfn_log_update"

// GfnLogUpdate mapped from table <gfn_log_update>
type GfnLogUpdate struct {
	ID         int64        `gorm:"column:id;type:bigint;primaryKey;comment:更新公告表id" json:"id"`                                       // 更新公告表id
	Title      string       `gorm:"column:title;type:character varying(100);not null;comment:更新公告标题" json:"title"`                    // 更新公告标题
	URL        string       `gorm:"column:url;type:character varying(255);not null;comment:更新公告文档地址" json:"url"`                      // 更新公告文档地址
	CreateTime cm.LocalTime `gorm:"column:create_time;type:int;type:unsigned;not null;autoCreateTime;comment:创建时间" json:"createTime"` // 创建时间
	UpdateTime cm.LocalTime `gorm:"column:update_time;type:int;type:unsigned;not null;autoUpdateTime;comment:更新时间" json:"updateTime"` // 更新时间
	Deleted    bool         `gorm:"column:deleted;type:boolean;not null;comment:软删除" json:"deleted"`                                  // 软删除
}

// TableName GfnLogUpdate's table name
func (*GfnLogUpdate) TableName() string {
	return TableNameGfnLogUpdate
}

type ChangeLogVo struct {
	Title      string       `json:"title"`
	URL        string       `json:"url"`
	CreateTime cm.LocalTime `json:"create_time"`
	UpdateTime cm.LocalTime `json:"update_time"`
}
