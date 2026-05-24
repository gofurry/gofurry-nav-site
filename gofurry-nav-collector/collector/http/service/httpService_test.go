package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net/http"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-collector/collector/http/models"
	"github.com/gofurry/gofurry-nav-collector/roof/env"
)

func TestBuildHTTPObservationPayloadAddsV2FieldsAndLimitsExternalText(t *testing.T) {
	longURL := "https://example.com/" + repeated("u", 2100)
	notBefore := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	notAfter := notBefore.Add(24 * time.Hour)
	result := models.HTTPModel{
		ResponseTime:    123,
		DNSLookupMS:     4,
		TCPConnectMS:    7,
		TLSHandshakeMS:  11,
		TTFBMS:          23,
		TransferMS:      31,
		HTTPProtocol:    "HTTP/2.0",
		RemoteAddr:      "203.0.113.5:443",
		RemoteIP:        "203.0.113.5",
		BodyReadBytes:   1987,
		BodyTruncated:   true,
		BodyLimitBytes:  1048576,
		ContentEncoding: "gzip",
		Compressed:      true,
		CacheControl:    repeated("c", 600),
		ETag:            repeated("e", 600),
		LastModified:    repeated("l", 600),
		Redirects: []string{
			longURL,
			"https://final.example/path",
		},
		FinalURL:    longURL,
		ContentType: repeated("c", 300),
		SecurityHeaders: map[string]bool{
			"strict_transport_security": true,
			"x_frame_options":           true,
		},
		SecurityHeaderValues: map[string]string{
			"strict_transport_security": "max-age=31536000; includeSubDomains; preload",
			"content_security_policy":   "default-src 'self'; script-src 'unsafe-inline' *",
			"x_frame_options":           "DENY",
			"x_content_type_options":    "nosniff",
			"referrer_policy":           "strict-origin-when-cross-origin",
			"permissions_policy":        "geolocation=()",
		},
		CertCollected:       true,
		CertVerified:        true,
		VerifyError:         repeated("e", 600),
		VerifyErrorCategory: "",
		TLSHandshake:        tlsHandshakeCollected,
		CertNotBefore:       notBefore,
		CertNotAfter:        notAfter,
		CertChainLen:        2,
		CertSubjectCN:       repeated("n", 300),
		CertSANCount:        5,
		OCSPStapled:         true,
		SCTCount:            3,
	}
	httpRecord := models.HTTPSaveModel{
		Domain:        "example.com",
		Url:           "https://example.com",
		StatusCode:    http.StatusOK,
		ContentLength: 2048,
		Title:         repeated("题", 300),
		Server:        repeated("s", 300),
		Redirects:     result.Redirects,
		Headers: map[string][]string{
			"Server":       {repeated("s", 600)},
			"Content-Type": {"text/html; charset=utf-8"},
		},
		Meta: map[string]string{
			"description": repeated("描", 600),
		},
		TLSVersion:  "TLS1.3",
		CipherSuite: "AES_128_GCM_SHA256",
		CertIssuer:  repeated("i", 300),
		CertIssuerOrg: []string{
			repeated("o", 300),
		},
		CertDNSNames: repeatedSlice("dns", 70),
		CertEmail: []string{
			repeated("m", 300),
		},
	}

	payload := buildHTTPObservationPayload(result, httpRecord)

	if payload["domain"] != "example.com" {
		t.Fatalf("兼容字段 domain 丢失，got %v", payload["domain"])
	}
	if payload["url"] != "https://example.com" {
		t.Fatalf("兼容字段 url 丢失，got %v", payload["url"])
	}
	if payload["status_code"] != int64(http.StatusOK) {
		t.Fatalf("兼容字段 status_code 丢失，got %v", payload["status_code"])
	}
	if payload["response_time_ms"] != int64(123) {
		t.Fatalf("兼容字段 response_time_ms 丢失，got %v", payload["response_time_ms"])
	}

	redirectChain, ok := payload["redirect_chain"].([]string)
	if !ok {
		t.Fatalf("redirect_chain 类型错误: %T", payload["redirect_chain"])
	}
	if len(redirectChain) != 2 {
		t.Fatalf("redirect_chain 数量错误: %d", len(redirectChain))
	}
	if got := runeLen(redirectChain[0]); got != maxHTTPPayloadURLLength {
		t.Fatalf("redirect_chain URL 未限长，got %d", got)
	}
	if payload["redirect_count"] != 2 {
		t.Fatalf("redirect_count 错误，got %v", payload["redirect_count"])
	}
	if got := runeLen(payload["final_url"].(string)); got != maxHTTPPayloadURLLength {
		t.Fatalf("final_url 未限长，got %d", got)
	}
	if got := runeLen(payload["content_type"].(string)); got != maxHTTPPayloadContentTypeLength {
		t.Fatalf("content_type 未限长，got %d", got)
	}
	if got := runeLen(payload["title"].(string)); got != maxHTTPPayloadTitleLength {
		t.Fatalf("title 未限长，got %d", got)
	}
	if got := runeLen(payload["server"].(string)); got != maxHTTPPayloadServerLength {
		t.Fatalf("server 未限长，got %d", got)
	}

	meta := payload["meta"].(map[string]string)
	if got := runeLen(meta["description"]); got != maxHTTPPayloadMetaValueLength {
		t.Fatalf("meta value 未限长，got %d", got)
	}
	headers := payload["headers"].(map[string][]string)
	if got := runeLen(headers["Server"][0]); got != maxHTTPPayloadHeaderValueLength {
		t.Fatalf("header value 未限长，got %d", got)
	}
	if got := runeLen(payload["cert_issuer"].(string)); got != maxHTTPPayloadCertTextLength {
		t.Fatalf("cert_issuer 未限长，got %d", got)
	}
	certIssuerOrg := payload["cert_issuer_org"].([]string)
	if got := runeLen(certIssuerOrg[0]); got != maxHTTPPayloadCertTextLength {
		t.Fatalf("cert_issuer_org 未限长，got %d", got)
	}
	certDNSNames := payload["cert_dns_names"].([]string)
	if got := len(certDNSNames); got != maxHTTPPayloadCertItems {
		t.Fatalf("cert_dns_names 未限制条数，got %d", got)
	}
	certEmail := payload["cert_email"].([]string)
	if got := runeLen(certEmail[0]); got != maxHTTPPayloadCertTextLength {
		t.Fatalf("cert_email 未限长，got %d", got)
	}
	if got := runeLen(payload["verify_error"].(string)); got != maxHTTPPayloadVerifyErrorLength {
		t.Fatalf("verify_error 未限长，got %d", got)
	}
	if payload["cert_collected"] != true {
		t.Fatalf("cert_collected 错误，got %v", payload["cert_collected"])
	}
	if payload["cert_verified"] != true {
		t.Fatalf("cert_verified 错误，got %v", payload["cert_verified"])
	}
	if payload["tls_handshake"] != tlsHandshakeCollected {
		t.Fatalf("tls_handshake 错误，got %v", payload["tls_handshake"])
	}
	if payload["http_protocol"] != "HTTP/2.0" {
		t.Fatalf("http_protocol 错误，got %v", payload["http_protocol"])
	}
	if payload["dns_lookup_ms"] != int64(4) || payload["tcp_connect_ms"] != int64(7) || payload["tls_handshake_ms"] != int64(11) {
		t.Fatalf("trace timings 错误，got dns=%v tcp=%v tls=%v", payload["dns_lookup_ms"], payload["tcp_connect_ms"], payload["tls_handshake_ms"])
	}
	if payload["ttfb_ms"] != int64(23) || payload["transfer_ms"] != int64(31) {
		t.Fatalf("ttfb/transfer 错误，got ttfb=%v transfer=%v", payload["ttfb_ms"], payload["transfer_ms"])
	}
	if payload["remote_addr"] != "203.0.113.5:443" || payload["remote_ip"] != "203.0.113.5" {
		t.Fatalf("remote addr/ip 错误，got addr=%v ip=%v", payload["remote_addr"], payload["remote_ip"])
	}
	if payload["body_read_bytes"] != int64(1987) {
		t.Fatalf("body_read_bytes 错误，got %v", payload["body_read_bytes"])
	}
	if payload["body_truncated"] != true {
		t.Fatalf("body_truncated 错误，got %v", payload["body_truncated"])
	}
	if payload["body_limit_bytes"] != int64(1048576) {
		t.Fatalf("body_limit_bytes 错误，got %v", payload["body_limit_bytes"])
	}
	if payload["compressed"] != true {
		t.Fatalf("compressed 错误，got %v", payload["compressed"])
	}
	if payload["content_encoding"] != "gzip" {
		t.Fatalf("content_encoding 错误，got %v", payload["content_encoding"])
	}
	cachePolicy, ok := payload["cache_policy"].(map[string]string)
	if !ok {
		t.Fatalf("cache_policy 类型错误: %T", payload["cache_policy"])
	}
	if got := runeLen(cachePolicy["cache_control"]); got != maxHTTPPayloadCacheValueLength {
		t.Fatalf("cache_control 未限长，got %d", got)
	}
	if got := runeLen(cachePolicy["etag"]); got != maxHTTPPayloadCacheValueLength {
		t.Fatalf("etag 未限长，got %d", got)
	}
	if got := runeLen(cachePolicy["last_modified"]); got != maxHTTPPayloadCacheValueLength {
		t.Fatalf("last_modified 未限长，got %d", got)
	}
	securityHeaderSummary, ok := payload["security_header_summary"].(map[string]any)
	if !ok {
		t.Fatalf("security_header_summary 类型错误: %T", payload["security_header_summary"])
	}
	hstsSummary, ok := securityHeaderSummary["hsts"].(map[string]any)
	if !ok {
		t.Fatalf("hsts summary 类型错误: %T", securityHeaderSummary["hsts"])
	}
	if hstsSummary["present"] != true || hstsSummary["max_age"] != int64(31536000) || hstsSummary["include_subdomains"] != true || hstsSummary["preload"] != true {
		t.Fatalf("hsts summary 错误: %+v", hstsSummary)
	}
	cspSummary, ok := securityHeaderSummary["content_security_policy"].(map[string]any)
	if !ok {
		t.Fatalf("csp summary 类型错误: %T", securityHeaderSummary["content_security_policy"])
	}
	if cspSummary["has_default_src"] != true || cspSummary["unsafe_inline"] != true || cspSummary["wildcard_source"] != true {
		t.Fatalf("csp summary 错误: %+v", cspSummary)
	}
	xfoSummary, ok := securityHeaderSummary["x_frame_options"].(map[string]any)
	if !ok {
		t.Fatalf("x_frame_options summary 类型错误: %T", securityHeaderSummary["x_frame_options"])
	}
	if xfoSummary["mode"] != "DENY" {
		t.Fatalf("x_frame_options mode 错误: %+v", xfoSummary)
	}
	xctoSummary, ok := securityHeaderSummary["x_content_type_options"].(map[string]any)
	if !ok {
		t.Fatalf("x_content_type_options summary 类型错误: %T", securityHeaderSummary["x_content_type_options"])
	}
	if xctoSummary["nosniff"] != true {
		t.Fatalf("x_content_type_options summary 错误: %+v", xctoSummary)
	}
	if payload["cert_not_before"] != notBefore.Format(time.RFC3339) {
		t.Fatalf("cert_not_before 错误，got %v", payload["cert_not_before"])
	}
	if payload["cert_not_after"] != notAfter.Format(time.RFC3339) {
		t.Fatalf("cert_not_after 错误，got %v", payload["cert_not_after"])
	}
	if payload["cert_chain_length"] != 2 {
		t.Fatalf("cert_chain_length 错误，got %v", payload["cert_chain_length"])
	}
	if got := runeLen(payload["cert_subject_cn"].(string)); got != maxHTTPPayloadCertTextLength {
		t.Fatalf("cert_subject_cn 未限长，got %d", got)
	}
	if payload["cert_san_count"] != 5 {
		t.Fatalf("cert_san_count 错误，got %v", payload["cert_san_count"])
	}
	if payload["ocsp_stapled"] != true || payload["sct_count"] != 3 {
		t.Fatalf("ocsp/sct 错误，got ocsp=%v sct=%v", payload["ocsp_stapled"], payload["sct_count"])
	}
	if payload["cert_public_key_algorithm"] != httpRecord.CertPubKeyAlg || payload["cert_signature_algorithm"] != httpRecord.CertSigAlg {
		t.Fatalf("规范 TLS 算法字段错误，got pub=%v sig=%v", payload["cert_public_key_algorithm"], payload["cert_signature_algorithm"])
	}
	if _, ok := payload["redirects"]; !ok {
		t.Fatal("兼容字段 redirects 丢失")
	}
	if _, ok := payload["headers"]; !ok {
		t.Fatal("兼容字段 headers 丢失")
	}
	if _, ok := payload["meta"]; !ok {
		t.Fatal("兼容字段 meta 丢失")
	}
}

