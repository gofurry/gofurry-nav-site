import { describe, expect, it } from 'vitest'
import { presentSiteSecurity } from '../../app/utils/siteSecurityPresentation'
import { buildSiteDetailQuery, parseSiteDetailRouteState, selectSiteSecurityView, selectSiteObservationView } from '../../app/utils/siteDetailRouteState'
import type { SiteDetailPageData } from '../../app/composables/useSiteDetailPage'
import { securityEvidence, securityState } from '../browser/fixtures/site-security-data'
import en from '../../i18n/locales/en.json'
import zh from '../../i18n/locales/zh.json'
type Source = Pick<SiteDetailPageData, 'domain' | 'targetLatestCore' | 'targetHealthSummary' | 'lightProbeState'>
const target = 'target.example'
const translate = (locale = 'en') => (key: string) => {
  const value = key.split('.').reduce<unknown>((value, part) => (value as Record<string, unknown>)[part], locale === 'en' ? en : zh)
  if (typeof value !== 'string') throw new Error('Missing translation ' + key)
  return value
}
function source(options: Partial<ReturnType<typeof securityState>> = {}): Source {
  const data = securityEvidence(target, { ...securityState(), ...options })
  return { domain: target, targetLatestCore: { site_id: 41, target, state: 'ready', protocols: { http: data.http } }, targetHealthSummary: null, lightProbeState: data.light }
}
const present = (options: Partial<ReturnType<typeof securityState>> = {}) => presentSiteSecurity(source(options), translate())
const payload = (data: Source) => data.targetLatestCore!.protocols.http!.payload as Record<string, unknown>

