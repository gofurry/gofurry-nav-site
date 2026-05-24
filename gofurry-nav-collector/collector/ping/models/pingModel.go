package models

import "github.com/gofurry/gofurry-nav-collector/common/models"

type PingVo struct {
	Domain string `json:"domain"`
}

type PingModel struct {
	Name                  string           `json:"name"`         // 对象名称
	PingTime              models.LocalTime `json:"pingTime"`     // ping时间
	AvgLossRate           float64          `json:"avgLossRate"`  // 平均丢包率
	AvgDelayTime          int64            `json:"avgDelayTime"` // 平均延迟
	MinRTTMS              int64            `json:"-"`            // v2 observation 最小 RTT
	MaxRTTMS              int64            `json:"-"`            // v2 observation 最大 RTT
	StdDevRTTMS           int64            `json:"-"`            // v2 observation RTT 标准差
	PacketsSent           int              `json:"-"`            // v2 observation 发送包数
	PacketsRecv           int              `json:"-"`            // v2 observation 接收包数
	PacketsRecvDuplicates int              `json:"-"`            // v2 observation 重复包数
	ResolvedIP            string           `json:"-"`            // v2 observation 实际解析 IP
	ResolvedIPs           []string         `json:"-"`            // v2 observation 已解析 IP 列表
	SelectedIP            string           `json:"-"`            // v2 observation 本次选择 IP
	IPFamily              string           `json:"-"`            // v2 observation IP 协议族
	ResolutionSource      string           `json:"-"`            // v2 observation 解析来源
	ProbeDurationMS       int64            `json:"-"`            // 单目标探测墙钟耗时
	ErrorCode             string           `json:"-"`            // v2 observation 错误码
	ErrorMessage          string           `json:"-"`            // v2 observation 错误信息
}

type PingSaveModel struct {
	Status string           `json:"status"` // 状态
	Time   models.LocalTime `json:"time"`   // ping时间
	Loss   string           `json:"loss"`   // 平均丢包率
	Delay  string           `json:"delay"`  // 平均延迟
}
