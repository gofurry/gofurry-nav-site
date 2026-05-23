package service

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofurry/gofurry-nav-collector/collector/http/dao"
	"github.com/gofurry/gofurry-nav-collector/collector/http/models"
	"github.com/gofurry/gofurry-nav-collector/common/log"
	cm "github.com/gofurry/gofurry-nav-collector/common/models"
	cs "github.com/gofurry/gofurry-nav-collector/common/service"
	"github.com/gofurry/gofurry-nav-collector/common/util"
	"github.com/gofurry/gofurry-nav-collector/roof/env"
	"github.com/sourcegraph/conc/pool"
)

var requestRunning atomic.Bool

// ============== HTTP模块 - 初始化部分 ==============

// 初始化
func InitHTTPOnStart() {
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("receive InitHttpOnStart recover: %v", err))
		}
	}()
	fmt.Println("Request 模块初始化开始...")

	//初始化后执行一次 Request
	go Request()
	go Delete()
	// 定时任务执行 Request
	cs.AddCronJob(time.Duration(env.GetServerConfig().Collector.Request.RequestInterval)*time.Hour, Request)
	cs.AddCronJob(48*time.Hour, Delete)

	fmt.Println("Request 模块初始化结束...")
}

// 每天清理一次日志表
func Delete() {
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("receive Ping Delete recover: %v", err))
		}
	}()

	// 每个域名仅保留 1500 条 request 记录
	count, deleteErr := dao.GetHTTPDao().DeleteByNum(env.GetServerConfig().Collector.Request.LogCount)
	if deleteErr != nil {
		log.Error("删除多余Request记录失败: ", deleteErr)
	} else {
		log.Info("删除多余Request记录成功, 共删除: ", count)
	}
}

// ============== HTTP模块 - 执行部分 ==============

// 执行 Request
func Request() {
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Sprintf("receive Request recover: %v", err))
		}
	}()
	if !requestRunning.CompareAndSwap(false, true) {
		log.WithFieldsMsg(map[string]interface{}{
			"protocol": "http",
			"status":   "skipped",
			"reason":   "previous_run_running",
		}, "Request skipped: previous run is still running")
		return
	}
	defer requestRunning.Store(false)

	requestList, err := dao.GetHTTPDao().GetList()
	if err != nil {
		log.Error("Request 获取站点列表失败: " + err.GetMsg())
		return
	}
	// 判空
	if cap(requestList) < 1 || len(requestList) < 1 {
		log.Info("Request 站点列表为空")
		return
	}
	log.Info("HTTP 采集开始")
	requestThread := pool.New().WithMaxGoroutines(env.GetServerConfig().Collector.Request.RequestThread)
	// 遍历站点列表, 每个站点开一个线程执行 request
	for _, v := range requestList {
		requestThread.Go(getRequestResult(v))
	}
	// 等待所有采集和解析执行完毕
	requestThread.Wait()
	log.Info("HTTP 采集结束")
}

// ============== HTTP模块 - 存储部分 ==============

// 解析 Request 采集结果
func getRequestResult(site models.GfnCollectorDomain) func() {
	return func() {
		defer func() {
			if err := recover(); err != nil {
				log.Error(fmt.Sprintf("receive RequestThread recover: %v", err))
			}
		}()

		// 执行 Request 获取结果
		result := performRequest(site)
		httpRecord := models.HTTPSaveModel{
			Domain:        result.Domain,
			Url:           result.Url,
			StatusCode:    result.StatusCode,
			ResponseTime:  util.Int642String(result.ResponseTime) + "ms",
			ContentLength: result.ContentLength,
			Title:         result.Title,
			Server:        result.Server,
			Redirects:     result.Redirects,
			Headers:       result.Headers,
			Meta:          result.Meta,
			TLSVersion:    result.TLSVersion,
			CipherSuite:   result.CipherSuite,
			CertExpiry:    result.CertExpiry.String(),
			CertDaysLeft:  util.Int642String(result.CertDaysLeft) + "天",
			CertIssuer:    result.CertIssuer,
			CertIssuerOrg: result.CertIssuerOrg,
			CertDNSNames:  result.CertDNSNames,
			CertPubKeyAlg: result.CertPubKeyAlg,
			CertSigAlg:    result.CertSigAlg,
			CertEmail:     result.CertEmail,
			CertIsCA:      result.CertIsCA,
		}
		jsonResult, _ := sonic.Marshal(httpRecord)

		var siteName string
		if site.Prefix != nil {
			siteName = *site.Prefix + site.Name
		} else {
			siteName = site.Name
		}

		httpSaveRecord := models.GfnCollectorLogHTTP{
			ID:         util.GenerateId(),
			Name:       siteName,
			Info:       string(jsonResult),
			CreateTime: result.StartTime,
		}

		if httpRecord.StatusCode == 0 || jsonResult == nil {
			httpSaveRecord.Status = "failure"
		} else {
			httpSaveRecord.Status = "success"
		}

		// 记录存redis
		cs.SetNX("request:"+siteName, string(jsonResult), 48*time.Hour)     // 创建记录
		cs.SetExpire("request:"+siteName, string(jsonResult), 48*time.Hour) // 更新记录

		// 存数据库
		err := dao.GetHTTPDao().Add(&httpSaveRecord)
		if err != nil {
			log.Error("添加http请求结果到数据库失败: ", err.GetMsg())
		}
	}
}

