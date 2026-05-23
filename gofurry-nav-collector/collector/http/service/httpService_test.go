package service

import (
	"net/http"
	"testing"

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

func runeLen(value string) int {
	return len([]rune(value))
}
