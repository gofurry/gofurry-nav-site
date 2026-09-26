import { describe, expect, it } from 'vitest'
import { presentSiteOverview } from '../../app/utils/siteOverviewPresentation'
import type { SiteHealthSummary, TargetHealthSummaryItem } from '../../app/types/nav'
import type { InsightChange, SiteInsights, SiteInsightCapabilityState } from '../../app/types/insights'
import en from '../../i18n/locales/en.json'
import zh from '../../i18n/locales/zh.json'

function translator(locale = 'en') {
  return (key: string, values: Record<string, string | number> = {}) => {
    const value = key.split('.').reduce<unknown>((value, key) => (value as Record<string, unknown>)[key], locale === 'en' ? en : zh)
    if (typeof value !== 'string') throw new Error('Missing translation ' + key)
    return value.replace(/\{(\w+)\}/g, (_, key: string) => String(values[key] ?? '{' + key + '}'))
  }
}
const target = (host: string, status: TargetHealthSummaryItem['status'] = 'healthy', extra: Partial<TargetHealthSummaryItem> = {}): TargetHealthSummaryItem => ({
  target: host, status, reason_codes: [], reason_messages: [], observed_at: '2026-09-26T10:00:00Z', ...extra,
})
const summary = (extra: Partial<SiteHealthSummary> = {}): SiteHealthSummary => ({
  state: 'ready', site_id: 41, status: 'healthy', target_count: 2,
  status_counts: { healthy: 2, warning: 0, degraded: 0, down: 0, unknown: 0 },
  targets: [target('a.example'), target('b.example')], reason_messages: [], reason_codes: [],
  generated_at: '2026-09-26T12:18:00Z', schema_version: 1, ...extra,
})
const facts = (extra: Partial<SiteInsights> = {}): SiteInsights => ({
  site: { id: 41, name: 'Site' }, capabilities: [], recent_changes: [], ...extra,
})
const change = (date: string, occurred_at: string | null = null, type = 'site.ipv6.enabled'): InsightChange => ({
  type, date, occurred_at, entity: { id: 41, name: 'Site' }, detail: null,
})
const present = (health: SiteHealthSummary | null = summary(), insights: SiteInsights | null = facts(), unavailable = false, locale = 'en') =>
  presentSiteOverview(health, insights, unavailable, translator(locale), locale)

describe('Site overview health and attention', () => {
  it('uses Site distribution and generated time, with no attention for a healthy Site', () => {
    const vm = present()
    expect(vm.health).toMatchObject({ summaryState: 'ready', status: 'healthy', targetCount: 2, summaryText: '2 / 2 targets healthy', generatedAt: '2026-09-26T12:18:00.000Z' })
    expect(vm.health.generatedLabel).toBe('2026-09-26 12:18:00 UTC')
    expect(vm.health.statusItems.map(item => item.status)).toEqual(['healthy'])
    expect(vm.attention).toEqual([])
  })
  it.each(['warning', 'degraded', 'down', 'unknown'] as const)('retains aggregate %s and omits zero count categories', status => {
    const vm = present(summary({ status, status_counts: { healthy: 1, [status]: 1 } }))
    expect(vm.health.status).toBe(status)
    expect(vm.health.statusItems.map(item => item.count)).toEqual([1, 1])
    expect(vm.health.summaryText).not.toContain('0 ')
    expect(vm.attention).toHaveLength(1)
  })
  it('keeps stale healthy distinct from ready healthy', () => {
    const vm = present(summary({ state: 'stale' }))
    expect(vm.health.statusLabel).toBe('Healthy')
    expect(vm.health.freshnessLabel).toBe('Summary may be stale')
    expect(vm.attention[0]?.message).toContain('may no longer reflect')
  })
  it.each([null, summary({ state: 'missing', status: 'unknown', generated_at: '0001-01-01T00:00:00Z' })])('represents missing summary independently of backend unknown', input => {
    const vm = present(input)
    expect(vm.health.summaryState).toBe('missing')
    expect(vm.health.status).toBeNull()
    expect(vm.health.statusLabel).toBe('No complete site health summary')
    expect(vm.health.generatedAt).toBeNull()
    expect(vm.attention).toHaveLength(1)
    expect(present(summary({ status: 'unknown' })).health.statusLabel).toBe('Unknown')
  })
  it('preserves zero targets without claiming 0/0 targets healthy', () => {
    const vm = present(summary({ target_count: 0, status_counts: {}, targets: [] }))
    expect(vm.health.targetCount).toBe(0)
    expect(vm.health.summaryText).toBe('No collected targets')
    expect(vm.health.statusItems).toEqual([])
  })
  it('prioritizes human messages and deduplicates target-specific fallback and repeated messages', () => {
    const vm = present(summary({
      status: 'degraded', reason_messages: ['b.example has failed', ' b.example has failed '],
      reason_codes: ['raw_code_must_be_hidden'],
      targets: [target('a.example'), target('b.example', 'degraded', { reason_messages: ['b.example has failed'], reason_codes: ['another_raw_code'] })],
    }))
    expect(vm.attention.map(item => item.message)).toEqual(['b.example has failed'])
  })
  it('deduplicates a shared human reason even if the Site message omits the hostname', () => {
    const vm = present(summary({ reason_messages: ['Certificate expired'], targets: [target('b.example', 'warning', { reason_messages: ['Certificate expired'] })] }))
    expect(vm.attention.map(item => item.message)).toEqual(['Certificate expired'])
  })
  it('falls back to reason codes only without human-readable evidence', () => {
    expect(present(summary({ reason_codes: ['probe_failure', 'probe_failure'] })).attention.map(item => item.message)).toEqual(['Reported signal: probe_failure'])
    const vm = present(summary({ reason_codes: ['probe_failure'], targets: [target('b.example', 'down', { reason_messages: ['Connection refused'] })] }))
    expect(vm.attention.map(item => item.message)).toEqual(['b.example · Connection refused'])
  })
  it('retains uncovered affected targets while filtering healthy and unobserved unknown noise', () => {
    const vm = present(summary({ targets: [
      target('healthy.example'), target('down.example', 'down'), target('unknown.example', 'unknown'),
      target('unobserved.example', 'unknown', { observed_at: '0001-01-01T00:00:00Z' }),
      target('stale.example', 'healthy', { reason_codes: ['http_stale'] }),
    ] }))
    expect(vm.attention.map(item => item.target)).toEqual(['down.example', 'unknown.example', 'stale.example'])
  })
  it('does not expose raw target codes alongside a human Site message', () => {
    const vm = present(summary({ reason_messages: ['a.example is degraded'], targets: [
      target('a.example', 'degraded'), target('b.example', 'down', { reason_codes: ['raw_code'] }),
    ] }))
    expect(vm.attention.map(item => item.message)).toEqual(['a.example is degraded', 'b.example · Down'])
  })
})

