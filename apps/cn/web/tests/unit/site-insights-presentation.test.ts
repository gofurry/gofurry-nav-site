import { describe, expect, it } from 'vitest'
import { presentSiteInsights, siteInsightsState, siteInsightAdoptionValues } from '../../app/utils/siteInsightsPresentation'
import { siteCapabilityKeys, siteCapabilityRegistry } from '../../app/utils/siteCapabilityRegistry'
import { buildSiteDetailQuery, parseSiteDetailRouteState, selectSiteInsightMetric, selectSiteInsightRange } from '../../app/utils/siteDetailRouteState'
import { siteInsightChangeCategory } from '../../app/utils/insightChanges'
import type { SiteInsights, SiteInsightCapabilityState } from '../../app/types/insights'
import en from '../../i18n/locales/en.json'
import zh from '../../i18n/locales/zh.json'
const t = (key: string) => key.split('.').reduce<unknown>((value, part) => (value as Record<string, unknown>)[part], en) as string
const source = (): SiteInsights => ({ site: { id: 41, name: 'Site' }, capabilities: [], recent_changes: [] })
const present = (data: SiteInsights | null, metric = siteCapabilityKeys[0]!) => presentSiteInsights(data, siteInsightsState(data), metric, t, 'en')
describe('Site intelligence facts', () => {
  it('projects all seven grouped registry rows, keeping a successful empty slice as missing facts', () => {
    const result = present(source())
    expect(result.state).toBe('empty')
    expect(result.rows.map(row => row.key)).toEqual(['ipv6', 'http2', 'tls13', 'certificate_verified', 'hsts', 'csp', 'security_txt'])
    expect(result.rows.every(row => row.state === 'missing' && row.stateLabel === 'No fact available')).toBe(true)
    expect(result.rows.map(row => row.category)).toEqual(['network', 'network', 'transport', 'transport', 'web_policy', 'web_policy', 'web_policy'])
    expect(result.rows).toHaveLength(siteCapabilityRegistry.length)
    expect(result.groups.map(group => [group.key, group.rows.length])).toEqual([['network', 2], ['transport', 2], ['web_policy', 3]])
  })
  it('never fabricates backend unavailable facts when the whole slice failed', () => {
    const result = present(null)
    expect(result.state).toBe('unavailable'); expect(result.changesState).toBe('unavailable')
    expect(result.rows.every(row => row.state === null && row.stateLabel === '—' && row.adoption === '—')).toBe(true)
  })
  it.each(['supported', 'unsupported', 'stale', 'not_probed', 'unavailable', 'unknown', 'not_applicable'] as SiteInsightCapabilityState[])('retains explicit %s without applying it to missing facts', state => {
    const data = source(); data.capabilities = [{ key: 'tls13', state, as_of: '2026-09-27', ecosystem: { value: 0, coverage: null } }]
    const result = present(data, 'tls13')
    expect(result.state).toBe('ready'); expect(result.selected).toMatchObject({ key: 'tls13', state, date: '2026-09-27', adoption: '0.0%', coverage: '—' })
    expect(result.rows.find(row => row.key === 'ipv6')?.state).toBe('missing')
  })
  it('keeps adoption and coverage separate and does not coerce null to zero', () => {
    const data = source(); data.capabilities = [{ key: 'ipv6', state: 'unknown', as_of: null, ecosystem: { value: null, coverage: 0 } }]
    expect(present(data).selected).toMatchObject({ adoption: '—', coverage: '0.0%', date: '—' })
  })
  it('retains the full sorted recent set, event fallback and day/exact precision', () => {
    const data = source()
    data.recent_changes = ['site.ipv6.enabled', 'site.primary_target.changed', 'site.tls_certificate.changed', 'site.unknown', 'site.http2.enabled', 'site.csp.added'].map((type, index) => ({
      type, date: `2026-09-2${index}`, occurred_at: index === 5 ? '2026-09-25T12:34:00Z' : null, entity: data.site, detail: null,
    }))
    const result = present(data)
    expect(result.state).toBe('ready'); expect(result.changes).toHaveLength(6)
    expect(result.changes[0]).toMatchObject({ category: 'capability', when: 'Sep 25, 2026, 12:34 PM UTC', precise: true })
    expect(result.changes[1]).toMatchObject({ when: '2026-09-24', precise: false })
    expect(result.changes[2]?.category).toBe('unknown')
    expect(result.changes[2]?.label).toBe(t('insights.changes.events.unknown'))
    expect(result.changes.map(item => item.category)).toContain('target')
    expect(result.changes.map(item => item.category)).toContain('certificate')
  })
  it.each([['site.hsts.removed', 'capability'], ['site.primary_target.changed', 'target'], ['site.tls_certificate.verification_failed', 'certificate'], ['site.future.enabled', null], ['game.free.enabled', null]])('maps public event %s to its presentation category', (event, category) => {
    expect(siteInsightChangeCategory(event!)).toBe(category)
  })
  it('plots only adoption and keeps true gaps, zero and nonfinite data distinct', () => {
    expect(siteInsightAdoptionValues([null, 0, .5, NaN].map(value => ({ date: '2026-09-27', value, coverage: .9 })))).toEqual([null, 0, 50, null])
  })
  it('has complete bilingual labels for all new UI copy', () => {
    function inspect(a: unknown, b: unknown) {
      if (typeof a === 'string') { expect(typeof b).toBe('string'); expect(b).not.toBe(''); return }
      for (const [key, child] of Object.entries(a as Record<string, unknown>)) inspect(child, (b as Record<string, unknown>)[key])
    }
    inspect(en.siteIntelligence, zh.siteIntelligence)
  })
})
describe('URL-owned Site insight selection', () => {
  it('preserves Target/range on metric selection and Target/metric on range selection', () => {
    const state = parseSiteDetailRouteState({ domain: 'a.example', tab: 'insights', metric: 'tls13', range: '90d' })
    expect(buildSiteDetailQuery(selectSiteInsightMetric(state, 'csp'))).toEqual({ domain: 'a.example', tab: 'insights', metric: 'csp', range: '90d' })
    expect(buildSiteDetailQuery(selectSiteInsightRange(state, 'all'))).toEqual({ domain: 'a.example', tab: 'insights', metric: 'tls13', range: 'all' })
  })
  it('clears foreign view state and omits default metric/range', () => {
    const state = parseSiteDetailRouteState({ domain: 'a.example', tab: 'security', view: 'tls' })
    expect(buildSiteDetailQuery(selectSiteInsightMetric(state, 'ipv6'))).toEqual({ domain: 'a.example', tab: 'insights' })
    expect(buildSiteDetailQuery(selectSiteInsightRange(state, '30d'))).toEqual({ domain: 'a.example', tab: 'insights' })
  })
})
