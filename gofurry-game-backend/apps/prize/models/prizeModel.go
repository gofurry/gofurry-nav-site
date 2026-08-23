package models

import (
	cm "github.com/gofurry/gofurry-game-backend/common/models"
)

const TableNameGfgPrize = "gfg_prize"

// GfgPrize mapped from table <gfg_prize>
type GfgPrize struct {
	ID         int64        `db:"id" json:"id"`                  // 抽奖活动表id
	Title      string       `db:"title" json:"title"`            // 标题
	Desc       string       `db:"desc" json:"desc"`              // 描述
	Prize      string       `db:"prize" json:"prize"`            // 奖品
	Key        string       `db:"key" json:"key"`                // 参与密钥
	StartTime  cm.LocalTime `db:"start_time" json:"startTime"`   // 开始时间
	EndTime    cm.LocalTime `db:"end_time" json:"endTime"`       // 结束时间
	CreateTime cm.LocalTime `db:"create_time" json:"createTime"` // 创建时间
	Status     bool         `db:"status" json:"status"`          // 状态
}

// TableName GfgPrize's table name
func (*GfgPrize) TableName() string {
	return TableNameGfgPrize
}

const TableNameGfgPrizeMember = "gfg_prize_member"

// GfgPrizeMember mapped from table <gfg_prize_member>
type GfgPrizeMember struct {
	ID         int64        `db:"id" json:"id"`                   // 抽奖活动参与表id
	PrizeID    int64        `db:"prize_id" json:"prizeId,string"` // 抽奖活动id
	Name       string       `db:"name" json:"name"`               // 参与者名称
	Email      string       `db:"email" json:"email"`             // 参与者邮箱
	IP         string       `db:"ip" json:"ip"`                   // 参与者ip
	Agent      string       `db:"agent" json:"agent"`             // User-Agent
	IsWinner   bool         `db:"is_winner" json:"isWinner"`      // 是否获奖
	PrizeKey   *string      `db:"prize_key" json:"prizeKey"`      // 获奖key
	CreateTime cm.LocalTime `db:"create_time" json:"createTime"`  // 创建时间
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