func TestBuildHTTPObservationPayloadSecurityHeadersHaveDefaults(t *testing.T) {
	payload := buildHTTPObservationPayload(models.HTTPModel{}, models.HTTPSaveModel{})

	securityHeaders, ok := payload["security_headers"].(map[string]bool)
	if !ok {
		t.Fatalf("security_headers 类型错误: %T", payload["security_headers"])
	}
	expectedKeys := []string{
		"strict_transport_security",
		"content_security_policy",
		"x_frame_options",
		"x_content_type_options",
		"referrer_policy",
		"permissions_policy",
	}
	for _, key := range expectedKeys {
		if securityHeaders[key] {
			t.Fatalf("缺失安全响应头时 %s 应为 false", key)
		}
	}
	securityHeaderSummary, ok := payload["security_header_summary"].(map[string]any)
	if !ok {
		t.Fatalf("security_header_summary 类型错误: %T", payload["security_header_summary"])
	}
	hstsSummary := securityHeaderSummary["hsts"].(map[string]any)
	if hstsSummary["present"] != false || hstsSummary["max_age"] != nil {
		t.Fatalf("缺省 hsts summary 错误: %+v", hstsSummary)
	}
	cspSummary := securityHeaderSummary["content_security_policy"].(map[string]any)
	if cspSummary["present"] != false || cspSummary["has_default_src"] != false || cspSummary["unsafe_inline"] != false {
		t.Fatalf("缺省 csp summary 错误: %+v", cspSummary)
	}
}

