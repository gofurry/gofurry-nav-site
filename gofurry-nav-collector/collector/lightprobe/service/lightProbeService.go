package service

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofurry/gofurry-nav-collector/collector/lightprobe/dao"
	"github.com/gofurry/gofurry-nav-collector/collector/lightprobe/models"
	"github.com/gofurry/gofurry-nav-collector/collector/observation"
	runstate "github.com/gofurry/gofurry-nav-collector/collector/scheduler"
	"github.com/gofurry/gofurry-nav-collector/common/log"
	cs "github.com/gofurry/gofurry-nav-collector/common/service"
	"github.com/gofurry/gofurry-nav-collector/roof/env"
	"golang.org/x/net/publicsuffix"
)

const (
	rdapBootstrapURLDefault = "https://data.iana.org/rdap/dns.json"

	lightProbeMaxTextLength = 512
	lightProbeMaxItems      = 16
	lightProbeRedirects     = 3

	rdapBootstrapCacheTTL = 24 * time.Hour
)

var (
	rdapRunning        atomic.Bool
	robotsRunning      atomic.Bool
	securityTXTRunning atomic.Bool
	pageAssetsRunning  atomic.Bool
	portCheckRunning   atomic.Bool

	rdapBootstrapURL = rdapBootstrapURLDefault
	rdapBootstrapMu  sync.Mutex
	rdapBootstrap    cachedRDAPBootstrap
)

type cachedRDAPBootstrap struct {
	expiresAt time.Time
	servers   map[string]string
}

type probeResult struct {
	Status       string
	DurationMS   int64
	ErrorCode    string
	ErrorMessage string
	Payload      map[string]any
}

type httpProbeResponse struct {
	StatusCode          int
	ContentType         string
	ContentLengthHeader int64
	Body                []byte
	BodyTruncated       bool
	DurationMS          int64
}

type pageAssetDeclaration struct {
	Rel   string `json:"rel,omitempty"`
	Href  string `json:"href,omitempty"`
	Type  string `json:"type,omitempty"`
	Sizes string `json:"sizes,omitempty"`
}

// InitLightProbeOnStart 注册默认关闭的 v2 低频轻探测任务。
func InitLightProbeOnStart() {
	cfg := env.GetServerConfig().Collector.V2
	if cfg.ProtocolEnabled(observation.ProtocolRDAP) {
		interval := cfg.LightProbe.RDAP.Interval()
		go RunRDAP()
		cs.AddCronJob(interval, RunRDAP)
		log.InfoFields(map[string]interface{}{
			"event":    "light_probe_registered",
			"interval": interval,
			"protocol": observation.ProtocolRDAP,
		}, "RDAP 低频轻探测已注册")
	}
	if cfg.ProtocolEnabled(observation.ProtocolRobots) {
		interval := cfg.LightProbe.Robots.Interval()
		go RunRobots()
		cs.AddCronJob(interval, RunRobots)
		log.InfoFields(map[string]interface{}{
			"event":    "light_probe_registered",
			"interval": interval,
			"protocol": observation.ProtocolRobots,
		}, "robots.txt 低频轻探测已注册")
	}
	if cfg.ProtocolEnabled(observation.ProtocolSecurityTXT) {
		interval := cfg.LightProbe.SecurityTXT.Interval()
		go RunSecurityTXT()
		cs.AddCronJob(interval, RunSecurityTXT)
		log.InfoFields(map[string]interface{}{
			"event":    "light_probe_registered",
			"interval": interval,
			"protocol": observation.ProtocolSecurityTXT,
		}, "security.txt 低频轻探测已注册")
	}
	if cfg.ProtocolEnabled(observation.ProtocolPageAssets) {
		interval := cfg.LightProbe.PageAssets.Interval()
		go RunPageAssets()
		cs.AddCronJob(interval, RunPageAssets)
		log.InfoFields(map[string]interface{}{
			"event":    "light_probe_registered",
			"interval": interval,
			"protocol": observation.ProtocolPageAssets,
		}, "页面资源声明低频轻探测已注册")
	}
	if cfg.ProtocolEnabled(observation.ProtocolPortCheck) {
		interval := cfg.LightProbe.PortCheck.Interval()
		go RunPortCheck()
		cs.AddCronJob(interval, RunPortCheck)
		log.InfoFields(map[string]interface{}{
			"event":    "light_probe_registered",
			"interval": interval,
			"protocol": observation.ProtocolPortCheck,
		}, "端口连通性低频轻探测已注册")
	}
}

func RunRDAP() {
	runLightProbe(observation.ProtocolRDAP, env.GetServerConfig().Collector.V2.LightProbe.RDAP.Interval(), &rdapRunning, runRDAPTargets)
}

func RunRobots() {
	runLightProbe(observation.ProtocolRobots, env.GetServerConfig().Collector.V2.LightProbe.Robots.Interval(), &robotsRunning, runRobotsTargets)
}

func RunSecurityTXT() {
	runLightProbe(observation.ProtocolSecurityTXT, env.GetServerConfig().Collector.V2.LightProbe.SecurityTXT.Interval(), &securityTXTRunning, runSecurityTXTTargets)
}

func RunPageAssets() {
	runLightProbe(observation.ProtocolPageAssets, env.GetServerConfig().Collector.V2.LightProbe.PageAssets.Interval(), &pageAssetsRunning, runPageAssetsTargets)
}

