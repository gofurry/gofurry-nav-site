import { describe, expect, it } from 'vitest'
import { presentSiteCdn, presentSiteProtocolStatus } from '../../app/utils/siteTargetSignals'
import { presentSiteTarget } from '../../app/utils/siteTargetPresentation'
import { presentSiteObservation } from '../../app/utils/siteObservationPresentation'
import type { TargetHealthSummary, TargetLatestResponse } from '../../app/types/nav'
import en from '../../i18n/locales/en.json'
import zh from '../../i18n/locales/zh.json'

const translate = (locale = 'en') => (key: string, values: Record<string, string | number> = {}) => {
  const value = key.split('.').reduce<unknown>((value, part) => (value as Record<string, unknown>)[part], locale === 'en' ? en : zh)
  if (typeof value !== 'string') throw new Error('Missing translation ' + key)
  return value.replace(/\{(\w+)\}/g, (_, key: string) => String(values[key]))
}
const hint = (provider: string, type = 'cdn', confidence = 'medium') => ({ provider, type, confidence })

describe('Current Target CDN hints', () => {
  it('selects high before medium, preserves tie order, and never infers type from provider', () => {
    const result = presentSiteCdn([hint('cloudflare', 'waf', 'high'), hint('tencent_cloud'), hint('fastly', 'cdn', 'high'), hint('aliyun', 'cdn', 'high')], translate())
    expect(result).toEqual({ provider: 'fastly', label: 'Fastly CDN' })
    expect(presentSiteCdn([hint('tencent_cloud'), hint('fastly', 'cdn', 'low')], translate('zh'))?.label).toBe('腾讯云 CDN')
  })
  it.each(['waf', 'hosting', 'reverse_proxy', 'response_header'])('does not turn %s into a CDN badge', type => {
    expect(presentSiteCdn([hint('cloudflare', type, 'high')], translate())).toBeNull()
  })
  it('hides absent, low and unrecognized confidence hints', () => {
    expect(presentSiteCdn([], translate())).toBeNull()
    expect(presentSiteCdn([hint('fastly', 'cdn', 'low'), hint('cloudflare', 'cdn', 'unknown'), hint(' ', 'cdn', 'high')], translate())).toBeNull()
  })
  it.each(['en', 'zh'])('localizes known CDN names and preserves a typed unknown provider in %s', locale => {
    for (const provider of ['cloudflare', 'tencent_cloud', 'aliyun', 'aws_cloudfront', 'fastly']) {
      expect(presentSiteCdn([hint(provider)], translate(locale))?.label).not.toContain('siteDetail.')
    }
    expect(presentSiteCdn([hint('Example Edge')], translate(locale))?.label).toBe('Example Edge CDN')
  })
})

describe('One protocol status presentation', () => {
  it.each([
    ['success', false, 'success', 'good', true], ['success', true, 'stale', 'warning', false],
    ['failure', false, 'failure', 'bad', false], ['skipped', false, 'skipped', 'muted', false],
    [undefined, false, 'unknown', 'muted', false], ['new_status', false, 'unknown', 'muted', false],
  ] as const)('projects %s / stale=%s', (value, stale, status, tone, success) => {
    expect(presentSiteProtocolStatus(value, stale, translate())).toMatchObject({ status, tone, success })
  })
  it.each(['success', 'failure', 'unknown', 'stale'])('keeps Overview and Context aligned for %s without coloring slow success green', status => {
    const source = { domain: 'a.example', siteHealthSummary: null, lightProbeState: null,
      targetHealthSummary: { target: 'a.example', state: 'ready', protocols: { ping: { status: status === 'stale' ? 'success' : status, stale: status === 'stale', duration_ms: 4302 } } } as TargetHealthSummary,
      targetLatestCore: { target: 'a.example', state: 'ready', protocols: {} } as TargetLatestResponse }
    const context = presentSiteTarget(source, translate()).protocolStates[0]!
    const overview = presentSiteObservation(source, translate()).protocols[0]!
    for (const key of ['status', 'statusLabel', 'tone', 'success', 'duration', 'durationTone'] as const) expect(overview[key]).toEqual(context[key])
    expect(context.durationTone).toBe('bad')
  })
})
