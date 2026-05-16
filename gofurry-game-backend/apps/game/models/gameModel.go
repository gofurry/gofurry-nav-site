package models

import (
	"time"

	rm "github.com/gofurry/gofurry-game-backend/apps/review/models"
	cm "github.com/gofurry/gofurry-game-backend/common/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const TableNameGfgGame = "gfg_game"

// GfgGame mapped from table <gfg_game>
type GfgGame struct {
	ID           int64        `gorm:"column:id;type:bigint;primaryKey;comment:游戏表ID" json:"id"`                                         // 游戏表ID
	Name         string       `gorm:"column:name;type:character varying(255);not null;comment:游戏名称" json:"name"`                        // 游戏名称
	NameEn       string       `gorm:"column:name_en;type:character varying(255);not null;comment:游戏英文名称" json:"nameEn"`                 // 游戏英文名称
	Info         string       `gorm:"column:info;type:character varying(300);not null;comment:游戏简介" json:"info"`                        // 游戏简介
	InfoEn       string       `gorm:"column:info_en;type:character varying(300);not null;comment:游戏英文简介" json:"infoEn"`                 // 游戏英文简介
	CreateTime   cm.LocalTime `gorm:"column:create_time;type:int;type:unsigned;not null;autoCreateTime;comment:创建时间" json:"createTime"` // 创建时间
	UpdateTime   cm.LocalTime `gorm:"column:update_time;type:int;type:unsigned;not null;autoUpdateTime;comment:更新时间" json:"updateTime"` // 更新时间
	Resources    *string      `gorm:"column:resources;type:json;comment:游戏相关资源" json:"resources"`                                       // 游戏相关资源
	Groups       *string      `gorm:"column:groups;type:json;comment:游戏相关社群" json:"groups"`                                             // 游戏相关社群
	ReleaseDate  string       `gorm:"column:release_date;type:character varying(255);not null;comment:发行日期" json:"releaseDate"`         // 发行日期
	Developers   string       `gorm:"column:developers;type:json;not null;comment:开发商" json:"developers"`                               // 开发商
	Publishers   string       `gorm:"column:publishers;type:json;not null;comment:发行商" json:"publishers"`                               // 发行商
	Appid        int64        `gorm:"column:appid;type:bigint;not null;comment:SteamAPI appid" json:"appid"`                            // SteamAPI appid
	Header       string       `gorm:"column:header;type:character varying(255);not null;comment:游戏封面图" json:"header"`                   // 游戏封面图
	Links        *string      `gorm:"column:links;type:json;comment:三方网站链接" json:"links"`                                               // 三方网站链接
	Weight       int64        `gorm:"column:weight;type:bigint;not null;comment:权重" json:"weight"`                                      // 权重
	PrimaryTag   int64        `gorm:"column:primary_tag;type:bigint;not null;comment:主标签" json:"primaryTag"`                            // 主标签
	SecondaryTag int64        `gorm:"column:secondary_tag;type:bigint;not null;comment:次标签" json:"secondaryTag"`                        // 次标签
}

// TableName GfgGame's table name
func (*GfgGame) TableName() string {
	return TableNameGfgGame
}

type GameRespVo struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Info        string        `json:"info"`
	CreateTime  cm.LocalTime  `json:"create_time"`
	UpdateTime  cm.LocalTime  `json:"update_time"`
	Resources   *[]cm.KvModel `json:"resources"`
	Groups      *[]cm.KvModel `json:"groups"`
	ReleaseDate string        `json:"release_date"`
	Developers  []string      `json:"developers"`
	Publishers  []string      `json:"publishers"`
	Appid       string        `json:"appid"`
	Header      string        `json:"header"`
	Links       *[]cm.KvModel `json:"links"`
}

const TableNameGfgGameRecord = "gfg_game_record"

