package models

import (
	"github.com/gofurry/gofurry-user/common/abstract"
	cm "github.com/gofurry/gofurry-user/common/models"
)

const TableNameGfUserOauth = "gf_user_oauth"

// GfUserOauth mapped from table <gf_user_oauth>
type GfUserOauth struct {
	abstract.IdModel
	UserID     int64        `gorm:"column:user_id;type:bigint;not null;comment:用户表id" json:"userId,string"`                           // 用户表id
	Provider   string       `gorm:"column:provider;type:character varying(50);not null;comment:三方平台名称" json:"provider"`               // 三方平台名称
	OpenID     string       `gorm:"column:open_id;type:character varying(255);not null;comment:三方唯一标识" json:"openId,string"`          // 三方唯一标识
	CreateTime cm.LocalTime `gorm:"column:create_time;type:int;type:unsigned;not null;autoCreateTime;comment:创建时间" json:"createTime"` // 创建时间
}

// TableName GfUserOauth's table name
func (*GfUserOauth) TableName() string {
	return TableNameGfUserOauth
}
