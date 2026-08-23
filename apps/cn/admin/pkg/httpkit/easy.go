package httpkit

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

func NewJSONOptions(headers map[string]string) *RequestOptions {
	opt := &RequestOptions{}
	if len(headers) > 0 {
		opt.Headers = make(map[string]string, len(headers)+1)
		for key, value := range headers {
			opt.Headers[key] = value
		}
	} else {
		opt.Headers = make(map[string]string, 1)
	}
	opt.Headers["Content-Type"] = "application/json"
	opt.Headers["Accept"] = "application/json"
	return opt
}

func (c *HTTPClient) GetJSON(rawURL string, target any, options ...*RequestOptions) (*Response, error) {
	resp, err := c.Get(rawURL, options...)
	if err != nil {
		return nil, err
	}
	if err := decodeJSONResponse(http.MethodGet, rawURL, resp, target); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *HTTPClient) PostJSON(rawURL string, body any, target any, options ...*RequestOptions) (*Response, error) {
	opt := mergeOptions(NewJSONOptions(nil), firstOptions(options...))
	resp, err := c.Post(rawURL, body, opt)
	if err != nil {
		return nil, err
	}
	if err := decodeJSONResponse(http.MethodPost, rawURL, resp, target); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *HTTPClient) PutJSON(rawURL string, body any, target any, options ...*RequestOptions) (*Response, error) {
	opt := mergeOptions(NewJSONOptions(nil), firstOptions(options...))
	resp, err := c.Put(rawURL, body, opt)
	if err != nil {
		return nil, err
	}
	if err := decodeJSONResponse(http.MethodPut, rawURL, resp, target); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *HTTPClient) PostForm(rawURL string, data url.Values, options ...*RequestOptions) (*Response, error) {
	opt := mergeOptions(&RequestOptions{
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
	}, firstOptions(options...))
	return c.Post(rawURL, data, opt)
}

func (c *HTTPClient) GetText(rawURL string, options ...*RequestOptions) (string, *Response, error) {
	resp, err := c.Get(rawURL, options...)
	if err != nil {
		return "", nil, err
	}
	if !resp.IsSuccess() {
		return "", resp, newStatusError(http.MethodGet, rawURL, resp)
	}
	return resp.String(), resp, nil
}

func (c *HTTPClient) PostText(rawURL string, body any, options ...*RequestOptions) (string, *Response, error) {
	resp, err := c.Post(rawURL, body, options...)
	if err != nil {
		return "", nil, err
	}
	if !resp.IsSuccess() {
		return "", resp, newStatusError(http.MethodPost, rawURL, resp)
	}
	return resp.String(), resp, nil
}

func GetJSON(rawURL string, target any, options ...*RequestOptions) (*Response, error) {
	return getDefaultClient().GetJSON(rawURL, target, options...)
}

func PostJSON(rawURL string, body any, target any, options ...*RequestOptions) (*Response, error) {
	return getDefaultClient().PostJSON(rawURL, body, target, options...)
}

func PutJSON(rawURL string, body any, target any, options ...*RequestOptions) (*Response, error) {
	return getDefaultClient().PutJSON(rawURL, body, target, options...)
}

func PostForm(rawURL string, data url.Values, options ...*RequestOptions) (*Response, error) {
	return getDefaultClient().PostForm(rawURL, data, options...)
}

func GetText(rawURL string, options ...*RequestOptions) (string, *Response, error) {
	return getDefaultClient().GetText(rawURL, options...)
}

func PostText(rawURL string, body any, options ...*RequestOptions) (string, *Response, error) {
	return getDefaultClient().PostText(rawURL, body, options...)
}

func decodeJSONResponse(method, rawURL string, resp *Response, target any) error {
	if !resp.IsSuccess() {
		return newStatusError(method, rawURL, resp)
	}
	if target == nil || len(resp.Body) == 0 {
		return nil
	}
	if err := resp.JSON(target); err != nil {
		return fmt.Errorf("decode response json failed: %w", err)
	}
	return nil
}

func newStatusError(method, rawURL string, resp *Response) error {
	if resp == nil {
		return errors.New("response is nil")
	}
	return &StatusError{
		Method:     method,
		URL:        rawURL,
		StatusCode: resp.StatusCode,
		Body:       append([]byte(nil), resp.Body...),
	}
}

func mergeOptions(base, incoming *RequestOptions) *RequestOptions {
	merged := cloneOptions(base)
	if incoming == nil {
		return merged
	}
	if merged == nil {
		merged = &RequestOptions{}
	}

	if len(incoming.Headers) > 0 {
		if merged.Headers == nil {
			merged.Headers = make(map[string]string, len(incoming.Headers))
		}
		for key, value := range incoming.Headers {
			merged.Headers[key] = value
		}
	}
	if len(incoming.QueryParams) > 0 {
		if merged.QueryParams == nil {
			merged.QueryParams = url.Values{}
		}
		for key, values := range incoming.QueryParams {
			for _, value := range values {
				merged.QueryParams.Add(key, value)
			}
		}
	}
	if incoming.Timeout > 0 {
		merged.Timeout = incoming.Timeout
	}
	if incoming.Proxy != nil {
		merged.Proxy = cloneProxyConfig(incoming.Proxy)
	}
	if incoming.Context != nil {
		merged.Context = incoming.Context
	}
	if incoming.TLSConfig != nil {
		merged.TLSConfig = incoming.TLSConfig
	}
	if incoming.Jar != nil {
		merged.Jar = incoming.Jar
	}
	if incoming.DisableKeepAlives != nil {
		merged.DisableKeepAlives = cloneBoolPtr(incoming.DisableKeepAlives)
	}
	if incoming.FollowRedirects != nil {
		merged.FollowRedirects = cloneBoolPtr(incoming.FollowRedirects)
	}
	if incoming.MaxRedirects > 0 {
		merged.MaxRedirects = incoming.MaxRedirects
	}
	if incoming.RetryCount > 0 {
		merged.RetryCount = incoming.RetryCount
	}
	if incoming.RetryDelay > 0 {
		merged.RetryDelay = incoming.RetryDelay
	}
	if incoming.MaxResponseSize > 0 {
		merged.MaxResponseSize = incoming.MaxResponseSize
	}
	return merged
}
