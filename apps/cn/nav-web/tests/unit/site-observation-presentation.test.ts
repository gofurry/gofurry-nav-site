import { describe, expect, it } from 'vitest'
import { normalizeObservationHeaders, presentPingHistory, presentSiteObservation } from '../../app/utils/siteObservationPresentation'
import type { CollectorEnvelope, TargetLatestResponse, TargetHealthSummary } from '../../app/types/nav'
import en from '../../i18n/locales/en.json'
import zh from '../../i18n/locales/zh.json'
const translate = (locale = 'en') => (key: string) => {
  const value = key.split('.').reduce<unknown>((value, part) => (value as Record<string, unknown>)[part], locale === 'en' ? en : zh)
  if (typeof value !== 'string') throw new Error('Missing translation ' + key)
  return value
}
const target = 'a.example'
const envelope = (protocol: string, payload: unknown, extra: Partial<CollectorEnvelope> = {}): CollectorEnvelope => ({
  site_id: 41, target, protocol, status: 'success', observed_at: '2026-09-26T12:00:00Z', duration_ms: 0, payload, schema_version: 1, ...extra,
})
const latest = (protocols: Record<string, CollectorEnvelope>): TargetLatestResponse => ({ target, site_id: 41, state: 'ready', protocols })
const source = (protocols: Record<string, CollectorEnvelope> = {}, summary: TargetHealthSummary | null = null, light: TargetLatestResponse | null = null) => ({
  domain: target, targetLatestCore: latest(protocols), targetHealthSummary: summary, lightProbeState: light,
})
const present = (protocols: Record<string, CollectorEnvelope> = {}) => presentSiteObservation(source(protocols), translate())

