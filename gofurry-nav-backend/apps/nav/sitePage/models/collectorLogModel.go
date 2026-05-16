package models

import (
	cm "github.com/gofurry/gofurry-nav-backend/common/models"
)

const TableNameGfnCollectorLogPing = "gfn_collector_log_ping"

// GfnCollectorLogPing mapped from table <gfn_collector_log_ping>
type GfnCollectorLogPing struct {
	ID         int64        `gorm:"column:id;type:bigint;primaryKey;comment:ping记录表id" json:"id"`                                     // ping记录表id
	Name       string       `gorm:"column:name;type:character varying(255);not null;comment:域名" json:"name"`                          // 域名
	Delay      string       `gorm:"column:delay;type:character varying(20);not null;comment:延迟" json:"delay"`                         // 延迟
	Loss       string       `gorm:"column:loss;type:character varying(20);not null;comment:丢包" json:"loss"`                           // 丢包
	Status     string       `gorm:"column:status;type:character varying(20);not null;comment:可达性 up down" json:"status"`              // 可达性 up down
	CreateTime cm.LocalTime `gorm:"column:create_time;type:int;type:unsigned;not null;autoCreateTime;comment:日志时间" json:"createTime"` // 日志时间
}

// TableName GfnCollectorLogPing's table name
func (*GfnCollectorLogPing) TableName() string {
	return TableNameGfnCollectorLogPing
}