// GfgGameRecord mapped from table <gfg_game_record>
type GfgGameRecord struct {
	ID          int64  `gorm:"column:id;type:bigint;primaryKey;comment:游戏记录表id" json:"id"`                              // 游戏记录表id
	GameID      int64  `gorm:"column:game_id;type:bigint;not null;comment:游戏表id" json:"gameId,string"`                  // 游戏表id
	Language    string `gorm:"column:language;type:text;not null;comment:支持语言" json:"language"`                         // 支持语言
	ReleaseDate string `gorm:"column:release_date;type:character varying(30);not null;comment:发行时间" json:"releaseDate"` // 发行时间
	Platform    string `gorm:"column:platform;type:character varying(50);not null;comment:支持平台" json:"platform"`        // 支持平台
	Developer   string `gorm:"column:developer;type:character varying(100);not null;comment:开发商" json:"developer"`      // 开发商
	Publisher   string `gorm:"column:publisher;type:character varying(100);not null;comment:发行商" json:"publisher"`      // 发行商
	Info        string `gorm:"column:info;type:text;not null;comment:游戏概述" json:"info"`                                 // 游戏概述
	Cover       string `gorm:"column:cover;type:character varying(255);comment:封面图" json:"cover"`                       // 封面图
	Lang        string `gorm:"column:lang;type:character varying(20);not null;comment:记录的语言" json:"lang"`               // 记录的语言
	PriceList   string `gorm:"column:price_list;type:json;not null;comment:游戏价格列表" json:"priceList"`                    // 游戏价格列表
	Initial     int64  `gorm:"column:initial;type:bigint;not null;comment:游戏价格" json:"initial"`                         // 游戏价格
	Final       int64  `gorm:"column:final;type:bigint;not null;comment:当前价格" json:"final"`                             // 当前价格
	Discount    int64  `gorm:"column:discount;type:bigint;not null;comment:折扣百分比" json:"discount"`                      // 折扣百分比
}

// TableName GfgGameRecord's table name
func (*GfgGameRecord) TableName() string {
	return TableNameGfgGameRecord
}

type GameMainInfoVo struct {
	Latest []rm.AvgScoreResult `json:"latest"`
	Recent []rm.AvgScoreResult `json:"recent"`
	Hot    []rm.AvgScoreResult `json:"hot"`
	Free   []rm.AvgScoreResult `json:"free"`
}

const TableNameGfgGamePlayerCount = "gfg_game_player_count"

// GfgGamePlayerCount mapped from table <gfg_game_player_count>
type GfgGamePlayerCount struct {
	ID         int64        `gorm:"column:id;type:bigint;primaryKey;comment:在线人数表ID" json:"id"`                                       // 在线人数表ID
	GameID     int64        `gorm:"column:game_id;type:bigint;not null;comment:游戏表ID" json:"gameId,string"`                           // 游戏表ID
	Count_     int64        `gorm:"column:count;type:bigint;not null;comment:在线人数" json:"count"`                                      // 在线人数
	CreateTime cm.LocalTime `gorm:"column:create_time;type:int;type:unsigned;not null;autoCreateTime;comment:创建时间" json:"createTime"` // 创建时间
}

// TableName GfgGamePlayerCount's table name
func (*GfgGamePlayerCount) TableName() string {
	return TableNameGfgGamePlayerCount
}

type PlayerTopCountVo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CountPeak   int64  `json:"count_peak"`
	CountRecent int64  `json:"count_recent"`
	CollectTime int64  `json:"collect_time"`
	Header      string `json:"header"`
}

type CachePriceVo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	GlobalPrice int64  `json:"global_price"`
	ChinaPrice  int64  `json:"china_price"`
	Discount    int64  `json:"discount"`
	Header      string `json:"header"`
}

type GameMainPanelVo struct {
	TopCount struct {
		One   []PlayerTopCountVo `json:"one"`
		Two   []PlayerTopCountVo `json:"two"`
		Three []PlayerTopCountVo `json:"three"`
		Four  []PlayerTopCountVo `json:"four"`
	} `json:"top_count"`
	TopDiscountVo []CachePriceVo `json:"top_discount_vo"`
	TopPriceVo    []CachePriceVo `json:"top_price_vo"`
	BottomPrice   struct {
		One   []CachePriceVo `json:"one"`
		Two   []CachePriceVo `json:"two"`
		Three []CachePriceVo `json:"three"`
		Four  []CachePriceVo `json:"four"`
	} `json:"bottom_price"`
}