// ============== HTTP模块 - 采集和解析部分 ==============

// 执行 Request 采集
func performRequest(site models.GfnCollectorDomain) (res models.HTTPModel) {
	res.Domain = site.Name
	if site.Prefix != nil {
		res.Url = *site.Prefix + site.Name
	} else {
		res.Url = site.Name
	}
	if site.TLS == "1" {
		res.Url = "https://" + res.Url
	} else {
		res.Url = "http://" + res.Url
	}
	res.Meta = make(map[string]string)
	res.Headers = make(map[string][]string)
	res.Redirects = []string{}
	probeBudget := env.GetServerConfig().Collector.ProbeBudget

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		TLSHandshakeTimeout: probeBudget.TLSHandshakeTimeout(),
	}
	// 设置代理
	if site.Proxy == "1" {
		proxyURL, _ := url.Parse(env.GetServerConfig().Collector.Proxy)
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	redirects := []string{}
	maxRedirects := probeBudget.MaxHTTPRedirects()
	client := &http.Client{
		Transport: transport,
		Timeout:   probeBudget.HTTPTimeout(),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			redirects = append(redirects, req.URL.String())
			if len(via) >= maxRedirects {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	// 构建请求
	req, err := http.NewRequest("GET", res.Url, nil)
	if err != nil {
		log.Error("创建请求失败: ", err)
		return
	}
	// 设置请求头
	for k, v := range models.HeadersMap {
		req.Header.Set(k, v)
	}

	// 请求开始
	start := time.Now()
	res.StartTime = cm.LocalTime(time.Now())
	resp, err := client.Do(req)
	if err != nil {
		log.Error("请求失败: ", err)
		return
	}
	defer resp.Body.Close()

	// 响应时间
	res.ResponseTime = time.Since(start).Milliseconds()
	res.StatusCode = int64(resp.StatusCode)
	res.Redirects = redirects

	for _, v := range models.CommonHeaders {
		if val := resp.Header.Values(v); len(val) > 0 {
			res.Headers[v] = val
		}
	}

	res.Server = resp.Header.Get("Server")

	// 读取响应体 限制 1MB
	body, err := io.ReadAll(io.LimitReader(resp.Body, probeBudget.MaxHTTPResponseBytes()))
	if err == nil {
		res.ContentLength = int64(len(body))

		// 提取 <title>
		re := regexp.MustCompile(`(?i)<title>(.*?)</title>`)
		matches := re.FindSubmatch(body)
		if len(matches) > 1 {
			res.Title = string(matches[1])
		}

		// 提取 <meta charset="...">
		reCharset := regexp.MustCompile(`(?i)<meta\s+[^>]*charset=["']?([^"'>\s]+)`)
		if m := reCharset.FindSubmatch(body); len(m) > 1 {
			res.Meta["charset"] = string(m[1])
		}

		// 提取 <meta name="description" content="...">
		reDesc := regexp.MustCompile(`(?i)<meta\s+name=["']description["']\s+content=["']([^"']+)["']`)
		if m := reDesc.FindSubmatch(body); len(m) > 1 {
			res.Meta["description"] = string(m[1])
		}

		// 提取 <meta name="keywords" content="...">
		reKeywords := regexp.MustCompile(`(?i)<meta\s+name=["']keywords["']\s+content=["']([^"']+)["']`)
		if m := reKeywords.FindSubmatch(body); len(m) > 1 {
			res.Meta["keywords"] = string(m[1])
		}
	}

	// TLS 证书检查
	if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
		cert := resp.TLS.PeerCertificates[0]                             // 服务器证书
		res.CertExpiry = cert.NotAfter                                   // 证书过期时间
		res.CertDaysLeft = int64(time.Until(cert.NotAfter).Hours() / 24) // 证书剩余可用天数
		res.CertIssuer = cert.Issuer.CommonName                          // 证书签发机构
		res.CertIssuerOrg = cert.Issuer.Organization                     // 证书签发组织
		res.CertDNSNames = cert.DNSNames                                 // 证书包含的域名
		res.CertPubKeyAlg = cert.PublicKeyAlgorithm.String()             // 证书的公钥算法
		res.CertSigAlg = cert.SignatureAlgorithm.String()                // 证书的签名算法
		res.CertEmail = cert.EmailAddresses                              // 证绑定的邮箱
		res.CertIsCA = cert.IsCA                                         // 是否CA证书

		if name, ok := models.TlsVersionMap[resp.TLS.Version]; ok {
			res.TLSVersion = name
		} else {
			res.TLSVersion = fmt.Sprintf("未知(%d)", resp.TLS.Version)
		}
		if cs, ok := models.CipherSuiteMap[resp.TLS.CipherSuite]; ok {
			res.CipherSuite = cs
		} else {
			res.CipherSuite = fmt.Sprintf("未知(%d)", resp.TLS.CipherSuite)
		}
	}

	return
}
