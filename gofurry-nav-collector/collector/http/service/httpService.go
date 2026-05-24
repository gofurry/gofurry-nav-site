package service

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofurry/gofurry-nav-collector/collector/http/dao"
	"github.com/gofurry/gofurry-nav-collector/collector/http/models"
	"github.com/gofurry/gofurry-nav-collector/collector/observation"
	"github.com/gofurry/gofurry-nav-collector/common/log"
	cm "github.com/gofurry/gofurry-nav-collector/common/models"
	cs "github.com/gofurry/gofurry-nav-collector/common/service"
	"github.com/gofurry/gofurry-nav-collector/common/util"
	"github.com/gofurry/gofurry-nav-collector/roof/env"
	"github.com/sourcegraph/conc/pool"
)

var requestRunning atomic.Bool

const (
	maxHTTPPayloadURLLength         = 2048
	maxHTTPPayloadContentTypeLength = 256
	maxHTTPPayloadTitleLength       = 256
	maxHTTPPayloadServerLength      = 256
	maxHTTPPayloadMetaValueLength   = 512
	maxHTTPPayloadHeaderValueLength = 512
	maxHTTPPayloadCertTextLength    = 256
	maxHTTPPayloadCertItems         = 64
	maxHTTPPayloadVerifyErrorLength = 512
	maxHTTPPayloadCacheValueLength  = 512

	tlsHandshakeNotTLS    = "not_tls"
	tlsHandshakeCollected = "collected"
	tlsHandshakeFailed    = "failed"
)

// ============== HTTP模块 - 初始化部分 ==============

// 初始化
func InitHTTPOnStart() {
	defer func() {
		if err := recover(); err != nil {
			log.ErrorFields(map[string]interface{}{
				"event":    "init_recovered",
				"protocol": "http",
			}, err)
		}
	}()
	log.InfoFields(map[string]interface{}{
		"event":           "module_init_start",
		"interval":        time.Duration(env.GetServerConfig().Collector.Request.RequestInterval) * time.Hour,
		"protocol":        "http",
		"retention_every": time.Hour * 48,
		"workers":         env.GetServerConfig().Collector.Request.RequestThread,
	}, "HTTP 采集模块初始化开始")

	//初始化后执行一次 Request
	go Request()
	go Delete()
	// 定时任务执行 Request
	cs.AddCronJob(time.Duration(env.GetServerConfig().Collector.Request.RequestInterval)*time.Hour, Request)
	cs.AddCronJob(48*time.Hour, Delete)

	log.InfoFields(map[string]interface{}{
		"event":    "module_init_complete",
		"protocol": "http",
	}, "HTTP 采集模块初始化完成")
}

// 每天清理一次日志表
func Delete() {
	defer func() {
		if err := recover(); err != nil {
			log.ErrorFields(map[string]interface{}{
				"event":    "retention_recovered",
				"protocol": "http",
			}, err)
		}
	}()

	start := time.Now()
	keepCount := env.GetServerConfig().Collector.Request.LogCount
	log.InfoFields(map[string]interface{}{
		"event":      "retention_start",
		"keep_count": keepCount,
		"protocol":   "http",
	}, "HTTP 历史日志保留清理开始")

	// 每个域名仅保留 1500 条 request 记录
	count, deleteErr := dao.GetHTTPDao().DeleteByNum(keepCount)
	if deleteErr != nil {
		log.ErrorFields(map[string]interface{}{
			"deleted":    count,
			"duration":   time.Since(start),
			"event":      "retention_failed",
			"keep_count": keepCount,
			"protocol":   "http",
		}, "HTTP 历史日志保留清理失败: "+deleteErr.GetMsg())
	} else {
		log.InfoFields(map[string]interface{}{
			"deleted":    count,
			"duration":   time.Since(start),
			"event":      "retention_complete",
			"keep_count": keepCount,
			"protocol":   "http",
		}, "HTTP 历史日志保留清理完成")
	}
	if env.GetServerConfig().Collector.V2.ProtocolEnabled(observation.ProtocolHTTP) {
		v2Count, v2DeleteErr := observation.DeleteByProtocolLimit(observation.ProtocolHTTP, keepCount)
		if v2DeleteErr != nil {
			log.ErrorFields(map[string]interface{}{
				"deleted":    v2Count,
				"duration":   time.Since(start),
				"event":      "v2_retention_failed",
				"keep_count": keepCount,
				"protocol":   "http",
			}, "HTTP v2 observation 保留清理失败: "+v2DeleteErr.GetMsg())
		} else if v2Count > 0 {
			log.InfoFields(map[string]interface{}{
				"deleted":    v2Count,
				"duration":   time.Since(start),
				"event":      "v2_retention_complete",
				"keep_count": keepCount,
				"protocol":   "http",
			}, "HTTP v2 observation 保留清理完成")
		}
	}
}