describe('Site capability snapshot', () => {
  it('emits the complete registry in its categories even for success-empty', () => {
    const vm = present()
    expect(vm.capabilities).toHaveLength(7)
    expect(vm.capabilities.every(item => item.state === 'missing' && item.stateLabel === 'No fact available')).toBe(true)
    expect(vm.capabilityState).toBe('empty')
    expect(vm.changesState).toBe('empty')
    expect(vm.capabilityGroups.map(group => [group.key, group.items.map(item => item.key)])).toEqual([
      ['network', ['ipv6', 'http2']], ['transport', ['tls13', 'certificate_verified']], ['web_policy', ['hsts', 'csp', 'security_txt']],
    ])
  })
  it.each(['supported', 'unsupported', 'stale', 'not_probed', 'unavailable', 'unknown', 'not_applicable'] as SiteInsightCapabilityState[])('retains backend %s separately from absent facts', state => {
    const vm = present(summary(), facts({ capabilities: [{ key: 'ipv6', state, as_of: '2026-09-26', ecosystem: { value: .75, coverage: .9 } }] }))
    expect(vm.capabilities[0]?.state).toBe(state)
    expect(vm.capabilities[1]?.state).toBe('missing')
    expect(vm.capabilities[0]).not.toHaveProperty('ecosystem')
    expect(vm.capabilities[0]?.tone).not.toBe('bad')
    expect(vm.capabilityState).toBe('ready')
  })
  it('optional failure only makes capability/changes unavailable and still yields seven rows', () => {
    const vm = present(summary(), null, true)
    expect(vm.health.status).toBe('healthy')
    expect(vm.attention).toEqual([])
    expect(vm.capabilityState).toBe('unavailable')
    expect(vm.changesState).toBe('unavailable')
    expect(vm.capabilities).toHaveLength(7)
    expect(vm.capabilities.every(item => item.state === 'unavailable')).toBe(true)
  })
})

describe('compact Site changes', () => {
  it('sorts a copy, limits it to four and preserves day-only and exact precision', () => {
    const items = [change('2026-09-21'), change('2026-09-25'), change('2026-09-24', '2026-09-24T12:34:00Z'),
      change('2026-09-22'), change('2026-09-23')]
    const original = [...items]
    const vm = present(summary(), facts({ recent_changes: items }))
    expect(items).toEqual(original)
    expect(vm.recentChanges).toHaveLength(4)
    expect(vm.recentChanges[0]).toMatchObject({ dateTime: '2026-09-25', when: '2026-09-25', precise: false })
    expect(vm.recentChanges[1]).toMatchObject({ dateTime: '2026-09-24T12:34:00Z', when: 'Sep 24, 2026, 12:34 PM', precise: true })
    expect(vm.recentChanges[0]?.label).toBe('enabled IPv6')
    expect(vm.changesState).toBe('ready')
  })
  it('keeps empty and unavailable distinct and uses the existing fallback event copy', () => {
    expect(present().changesState).toBe('empty')
    expect(present(summary(), facts(), true).changesState).toBe('unavailable')
    const vm = present(summary(), facts({ recent_changes: [change('2026-09-25', null, 'future.event')] }))
    expect(vm.recentChanges[0]?.label).toBe(translator()('insights.changes.events.unknown'))
  })
  it('has complete Chinese presentation without affecting fact identity or precision', () => {
    const vm = present(summary(), facts({ recent_changes: [change('2026-09-25')] }), false, 'zh')
    expect(vm.health.summaryText).toBe('2 / 2 个目标健康')
    expect(vm.capabilities[0]?.stateLabel).toBe('暂无事实')
    expect(vm.recentChanges[0]?.when).toBe('2026-09-25')
  })
})