func TestBuildHTTPObservationPayloadTLSDefaultsForNoTLS(t *testing.T) {
	payload := buildHTTPObservationPayload(models.HTTPModel{
		TLSHandshake: tlsHandshakeNotTLS,
	}, models.HTTPSaveModel{})

	if payload["cert_collected"] != false {
		t.Fatalf("无 TLS 时 cert_collected 应为 false，got %v", payload["cert_collected"])
	}
	if payload["cert_verified"] != false {
		t.Fatalf("无 TLS 时 cert_verified 应为 false，got %v", payload["cert_verified"])
	}
	if payload["tls_handshake"] != tlsHandshakeNotTLS {
		t.Fatalf("无 TLS 时 tls_handshake 应为 not_tls，got %v", payload["tls_handshake"])
	}
}

func TestEnrichHTTPPageDetailsExtractsPageSemantics(t *testing.T) {
	body := []byte(`<!doctype html>
<html lang="zh-CN">
<head>
  <title>  示例站点  </title>
  <meta charset="utf-8">
  <meta name="description" content="  面向访客的介绍  ">
  <meta name="keywords" content="furry, nav">
  <meta name="author" content="GoFurry">
  <meta name="generator" content="Nuxt">
  <meta name="application-name" content="GoFurry Nav">
  <meta name="theme-color" content="#ffffff">
  <meta name="robots" content="index,follow">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta property="og:title" content="OG 标题">
  <meta property="og:description" content="OG 描述">
  <meta property="og:site_name" content="GoFurry">
  <meta property="og:type" content="website">
  <meta property="og:image" content="/og.png">
  <meta property="og:url" content="/share">
  <meta name="twitter:card" content="summary_large_image">
  <meta name="twitter:title" content="Twitter 标题">
  <meta name="twitter:description" content="Twitter 描述">
  <meta name="twitter:image" content="/twitter.png">
  <meta name="twitter:site" content="@gofurry">
  <meta http-equiv="refresh" content="5; url=/jump">
  <link rel="canonical" href="/canonical">
  <link rel="icon" href="/favicon.ico" type="image/x-icon" sizes="32x32">
  <link rel="apple-touch-icon" href="/apple.png">
</head><body></body></html>`)
	result := models.HTTPModel{
		Domain:          "example.com",
		Url:             "https://example.com",
		FinalURL:        "https://example.com/path/index.html",
		Meta:            map[string]string{},
		ServerHints:     &models.HTTPServerHints{Server: "nginx"},
		ContentLanguage: "",
	}

	enrichHTTPPageDetails(&result, body)

	if result.Title != "示例站点" {
		t.Fatalf("title = %q", result.Title)
	}
	if result.HTMLLang != "zh-CN" || result.ContentLanguage != "zh-CN" {
		t.Fatalf("语言提取错误 html=%q effective=%q", result.HTMLLang, result.ContentLanguage)
	}
	if result.Meta["charset"] != "utf-8" || result.Meta["description"] != "面向访客的介绍" {
		t.Fatalf("meta 提取错误: %+v", result.Meta)
	}
	if result.Meta["application_name"] != "GoFurry Nav" || result.Meta["theme_color"] != "#ffffff" || result.Meta["robots"] != "index,follow" {
		t.Fatalf("扩展 meta 提取错误: %+v", result.Meta)
	}
	if result.OpenGraph["image"] != "https://example.com/og.png" || result.OpenGraph["url"] != "https://example.com/share" {
		t.Fatalf("OpenGraph URL 解析错误: %+v", result.OpenGraph)
	}
	if result.TwitterCard["card"] != "summary_large_image" || result.TwitterCard["image"] != "https://example.com/twitter.png" {
		t.Fatalf("Twitter Card 提取错误: %+v", result.TwitterCard)
	}
	if result.CanonicalURL != "https://example.com/canonical" {
		t.Fatalf("canonical URL = %q", result.CanonicalURL)
	}
	if result.MetaRefresh == nil || !result.MetaRefresh.Present || result.MetaRefresh.DelaySeconds == nil || *result.MetaRefresh.DelaySeconds != 5 || result.MetaRefresh.URL != "https://example.com/jump" {
		t.Fatalf("meta refresh 提取错误: %+v", result.MetaRefresh)
	}
	if len(result.IconLinks) != 2 || result.IconLinks[0].Href != "https://example.com/favicon.ico" {
		t.Fatalf("icon links 提取错误: %+v", result.IconLinks)
	}
	if result.SharePreview == nil || result.SharePreview.Title != "OG 标题" || result.SharePreview.Image != "https://example.com/og.png" {
		t.Fatalf("share preview 提取错误: %+v", result.SharePreview)
	}
	if result.PageTextSummary == "" {
		t.Fatal("page text summary 不应为空")
	}
	if result.RedirectHint == nil || !result.RedirectHint.CanonicalURLDifferent || !result.RedirectHint.MetaRefreshPresent {
		t.Fatalf("redirect hint 提取错误: %+v", result.RedirectHint)
	}
	if result.ServerHints.Generator != "Nuxt" {
		t.Fatalf("server hints generator = %q", result.ServerHints.Generator)
	}
}

