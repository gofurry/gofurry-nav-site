package models

import (
	"time"

	"github.com/miekg/dns"
)

type DNSRecord struct {
	Type         string        `json:"type"`          // 记录类型，如 A/AAAA/MX 等
	Value        string        `json:"value"`         // 记录值，如 IP / 域名
	TTL          uint32        `json:"ttl"`           // TTL 值
	DNSSEC       bool          `json:"dnssec"`        // 是否启用 DNSSEC
	ASN          string        `json:"asn"`           // IP 所属 ASN
	Country      string        `json:"country"`       // IP 国家
	City         string        `json:"city"`          // IP 城市
	ProviderType string        `json:"provider_type"` // 类型判定：CDN / Origin
	ISP          string        `json:"isp"`           // ISP 名称
	Duration     time.Duration `json:"duration"`      // 查询耗时
	Children     []DNSRecord   `json:"children"`      // 子记录（递归查询产生）
	ReversePTR   string        `json:"reverse_ptr"`   // 反向 PTR
	Hijacked     bool          `json:"hijacked"`      // 劫持检测标记
	RiskFlags    []string      `json:"-"`             // v2 observation 风险标记
}

// DNSStatistics 统计结果
type DNSStatistics struct {
	MinTTL    uint32        `json:"min_ttl"`    // 最小 TTL
	MaxTTL    uint32        `json:"max_ttl"`    // 最大 TTL
	AvgTTL    float64       `json:"avg_ttl"`    // 平均 TTL
	MinTime   time.Duration `json:"min_time"`   // 单条最小耗时
	MaxTime   time.Duration `json:"max_time"`   // 单条最大耗时
	AvgTime   time.Duration `json:"avg_time"`   // 单条平均耗时
	TotalTime time.Duration `json:"total_time"` // 查询总耗时
}

type RecordType struct {
	Type uint16
	Name string
}

var RecordTypes = []RecordType{
	{dns.TypeA, "A"},
	{dns.TypeAAAA, "AAAA"},
	{dns.TypeMX, "MX"},
	{dns.TypeNS, "NS"},
	{dns.TypeTXT, "TXT"},
	{dns.TypeCNAME, "CNAME"},
	{dns.TypeSOA, "SOA"},
	{dns.TypeCAA, "CAA"},
}

// CDN 提供商列表，用于检测 IP 是否为 CDN
var CdnProviders = []string{
	"Cloudflare", "Akamai", "Fastly", "EdgeCast",
	"Tencent", "Alibaba", "Baidu", "ChinaCache",
	"Huawei", "JD", "Kingsoft", "Wangsu", "Sangfor",
}