func RunPortCheck() {
	runLightProbe(observation.ProtocolPortCheck, env.GetServerConfig().Collector.V2.LightProbe.PortCheck.Interval(), &portCheckRunning, runPortCheckTargets)
}

func runLightProbe(protocol string, interval time.Duration, running *atomic.Bool, executor func([]models.GfnCollectorDomain, *runstate.Run)) {
	if !env.GetServerConfig().Collector.V2.ProtocolEnabled(protocol) {
		return
	}
	run := runstate.NewRun(protocol, interval)
	if !running.CompareAndSwap(false, true) {
		run.Skip("previous_run_running", 0)
		fields := run.Fields()
		fields["event"] = "run_skipped"
		fields["skipped_reason"] = "previous_run_running"
		fields["status"] = "skipped"
		log.WarnFields(fields, "低频轻探测已跳过：上一轮仍在运行")
		return
	}
	defer running.Store(false)
	if !run.AcquireLeaseOrSkip() {
		fields := run.Fields()
		fields["event"] = "run_skipped"
		fields["skipped_reason"] = "lease_held_by_other_collector"
		fields["status"] = "skipped"
		log.WarnFields(fields, "低频轻探测已跳过：采集 lease 已被其他实例持有")
		return
	}
	defer run.ReleaseLease()
	run.Start()

	start := time.Now()
	targets, err := dao.GetLightProbeDao().GetList()
	if err != nil {
		run.Fail("load_targets", 0)
		fields := run.Fields()
		fields["duration"] = time.Since(start)
		fields["event"] = "run_failed"
		log.ErrorFields(fields, "低频轻探测目标读取失败: "+err.GetMsg())
		return
	}
	if len(targets) == 0 {
		run.Complete(0)
		fields := run.Fields()
		fields["duration"] = time.Since(start)
		fields["event"] = "run_complete"
		fields["targets"] = 0
		log.InfoFields(fields, "低频轻探测完成：没有需要探测的目标")
		return
	}

	run.SetTargetCount(len(targets))
	fields := run.Fields()
	fields["event"] = "run_start"
	fields["targets"] = len(targets)
	log.InfoFields(fields, "低频轻探测运行开始")
	executor(targets, run)
	run.Complete(len(targets))
	snapshot := run.Snapshot(runstate.StatusComplete, "")
	fields = run.Fields()
	fields["duration"] = time.Since(start)
	fields["event"] = "run_complete"
	fields["failure_count"] = snapshot.FailureCount
	fields["success_count"] = snapshot.SuccessCount
	fields["targets"] = len(targets)
	log.InfoFields(fields, "低频轻探测运行完成")
}

func runRDAPTargets(targets []models.GfnCollectorDomain, run *runstate.Run) {
	cfg := env.GetServerConfig().Collector.V2.LightProbe.RDAP
	client := rdapHTTPClient(cfg.Timeout())
	results := map[string]probeResult{}
	for _, target := range targets {
		domain, domainErr := registrableDomain(target.TargetName())
		if domainErr != nil {
			result := failureResult("rdap_no_server", domainErr.Error(), map[string]any{
				"registrable_domain": "",
				"rdap_server":        "",
				"raw_truncated":      false,
			})
			saveLightProbeResult(observation.ProtocolRDAP, target, result, run)
			continue
		}
		result, ok := results[domain]
		if !ok {
			result = probeRDAP(client, domain)
			results[domain] = result
		}
		saveLightProbeResult(observation.ProtocolRDAP, target, result, run)
	}
}

func runRobotsTargets(targets []models.GfnCollectorDomain, run *runstate.Run) {
	cfg := env.GetServerConfig().Collector.V2.LightProbe.Robots
	for _, target := range targets {
		result := probeRobots(target, cfg.Timeout(), cfg.MaxResponseSize(), cfg.MaxSitemaps())
		saveLightProbeResult(observation.ProtocolRobots, target, result, run)
	}
}

func runSecurityTXTTargets(targets []models.GfnCollectorDomain, run *runstate.Run) {
	cfg := env.GetServerConfig().Collector.V2.LightProbe.SecurityTXT
	for _, target := range targets {
		result := probeSecurityTXT(target, cfg.Timeout(), cfg.MaxResponseSize())
		saveLightProbeResult(observation.ProtocolSecurityTXT, target, result, run)
	}
}

func runPageAssetsTargets(targets []models.GfnCollectorDomain, run *runstate.Run) {
	cfg := env.GetServerConfig().Collector.V2.LightProbe.PageAssets
	for _, target := range targets {
		result := probePageAssets(target, cfg)
		saveLightProbeResult(observation.ProtocolPageAssets, target, result, run)
	}
}

func runPortCheckTargets(targets []models.GfnCollectorDomain, run *runstate.Run) {
	cfg := env.GetServerConfig().Collector.V2.LightProbe.PortCheck
	for _, target := range targets {
		result := probePortCheck(target, cfg)
		saveLightProbeResult(observation.ProtocolPortCheck, target, result, run)
	}
}