func TestBuildHTTPCookieSummaryDoesNotExposeRawCookie(t *testing.T) {
	header := http.Header{}
	header.Add("Set-Cookie", "sid=secret; Path=/; Secure; HttpOnly; SameSite=Lax")
	header.Add("Set-Cookie", "theme=dark; Path=/; SameSite=None")
	header.Add("Set-Cookie", "strict=1; Path=/; SameSite=Strict")
	summary := buildHTTPCookieSummary(header.Values("Set-Cookie"))

	if summary == nil {
		t.Fatal("cookie summary 不应为空")
	}
	if summary.SetCookieCount != 3 || summary.SecureCount != 1 || summary.HTTPOnlyCount != 1 {
		t.Fatalf("cookie summary 基础计数错误: %+v", summary)
	}
	if summary.SameSiteLaxCount != 1 || summary.SameSiteNoneCount != 1 || summary.SameSiteStrictCount != 1 {
		t.Fatalf("SameSite 计数错误: %+v", summary)
	}
}

func TestBuildHTTPObservationPayloadIncludesPageDetailFields(t *testing.T) {
	delay := int64(3)
	httpRecord := models.HTTPSaveModel{
		Domain:       "example.com",
		Url:          "https://example.com",
		StatusCode:   http.StatusOK,
		Title:        repeated("t", 300),
		Headers:      map[string][]string{"X-Powered-By": {"Next.js"}},
		Meta:         map[string]string{"description": repeated("d", 600), "viewport": "width=device-width"},
		OpenGraph:    map[string]string{"title": repeated("o", 600), "image": "https://example.com/og.png"},
		TwitterCard:  map[string]string{"card": "summary", "site": "@gofurry"},
		CanonicalURL: "https://example.com/canonical",
		HTMLLang:     "zh-CN",
		MetaRefresh:  &models.HTTPMetaRefresh{Present: true, DelaySeconds: &delay, URL: "https://example.com/jump"},
		IconLinks: []models.HTTPLinkInfo{
			{Rel: "icon", Href: "https://example.com/favicon.ico", Type: "image/x-icon", Sizes: "32x32"},
		},
		CookieSummary: &models.HTTPCookieSummary{SetCookieCount: 2, SecureCount: 1},
		ServerHints:   &models.HTTPServerHints{Server: "nginx", XPoweredBy: "Next.js", Generator: "Nuxt"},
		CrossOrigin:   &models.HTTPCrossOrigin{CrossOriginOpenerPolicy: "same-origin", AccessControlAllowOrigin: "*"},
		ContentLang:   "zh-CN",
		PageSummary:   repeated("摘", 1200),
		SharePreview:  &models.HTTPSharePreview{Title: "分享标题", Description: "分享描述", SiteName: "GoFurry", Image: "https://example.com/og.png", URL: "https://example.com"},
		RedirectHint:  &models.HTTPRedirectHint{FinalURLDifferent: true, MetaRefreshPresent: true, MetaRefreshURL: "https://example.com/jump"},
	}

	payload := buildHTTPObservationPayload(models.HTTPModel{}, httpRecord)

	if got := runeLen(payload["canonical_url"].(string)); got == 0 || got > maxHTTPPayloadURLLength {
		t.Fatalf("canonical_url 限长错误，got %d", got)
	}
	meta := payload["meta"].(map[string]string)
	if got := runeLen(meta["description"]); got != maxHTTPPayloadMetaValueLength {
		t.Fatalf("meta description 未限长，got %d", got)
	}
	openGraph := payload["open_graph"].(map[string]string)
	if got := runeLen(openGraph["title"]); got != maxHTTPPayloadMetaValueLength {
		t.Fatalf("open_graph title 未限长，got %d", got)
	}
	if payload["html_lang"] != "zh-CN" || payload["content_language_effective"] != "zh-CN" {
		t.Fatalf("语言字段错误 html=%v effective=%v", payload["html_lang"], payload["content_language_effective"])
	}
	metaRefresh := payload["meta_refresh"].(*models.HTTPMetaRefresh)
	if metaRefresh.URL != "https://example.com/jump" || metaRefresh.DelaySeconds == nil || *metaRefresh.DelaySeconds != 3 {
		t.Fatalf("meta_refresh 错误: %+v", metaRefresh)
	}
	iconLinks := payload["icon_links"].([]models.HTTPLinkInfo)
	if len(iconLinks) != 1 || iconLinks[0].Href != "https://example.com/favicon.ico" {
		t.Fatalf("icon_links 错误: %+v", iconLinks)
	}
	if payload["cookie_summary"].(*models.HTTPCookieSummary).SetCookieCount != 2 {
		t.Fatalf("cookie_summary 错误: %+v", payload["cookie_summary"])
	}
	serverHints := payload["server_hints"].(*models.HTTPServerHints)
	if serverHints.XPoweredBy != "Next.js" || serverHints.Generator != "Nuxt" {
		t.Fatalf("server_hints 错误: %+v", serverHints)
	}
	crossOrigin := payload["cross_origin_summary"].(*models.HTTPCrossOrigin)
	if crossOrigin.CrossOriginOpenerPolicy != "same-origin" || crossOrigin.AccessControlAllowOrigin != "*" {
		t.Fatalf("cross_origin_summary 错误: %+v", crossOrigin)
	}
	if got := runeLen(payload["page_text_summary"].(string)); got != maxHTTPPayloadPageSummaryLength {
		t.Fatalf("page_text_summary 未限长，got %d", got)
	}
	sharePreview := payload["share_preview"].(*models.HTTPSharePreview)
	if sharePreview.Title != "分享标题" || sharePreview.Image != "https://example.com/og.png" {
		t.Fatalf("share_preview 错误: %+v", sharePreview)
	}
	redirectHint := payload["redirect_hint"].(*models.HTTPRedirectHint)
	if !redirectHint.FinalURLDifferent || !redirectHint.MetaRefreshPresent {
		t.Fatalf("redirect_hint 错误: %+v", redirectHint)
	}
}

