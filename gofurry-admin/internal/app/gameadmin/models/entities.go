package models

import (
	pkgmodels "github.com/gofurry/awesome-fiber-template/v3/medium/pkg/models"
)

type Game struct {
	ID           int64               `gorm:"column:id;primaryKey" json:"id"`
	Name         string              `gorm:"column:name;not null" json:"name"`
	NameEn       string              `gorm:"column:name_en;not null" json:"name_en"`
	Info         string              `gorm:"column:info;not null" json:"info"`
	InfoEn       string              `gorm:"column:info_en;not null" json:"info_en"`
	CreateTime   pkgmodels.LocalTime `gorm:"column:create_time;autoCreateTime" json:"create_time"`
	UpdateTime   pkgmodels.LocalTime `gorm:"column:update_time;autoUpdateTime" json:"update_time"`
	Resources    *string             `gorm:"column:resources" json:"-"`
	Groups       *string             `gorm:"column:groups" json:"-"`
	ReleaseDate  string              `gorm:"column:release_date;not null" json:"release_date"`
	Developers   string              `gorm:"column:developers;not null" json:"-"`
	Publishers   string              `gorm:"column:publishers;not null" json:"-"`
	Appid        int64               `gorm:"column:appid;not null" json:"appid"`
	Header       string              `gorm:"column:header;not null" json:"header"`
	Links        *string             `gorm:"column:links" json:"-"`
	Weight       int64               `gorm:"column:weight;not null" json:"weight"`
	PrimaryTag   int64               `gorm:"column:primary_tag;not null" json:"primary_tag"`
	SecondaryTag int64               `gorm:"column:secondary_tag;not null" json:"secondary_tag"`
}

func (*Game) TableName() string { return "gfg_game" }

type GameDTO struct {
	ID           int64               `json:"id"`
	Name         string              `json:"name"`
	NameEn       string              `json:"name_en"`
	Info         string              `json:"info"`
	InfoEn       string              `json:"info_en"`
	CreateTime   pkgmodels.LocalTime `json:"create_time"`
	UpdateTime   pkgmodels.LocalTime `json:"update_time"`
	Resources    []pkgmodels.KvModel `json:"resources"`
	Groups       []pkgmodels.KvModel `json:"groups"`
	ReleaseDate  string              `json:"release_date"`
	Developers   []string            `json:"developers"`
	Publishers   []string            `json:"publishers"`
	Appid        int64               `json:"appid"`
	Header       string              `json:"header"`
	Links        []pkgmodels.KvModel `json:"links"`
	Weight       int64               `json:"weight"`
	PrimaryTag   int64               `json:"primary_tag"`
	SecondaryTag int64               `json:"secondary_tag"`
}

type GamePayload struct {
	Name         string              `json:"name"`
	NameEn       string              `json:"name_en"`
	Info         string              `json:"info"`
	InfoEn       string              `json:"info_en"`
	Resources    []pkgmodels.KvModel `json:"resources"`
	Groups       []pkgmodels.KvModel `json:"groups"`
	ReleaseDate  string              `json:"release_date"`
	Developers   []string            `json:"developers"`
	Publishers   []string            `json:"publishers"`
	Appid        int64               `json:"appid"`
	Header       string              `json:"header"`
	Links        []pkgmodels.KvModel `json:"links"`
	Weight       int64               `json:"weight"`
	PrimaryTag   int64               `json:"primary_tag"`
	SecondaryTag int64               `json:"secondary_tag"`
}

type GameComment struct {
	ID         int64               `gorm:"column:id;primaryKey" json:"id"`
	Region     string              `gorm:"column:region;not null" json:"region"`
	Content    string              `gorm:"column:content;not null" json:"content"`
	Score      float64             `gorm:"column:score;not null" json:"score"`
	CreateTime pkgmodels.LocalTime `gorm:"column:create_time;autoCreateTime" json:"create_time"`
	GameID     int64               `gorm:"column:game_id;not null" json:"game_id"`
	IP         string              `gorm:"column:ip;not null" json:"ip"`
	Name       string              `gorm:"column:name" json:"name"`
}

