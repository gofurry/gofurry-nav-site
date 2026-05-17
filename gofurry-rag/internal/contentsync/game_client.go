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

type HTTPGameClient struct {
	baseURL string
	client  *http.Client
}

type GameSummary struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Info        string   `json:"info"`
	ReleaseDate string   `json:"release_date"`
	Developers  []string `json:"developers"`
	Publishers  []string `json:"publishers"`
}

type GameKV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type GameTag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Desc string `json:"desc"`
}

type GamePCRequirements struct {
	Minimum     string `json:"minimum"`
	Recommended string `json:"recommended"`
}

type GameDetail struct {
	Name                string             `json:"name"`
	Info                string             `json:"info"`
	Resources           []GameKV           `json:"resources"`
	Groups              []GameKV           `json:"groups"`
	ReleaseDate         string             `json:"release_date"`
	Developers          []string           `json:"developers"`
	Publishers          []string           `json:"publishers"`
	Links               []GameKV           `json:"links"`
	Platform            string             `json:"platform"`
	Tags                []GameTag          `json:"tags"`
	SupportedLanguages  string             `json:"supported_languages"`
	Website             string             `json:"website"`
	DetailedDescription string             `json:"detailed_description"`
	AboutTheGame        string             `json:"about_the_game"`
	PcRequirements      GamePCRequirements `json:"pc_requirements"`
}

type GameNews struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	PostTime string `json:"post_time"`
	Headline string `json:"headline"`
	Author   string `json:"author"`
	Content  string `json:"content"`
	URL      string `json:"url"`
}

type GameCreator struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Info    string   `json:"info"`
	URL     string   `json:"url"`
	Avatar  string   `json:"avatar"`
	Links   []GameKV `json:"links"`
	Contact []GameKV `json:"contact"`
	Type    int64    `json:"type"`
}

func NewHTTPGameClient(baseURL string, timeout time.Duration) *HTTPGameClient {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &HTTPGameClient{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		client:  &http.Client{Timeout: timeout},
	}
}

func (c *HTTPGameClient) ListGames(ctx context.Context, locale string) ([]GameSummary, error) {
	var data []GameSummary
	if err := c.fetchJSON(ctx, "/game/sync/list", map[string]string{"lang": locale}, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (c *HTTPGameClient) GetGameInfo(ctx context.Context, id, locale string) (GameDetail, error) {
	var data GameDetail
	err := c.fetchJSON(ctx, "/game/sync/info", map[string]string{"id": id, "lang": locale}, &data)
	return data, err
}

func (c *HTTPGameClient) ListGameNews(ctx context.Context, locale string) ([]GameNews, error) {
	var data []GameNews
	if err := c.fetchJSON(ctx, "/game/sync/news", map[string]string{"lang": locale}, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (c *HTTPGameClient) ListCreators(ctx context.Context, locale string) ([]GameCreator, error) {
	var data []GameCreator
	if err := c.fetchJSON(ctx, "/game/sync/creators", map[string]string{"lang": locale}, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (c *HTTPGameClient) fetchJSON(ctx context.Context, endpoint string, query map[string]string, target any) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("game client is not configured")
	}
	if c.baseURL == "" {
		return fmt.Errorf("rag.sync_game_base_url is not configured")
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