func TestVerifyTLSCertificateWithRoots(t *testing.T) {
	cert := newTestCertificate(t, "example.com")
	roots := x509.NewCertPool()
	roots.AddCert(cert)
	state := &tls.ConnectionState{PeerCertificates: []*x509.Certificate{cert}}

	verified, verifyError := verifyTLSCertificateWithRoots(state, "example.com", roots)
	if !verified {
		t.Fatalf("证书应校验通过: %s", verifyError)
	}
	if verifyError != "" {
		t.Fatalf("校验通过时 verify_error 应为空，got %q", verifyError)
	}

	verified, verifyError = verifyTLSCertificateWithRoots(state, "wrong.example.com", roots)
	if verified {
		t.Fatal("域名不匹配时证书不应校验通过")
	}
	if verifyError == "" {
		t.Fatal("域名不匹配时 verify_error 不应为空")
	}
}

func TestVerifyTLSCertificateMissingCertificate(t *testing.T) {
	verified, verifyError := verifyTLSCertificateWithRoots(&tls.ConnectionState{}, "example.com", nil)
	if verified {
		t.Fatal("没有证书时不应校验通过")
	}
	if verifyError == "" {
		t.Fatal("没有证书时 verify_error 不应为空")
	}
}

func TestVerifyTLSCertificateDetailedClassifiesErrors(t *testing.T) {
	cert := newTestCertificate(t, "example.com")
	roots := x509.NewCertPool()
	roots.AddCert(cert)
	state := &tls.ConnectionState{PeerCertificates: []*x509.Certificate{cert}}

	verified, verifyError, category := verifyTLSCertificateDetailed(state, "wrong.example.com", roots)
	if verified {
		t.Fatal("域名不匹配时不应校验通过")
	}
	if verifyError == "" {
		t.Fatal("域名不匹配时 verify_error 不应为空")
	}
	if category != "hostname_mismatch" {
		t.Fatalf("category = %q, want hostname_mismatch", category)
	}
}

