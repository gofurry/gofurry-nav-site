import { describe, expect, it } from 'vitest'
import { presentSiteTarget } from '../../app/utils/siteTargetPresentation'
import type { SiteHealthSummary, TargetHealthSummary, TargetLatestResponse } from '../../app/types/nav'

const t = (key: string) => key
function source(payload: unknown = {}) {
  return {
    domain: 'a.example',
    siteHealthSummary: { status: 'down', targets: [{ target: 'a.example', status: 'down' }, { target: 'b.example' }] } as SiteHealthSummary,
    targetHealthSummary: { state: 'ready', target: 'a.example', status: 'healthy', protocols: {} } as TargetHealthSummary,
    targetLatestCore: { target: 'a.example', protocols: { http: { target: 'a.example', payload } } } as TargetLatestResponse,
  }
}
describe('Current Target presentation', () => {
  it('accepts the backend missing-summary shape with nil targets and zero time', () => {
    const data = source()
    data.siteHealthSummary = JSON.parse('{"state":"missing","status":"unknown","targets":null}') as SiteHealthSummary
    data.targetHealthSummary.state = 'missing'
    data.targetHealthSummary.observed_at = '0001-01-01T00:00:00Z'
    const result = presentSiteTarget(data, t)
    expect(result.status).toBe('unknown')
    expect(result.observedAt).toBe('—')
    expect(result.targetList).toEqual([{ target: 'a.example', relation: '' }])
    data.targetLatestCore.protocols.http!.observed_at = '2026-08-30T12:00:00Z'
    expect(presentSiteTarget(data, t).observedAt).toBe('2026-08-30 12:00:00 UTC')
  })
  it('never substitutes Site aggregate health or an unrelated Target payload', () => {
    const data = source({ status_code: 200 })
    expect(presentSiteTarget(data, t).status).toBe('healthy')
    data.targetHealthSummary.target = 'b.example'
    data.targetLatestCore.target = 'b.example'
    const result = presentSiteTarget(data, t)
    expect(result.status).toBe('unknown')
    expect(result.httpStatus).toBeNull()
    expect(result.targetList.map(item => item.target)).toEqual(['a.example', 'b.example'])
  })

  it('keeps absence, zero and a failed verification distinct', () => {
    const missing = presentSiteTarget(source(), t)
    expect(missing.certificateDays).toBeNull()
    expect(missing.certificateVerified).toBeNull()
    expect(missing.certificateLabel).toBe('siteDetail.notObserved')
    expect(missing.latency).toBeNull()
    expect(missing.health.map(item => item.key)).toEqual(['status', 'latency', 'http', 'tls', 'certificate', 'observed'])
    expect(missing.health.slice(1).every(item => item.value === '—')).toBe(true)
    const failed = presentSiteTarget(source({ response_time_ms: 0, cert_days_left: 0, cert_verified: false }), t)
    expect(failed.latency).toBe(0)
    expect(failed.certificateDays).toBe(0)
    expect(failed.certificateLabel).toBe('siteDetail.notVerified')
    expect(failed.health[4]?.tone).toBe('warning')
  })

  it('preserves target staleness and protocol state instead of inferring success', () => {
    const data = source()
    data.targetHealthSummary.state = 'stale'
    data.targetHealthSummary.protocols.http = { protocol: 'http', status: 'success', stale: true, duration_ms: 12, observed_at: 'invalid', stale_after_seconds: 60 }
    const result = presentSiteTarget(data, t)
    expect(result.status).toBe('stale')
    expect(result.protocolStates[1]).toMatchObject({ status: 'stale', duration: '12 ms', observedAt: '—' })
    expect(result.protocolStates[0]).toMatchObject({ status: 'unknown', tone: 'neutral', duration: '—' })
  })

  it('formats deterministic evidence and keeps infrastructure confidence and relations', () => {
    const data = source({ final_url: 'https://final.example/', tls_version: 'TLS 1.3', cert_days_left: 42, cert_verified: true })
    data.targetHealthSummary.observed_at = '2026-08-30T20:00:00+08:00'
    data.targetHealthSummary.edge_provider_hints = [{ provider: 'cloudflare', hint_type: 'header', confidence: 'medium', evidence: [{ source: 'http', field: 'server', value: 'cloudflare' }] }]
    data.siteHealthSummary.target_relation_hints = [{ relation: 'shared_canonical', host: 'a.example', targets: ['a.example', 'b.example'] }]
    const result = presentSiteTarget(data, t)
    expect(result.visitUrl).toBe('https://final.example/')
    expect(result.observedAt).toBe('2026-08-30 12:00:00 UTC')
    expect(result.edgeProviderHints[0]).toMatchObject({ confidence: 'medium', evidence: 'http: server = cloudflare' })
    expect(result.targetList[1]?.relation).toBe('shared_canonical · a.example')
    expect(result.certificateLabel).toBe('siteDetail.verified')
    expect(presentSiteTarget(source({ final_url: 'javascript:alert(1)' }), t).visitUrl).toBe('https://a.example')
  })
})
