package service

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gofurry/gofurry-nav-collector/collector/http/models"
	"github.com/gofurry/gofurry-nav-collector/roof/env"
)

// Every scenario calls the production performRequest, including its Transport,
// trace, proxy, redirect, body-budget and independent certificate diagnostics.
func TestHTTP2ProbeNegotiation(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Connection") != "" {
			t.Error("static hop-by-hop Connection header reached the endpoint")
		}
		w.WriteHeader(http.StatusForbidden) // HTTP status is not negotiation evidence.
		_, _ = io.WriteString(w, "<title>probe fixture</title>")
	})
	h2 := httptest.NewUnstartedServer(handler)
	h2.EnableHTTP2 = true
	h2.StartTLS()
	t.Cleanup(h2.Close)
	h1 := httptest.NewTLSServer(handler)
	t.Cleanup(h1.Close)
	plain := httptest.NewServer(handler)
	t.Cleanup(plain.Close)
	redirectH2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, h2.URL+"/final", http.StatusFound)
	}))
	t.Cleanup(redirectH2.Close)
	redirectHTTP := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, plain.URL+"/final", http.StatusMovedPermanently)
	}))
	redirectHTTP.EnableHTTP2 = true
	redirectHTTP.StartTLS()
	t.Cleanup(redirectHTTP.Close)
	for _, tc := range []struct{ name, initial, final, proto, alpn string }{
		{"TLS H2", h2.URL, h2.URL, "HTTP/2.0", "h2"},
		{"TLS H1", h1.URL, h1.URL, "HTTP/1.1", "http/1.1"},
		{"HTTP", plain.URL, plain.URL, "HTTP/1.1", ""},
		{"HTTP to H2", redirectH2.URL, h2.URL + "/final", "HTTP/2.0", "h2"},
		{"H2 to HTTP", redirectHTTP.URL, plain.URL + "/final", "HTTP/1.1", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := performRequest(http2FixtureSite(t, tc.initial, false))
			assertHTTP2Evidence(t, result, tc.final, tc.proto, tc.alpn)
			if strings.HasPrefix(tc.final, "https:") && (!result.CertCollected || result.CertVerified) {
				t.Fatalf("self-signed certificate diagnostics changed: collected=%v verified=%v", result.CertCollected, result.CertVerified)
			}
		})
	}
}

func TestHTTP2ProbeConnectProxy(t *testing.T) {
	target := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Proxy-Authorization") != "" || r.Header.Get("Connection") != "" {
			t.Error("proxy credentials/hop-by-hop header reached target")
		}
		w.WriteHeader(http.StatusForbidden)
	}))
	target.EnableHTTP2 = true
	target.StartTLS()
	t.Cleanup(target.Close)
	for _, secureProxy := range []bool{false, true} {
		t.Run(map[bool]string{false: "HTTP CONNECT", true: "HTTPS CONNECT"}[secureProxy], func(t *testing.T) {
			connected := make(chan string, 1)
			proxy := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodConnect || r.Host != strings.TrimPrefix(target.URL, "https://") {
					http.Error(w, "unexpected tunnel", http.StatusBadGateway)
					return
				}
				upstream, err := net.Dial("tcp", r.Host)
				if err != nil {
					t.Error(err)
					return
				}
				defer upstream.Close()
				client, buffered, err := w.(http.Hijacker).Hijack()
				if err != nil {
					t.Error(err)
					return
				}
				defer client.Close()
				_, _ = buffered.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n")
				_ = buffered.Flush()
				connected <- r.Method
				go func() { _, _ = io.Copy(upstream, buffered) }()
				_, _ = io.Copy(client, upstream)
			}))
			if secureProxy {
				proxy.StartTLS()
			} else {
				proxy.Start()
			}
			t.Cleanup(proxy.Close)
			original := env.GetServerConfig().Collector.Proxy
			env.GetServerConfig().Collector.Proxy = proxy.URL
			t.Cleanup(func() { env.GetServerConfig().Collector.Proxy = original })
			result := performRequest(http2FixtureSite(t, target.URL, true))
			assertHTTP2Evidence(t, result, target.URL, "HTTP/2.0", "h2")
			select {
			case <-connected:
			default:
				t.Fatal("probe bypassed configured proxy")
			}
			// Close the production client's idle tunnel before fixture cleanup.
			target.CloseClientConnections()
		})
	}
}

