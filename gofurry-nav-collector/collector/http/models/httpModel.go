package models

import (
	"crypto/tls"
	"time"

	"github.com/gofurry/gofurry-nav-collector/common/models"
)

// HTTP 采集结果
type HTTPModel struct {
	// HTTP 基本信息
	Domain          string              `json:"domain"`        // 域名
	Url             string              `json:"url"`           // url
	StatusCode      int64               `json:"statusCode"`    // 状态码
	ResponseTime    int64               `json:"responseTime"`  // 响应时间
	ContentLength   int64               `json:"contentLength"` // 页面大小
	Title           string              `json:"title"`         // 标题
	Server          string              `json:"server"`        // 服务器类型
	Redirects       []string            `json:"redirects"`     // 重定向链
	Headers         map[string][]string `json:"headers"`       // 响应头
	Meta            map[string]string   `json:"meta"`          // meta 标签
	FinalURL        string              `json:"-"`             // v2 observation 最终 URL
	ContentType     string              `json:"-"`             // v2 observation Content-Type
	SecurityHeaders map[string]bool     `json:"-"`             // v2 observation 常见安全响应头是否存在

	// TLS
	TLSVersion    string    `json:"tlsVersion"`    // TLS 版本
	CipherSuite   string    `json:"cipherSuite"`   // 加密套件
	CertExpiry    time.Time `json:"certExpiry"`    // 证书过期时间
	CertDaysLeft  int64     `json:"certDaysLeft"`  // 证书剩余天数
	CertIssuer    string    `json:"certIssuer"`    // 签发机构
	CertIssuerOrg []string  `json:"certIssuerOrg"` // 签发组织
	CertDNSNames  []string  `json:"certDNSNames"`  // 绑定域名
	CertPubKeyAlg string    `json:"certPubKeyAlg"` // 公钥算法
	CertSigAlg    string    `json:"certSigAlg"`    // 签名算法
	CertEmail     []string  `json:"certEmail"`     // 绑定邮箱
	CertIsCA      bool      `json:"certIsCA"`      // 是否CA

	// 其他
	StartTime    models.LocalTime `json:"startTime"` // 请求开始时间
	ErrorCode    string           `json:"-"`         // v2 observation 错误码
	ErrorMessage string           `json:"-"`         // v2 observation 错误信息
}

type HTTPSaveModel struct {
	// HTTP 基本信息
	Domain        string              `json:"domain"`        // 域名
	Url           string              `json:"url"`           // url
	StatusCode    int64               `json:"statusCode"`    // 状态码
	ResponseTime  string              `json:"responseTime"`  // 响应时间
	ContentLength int64               `json:"contentLength"` // 页面大小
	Title         string              `json:"title"`         // 标题
	Server        string              `json:"server"`        // 服务器类型
	Redirects     []string            `json:"redirects"`     // 重定向链
	Headers       map[string][]string `json:"headers"`       // 响应头
	Meta          map[string]string   `json:"meta"`          // meta 标签

	// TLS
	TLSVersion    string   `json:"tlsVersion"`    // TLS 版本
	CipherSuite   string   `json:"cipherSuite"`   // 加密套件
	CertExpiry    string   `json:"certExpiry"`    // 证书过期时间
	CertDaysLeft  string   `json:"certDaysLeft"`  // 证书剩余天数
	CertIssuer    string   `json:"certIssuer"`    // 签发机构
	CertIssuerOrg []string `json:"certIssuerOrg"` // 签发组织
	CertDNSNames  []string `json:"certDNSNames"`  // 绑定域名
	CertPubKeyAlg string   `json:"certPubKeyAlg"` // 公钥算法
	CertSigAlg    string   `json:"certSigAlg"`    // 签名算法
	CertEmail     []string `json:"certEmail"`     // 绑定邮箱
	CertIsCA      bool     `json:"certIsCA"`      // 是否CA
}

// TLS 版本映射
var TlsVersionMap = map[uint16]string{
	tls.VersionTLS10: "TLS1.0",
	tls.VersionTLS11: "TLS1.1",
	tls.VersionTLS12: "TLS1.2",
	tls.VersionTLS13: "TLS1.3",
}

// CipherSuite 映射
var CipherSuiteMap = map[uint16]string{
	tls.TLS_AES_128_GCM_SHA256:                "AES_128_GCM_SHA256",
	tls.TLS_AES_256_GCM_SHA384:                "AES_256_GCM_SHA384",
	tls.TLS_CHACHA20_POLY1305_SHA256:          "CHACHA20_POLY1305_SHA256",
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256: "ECDHE_RSA_AES_128_GCM_SHA256",
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384: "ECDHE_RSA_AES_256_GCM_SHA384",
}

// 请求头
var HeadersMap = map[string]string{
	"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
	"Connection":      "Keep-Alive",
	"Accept-Language": "zh-CN,zh;q=0.9",
}

// 需要解析的响应头
var CommonHeaders = []string{
	"Server", "Content-Type", "Content-Language",
}
