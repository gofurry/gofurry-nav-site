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
)

func TestBuildHTTPObservationPayloadAddsV2FieldsAndLimitsExternalText(t *testing.T) {
	longURL := "https://example.com/" + repeated("u", 2100)
	result := models.HTTPModel{
		ResponseTime: 123,
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
		CertCollected: true,
		CertVerified:  true,
		VerifyError:   repeated("e", 600),
		TLSHandshake:  tlsHandshakeCollected,
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
