package models

import (
	cm "github.com/gofurry/gofurry-game-backend/common/models"
)

const TableNameGfgGameComment = "gfg_game_comment"

// GfgGameComment mapped from table <gfg_game_comment>
type GfgGameComment struct {
	ID         int64        `db:"id" json:"id"`                  // 评论表ID
	Region     string       `db:"region" json:"region"`          // 地区
	Content    string       `db:"content" json:"content"`        // 评论
	Score      float64      `db:"score" json:"score"`            // 评分
	CreateTime cm.LocalTime `db:"create_time" json:"createTime"` // 创建时间
	GameID     int64        `db:"game_id" json:"gameId,string"`  // 游戏表ID
	IP         string       `db:"ip" json:"ip"`                  // ip
	Name       string       `db:"name" json:"name"`              // 评论人名
}

// TableName GfgGameComment's table name
func (*GfgGameComment) TableName() string {
	return TableNameGfgGameComment
}

type AvgScoreResult struct {
	GameID       string  `db:"game_id" json:"game_id"`
	AvgScore     float64 `db:"avg_score" json:"avg_score"`
	CommentCount int64   `db:"comment_count" json:"comment_count"`
	Name         string  `db:"name" json:"name"`
	NameEn       string  `db:"name_en" json:"name_en"`
	Info         string  `db:"info" json:"info"`
	InfoEn       string  `db:"info_en" json:"info_en"`
	Header       string  `db:"header" json:"header"`
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