func saveLightProbeResult(protocol string, target models.GfnCollectorDomain, result probeResult, run *runstate.Run) {
	if result.Status == observation.StatusSuccess {
		run.RecordSuccess()
	} else {
		run.RecordFailure()
	}
	collectorID, jobID := "", ""
	if run != nil {
		collectorID = run.CollectorID
		jobID = run.JobID
	}
	if err := observation.SaveIfEnabled(observation.Input{
		SiteID:       target.SiteID,
		Target:       target.TargetName(),
		Protocol:     protocol,
		Status:       result.Status,
		ObservedAt:   time.Now(),
		DurationMS:   result.DurationMS,
		ErrorCode:    result.ErrorCode,
		ErrorMessage: result.ErrorMessage,
		Payload:      result.Payload,
		CollectorID:  collectorID,
		JobID:        jobID,
	}); err != nil {
		log.WarnFields(map[string]interface{}{
			"event":    "light_probe_observation_write_failed",
			"protocol": protocol,
			"site_id":  target.SiteID,
			"target":   target.TargetName(),
		}, "低频轻探测 observation 旁路写入失败: "+err.GetMsg())
	}
}

func probeRDAP(client *http.Client, domain string) probeResult {
	start := time.Now()
	tld := domainTLD(domain)
	server, err := rdapServerForTLD(client, tld)
	if err != nil {
		code := "rdap_bootstrap_failed"
		if errors.Is(err, errRDAPNoServer) {
			code = "rdap_no_server"
		}
		result := failureResult(code, err.Error(), map[string]any{
			"registrable_domain": domain,
			"rdap_server":        "",
			"raw_truncated":      false,
		})
		result.DurationMS = time.Since(start).Milliseconds()
		return result
	}

	queryURL := strings.TrimRight(server, "/") + "/domain/" + url.PathEscape(domain)
	resp, err := client.Get(queryURL)
	if err != nil {
		result := failureResult("rdap_request_failed", err.Error(), map[string]any{
			"registrable_domain": domain,
			"rdap_server":        server,
			"raw_truncated":      false,
		})
		result.DurationMS = time.Since(start).Milliseconds()
		return result
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	if readErr != nil {
		result := failureResult("rdap_request_failed", readErr.Error(), map[string]any{
			"registrable_domain": domain,
			"rdap_server":        server,
			"status_code":        resp.StatusCode,
			"raw_truncated":      false,
		})
		result.DurationMS = time.Since(start).Milliseconds()
		return result
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		result := failureResult("rdap_request_failed", fmt.Sprintf("RDAP status %d", resp.StatusCode), map[string]any{
			"registrable_domain": domain,
			"rdap_server":        server,
			"status_code":        resp.StatusCode,
			"raw_truncated":      false,
		})
		result.DurationMS = time.Since(start).Milliseconds()
		return result
	}

	payload, err := parseRDAPPayload(body, domain, server, resp.StatusCode)
	if err != nil {
		result := failureResult("rdap_decode_failed", err.Error(), map[string]any{
			"registrable_domain": domain,
			"rdap_server":        server,
			"status_code":        resp.StatusCode,
			"raw_truncated":      false,
		})
		result.DurationMS = time.Since(start).Milliseconds()
		return result
	}
	return probeResult{
		Status:     observation.StatusSuccess,
		DurationMS: time.Since(start).Milliseconds(),
		Payload:    payload,
	}
}

func probeRobots(target models.GfnCollectorDomain, timeout time.Duration, maxBytes int64, maxSitemaps int) probeResult {
	resp, err := probeHTTPGet(target, "/robots.txt", timeout, maxBytes)
	if err != nil {
		return failureResult("robots_request_failed", err.Error(), map[string]any{
			"exists": false,
		})
	}
	payload := parseRobotsPayload(resp.Body, maxSitemaps)
	payload["exists"] = resp.StatusCode >= 200 && resp.StatusCode < 300
	payload["status_code"] = resp.StatusCode
	payload["content_type"] = limitLightText(resp.ContentType)
	payload["body_truncated"] = resp.BodyTruncated
	return probeResult{
		Status:     observation.StatusSuccess,
		DurationMS: resp.DurationMS,
		Payload:    payload,
	}
}

func probeSecurityTXT(target models.GfnCollectorDomain, timeout time.Duration, maxBytes int64) probeResult {
	resp, err := probeHTTPGet(target, "/.well-known/security.txt", timeout, maxBytes)
	pathUsed := "/.well-known/security.txt"
	if err != nil {
		return failureResult("security_txt_request_failed", err.Error(), map[string]any{
			"exists":    false,
			"path_used": pathUsed,
		})
	}
	if resp.StatusCode == http.StatusNotFound {
		fallback, fallbackErr := probeHTTPGet(target, "/security.txt", timeout, maxBytes)
		if fallbackErr != nil {
			return failureResult("security_txt_request_failed", fallbackErr.Error(), map[string]any{
				"exists":    false,
				"path_used": "/security.txt",
			})
		}
		resp = fallback
		pathUsed = "/security.txt"
	}

	payload := parseSecurityTXTPayload(resp.Body)
	payload["exists"] = resp.StatusCode >= 200 && resp.StatusCode < 300
	payload["path_used"] = pathUsed
	payload["status_code"] = resp.StatusCode
	payload["content_type"] = limitLightText(resp.ContentType)
	payload["body_truncated"] = resp.BodyTruncated
	return probeResult{
		Status:     observation.StatusSuccess,
		DurationMS: resp.DurationMS,
		Payload:    payload,
	}
}

func probePageAssets(target models.GfnCollectorDomain, cfg env.LightProbePageAssetsConfig) probeResult {
	start := time.Now()
	raw, err := cs.Get(observation.TargetLatestKey(observation.ProtocolHTTP, target.SiteID, target.TargetName()))
	if err != nil {
		result := failureResult("page_assets_http_latest_read_failed", err.GetMsg(), emptyPageAssetsPayload("http_latest_read_failed"))
		result.DurationMS = time.Since(start).Milliseconds()
		return result
	}
	if raw == "" {
		return probeResult{
			Status:     observation.StatusSuccess,
			DurationMS: time.Since(start).Milliseconds(),
			Payload:    emptyPageAssetsPayload("http_latest_missing"),
		}
	}

	var latest observation.LatestDocument
	if err := sonic.UnmarshalString(raw, &latest); err != nil {
		result := failureResult("page_assets_http_latest_decode_failed", err.Error(), emptyPageAssetsPayload("http_latest_decode_failed"))
		result.DurationMS = time.Since(start).Milliseconds()
		return result
	}
	payloadMap, ok := latest.Payload.(map[string]any)
	if !ok {
		return probeResult{
			Status:     observation.StatusSuccess,
			DurationMS: time.Since(start).Milliseconds(),
			Payload:    emptyPageAssetsPayload("http_latest_payload_not_object"),
		}
	}

	payload := buildPageAssetsPayloadFromHTTPPayload(target, payloadMap, cfg)
	payload["http_latest_found"] = true
	return probeResult{
		Status:     observation.StatusSuccess,
		DurationMS: time.Since(start).Milliseconds(),
		Payload:    payload,
	}
}

func probePortCheck(target models.GfnCollectorDomain, cfg env.LightProbePortCheckConfig) probeResult {
	start := time.Now()
	ports, meta := normalizePortCheckPorts(cfg.Ports, cfg.MaxPorts())
	payload := map[string]any{
		"ports_configured":         len(cfg.Ports),
		"ports_checked":            len(ports),
		"open_count":               0,
		"closed_count":             0,
		"timeout_count":            0,
		"filtered_suspected_count": 0,
		"skipped_count":            meta.SkippedCount(),
		"invalid_port_count":       meta.InvalidCount,
		"duplicate_port_count":     meta.DuplicateCount,
		"truncated_port_count":     meta.TruncatedCount,
		"truncated":                meta.Truncated,
		"results":                  []map[string]any{},
	}
	if len(ports) == 0 {
		payload["skipped_reason"] = "port_list_empty"
		return probeResult{
			Status:     observation.StatusSuccess,
			DurationMS: time.Since(start).Milliseconds(),
			Payload:    payload,
		}
	}

	host := hostnameOnly(target.TargetName())
	results := probePorts(host, ports, cfg.Timeout(), cfg.WorkerCount())
	counts := portCheckCounts(results)
	payload["open_count"] = counts["open"]
	payload["closed_count"] = counts["closed"]
	payload["timeout_count"] = counts["timeout"]
	payload["filtered_suspected_count"] = counts["filtered_suspected"]
	payload["results"] = results
	return probeResult{
		Status:     observation.StatusSuccess,
		DurationMS: time.Since(start).Milliseconds(),
		Payload:    payload,
	}
}

type portCheckNormalizeMeta struct {
	InvalidCount   int
	DuplicateCount int
	TruncatedCount int
	Truncated      bool
}

func (m portCheckNormalizeMeta) SkippedCount() int {
	return m.InvalidCount + m.DuplicateCount + m.TruncatedCount
}

func normalizePortCheckPorts(raw []int, maxPorts int) ([]int, portCheckNormalizeMeta) {
	if maxPorts <= 0 {
		maxPorts = 24
	}
	seen := map[int]bool{}
	ports := make([]int, 0, len(raw))
	meta := portCheckNormalizeMeta{}
	for _, port := range raw {
		if port < 1 || port > 65535 {
			meta.InvalidCount++
			continue
		}
		if seen[port] {
			meta.DuplicateCount++
			continue
		}
		seen[port] = true
		if len(ports) >= maxPorts {
			meta.TruncatedCount++
			meta.Truncated = true
			continue
		}
		ports = append(ports, port)
	}
	sort.Ints(ports)
	return ports, meta
}

func probePorts(host string, ports []int, timeout time.Duration, concurrency int) []map[string]any {
	if concurrency <= 0 {
		concurrency = 8
	}
	if concurrency > len(ports) {
		concurrency = len(ports)
	}
	results := make([]map[string]any, len(ports))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	for idx, port := range ports {
		wg.Add(1)
		sem <- struct{}{}
		go func(index int, p int) {
			defer wg.Done()
			defer func() { <-sem }()
			results[index] = probeSinglePort(host, p, timeout)
		}(idx, port)
	}
	wg.Wait()
	return results
}

func probeSinglePort(host string, port int, timeout time.Duration) map[string]any {
	start := time.Now()
	result := map[string]any{
		"port":          port,
		"service_hint":  serviceHintForPort(port),
		"status":        "filtered_suspected",
		"duration_ms":   int64(0),
		"error_code":    "",
		"error_message": "",
	}
	if host == "" {
		result["status"] = "skipped"
		result["error_code"] = "target_host_empty"
		return result
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	address := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := (&net.Dialer{Timeout: timeout}).DialContext(ctx, "tcp", address)
	result["duration_ms"] = time.Since(start).Milliseconds()
	if err == nil {
		_ = conn.Close()
		result["status"] = "open"
		return result
	}
	status, code := classifyPortCheckError(err)
	result["status"] = status
	result["error_code"] = code
	result["error_message"] = limitLightText(err.Error())
	return result
}

func classifyPortCheckError(err error) (string, string) {
	if err == nil {
		return "open", ""
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout", "timeout"
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "timeout", "timeout"
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return "filtered_suspected", "dns_failed"
	}
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "connection refused") || strings.Contains(message, "actively refused") || strings.Contains(message, "connectex: no connection"):
		return "closed", "connection_refused"
	case strings.Contains(message, "i/o timeout"):
		return "timeout", "timeout"
	case strings.Contains(message, "no route to host") || strings.Contains(message, "network is unreachable") || strings.Contains(message, "host is unreachable"):
		return "filtered_suspected", "network_unreachable"
	default:
		return "filtered_suspected", "connect_failed"
	}
}

