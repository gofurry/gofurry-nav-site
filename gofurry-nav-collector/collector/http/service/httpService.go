package service

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/PuerkitoBio/goquery"
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
	maxHTTPPayloadPageSummaryLength = 1024
	maxHTTPPayloadCertTextLength    = 256
	maxHTTPPayloadCertItems         = 64
	maxHTTPPayloadVerifyErrorLength = 512
	maxHTTPPayloadCacheValueLength  = 512
	maxHTTPPayloadIconLinks         = 16

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
			OpenGraph:     result.OpenGraph,
			TwitterCard:   result.TwitterCard,
			CanonicalURL:  result.CanonicalURL,
			HTMLLang:      result.HTMLLang,
			MetaRefresh:   result.MetaRefresh,
			IconLinks:     result.IconLinks,
			CookieSummary: result.CookieSummary,
			ServerHints:   result.ServerHints,
			CrossOrigin:   result.CrossOriginSummary,
			ContentLang:   result.ContentLanguage,
			PageSummary:   result.PageTextSummary,
			SharePreview:  result.SharePreview,
			RedirectHint:  result.RedirectHint,
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
		"domain":                     httpRecord.Domain,
		"url":                        httpRecord.Url,
		"status_code":                httpRecord.StatusCode,
		"response_time_ms":           result.ResponseTime,
		"content_length":             httpRecord.ContentLength,
		"title":                      limitString(httpRecord.Title, maxHTTPPayloadTitleLength),
		"server":                     limitString(httpRecord.Server, maxHTTPPayloadServerLength),
		"redirects":                  redirectChain,
		"headers":                    limitHeaderValues(httpRecord.Headers, maxHTTPPayloadHeaderValueLength),
		"meta":                       limitStringMapValues(httpRecord.Meta, maxHTTPPayloadMetaValueLength),
		"open_graph":                 limitStringMapValues(httpRecord.OpenGraph, maxHTTPPayloadMetaValueLength),
		"twitter_card":               limitStringMapValues(httpRecord.TwitterCard, maxHTTPPayloadMetaValueLength),
		"canonical_url":              limitString(httpRecord.CanonicalURL, maxHTTPPayloadURLLength),
		"html_lang":                  limitString(httpRecord.HTMLLang, maxHTTPPayloadMetaValueLength),
		"meta_refresh":               limitHTTPMetaRefresh(httpRecord.MetaRefresh),
		"icon_links":                 limitHTTPLinkInfos(httpRecord.IconLinks),
		"cookie_summary":             httpRecord.CookieSummary,
		"server_hints":               limitHTTPServerHints(httpRecord.ServerHints),
		"cross_origin_summary":       limitHTTPCrossOrigin(httpRecord.CrossOrigin),
		"content_language_effective": limitString(httpRecord.ContentLang, maxHTTPPayloadMetaValueLength),
		"page_text_summary":          limitString(httpRecord.PageSummary, maxHTTPPayloadPageSummaryLength),
		"share_preview":              limitHTTPSharePreview(httpRecord.SharePreview),
		"redirect_hint":              limitHTTPRedirectHint(httpRecord.RedirectHint),
		"redirect_chain":             redirectChain,
		"redirect_count":             len(result.Redirects),
		"final_url":                  limitString(finalURL, maxHTTPPayloadURLLength),
		"content_type":               limitString(result.ContentType, maxHTTPPayloadContentTypeLength),
		"security_headers":           securityHeadersWithDefaults(result.SecurityHeaders),
		"security_header_summary":    buildHTTPSecurityHeaderSummary(result.SecurityHeaderValues),
		"http_protocol":              result.HTTPProtocol,
		"dns_lookup_ms":              result.DNSLookupMS,
		"tcp_connect_ms":             result.TCPConnectMS,
		"tls_handshake_ms":           result.TLSHandshakeMS,
		"ttfb_ms":                    result.TTFBMS,
		"transfer_ms":                result.TransferMS,
		"remote_addr":                result.RemoteAddr,
		"remote_ip":                  result.RemoteIP,
		"body_read_bytes":            result.BodyReadBytes,
		"compressed":                 result.Compressed,
		"content_encoding":           limitString(result.ContentEncoding, maxHTTPPayloadHeaderValueLength),
		"cache_policy":               buildHTTPCachePolicy(result),
		"tls_version":                httpRecord.TLSVersion,
		"cipher_suite":               httpRecord.CipherSuite,
		"cert_expiry":                httpRecord.CertExpiry,
		"cert_days_left":             httpRecord.CertDaysLeft,
		"cert_issuer":                limitString(httpRecord.CertIssuer, maxHTTPPayloadCertTextLength),
		"cert_issuer_org":            limitStringSliceItems(httpRecord.CertIssuerOrg, maxHTTPPayloadCertTextLength, maxHTTPPayloadCertItems),
		"cert_dns_names":             limitStringSliceItems(httpRecord.CertDNSNames, maxHTTPPayloadCertTextLength, maxHTTPPayloadCertItems),
		"cert_pub_key_alg":           httpRecord.CertPubKeyAlg,
		"cert_sig_alg":               httpRecord.CertSigAlg,
		"cert_public_key_algorithm":  httpRecord.CertPubKeyAlg,
		"cert_signature_algorithm":   httpRecord.CertSigAlg,
		"cert_email":                 limitStringSliceItems(httpRecord.CertEmail, maxHTTPPayloadCertTextLength, maxHTTPPayloadCertItems),
		"cert_is_ca":                 httpRecord.CertIsCA,
		"cert_collected":             result.CertCollected,
		"cert_verified":              result.CertVerified,
		"verify_error":               limitString(result.VerifyError, maxHTTPPayloadVerifyErrorLength),
		"verify_error_category":      result.VerifyErrorCategory,
		"tls_handshake":              result.TLSHandshake,
		"cert_not_before":            formatObservationTime(result.CertNotBefore),
		"cert_not_after":             formatObservationTime(result.CertNotAfter),
		"cert_chain_length":          result.CertChainLen,
		"cert_subject_cn":            limitString(result.CertSubjectCN, maxHTTPPayloadCertTextLength),
		"cert_san_count":             result.CertSANCount,
		"ocsp_stapled":               result.OCSPStapled,
		"sct_count":                  result.SCTCount,
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

