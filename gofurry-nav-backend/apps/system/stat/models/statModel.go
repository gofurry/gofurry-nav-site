package models

import cm "github.com/gofurry/gofurry-nav-backend/common/models"

type ViewsCountVo struct {
	Total      int64    `json:"total"`
	YearCount  int64    `json:"year_count"`
	MonthCount int64    `json:"month_count"`
	Date       []string `json:"date"`
	Count      []int64  `json:"count"`
}

type GroupCountVo struct {
	Name  string `gorm:"column:name" json:"name"`
	Count int64  `gorm:"column:count" json:"count"`
}

type RegionCountVo struct {
	RegionMap map[string]int64 `gorm:"column:region_map" json:"region_map"`
}

type SiteListVo struct {
	Name       string       `gorm:"column:name" json:"name"`
	Country    string       `gorm:"column:country" json:"country"`
	CreateTime cm.LocalTime `gorm:"column:create_time" json:"create_time"`
}

type SiteTypeModel struct {
	NSFW    string `gorm:"column:nsfw" json:"nsfw"`
	Welfare string `gorm:"column:welfare" json:"welfare"`
}

type SiteCommonInfoVo struct {
	CommonCountModel
	SiteReachRate  float64 `json:"site_reach_rate"`
	NonProfitRatio float64 `json:"non_profit_business_ratio"`
	SfwNsfwRatio   float64 `json:"sfw_nsfw_ratio"`
}

type CommonCountModel struct {
	SiteCount   int64 `json:"site_count"`
	DomainCount int64 `json:"domain_count"`
	DNSCount    int64 `json:"dns_count"`
	HTTPCount   int64 `json:"http_count"`
	PingCount   int64 `json:"ping_count"`
}

type PingStatusModel struct {
	Name       string       `gorm:"column:name" json:"name"`
	Status     string       `gorm:"column:status" json:"status"`
	CreateTime cm.LocalTime `gorm:"column:create_time" json:"create_time"`
}

type PingLogVo struct {
	Name       string       `json:"name"`
	Status     string       `json:"status"`
	CreateTime cm.LocalTime `json:"createTime"`
	Loss       string       `json:"loss"`
	Delay      string       `json:"delay"`
}

type PromMetricsVo struct {
	Node     map[string]string            `json:"node"`
	Nav      map[string]string            `json:"nav"`
	Game     map[string]string            `json:"game"`
	NavPath  map[string]map[string]string `json:"nav_path"`
	GamePath map[string]map[string]string `json:"game_path"`
}

type PromMetricsHistoryVo struct {
	CPU struct {
		TwentyMinutes []MetricsModel `json:"twenty_minutes"`
		OneHour       []MetricsModel `json:"one_hour"`
		TwentyHours   []MetricsModel `json:"twenty_hours"`
	} `json:"cpu"`
	Connect struct {
		TwentyMinutes []MetricsModel `json:"twenty_minutes"`
		OneHour       []MetricsModel `json:"one_hour"`
		TwentyHours   []MetricsModel `json:"twenty_hours"`
	} `json:"connect"`
	Memory struct {
		TwentyMinutes []MetricsModel `json:"twenty_minutes"`
		OneHour       []MetricsModel `json:"one_hour"`
		TwentyHours   []MetricsModel `json:"twenty_hours"`
	} `json:"memory"`
}

type MetricsModel struct {
	Time  int64   `json:"time"`
	Usage float64 `json:"usage"`
}