func portCheckCounts(results []map[string]any) map[string]int {
	counts := map[string]int{
		"open":               0,
		"closed":             0,
		"timeout":            0,
		"filtered_suspected": 0,
	}
	for _, result := range results {
		status, _ := result["status"].(string)
		if _, ok := counts[status]; ok {
			counts[status]++
		}
	}
	return counts
}

func serviceHintForPort(port int) string {
	switch port {
	case 22:
		return "ssh"
	case 25:
		return "smtp"
	case 53:
		return "dns"
	case 80:
		return "http"
	case 443:
		return "https"
	case 465:
		return "smtps"
	case 587:
		return "submission"
	case 993:
		return "imaps"
	case 995:
		return "pop3s"
	case 3000:
		return "grafana"
	case 3306:
		return "mysql"
	case 5432:
		return "postgresql"
	case 6379:
		return "redis"
	case 8080:
		return "http_alt"
	case 8443:
		return "https_alt"
	case 9090:
		return "prometheus"
	case 9092:
		return "kafka"
	case 27017:
		return "mongodb"
	default:
		return ""
	}
}

func probeHTTPGet(target models.GfnCollectorDomain, path string, timeout time.Duration, maxBytes int64) (httpProbeResponse, error) {
	client, err := httpClientForTarget(timeout, target.Proxy == "1")
	if err != nil {
		return httpProbeResponse{}, err
	}
	return probeHTTPGetURL(client, targetURL(target, path), maxBytes)
}

