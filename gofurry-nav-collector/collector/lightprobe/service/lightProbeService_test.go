package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofurry/gofurry-nav-collector/collector/lightprobe/models"
)

func TestParseRDAPBootstrapSelectsServer(t *testing.T) {
	servers, err := parseRDAPBootstrap([]byte(`{
		"services": [
			[["com", "net"], ["https://rdap.example/"]],
			[["org"], ["https://rdap-org.example/"]]
		]
	}`))
	if err != nil {
		t.Fatalf("parseRDAPBootstrap() error = %v", err)
	}
	if servers["com"] != "https://rdap.example/" || servers["net"] != "https://rdap.example/" || servers["org"] != "https://rdap-org.example/" {
		t.Fatalf("unexpected RDAP servers: %+v", servers)
	}
}

func TestRegistrableDomainGroupsSubdomains(t *testing.T) {
	first, err := registrableDomain("www.example.com")
	if err != nil {
		t.Fatalf("registrableDomain(www.example.com) error = %v", err)
	}
	second, err := registrableDomain("api.example.com")
	if err != nil {
		t.Fatalf("registrableDomain(api.example.com) error = %v", err)
	}
	if first != "example.com" || second != first {
		t.Fatalf("registrable domains = %q and %q, want both example.com", first, second)
	}
}

func TestProbeRDAPExtractsGovernanceFields(t *testing.T) {
	var bootstrapHits int
	var rdapHits int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/bootstrap":
			bootstrapHits++
			_, _ = w.Write([]byte(`{"services":[[["test"],["` + "http://" + r.Host + `" ]]]}`))
		case "/domain/example.test":
			rdapHits++
			_, _ = w.Write([]byte(`{
				"status": ["active", "client transfer prohibited"],
				"events": [
					{"eventAction": "registration", "eventDate": "2020-01-01T00:00:00Z"},
					{"eventAction": "expiration", "eventDate": "2030-01-01T00:00:00Z"}
				],
				"nameservers": [{"ldhName": "ns2.example.test"}, {"ldhName": "ns1.example.test"}],
				"secureDNS": {"delegationSigned": true},
				"entities": [{
					"roles": ["registrar"],
					"vcardArray": ["vcard", [["fn", {}, "text", "Example Registrar"]]]
				}]
			}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	restoreRDAPBootstrapForTest(server.URL + "/bootstrap")
	defer restoreRDAPBootstrapForTest(rdapBootstrapURLDefault)

	result := probeRDAP(&http.Client{Timeout: time.Second}, "example.test")
	if result.Status != "success" || result.ErrorCode != "" {
		t.Fatalf("probeRDAP() = %+v, want success", result)
	}
	if bootstrapHits != 1 || rdapHits != 1 {
		t.Fatalf("bootstrapHits=%d rdapHits=%d, want 1 each", bootstrapHits, rdapHits)
	}
	if result.Payload["registrar"] != "Example Registrar" || result.Payload["expires_at"] != "2030-01-01T00:00:00Z" {
		t.Fatalf("unexpected RDAP payload: %+v", result.Payload)
	}
	nameservers, ok := result.Payload["nameservers"].([]string)
	if !ok || len(nameservers) != 2 || nameservers[0] != "ns1.example.test" {
		t.Fatalf("unexpected nameservers: %#v", result.Payload["nameservers"])
	}
	if signed, _ := result.Payload["dnssec_delegation_signed"].(bool); !signed {
		t.Fatalf("dnssec_delegation_signed should be true: %+v", result.Payload)
	}
}

func TestProbeRDAPNoServerReturnsFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"services":[[["com"],["https://rdap.example/"]]]}`))
	}))
	defer server.Close()

	restoreRDAPBootstrapForTest(server.URL)
	defer restoreRDAPBootstrapForTest(rdapBootstrapURLDefault)

	result := probeRDAP(&http.Client{Timeout: time.Second}, "example.test")
	if result.Status != "failure" || result.ErrorCode != "rdap_no_server" {
		t.Fatalf("probeRDAP() = %+v, want rdap_no_server failure", result)
	}
}

