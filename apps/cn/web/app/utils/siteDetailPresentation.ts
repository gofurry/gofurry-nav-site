export type SiteDetailTone = 'good' | 'warning' | 'bad' | 'neutral' | 'muted' | 'info'
type Translate = (key: string, values?: Record<string, string | number>) => string

const healthReasons = new Set([
  'http_missing_or_stale', 'http_failed', 'dns_failed', 'dns_missing_or_stale',
  'dns_failed_but_http_ok', 'ping_failed_but_http_ok', 'dns_risk_private_ip',
  'dns_risk_low_ttl', 'dns_risk_nxdomain_with_answer', 'dns_risk_ptr_empty', 'dns_risk_other',
  'tls_verify_expired', 'tls_verify_not_yet_valid', 'tls_verify_hostname_mismatch',
  'tls_verify_unknown_authority', 'tls_verify_incompatible_usage', 'tls_verify_other',
  'tls_cert_expired', 'tls_cert_expiring_soon', 'no_target_summary', 'all_targets_down',
  'all_targets_unknown', 'some_targets_degraded', 'some_targets_warning',
])

/** Localized copy only: Collector remains the sole health classification owner. */
export function siteHealthReasonLabel(code: string, t: Translate) {
  if (code === 'summary_stale') return t('siteOverview.staleAttention')
  if (code === 'summary_missing') return t('siteOverview.missingAttention')
  return healthReasons.has(code) ? t('siteDetail.reasons.' + code) : t('siteOverview.reasonCode', { code })
}

/** UI response-time guidance, not a new health verdict or network standard. */
export const siteLatencyTone = (value: number | null): SiteDetailTone => value === null || !Number.isFinite(value) || value < 0 ? 'neutral'
  : value <= 800 ? 'good' : value <= 2000 ? 'warning' : 'bad'
export const siteHttpTone = (value: number | null): SiteDetailTone => value === null ? 'neutral'
  : value >= 400 && value < 600 ? 'bad' : value >= 200 && value < 300 ? 'good' : 'neutral'
export const siteProtocolTone = (value: string): SiteDetailTone => ['healthy', 'success', 'present', 'verified'].includes(value) ? 'good'
  : ['down', 'degraded', 'failure', 'failed'].includes(value) ? 'bad'
    : ['warning', 'stale', 'unavailable'].includes(value) ? 'warning' : 'neutral'

export function siteDnsSignal(flag: string, t: Translate) {
  const warning = flag === 'private_ip' || flag === 'nxdomain_with_answer'
  const known = warning || flag === 'ptr_empty' || flag === 'low_ttl'
  return { code: flag, tone: warning ? 'warning' as const : 'neutral' as const,
    kind: t('siteObservation.' + (warning ? 'warningSignal' : 'informationalSignal')),
    label: t('siteObservation.dnsSignals.' + (known ? flag : 'other')) }
}
