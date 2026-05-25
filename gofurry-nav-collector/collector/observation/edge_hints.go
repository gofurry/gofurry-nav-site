package observation

import (
	"fmt"
	"sort"
	"strings"
)

const edgeEvidenceValueLimit = 256

type edgeHintCollector struct {
	hints map[string]*edgeHintBuilder
}

type edgeHintBuilder struct {
	provider   string
	hintType   string
	maxScore   int
	evidence   []EdgeProviderEvidence
	evidenceID map[string]struct{}
	sources    map[string]struct{}
}

// BuildEdgeProviderHints returns conservative CDN/WAF/hosting hints from existing v2 latest documents.
func BuildEdgeProviderHints(docs map[string]LatestDocument) []EdgeProviderHint {
	collector := edgeHintCollector{hints: map[string]*edgeHintBuilder{}}
	if doc, ok := docs[ProtocolHTTP]; ok {
		collector.collectHTTP(doc.Payload)
	}
	if doc, ok := docs[ProtocolDNS]; ok {
		collector.collectDNS(doc.Payload)
	}
	return collector.result()
}

func (c *edgeHintCollector) collectHTTP(payload any) {
	payloadMap := mapFromAny(payload)
	if len(payloadMap) == 0 {
		return
	}
	for key, value := range mapFromAny(payloadMap["headers"]) {
		field := "headers." + strings.ToLower(key)
		c.matchHTTPHeader(strings.ToLower(key), stringFromAny(value), field)
	}
	for _, key := range []string{"server", "final_url", "cert_issuer", "cert_issuer_cn"} {
		c.matchGeneralString("http", key, stringFromAny(payloadMap[key]), 1)
	}
	for _, key := range []string{"cert_dns_names", "cert_subject_org", "cert_chain_issuers"} {
		for _, value := range stringSliceFromAny(payloadMap[key]) {
			c.matchGeneralString("tls", key, value, 1)
		}
	}
	for key, value := range mapFromAny(payloadMap["server_hints"]) {
		c.matchGeneralString("http", "server_hints."+strings.ToLower(key), stringFromAny(value), 1)
	}
}

func (c *edgeHintCollector) matchHTTPHeader(key string, value string, field string) {
	lowerValue := strings.ToLower(value)
	switch {
	case key == "cf-ray" || key == "cf-cache-status":
		c.add("cloudflare", "cdn", 3, "http", field, value)
	case key == "server" && strings.Contains(lowerValue, "cloudflare"):
		c.add("cloudflare", "cdn", 2, "http", field, value)
	case key == "x-vercel-id" || key == "x-vercel-cache":
		c.add("vercel", "hosting_platform", 3, "http", field, value)
	case key == "x-nf-request-id":
		c.add("netlify", "hosting_platform", 3, "http", field, value)
	case strings.HasPrefix(key, "x-amz-cf-") || strings.Contains(lowerValue, "cloudfront"):
		c.add("aws_cloudfront", "cdn", 3, "http", field, value)
	case key == "x-fastly-request-id" || strings.Contains(lowerValue, "fastly"):
		c.add("fastly", "cdn", 3, "http", field, value)
	case key == "x-github-request-id" || (key == "server" && strings.Contains(lowerValue, "github")):
		c.add("github_pages", "hosting_platform", 2, "http", field, value)
	case key == "x-oss-request-id":
		c.add("aliyun", "object_storage", 3, "http", field, value)
	case key == "eagleid" || key == "x-swift-cachetime" || key == "x-acs-request-id":
		c.add("aliyun", "cdn", 2, "http", field, value)
	case key == "x-nws-log-uuid" || strings.Contains(lowerValue, "tencent"):
		c.add("tencent_cloud", "cdn", 2, "http", field, value)
	}
}

func (c *edgeHintCollector) collectDNS(payload any) {
	payloadMap := mapFromAny(payload)
	if len(payloadMap) == 0 {
		return
	}
	for _, key := range []string{"cname_terminal", "name_server_hosts", "mx_hosts"} {
		for _, value := range valuesFromAny(payloadMap[key]) {
			c.matchGeneralString("dns", key, value, 1)
		}
	}
	for _, recordType := range []string{"A", "AAAA", "CNAME", "NS"} {
		for _, item := range listFromAny(payloadMap[recordType]) {
			c.collectDNSRecord(recordType, mapFromAny(item))
		}
	}
}

func (c *edgeHintCollector) collectDNSRecord(recordType string, record map[string]any) {
	for _, key := range []string{"value", "asn", "isp", "provider_type"} {
		c.matchGeneralString("dns", recordType+"."+key, stringFromAny(record[key]), 1)
	}
	for _, child := range listFromAny(record["children"]) {
		c.collectDNSRecord(recordType+".children", mapFromAny(child))
	}
}