func TestHTTP2ProbeFailuresHaveNoTrustedProfile(t *testing.T) {
	brokenTLS := httptest.NewUnstartedServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	brokenTLS.TLS = &tls.Config{MinVersion: tls.VersionTLS12, MaxVersion: tls.VersionTLS12, CipherSuites: []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256}}
	brokenTLS.StartTLS()
	t.Cleanup(brokenTLS.Close)
	result := performRequest(http2FixtureSite(t, brokenTLS.URL, false))
	if result.ErrorCode != "http_probe_failed" || result.HTTP2ProbeProfile != "" || result.StatusCode != 0 {
		t.Fatalf("TLS failure became a trusted response: %+v", result)
	}
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusProxyAuthRequired)
	}))
	t.Cleanup(proxy.Close)
	original := env.GetServerConfig().Collector.Proxy
	t.Cleanup(func() { env.GetServerConfig().Collector.Proxy = original })
	for _, tc := range []struct{ proxy, code string }{
		{proxy.URL, "http_probe_failed"},
		{"http://test-user:do-not-log@%invalid", "http_proxy_config_invalid"},
	} {
		env.GetServerConfig().Collector.Proxy = tc.proxy
		result = performRequest(http2FixtureSite(t, brokenTLS.URL, true))
		if result.ErrorCode != tc.code || result.HTTP2ProbeProfile != "" || strings.Contains(result.ErrorMessage, "do-not-log") {
			t.Fatalf("unsafe/incorrect proxy failure: code=%s profile=%s message=%s", result.ErrorCode, result.HTTP2ProbeProfile, result.ErrorMessage)
		}
	}
}

func TestHTTP2ProbeRuntimeDisabledIsNotTrusted(t *testing.T) {
	t.Setenv("GODEBUG", "http2client=0")
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	server.EnableHTTP2 = true
	server.StartTLS()
	t.Cleanup(server.Close)
	result := performRequest(http2FixtureSite(t, server.URL, false))
	if result.ErrorCode != "" || result.HTTPProtocol != "HTTP/1.1" || result.HTTP2ProbeProfile != "" {
		t.Fatalf("disabled H2 mislabeled: proto=%s profile=%s error=%s", result.HTTPProtocol, result.HTTP2ProbeProfile, result.ErrorCode)
	}
}

func http2FixtureSite(t *testing.T, raw string, proxy bool) models.GfnCollectorDomain {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	site := models.GfnCollectorDomain{Name: u.Host, TLS: "0", Proxy: "0"}
	if u.Scheme == "https" {
		site.TLS = "1"
	}
	if proxy {
		site.Proxy = "1"
	}
	return site
}

func assertHTTP2Evidence(t *testing.T, result models.HTTPModel, final, proto, alpn string) {
	t.Helper()
	scheme := strings.SplitN(final, ":", 2)[0]
	if result.ErrorCode != "" || result.StatusCode != http.StatusForbidden || result.FinalURL != final || result.FinalScheme != scheme || result.HTTPProtocol != proto || result.TLSALPN != alpn || result.HTTP2ProbeProfile != "go_h2_enabled_v1" {
		t.Fatalf("final=%s scheme=%s proto=%s alpn=%q profile=%s status=%d error=%s", result.FinalURL, result.FinalScheme, result.HTTPProtocol, result.TLSALPN, result.HTTP2ProbeProfile, result.StatusCode, result.ErrorCode)
	}
	payload := buildHTTPObservationPayload(result, models.HTTPSaveModel{})
	for key, want := range map[string]string{"http2_probe_profile": "go_h2_enabled_v1", "final_url": final, "final_scheme": scheme, "http_protocol": proto, "tls_alpn": alpn} {
		if payload[key] != want {
			t.Fatalf("payload[%s]=%v want %s", key, payload[key], want)
		}
	}
	t.Logf("final scheme=%s resp.Proto=%s TLS ALPN=%q", scheme, proto, alpn)
}
