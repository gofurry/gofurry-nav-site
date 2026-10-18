import { describe, expect, it } from 'vitest'
import { siteDnsSignal, siteHealthReasonLabel, siteHttpTone, siteLatencyTone } from '../../app/utils/siteDetailPresentation'
import { presentSiteCertificateExpiry, siteSecurityTxtValidationLabel } from '../../app/utils/siteSecurityPresentation'
import { presentSiteTarget } from '../../app/utils/siteTargetPresentation'
import type { TargetLatestResponse } from '../../app/types/nav'
import en from '../../i18n/locales/en.json'
import zh from '../../i18n/locales/zh.json'

const translate = (locale: 'en' | 'zh' = 'en') => (key: string, values: Record<string, string | number> = {}) => {
  const value = key.split('.').reduce<unknown>((value, part) => (value as Record<string, unknown>)[part], locale === 'en' ? en : zh)
  if (typeof value !== 'string') throw new Error('Missing translation ' + key)
  return value.replace(/\{(\w+)\}/g, (_, key: string) => String(values[key] ?? '{' + key + '}'))
}
function strip(payload: Record<string, unknown>) {
  return presentSiteTarget({ domain: 'a.example', siteHealthSummary: null, targetHealthSummary: null,
    targetLatestCore: { target: 'a.example', protocols: { http: { target: 'a.example', payload } } } as TargetLatestResponse }, translate()).health
}

describe('Site Detail refinement evidence presentation', () => {
  it.each([[null, 'neutral'], [-1, 'neutral'], [NaN, 'neutral'], [0, 'good'], [800, 'good'], [801, 'warning'], [2000, 'warning'], [2001, 'bad']] as const)('response latency %s is %s', (value, tone) => {
    expect(siteLatencyTone(value)).toBe(tone)
  })
  it.each([[null, 'neutral'], [200, 'good'], [299, 'good'], [301, 'neutral'], [404, 'bad'], [503, 'bad'], [999, 'neutral']] as const)('HTTP %s is %s', (value, tone) => {
    expect(siteHttpTone(value)).toBe(tone)
  })
  it.each([[31, 'normal', 'good'], [30, 'attention', 'warning'], [8, 'attention', 'warning'], [7, 'warning', 'bad'], [1, 'warning', 'bad'], [0, 'expired', 'bad'], [-1, 'expired', 'bad']] as const)('shares certificate %s days between Security and Health Strip', (days, state, tone) => {
    const payload = { cert_collected: true, cert_days_left: days, cert_verified: true }
    const expiry = presentSiteCertificateExpiry(payload, translate())
    expect(expiry).toMatchObject({ state, tone, days })
    expect(strip(payload).find(item => item.key === 'certificate')).toMatchObject({ value: days <= 0 ? 'Expired' : `${days} days left`, tone: expiry.tone })
  })
  it.each([{}, { cert_collected: false, cert_days_left: 0 }, { tls_handshake: 'not_tls', cert_days_left: 0 }])('never substitutes default zero or a dash for unobserved certificate: %j', payload => {
    expect(presentSiteCertificateExpiry(payload, translate())).toMatchObject({ days: null, value: 'Not observed', tone: 'neutral' })
    expect(strip(payload).find(item => item.key === 'certificate')?.value).toBe('Not observed')
  })
  it.each(['TLS 1.2', 'TLS 1.3', 'TLS1.3'])('keeps successful %s positive without downgrading TLS 1.2', tls_version => {
    expect(strip({ tls_version, tls_handshake: 'collected' }).find(item => item.key === 'tls')?.tone).toBe('good')
    expect(strip({ tls_version, tls_handshake: 'failed' }).find(item => item.key === 'tls')).toMatchObject({ tone: 'bad', value: 'Failed' })
  })
  it.each(['en', 'zh'] as const)('has explicit known-code copy and preserves unknown fallback in %s', locale => {
    const t = translate(locale)
    for (const code of Object.keys(en.siteDetail.reasons)) {
      expect(siteHealthReasonLabel(code, t)).toBe(t('siteDetail.reasons.' + code))
      expect(siteHealthReasonLabel(code, t)).not.toBe(code)
    }
    expect(siteHealthReasonLabel('future_reason', t)).toContain('future_reason')
    for (const code of Object.keys(en.siteSecurity.validation)) expect(siteSecurityTxtValidationLabel(code, t)).toBe(t('siteSecurity.validation.' + code))
    expect(siteSecurityTxtValidationLabel('future_validation', t)).toBe('future_validation')
  })
  it('distinguishes raw diagnostic DNS evidence without inventing a health verdict', () => {
    for (const code of ['ptr_empty', 'low_ttl', 'other', 'future_flag']) expect(siteDnsSignal(code, translate())).toMatchObject({ code, tone: 'neutral', kind: 'Informational' })
    for (const code of ['private_ip', 'nxdomain_with_answer']) expect(siteDnsSignal(code, translate())).toMatchObject({ code, tone: 'warning', kind: 'Needs attention' })
  })
})