func probeHTTPGetURL(client *http.Client, rawURL string, maxBytes int64) (httpProbeResponse, error) {
	start := time.Now()
	resp, err := client.Get(rawURL)
	if err != nil {
		return httpProbeResponse{}, err
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if readErr != nil {
		return httpProbeResponse{}, readErr
	}
	truncated := int64(len(body)) > maxBytes
	if truncated {
		body = body[:maxBytes]
	}
	return httpProbeResponse{
		StatusCode:          resp.StatusCode,
		ContentType:         resp.Header.Get("Content-Type"),
		ContentLengthHeader: resp.ContentLength,
		Body:                body,
		BodyTruncated:       truncated,
		DurationMS:          time.Since(start).Milliseconds(),
	}, nil
}

func buildPageAssetsPayloadFromHTTPPayload(target models.GfnCollectorDomain, httpPayload map[string]any, cfg env.LightProbePageAssetsConfig) map[string]any {
	iconDecl := selectPageAssetIcon(httpPayload)
	manifestDecl := selectPageAssetManifest(httpPayload)
	return map[string]any{
		"icon":     probeDeclaredAsset(target, iconDecl, cfg.Timeout(), cfg.MaxIconSize(), cfg.AllowedAssetHosts, isAllowedIconContentType, nil),
		"manifest": probeDeclaredAsset(target, manifestDecl, cfg.Timeout(), cfg.MaxManifestSize(), cfg.AllowedAssetHosts, isAllowedManifestContentType, parseManifestSummary),
	}
}

func emptyPageAssetsPayload(reason string) map[string]any {
	return map[string]any{
		"http_latest_found": false,
		"icon": map[string]any{
			"exists":         false,
			"skipped_reason": reason,
		},
		"manifest": map[string]any{
			"exists":         false,
			"skipped_reason": reason,
		},
	}
}

func probeDeclaredAsset(target models.GfnCollectorDomain, declaration pageAssetDeclaration, timeout time.Duration, maxBytes int64, allowedHosts []string, contentTypeAllowed func(string) bool, parser func([]byte, string) map[string]any) map[string]any {
	payload := map[string]any{
		"exists":         false,
		"source_url":     limitLightURL(declaration.Href),
		"selected_rel":   limitLightText(declaration.Rel),
		"selected_sizes": limitLightText(declaration.Sizes),
	}
	if declaration.Href == "" {
		payload["skipped_reason"] = "asset_link_missing"
		return payload
	}
	if !assetURLAllowed(target.TargetName(), declaration.Href, allowedHosts) {
		payload["skipped_reason"] = "asset_host_not_allowed"
		return payload
	}

	client, err := httpClientForTarget(timeout, target.Proxy == "1")
	if err != nil {
		payload["skipped_reason"] = "http_client_config_invalid"
		payload["error_message"] = limitLightText(err.Error())
		return payload
	}
	resp, err := probeHTTPGetURL(client, declaration.Href, maxBytes)
	if err != nil {
		payload["skipped_reason"] = "asset_request_failed"
		payload["error_message"] = limitLightText(err.Error())
		return payload
	}

	payload["status_code"] = resp.StatusCode
	payload["content_type"] = limitLightText(resp.ContentType)
	payload["content_length_header"] = resp.ContentLengthHeader
	payload["body_read_bytes"] = len(resp.Body)
	payload["body_truncated"] = resp.BodyTruncated
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		payload["skipped_reason"] = "asset_not_success"
		return payload
	}
	if !contentTypeAllowed(resp.ContentType) {
		payload["skipped_reason"] = "content_type_not_allowed"
		return payload
	}
	payload["exists"] = true
	payload["sha256"] = sha256Hex(resp.Body)
	if parser != nil {
		for key, value := range parser(resp.Body, declaration.Href) {
			payload[key] = value
		}
	}
	return payload
}