func limitHTTPLinkInfos(values []models.HTTPLinkInfo) []models.HTTPLinkInfo {
	if values == nil {
		return nil
	}
	if len(values) > maxHTTPPayloadIconLinks {
		values = values[:maxHTTPPayloadIconLinks]
	}
	limited := make([]models.HTTPLinkInfo, 0, len(values))
	for _, value := range values {
		limited = append(limited, models.HTTPLinkInfo{
			Rel:   limitString(value.Rel, maxHTTPPayloadMetaValueLength),
			Href:  limitString(value.Href, maxHTTPPayloadURLLength),
			Type:  limitString(value.Type, maxHTTPPayloadMetaValueLength),
			Sizes: limitString(value.Sizes, maxHTTPPayloadMetaValueLength),
		})
	}
	return limited
}

func limitHTTPMetaRefresh(value *models.HTTPMetaRefresh) *models.HTTPMetaRefresh {
	if value == nil {
		return nil
	}
	return &models.HTTPMetaRefresh{
		Present:      value.Present,
		DelaySeconds: value.DelaySeconds,
		URL:          limitString(value.URL, maxHTTPPayloadURLLength),
	}
}

func limitHTTPServerHints(value *models.HTTPServerHints) *models.HTTPServerHints {
	if value == nil {
		return nil
	}
	return &models.HTTPServerHints{
		Server:     limitString(value.Server, maxHTTPPayloadServerLength),
		XPoweredBy: limitString(value.XPoweredBy, maxHTTPPayloadServerLength),
		Generator:  limitString(value.Generator, maxHTTPPayloadMetaValueLength),
	}
}

func limitHTTPCrossOrigin(value *models.HTTPCrossOrigin) *models.HTTPCrossOrigin {
	if value == nil {
		return nil
	}
	return &models.HTTPCrossOrigin{
		CrossOriginOpenerPolicy:   limitString(value.CrossOriginOpenerPolicy, maxHTTPPayloadHeaderValueLength),
		CrossOriginEmbedderPolicy: limitString(value.CrossOriginEmbedderPolicy, maxHTTPPayloadHeaderValueLength),
		CrossOriginResourcePolicy: limitString(value.CrossOriginResourcePolicy, maxHTTPPayloadHeaderValueLength),
		AccessControlAllowOrigin:  limitString(value.AccessControlAllowOrigin, maxHTTPPayloadHeaderValueLength),
	}
}