describe('Security route ownership', () => {
  it('keeps Target and clears foreign state for all Security views, omitting overview', () => {
    const from = parseSiteDetailRouteState({ domain: target, tab: 'insights', metric: 'hsts', range: 'all' })
    for (const view of ['overview', 'tls', 'web', 'exposure'] as const) {
      expect(buildSiteDetailQuery(selectSiteSecurityView(from, view))).toEqual({ domain: target, tab: 'security', ...(view === 'overview' ? {} : { view }) })
    }
  })
  it('crosses to raw HTTP through the same route owner and retains Target', () => {
    expect(buildSiteDetailQuery(selectSiteObservationView(parseSiteDetailRouteState({ domain: target, tab: 'security', view: 'web' }), 'http'))).toEqual({ domain: target, tab: 'observation', view: 'http' })
  })
})
describe('Current Target security projection', () => {
  it('never exposes conclusion fields anywhere in the model', () => {
    function inspect(value: unknown) {
      if (!value || typeof value !== 'object') return
      for (const [key, child] of Object.entries(value)) {
        expect(['score', 'grade', 'rating', 'riskPercent', 'overallLevel', 'wafEnabled']).not.toContain(key)
        inspect(child)
      }
    }
    inspect(present())
    expect(present().overview.attention).toEqual([])
    expect(present().overview.sections.find(item => item.key === 'certificate')?.value).toBe('Verified')
    expect(present().overview.sections.find(item => item.key === 'headers')?.value).toBe('6 / 6 observed')
    expect(present().overview.sections.find(item => item.key === 'transport')?.value).toBe('TLS 1.3')
  })
  it.each(['targetLatestCore', 'lightProbeState'] as const)('rejects foreign %s identity', key => {
    const data = source(); data[key]!.target = 'foreign'
    const vm = presentSiteSecurity(data, translate())
    expect(key === 'targetLatestCore' ? vm.certificate.verification.state : vm.securityTxt.state).toBe('not_observed')
  })
  it('rejects mismatched per-envelope Target/protocol and handles missing envelopes', () => {
    const data = source(); data.targetLatestCore!.protocols.http!.target = 'foreign'
    data.lightProbeState!.protocols.security_txt!.protocol = 'robots'
    delete data.lightProbeState!.protocols.port_check; delete data.lightProbeState!.protocols.waf_canary
    const vm = presentSiteSecurity(data, translate())
    expect([vm.transport.state, vm.securityTxt.state, vm.portCheck.state, vm.wafCanary.state]).toEqual(Array(4).fill('not_observed'))
  })
  it.each([
    ['verified', 'verified'], ['failed', 'failed'], ['missing_verification', 'not_observed'], ['not_collected', 'not_observed'], ['missing', 'not_observed'], ['not_tls', 'not_applicable'],
  ] as const)('certificate %s is %s without interpreting Go zero defaults', (tls, expected) => {
    const vm = present({ tls })
    expect(vm.certificate.verification.state).toBe(expected)
    if (tls === 'not_collected' || tls === 'not_tls') {
      expect(vm.certificate.days).toBeNull(); expect(vm.certificate.crypto.every(item => item.value === '—')).toBe(true)
      expect(vm.overview.attention).toEqual([])
    }
    if (tls === 'failed') expect(vm.certificate.errors.map(item => item.value)).toContain('unknown_authority')
  })
  it('keeps an explicit verification fact when the collection flag is absent', () => {
    const data = source(); delete payload(data).cert_collected
    expect(presentSiteSecurity(data, translate()).certificate.verification.state).toBe('verified')
  })
  it.each([[31, 'normal'], [30, 'attention'], [8, 'attention'], [7, 'warning'], [1, 'warning'], [0, 'expired'], [-2, 'expired']] as const)('expiry %s stays separate from verification', (days, state) => {
    const vm = present({ days })
    expect(vm.certificate.verification.state).toBe('verified')
    expect(vm.certificate.expiry.state).toBe(state)
    expect(vm.certificate.days).toBe(days)
    expect(vm.overview.attention.some(item => item.key === 'expiry')).toBe(['attention', 'warning', 'expired'].includes(state))
  })
  it('retains SAN, chain, crypto, false OCSP and zero SCT without extra Attention', () => {
    const vm = present()
    expect(vm.certificate.san).toHaveLength(2); expect(vm.certificate.chain).toHaveLength(2)
    expect(vm.certificate.crypto.find(item => item.key === 'ocsp_stapled')?.value).toBe('No')
    expect(vm.certificate.crypto.find(item => item.key === 'sct_count')?.value).toBe('0')
    expect(vm.certificate.crypto.find(item => item.key === 'cert_fingerprint_sha256')?.value).toHaveLength(64)
    expect(vm.overview.attention).toEqual([])
  })
  it.each([['all', 6, 0], ['some', 2, 4], ['missing', 0, 6], ['not_observed', 0, 0]] as const)('headers %s preserve observed absence versus missing evidence', (headers, found, missing) => {
    const vm = present({ headers })
    expect(vm.headers.rows.filter(row => row.state === 'present')).toHaveLength(found)
    expect(vm.headers.rows.filter(row => row.state === 'missing')).toHaveLength(missing)
    if (headers === 'not_observed') expect(vm.headers.rows.every(row => row.state === 'not_observed')).toBe(true)
    expect(vm.overview.attention).toEqual([])
  })
  it('prioritizes collector presence summary and retains raw header values without reanalysis', () => {
    const data = source(), http = payload(data)
    http.security_header_summary = { hsts: { present: false } }
    const vm = presentSiteSecurity(data, translate())
    expect(vm.headers.rows[0]).toMatchObject({ state: 'missing', value: 'max-age=31536000; includeSubDomains' })
    delete http.security_header_summary
    expect(presentSiteSecurity(data, translate()).headers.rows[0]?.state).toBe('present')
  })
  it('marks failed HTTP headers unavailable, never Missing', () => {
    expect(present({ httpFailure: true }).headers.rows.every(row => row.state === 'unavailable')).toBe(true)
  })
  it.each([['found', 'found'], ['issues', 'found_with_issues'], ['not_found', 'not_found'], ['unavailable', 'unavailable'], ['not_observed', 'not_observed']] as const)('security.txt %s maps to %s', (txt, state) => {
    const vm = present({ txt }); expect(vm.securityTxt.state).toBe(state)
    expect(vm.overview.attention.some(item => item.key === 'securityTxt')).toBe(txt === 'issues')
  })
  it('does not revalidate RFC fields or replace reported errors', () => {
    const data = source(), txt = data.lightProbeState!.protocols.security_txt!.payload as Record<string, unknown>
    txt.expires = '1999-01-01'; txt.contact = []; txt.validation_errors = []
    expect(presentSiteSecurity(data, translate()).securityTxt.state).toBe('found')
    txt.validation_errors = ['collector_specific_issue']
    expect(presentSiteSecurity(data, translate()).securityTxt.validation).toEqual(['collector_specific_issue'])
  })
  it('keeps open ports neutral, distinguishes connection outcomes and discloses metadata', () => {
    const vm = present()
    expect(vm.portCheck.results.map(row => row.status)).toEqual(['open', 'closed', 'timeout', 'filtered_suspected', 'skipped'])
    expect(vm.portCheck.results.map(row => row.tone)).toEqual(['neutral', 'muted', 'warning', 'muted', 'muted'])
    expect(vm.portCheck.metadata.find(item => item.key === 'duplicate_port_count')?.value).toBe('1')
    expect(vm.portCheck.truncated).toBe(true); expect(vm.overview.attention).toEqual([])
  })
  it.each([['empty', 'empty'], ['skipped', 'skipped'], ['unavailable', 'unavailable'], ['not_observed', 'not_observed']] as const)('port %s maps to %s', (ports, state) => {
    expect(present({ ports }).portCheck.state).toBe(state)
  })
  it.each(['unexpected_pass', 'network_error', 'unexpected_status'] as const)('only reported WAF %s produces Attention', waf => {
    const vm = present({ waf }); expect(vm.wafCanary.state).toBe('mismatch'); expect(vm.overview.attention).toHaveLength(1)
  })
  it.each([['matched', 'matched'], ['truncated', 'incomplete'], ['unavailable', 'unavailable'], ['not_observed', 'not_observed']] as const)('WAF %s maps to %s', (waf, state) => {
    expect(present({ waf }).wafCanary.state).toBe(state)
  })
  it('does not claim matching when counts are unknown, zero, partial or only target-count truncation exists', () => {
    const data = source(), waf = data.lightProbeState!.protocols.waf_canary!.payload as Record<string, unknown>
    waf.network_error_count = undefined
    expect(presentSiteSecurity(data, translate()).wafCanary.state).toBe('present')
    waf.network_error_count = 0; waf.cases_executed = 1
    expect(presentSiteSecurity(data, translate()).wafCanary.state).toBe('present')
    waf.cases_executed = 2; waf.expected_blocked_count = 0
    expect(presentSiteSecurity(data, translate()).wafCanary.state).toBe('present')
    waf.truncated_target_count = 1
    expect(presentSiteSecurity(data, translate()).wafCanary.state).toBe('incomplete')
  })
  it('keeps stale evidence separate and translates both locales without a client clock', () => {
    const data = source({ stale: true }); data.targetLatestCore!.state = 'stale'
    const vm = presentSiteSecurity(data, translate('zh'))
    expect(vm.transport.stale).toBe(true); expect(vm.securityTxt.stale).toBe(true)
    expect(vm.transport.observed).toBe('2026-08-30 12:00:00 UTC'); expect(vm.certificate.verification.label).toBe('已验证')
  })
})