func selectPageAssetIcon(payload map[string]any) pageAssetDeclaration {
	for _, item := range pageAssetLinksFromPayload(payload["icon_links"]) {
		rel := strings.ToLower(item.Rel)
		if strings.Contains(rel, "manifest") {
			continue
		}
		if strings.Contains(rel, "apple-touch-icon") {
			return item
		}
	}
	for _, item := range pageAssetLinksFromPayload(payload["icon_links"]) {
		rel := strings.ToLower(item.Rel)
		if strings.Contains(rel, "manifest") {
			continue
		}
		if strings.Contains(rel, "icon") || strings.Contains(rel, "shortcut") {
			return item
		}
	}
	return pageAssetDeclaration{}
}

func selectPageAssetManifest(payload map[string]any) pageAssetDeclaration {
	if manifest := pageAssetLinkFromPayload(payload["manifest_link"]); manifest.Href != "" {
		return manifest
	}
	for _, item := range pageAssetLinksFromPayload(payload["icon_links"]) {
		if strings.Contains(strings.ToLower(item.Rel), "manifest") {
			return item
		}
	}
	return pageAssetDeclaration{}
}

func pageAssetLinksFromPayload(raw any) []pageAssetDeclaration {
	values, ok := raw.([]any)
	if !ok {
		return nil
	}
	links := make([]pageAssetDeclaration, 0, len(values))
	for _, value := range values {
		link := pageAssetLinkFromPayload(value)
		if link.Href != "" {
			links = append(links, link)
		}
	}
	return links
}

func pageAssetLinkFromPayload(raw any) pageAssetDeclaration {
	values, ok := raw.(map[string]any)
	if !ok {
		return pageAssetDeclaration{}
	}
	return pageAssetDeclaration{
		Rel:   stringFromAny(values["rel"]),
		Href:  stringFromAny(values["href"]),
		Type:  stringFromAny(values["type"]),
		Sizes: stringFromAny(values["sizes"]),
	}
}

func assetURLAllowed(target string, rawURL string, allowedHosts []string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil || !parsed.IsAbs() {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	assetHost := strings.ToLower(parsed.Hostname())
	if assetHost == "" {
		return false
	}
	targetHost := strings.ToLower(hostnameOnly(target))
	if assetHost == targetHost {
		return true
	}
	for _, host := range allowedHosts {
		if assetHost == strings.ToLower(hostnameOnly(host)) {
			return true
		}
	}
	return sameRegistrableDomain(assetHost, targetHost)
}

func sameRegistrableDomain(first string, second string) bool {
	if first == "" || second == "" || net.ParseIP(first) != nil || net.ParseIP(second) != nil {
		return false
	}
	firstDomain, firstErr := publicsuffix.EffectiveTLDPlusOne(first)
	secondDomain, secondErr := publicsuffix.EffectiveTLDPlusOne(second)
	return firstErr == nil && secondErr == nil && firstDomain == secondDomain
}

func hostnameOnly(value string) string {
	value = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(value, "https://"), "http://"))
	if host, _, err := net.SplitHostPort(value); err == nil {
		return host
	}
	if parsed, err := url.Parse(value); err == nil && parsed.Hostname() != "" {
		return parsed.Hostname()
	}
	return strings.Trim(value, "[]")
}

func isAllowedIconContentType(value string) bool {
	contentType := strings.ToLower(strings.TrimSpace(strings.Split(value, ";")[0]))
	switch contentType {
	case "image/x-icon", "image/vnd.microsoft.icon", "image/png", "image/jpeg", "image/gif", "image/webp", "image/svg+xml":
		return true
	default:
		return false
	}
}

func isAllowedManifestContentType(value string) bool {
	contentType := strings.ToLower(strings.TrimSpace(strings.Split(value, ";")[0]))
	switch contentType {
	case "application/manifest+json", "application/json", "text/json":
		return true
	default:
		return false
	}
}

func parseManifestSummary(body []byte, sourceURL string) map[string]any {
	var parsed struct {
		Name            string `json:"name"`
		ShortName       string `json:"short_name"`
		ThemeColor      string `json:"theme_color"`
		BackgroundColor string `json:"background_color"`
		Display         string `json:"display"`
		StartURL        string `json:"start_url"`
		Scope           string `json:"scope"`
		Icons           []any  `json:"icons"`
	}
	if err := sonic.Unmarshal(body, &parsed); err != nil {
		return map[string]any{
			"manifest_decode_error": limitLightText(err.Error()),
		}
	}
	return map[string]any{
		"name":             limitLightText(parsed.Name),
		"short_name":       limitLightText(parsed.ShortName),
		"theme_color":      limitLightText(parsed.ThemeColor),
		"background_color": limitLightText(parsed.BackgroundColor),
		"display":          limitLightText(parsed.Display),
		"start_url":        limitLightURL(resolveManifestURL(parsed.StartURL, sourceURL)),
		"scope":            limitLightURL(resolveManifestURL(parsed.Scope, sourceURL)),
		"icons_count":      len(parsed.Icons),
	}
}

