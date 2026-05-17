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
	baseURL              string
	client               *http.Client
	allowedMarkdownHosts map[string]struct{}
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
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Sites []string `json:"sites"`
}

type NavSiteDetail struct {
	Name    string `json:"name"`
	Info    string `json:"info"`
	Country string `json:"country"`
	NSFW    string `json:"nsfw"`
	Welfare string `json:"welfare"`
}

type NavHTTPRecord struct {
	Domain string `json:"domain"`
	URL    string `json:"url"`
	Title  string `json:"title"`
	Meta   struct {
		Description string `json:"description"`
	} `json:"meta"`
}

type ChangeLog struct {
	Title      string `json:"title"`
	URL        string `json:"url"`
	CreateTime string `json:"create_time"`
	UpdateTime string `json:"update_time"`
}

func NewHTTPNavClient(baseURL string, timeout time.Duration, allowedMarkdownHosts []string) *HTTPNavClient {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	return &HTTPNavClient{
		baseURL:              baseURL,
		client:               &http.Client{Timeout: timeout},
		allowedMarkdownHosts: buildAllowedMarkdownHosts(baseURL, allowedMarkdownHosts),
	}
}

func (c *HTTPNavClient) ListSites(ctx context.Context, locale string) ([]NavSite, error) {
	var data []NavSite
	if err := c.fetchJSON(ctx, "/nav/page/site/list", map[string]string{"lang": locale}, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (c *HTTPNavClient) ListGroups(ctx context.Context, locale string) ([]NavGroup, error) {
	var data []NavGroup
	if err := c.fetchJSON(ctx, "/nav/page/group/list", map[string]string{"lang": locale}, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (c *HTTPNavClient) GetSiteDetail(ctx context.Context, id, locale string) (NavSiteDetail, error) {
	var data NavSiteDetail
	err := c.fetchJSON(ctx, "/nav/site/getSiteDetail", map[string]string{"id": id, "lang": locale}, &data)
	return data, err
}

func (c *HTTPNavClient) GetSiteHTTP(ctx context.Context, domain string) (NavHTTPRecord, error) {
	var data NavHTTPRecord
	err := c.fetchJSON(ctx, "/nav/site/getSiteHttpRecord", map[string]string{"domain": domain}, &data)
	return data, err
}

func (c *HTTPNavClient) ListChangelogs(ctx context.Context) ([]ChangeLog, error) {
	var data []ChangeLog
	if err := c.fetchJSON(ctx, "/site/changelog", nil, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (c *HTTPNavClient) FetchMarkdown(ctx context.Context, rawURL string) (string, error) {
	target, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", err
	}
	if target.Scheme != "http" && target.Scheme != "https" {
		return "", fmt.Errorf("unsupported markdown url scheme %q", target.Scheme)
	}
	host := strings.ToLower(target.Host)
	if _, ok := c.allowedMarkdownHosts[host]; !ok {
		return "", fmt.Errorf("markdown url host %q is not allowed", host)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "text/markdown,text/plain;q=0.9,*/*;q=0.8")
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	body, err := readLimitedString(resp.Body, maxMarkdownBytes)
	if err != nil {
		return "", err
	}
	return body, nil
}

func (c *HTTPNavClient) fetchJSON(ctx context.Context, endpoint string, query map[string]string, target any) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("nav client is not configured")
	}
	if c.baseURL == "" {
		return fmt.Errorf("rag.sync_nav_base_url is not configured")
	}
	fullURL := c.baseURL + endpoint
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

func buildAllowedMarkdownHosts(baseURL string, configured []string) map[string]struct{} {
	result := map[string]struct{}{
		"raw.githubusercontent.com": {},
	}
	if parsed, err := url.Parse(strings.TrimSpace(baseURL)); err == nil && parsed.Host != "" {
		result[strings.ToLower(parsed.Host)] = struct{}{}
	}
	for _, host := range configured {
		host = strings.ToLower(strings.TrimSpace(host))
		host = strings.TrimPrefix(host, "http://")
		host = strings.TrimPrefix(host, "https://")
		host = strings.TrimRight(host, "/")
		if idx := strings.IndexByte(host, '/'); idx >= 0 {
			host = host[:idx]
		}
		if host != "" {
			result[host] = struct{}{}
		}
	}
	return result
}
