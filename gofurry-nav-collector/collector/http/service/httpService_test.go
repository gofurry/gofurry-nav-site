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