// ============== HTTP模块 - 执行部分 ==============

// 执行 Request
func Request() {
	defer func() {
		if err := recover(); err != nil {
			log.ErrorFields(map[string]interface{}{
				"event":    "run_recovered",
				"protocol": "http",
			}, fmt.Sprintf("HTTP 采集运行触发 panic，已恢复: %v", err))
		}
	}()
	if !requestRunning.CompareAndSwap(false, true) {
		log.WarnFields(map[string]interface{}{
			"event":    "run_skipped",
			"protocol": "http",
			"reason":   "上一轮采集仍在运行",
			"status":   "skipped",
		}, "HTTP 采集已跳过：上一轮仍在运行")
		return
	}
	defer requestRunning.Store(false)

	start := time.Now()
	log.InfoFields(map[string]interface{}{
		"event":     "run_start",
		"protocol":  "http",
		"timeout":   env.GetServerConfig().Collector.ProbeBudget.HTTPTimeout(),
		"workers":   env.GetServerConfig().Collector.Request.RequestThread,
		"redirects": env.GetServerConfig().Collector.ProbeBudget.MaxHTTPRedirects(),
	}, "HTTP 采集运行开始")

	requestList, err := dao.GetHTTPDao().GetList()
	if err != nil {
		log.ErrorFields(map[string]interface{}{
			"duration": time.Since(start),
			"event":    "run_failed",
			"protocol": "http",
			"stage":    "load_targets",
		}, "HTTP 目标列表读取失败: "+err.GetMsg())
		return
	}
	// 判空
	if len(requestList) < 1 {
		log.InfoFields(map[string]interface{}{
			"duration": time.Since(start),
			"event":    "run_complete",
			"protocol": "http",
			"reason":   "目标列表为空",
			"targets":  0,
		}, "HTTP 采集完成：没有需要探测的目标")
		return
	}
	log.InfoFields(map[string]interface{}{
		"event":    "probe_start",
		"protocol": "http",
		"targets":  len(requestList),
		"workers":  env.GetServerConfig().Collector.Request.RequestThread,
	}, "HTTP 探测开始")
	requestThread := pool.New().WithMaxGoroutines(env.GetServerConfig().Collector.Request.RequestThread)
	// 遍历站点列表, 每个站点开一个线程执行 request
	for _, v := range requestList {
		requestThread.Go(getRequestResult(v))
	}
	// 等待所有采集和解析执行完毕
	requestThread.Wait()
	log.InfoFields(map[string]interface{}{
		"duration": time.Since(start),
		"event":    "run_complete",
		"protocol": "http",
		"targets":  len(requestList),
	}, "HTTP 采集运行完成")
}

// ============== HTTP模块 - 存储部分 ==============