type UpdateNewsModels struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	PostTime cm.LocalTime `json:"post_time"`
	Headline string       `json:"headline"`
	Header   string       `json:"header"`
	Author   string       `json:"author"`
	Content  string       `json:"content"`
	Url      string       `json:"url"`
}

type UpdateNewsVo struct {
	NewsZh []UpdateNewsModels `json:"news_zh"`
	NewsEn []UpdateNewsModels `json:"news_en"`
}

const TableNameGfgGameNews = "gfg_game_news"

// GfgGameNews mapped from table <gfg_game_news>
type GfgGameNews struct {
	ID         int64        `gorm:"column:id;type:bigint;primaryKey;comment:游戏更新公告记录表id" json:"id"`                                   // 游戏更新公告记录表id
	GameID     int64        `gorm:"column:game_id;type:bigint;not null;comment:游戏表id" json:"gameId,string"`                           // 游戏表id
	Headline   string       `gorm:"column:headline;type:character varying(255);not null;comment:更新公告标题" json:"headline"`              // 更新公告标题
	Content    string       `gorm:"column:content;type:text;not null;comment:更新公告内容" json:"content"`                                  // 更新公告内容
	Index      int64        `gorm:"column:index;type:bigint;not null;comment:更新公告编号" json:"index"`                                    // 更新公告编号
	PostTime   cm.LocalTime `gorm:"column:post_time;type:timestamp(0) without time zone;not null;comment:更新公告上传日期" json:"postTime"`   // 更新公告上传日期
	CreateTime cm.LocalTime `gorm:"column:create_time;type:int;type:unsigned;not null;autoCreateTime;comment:采集时间" json:"createTime"` // 采集时间
	Author     string       `gorm:"column:author;type:character varying(50);not null;comment:公告作者" json:"author"`                     // 公告作者
	URL        string       `gorm:"column:url;type:character varying(255);not null;comment:更新公告原始地址" json:"url"`                      // 更新公告原始地址
	Total      int64        `gorm:"column:total;type:bigint;not null;comment:公告总数" json:"total"`                                      // 公告总数
	Lang       string       `gorm:"column:lang;type:character varying(30);not null;comment:记录的语言" json:"lang"`                        // 记录的语言
}

// TableName GfgGameNews's table name
func (*GfgGameNews) TableName() string {
	return TableNameGfgGameNews
}

type TagModelVo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Prefix    string `json:"prefix"`
	GameCount int    `json:"game_count"`
}

type GameBaseInfoVo struct {
	Name       string       `json:"name"`
	Info       string       `json:"info"`
	CreateTime cm.LocalTime `json:"create_time"`
	UpdateTime cm.LocalTime `json:"update_time"`
	Resources  []struct {
		cm.KvModel
	} `json:"resources"`
	Groups []struct {
		cm.KvModel
	} `json:"groups"`
	ReleaseDate string   `json:"release_date"`
	Developers  []string `json:"developers"`
	Publishers  []string `json:"publishers"`
	Appid       int64    `json:"appid"`
	Cover       string   `json:"cover"`
	Links       []struct {
		cm.KvModel
	} `json:"links"`
	Platform            string               `json:"platform"`
	PriceList           []PriceModel         `json:"price_list"`
	News                []NewsVo             `json:"news"`
	Tags                []TagVo              `json:"tags"`
	Support             SteamAppSupport      `json:"support"`
	Screenshots         []SteamAppScreenshot `json:"screenshots"`
	Movies              []SteamAppMovie      `json:"movies"`
	SupportedLanguages  string               `json:"supported_languages"`
	RequiredAge         string               `json:"required_age"`
	Website             string               `json:"website"`
	DetailedDescription string               `json:"detailed_description"`
	AboutTheGame        string               `json:"about_the_game"`
	PcRequirements      PcRequirementModel   `json:"pc_requirements"`
	OnlineCount         int64                `json:"online_count"`
	CountCollectTime    string               `json:"count_collect_time"`
	ViewCount           int64                `json:"view_count"`
}

