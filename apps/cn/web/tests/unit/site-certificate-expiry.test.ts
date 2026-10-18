import { describe, expect, it, vi } from 'vitest'
import { presentSiteCertificateExpiry, presentSiteSecurity } from '../../app/utils/siteSecurityPresentation'
import { presentSiteTarget } from '../../app/utils/siteTargetPresentation'
import type { CollectorEnvelope, TargetLatestResponse } from '../../app/types/nav'
import en from '../../i18n/locales/en.json'

const t = (key: string) => key.split('.').reduce<unknown>((value, part) => (value as Record<string, unknown>)[part], en) as string
const observedAt = '2026-09-27T12:00:00Z'
const certificate = { cert_verified: true, cert_not_before: '2026-08-01T00:00:00Z', cert_not_after: '2026-11-11T12:00:00Z' }

function projections(payload: Record<string, unknown>, observed: unknown = observedAt) {
  const http = { target: 'a.example', protocol: 'http', status: 'success', observed_at: observed, payload } as CollectorEnvelope
  const source = { domain: 'a.example', siteHealthSummary: null, targetHealthSummary: null, lightProbeState: null,
    targetLatestCore: { target: 'a.example', protocols: { http } } as TargetLatestResponse }
  return { target: presentSiteTarget(source, t), security: presentSiteSecurity(source, t) }
}

describe('Certificate remaining days at HTTP observation time', () => {
  it.each([undefined, null, NaN, Infinity, 'invalid'])('derives V2 days when legacy days are %s across all active projections', legacy => {
    const payload: Record<string, unknown> = { ...certificate }
    if (legacy !== undefined) payload.cert_days_left = legacy
    const { target, security } = projections(payload)
    expect(target.certificateDays).toBe(45)
    expect(target.health.find(item => item.key === 'certificate')).toMatchObject({ value: '45 days left', tone: 'good' })
    expect(security.certificate).toMatchObject({ days: 45, verification: { state: 'verified' }, expiry: { days: 45, state: 'normal', value: '45 days left' } })
    expect(security.certificate.validity.find(item => item.key === 'cert_days_left')?.value).toBe('45')
    expect(security.overview.sections.find(item => item.key === 'certificate')?.detail).toBe('45 days left')
    expect(payload.cert_days_left).toBe(legacy)
  })

  it.each([[31, 'normal'], [30, 'attention'], [8, 'attention'], [7, 'warning'], [1, 'warning'], [0, 'expired'], [-2, 'expired']] as const)('explicit %s days has priority and keeps %s urgency', (days, state) => {
    const { target, security } = projections({ ...certificate, cert_days_left: days })
    expect(target.certificateDays).toBe(days)
    expect(security.certificate.expiry).toMatchObject({ days, state })
    expect(security.certificate.validity.find(item => item.key === 'cert_days_left')?.value).toBe(String(days))
    expect(security.certificate.verification.state).toBe('verified')
  })

  it('uses explicit days even without either timestamp', () => {
    expect(presentSiteCertificateExpiry({ cert_verified: true, cert_days_left: 7 }, t)).toMatchObject({ days: 7, state: 'warning' })
  })

  it.each([
    ['2026-10-28T12:00:00Z', 31, 'normal'], ['2026-10-27T12:00:00Z', 30, 'attention'],
    ['2026-10-05T12:00:00Z', 8, 'attention'], ['2026-10-04T12:00:00Z', 7, 'warning'],
    ['2026-09-28T12:00:00Z', 1, 'warning'], ['2026-09-28T11:59:59Z', 0, 'expired'],
    ['2026-09-27T12:00:00Z', 0, 'expired'], ['2026-09-27T11:59:59Z', -1, 'expired'],
  ] as const)('derived whole days for %s preserve the existing threshold', (expires, days, state) => {
    const { target, security } = projections({ ...certificate, cert_not_after: expires })
    expect(target.certificateDays).toBe(days)
    expect(security.certificate.expiry).toMatchObject({ days, state })
    expect(security.certificate.validity.find(item => item.key === 'cert_days_left')?.value).toBe(String(days))
    expect(security.overview.attention.some(item => item.key === 'expiry')).toBe(state !== 'normal')
  })

  it('uses timezone offsets and HTTP evidence time, never the client clock', () => {
    const clock = vi.spyOn(Date, 'now').mockImplementation(() => { throw new Error('Client clock must not be read') })
    const payload = { ...certificate, cert_not_after: '2026-11-11T20:00:00+08:00' }
    expect(projections(payload, '2026-09-27T05:00:00-07:00').security.certificate.days).toBe(45)
    expect(clock).not.toHaveBeenCalled()
  })

  it.each([undefined, null, '', 'invalid', '0001-01-01T00:00:00Z', '2026-09-27T12:00:00', '2026-13-01T00:00:00Z', '2026-02-30T00:00:00Z', '2026-09-27T24:00:00Z'])('missing/invalid time %s keeps validity not observed', invalid => {
    expect(presentSiteCertificateExpiry(certificate, t, invalid)).toMatchObject({ state: 'not_observed', days: null })
    for (const [expires, observed] of [[invalid, observedAt], [certificate.cert_not_after, invalid]]) {
      const { target, security } = projections({ ...certificate, cert_not_after: expires }, observed === undefined ? null : observed)
      expect(target.certificateDays).toBeNull()
      expect(target.health.find(item => item.key === 'certificate')).toMatchObject({ value: 'Not observed', tone: 'neutral' })
      expect(security.certificate.expiry).toMatchObject({ state: 'not_observed', days: null })
      expect(security.certificate.validity.find(item => item.key === 'cert_days_left')?.value).toBe('—')
      expect(security.certificate.verification.state).toBe('verified')
    }
  })

  it.each([{ cert_collected: false }, { tls_handshake: 'not_tls' }])('does not derive days from uncollected defaults: %j', marker => {
    const { target, security } = projections({ ...certificate, ...marker })
    expect(target.certificateDays).toBeNull()
    expect(security.certificate.expiry.state).toBe('not_observed')
  })
})