func TestDetectHTTPSecurityHeaders(t *testing.T) {
	header := http.Header{}
	header.Set("Strict-Transport-Security", "max-age=31536000")
	header.Set("Content-Security-Policy", "default-src 'self'")
	header.Set("X-Frame-Options", "DENY")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	header.Set("Permissions-Policy", "geolocation=()")

	securityHeaders := detectHTTPSecurityHeaders(header)
	for key, value := range securityHeaders {
		if !value {
			t.Fatalf("安全响应头 %s 应为 true", key)
		}
	}
}

func TestPerformRequestReturnsFailureForInvalidProxyConfig(t *testing.T) {
	oldProxy := env.GetServerConfig().Collector.Proxy
	env.GetServerConfig().Collector.Proxy = "://bad proxy"
	t.Cleanup(func() {
		env.GetServerConfig().Collector.Proxy = oldProxy
	})

	result := performRequest(models.GfnCollectorDomain{
		Name:  "example.com",
		Proxy: "1",
		TLS:   "1",
	})

	if result.ErrorCode != "http_proxy_config_invalid" {
		t.Fatalf("ErrorCode = %q, want http_proxy_config_invalid", result.ErrorCode)
	}
	if result.StatusCode != 0 {
		t.Fatalf("StatusCode = %d, want 0", result.StatusCode)
	}
	if result.ErrorMessage == "" {
		t.Fatal("ErrorMessage should explain invalid proxy config")
	}
}

func repeated(value string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += value
	}
	return result
}

func repeatedSlice(value string, count int) []string {
	result := make([]string, 0, count)
	for i := 0; i < count; i++ {
		result = append(result, repeated(value, 300))
	}
	return result
}

func runeLen(value string) int {
	return len([]rune(value))
}

func newTestCertificate(t *testing.T, dnsName string) *x509.Certificate {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成测试私钥失败: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: dnsName,
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		DNSNames:              []string{dnsName},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("生成测试证书失败: %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("解析测试证书失败: %v", err)
	}
	return cert
}