type NewsVo struct {
	Headline string       `json:"headline"`
	Content  string       `json:"content"`
	PostTime cm.LocalTime `json:"post_time"`
	Author   string       `json:"author"`
	URL      string       `json:"url"`
}

type TagVo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Desc string `json:"desc"`
}

type GameRemarkVo struct {
	Total    int           `json:"total"`
	AvgScore float64       `json:"avg_score"`
	Remarks  []CommentItem `json:"remarks"`
}

type CommentItem struct {
	Region     string       `json:"region"`
	Content    string       `json:"content"`
	Score      float64      `json:"score"`
	CreateTime cm.LocalTime `json:"create_time"`
	IP         string       `json:"ip"`
	Name       *string      `json:"name"`
}

// GameIntro 游戏简介HTML存储模型
type GameIntro struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`        // MongoDB自动生成的ID
	GameID     int64              `bson:"game_id" json:"game_id"`         // 游戏表ID（核心业务标识）
	Content    string             `bson:"content" json:"content"`         // HTML内容
	Lang       string             `bson:"lang" json:"lang"`               // 语言
	CreateTime time.Time          `bson:"create_time" json:"create_time"` // 创建时间
	UpdateTime time.Time          `bson:"update_time" json:"update_time"` // 更新时间
}

// TableName 返回集合名
func (GameIntro) TableName() string {
	return "game_intro"
}

type GameIntroVo struct {
	Content    string       `json:"content"`
	UpdateTime cm.LocalTime `json:"update_time"`
}

const TableNameGfgGameCreator = "gfg_game_creator"

// GfgGameCreator mapped from table <gfg_game_creator>
type GfgGameCreator struct {
	ID         int64        `gorm:"column:id;type:bigint;primaryKey;comment:相关作者表id" json:"id"`                                          // 相关作者表id
	Name       string       `gorm:"column:name;type:character varying(100);not null;comment:名称" json:"name"`                             // 名称
	Info       string       `gorm:"column:info;type:character varying(700);not null;comment:相关描述" json:"info"`                           // 相关描述
	MainURL    string       `gorm:"column:main_url;type:character varying(255);not null;comment:主链接" json:"mainUrl"`                     // 主链接
	Links      *string      `gorm:"column:links;type:jsonb;comment:其他链接" json:"links"`                                                   // 其他链接
	Cover      string       `gorm:"column:cover;type:character varying(255);not null;comment:封面图" json:"cover"`                          // 封面图
	Contact    *string      `gorm:"column:contact;type:jsonb;comment:联系方式" json:"contact"`                                               // 联系方式
	CreateTime cm.LocalTime `gorm:"column:create_time;type:int;type:unsigned;not null;autoCreateTime;comment:创建时间" json:"createTime"`    // 创建时间
	UpdateTime cm.LocalTime `gorm:"column:update_time;type:int;type:unsigned;not null;autoUpdateTime;comment:修改时间" json:"updateTime"`    // 修改时间
	Type       int64        `gorm:"column:type;type:bigint;not null;comment:类型描述 1=Steam鉴赏家 2=博主 3=开发者 4=发行者 5=汉化者 6=内容创作者" json:"type"` // 类型描述 1=Steam鉴赏家 2=博主 3=开发者 4=发行者 5=汉化者 6=内容创作者
	NameEn     string       `gorm:"column:name_en;type:character varying(100);comment:英文名" json:"nameEn"`                                // 英文名
	InfoEn     string       `gorm:"column:info_en;type:character varying(700);comment:英文描述" json:"infoEn"`                               // 英文描述
	Deleted    bool         `gorm:"column:deleted;type:boolean;comment:软删除" json:"deleted"`
}

// TableName GfgGameCreator's table name
func (*GfgGameCreator) TableName() string {
	return TableNameGfgGameCreator
}

type CreatorVo struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Info       string       `json:"info"`
	URL        string       `json:"url"`
	Avatar     string       `json:"avatar"`
	Links      []cm.KvModel `json:"links"`
	Contact    []cm.KvModel `json:"contact"`
	Type       int64        `json:"type"`
	CreateTime cm.LocalTime `json:"create_time"`
	UpdateTime cm.LocalTime `json:"update_time"`
}

type TempCreator struct {
	ID         int64        `gorm:"column:id"`
	Name       string       `gorm:"column:name"`
	Info       string       `gorm:"column:info"`
	URL        string       `gorm:"column:url"`
	Avatar     string       `gorm:"column:avatar"`
	Links      *string      `gorm:"column:links"`
	Contact    *string      `gorm:"column:contact"`
	Type       int64        `gorm:"column:type"`
	CreateTime cm.LocalTime `gorm:"column:create_time"`
	UpdateTime cm.LocalTime `gorm:"column:update_time"`
}

type UpdateCreatorVo struct {
	CreatorZh []CreatorVo `json:"creator_zh"`
	CreatorEn []CreatorVo `json:"creator_en"`
}

// ============================= redis 记录
type SteamAppPrice struct {
	Initial          int64  `json:"initial"`
	Final            int64  `json:"final"`
	Currency         string `json:"currency"`
	DiscountPercent  int64  `json:"discount_percent"`
	InitialFormatted string `json:"initial_formatted"`
	FinalFormatted   string `json:"final_formatted"`
}

type SteamAppRelease struct {
	ComingSoon bool   `json:"coming_soon"`
	Date       string `json:"date"`
}

type SteamAppPlatform struct {
	Windows bool `json:"windows"`
	Mac     bool `json:"mac"`
	Linux   bool `json:"linux"`
}

type SteamAppSupport struct {
	URL   string `json:"url"`
	Email string `json:"email"`
}

type SteamAppScreenshot struct {
	ID            int64  `json:"id"`
	PathThumbnail string `json:"path_thumbnail"`
	PathFull      string `json:"path_full"`
}

type SteamAppMovie struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Thumbnail string `json:"thumbnail"`
	DashAv1   string `json:"dash_av1"`
	DashH264  string `json:"dash_h264"`
	HlsH264   string `json:"hls_h264"`
}

type PriceModel struct {
	Price   string `json:"price"`
	Country string `json:"country"`
}

type PcRequirementModel struct {
	Minimum     string `json:"minimum"`
	Recommended string `json:"recommended"`
}

type GameSaveModel struct {
	Price               SteamAppPrice        `json:"price"`
	Support             SteamAppSupport      `json:"support"`
	Screenshots         []SteamAppScreenshot `json:"screenshots"`
	Movies              []SteamAppMovie      `json:"movies"`
	PriceList           string               `json:"price_list"`
	SupportedLanguages  string               `json:"supported_languages"`
	Developers          string               `json:"developers"`
	Publishers          string               `json:"publishers"`
	HeaderImage         string               `json:"header_image"`
	ShortDescription    string               `json:"short_description"`
	Date                string               `json:"date"`
	Platforms           string               `json:"platforms"`
	RequiredAge         string               `json:"required_age"`
	Website             string               `json:"website"`
	ContentDescriptors  string               `json:"content_descriptors"`
	DetailedDescription string               `json:"detailed_description"`
	AboutTheGame        string               `json:"about_the_game"`
	PcRequirements      PcRequirementModel   `json:"pc_requirements"`

	CollectDate cm.LocalTime `json:"collect_date"`
}

type GameOnlineModel struct {
	ID         int64  `json:"id"`
	GameID     string `json:"gameId"`
	Count      int64  `json:"count"`
	CreateTime string `json:"createTime"`
}
