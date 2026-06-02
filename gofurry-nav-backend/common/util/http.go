package util

/*
 * @Desc: http工具类
 * @author: 福狼
 * @version: v1.0.0
 */

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/bytedance/sonic"
)

const defaultHTTPHelperMaxBodyBytes int64 = 1024 * 1024

func GetByHttp(url string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "send get failed, err:%v", err)
		return "", err
	}
	defer resp.Body.Close()

	// 响应码和HTTP协议版本
	fmt.Println(resp.StatusCode, resp.Proto)

	// 读取响应内容
	body, err := io.ReadAll(io.LimitReader(resp.Body, defaultHTTPHelperMaxBodyBytes))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "get resp failed, err:%v", err)
		return "", err
	}

	return string(body), nil
}

func PostByHttp(url, contentType string, params map[string]any) (string, error) {
	// map转json
	jsonData, err := sonic.Marshal(params)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "marshal json failed, err:%v", err)
		return "", err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, contentType, strings.NewReader(string(jsonData)))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "send post failed, err:%v", err)
		return "", err
	}
	defer resp.Body.Close()

	// 响应码和HTTP协议版本
	fmt.Println(resp.StatusCode, resp.Proto)

	// 读取响应内容
	body, err := io.ReadAll(io.LimitReader(resp.Body, defaultHTTPHelperMaxBodyBytes))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "get resp failed, err:%v", err)
		return "", err
	}

	return string(body), nil
}

// Get 请求头 参数 返回: string
func GetByHttpWithParams(apiUrl string, headers map[string]string, params map[string]string, timeout time.Duration) (string, error) {
	// 构建表单
	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}

	// 拼接url和参数
	paramStr := values.Encode()
	if paramStr != "" {
		apiUrl = fmt.Sprintf("%s?%s", apiUrl, paramStr)
	}

	// 创建请求
	client := &http.Client{
		Timeout: timeout,
	}
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "create request failed, err:%v", err)
		return "", err
	}
	// 增加请求头
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	// 请求
	resp, err := client.Do(req)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "send get failed, err:%v", err)
		return "", err
	}
	defer resp.Body.Close()

	// 响应码和http协议版本
	fmt.Println(resp.StatusCode, resp.Proto)

	body, err := io.ReadAll(io.LimitReader(resp.Body, defaultHTTPHelperMaxBodyBytes))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "get resp failed, err:%v", err)
		return "", err
	}

	return string(body), nil
}

// Get 请求头 参数 返回: *goquery.Document
func GetByHttpWithParamsBackDoc(apiUrl string, headers map[string]string, params map[string]string, timeout time.Duration) (*goquery.Document, error) {
	// 构建表单
	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}

	// 拼接url和参数
	paramStr := values.Encode()
	if paramStr != "" {
		apiUrl = fmt.Sprintf("%s?%s", apiUrl, paramStr)
	}

	// 创建请求
	client := &http.Client{
		Timeout: timeout,
	}
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "create request failed, err:%v", err)
		return nil, err
	}
	// 增加请求头
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	// 请求
	resp, err := client.Do(req)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "send get failed, err:%v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// 响应码和http协议版本
	fmt.Println(resp.StatusCode, resp.Proto)

	// 解析 HTML 内容
	doc, err := goquery.NewDocumentFromReader(io.LimitReader(resp.Body, defaultHTTPHelperMaxBodyBytes))
	if err != nil {
		return nil, err
	}

	return doc, nil
}

// Post 请求头 参数 返回: string
func PostByHttpWithParams(apiUrl string, headers map[string]string, params map[string]string, timeout time.Duration) (string, error) {
	// 构建表单
	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}

	// 拼接url和参数
	paramStr := values.Encode()
	if paramStr != "" {
		apiUrl = fmt.Sprintf("%s?%s", apiUrl, paramStr)
	}

	// 创建请求
	client := &http.Client{
		Timeout: timeout,
	}
	req, err := http.NewRequest("POST", apiUrl, nil)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "create request failed, err:%v", err)
		return "", err
	}
	// 增加请求头
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	// 请求
	resp, err := client.Do(req)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "send get failed, err:%v", err)
		return "", err
	}
	defer resp.Body.Close()

	// 响应码和http协议版本
	fmt.Println(resp.StatusCode, resp.Proto)

	body, err := io.ReadAll(io.LimitReader(resp.Body, defaultHTTPHelperMaxBodyBytes))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "get resp failed, err:%v", err)
		return "", err
	}

	return string(body), nil
}
