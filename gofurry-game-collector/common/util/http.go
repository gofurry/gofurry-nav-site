package util

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/bytedance/sonic"
)

/*
 * @Desc: http工具类
 * @author: 福狼
 * @version: v1.0.0
 */

// 全局HTTP客户端
var (
	defaultTransport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 若需验证证书设为false
		},
		MaxIdleConns:        100,              // 最大空闲连接数
		MaxConnsPerHost:     20,               // 每个主机的最大连接数
		IdleConnTimeout:     30 * time.Second, // 空闲连接超时时间
		TLSHandshakeTimeout: 5 * time.Second,  // TLS握手超时
	}

	defaultClient = &http.Client{
		Transport: defaultTransport,
		Timeout:   30 * time.Second, // 默认超时时间
	}
)

// GetByHttp 基础GET请求
func GetByHttp(url string) (string, error) {
	resp, err := defaultClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("发送GET请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应体失败: %w", err)
	}

	return string(body), nil
}

// PostByHttp 基础POST请求
func PostByHttp(url, contentType string, params map[string]any) (string, error) {
	jsonData, err := sonic.Marshal(params)
	if err != nil {
		return "", fmt.Errorf("参数JSON序列化失败: %w", err)
	}

	resp, err := defaultClient.Post(url, contentType, strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("发送POST请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应体失败: %w", err)
	}

	return string(body), nil
}

// GetByHttpWithParams 带请求头、参数和超时的GET请求
func GetByHttpWithParams(apiUrl string, headers map[string]string, params map[string]string, timeout time.Duration, proxy *string) (string, error) {
	// 构建查询参数
	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}
	paramStr := values.Encode()
	if paramStr != "" {
		apiUrl = fmt.Sprintf("%s?%s", apiUrl, paramStr)
	}

	// 创建请求
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	// 获取客户端
	client := getClientWithProxy(proxy, timeout)

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送GET请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应体失败: %w", err)
	}

	return string(body), nil
}

// GetByHttpWithParamsBackDoc 带参数的GET请求，返回goquery.Document
func GetByHttpWithParamsBackDoc(apiUrl string, headers map[string]string, params map[string]string, timeout time.Duration, proxy *string) (*goquery.Document, error) {
	// 构建查询参数
	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}
	paramStr := values.Encode()
	if paramStr != "" {
		apiUrl = fmt.Sprintf("%s?%s", apiUrl, paramStr)
	}

	// 创建请求
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	// 获取客户端（根据代理动态选择Transport）
	client := getClientWithProxy(proxy, timeout)

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送GET请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 解析为goquery.Document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("解析响应为HTML失败: %w", err)
	}

	return doc, nil
}

// PostByHttpWithParams 带参数的POST请求
func PostByHttpWithParams(apiUrl string, headers map[string]string, params map[string]string, timeout time.Duration, proxy *string) (string, error) {
	// 构建表单参数
	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}
	paramStr := values.Encode()

	// 创建请求（参数放在Body中，而非URL）
	req, err := http.NewRequest("POST", apiUrl, strings.NewReader(paramStr))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置默认Content-Type（表单提交）
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// 覆盖自定义请求头
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// 获取客户端
	client := getClientWithProxy(proxy, timeout)

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送POST请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应体失败: %w", err)
	}

	return string(body), nil
}

// 工具函数：根据代理和超时时间获取客户端
func getClientWithProxy(proxy *string, timeout time.Duration) *http.Client {
	// 无需代理使用默认客户端
	if proxy == nil || *proxy == "" {
		return &http.Client{
			Transport: defaultTransport,
			Timeout:   timeout,
		}
	}

	// 有代理时创建带代理的Transport
	proxyURL, err := url.Parse(*proxy)
	if err != nil {
		// 代理解析失败 降级使用默认客户端
		return &http.Client{
			Transport: defaultTransport,
			Timeout:   timeout,
		}
	}

	// 仅修改Proxy
	proxyTransport := defaultTransport.Clone()
	proxyTransport.Proxy = http.ProxyURL(proxyURL)

	return &http.Client{
		Transport: proxyTransport,
		Timeout:   timeout,
	}
}