func resolveManifestURL(value string, baseURL string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return value
	}
	if parsed.IsAbs() {
		return parsed.String()
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return value
	}
	return base.ResolveReference(parsed).String()
}

func sha256Hex(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func stringFromAny(value any) string {
	if raw, ok := value.(string); ok {
		return raw
	}
	return ""
}

func limitLightURL(value string) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= 2048 {
		return string(runes)
	}
	return string(runes[:2048])
}

func httpClientForTarget(timeout time.Duration, useProxy bool) (*http.Client, error) {
	proxyRaw := ""
	if useProxy {
		proxyRaw = env.GetServerConfig().Collector.Proxy
	}
	return httpClientWithError(timeout, proxyRaw, useProxy)
}

func rdapHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= lightProbeRedirects {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
}

func httpClientWithError(timeout time.Duration, proxyRaw string, useProxy bool) (*http.Client, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	if useProxy {
		proxyURL, err := url.Parse(proxyRaw)
		if err != nil || proxyURL.Scheme == "" || proxyURL.Host == "" {
			return nil, fmt.Errorf("代理地址无效")
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= lightProbeRedirects {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}, nil
}

func targetURL(target models.GfnCollectorDomain, path string) string {
	scheme := "http"
	if target.TLS == "1" {
		scheme = "https"
	}
	return scheme + "://" + target.TargetName() + path
}

func failureResult(code string, message string, payload map[string]any) probeResult {
	if payload == nil {
		payload = map[string]any{}
	}
	payload["error_code"] = code
	return probeResult{
		Status:       observation.StatusFailure,
		ErrorCode:    code,
		ErrorMessage: message,
		Payload:      payload,
	}
}

func parseRobotsPayload(body []byte, maxSitemaps int) map[string]any {
	sitemaps := []string{}
	sitemapCount := 0
	userAgentStarPresent := false
	globalDisallowAll := false
	inStarBlock := false
	scanner := bufio.NewScanner(bytes.NewReader(body))
	scanner.Buffer(make([]byte, 1024), 128*1024)
	for scanner.Scan() {
		key, value, ok := parseLightProbeLine(scanner.Text())
		if !ok {
			continue
		}
		switch strings.ToLower(key) {
		case "user-agent":
			inStarBlock = strings.TrimSpace(value) == "*"
			if inStarBlock {
				userAgentStarPresent = true
			}
		case "disallow":
			if inStarBlock && strings.TrimSpace(value) == "/" {
				globalDisallowAll = true
			}
		case "sitemap":
			sitemapCount++
			if len(sitemaps) < maxSitemaps {
				sitemaps = append(sitemaps, limitLightText(value))
			}
		}
	}
	return map[string]any{
		"sitemap_count":           sitemapCount,
		"sitemaps":                sitemaps,
		"global_disallow_all":     globalDisallowAll,
		"user_agent_star_present": userAgentStarPresent,
	}
}

func parseSecurityTXTPayload(body []byte) map[string]any {
	contacts := []string{}
	policies := []string{}
	preferredLanguages := []string{}
	canonicals := []string{}
	expires := ""
	scanner := bufio.NewScanner(bytes.NewReader(body))
	scanner.Buffer(make([]byte, 1024), 128*1024)
	for scanner.Scan() {
		key, value, ok := parseLightProbeLine(scanner.Text())
		if !ok {
			continue
		}
		switch strings.ToLower(key) {
		case "contact":
			contacts = appendLimitedLightItem(contacts, value)
		case "expires":
			if expires == "" {
				expires = limitLightText(value)
			}
		case "policy":
			policies = appendLimitedLightItem(policies, value)
		case "preferred-languages":
			preferredLanguages = appendLimitedLightItem(preferredLanguages, value)
		case "canonical":
			canonicals = appendLimitedLightItem(canonicals, value)
		}
	}
	return map[string]any{
		"contact":             contacts,
		"expires":             expires,
		"policy":              policies,
		"preferred_languages": preferredLanguages,
		"canonical":           canonicals,
	}
}

func parseLightProbeLine(line string) (string, string, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false
	}
	if idx := strings.Index(line, "#"); idx >= 0 {
		line = strings.TrimSpace(line[:idx])
	}
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	if key == "" || value == "" {
		return "", "", false
	}
	return key, value, true
}

func appendLimitedLightItem(values []string, value string) []string {
	if len(values) >= lightProbeMaxItems {
		return values
	}
	return append(values, limitLightText(value))
}

func limitLightText(value string) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= lightProbeMaxTextLength {
		return string(runes)
	}
	return string(runes[:lightProbeMaxTextLength])
}

func registrableDomain(target string) (string, error) {
	host := strings.Trim(strings.ToLower(strings.TrimSpace(target)), ".")
	if splitHost, _, err := net.SplitHostPort(host); err == nil {
		host = splitHost
	}
	if host == "" || net.ParseIP(host) != nil {
		return "", fmt.Errorf("目标不是可查询 RDAP 的域名: %s", target)
	}
	domain, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		return "", err
	}
	return domain, nil
}

