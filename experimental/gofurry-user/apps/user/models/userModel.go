package models

import (
	"github.com/gofurry/gofurry-user/common/abstract"
	cm "github.com/gofurry/gofurry-user/common/models"
)

const TableNameGfUser = "gf_user"

// GfUser mapped from table <gf_user>
type GfUser struct {
	abstract.DefaultModel
	Nickname   string       `gorm:"column:nickname;type:character varying(60);not null;comment:用户名" json:"nickname"`                  // 用户名
	Email      *string      `gorm:"column:email;type:character varying(100);comment:用户邮箱" json:"email"`                               // 用户邮箱
	Oauth      bool         `gorm:"column:oauth;type:boolean;not null;comment:是否三方登录" json:"oauth"`                                   // 是否三方登录
	Password   string       `gorm:"column:password;type:character varying(255);not null;comment:用户密码" json:"password"`                // 用户密码
	Role       string       `gorm:"column:role;type:character varying(50);comment:用户身份" json:"role"`                                  // 用户身份
	Info       *string      `gorm:"column:info;type:character varying(255);comment:用户信息" json:"info"`                                 // 用户信息
	CreateTime cm.LocalTime `gorm:"column:create_time;type:int;type:unsigned;not null;autoCreateTime;comment:创建时间" json:"createTime"` // 创建时间
	UpdateTime cm.LocalTime `gorm:"column:update_time;type:int;type:unsigned;not null;autoUpdateTime;comment:更新时间" json:"updateTime"` // 更新时间
	Status     string       `gorm:"column:status;type:character varying(20);not null;comment:用户状态" json:"status"`                     // 用户状态
	Avatar     string       `gorm:"column:avatar;type:character varying(255);not null;comment:用户头像" json:"avatar"`                    // 用户头像
}

// TableName GfUser's table name
func (*GfUser) TableName() string {
	return TableNameGfUser
}

const TableNameGfLoginLog = "gf_login_log"

// GfLoginLog mapped from table <gf_login_log>
type GfLoginLog struct {
	abstract.IdModel
	UserID     int64        `gorm:"column:user_id;type:bigint;not null;comment:用户表id" json:"userId,string"`                             // 用户表id
	Agent      string       `gorm:"column:agent;type:character varying(255);not null;comment:浏览器信息" json:"agent"`                       // 浏览器信息
	IP         string       `gorm:"column:ip;type:character varying(255);not null;comment:登录 ip" json:"ip"`                             // 登录 ip
	CreateTime cm.LocalTime `gorm:"column:create_time;type:int;type:unsigned;not null;autoCreateTime;comment:日志创建时间" json:"createTime"` // 日志创建时间
	LoginType  string       `gorm:"column:login_type;type:character varying(20);not null;comment:记录登录方式" json:"loginType"`              // 记录登录方式
}

// TableName GfLoginLog's table name
func (*GfLoginLog) TableName() string {
	return TableNameGfLoginLog
}

type CurrentUser struct {
	Name string `json:"name"`
	ID   int64  `json:"id,string"`
}

type UserLoginRequest struct {
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UserRegisterRequest struct {
	Email    string `json:"email" validate:"required,email,min=1,max=100"`
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required"`
	Code     string `json:"code" validate:"required,len=6"`
	Role     string `json:"role" validate:"required"`
}