// 解析 Request 采集结果
func getRequestResult(site models.GfnCollectorDomain) func() {
	return func() {
		defer func() {
			if err := recover(); err != nil {
				log.ErrorFields(map[string]interface{}{
					"event":    "probe_recovered",
					"protocol": "http",
					"site":     site.TargetName(),
				}, fmt.Sprintf("HTTP 单目标探测触发 panic，已恢复: %v", err))
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

		siteName := site.TargetName()

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
			log.ErrorFields(map[string]interface{}{
				"event":    "db_write_failed",
				"protocol": "http",
				"site":     siteName,
				"status":   httpSaveRecord.Status,
				"url":      result.Url,
			}, "HTTP 探测结果写入数据库失败: "+err.GetMsg())
		}
		saveErr := observation.SaveIfEnabled(observation.Input{
			SiteID:       site.SiteID,
			Target:       siteName,
			Protocol:     observation.ProtocolHTTP,
			Status:       httpSaveRecord.Status,
			ObservedAt:   time.Time(result.StartTime),
			DurationMS:   result.ResponseTime,
			ErrorCode:    result.ErrorCode,
			ErrorMessage: result.ErrorMessage,
			Payload:      buildHTTPObservationPayload(result, httpRecord),
		})
		if saveErr != nil {
			log.WarnFields(map[string]interface{}{
				"event":    "v2_observation_write_failed",
				"protocol": "http",
				"site_id":  site.SiteID,
				"site":     siteName,
				"url":      result.Url,
			}, "HTTP v2 observation 旁路写入失败: "+saveErr.GetMsg())
		}
	}
}

func buildHTTPObservationPayload(result models.HTTPModel, httpRecord models.HTTPSaveModel) map[string]any {
	finalURL := result.FinalURL
	if finalURL == "" {
		finalURL = httpRecord.Url
	}
	redirectChain := limitStringSlice(result.Redirects, maxHTTPPayloadURLLength)

	return map[string]any{
		"domain":                    httpRecord.Domain,
		"url":                       httpRecord.Url,
		"status_code":               httpRecord.StatusCode,
		"response_time_ms":          result.ResponseTime,
		"content_length":            httpRecord.ContentLength,
		"title":                     limitString(httpRecord.Title, maxHTTPPayloadTitleLength),
		"server":                    limitString(httpRecord.Server, maxHTTPPayloadServerLength),
		"redirects":                 redirectChain,
		"headers":                   limitHeaderValues(httpRecord.Headers, maxHTTPPayloadHeaderValueLength),
		"meta":                      limitStringMapValues(httpRecord.Meta, maxHTTPPayloadMetaValueLength),
		"redirect_chain":            redirectChain,
		"redirect_count":            len(result.Redirects),
		"final_url":                 limitString(finalURL, maxHTTPPayloadURLLength),
		"content_type":              limitString(result.ContentType, maxHTTPPayloadContentTypeLength),
		"security_headers":          securityHeadersWithDefaults(result.SecurityHeaders),
		"http_protocol":             result.HTTPProtocol,
		"dns_lookup_ms":             result.DNSLookupMS,
		"tcp_connect_ms":            result.TCPConnectMS,
		"tls_handshake_ms":          result.TLSHandshakeMS,
		"ttfb_ms":                   result.TTFBMS,
		"transfer_ms":               result.TransferMS,
		"remote_addr":               result.RemoteAddr,
		"remote_ip":                 result.RemoteIP,
		"body_read_bytes":           result.BodyReadBytes,
		"compressed":                result.Compressed,
		"content_encoding":          limitString(result.ContentEncoding, maxHTTPPayloadHeaderValueLength),
		"cache_policy":              buildHTTPCachePolicy(result),
		"tls_version":               httpRecord.TLSVersion,
		"cipher_suite":              httpRecord.CipherSuite,
		"cert_expiry":               httpRecord.CertExpiry,
		"cert_days_left":            httpRecord.CertDaysLeft,
		"cert_issuer":               limitString(httpRecord.CertIssuer, maxHTTPPayloadCertTextLength),
		"cert_issuer_org":           limitStringSliceItems(httpRecord.CertIssuerOrg, maxHTTPPayloadCertTextLength, maxHTTPPayloadCertItems),
		"cert_dns_names":            limitStringSliceItems(httpRecord.CertDNSNames, maxHTTPPayloadCertTextLength, maxHTTPPayloadCertItems),
		"cert_pub_key_alg":          httpRecord.CertPubKeyAlg,
		"cert_sig_alg":              httpRecord.CertSigAlg,
		"cert_public_key_algorithm": httpRecord.CertPubKeyAlg,
		"cert_signature_algorithm":  httpRecord.CertSigAlg,
		"cert_email":                limitStringSliceItems(httpRecord.CertEmail, maxHTTPPayloadCertTextLength, maxHTTPPayloadCertItems),
		"cert_is_ca":                httpRecord.CertIsCA,
		"cert_collected":            result.CertCollected,
		"cert_verified":             result.CertVerified,
		"verify_error":              limitString(result.VerifyError, maxHTTPPayloadVerifyErrorLength),
		"verify_error_category":     result.VerifyErrorCategory,
		"tls_handshake":             result.TLSHandshake,
		"cert_not_before":           formatObservationTime(result.CertNotBefore),
		"cert_not_after":            formatObservationTime(result.CertNotAfter),
		"cert_chain_length":         result.CertChainLen,
		"cert_subject_cn":           limitString(result.CertSubjectCN, maxHTTPPayloadCertTextLength),
		"cert_san_count":            result.CertSANCount,
		"ocsp_stapled":              result.OCSPStapled,
		"sct_count":                 result.SCTCount,
	}
}

func buildHTTPCachePolicy(result models.HTTPModel) map[string]string {
	return map[string]string{
		"cache_control": limitString(result.CacheControl, maxHTTPPayloadCacheValueLength),
		"etag":          limitString(result.ETag, maxHTTPPayloadCacheValueLength),
		"last_modified": limitString(result.LastModified, maxHTTPPayloadCacheValueLength),
	}
}

func formatObservationTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func limitString(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	if len([]rune(value)) <= limit {
		return value
	}
	return string([]rune(value)[:limit])
}

func limitStringSlice(values []string, limit int) []string {
	return limitStringSliceItems(values, limit, len(values))
}

func limitStringSliceItems(values []string, limit int, maxItems int) []string {
	if values == nil {
		return nil
	}
	if maxItems < 0 {
		maxItems = 0
	}
	if len(values) < maxItems {
		maxItems = len(values)
	}
	limited := make([]string, 0, maxItems)
	for _, value := range values[:maxItems] {
		limited = append(limited, limitString(value, limit))
	}
	return limited
}

func limitStringMapValues(values map[string]string, limit int) map[string]string {
	if values == nil {
		return nil
	}
	limited := make(map[string]string, len(values))
	for key, value := range values {
		limited[key] = limitString(value, limit)
	}
	return limited
}

func limitHeaderValues(values map[string][]string, limit int) map[string][]string {
	if values == nil {
		return nil
	}
	limited := make(map[string][]string, len(values))
	for key, headerValues := range values {
		limited[key] = limitStringSlice(headerValues, limit)
	}
	return limited
}

func detectHTTPSecurityHeaders(header http.Header) map[string]bool {
	return map[string]bool{
		"strict_transport_security": header.Get("Strict-Transport-Security") != "",
		"content_security_policy":   header.Get("Content-Security-Policy") != "",
		"x_frame_options":           header.Get("X-Frame-Options") != "",
		"x_content_type_options":    header.Get("X-Content-Type-Options") != "",
		"referrer_policy":           header.Get("Referrer-Policy") != "",
		"permissions_policy":        header.Get("Permissions-Policy") != "",
	}
}

func securityHeadersWithDefaults(values map[string]bool) map[string]bool {
	defaults := detectHTTPSecurityHeaders(http.Header{})
	for key, value := range values {
		if _, ok := defaults[key]; ok {
			defaults[key] = value
		}
	}
	return defaults
}

func verifyTLSCertificate(state *tls.ConnectionState, dnsName string) (bool, string) {
	verified, verifyError, _ := verifyTLSCertificateDetailed(state, dnsName, nil)
	return verified, verifyError
}

func verifyTLSCertificateWithRoots(state *tls.ConnectionState, dnsName string, roots *x509.CertPool) (bool, string) {
	verified, verifyError, _ := verifyTLSCertificateDetailed(state, dnsName, roots)
	return verified, verifyError
}

func verifyTLSCertificateDetailed(state *tls.ConnectionState, dnsName string, roots *x509.CertPool) (bool, string, string) {
	if state == nil || len(state.PeerCertificates) == 0 {
		return false, "未采集到服务端证书", "other"
	}
	if dnsName == "" {
		return false, "无法确定证书校验域名", "other"
	}

	intermediates := x509.NewCertPool()
	for _, cert := range state.PeerCertificates[1:] {
		intermediates.AddCert(cert)
	}
	_, err := state.PeerCertificates[0].Verify(x509.VerifyOptions{
		DNSName:       dnsName,
		Roots:         roots,
		Intermediates: intermediates,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	})
	if err != nil {
		return false, err.Error(), classifyTLSVerifyError(err, state.PeerCertificates[0])
	}
	return true, "", ""
}

func classifyTLSVerifyError(err error, cert *x509.Certificate) string {
	var hostnameErr x509.HostnameError
	if errors.As(err, &hostnameErr) {
		return "hostname_mismatch"
	}

	var unknownAuthorityErr x509.UnknownAuthorityError
	if errors.As(err, &unknownAuthorityErr) {
		return "unknown_authority"
	}

	var invalidErr x509.CertificateInvalidError
	if errors.As(err, &invalidErr) {
		switch invalidErr.Reason {
		case x509.Expired:
			now := time.Now()
			if cert != nil && now.Before(cert.NotBefore) {
				return "not_yet_valid"
			}
			return "expired"
		case x509.IncompatibleUsage:
			return "incompatible_usage"
		default:
			return "other"
		}
	}

	return "other"
}

func tlsVerifyHost(finalURL string, fallback string) string {
	if parsed, err := url.Parse(finalURL); err == nil && parsed.Host != "" {
		return parsed.Hostname()
	}
	return hostWithoutPort(fallback)
}

func hostWithoutPort(host string) string {
	if host == "" {
		return ""
	}
	if splitHost, _, err := net.SplitHostPort(host); err == nil {
		return splitHost
	}
	return host
}

func redactProxyURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "(invalid)"
	}
	if parsed.User != nil {
		parsed.User = url.UserPassword("****", "****")
	}
	return parsed.String()
}

