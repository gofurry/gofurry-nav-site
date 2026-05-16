package models

import cm "github.com/gofurry/gofurry-game-backend/common/models"

type SearchGameVo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Info  string `json:"info"`
	Cover string `json:"cover"`
}

type SearchRequest struct {
	Txt  string `json:"txt"`
	Lang string `json:"lang"`
}

type SearchPageQueryRequest struct {
	cm.PageReq
	Content         *string      `json:"content"`        // 名称、内容、开发商、发行商
	PubStartTime    cm.LocalTime `json:"pub_start_time"` // 发售时间
	PubEndTime      cm.LocalTime `json:"pub_end_time"`
	UpdateStartTime cm.LocalTime `json:"update_start_time"` // 更新时间
	UpdateEndTime   cm.LocalTime `json:"update_end_time"`
	ScoreOrder      bool         `json:"score"`        // 评分排序
	RemarkOrder     bool         `json:"remark_order"` // 评论数排序
	TimeOrder       bool         `json:"time_order"`   // 更新日期排序
	TagList         []int64      `json:"tag_list"`     // 标签列表
	Lang            string       `json:"lang"`
}

type GamePageQueryVo struct {
	ID           string       `json:"id"` // 游戏 ID
	Name         string       `json:"name"`
	Info         string       `json:"info"`
	Cover        string       `json:"cover"` // 封面图
	Appid        int64        `json:"appid"`
	UpdateTime   cm.LocalTime `json:"update_time"`
	ReleaseDate  string       `json:"release_date"`
	RemarkCount  int          `json:"remark_count"`  // 评论数量
	AvgScore     float64      `json:"avg_score"`     // 评论平均分
	PrimaryTag   string       `json:"primary_tag"`   // 主标签名字
	SecondaryTag string       `json:"secondary_tag"` // 次标签名字
}