func limitHTTPSharePreview(value *models.HTTPSharePreview) *models.HTTPSharePreview {
	if value == nil {
		return nil
	}
	return &models.HTTPSharePreview{
		Title:       limitString(value.Title, maxHTTPPayloadTitleLength),
		Description: limitString(value.Description, maxHTTPPayloadMetaValueLength),
		SiteName:    limitString(value.SiteName, maxHTTPPayloadMetaValueLength),
		Image:       limitString(value.Image, maxHTTPPayloadURLLength),
		URL:         limitString(value.URL, maxHTTPPayloadURLLength),
	}
}

func limitHTTPRedirectHint(value *models.HTTPRedirectHint) *models.HTTPRedirectHint {
	if value == nil {
		return nil
	}
	return &models.HTTPRedirectHint{
		FinalURLDifferent:     value.FinalURLDifferent,
		CanonicalURLDifferent: value.CanonicalURLDifferent,
		MetaRefreshPresent:    value.MetaRefreshPresent,
		MetaRefreshURL:        limitString(value.MetaRefreshURL, maxHTTPPayloadURLLength),
	}
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

func captureHTTPSecurityHeaderValues(header http.Header) map[string]string {
	return map[string]string{
		"strict_transport_security": header.Get("Strict-Transport-Security"),
		"content_security_policy":   header.Get("Content-Security-Policy"),
		"x_frame_options":           header.Get("X-Frame-Options"),
		"x_content_type_options":    header.Get("X-Content-Type-Options"),
		"referrer_policy":           header.Get("Referrer-Policy"),
		"permissions_policy":        header.Get("Permissions-Policy"),
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

func buildHTTPSecurityHeaderSummary(values map[string]string) map[string]any {
	return map[string]any{
		"hsts":                    summarizeHSTS(values["strict_transport_security"]),
		"content_security_policy": summarizeCSP(values["content_security_policy"]),
		"x_frame_options":         summarizeXFrameOptions(values["x_frame_options"]),
		"x_content_type_options":  summarizeXContentTypeOptions(values["x_content_type_options"]),
		"referrer_policy":         summarizeReferrerPolicy(values["referrer_policy"]),
		"permissions_policy":      summarizePermissionsPolicy(values["permissions_policy"]),
	}
}

func summarizeHSTS(value string) map[string]any {
	summary := map[string]any{
		"present":            value != "",
		"max_age":            nil,
		"include_subdomains": false,
		"preload":            false,
	}
	if value == "" {
		return summary
	}

	for _, part := range strings.Split(value, ";") {
		token := strings.TrimSpace(part)
		lowerToken := strings.ToLower(token)
		switch {
		case strings.HasPrefix(lowerToken, "max-age="):
			raw := strings.TrimSpace(token[len("max-age="):])
			if maxAge, err := strconv.ParseInt(raw, 10, 64); err == nil {
				summary["max_age"] = maxAge
			}
		case lowerToken == "includesubdomains":
			summary["include_subdomains"] = true
		case lowerToken == "preload":
			summary["preload"] = true
		}
	}
	return summary
}

func summarizeCSP(value string) map[string]any {
	summary := map[string]any{
		"present":         value != "",
		"has_default_src": false,
		"unsafe_inline":   false,
		"unsafe_eval":     false,
		"wildcard_source": false,
	}
	if value == "" {
		return summary
	}

	directives := strings.Split(value, ";")
	for _, directive := range directives {
		trimmed := strings.TrimSpace(directive)
		if trimmed == "" {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) == 0 {
			continue
		}
		if strings.ToLower(fields[0]) == "default-src" {
			summary["has_default_src"] = true
		}
		for _, field := range fields[1:] {
			switch strings.ToLower(field) {
			case "'unsafe-inline'":
				summary["unsafe_inline"] = true
			case "'unsafe-eval'":
				summary["unsafe_eval"] = true
			case "*":
				summary["wildcard_source"] = true
			}
		}
	}
	return summary
}

func summarizeXFrameOptions(value string) map[string]any {
	mode := ""
	if value != "" {
		mode = strings.ToUpper(strings.TrimSpace(value))
	}
	return map[string]any{
		"present": value != "",
		"mode":    mode,
	}
}

func summarizeXContentTypeOptions(value string) map[string]any {
	mode := strings.ToLower(strings.TrimSpace(value))
	return map[string]any{
		"present": value != "",
		"nosniff": mode == "nosniff",
	}
}

func summarizeReferrerPolicy(value string) map[string]any {
	return map[string]any{
		"present": value != "",
		"policy":  limitString(value, maxHTTPPayloadHeaderValueLength),
	}
}

func summarizePermissionsPolicy(value string) map[string]any {
	return map[string]any{
		"present": value != "",
		"policy":  limitString(value, maxHTTPPayloadHeaderValueLength),
	}
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

func enrichHTTPPageDetails(res *models.HTTPModel, body []byte) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return
	}
	if res.Meta == nil {
		res.Meta = make(map[string]string)
	}
	res.OpenGraph = make(map[string]string)
	res.TwitterCard = make(map[string]string)

	if title := normalizeHTTPText(doc.Find("title").First().Text()); title != "" {
		res.Title = limitString(title, maxHTTPPayloadTitleLength)
	}
	if htmlLang, ok := doc.Find("html").First().Attr("lang"); ok {
		res.HTMLLang = limitString(normalizeHTTPText(htmlLang), maxHTTPPayloadMetaValueLength)
	}

	doc.Find("meta").Each(func(_ int, selection *goquery.Selection) {
		content := normalizeHTTPText(selection.AttrOr("content", ""))
		if content == "" {
			if charset := normalizeHTTPText(selection.AttrOr("charset", "")); charset != "" {
				res.Meta["charset"] = limitString(charset, maxHTTPPayloadMetaValueLength)
			}
			return
		}

		name := strings.ToLower(strings.TrimSpace(selection.AttrOr("name", "")))
		property := strings.ToLower(strings.TrimSpace(selection.AttrOr("property", "")))
		httpEquiv := strings.ToLower(strings.TrimSpace(selection.AttrOr("http-equiv", "")))

		switch name {
		case "description", "keywords", "author", "generator", "theme-color", "robots", "viewport", "application-name":
			res.Meta[normalizeHTTPMetaKey(name)] = limitString(content, maxHTTPPayloadMetaValueLength)
		}
		if httpEquiv == "content-type" {
			if charset := charsetFromContentType(content); charset != "" {
				res.Meta["charset"] = limitString(charset, maxHTTPPayloadMetaValueLength)
			}
		}
		if httpEquiv == "refresh" {
			res.MetaRefresh = parseHTTPMetaRefresh(content, res.FinalURL)
		}
		if strings.HasPrefix(property, "og:") {
			addAllowedHTTPMapValue(res.OpenGraph, strings.TrimPrefix(property, "og:"), content, res.FinalURL)
		}
		if strings.HasPrefix(name, "twitter:") {
			addAllowedHTTPMapValue(res.TwitterCard, strings.TrimPrefix(name, "twitter:"), content, res.FinalURL)
		}
	})

	doc.Find("link").EachWithBreak(func(_ int, selection *goquery.Selection) bool {
		rel := normalizeHTTPText(selection.AttrOr("rel", ""))
		href := normalizeHTTPText(selection.AttrOr("href", ""))
		if href == "" {
			return true
		}
		relLower := strings.ToLower(rel)
		if strings.Contains(relLower, "canonical") && res.CanonicalURL == "" {
			res.CanonicalURL = resolveHTTPURL(href, res.FinalURL)
		}
		if isHTTPIconRel(relLower) && len(res.IconLinks) < maxHTTPPayloadIconLinks {
			res.IconLinks = append(res.IconLinks, models.HTTPLinkInfo{
				Rel:   limitString(rel, maxHTTPPayloadMetaValueLength),
				Href:  limitString(resolveHTTPURL(href, res.FinalURL), maxHTTPPayloadURLLength),
				Type:  limitString(normalizeHTTPText(selection.AttrOr("type", "")), maxHTTPPayloadMetaValueLength),
				Sizes: limitString(normalizeHTTPText(selection.AttrOr("sizes", "")), maxHTTPPayloadMetaValueLength),
			})
		}
		return true
	})

	if res.ServerHints == nil {
		res.ServerHints = &models.HTTPServerHints{}
	}
	res.ServerHints.Generator = res.Meta["generator"]
	if res.ContentLanguage == "" {
		res.ContentLanguage = res.HTMLLang
	}
	res.SharePreview = buildHTTPSharePreview(res)
	res.PageTextSummary = buildHTTPPageTextSummary(res)
	res.RedirectHint = buildHTTPRedirectHint(res)
	if len(res.OpenGraph) == 0 {
		res.OpenGraph = nil
	}
	if len(res.TwitterCard) == 0 {
		res.TwitterCard = nil
	}
}

func normalizeHTTPText(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func normalizeHTTPMetaKey(value string) string {
	return strings.ReplaceAll(value, "-", "_")
}

func charsetFromContentType(value string) string {
	for _, part := range strings.Split(value, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(strings.ToLower(part), "charset=") {
			return strings.Trim(strings.TrimSpace(part[len("charset="):]), `"'`)
		}
	}
	return ""
}

func addAllowedHTTPMapValue(values map[string]string, key string, value string, baseURL string) {
	allowed := map[string]bool{
		"title":       true,
		"description": true,
		"site_name":   true,
		"type":        true,
		"image":       true,
		"url":         true,
		"card":        true,
		"site":        true,
	}
	if !allowed[key] || value == "" {
		return
	}
	if key == "image" || key == "url" {
		value = resolveHTTPURL(value, baseURL)
		values[key] = limitString(value, maxHTTPPayloadURLLength)
		return
	}
	values[key] = limitString(value, maxHTTPPayloadMetaValueLength)
}

func resolveHTTPURL(value string, baseURL string) string {
	parsed, err := url.Parse(value)
	if err != nil {
		return limitString(value, maxHTTPPayloadURLLength)
	}
	if parsed.IsAbs() {
		return limitString(parsed.String(), maxHTTPPayloadURLLength)
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return limitString(value, maxHTTPPayloadURLLength)
	}
	return limitString(base.ResolveReference(parsed).String(), maxHTTPPayloadURLLength)
}

func parseHTTPMetaRefresh(content string, baseURL string) *models.HTTPMetaRefresh {
	refresh := &models.HTTPMetaRefresh{Present: true}
	parts := strings.Split(content, ";")
	if len(parts) > 0 {
		if delay, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64); err == nil {
			refresh.DelaySeconds = &delay
		}
	}
	for _, part := range parts[1:] {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(strings.ToLower(part), "url=") {
			refresh.URL = resolveHTTPURL(strings.Trim(strings.TrimSpace(part[len("url="):]), `"'`), baseURL)
			break
		}
	}
	return refresh
}

func isHTTPIconRel(rel string) bool {
	if rel == "manifest" {
		return true
	}
	for _, token := range strings.Fields(rel) {
		switch token {
		case "icon", "shortcut", "apple-touch-icon":
			return true
		}
	}
	return false
}

func buildHTTPCookieSummary(values []string) *models.HTTPCookieSummary {
	if len(values) == 0 {
		return nil
	}
	summary := &models.HTTPCookieSummary{SetCookieCount: len(values)}
	for _, value := range values {
		lowerValue := strings.ToLower(value)
		if strings.Contains(lowerValue, "; secure") {
			summary.SecureCount++
		}
		if strings.Contains(lowerValue, "; httponly") {
			summary.HTTPOnlyCount++
		}
		switch {
		case strings.Contains(lowerValue, "samesite=lax"):
			summary.SameSiteLaxCount++
		case strings.Contains(lowerValue, "samesite=strict"):
			summary.SameSiteStrictCount++
		case strings.Contains(lowerValue, "samesite=none"):
			summary.SameSiteNoneCount++
		}
	}
	return summary
}

func buildHTTPCrossOriginSummary(header http.Header) *models.HTTPCrossOrigin {
	summary := &models.HTTPCrossOrigin{
		CrossOriginOpenerPolicy:   header.Get("Cross-Origin-Opener-Policy"),
		CrossOriginEmbedderPolicy: header.Get("Cross-Origin-Embedder-Policy"),
		CrossOriginResourcePolicy: header.Get("Cross-Origin-Resource-Policy"),
		AccessControlAllowOrigin:  header.Get("Access-Control-Allow-Origin"),
	}
	if summary.CrossOriginOpenerPolicy == "" &&
		summary.CrossOriginEmbedderPolicy == "" &&
		summary.CrossOriginResourcePolicy == "" &&
		summary.AccessControlAllowOrigin == "" {
		return nil
	}
	return summary
}

func buildHTTPSharePreview(res *models.HTTPModel) *models.HTTPSharePreview {
	preview := &models.HTTPSharePreview{
		Title:       firstNonEmpty(res.OpenGraph["title"], res.TwitterCard["title"], res.Title),
		Description: firstNonEmpty(res.OpenGraph["description"], res.TwitterCard["description"], res.Meta["description"]),
		SiteName:    firstNonEmpty(res.OpenGraph["site_name"], res.Domain),
		Image:       firstNonEmpty(res.OpenGraph["image"], res.TwitterCard["image"]),
		URL:         firstNonEmpty(res.OpenGraph["url"], res.CanonicalURL, res.FinalURL),
	}
	if preview.Title == "" && preview.Description == "" && preview.Image == "" && preview.URL == "" {
		return nil
	}
	return limitHTTPSharePreview(preview)
}

func buildHTTPPageTextSummary(res *models.HTTPModel) string {
	parts := []string{
		res.Title,
		res.Meta["description"],
		res.Meta["keywords"],
		res.OpenGraph["site_name"],
	}
	return limitString(normalizeHTTPText(strings.Join(nonEmptyHTTPStrings(parts), " ")), maxHTTPPayloadPageSummaryLength)
}

func buildHTTPRedirectHint(res *models.HTTPModel) *models.HTTPRedirectHint {
	hint := &models.HTTPRedirectHint{
		FinalURLDifferent:     res.FinalURL != "" && res.Url != "" && res.FinalURL != res.Url,
		CanonicalURLDifferent: res.CanonicalURL != "" && res.FinalURL != "" && res.CanonicalURL != res.FinalURL,
	}
	if res.MetaRefresh != nil {
		hint.MetaRefreshPresent = true
		hint.MetaRefreshURL = res.MetaRefresh.URL
	}
	if !hint.FinalURLDifferent && !hint.CanonicalURLDifferent && !hint.MetaRefreshPresent {
		return nil
	}
	return limitHTTPRedirectHint(hint)
}

func nonEmptyHTTPStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
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
	res.SecurityHeaderValues = captureHTTPSecurityHeaderValues(resp.Header)
	res.HTTPProtocol = resp.Proto
	res.ContentEncoding = resp.Header.Get("Content-Encoding")
	res.Compressed = resp.Uncompressed || res.ContentEncoding != ""
	res.CacheControl = resp.Header.Get("Cache-Control")
	res.ETag = resp.Header.Get("ETag")
	res.LastModified = resp.Header.Get("Last-Modified")
	res.CookieSummary = buildHTTPCookieSummary(resp.Header.Values("Set-Cookie"))
	res.CrossOriginSummary = buildHTTPCrossOriginSummary(resp.Header)
	res.ContentLanguage = resp.Header.Get("Content-Language")

	for _, v := range models.CommonHeaders {
		if val := resp.Header.Values(v); len(val) > 0 {
			res.Headers[v] = limitStringSlice(val, maxHTTPPayloadHeaderValueLength)
		}
	}

	res.Server = resp.Header.Get("Server")
	res.ServerHints = &models.HTTPServerHints{
		Server:     res.Server,
		XPoweredBy: resp.Header.Get("X-Powered-By"),
	}

	// 读取响应体 限制 1MB
	bodyReadStart := time.Now()
	body, err := io.ReadAll(io.LimitReader(resp.Body, probeBudget.MaxHTTPResponseBytes()))
	res.TransferMS = time.Since(bodyReadStart).Milliseconds()
	if err == nil {
		res.ContentLength = int64(len(body))
		res.BodyReadBytes = int64(len(body))
		enrichHTTPPageDetails(&res, body)
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
