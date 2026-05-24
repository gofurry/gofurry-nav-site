package models

import (
	"crypto/tls"
	"time"

	"github.com/gofurry/gofurry-nav-collector/common/models"
)

// HTTP 采集结果
type HTTPModel struct {
	// HTTP 基本信息
	Domain               string              `json:"domain"`        // 域名
	Url                  string              `json:"url"`           // url
	StatusCode           int64               `json:"statusCode"`    // 状态码
	ResponseTime         int64               `json:"responseTime"`  // 响应时间
	ContentLength        int64               `json:"contentLength"` // 页面大小
	Title                string              `json:"title"`         // 标题
	Server               string              `json:"server"`        // 服务器类型
	Redirects            []string            `json:"redirects"`     // 重定向链
	Headers              map[string][]string `json:"headers"`       // 响应头
	Meta                 map[string]string   `json:"meta"`          // meta 标签
	OpenGraph            map[string]string   `json:"openGraph,omitempty"`
	TwitterCard          map[string]string   `json:"twitterCard,omitempty"`
	CanonicalURL         string              `json:"canonicalUrl,omitempty"`
	HTMLLang             string              `json:"htmlLang,omitempty"`
	MetaRefresh          *HTTPMetaRefresh    `json:"metaRefresh,omitempty"`
	IconLinks            []HTTPLinkInfo      `json:"iconLinks,omitempty"`
	CookieSummary        *HTTPCookieSummary  `json:"cookieSummary,omitempty"`
	ServerHints          *HTTPServerHints    `json:"serverHints,omitempty"`
	CrossOriginSummary   *HTTPCrossOrigin    `json:"crossOriginSummary,omitempty"`
	ContentLanguage      string              `json:"contentLanguageEffective,omitempty"`
	PageTextSummary      string              `json:"pageTextSummary,omitempty"`
	SharePreview         *HTTPSharePreview   `json:"sharePreview,omitempty"`
	RedirectHint         *HTTPRedirectHint   `json:"redirectHint,omitempty"`
	FinalURL             string              `json:"-"` // v2 observation 最终 URL
	ContentType          string              `json:"-"` // v2 observation Content-Type
	SecurityHeaders      map[string]bool     `json:"-"` // v2 observation 常见安全响应头是否存在
	SecurityHeaderValues map[string]string   `json:"-"` // v2 observation 常见安全响应头原始值
	DNSLookupMS          int64               `json:"-"` // v2 observation DNS 查询耗时
	TCPConnectMS         int64               `json:"-"` // v2 observation TCP 连接耗时
	TLSHandshakeMS       int64               `json:"-"` // v2 observation TLS 握手耗时
	TTFBMS               int64               `json:"-"` // v2 observation 首字节耗时
	TransferMS           int64               `json:"-"` // v2 observation 响应体读取耗时
	HTTPProtocol         string              `json:"-"` // v2 observation HTTP 协议
	RemoteAddr           string              `json:"-"` // v2 observation 实际 TCP 对端
	RemoteIP             string              `json:"-"` // v2 observation 实际 TCP 对端 IP
	BodyReadBytes        int64               `json:"-"` // v2 observation 实际读取字节数
	BodyTruncated        bool                `json:"-"` // v2 observation 响应体是否被采集上限截断
	BodyLimitBytes       int64               `json:"-"` // v2 observation 响应体读取上限
	ContentEncoding      string              `json:"-"` // v2 observation Content-Encoding
	Compressed           bool                `json:"-"` // v2 observation 是否压缩响应
	CacheControl         string              `json:"-"` // v2 observation Cache-Control
	ETag                 string              `json:"-"` // v2 observation ETag
	LastModified         string              `json:"-"` // v2 observation Last-Modified
	ContentLengthHeader  int64               `json:"-"` // v2 observation Content-Length 响应头值
	TransferEncoding     []string            `json:"-"` // v2 observation Transfer-Encoding
	IsChunked            bool                `json:"-"` // v2 observation 是否 chunked 传输
	HTMLCharset          string              `json:"-"` // v2 observation HTML charset
	Doctype              string              `json:"-"` // v2 observation doctype
	RobotsMetaPolicy     string              `json:"-"` // v2 observation robots meta 策略
	CompressionRatio     float64             `json:"-"` // v2 observation 压缩比例估算

	// TLS
	TLSVersion            string    `json:"tlsVersion"`    // TLS 版本
	CipherSuite           string    `json:"cipherSuite"`   // 加密套件
	CertExpiry            time.Time `json:"certExpiry"`    // 证书过期时间
	CertDaysLeft          int64     `json:"certDaysLeft"`  // 证书剩余天数
	CertIssuer            string    `json:"certIssuer"`    // 签发机构
	CertIssuerOrg         []string  `json:"certIssuerOrg"` // 签发组织
	CertDNSNames          []string  `json:"certDNSNames"`  // 绑定域名
	CertPubKeyAlg         string    `json:"certPubKeyAlg"` // 公钥算法
	CertSigAlg            string    `json:"certSigAlg"`    // 签名算法
	CertEmail             []string  `json:"certEmail"`     // 绑定邮箱
	CertIsCA              bool      `json:"certIsCA"`      // 是否CA
	CertCollected         bool      `json:"-"`             // v2 observation 是否采集到证书
	CertVerified          bool      `json:"-"`             // v2 observation 证书是否校验通过
	VerifyError           string    `json:"-"`             // v2 observation 证书校验失败原因
	TLSHandshake          string    `json:"-"`             // v2 observation TLS 握手状态
	CertNotBefore         time.Time `json:"-"`             // v2 observation 证书生效时间
	CertNotAfter          time.Time `json:"-"`             // v2 observation 证书过期时间
	CertChainLen          int       `json:"-"`             // v2 observation 证书链长度
	CertSubjectCN         string    `json:"-"`             // v2 observation 证书 Subject CN
	CertSANCount          int       `json:"-"`             // v2 observation 证书 SAN 数量
	OCSPStapled           bool      `json:"-"`             // v2 observation 是否带 OCSP Staple
	SCTCount              int       `json:"-"`             // v2 observation SCT 数量
	VerifyErrorCategory   string    `json:"-"`             // v2 observation 证书校验错误分类
	CertSerialNumber      string    `json:"-"`             // v2 observation 证书序列号
	CertFingerprintSHA256 string    `json:"-"`             // v2 observation 证书 SHA256 指纹
	CertSPKISHA256        string    `json:"-"`             // v2 observation SPKI SHA256 指纹
	CertPublicKeyBits     int       `json:"-"`             // v2 observation 公钥位数
	CertSubjectOrg        []string  `json:"-"`             // v2 observation Subject 组织
	CertIssuerCN          string    `json:"-"`             // v2 observation Issuer CN
	CertChainIssuers      []string  `json:"-"`             // v2 observation 证书链签发者

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
	OpenGraph     map[string]string   `json:"openGraph,omitempty"`
	TwitterCard   map[string]string   `json:"twitterCard,omitempty"`
	CanonicalURL  string              `json:"canonicalUrl,omitempty"`
	HTMLLang      string              `json:"htmlLang,omitempty"`
	MetaRefresh   *HTTPMetaRefresh    `json:"metaRefresh,omitempty"`
	IconLinks     []HTTPLinkInfo      `json:"iconLinks,omitempty"`
	CookieSummary *HTTPCookieSummary  `json:"cookieSummary,omitempty"`
	ServerHints   *HTTPServerHints    `json:"serverHints,omitempty"`
	CrossOrigin   *HTTPCrossOrigin    `json:"crossOriginSummary,omitempty"`
	ContentLang   string              `json:"contentLanguageEffective,omitempty"`
	PageSummary   string              `json:"pageTextSummary,omitempty"`
	SharePreview  *HTTPSharePreview   `json:"sharePreview,omitempty"`
	RedirectHint  *HTTPRedirectHint   `json:"redirectHint,omitempty"`

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

type HTTPLinkInfo struct {
	Rel   string `json:"rel,omitempty"`
	Href  string `json:"href,omitempty"`
	Type  string `json:"type,omitempty"`
	Sizes string `json:"sizes,omitempty"`
}

type HTTPMetaRefresh struct {
	Present      bool   `json:"present"`
	DelaySeconds *int64 `json:"delay_seconds,omitempty"`
	URL          string `json:"url,omitempty"`
}

type HTTPCookieSummary struct {
	SetCookieCount      int `json:"set_cookie_count"`
	SecureCount         int `json:"secure_count"`
	HTTPOnlyCount       int `json:"http_only_count"`
	SameSiteLaxCount    int `json:"same_site_lax_count"`
	SameSiteStrictCount int `json:"same_site_strict_count"`
	SameSiteNoneCount   int `json:"same_site_none_count"`
}

type HTTPServerHints struct {
	Server     string `json:"server,omitempty"`
	XPoweredBy string `json:"x_powered_by,omitempty"`
	Generator  string `json:"generator,omitempty"`
}

type HTTPCrossOrigin struct {
	CrossOriginOpenerPolicy   string `json:"cross_origin_opener_policy,omitempty"`
	CrossOriginEmbedderPolicy string `json:"cross_origin_embedder_policy,omitempty"`
	CrossOriginResourcePolicy string `json:"cross_origin_resource_policy,omitempty"`
	AccessControlAllowOrigin  string `json:"access_control_allow_origin,omitempty"`
}

type HTTPSharePreview struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	SiteName    string `json:"site_name,omitempty"`
	Image       string `json:"image,omitempty"`
	URL         string `json:"url,omitempty"`
}

type HTTPRedirectHint struct {
	FinalURLDifferent     bool   `json:"final_url_different"`
	CanonicalURLDifferent bool   `json:"canonical_url_different"`
	MetaRefreshPresent    bool   `json:"meta_refresh_present"`
	MetaRefreshURL        string `json:"meta_refresh_url,omitempty"`
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
	"Server",
	"Content-Type",
	"Content-Language",
	"Date",
	"Last-Modified",
	"ETag",
	"Cache-Control",
	"Vary",
	"Content-Encoding",
	"X-Robots-Tag",
	"Link",
	"Alt-Svc",
	"X-Powered-By",
}