func domainTLD(domain string) string {
	domain = strings.Trim(domain, ".")
	idx := strings.LastIndex(domain, ".")
	if idx < 0 || idx == len(domain)-1 {
		return domain
	}
	return domain[idx+1:]
}

var errRDAPNoServer = errors.New("未找到 TLD 对应的 RDAP 服务")

func rdapServerForTLD(client *http.Client, tld string) (string, error) {
	servers, err := rdapBootstrapServers(client)
	if err != nil {
		return "", err
	}
	server := servers[strings.ToLower(strings.Trim(tld, "."))]
	if server == "" {
		return "", errRDAPNoServer
	}
	return server, nil
}

func rdapBootstrapServers(client *http.Client) (map[string]string, error) {
	rdapBootstrapMu.Lock()
	defer rdapBootstrapMu.Unlock()
	if rdapBootstrap.servers != nil && time.Now().Before(rdapBootstrap.expiresAt) {
		return rdapBootstrap.servers, nil
	}
	timeout := client.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rdapBootstrapURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("IANA RDAP bootstrap status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return nil, err
	}
	servers, err := parseRDAPBootstrap(body)
	if err != nil {
		return nil, err
	}
	rdapBootstrap = cachedRDAPBootstrap{
		expiresAt: time.Now().Add(rdapBootstrapCacheTTL),
		servers:   servers,
	}
	return servers, nil
}

func parseRDAPBootstrap(body []byte) (map[string]string, error) {
	var parsed struct {
		Services [][][]string `json:"services"`
	}
	if err := sonic.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	servers := map[string]string{}
	for _, service := range parsed.Services {
		if len(service) < 2 || len(service[1]) == 0 {
			continue
		}
		server := service[1][0]
		for _, tld := range service[0] {
			tld = strings.ToLower(strings.Trim(tld, "."))
			if tld != "" && server != "" {
				servers[tld] = server
			}
		}
	}
	if len(servers) == 0 {
		return nil, errors.New("RDAP bootstrap 未包含可用服务")
	}
	return servers, nil
}

func parseRDAPPayload(body []byte, domain string, server string, statusCode int) (map[string]any, error) {
	var parsed rdapDomainResponse
	if err := sonic.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	return map[string]any{
		"registrable_domain":       domain,
		"rdap_server":              server,
		"status_code":              statusCode,
		"registrar":                limitLightText(parsed.Registrar()),
		"statuses":                 limitLightItems(parsed.Status),
		"expires_at":               limitLightText(parsed.EventDate("expiration")),
		"nameservers":              limitLightItems(parsed.NameServers()),
		"dnssec_delegation_signed": parsed.SecureDNS != nil && parsed.SecureDNS.DelegationSigned,
		"events_summary":           parsed.EventsSummary(),
		"raw_truncated":            false,
	}, nil
}

type rdapDomainResponse struct {
	Status      []string       `json:"status"`
	Events      []rdapEvent    `json:"events"`
	Nameservers []rdapNS       `json:"nameservers"`
	SecureDNS   *rdapSecureDNS `json:"secureDNS"`
	Entities    []rdapEntity   `json:"entities"`
}

type rdapEvent struct {
	EventAction string `json:"eventAction"`
	EventDate   string `json:"eventDate"`
}

type rdapNS struct {
	LDHName string `json:"ldhName"`
}

type rdapSecureDNS struct {
	DelegationSigned bool `json:"delegationSigned"`
}

type rdapEntity struct {
	Roles      []string `json:"roles"`
	VCardArray []any    `json:"vcardArray"`
}

func (r rdapDomainResponse) Registrar() string {
	for _, entity := range r.Entities {
		if !containsFold(entity.Roles, "registrar") {
			continue
		}
		if name := rdapVCardFN(entity.VCardArray); name != "" {
			return name
		}
	}
	return ""
}

func (r rdapDomainResponse) EventDate(action string) string {
	for _, event := range r.Events {
		if strings.EqualFold(event.EventAction, action) {
			return event.EventDate
		}
	}
	return ""
}

func (r rdapDomainResponse) EventsSummary() map[string]string {
	summary := map[string]string{}
	for _, event := range r.Events {
		action := limitLightText(event.EventAction)
		if action != "" && summary[action] == "" {
			summary[action] = limitLightText(event.EventDate)
		}
	}
	return summary
}

func (r rdapDomainResponse) NameServers() []string {
	names := make([]string, 0, len(r.Nameservers))
	for _, ns := range r.Nameservers {
		if ns.LDHName != "" {
			names = append(names, ns.LDHName)
		}
	}
	sort.Strings(names)
	return names
}

func rdapVCardFN(vcard []any) string {
	if len(vcard) < 2 {
		return ""
	}
	props, ok := vcard[1].([]any)
	if !ok {
		return ""
	}
	for _, prop := range props {
		items, ok := prop.([]any)
		if !ok || len(items) < 4 {
			continue
		}
		name, _ := items[0].(string)
		if !strings.EqualFold(name, "fn") {
			continue
		}
		value, _ := items[3].(string)
		return value
	}
	return ""
}

func containsFold(values []string, expected string) bool {
	for _, value := range values {
		if strings.EqualFold(value, expected) {
			return true
		}
	}
	return false
}

func limitLightItems(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	if len(values) > lightProbeMaxItems {
		values = values[:lightProbeMaxItems]
	}
	limited := make([]string, 0, len(values))
	for _, value := range values {
		limited = append(limited, limitLightText(value))
	}
	return limited
}