func (*GameComment) TableName() string { return "gfg_game_comment" }

type GameCommentPayload struct {
	Region  string  `json:"region"`
	Content string  `json:"content"`
	Score   float64 `json:"score"`
	GameID  int64   `json:"game_id"`
	IP      string  `json:"ip"`
	Name    string  `json:"name"`
}

type Prize struct {
	ID         int64               `gorm:"column:id;primaryKey" json:"id"`
	Title      string              `gorm:"column:title;not null" json:"title"`
	Desc       string              `gorm:"column:desc;not null" json:"desc"`
	Prize      string              `gorm:"column:prize;not null" json:"-"`
	Key        string              `gorm:"column:key;not null" json:"key"`
	StartTime  pkgmodels.LocalTime `gorm:"column:start_time;not null" json:"start_time"`
	EndTime    pkgmodels.LocalTime `gorm:"column:end_time;not null" json:"end_time"`
	CreateTime pkgmodels.LocalTime `gorm:"column:create_time;autoCreateTime" json:"create_time"`
	Status     bool                `gorm:"column:status;not null" json:"status"`
}

func (*Prize) TableName() string { return "gfg_prize" }

type PrizeBody struct {
	Keys     []string `json:"keys"`
	Title    string   `json:"title"`
	Platform string   `json:"platform"`
}

type PrizeDTO struct {
	ID         int64               `json:"id"`
	Title      string              `json:"title"`
	Desc       string              `json:"desc"`
	Prize      PrizeBody           `json:"prize"`
	Key        string              `json:"key"`
	StartTime  pkgmodels.LocalTime `json:"start_time"`
	EndTime    pkgmodels.LocalTime `json:"end_time"`
	CreateTime pkgmodels.LocalTime `json:"create_time"`
	Status     bool                `json:"status"`
}

type PrizePayload struct {
	Title     string    `json:"title"`
	Desc      string    `json:"desc"`
	Prize     PrizeBody `json:"prize"`
	Key       string    `json:"key"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
	Status    bool      `json:"status"`
}

type Tag struct {
	ID         int64               `gorm:"column:id;primaryKey" json:"id"`
	Name       string              `gorm:"column:name;not null" json:"name"`
	NameEn     string              `gorm:"column:name_en;not null" json:"name_en"`
	Info       string              `gorm:"column:info;not null" json:"info"`
	InfoEn     string              `gorm:"column:info_en;not null" json:"info_en"`
	Prefix     int64               `gorm:"column:prefix;not null" json:"prefix"`
	CreateTime pkgmodels.LocalTime `gorm:"column:create_time;autoCreateTime" json:"create_time"`
	UpdateTime pkgmodels.LocalTime `gorm:"column:update_time;autoUpdateTime" json:"update_time"`
}

func (*Tag) TableName() string { return "gfg_tag" }

type TagPayload struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	NameEn string `json:"name_en"`
	Info   string `json:"info"`
	InfoEn string `json:"info_en"`
	Prefix int64  `json:"prefix"`
}

type TagMap struct {
	ID         int64               `gorm:"column:id;primaryKey" json:"id"`
	GameID     int64               `gorm:"column:game_id;not null" json:"game_id"`
	TagID      int64               `gorm:"column:tag_id;not null" json:"tag_id"`
	CreateTime pkgmodels.LocalTime `gorm:"column:create_time;autoCreateTime" json:"create_time"`
	UpdateTime pkgmodels.LocalTime `gorm:"column:update_time;autoUpdateTime" json:"update_time"`
}

func (*TagMap) TableName() string { return "gfg_tag_map" }

type TagMapPayload struct {
	GameID int64 `json:"game_id"`
	TagID  int64 `json:"tag_id"`
}

type TagMapDTO struct {
	ID         int64               `json:"id"`
	GameID     int64               `json:"game_id"`
	TagID      int64               `json:"tag_id"`
	GameName   string              `json:"game_name"`
	TagName    string              `json:"tag_name"`
	CreateTime pkgmodels.LocalTime `json:"create_time"`
	UpdateTime pkgmodels.LocalTime `json:"update_time"`
}
