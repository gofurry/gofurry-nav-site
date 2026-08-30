package models

import (
	pkgmodels "github.com/gofurry/gofurry-admin/pkg/models"
)

type Game struct {
	ID           int64               `json:"id"`
	Name         string              `json:"name"`
	NameEn       string              `json:"name_en"`
	Info         string              `json:"info"`
	InfoEn       string              `json:"info_en"`
	CreateTime   pkgmodels.LocalTime `json:"create_time"`
	UpdateTime   pkgmodels.LocalTime `json:"update_time"`
	Resources    *string             `json:"-"`
	Groups       *string             `json:"-"`
	Developers   string              `json:"-"`
	Publishers   string              `json:"-"`
	Appid        int64               `json:"appid"`
	Header       string              `json:"header"`
	Links        *string             `json:"-"`
	Weight       int64               `json:"weight"`
	PrimaryTag   int64               `json:"primary_tag"`
	SecondaryTag int64               `json:"secondary_tag"`
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
	Developers   []string            `json:"developers"`
	Publishers   []string            `json:"publishers"`
	Appid        int64               `json:"appid"`
	Header       string              `json:"header"`
	Links        []pkgmodels.KvModel `json:"links"`
	Weight       int64               `json:"weight"`
	PrimaryTag   int64               `json:"primary_tag"`
	SecondaryTag int64               `json:"secondary_tag"`
}

type GameWorkspaceTag struct {
	ID      int64  `json:"id"`
	GameID  int64  `json:"game_id"`
	TagID   int64  `json:"tag_id"`
	TagName string `json:"tag_name"`
}

type GameWorkspace struct {
	Game GameDTO            `json:"game"`
	Tags []GameWorkspaceTag `json:"tags"`
}

type GamePayload struct {
	Name         string              `json:"name"`
	NameEn       string              `json:"name_en"`
	Info         string              `json:"info"`
	InfoEn       string              `json:"info_en"`
	Resources    []pkgmodels.KvModel `json:"resources"`
	Groups       []pkgmodels.KvModel `json:"groups"`
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
	ID         int64               `json:"id"`
	Region     string              `json:"region"`
	Content    string              `json:"content"`
	Score      float64             `json:"score"`
	CreateTime pkgmodels.LocalTime `json:"create_time"`
	GameID     int64               `json:"game_id"`
	IP         string              `json:"ip"`
	Name       string              `json:"name"`
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
	ID         int64               `json:"id"`
	Title      string              `json:"title"`
	Desc       string              `json:"desc"`
	Prize      string              `json:"-"`
	Key        string              `json:"key"`
	StartTime  pkgmodels.LocalTime `json:"start_time"`
	EndTime    pkgmodels.LocalTime `json:"end_time"`
	CreateTime pkgmodels.LocalTime `json:"create_time"`
	Status     bool                `json:"status"`
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
	ID         int64               `json:"id"`
	Name       string              `json:"name"`
	NameEn     string              `json:"name_en"`
	Info       string              `json:"info"`
	InfoEn     string              `json:"info_en"`
	Prefix     int64               `json:"prefix"`
	CreateTime pkgmodels.LocalTime `json:"create_time"`
	UpdateTime pkgmodels.LocalTime `json:"update_time"`
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
	ID         int64               `json:"id"`
	GameID     int64               `json:"game_id"`
	TagID      int64               `json:"tag_id"`
	CreateTime pkgmodels.LocalTime `json:"create_time"`
	UpdateTime pkgmodels.LocalTime `json:"update_time"`
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
