package contentsync

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HTTPNavClient struct {
	baseURL string
	client  *http.Client
}

type apiEnvelope[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type NavSite struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Domain  string `json:"domain"`
	Info    string `json:"info"`
	Country string `json:"country"`
	NSFW    string `json:"nsfw"`
	Welfare string `json:"welfare"`
}

type NavGroup struct {
	ID    string        `json:"id"`
	Name  string        `json:"name"`
	Sites NavGroupSites `json:"sites"`
}

type NavGroupSites []string

func (sites *NavGroupSites) UnmarshalJSON(data []byte) error {
	data = []byte(strings.TrimSpace(string(data)))
	if len(data) == 0 || string(data) == "null" {
		*sites = nil
		return nil
	}
	var items []any
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		if id := navGroupSiteID(item); id != "" {
			result = append(result, id)
		}
	}
	*sites = result
	return nil
}

func navGroupSiteID(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		return strings.TrimSpace(fmt.Sprint(typed))
	case map[string]any:
		for _, key := range []string{"id", "site_id"} {
			if id := navGroupSiteID(typed[key]); id != "" {
				return id
			}
		}
	}
	return ""
}

type NavSiteDetail struct {
	Name    string `json:"name"`
	Info    string `json:"info"`
	Country string `json:"country"`
	NSFW    string `json:"nsfw"`
	Welfare string `json:"welfare"`
}

type navV2SiteDetailPayload struct {
	Site struct {
		Name    string  `json:"name"`
		Info    string  `json:"info"`
		Country *string `json:"country"`
		NSFW    string  `json:"nsfw"`
		Welfare string  `json:"welfare"`
	} `json:"site"`
}

type NavHTTPRecord struct {
	Domain string `json:"domain"`
	URL    string `json:"url"`
	Title  string `json:"title"`
	Meta   struct {
		Description string `json:"description"`
	} `json:"meta"`
}

func NewHTTPNavClient(baseURL string, timeout time.Duration) *HTTPNavClient {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	return &HTTPNavClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: timeout},
	}
}

func (c *HTTPNavClient) ListSites(ctx context.Context, locale string) ([]NavSite, error) {
	var data []NavSite
	err := c.fetchJSONWithBase(ctx, c.versionedBaseURL("v2"), "/nav/sync/sites", map[string]string{"lang": normalizeNavLocale(locale)}, &data)
	return data, err
}

func (c *HTTPNavClient) ListGroups(ctx context.Context, locale string) ([]NavGroup, error) {
	var data []NavGroup
	err := c.fetchJSONWithBase(ctx, c.versionedBaseURL("v2"), "/nav/sync/site-groups", map[string]string{"lang": normalizeNavLocale(locale)}, &data)
	return data, err
}

func (c *HTTPNavClient) GetSiteDetail(ctx context.Context, id, locale string) (NavSiteDetail, error) {
	var data navV2SiteDetailPayload
	err := c.fetchJSONWithBase(
		ctx,
		c.versionedBaseURL("v2"),
		"/nav/sites/"+url.PathEscape(strings.TrimSpace(id))+"/detail",
		map[string]string{"lang": normalizeNavLocale(locale)},
		&data,
	)
	return NavSiteDetail{
		Name:    data.Site.Name,
		Info:    data.Site.Info,
		Country: stringPtrValue(data.Site.Country),
		NSFW:    data.Site.NSFW,
		Welfare: data.Site.Welfare,
	}, err
}

func (c *HTTPNavClient) GetSiteHTTP(ctx context.Context, domain string) (NavHTTPRecord, error) {
	var data NavHTTPRecord
	err := c.fetchJSON(ctx, "/nav/site/getSiteHttpRecord", map[string]string{"domain": domain}, &data)
	if err != nil && isOptionalLegacyNavHTTPError(err) {
		return NavHTTPRecord{}, nil
	}
	return data, err
}

func (c *HTTPNavClient) fetchJSON(ctx context.Context, endpoint string, query map[string]string, target any) error {
	return c.fetchJSONWithBase(ctx, c.baseURL, endpoint, query, target)
}

func (c *HTTPNavClient) fetchJSONWithBase(ctx context.Context, baseURL string, endpoint string, query map[string]string, target any) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("nav client is not configured")
	}
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return fmt.Errorf("rag.sync_nav_base_url is not configured")
	}
	fullURL := baseURL + endpoint
	if len(query) > 0 {
		values := url.Values{}
		for key, value := range query {
			if value = strings.TrimSpace(value); value != "" {
				values.Set(key, value)
			}
		}
		if encoded := values.Encode(); encoded != "" {
			fullURL += "?" + encoded
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, readErrorBody(resp.Body))
	}
	var envelope apiEnvelope[json.RawMessage]
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return err
	}
	if envelope.Code != 1 {
		message := strings.TrimSpace(envelope.Message)
		if message == "" {
			message = "upstream api returned failure"
		}
		return fmt.Errorf("%s", message)
	}
	if len(envelope.Data) == 0 {
		envelope.Data = []byte("null")
	}
	return json.Unmarshal(envelope.Data, target)
}

func (c *HTTPNavClient) versionedBaseURL(version string) string {
	base := strings.TrimRight(strings.TrimSpace(c.baseURL), "/")
	switch {
	case strings.Contains(base, "/api/v1"):
		return strings.Replace(base, "/api/v1", "/api/"+version, 1)
	case strings.Contains(base, "/api/v2"):
		return strings.Replace(base, "/api/v2", "/api/"+version, 1)
	default:
		return base + "/api/" + version
	}
}

func normalizeNavLocale(locale string) string {
	locale = strings.ToLower(strings.TrimSpace(locale))
	if strings.HasPrefix(locale, "en") {
		return "en"
	}
	return "zh"
}

func stringPtrValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func isOptionalLegacyNavHTTPError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unexpected status 404") || strings.Contains(message, "链接不存在")
}
