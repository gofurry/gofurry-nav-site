package observation

const (
	ReasonSeverityInfo     = "info"
	ReasonSeverityWarning  = "warning"
	ReasonSeverityDegraded = "degraded"
	ReasonSeverityDown     = "down"
	ReasonSeverityUnknown  = "unknown"

	ReasonScopeTarget = "target"
	ReasonScopeSite   = "site"
)

type ReasonDefinition struct {
	Code          string
	MessageZH     string
	Severity      string
	Scope         string
	AffectsHealth bool
	DescriptionZH string
}

var reasonDefinitions = []ReasonDefinition{
	{Code: "http_missing_or_stale", MessageZH: "HTTP 观测缺失或已过期，无法判断访客是否可打开", Severity: ReasonSeverityUnknown, Scope: ReasonScopeTarget, AffectsHealth: true, DescriptionZH: "缺少可用 HTTP latest，或 HTTP latest 已超过协议过期窗口。"},
	{Code: "http_failed", MessageZH: "HTTP 访问失败", Severity: ReasonSeverityDegraded, Scope: ReasonScopeTarget, AffectsHealth: true, DescriptionZH: "最新 HTTP observation 为失败，访客访问链路存在明确问题。"},
	{Code: "dns_failed", MessageZH: "DNS 解析也失败", Severity: ReasonSeverityDown, Scope: ReasonScopeTarget, AffectsHealth: true, DescriptionZH: "HTTP 失败且 DNS latest 也失败，目标被判定为 down 的核心依据之一。"},
	{Code: "dns_missing_or_stale", MessageZH: "DNS 观测缺失或已过期", Severity: ReasonSeverityWarning, Scope: ReasonScopeTarget, AffectsHealth: true, DescriptionZH: "DNS latest 缺失或超过协议过期窗口，当前 DNS 信息不足以辅助判断。"},
	{Code: "dns_failed_but_http_ok", MessageZH: "DNS 失败但 HTTP 当前仍可访问", Severity: ReasonSeverityWarning, Scope: ReasonScopeTarget, AffectsHealth: true, DescriptionZH: "DNS latest 失败，但 HTTP latest 成功，说明解析链路存在需要关注的观测信号。"},
	{Code: "ping_failed_but_http_ok", MessageZH: "Ping 失败但 HTTP 当前仍可访问", Severity: ReasonSeverityWarning, Scope: ReasonScopeTarget, AffectsHealth: true, DescriptionZH: "Ping latest 失败，但 HTTP latest 成功；通常只表示 ICMP 不通或被限制，不单独判定站点不可访问。"},
	{Code: "dns_risk_private_ip", MessageZH: "DNS observation 出现风险信号: private_ip", Severity: ReasonSeverityWarning, Scope: ReasonScopeTarget, AffectsHealth: true, DescriptionZH: "DNS v2 risk_flags 中出现 private_ip，表示解析结果含私网或特殊地址，需要结合部署环境人工确认。"},
	{Code: "dns_risk_low_ttl", MessageZH: "DNS observation 出现风险信号: low_ttl", Severity: ReasonSeverityWarning, Scope: ReasonScopeTarget, AffectsHealth: true, DescriptionZH: "DNS v2 risk_flags 中出现 low_ttl，表示 TTL 很低，可能是正常调度策略，也可能需要关注解析稳定性。"},
	{Code: "dns_risk_nxdomain_with_answer", MessageZH: "DNS observation 出现风险信号: nxdomain_with_answer", Severity: ReasonSeverityWarning, Scope: ReasonScopeTarget, AffectsHealth: true, DescriptionZH: "DNS v2 risk_flags 中出现 nxdomain_with_answer，表示响应语义不常见，需要人工确认 resolver 或权威响应行为。"},
	{Code: "dns_risk_ptr_empty", MessageZH: "DNS observation 出现风险信号: ptr_empty", Severity: ReasonSeverityWarning, Scope: ReasonScopeTarget, AffectsHealth: true, DescriptionZH: "DNS v2 risk_flags 中出现 ptr_empty，表示反向解析为空；这是治理参考，不等同于故障。"},
	{Code: "dns_risk_other", MessageZH: "DNS observation 出现未知风险信号", Severity: ReasonSeverityWarning, Scope: ReasonScopeTarget, AffectsHealth: true, DescriptionZH: "DNS v2 risk_flags 出现当前字典未细分的风险标记，统一收敛到 other，避免 reason code 无限扩张。"},
	{Code: "tls_verify_expired", MessageZH: "TLS 证书校验未通过: expired", Severity: ReasonSeverityDegraded, Scope: ReasonScopeTarget, AffectsHealth: true, DescriptionZH: "TLS 证书链或域名校验失败，分类为证书过期。"},
	{Code: "tls_verify_not_yet_valid", MessageZH: "TLS 证书校验未通过: not_yet_valid", Severity: ReasonSeverityDegraded, Scope: ReasonScopeTarget, AffectsHealth: true, DescriptionZH: "TLS 证书链或域名校验失败，分类为证书尚未生效。"},
	{Code: "tls_verify_hostname_mismatch", MessageZH: "TLS 证书校验未通过: hostname_mismatch", Severity: ReasonSeverityDegraded, Scope: ReasonScopeTarget, AffectsHealth: true, DescriptionZH: "TLS 证书链或域名校验失败，分类为证书域名不匹配。"},
	{Code: "tls_verify_unknown_authority", MessageZH: "TLS 证书校验未通过: unknown_authority", Severity: ReasonSeverityDegraded, Scope: ReasonScopeTarget, AffectsHealth: true, DescriptionZH: "TLS 证书链或域名校验失败，分类为未知签发机构。"},
	{Code: "tls_verify_incompatible_usage", MessageZH: "TLS 证书校验未通过: incompatible_usage", Severity: ReasonSeverityDegraded, Scope: ReasonScopeTarget, AffectsHealth: true, DescriptionZH: "TLS 证书链或域名校验失败，分类为证书用途不兼容。"},
	{Code: "tls_verify_other", MessageZH: "TLS 证书校验未通过: other", Severity: ReasonSeverityDegraded, Scope: ReasonScopeTarget, AffectsHealth: true, DescriptionZH: "TLS 证书链或域名校验失败，但无法归入更具体分类。"},
	{Code: "tls_cert_expired", MessageZH: "TLS 证书已过期", Severity: ReasonSeverityDegraded, Scope: ReasonScopeTarget, AffectsHealth: true, DescriptionZH: "根据证书 not_after 判断，证书已经过期。"},
	{Code: "tls_cert_expiring_soon", MessageZH: "TLS 证书将在 30 天内过期", Severity: ReasonSeverityWarning, Scope: ReasonScopeTarget, AffectsHealth: true, DescriptionZH: "根据证书 not_after 判断，证书将在 30 天内过期。"},
	{Code: "no_target_summary", MessageZH: "没有可用的采集目标健康摘要", Severity: ReasonSeverityUnknown, Scope: ReasonScopeSite, AffectsHealth: true, DescriptionZH: "站点没有可用于聚合的 target summary。"},
	{Code: "all_targets_down", MessageZH: "所有采集目标都判定为 down", Severity: ReasonSeverityDown, Scope: ReasonScopeSite, AffectsHealth: true, DescriptionZH: "站点下所有 target summary 均为 down。"},
	{Code: "all_targets_unknown", MessageZH: "所有采集目标状态未知", Severity: ReasonSeverityUnknown, Scope: ReasonScopeSite, AffectsHealth: true, DescriptionZH: "站点下所有 target summary 均为 unknown。"},
	{Code: "some_targets_degraded", MessageZH: "部分采集目标不可用或降级", Severity: ReasonSeverityDegraded, Scope: ReasonScopeSite, AffectsHealth: true, DescriptionZH: "站点下至少一个 target 为 down 或 degraded。"},
	{Code: "some_targets_warning", MessageZH: "部分采集目标存在需要关注的观测信号", Severity: ReasonSeverityWarning, Scope: ReasonScopeSite, AffectsHealth: true, DescriptionZH: "站点下至少一个 target 为 warning 或 unknown。"},
}

var reasonDefinitionsByCode = buildReasonDefinitionsByCode()

func AllReasonDefinitions() []ReasonDefinition {
	copied := make([]ReasonDefinition, len(reasonDefinitions))
	copy(copied, reasonDefinitions)
	return copied
}

func ReasonDefinitionByCode(code string) (ReasonDefinition, bool) {
	definition, ok := reasonDefinitionsByCode[code]
	return definition, ok
}

func buildReasonDefinitionsByCode() map[string]ReasonDefinition {
	values := make(map[string]ReasonDefinition, len(reasonDefinitions))
	for _, definition := range reasonDefinitions {
		values[definition.Code] = definition
	}
	return values
}

func dnsRiskReasonCode(risk string) string {
	switch risk {
	case "private_ip", "low_ttl", "nxdomain_with_answer", "ptr_empty":
		return "dns_risk_" + risk
	default:
		return "dns_risk_other"
	}
}

func tlsVerifyReasonCode(category string) string {
	switch category {
	case "expired", "not_yet_valid", "hostname_mismatch", "unknown_authority", "incompatible_usage":
		return "tls_verify_" + category
	default:
		return "tls_verify_other"
	}
}