func (c *edgeHintCollector) matchGeneralString(source string, field string, value string, score int) {
	lower := strings.ToLower(value)
	if lower == "" {
		return
	}
	switch {
	case strings.Contains(lower, "cloudflare"):
		c.add("cloudflare", "cdn", score, source, field, value)
	case strings.Contains(lower, "cloudfront") || strings.Contains(lower, "amazon") || strings.Contains(lower, "aws"):
		c.add("aws_cloudfront", "cdn", score, source, field, value)
	case strings.Contains(lower, "fastly"):
		c.add("fastly", "cdn", score, source, field, value)
	case strings.Contains(lower, "vercel"):
		c.add("vercel", "hosting_platform", score, source, field, value)
	case strings.Contains(lower, "netlify"):
		c.add("netlify", "hosting_platform", score, source, field, value)
	case strings.Contains(lower, "github.io") || strings.Contains(lower, "github pages"):
		c.add("github_pages", "hosting_platform", score, source, field, value)
	case strings.Contains(lower, "aliyun") || strings.Contains(lower, "alibaba") || strings.Contains(lower, "alidns") || strings.Contains(lower, "hichina"):
		c.add("aliyun", "cdn", score, source, field, value)
	case strings.Contains(lower, "tencent") || strings.Contains(lower, "qcloud") || strings.Contains(lower, "dnsv1.com"):
		c.add("tencent_cloud", "cdn", score, source, field, value)
	}
}

func (c *edgeHintCollector) add(provider string, hintType string, score int, source string, field string, value string) {
	if provider == "" || hintType == "" || value == "" {
		return
	}
	if score <= 0 {
		score = 1
	}
	key := provider + "|" + hintType
	builder := c.hints[key]
	if builder == nil {
		builder = &edgeHintBuilder{
			provider:   provider,
			hintType:   hintType,
			evidenceID: map[string]struct{}{},
			sources:    map[string]struct{}{},
		}
		c.hints[key] = builder
	}
	if score > builder.maxScore {
		builder.maxScore = score
	}
	builder.sources[source] = struct{}{}
	evidence := EdgeProviderEvidence{
		Source: source,
		Field:  field,
		Value:  limitEdgeEvidenceValue(value),
	}
	id := fmt.Sprintf("%s|%s|%s", evidence.Source, evidence.Field, strings.ToLower(evidence.Value))
	if _, ok := builder.evidenceID[id]; ok {
		return
	}
	builder.evidenceID[id] = struct{}{}
	builder.evidence = append(builder.evidence, evidence)
}

func (c *edgeHintCollector) result() []EdgeProviderHint {
	if len(c.hints) == 0 {
		return nil
	}
	keys := make([]string, 0, len(c.hints))
	for key := range c.hints {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	hints := make([]EdgeProviderHint, 0, len(keys))
	for _, key := range keys {
		builder := c.hints[key]
		hints = append(hints, EdgeProviderHint{
			Provider:   builder.provider,
			HintType:   builder.hintType,
			Confidence: builder.confidence(),
			Evidence:   builder.evidence,
		})
	}
	return hints
}

func (b *edgeHintBuilder) confidence() string {
	if b.maxScore >= 3 {
		return "high"
	}
	if b.maxScore >= 2 || len(b.sources) >= 2 || len(b.evidence) >= 2 {
		return "medium"
	}
	return "low"
}

func mapFromAny(value any) map[string]any {
	if values, ok := value.(map[string]any); ok {
		return values
	}
	if values, ok := value.(map[string]string); ok {
		result := make(map[string]any, len(values))
		for key, item := range values {
			result[key] = item
		}
		return result
	}
	return nil
}

func listFromAny(value any) []any {
	if values, ok := value.([]any); ok {
		return values
	}
	return nil
}

func valuesFromAny(value any) []string {
	if single := stringFromAny(value); single != "" {
		return []string{single}
	}
	return stringSliceFromAny(value)
}

func stringFromAny(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	default:
		return ""
	}
}

func stringSliceFromAny(value any) []string {
	switch values := value.(type) {
	case []string:
		return values
	case []any:
		result := make([]string, 0, len(values))
		for _, item := range values {
			if text := stringFromAny(item); text != "" {
				result = append(result, text)
			}
		}
		return result
	default:
		return nil
	}
}

func limitEdgeEvidenceValue(value string) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= edgeEvidenceValueLimit {
		return string(runes)
	}
	return string(runes[:edgeEvidenceValueLimit])
}
