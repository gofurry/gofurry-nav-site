import type { CollectorEnvelope, TargetLatestResponse } from '~/types/nav'
import type { SiteDetailPageData } from '~/composables/useSiteDetailPage'
import { normalizeObservationHeaders, observationRecord as record, observationTime, type ObservationFact } from './siteObservationPresentation'
import { siteProtocolTone } from './siteDetailPresentation'

type Source = Pick<SiteDetailPageData, 'domain' | 'targetLatestCore' | 'targetHealthSummary' | 'lightProbeState'>
type Translate = (key: string) => string
const text = (value: unknown) => typeof value === 'string' ? value.trim() : ''
const strings = (value: unknown) => Array.isArray(value) ? value.map(text).filter(Boolean) : []
const rows = (value: unknown) => Array.isArray(value) ? value.map(record) : []
const bool = (value: unknown) => typeof value === 'boolean' ? value : null
const number = (value: unknown) => typeof value === 'number' && Number.isFinite(value) ? value : null
const count = (value: unknown) => number(value) !== null && (value as number) >= 0 ? value as number : null
const securityTxtValidationCodes = new Set(['content_type_not_text_plain', 'body_empty', 'body_truncated', 'contact_missing_or_invalid', 'expires_missing_or_invalid', 'security_txt_expired'])
export const siteSecurityTxtValidationLabel = (code: string, t: Translate) => securityTxtValidationCodes.has(code) ? t('siteSecurity.validation.' + code) : code
const headerCatalog = [
  ['hsts', 'strict_transport_security', 'strict-transport-security', 'HSTS'],
  ['content_security_policy', 'content_security_policy', 'content-security-policy', 'Content-Security-Policy'],
  ['x_frame_options', 'x_frame_options', 'x-frame-options', 'X-Frame-Options'],
  ['x_content_type_options', 'x_content_type_options', 'x-content-type-options', 'X-Content-Type-Options'],
  ['referrer_policy', 'referrer_policy', 'referrer-policy', 'Referrer-Policy'],
  ['permissions_policy', 'permissions_policy', 'permissions-policy', 'Permissions-Policy'],
] as const

/** Shared by Security and Target context: collector defaults are not observations. */
export function readSiteCertificateEvidence(value: unknown): Record<string, unknown> | null {
  const payload = record(value), collected = bool(payload.cert_collected)
  if (collected === false || text(payload.tls_handshake) === 'not_tls') return null
  const observed = collected === true || bool(payload.cert_verified) !== null || number(payload.cert_days_left) !== null
    || Boolean(text(payload.cert_subject_cn) || text(payload.cert_not_after) || text(payload.cert_fingerprint_sha256))
  return observed ? payload : null
}