describe('Current Target evidence projection', () => {
  it('keeps all four protocol fields, prefers summary and preserves status separately from freshness', () => {
    const summary = { target, state: 'ready', status: 'warning', reason_messages: ['Readable reason', 'Readable reason'], reason_codes: ['hidden_raw'],
      protocols: { ping: { status: 'success', duration_ms: 5, observed_at: '2026-09-25T12:00:00Z', stale: true } } } as TargetHealthSummary
    const vm = presentSiteObservation(source({ ping: envelope('ping', {}, { duration_ms: 100 }) }, summary), translate())
    expect(vm.protocols[0]).toMatchObject({ status: 'success', duration: '5 ms', observed: '2026-09-25 12:00:00 UTC', freshness: 'Stale' })
    expect(vm.risks).toEqual(['Readable reason'])
    expect(vm.status).toBe('warning')
  })
  it('rejects unrelated Target summaries, envelopes and light probes', () => {
    const data = source({ http: envelope('http', { final_url: 'foreign' }, { target: 'foreign' }) }, { target: 'foreign', status: 'down' } as TargetHealthSummary,
      { ...latest({ robots: envelope('robots', { exists: true }) }), target: 'foreign' })
    const vm = presentSiteObservation(data, translate())
    expect(vm.endpoint).toEqual([])
    expect(vm.status).toBe('unknown')
    expect(vm.web.probes[0]?.status).toBe('unknown')
    data.targetLatestCore = { ...data.targetLatestCore, target: 'foreign' }
    expect(presentSiteObservation(data, translate()).http.facts.find(item => item.key === 'status')?.value).toBe('—')
  })
  it('keeps absent facts unknown and valid zero/false values visible', () => {
    const empty = present()
    expect(empty.endpoint).toEqual([])
    expect(empty.risks).toEqual([])
    expect(empty.protocols.every(item => item.freshness === 'Freshness unknown')).toBe(true)
    const vm = present({ ping: envelope('ping', { avg_rtt_ms: 0, jitter_ms: 0, loss_rate: 0 }), http: envelope('http', { response_time_ms: 0, body_read_bytes: 0 }) })
    expect(vm.kpis.map(item => item.value)).toEqual(['0 ms', '0 ms', '0 ms', '0%'])
    expect(vm.http.facts.find(item => item.key === 'bytes')?.value).toBe('0 B')
  })
  it.each([0, .1, 1, 10, 100])('retains collector percentage %s without guessing a ratio', loss => {
    expect(present({ ping: envelope('ping', { loss_rate: loss }) }).kpis[3]?.value).toBe(loss + '%')
  })
  it('retains independent timings rather than synthesizing a total', () => {
    const vm = present({ http: envelope('http', { dns_lookup_ms: 5, tcp_connect_ms: 7, tls_handshake_ms: 11, ttfb_ms: 90, transfer_ms: 13, response_time_ms: 120 }) })
    expect(vm.timings.map(item => item.value)).toEqual([5, 7, 11, 90, 13, 120])
    expect(vm.timings.at(-1)?.value).not.toBe(126)
    expect(vm.timings.every(item => item.fraction >= 0 && item.fraction <= 1)).toBe(true)
  })
  it('normalizes header case and repeated values and keeps redirects conditional', () => {
    expect(normalizeObservationHeaders({ Server: ['edge'], server: 'origin', Empty: [], 'Content-Type': ['text/html'] })).toEqual([
      { key: 'content-type', label: 'content-type', value: 'text/html' }, { key: 'server', label: 'server', value: 'edge\norigin' },
    ])
    const vm = present({ http: envelope('http', { headers: { Server: ['edge'], 'X-Probe': ['a'] }, redirect_chain: ['http://a.example', 'https://a.example'] }) })
    expect(vm.http.commonHeaders.map(item => item.key)).toEqual(['server'])
    expect(vm.http.headers).toHaveLength(2)
    expect(vm.http.redirects).toHaveLength(2)
    expect(present().http.redirects).toEqual([])
  })
  it('groups nested DNS evidence and preserves chains, secondary detail and backend risks', () => {
    const vm = present({ dns: envelope('dns', { A: [], CNAME: [{ type: 'CNAME', value: 'edge.example', ttl: 60, children: [
      { type: 'A', value: '203.0.113.1', ttl: 0, dnssec: false, asn: 'AS64496', risk_flags: ['reported_flag'] },
    ] }], risk_flags: ['reported_flag'] }) })
    expect(vm.dns.groups.map(group => group.type)).toEqual(['A', 'CNAME'])
    expect(vm.dns.chains).toEqual([['a.example', 'edge.example', '203.0.113.1']])
    expect(vm.dns.groups[0]?.records[0]?.ttl).toBe('0 s')
    expect(vm.dns.groups[0]?.records[0]?.details).toContainEqual({ key: 'dnssec', label: 'DNSSEC', value: 'No' })
    expect(vm.dns.risks).toEqual(['reported_flag'])
    expect(present().dns.groups).toEqual([])
    expect(present().dns.chains).toEqual([])
  })
  it('uses a strict Web allowlist and retains robots/llms/assets/RDAP facts and disclosures', () => {
    const probes = latest({ robots: envelope('robots', { exists: false, global_disallow_all: false, sitemaps: ['https://a.example/map'] }),
      llms_txt: envelope('llms_txt', { headings: ['About'], links: ['https://a.example'], optional_section_present: false, body_read_bytes: 0 }),
      page_assets: envelope('page_assets', { manifest: { name: 'App', theme_color: '#123456', icons_count: 0 } }),
      rdap: envelope('rdap', { registrable_domain: 'a.example', dnssec_delegation_signed: false }),
      security_txt: envelope('security_txt', { contact: ['SECRET'] }), port_check: envelope('port_check', { ports: ['SECRET'] }), waf_canary: envelope('waf_canary', { text: 'SECRET' }) })
    const vm = presentSiteObservation(source({}, null, probes), translate())
    expect(vm.web.probes.map(probe => probe.protocol)).toEqual(['robots', 'llms_txt', 'page_assets', 'rdap'])
    expect(JSON.stringify(vm.web)).not.toContain('SECRET')
    expect(vm.web.probes[1]?.sections.find(section => section.disclosure)?.items).toHaveLength(2)
    expect(JSON.stringify(vm.web)).toContain('No')
    expect(presentSiteObservation(source({}, null, probes), translate('zh')).web.probes[0]?.sections[0]?.items[0]?.value).toBe('否')
  })
  it('keeps history null measurements, identity, percentage and descending UTC order', () => {
    const rows = presentPingHistory([envelope('ping', {}, { observed_at: '2026-09-25T12:00:00Z', status: 'failure' }),
      envelope('ping', { avg_rtt_ms: 0, loss_rate: .1 }), envelope('http', { avg_rtt_ms: 10 }),
      envelope('ping', { avg_rtt_ms: 10 }, { target: 'foreign' })], target)
    expect(rows).toHaveLength(2)
    expect(rows[0]).toMatchObject({ rtt: 0, loss: .1, time: '2026-09-26 12:00:00 UTC' })
    expect(rows[1]).toMatchObject({ rtt: null, loss: null, status: 'failure' })
    expect(presentPingHistory([], target)).toEqual([])
  })
})
