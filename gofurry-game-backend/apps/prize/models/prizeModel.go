package models

import (
	cm "github.com/gofurry/gofurry-game-backend/common/models"
)

const TableNameGfgPrize = "gfg_prize"

// GfgPrize mapped from table <gfg_prize>
type GfgPrize struct {
	ID         int64        `gorm:"column:id;type:bigint;primaryKey;comment:抽奖活动表id" json:"id"`                                       // 抽奖活动表id
	Title      string       `gorm:"column:title;type:character varying(100);not null;comment:标题" json:"title"`                        // 标题
	Desc       string       `gorm:"column:desc;type:text;not null;comment:描述" json:"desc"`                                            // 描述
	Prize      string       `gorm:"column:prize;type:jsonb;not null;comment:奖品" json:"prize"`                                         // 奖品
	Key        string       `gorm:"column:key;type:character varying(255);not null;comment:参与密钥" json:"key"`                          // 参与密钥
	StartTime  cm.LocalTime `gorm:"column:start_time;type:timestamp(0) without time zone;not null;comment:开始时间" json:"startTime"`     // 开始时间
	EndTime    cm.LocalTime `gorm:"column:end_time;type:timestamp(0) without time zone;not null;comment:结束时间" json:"endTime"`         // 结束时间
	CreateTime cm.LocalTime `gorm:"column:create_time;type:int;type:unsigned;not null;autoCreateTime;comment:创建时间" json:"createTime"` // 创建时间
	Status     bool         `gorm:"column:status;type:boolean;not null;comment:状态" json:"status"`                                     // 状态
}

// TableName GfgPrize's table name
func (*GfgPrize) TableName() string {
	return TableNameGfgPrize
}

const TableNameGfgPrizeMember = "gfg_prize_member"

// GfgPrizeMember mapped from table <gfg_prize_member>
type GfgPrizeMember struct {
	ID         int64        `gorm:"column:id;type:bigint;primaryKey;comment:抽奖活动参与表id" json:"id"`                                     // 抽奖活动参与表id
	PrizeID    int64        `gorm:"column:prize_id;type:bigint;not null;comment:抽奖活动id" json:"prizeId,string"`                        // 抽奖活动id
	Name       string       `gorm:"column:name;type:character varying(50);not null;comment:参与者名称" json:"name"`                        // 参与者名称
	Email      string       `gorm:"column:email;type:character varying(255);not null;comment:参与者邮箱" json:"email"`                     // 参与者邮箱
	IP         string       `gorm:"column:ip;type:character varying(50);not null;comment:参与者ip" json:"ip"`                            // 参与者ip
	Agent      string       `gorm:"column:agent;type:character varying(700);not null;comment:User-Agent" json:"agent"`                // User-Agent
	IsWinner   bool         `gorm:"column:is_winner;type:boolean;not null;comment:是否获奖" json:"isWinner"`                              // 是否获奖
	PrizeKey   *string      `gorm:"column:prize_key;type:character varying(255);comment:获奖key" json:"prizeKey"`                       // 获奖key
	CreateTime cm.LocalTime `gorm:"column:create_time;type:int;type:unsigned;not null;autoCreateTime;comment:创建时间" json:"createTime"` // 创建时间
}

// TableName GfgPrizeMember's table name
func (*GfgPrizeMember) TableName() string {
	return TableNameGfgPrizeMember
}

type PrizeParticipationRequest struct {
	ID    int64  `json:"id" validate:"required"`
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email,min=0,max=255"`
	Key   string `json:"key" validate:"required"`
}

type ParticipationCacheSaveModel struct {
	PrizeId int64  `json:"prize_id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	IP      string `json:"ip"`
	Agent   string `json:"agent"`
}

type PrizeModel struct {
	Keys     []string `json:"keys"`
	Title    string   `json:"title"`    // 奖品名称
	Platform string   `json:"platform"` // 奖品兑换平台
}

type PrizeCacheModel struct {
	ID      int64        `json:"id"`
	Title   string       `json:"title"`
	Desc    string       `json:"desc"`
	EndTime cm.LocalTime `json:"end_time"`
	Prize   string       `json:"prize"`
}

type WinnerCacheModel struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type PrizeCacheSaveModel struct {
	Name    string       `json:"name"`
	Desc    string       `json:"desc"`
	EndTime cm.LocalTime `json:"end_time"`
	Prize   struct {
		Title    string `json:"title"`
		Platform string `json:"platform"`
		Count    int    `json:"count"`
	} `json:"prize"`
	Winner []WinnerCacheModel `json:"winner"`
	Count  int                `json:"count"`
}

type PrizeWinnerCacheSaveModel struct {
	Prize      []PrizeCacheSaveModel `json:"prize"`
	PrizeCount int                   `json:"prize_count"`
}

type LotteryResp struct {
	History PrizeWinnerCacheSaveModel `json:"history"`
	Active  []ActiveVo                `json:"active"`
}

type ActiveLotteryVo struct {
	ID        int64        `json:"id"`
	Title     string       `json:"title"`
	Desc      string       `json:"desc"`
	StartTime cm.LocalTime `json:"start_time"`
	EndTime   cm.LocalTime `json:"end_time"`
	Prize     string       `json:"prize"`
}

type ActiveVo struct {
	Lottery LotteryVo          `json:"lottery"`
	Member  []WinnerCacheModel `json:"member"`
	Count   int                `json:"count"`
}

type LotteryVo struct {
	ID        int64        `json:"id"`
	Title     string       `json:"title"`
	Desc      string       `json:"desc"`
	StartTime cm.LocalTime `json:"start_time"`
	EndTime   cm.LocalTime `json:"end_time"`
	Prize     struct {
		Title    string `json:"title"`
		Platform string `json:"platform"`
		Count    int    `json:"count"`
	} `json:"prize"`
}