// Require an explicit timezone so SSR and the browser use the same evidence time.
const certificateTimestamp = (value: unknown) => {
  const raw = text(value)
  if (!/^\d{4}-\d{2}-\d{2}T(?:[01]\d|2[0-3]):[0-5]\d:[0-5]\d(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$/.test(raw) || raw.startsWith('0001-01-01T')) return NaN
  const timestamp = Date.parse(raw), day = raw.slice(0, 10)
  // Date.parse normalizes impossible dates such as February 30; those are not evidence.
  return Number.isFinite(timestamp) && new Date(day).toISOString().slice(0, 10) === day ? timestamp : NaN
}

/** Legacy days take precedence; V2 falls back to whole days at HTTP observation time. */
export function presentSiteCertificateExpiry(payload: unknown, t: Translate, observedAt?: unknown) {
  const certificate = readSiteCertificateEvidence(payload)
  let days = number(certificate?.cert_days_left)
  if (days === null && certificate) {
    const expires = certificateTimestamp(certificate.cert_not_after), observed = certificateTimestamp(observedAt)
    if (Number.isFinite(expires) && Number.isFinite(observed)) days = Math.floor((expires - observed) / 86_400_000)
  }
  const state = days === null ? 'not_observed' : days <= 0 ? 'expired' : days <= 7 ? 'warning' : days <= 30 ? 'attention' : 'normal'
  return { state, days, label: t('siteSecurity.states.' + state),
    value: days === null ? t('siteDetail.notObserved') : days <= 0 ? t('siteDetail.expired') : `${days} ${t('siteDetail.days')}`,
    tone: state === 'normal' ? 'good' as const : state === 'attention' ? 'warning' as const
      : state === 'warning' || state === 'expired' ? 'bad' as const : 'neutral' as const }
}

/** Current Target evidence only. No I/O, client-clock validity calculation or security verdict. */
export function presentSiteSecurity(source: Source, t: Translate) {
  const target = source.domain
  const latest = source.targetLatestCore?.target === target ? source.targetLatestCore : null
  const light = source.lightProbeState?.target === target ? source.lightProbeState : null
  const summary = source.targetHealthSummary?.target === target ? source.targetHealthSummary : null
  const envelope = (snapshot: TargetLatestResponse | null, protocol: string) => {
    const item = snapshot?.protocols?.[protocol]
    return item && (!item.target || item.target === target) && (!item.protocol || item.protocol === protocol) ? item : undefined
  }
  const state = (value: string) => ({ state: value, label: t('siteSecurity.states.' + value) })
  const display = (value: unknown): string => typeof value === 'boolean' ? t('siteSecurity.' + (value ? 'yes' : 'no'))
    : number(value) !== null ? String(value) : Array.isArray(value) ? strings(value).join('\n') || '—' : text(value) || '—'
  const fact = (key: string, value: unknown): ObservationFact => ({ key, label: t('siteSecurity.fields.' + key), value: display(value) })
  const facts = (data: Record<string, unknown>, keys: string[]) => keys.map(key => fact(key, data[key]))
  const compact = (items: ObservationFact[]) => items.filter(item => item.value !== '—')
  const ms = (value: unknown) => count(value) === null ? '—' : `${value} ms`
  const statusOf = (item?: CollectorEnvelope) => !item ? 'not_observed' : item.status === 'failure' ? 'unavailable'
    : item.status === 'skipped' ? 'skipped' : item.status === 'success' ? 'present' : 'unknown'
  const meta = (item?: CollectorEnvelope, stale = false) => ({ ...state(statusOf(item)), stale,
    observed: observationTime(item?.observed_at),
    facts: [fact('observed_at', observationTime(item?.observed_at)), fact('duration_ms', ms(item?.duration_ms))],
    errors: compact([fact('error_code', item?.error_code), fact('error_message', item?.error_message)]),
    truncated: item?.payload_truncated === true,
  })
  const http = envelope(latest, 'http'), payload = record(http?.payload)
  const httpMeta = meta(http, summary?.protocols?.http?.stale === true || latest?.state === 'stale')
  const handshake = text(payload.tls_handshake)
  const transportState = handshake === 'failed' ? 'failed' : handshake === 'not_tls' ? 'not_applicable'
    : text(payload.tls_version) || handshake === 'collected' ? 'present' : httpMeta.state === 'present' ? 'not_observed' : httpMeta.state
  const transport = { ...httpMeta, ...state(transportState), value: text(payload.tls_version) || state(transportState).label,
    tone: transportState === 'not_observed' ? 'muted' as const : siteProtocolTone(transportState), facts: [fact('tls_version', payload.tls_version), fact('cipher_suite', payload.cipher_suite),
    fact('tls_handshake', handshake), fact('observed_at', httpMeta.observed)] }

  const collected = bool(payload.cert_collected)
  const evidence = readSiteCertificateEvidence(payload), hasCertificate = evidence !== null
  const cert = evidence ?? {}
  const verified = bool(cert.cert_verified)
  const verification = verified === true ? 'verified' : verified === false ? 'failed'
    : httpMeta.state === 'unavailable' && !hasCertificate ? 'unavailable' : handshake === 'not_tls' ? 'not_applicable' : 'not_observed'
  const expiryPresentation = presentSiteCertificateExpiry(payload, t, http?.observed_at), expiry = expiryPresentation.state, days = expiryPresentation.days
  const certificate = {
    verification: { ...state(verification), tone: siteProtocolTone(verification) }, expiry: expiryPresentation, days, collected,
    errors: compact(facts(cert, ['verify_error_category', 'verify_error'])),
    validity: [fact('cert_not_before', observationTime(cert.cert_not_before)), fact('cert_not_after', observationTime(cert.cert_not_after)), fact('cert_days_left', days)],
    subject: facts(cert, ['cert_subject_cn', 'cert_subject_org', 'cert_san_count']), san: strings(cert.cert_dns_names),
    issuer: facts(cert, ['cert_issuer_cn', 'cert_issuer_org', 'cert_chain_length']), chain: strings(cert.cert_chain_issuers),
    crypto: facts(cert, ['cert_public_key_algorithm', 'cert_public_key_bits', 'cert_signature_algorithm', 'ocsp_stapled', 'sct_count', 'cert_serial_number', 'cert_fingerprint_sha256', 'cert_spki_sha256']),
  }

  const headerSummary = record(payload.security_header_summary), headerFlags = record(payload.security_headers)
  const rawHeaders = normalizeObservationHeaders(payload.headers)
  const hasRawHeaders = payload.headers !== null && typeof payload.headers === 'object' && !Array.isArray(payload.headers)
  const headerRows = headerCatalog.map(([key, flag, rawKey, label]) => {
    const item = record(headerSummary[key])
    const value = text(headerFlags[flag]) || rawHeaders.find(row => row.key === rawKey)?.value || ''
    const present = bool(item.present) ?? bool(headerFlags[flag]) ?? (typeof headerFlags[flag] === 'string' || hasRawHeaders ? Boolean(value) : null)
    const headerState = httpMeta.state !== 'present' ? httpMeta.state : present === null ? 'not_observed' : present ? 'present' : 'missing'
    return { key, ...state(headerState), name: label, value: value || '—',
      tone: headerState === 'present' ? 'good' : headerState === 'unavailable' ? 'warning' : headerState === 'not_observed' ? 'muted' : 'neutral',
      details: compact(facts(item, ['max_age', 'include_subdomains', 'preload', 'has_default_src', 'unsafe_inline', 'unsafe_eval', 'wildcard_source', 'mode', 'nosniff', 'policy', 'directive_count'])) }
  })
  const headers = { ...httpMeta, rows: headerRows }

  const txtEnvelope = envelope(light, 'security_txt'), txt = record(txtEnvelope?.payload)
  const txtMeta = meta(txtEnvelope, light?.state === 'stale'), validation = strings(txt.validation_errors)
  const validationLabels = validation.map(code => siteSecurityTxtValidationLabel(code, t))
  const txtState = txtMeta.state !== 'present' ? txtMeta.state : bool(txt.exists) === true ? validation.length ? 'found_with_issues' : 'found'
    : bool(txt.exists) === false ? 'not_found' : 'unknown'
  const securityTxt = { ...txtMeta, ...state(txtState), metaFacts: txtMeta.facts, validation: txtMeta.state === 'present' ? validation : [],
    tone: txtState === 'found' ? 'info' as const : ['unavailable', 'found_with_issues'].includes(txtState) ? 'warning' as const : txtState === 'not_observed' ? 'muted' as const : 'neutral' as const,
    validationLabels: txtMeta.state === 'present' ? validationLabels : [],
    facts: facts(txt, ['exists', 'recognition', 'path_used', 'status_code', 'content_type', 'contact', 'expires', 'policy', 'canonical', 'preferred_languages']),
    truncated: txtMeta.truncated || txt.body_truncated === true }

  const portEnvelope = envelope(light, 'port_check'), ports = record(portEnvelope?.payload)
  const portMeta = meta(portEnvelope, light?.state === 'stale')
  const portRows = rows(ports.results).map((item, index) => ({ key: String(index),
    status: ['open', 'closed', 'timeout', 'filtered_suspected', 'skipped'].includes(text(item.status)) ? text(item.status) : 'unknown',
    tone: item.status === 'timeout' ? 'warning' : ['closed', 'filtered_suspected', 'skipped'].includes(text(item.status)) ? 'muted' : 'neutral',
    facts: [fact('port', item.port), fact('service_hint', item.service_hint), fact('status', t('siteSecurity.states.' + (['open', 'closed', 'timeout', 'filtered_suspected', 'skipped'].includes(text(item.status)) ? text(item.status) : 'unknown'))), fact('duration_ms', ms(item.duration_ms))],
    errors: compact(facts(item, ['error_code', 'error_message'])) }))
  const portCheck = { ...portMeta, ...state(portMeta.state === 'present' && Array.isArray(ports.results) && !portRows.length ? 'empty' : portMeta.state),
    summary: facts(ports, ['ports_checked', 'open_count', 'closed_count', 'timeout_count', 'filtered_suspected_count', 'skipped_count']),
    results: portRows, metadata: facts(ports, ['ports_configured', 'invalid_port_count', 'duplicate_port_count', 'truncated_port_count', 'truncated', 'skipped_reason']),
    truncated: portMeta.truncated || ports.truncated === true }

  const wafEnvelope = envelope(light, 'waf_canary'), waf = record(wafEnvelope?.payload)
  const wafMeta = meta(wafEnvelope, light?.state === 'stale')
  const truncated = wafMeta.truncated || waf.target_run_truncated === true || (count(waf.truncated_target_count) ?? 0) > 0
  const mismatchKeys = ['unexpected_pass_count', 'network_error_count', 'status_code_unexpected_count']
  const mismatches = wafMeta.state === 'present' ? mismatchKeys.filter(key => (count(waf[key]) ?? 0) > 0) : []
  // Matching is bounded to this complete, successful sample, never proof of a WAF.
  const matched = wafMeta.state === 'present' && !truncated && (count(waf.expected_blocked_count) ?? 0) > 0
    && count(waf.expected_blocked_matched_count) === count(waf.expected_blocked_count)
    && count(waf.cases_executed) === count(waf.cases_total) && (count(waf.cases_total) ?? 0) > 0
    && mismatchKeys.every(key => count(waf[key]) === 0)
  const wafState = wafMeta.state !== 'present' ? wafMeta.state : truncated ? 'incomplete' : mismatches.length ? 'mismatch' : matched ? 'matched' : 'present'
  const wafCanary = { ...wafMeta, ...state(wafState), truncated,
    tone: wafState === 'matched' ? 'info' : ['mismatch', 'unavailable', 'incomplete'].includes(wafState) ? 'warning' : 'neutral',
    summaryText: wafMeta.state === 'present' && count(waf.expected_blocked_matched_count) !== null && count(waf.expected_blocked_count) !== null
      ? `${waf.expected_blocked_matched_count} / ${waf.expected_blocked_count} ${t('siteSecurity.expectedMatched')}` : state(wafState).label,
    summary: facts(waf, ['cases_total', 'cases_executed', 'blocked_count', 'expected_blocked_count', 'expected_blocked_matched_count', ...mismatchKeys]),
    metadata: facts(waf, ['target_run_truncated', 'truncated_target_count']),
    cases: rows(waf.cases).map((item, index) => ({ key: String(index), facts: [...facts(item, ['case_id', 'category', 'method', 'status_code', 'blocked', 'expected_blocked', 'matched_expected']), fact('duration_ms', ms(item.duration_ms))], errors: compact(facts(item, ['error_code', 'error_message'])) })),
  }

  const attention: { key: string; message: string; detail: string }[] = []
  const addAttention = (key: string, detail = '') => attention.push({ key, message: t('siteSecurity.attention.' + key), detail })
  if (verification === 'failed') addAttention('verification', text(cert.verify_error) || text(cert.verify_error_category))
  if (['attention', 'warning', 'expired'].includes(expiry)) addAttention('expiry', `${days} ${t('siteDetail.days')}`)
  if (txtState === 'found_with_issues') addAttention('securityTxt', validationLabels.join('; '))
  for (const key of mismatches) addAttention(key, display(waf[key]))
  const overview = { attention, sections: [
    { key: 'transport', title: t('siteSecurity.transport'), value: transport.value, detail: '', tone: transport.tone },
    { key: 'certificate', title: t('siteSecurity.certificate'), value: certificate.verification.label, detail: expiryPresentation.value, tone: certificate.verification.tone },
    { key: 'headers', title: t('siteSecurity.headers'), value: headerRows.some(row => row.state === 'present' || row.state === 'missing')
      ? `${headerRows.filter(row => row.state === 'present').length} / ${headerRows.length} ${t('siteSecurity.observedCount')}`
      : state(httpMeta.state === 'present' ? 'not_observed' : httpMeta.state).label, detail: '' },
    { key: 'securityTxt', title: 'security.txt', value: securityTxt.label, detail: '', tone: securityTxt.tone },
    { key: 'portCheck', title: t('siteSecurity.portCheck'), value: portMeta.state === 'present' && count(ports.ports_checked) !== null
      ? `${ports.ports_checked} ${t('siteSecurity.checkedCount')}` : portCheck.label, detail: '' },
    { key: 'wafCanary', title: t('siteSecurity.wafCanary'), value: wafCanary.summaryText, detail: wafState === 'matched' ? '' : wafCanary.label, tone: wafCanary.tone },
  ] }
  return { target, overview, transport, certificate, headers, securityTxt, portCheck, wafCanary }
}
export type SiteSecurityPresentation = ReturnType<typeof presentSiteSecurity>
