package models

import (
	cm "github.com/gofurry/gofurry-game-backend/common/models"
)

const TableNameGfgGameComment = "gfg_game_comment"

// GfgGameComment mapped from table <gfg_game_comment>
type GfgGameComment struct {
	ID         int64        `gorm:"column:id;type:bigint;primaryKey;comment:评论表ID" json:"id"`                                         // 评论表ID
	Region     string       `gorm:"column:region;type:character varying(50);not null;comment:地区" json:"region"`                       // 地区
	Content    string       `gorm:"column:content;type:text;not null;comment:评论" json:"content"`                                      // 评论
	Score      float64      `gorm:"column:score;type:double precision;not null;comment:评分" json:"score"`                              // 评分
	CreateTime cm.LocalTime `gorm:"column:create_time;type:int;type:unsigned;not null;autoCreateTime;comment:创建时间" json:"createTime"` // 创建时间
	GameID     int64        `gorm:"column:game_id;type:bigint;not null;comment:游戏表ID" json:"gameId,string"`                           // 游戏表ID
	IP         string       `gorm:"column:ip;type:character varying(50);not null;comment:ip" json:"ip"`                               // ip
	Name       string       `gorm:"column:name;type:character varying(50);comment:评论人名称" json:"name"`                                 // 评论人名
}

// TableName GfgGameComment's table name
func (*GfgGameComment) TableName() string {
	return TableNameGfgGameComment
}

type AvgScoreResult struct {
	GameID       string  `gorm:"column:game_id" json:"game_id"`
	AvgScore     float64 `gorm:"column:avg_score" json:"avg_score"`
	CommentCount int64   `gorm:"column:comment_count" json:"comment_count"`
	Name         string  `gorm:"column:name" json:"name"`
	NameEn       string  `gorm:"column:name_en" json:"name_en"`
	Info         string  `gorm:"column:info" json:"info"`
	InfoEn       string  `gorm:"column:info_en" json:"info_en"`
	Header       string  `gorm:"column:header" json:"header"`
}

type AnonymousReviewRequest struct {
	ID      string  `json:"id"` // 游戏 ID
	Content string  `json:"content"`
	Score   float64 `json:"score"`
	Name    string  `json:"name"`
}

type AnonymousReviewResponse struct {
	Region    string       `json:"region"`
	Score     float64      `json:"score"`
	Content   string       `json:"content"`
	IP        string       `json:"ip"`
	Time      cm.LocalTime `json:"time"`
	GameName  string       `json:"game_name"`
	GameCover string       `json:"game_cover"`
}