describe('First-round Security composition facts', () => {
  it('shows actual TLS independently from certificate verification and expiry', () => {
    const data = source({ tls: 'failed', days: 45 })
    payload(data).tls_version = 'TLS 1.2'
    const vm = presentSiteSecurity(data, translate())
    expect(vm.transport).toMatchObject({ value: 'TLS 1.2', tone: 'good' })
    expect(vm.certificate.verification.state).toBe('failed')
    expect(vm.certificate.expiry.state).toBe('normal')
    expect(present({ tls: 'missing' }).transport).toMatchObject({ state: 'not_observed', tone: 'muted' })
  })
  it('keeps configured ports as metadata while exposing the six requested summary facts', () => {
    const vm = present()
    expect(vm.portCheck.summary.map(item => item.key)).toEqual(['ports_checked', 'open_count', 'closed_count', 'timeout_count', 'filtered_suspected_count', 'skipped_count'])
    expect(vm.portCheck.metadata.some(item => item.key === 'ports_configured')).toBe(true)
    expect(vm.portCheck.results.find(item => item.status === 'open')?.tone).toBe('neutral')
    expect(vm.wafCanary.tone).toBe('info')
    expect(vm.wafCanary.summaryText).toBe('1 / 1 expected behaviors matched')
  })
  it.each([['found', 'info'], ['not_found', 'neutral'], ['unavailable', 'warning'], ['not_observed', 'muted']] as const)('keeps security.txt %s tone factual', (txt, tone) => {
    expect(present({ txt }).securityTxt.tone).toBe(tone)
  })
})
