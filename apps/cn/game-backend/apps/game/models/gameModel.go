package models

import cm "github.com/gofurry/gofurry-game-backend/common/models"

const TableNameGfgGame = "gfg_game"

// GfgGame is the site-maintained game profile table.
type GfgGame struct {
	ID           int64        `db:"id" json:"id"`
	Name         string       `db:"name" json:"name"`
	NameEn       string       `db:"name_en" json:"nameEn"`
	Info         string       `db:"info" json:"info"`
	InfoEn       string       `db:"info_en" json:"infoEn"`
	CreateTime   cm.LocalTime `db:"create_time" json:"createTime"`
	UpdateTime   cm.LocalTime `db:"update_time" json:"updateTime"`
	Resources    *string      `db:"resources" json:"resources"`
	Groups       *string      `db:"groups" json:"groups"`
	ReleaseDate  string       `db:"release_date" json:"releaseDate"`
	Developers   string       `db:"developers" json:"developers"`
	Publishers   string       `db:"publishers" json:"publishers"`
	Appid        int64        `db:"appid" json:"appid"`
	Header       string       `db:"header" json:"header"`
	Links        *string      `db:"links" json:"links"`
	Weight       int64        `db:"weight" json:"weight"`
	PrimaryTag   int64        `db:"primary_tag" json:"primaryTag"`
	SecondaryTag int64        `db:"secondary_tag" json:"secondaryTag"`
}

func (*GfgGame) TableName() string {
	return TableNameGfgGame
}
