package httpkit

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/net/proxy"
)

const (
	DefaultTimeout             = 30 * time.Second
	DefaultMaxIdleConns        = 100
	DefaultMaxIdleConnsPerHost = 10
	DefaultIdleConnTimeout     = 90 * time.Second
	maxCachedClients           = 32
)

var (
	defaultClientMu sync.RWMutex
	DefaultClient   = NewHTTPClient(nil)
)

type HTTPClient struct {
	client                   *http.Client
	baseTransport            *http.Transport
	defaultProxy             *ProxyConfig
	defaultTimeout           time.Duration
	defaultTLSConfig         *tls.Config
	defaultJar               http.CookieJar
	defaultDisableKeepAlives bool
	defaultFollowRedirects   *bool
	defaultMaxRedirects      int
	clients                  map[string]*http.Client
	mu                       sync.RWMutex
}

type ProxyConfig struct {
	Type     string `json:"type"`
	Network  string `json:"network"`
	Address  string `json:"address"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type RequestOptions struct {
	Headers           map[string]string `json:"headers"`
	QueryParams       url.Values        `json:"-"`
	Timeout           time.Duration     `json:"timeout"`
	Proxy             *ProxyConfig      `json:"proxy"`
	Context           context.Context   `json:"-"`
	TLSConfig         *tls.Config       `json:"-"`
	Jar               http.CookieJar    `json:"-"`
	DisableKeepAlives *bool             `json:"disable_keep_alives"`
	FollowRedirects   *bool             `json:"follow_redirects"`
	MaxRedirects      int               `json:"max_redirects"`
	RetryCount        int               `json:"retry_count"`
	RetryDelay        time.Duration     `json:"retry_delay"`
	MaxResponseSize   int64             `json:"max_response_size"`
}

type Response struct {
	StatusCode int           `json:"status_code"`
	Headers    http.Header   `json:"headers"`
	Body       []byte        `json:"body"`
	Duration   time.Duration `json:"duration"`
}

type StatusError struct {
	Method     string
	URL        string
	StatusCode int
	Body       []byte
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("http %s %s returned status %d", strings.ToUpper(e.Method), e.URL, e.StatusCode)
}

func (e *StatusError) BodyString() string {
	return string(e.Body)
}

func NewHTTPClient(options *RequestOptions) *HTTPClient {
	defaultTimeout := DefaultTimeout
	if options != nil && options.Timeout > 0 {
		defaultTimeout = options.Timeout
	}

	defaultTLSConfig := (*tls.Config)(nil)
	if options != nil && options.TLSConfig != nil {
		defaultTLSConfig = options.TLSConfig
	}

	defaultJar := http.CookieJar(nil)
	if options != nil && options.Jar != nil {
		defaultJar = options.Jar
	}

	defaultDisableKeepAlives := false
	if options != nil && options.DisableKeepAlives != nil {
		defaultDisableKeepAlives = *options.DisableKeepAlives
	}

	var defaultFollowRedirects *bool
	if options != nil && options.FollowRedirects != nil {
		value := *options.FollowRedirects
		defaultFollowRedirects = &value
	}

	defaultMaxRedirects := 0
	if options != nil && options.MaxRedirects > 0 {
		defaultMaxRedirects = options.MaxRedirects
	}

	baseTransport := &http.Transport{
		MaxIdleConns:        DefaultMaxIdleConns,
		MaxIdleConnsPerHost: DefaultMaxIdleConnsPerHost,
		IdleConnTimeout:     DefaultIdleConnTimeout,
		DisableCompression:  false,
		DisableKeepAlives:   defaultDisableKeepAlives,
		TLSClientConfig:     defaultTLSConfig,
	}

	var defaultProxy *ProxyConfig
	clientTransport := baseTransport
	if options != nil && options.Proxy != nil {
		defaultProxy = cloneProxyConfig(options.Proxy)
		proxyTransport := baseTransport.Clone()
		if err := configureProxy(proxyTransport, defaultProxy); err == nil {
			clientTransport = proxyTransport
		} else {
			defaultProxy = nil
		}
	}

	client := &http.Client{
		Transport:     clientTransport,
		Jar:           defaultJar,
		CheckRedirect: buildRedirectPolicy(defaultFollowRedirects, defaultMaxRedirects),
	}

	defaultKey := clientKey(defaultProxy, defaultTLSConfig, defaultDisableKeepAlives, defaultJar, defaultFollowRedirects, defaultMaxRedirects)
	clients := map[string]*http.Client{defaultKey: client}

	return &HTTPClient{
		client:                   client,
		baseTransport:            baseTransport,
		defaultProxy:             defaultProxy,
		defaultTimeout:           defaultTimeout,
		defaultTLSConfig:         defaultTLSConfig,
		defaultJar:               defaultJar,
		defaultDisableKeepAlives: defaultDisableKeepAlives,
		defaultFollowRedirects:   defaultFollowRedirects,
		defaultMaxRedirects:      defaultMaxRedirects,
		clients:                  clients,
	}
}

func NewStdClient(options *RequestOptions) *http.Client {
	client := NewHTTPClient(options)
	if options == nil {
		return client.getClient()
	}
	return client.getClientForOptions(options)
}

func ConfigureDefault(options *RequestOptions) {
	setDefaultClient(NewHTTPClient(options))
}

func ResetDefault() {
	setDefaultClient(NewHTTPClient(nil))
}

func SetDefaultTimeout(timeout time.Duration) {
	if timeout <= 0 {
		return
	}
	options := getDefaultClient().defaultOptions()
	options.Timeout = timeout
	setDefaultClient(NewHTTPClient(options))
}

func SetDefaultProxy(proxyConfig *ProxyConfig) {
	options := getDefaultClient().defaultOptions()
	options.Proxy = cloneProxyConfig(proxyConfig)
	setDefaultClient(NewHTTPClient(options))
}

func (c *HTTPClient) Get(rawURL string, options ...*RequestOptions) (*Response, error) {
	return c.Request(http.MethodGet, rawURL, nil, options...)
}

func (c *HTTPClient) Post(rawURL string, body any, options ...*RequestOptions) (*Response, error) {
	return c.Request(http.MethodPost, rawURL, body, options...)
}

func (c *HTTPClient) Put(rawURL string, body any, options ...*RequestOptions) (*Response, error) {
	return c.Request(http.MethodPut, rawURL, body, options...)
}

func (c *HTTPClient) Delete(rawURL string, options ...*RequestOptions) (*Response, error) {
	return c.Request(http.MethodDelete, rawURL, nil, options...)
}

func Get(rawURL string, options ...*RequestOptions) (*Response, error) {
	return getDefaultClient().Get(rawURL, options...)
}

func Post(rawURL string, body any, options ...*RequestOptions) (*Response, error) {
	return getDefaultClient().Post(rawURL, body, options...)
}

func Put(rawURL string, body any, options ...*RequestOptions) (*Response, error) {
	return getDefaultClient().Put(rawURL, body, options...)
}

func Delete(rawURL string, options ...*RequestOptions) (*Response, error) {
	return getDefaultClient().Delete(rawURL, options...)
}

func (c *HTTPClient) Request(method, rawURL string, body any, options ...*RequestOptions) (*Response, error) {
	opt := firstOptions(options...)
	ctx := opt.Context
	if ctx == nil {
		ctx = context.Background()
	}

	timeout := c.getDefaultTimeout()
	if opt.Timeout > 0 {
		timeout = opt.Timeout
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	retryCount := opt.RetryCount
	if retryCount <= 0 {
		retryCount = 1
	}
	if !isIdempotentMethod(method) && retryCount > 1 {
		retryCount = 1
	}

	bodyFn, err := buildBodyProvider(body, retryCount)
	if err != nil {
		return nil, fmt.Errorf("prepare request body failed: %w", err)
	}

	targetURL, err := mergeQueryParams(rawURL, opt.QueryParams)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for attempt := 0; attempt < retryCount; attempt++ {
		if attempt > 0 && opt.RetryDelay > 0 {
			select {
			case <-time.After(opt.RetryDelay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		resp, reqErr := c.doSingleRequest(ctx, method, targetURL, bodyFn, opt)
		if reqErr == nil {
			return resp, nil
		}
		lastErr = reqErr

		if !shouldRetry(reqErr) {
			break
		}
	}

	return nil, lastErr
}

type bodyProvider func() (io.Reader, string, error)

func buildBodyProvider(body any, retryCount int) (bodyProvider, error) {
	if body == nil {
		return nil, nil
	}

	switch value := body.(type) {
	case string:
		data := value
		return func() (io.Reader, string, error) {
			return strings.NewReader(data), "text/plain; charset=utf-8", nil
		}, nil
	case []byte:
		data := append([]byte(nil), value...)
		return func() (io.Reader, string, error) {
			return bytes.NewReader(data), "application/octet-stream", nil
		}, nil
	case io.ReadSeeker:
		return func() (io.Reader, string, error) {
			if _, err := value.Seek(0, io.SeekStart); err != nil {
				return nil, "", err
			}
			return value, "", nil
		}, nil
	case io.Reader:
		if retryCount > 1 {
			data, err := io.ReadAll(value)
			if err != nil {
				return nil, err
			}
			return func() (io.Reader, string, error) {
				return bytes.NewReader(data), "", nil
			}, nil
		}
		return func() (io.Reader, string, error) {
			return value, "", nil
		}, nil
	case url.Values:
		data := value.Encode()
		return func() (io.Reader, string, error) {
			return strings.NewReader(data), "application/x-www-form-urlencoded", nil
		}, nil
	default:
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		return func() (io.Reader, string, error) {
			return bytes.NewReader(data), "application/json", nil
		}, nil
	}
}

func (c *HTTPClient) doSingleRequest(ctx context.Context, method, rawURL string, bodyFn bodyProvider, opt *RequestOptions) (*Response, error) {
	start := time.Now()

	var (
		bodyReader  io.Reader
		contentType string
		err         error
	)
	if bodyFn != nil {
		bodyReader, contentType, err = bodyFn()
		if err != nil {
			return nil, fmt.Errorf("prepare request body failed: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, rawURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	for key, value := range opt.Headers {
		req.Header.Set(key, value)
	}
	if bodyFn != nil && req.Header.Get("Content-Type") == "" && contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	client := c.getClientForOptions(opt)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := readResponseBody(resp.Body, opt.MaxResponseSize)
	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header.Clone(),
		Body:       bodyBytes,
		Duration:   time.Since(start),
	}, nil
}

func readResponseBody(body io.Reader, maxResponseSize int64) ([]byte, error) {
	if maxResponseSize <= 0 {
		bodyBytes, err := io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("read response body failed: %w", err)
		}
		return bodyBytes, nil
	}

	limit := maxResponseSize + 1
	if limit <= 0 {
		limit = maxResponseSize
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(body, limit))
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}
	if int64(len(bodyBytes)) > maxResponseSize {
		return nil, fmt.Errorf("response body exceeds max size %d bytes", maxResponseSize)
	}
	return bodyBytes, nil
}

func (c *HTTPClient) getClient() *http.Client {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.client
}

func (c *HTTPClient) getDefaultTimeout() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.defaultTimeout
}

func (c *HTTPClient) defaultOptions() *RequestOptions {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return &RequestOptions{
		Timeout:           c.defaultTimeout,
		Proxy:             cloneProxyConfig(c.defaultProxy),
		TLSConfig:         c.defaultTLSConfig,
		Jar:               c.defaultJar,
		DisableKeepAlives: boolPtr(c.defaultDisableKeepAlives),
		FollowRedirects:   cloneBoolPtr(c.defaultFollowRedirects),
		MaxRedirects:      c.defaultMaxRedirects,
	}
}

func (c *HTTPClient) getClientForOptions(opt *RequestOptions) *http.Client {
	c.mu.RLock()
	defaultProxy := c.defaultProxy
	defaultTLSConfig := c.defaultTLSConfig
	defaultJar := c.defaultJar
	defaultDisableKeepAlives := c.defaultDisableKeepAlives
	defaultFollowRedirects := c.defaultFollowRedirects
	defaultMaxRedirects := c.defaultMaxRedirects
	baseTransport := c.baseTransport
	c.mu.RUnlock()

	effectiveProxy := defaultProxy
	if opt != nil && opt.Proxy != nil {
		effectiveProxy = opt.Proxy
	}

	effectiveTLSConfig := defaultTLSConfig
	if opt != nil && opt.TLSConfig != nil {
		effectiveTLSConfig = opt.TLSConfig
	}

	effectiveJar := defaultJar
	if opt != nil && opt.Jar != nil {
		effectiveJar = opt.Jar
	}

	effectiveDisableKeepAlives := defaultDisableKeepAlives
	if opt != nil && opt.DisableKeepAlives != nil {
		effectiveDisableKeepAlives = *opt.DisableKeepAlives
	}

	effectiveFollowRedirects := defaultFollowRedirects
	if opt != nil && opt.FollowRedirects != nil {
		effectiveFollowRedirects = opt.FollowRedirects
	}

	effectiveMaxRedirects := defaultMaxRedirects
	if opt != nil && opt.MaxRedirects > 0 {
		effectiveMaxRedirects = opt.MaxRedirects
	}

	key := clientKey(effectiveProxy, effectiveTLSConfig, effectiveDisableKeepAlives, effectiveJar, effectiveFollowRedirects, effectiveMaxRedirects)
	c.mu.RLock()
	if client, ok := c.clients[key]; ok {
		c.mu.RUnlock()
		return client
	}
	c.mu.RUnlock()

	if baseTransport == nil {
		return c.getClient()
	}

	transport := baseTransport.Clone()
	transport.DisableKeepAlives = effectiveDisableKeepAlives
	transport.TLSClientConfig = effectiveTLSConfig
	if effectiveProxy != nil {
		if err := configureProxy(transport, effectiveProxy); err != nil {
			return c.getClient()
		}
	}

	client := &http.Client{
		Transport:     transport,
		Jar:           effectiveJar,
		CheckRedirect: buildRedirectPolicy(effectiveFollowRedirects, effectiveMaxRedirects),
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if existing, ok := c.clients[key]; ok {
		return existing
	}
	if len(c.clients) >= maxCachedClients {
		return client
	}
	c.clients[key] = client
	return client
}

func (c *HTTPClient) CloseIdleConnections() {
	c.mu.RLock()
	clients := make([]*http.Client, 0, len(c.clients))
	for _, client := range c.clients {
		clients = append(clients, client)
	}
	c.mu.RUnlock()

	seen := make(map[*http.Client]struct{}, len(clients))
	for _, client := range clients {
		if client == nil {
			continue
		}
		if _, ok := seen[client]; ok {
			continue
		}
		seen[client] = struct{}{}
		client.CloseIdleConnections()
	}
}

func getDefaultClient() *HTTPClient {
	defaultClientMu.RLock()
	defer defaultClientMu.RUnlock()
	return DefaultClient
}

func setDefaultClient(client *HTTPClient) {
	if client == nil {
		client = NewHTTPClient(nil)
	}

	defaultClientMu.Lock()
	previous := DefaultClient
	DefaultClient = client
	defaultClientMu.Unlock()

	if previous != nil {
		previous.CloseIdleConnections()
	}
}

func buildRedirectPolicy(followRedirects *bool, maxRedirects int) func(req *http.Request, via []*http.Request) error {
	if followRedirects != nil && !*followRedirects {
		return func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	if maxRedirects > 0 {
		return func(req *http.Request, via []*http.Request) error {
			if len(via) > maxRedirects {
				return fmt.Errorf("stopped after %d redirects", maxRedirects)
			}
			return nil
		}
	}
	return nil
}

func configureProxy(transport *http.Transport, proxyConfig *ProxyConfig) error {
	switch strings.ToLower(proxyConfig.Type) {
	case "http", "https":
		scheme := strings.ToLower(proxyConfig.Type)
		proxyURL := fmt.Sprintf("%s://%s", scheme, proxyConfig.Address)
		if proxyConfig.Username != "" && proxyConfig.Password != "" {
			proxyURL = fmt.Sprintf("%s://%s:%s@%s",
				scheme,
				url.QueryEscape(proxyConfig.Username),
				url.QueryEscape(proxyConfig.Password),
				proxyConfig.Address,
			)
		}

		parsedURL, err := url.Parse(proxyURL)
		if err != nil {
			return fmt.Errorf("invalid proxy url: %w", err)
		}
		transport.Proxy = http.ProxyURL(parsedURL)
	case "socks5":
		var auth *proxy.Auth
		if proxyConfig.Username != "" && proxyConfig.Password != "" {
			auth = &proxy.Auth{
				User:     proxyConfig.Username,
				Password: proxyConfig.Password,
			}
		}

		network := "tcp"
		if proxyConfig.Network != "" {
			network = proxyConfig.Network
		}

		dialer, err := proxy.SOCKS5(network, proxyConfig.Address, auth, proxy.Direct)
		if err != nil {
			return fmt.Errorf("create socks5 dialer failed: %w", err)
		}

		contextDialer, ok := dialer.(proxy.ContextDialer)
		if !ok {
			return fmt.Errorf("socks5 dialer does not support context")
		}
		transport.DialContext = contextDialer.DialContext
	default:
		return fmt.Errorf("unsupported proxy type: %s", proxyConfig.Type)
	}

	return nil
}

func mergeQueryParams(rawURL string, params url.Values) (string, error) {
	if len(params) == 0 {
		return rawURL, nil
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse request url failed: %w", err)
	}

	query := parsedURL.Query()
	for key, values := range params {
		query.Del(key)
		for _, value := range values {
			query.Add(key, value)
		}
	}
	parsedURL.RawQuery = query.Encode()
	return parsedURL.String(), nil
}

func firstOptions(options ...*RequestOptions) *RequestOptions {
	if len(options) == 0 || options[0] == nil {
		return &RequestOptions{}
	}
	return cloneOptions(options[0])
}

func cloneOptions(opt *RequestOptions) *RequestOptions {
	if opt == nil {
		return &RequestOptions{}
	}

	cloned := *opt
	if opt.Headers != nil {
		cloned.Headers = make(map[string]string, len(opt.Headers))
		for key, value := range opt.Headers {
			cloned.Headers[key] = value
		}
	}
	if opt.QueryParams != nil {
		cloned.QueryParams = make(url.Values, len(opt.QueryParams))
		for key, values := range opt.QueryParams {
			cloned.QueryParams[key] = append([]string(nil), values...)
		}
	}
	if opt.Proxy != nil {
		cloned.Proxy = cloneProxyConfig(opt.Proxy)
	}
	if opt.DisableKeepAlives != nil {
		value := *opt.DisableKeepAlives
		cloned.DisableKeepAlives = &value
	}
	if opt.FollowRedirects != nil {
		value := *opt.FollowRedirects
		cloned.FollowRedirects = &value
	}
	return &cloned
}

func cloneProxyConfig(cfg *ProxyConfig) *ProxyConfig {
	if cfg == nil {
		return nil
	}
	cloned := *cfg
	return &cloned
}

func cloneBoolPtr(value *bool) *bool {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func boolPtr(value bool) *bool {
	return &value
}

func proxyKey(proxyConfig *ProxyConfig) string {
	if proxyConfig == nil {
		return ""
	}
	return strings.Join([]string{
		strings.ToLower(proxyConfig.Type),
		strings.ToLower(proxyConfig.Network),
		proxyConfig.Address,
		proxyConfig.Username,
		proxyConfig.Password,
	}, "|")
}

func boolPtrKey(value *bool) string {
	if value == nil {
		return "nil"
	}
	if *value {
		return "true"
	}
	return "false"
}

func clientKey(proxyConfig *ProxyConfig, tlsConfig *tls.Config, disableKeepAlives bool, jar http.CookieJar, followRedirects *bool, maxRedirects int) string {
	return strings.Join([]string{
		proxyKey(proxyConfig),
		fmt.Sprintf("%p", tlsConfig),
		fmt.Sprintf("%p", jar),
		strconv.FormatBool(disableKeepAlives),
		boolPtrKey(followRedirects),
		strconv.Itoa(maxRedirects),
	}, "|")
}

func isIdempotentMethod(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

func shouldRetry(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() || netErr.Temporary() {
			return true
		}
	}

	if errors.Is(err, io.EOF) {
		return true
	}

	if errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.EPIPE) {
		return true
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "connection refused") ||
		strings.Contains(message, "timeout") ||
		strings.Contains(message, "temporary failure")
}

func (r *Response) String() string {
	return string(r.Body)
}

func (r *Response) JSON(target any) error {
	return json.Unmarshal(r.Body, target)
}

func (r *Response) IsSuccess() bool {
	return r.StatusCode >= http.StatusOK && r.StatusCode < http.StatusMultipleChoices
}