// ============== HTTP模块 - 采集和解析部分 ==============

// 执行 Request 采集
func performRequest(site models.GfnCollectorDomain) (res models.HTTPModel) {
	probeStart := time.Now()
	res.Domain = site.TargetName()
	res.Url = res.Domain
	if site.TLS == "1" {
		res.Url = "https://" + res.Url
	} else {
		res.Url = "http://" + res.Url
	}
	res.FinalURL = res.Url
	res.Meta = make(map[string]string)
	res.Headers = make(map[string][]string)
	res.Redirects = []string{}
	res.SecurityHeaders = securityHeadersWithDefaults(nil)
	res.TLSHandshake = tlsHandshakeNotTLS
	if site.TLS == "1" {
		res.TLSHandshake = tlsHandshakeFailed
	}
	probeBudget := env.GetServerConfig().Collector.ProbeBudget

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		TLSHandshakeTimeout: probeBudget.TLSHandshakeTimeout(),
	}
	// 设置代理
	if site.Proxy == "1" {
		proxyRaw := env.GetServerConfig().Collector.Proxy
		proxyURL, err := url.Parse(proxyRaw)
		if err != nil || proxyURL.Scheme == "" || proxyURL.Host == "" {
			res.StartTime = cm.LocalTime(time.Now())
			res.ResponseTime = time.Since(probeStart).Milliseconds()
			res.ErrorCode = "http_proxy_config_invalid"
			if err != nil {
				res.ErrorMessage = err.Error()
			} else {
				res.ErrorMessage = "代理地址缺少 scheme 或 host"
			}
			log.ErrorFields(map[string]interface{}{
				"event":    "proxy_config_invalid",
				"protocol": "http",
				"proxy":    redactProxyURL(proxyRaw),
				"site_id":  site.SiteID,
				"target":   res.Domain,
				"url":      res.Url,
			}, "HTTP 代理配置无效: "+res.ErrorMessage)
			return
		}
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
		res.ErrorCode = "http_request_create_failed"
		res.ErrorMessage = err.Error()
		log.ErrorFields(map[string]interface{}{
			"event":    "request_create_failed",
			"protocol": "http",
			"url":      res.Url,
		}, "HTTP 请求创建失败: "+err.Error())
		return
	}
	// 设置请求头
	for k, v := range models.HeadersMap {
		req.Header.Set(k, v)
	}

	var start time.Time
	var requestStartedAt time.Time
	var dnsStartedAt time.Time
	var tcpStartedAt time.Time
	var tlsStartedAt time.Time
	trace := &httptrace.ClientTrace{
		GetConn: func(string) {
			if requestStartedAt.IsZero() {
				requestStartedAt = time.Now()
			}
		},
		DNSStart: func(httptrace.DNSStartInfo) {
			dnsStartedAt = time.Now()
		},
		DNSDone: func(httptrace.DNSDoneInfo) {
			if !dnsStartedAt.IsZero() {
				res.DNSLookupMS = time.Since(dnsStartedAt).Milliseconds()
			}
		},
		ConnectStart: func(string, string) {
			if tcpStartedAt.IsZero() {
				tcpStartedAt = time.Now()
			}
		},
		ConnectDone: func(string, string, error) {
			if !tcpStartedAt.IsZero() && res.TCPConnectMS == 0 {
				res.TCPConnectMS = time.Since(tcpStartedAt).Milliseconds()
			}
		},
		TLSHandshakeStart: func() {
			tlsStartedAt = time.Now()
		},
		TLSHandshakeDone: func(tls.ConnectionState, error) {
			if !tlsStartedAt.IsZero() {
				res.TLSHandshakeMS = time.Since(tlsStartedAt).Milliseconds()
			}
		},
		GotConn: func(info httptrace.GotConnInfo) {
			if info.Conn != nil {
				res.RemoteAddr = info.Conn.RemoteAddr().String()
				res.RemoteIP = hostWithoutPort(res.RemoteAddr)
			}
		},
		GotFirstResponseByte: func() {
			if requestStartedAt.IsZero() {
				requestStartedAt = start
			}
			res.TTFBMS = time.Since(requestStartedAt).Milliseconds()
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	// 请求开始
	start = time.Now()
	requestStartedAt = start
	res.StartTime = cm.LocalTime(time.Now())
	resp, err := client.Do(req)
	if err != nil {
		res.ResponseTime = time.Since(start).Milliseconds()
		res.Redirects = redirects
		res.ErrorCode = "http_probe_failed"
		res.ErrorMessage = err.Error()
		log.WarnFields(map[string]interface{}{
			"duration": time.Since(start),
			"event":    "probe_failed",
			"protocol": "http",
			"url":      res.Url,
		}, "HTTP 探测失败: "+err.Error())
		return
	}
	defer resp.Body.Close()

	// 响应时间
	res.ResponseTime = time.Since(start).Milliseconds()
	res.StatusCode = int64(resp.StatusCode)
	res.Redirects = redirects
	if resp.Request != nil && resp.Request.URL != nil {
		res.FinalURL = resp.Request.URL.String()
	}
	res.ContentType = resp.Header.Get("Content-Type")
	res.SecurityHeaders = detectHTTPSecurityHeaders(resp.Header)
	res.HTTPProtocol = resp.Proto
	res.ContentEncoding = resp.Header.Get("Content-Encoding")
	res.Compressed = resp.Uncompressed || res.ContentEncoding != ""
	res.CacheControl = resp.Header.Get("Cache-Control")
	res.ETag = resp.Header.Get("ETag")
	res.LastModified = resp.Header.Get("Last-Modified")

	for _, v := range models.CommonHeaders {
		if val := resp.Header.Values(v); len(val) > 0 {
			res.Headers[v] = val
		}
	}

	res.Server = resp.Header.Get("Server")

	// 读取响应体 限制 1MB
	bodyReadStart := time.Now()
	body, err := io.ReadAll(io.LimitReader(resp.Body, probeBudget.MaxHTTPResponseBytes()))
	res.TransferMS = time.Since(bodyReadStart).Milliseconds()
	if err == nil {
		res.ContentLength = int64(len(body))
		res.BodyReadBytes = int64(len(body))

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
		res.TLSHandshake = tlsHandshakeCollected
		res.CertCollected = true
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
		res.CertNotBefore = cert.NotBefore
		res.CertNotAfter = cert.NotAfter
		res.CertChainLen = len(resp.TLS.PeerCertificates)
		res.CertSubjectCN = cert.Subject.CommonName
		res.CertSANCount = len(cert.DNSNames)
		res.OCSPStapled = len(resp.TLS.OCSPResponse) > 0
		res.SCTCount = len(resp.TLS.SignedCertificateTimestamps)

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

		res.CertVerified, res.VerifyError, res.VerifyErrorCategory = verifyTLSCertificateDetailed(resp.TLS, tlsVerifyHost(res.FinalURL, res.Domain), nil)
	}

	return
}
