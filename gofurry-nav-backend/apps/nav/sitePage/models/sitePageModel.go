package models

import (
	cm "github.com/gofurry/gofurry-nav-backend/common/models"
)

type SiteInfoVo struct {
	Name      string  `form:"name"  json:"name"`
	Info      string  `form:"info"  json:"info"`
	Icon      *string `form:"icon"  json:"icon"`
	Country   *string `form:"country"  json:"country"`
	Nsfw      string  `form:"nsfw"  json:"nsfw"`
	Welfare   string  `form:"welfare"  json:"welfare"`
	ViewCount int64   `form:"viewCount" json:"view_count"`
}

type SiteDnsVo struct {
	A     string `form:"a"  json:"a"`
	AAAA  string `form:"AAAA"  json:"AAAA"`
	CNAME string `form:"CNAME"  json:"CNAME"`
	TXT   string `form:"txt"  json:"txt"`
	MX    string `form:"MX"  json:"MX"`
	NS    string `form:"ns"  json:"ns"`
	SOA   string `form:"SOA"  json:"SOA"`
	CAA   string `form:"caa"  json:"caa"`
}

type SiteDelayVo struct {
	Twenty  SiteDelayModel `form:"twenty"  json:"twenty"`
	Sixty   SiteDelayModel `form:"sixty"  json:"sixty"`
	Hundred SiteDelayModel `form:"hundred"  json:"hundred"`
}

type SiteDelayModel struct {
	DelayModel []SiteDelay
	AvgDelay   string `form:"avgDelay"  json:"avgDelay"`
	AvgLoss    string `form:"avgLoss"  json:"avgLoss"`
}

type SiteDelay struct {
	Delay  int          `form:"delay"  json:"delay"`
	Loss   int          `form:"loss"  json:"loss"`
	Status string       `form:"status"  json:"status"`
	Time   cm.LocalTime `form:"time"  json:"time"`
}