func TestParseRobotsPayload(t *testing.T) {
	payload := parseRobotsPayload([]byte(`
User-agent: *
Disallow: /
Sitemap: https://example.com/sitemap.xml
Sitemap: https://example.com/sitemap-news.xml
`), 1)

	if payload["global_disallow_all"] != true || payload["user_agent_star_present"] != true {
		t.Fatalf("unexpected robots flags: %+v", payload)
	}
	if payload["sitemap_count"] != 2 {
		t.Fatalf("sitemap_count = %v, want 2", payload["sitemap_count"])
	}
	sitemaps, _ := payload["sitemaps"].([]string)
	if len(sitemaps) != 1 {
		t.Fatalf("stored sitemaps = %#v, want one limited item", sitemaps)
	}
}

func TestProbeRobotsNotFoundIsSuccessWithExistsFalse(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	target := targetFromTestServer(server.URL)
	result := probeRobots(target, time.Second, 1024, 20)
	if result.Status != "success" {
		t.Fatalf("probeRobots() status = %q, want success", result.Status)
	}
	if result.Payload["exists"] != false || result.Payload["status_code"] != http.StatusNotFound {
		t.Fatalf("unexpected robots payload: %+v", result.Payload)
	}
}

func TestProbeRobotsTruncatesBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("User-agent: *\nSitemap: https://example.com/" + strings.Repeat("a", 2000)))
	}))
	defer server.Close()

	result := probeRobots(targetFromTestServer(server.URL), time.Second, 32, 20)
	if result.Status != "success" || result.Payload["body_truncated"] != true {
		t.Fatalf("probeRobots() = %+v, want truncated success", result)
	}
}

func TestProbeSecurityTXTFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/security.txt" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Path == "/security.txt" {
			_, _ = w.Write([]byte("Contact: mailto:security@example.com\nExpires: 2030-01-01T00:00:00Z\nPolicy: https://example.com/policy\nPreferred-Languages: zh,en\nCanonical: https://example.com/security.txt\n"))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	result := probeSecurityTXT(targetFromTestServer(server.URL), time.Second, 4096)
	if result.Status != "success" || result.Payload["exists"] != true || result.Payload["path_used"] != "/security.txt" {
		t.Fatalf("probeSecurityTXT() = %+v, want fallback success", result)
	}
	contacts, _ := result.Payload["contact"].([]string)
	if len(contacts) != 1 || contacts[0] != "mailto:security@example.com" {
		t.Fatalf("unexpected contacts: %#v", result.Payload["contact"])
	}
}

func TestParseSecurityTXTPayloadLimitsFields(t *testing.T) {
	longValue := strings.Repeat("a", lightProbeMaxTextLength+20)
	payload := parseSecurityTXTPayload([]byte("Contact: " + longValue + "\nPolicy: " + longValue + "\n"))

	contacts, _ := payload["contact"].([]string)
	policies, _ := payload["policy"].([]string)
	if len(contacts) != 1 || len([]rune(contacts[0])) != lightProbeMaxTextLength {
		t.Fatalf("contact not limited: %#v", contacts)
	}
	if len(policies) != 1 || len([]rune(policies[0])) != lightProbeMaxTextLength {
		t.Fatalf("policy not limited: %#v", policies)
	}
}

func TestHTTPClientInvalidProxyReturnsError(t *testing.T) {
	client, err := httpClientWithError(time.Second, "://bad-proxy", true)
	if err == nil || client != nil {
		t.Fatalf("httpClientWithError() client=%v err=%v, want invalid proxy error", client, err)
	}
}

func restoreRDAPBootstrapForTest(rawURL string) {
	rdapBootstrapMu.Lock()
	defer rdapBootstrapMu.Unlock()
	rdapBootstrapURL = rawURL
	rdapBootstrap = cachedRDAPBootstrap{}
}

func targetFromTestServer(rawURL string) models.GfnCollectorDomain {
	return models.GfnCollectorDomain{
		SiteID: 1,
		Name:   strings.TrimPrefix(rawURL, "http://"),
		TLS:    "0",
	}
}
