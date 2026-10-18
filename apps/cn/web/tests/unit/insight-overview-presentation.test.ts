import { describe, expect, it } from 'vitest'
import type { GameV2ListItem, GameV2PanelRecord } from '../../app/types/game'
import type { InsightFeedItem, InsightMetric, InsightOverview } from '../../app/types/insights'
import { auditOverviewGameCover, formatOverviewPercentagePoints, overviewActivitySummary, overviewActivityWithoutGameArt, overviewDirectoryGroups, overviewEntityCount, overviewFactDate, overviewHeroGameKeys, overviewHeroSiteKeys, overviewPlayerObservation, overviewRatio, overviewSample, selectOverviewEcosystemMetrics, selectOverviewMetric, selectOverviewPulse } from '../../app/utils/insightOverviewPresentation'
import { overviewActivity, overviewExploreGroups } from '../../app/utils/insightOverview'

const now = Date.parse('2026-09-04T09:00:00Z')
const proposedWindow = 72 * 60 * 60 * 1000
const game = (extra: Partial<GameV2ListItem> = {}) => ({
  id: '91', name: 'Observation', tags: [], header_url: 'https://example.test/header.jpg', capsule_url: '',
  online_count: { status: 'success', count: 0, collected_at: '2026-09-01T09:00:00Z' }, ...extra,
}) as GameV2ListItem
const metric = (extra: Partial<InsightMetric> = {}): InsightMetric => ({
  key: 'ipv6', value: 0, coverage: .8, known: 8, eligible: 10, as_of: '2026-09-01', delta_30d: .042, available_from: null, ...extra,
})
const overview = (metrics: InsightMetric[]): InsightOverview => ({ metrics, entity_count: 10, changes_7d: 0, generated_at: '2026-09-02T00:00:00Z', recent_changes: [] })

describe('Overview activity and directory ownership', () => {
  it('sums only two complete reliable sources, keeping real zero and independent availability', () => {
    const nav = { ...overview([]), changes_7d: 20 }, game = { ...overview([]), changes_7d: 27 }
    expect(overviewActivitySummary(nav, game)).toEqual({ availability: 'complete', total: 47 })
    expect(overviewActivitySummary(overview([]), overview([]))).toEqual({ availability: 'complete', total: 0 })
    expect(overviewActivitySummary(nav, null)).toEqual({ availability: 'site-only', total: null })
    expect(overviewActivitySummary(null, game)).toEqual({ availability: 'game-only', total: null })
    expect(overviewActivitySummary(null, null)).toEqual({ availability: 'unavailable', total: null })
    for (const changes_7d of [NaN, Infinity, -1, .5, Number.MAX_SAFE_INTEGER]) {
      expect(overviewActivitySummary({ ...nav, changes_7d }, game).total).toBeNull()
      expect(overviewActivitySummary(nav, { ...game, changes_7d }).total).toBeNull()
    }
  })
  it('keeps the newest five real events and their original day/exact precision without mutating either feed', () => {
    const event = (id: number, date: string, occurred_at: string | null) => ({ entity: { id, name: `Entity ${id}` }, type: 'unknown', date, occurred_at, detail: null })
    const nav = { ...overview([]), recent_changes: [event(1, '2026-09-01', null), event(2, '2026-08-31', null)] }
    const game = { ...overview([]), recent_changes: [event(3, '2026-09-01', '2026-09-01T01:00:00Z'), event(4, '2026-09-02', null), event(5, '2026-08-30', null), event(6, '2026-09-01', '2026-09-01T02:00:00Z')] }
    const before = JSON.stringify([nav, game])
    const items = overviewActivity(nav, game)
    expect(items.map(item => item.entity.id)).toEqual([4, 6, 3, 1, 2])
    expect(items.find(item => item.entity.id === 1)).toMatchObject({ date: '2026-09-01', occurred_at: null, domain: 'site' })
    expect(items.find(item => item.entity.id === 3)?.occurred_at).toBe('2026-09-01T01:00:00Z')
    expect(overviewActivity(nav, null)).toHaveLength(2)
    expect(overviewActivity(null, null)).toEqual([])
    expect(JSON.stringify([nav, game])).toBe(before)
  })
  it('derives six homepage topics without changing the shared Domain destinations', () => {
    expect(Object.keys(overviewDirectoryGroups)).toEqual(['game', 'site'])
    expect(overviewDirectoryGroups.game.map(item => item.path)).toEqual(['/insights/games/players', '/insights/games/prices', '/insights/games/languages', '/insights/games/compare'])
    expect(overviewDirectoryGroups.site.map(item => item.path)).toEqual(['/insights/sites/certificates', '/insights/sites/compare'])
    expect(overviewExploreGroups.game.map(item => item.path)).toEqual(['/insights/games', ...overviewDirectoryGroups.game.map(item => item.path)])
    expect(overviewExploreGroups.site.map(item => item.path)).toEqual(['/insights/sites', ...overviewDirectoryGroups.site.map(item => item.path)])
  })
})

describe('Overview dual ecosystem projection', () => {
  it('selects at most three game metrics in stable priority, keeping zero and skipping unknown keys', () => {
    const data = overview([metric({ key: 'linux', value: .3 }), metric({ key: 'mac', value: .4 }), metric({ key: 'windows', value: 0 }), metric({ key: 'free', value: 0 }), metric()])
    const before = JSON.stringify(data)
    expect(selectOverviewEcosystemMetrics(data, 'game').map(item => item.key)).toEqual(['windows', 'mac', 'linux'])
    expect(selectOverviewEcosystemMetrics(data, 'game')[0]?.value).toBe(0)
    expect(JSON.stringify(data)).toBe(before)
  })
  it('fills missing game metrics from the existing catalog without inventing a minimum count', () => {
    const data = overview([metric({ key: 'free', value: null }), metric({ key: 'windows', value: NaN }), metric({ key: 'mac', value: 0 }), metric({ key: 'linux', value: .3 })])
    expect(selectOverviewMetric(data, overviewHeroGameKeys)?.key).toBe('mac')
    expect(selectOverviewEcosystemMetrics(data, 'game').map(item => item.key)).toEqual(['linux'])
    expect(selectOverviewEcosystemMetrics(overview([metric({ key: 'linux' })]), 'game')).toEqual([])
  })
  it.each(['game', 'site'] as const)('excludes the exact %s Hero for every fallback, keeps zero and deduplicates source keys', domain => {
    const keys = domain === 'game' ? overviewHeroGameKeys : overviewHeroSiteKeys
    for (let selected = 0; selected < keys.length; selected++) {
      const rows = keys.map((key, index) => metric({ key, value: index < selected ? null : 0 }))
      const data = overview([...rows].reverse().concat(rows))
      expect(selectOverviewMetric(data, keys)?.key).toBe(keys[selected])
      const result = selectOverviewEcosystemMetrics(data, domain)
      expect(result.map(item => item.key)).toEqual(keys.slice(selected + 1, selected + 4))
      expect(result.every(item => item.value === 0)).toBe(true)
    }
  })
  it('preserves each metric date and coverage, without borrowing the Hero or generated time', () => {
    const data = overview([metric(), metric({ key: 'tls13', as_of: '2026-08-30', coverage: .75 }), metric({ key: 'http2', as_of: '', coverage: 2 })])
    const result = selectOverviewEcosystemMetrics(data, 'site')
    expect(result[0]).toMatchObject({ key: 'tls13', as_of: '2026-08-30', coverage: .75 })
    expect(result[1]).toMatchObject({ key: 'http2', as_of: null, coverage: null })
  })
  it('uses empty lists for unavailable or invalid metrics without manufacturing zero', () => {
    for (const domain of ['game', 'site'] as const) {
      expect(selectOverviewEcosystemMetrics(null, domain)).toEqual([])
      expect(selectOverviewEcosystemMetrics(overview([]), domain)).toEqual([])
      expect(selectOverviewEcosystemMetrics(overview([metric({ key: domain === 'game' ? 'free' : 'ipv6', value: -1 }), metric({ key: domain === 'game' ? 'linux' : 'http2', value: Infinity })]), domain)).toEqual([])
    }
  })
  it('keeps real zero entity counts and rejects missing, fractional or negative counts', () => {
    expect(overviewEntityCount(0)).toBe(0)
    expect(overviewEntityCount(213)).toBe(213)
    for (const value of [null, undefined, -1, .5, NaN, Infinity, '213', Number.MAX_SAFE_INTEGER + 1]) expect(overviewEntityCount(value)).toBeNull()
  })
})

describe('Overview metric evidence', () => {
  it('selects in stable configured order, keeps zero, and rejects invalid ratios', () => {
    const data = overview([metric({ key: 'tls13', value: .4 }), metric()])
    expect(selectOverviewMetric(data, overviewHeroSiteKeys)?.key).toBe('ipv6')
    expect(selectOverviewMetric(data, overviewHeroSiteKeys)?.value).toBe(0)
    for (const value of [NaN, Infinity, -1, 1.1, null]) {
      expect(selectOverviewMetric(overview([metric({ value }), metric({ key: 'tls13', value: .4 })]), overviewHeroSiteKeys)?.key).toBe('tls13')
    }
    expect(selectOverviewMetric(data, overviewHeroGameKeys)).toBeNull()
    expect(selectOverviewMetric(null, overviewHeroSiteKeys)).toBeNull()
    expect(overviewRatio('0')).toBeNull()
  })
  it('never substitutes generated_at for missing or invalid fact dates and coverage', () => {
    for (const as_of of ['', '2026-02-30', '2026-09-01T00:00:00Z']) {
      expect(selectOverviewMetric(overview([metric({ as_of, coverage: NaN })]), overviewHeroSiteKeys)).toMatchObject({ as_of: null, coverage: null })
    }
    expect(overviewFactDate('2024-02-29')).toBe('2024-02-29')
    expect(overviewSample(metric())).toBe('8 / 10')
    expect(overviewSample(metric({ known: 0, eligible: 0 }))).toBe('0 / 0')
    expect(overviewSample(metric({ known: 11 }))).toBeNull()
  })
  it.each(['zh', 'en'])('formats signed percentage-point differences in %s', locale => {
    expect(formatOverviewPercentagePoints(.042, locale)).toBe('+4.2')
    expect(formatOverviewPercentagePoints(-.011, locale)).toBe('-1.1')
    expect(formatOverviewPercentagePoints(0, locale)).toBe('0.0')
    expect(formatOverviewPercentagePoints(-.00001, locale)).toBe('0.0')
    for (const value of [null, undefined, NaN, Infinity, 2]) expect(formatOverviewPercentagePoints(value, locale)).toBeNull()
  })
})

describe('Overview observation and cover qualification', () => {
  it('uses the injected clock, keeps successful zero, and does not assume a freshness window', () => {
    expect(overviewPlayerObservation(game(), now)).toMatchObject({ count: 0, freshness: 'unverified' })
    expect(overviewPlayerObservation(game(), now, proposedWindow)?.freshness).toBe('fresh')
    expect(overviewPlayerObservation(game(), now - proposedWindow, proposedWindow)?.freshness).toBe('fresh')
    expect(overviewPlayerObservation(game(), now)?.label).toBe('2026-09-01 09:00:00 UTC')
    expect(overviewPlayerObservation(game(), now + 1, proposedWindow)?.freshness).toBe('stale')
    expect(overviewPlayerObservation(game(), now - proposedWindow - 1, proposedWindow)).toBeNull()
    expect(overviewPlayerObservation(game(), NaN, proposedWindow)).toBeNull()
  })
  it.each([-1, NaN, Infinity])('rejects invalid player count %s', count => {
    expect(overviewPlayerObservation(game({ online_count: { ...game().online_count, count } }), now)).toBeNull()
  })
  it.each(['', 'invalid', '2026-09-01', '2026-02-30T00:00:00Z', '2026-09-01T09:00:00'])('rejects missing, ambiguous or invalid observation time %s', collected_at => {
    expect(overviewPlayerObservation(game({ online_count: { ...game().online_count, collected_at } }), now)).toBeNull()
  })
  it('requires success and retains precise timestamps without timezone assumptions', () => {
    expect(overviewPlayerObservation(game({ online_count: { ...game().online_count, status: 'failed' } }), now)).toBeNull()
    expect(overviewPlayerObservation(game({ online_count: { ...game().online_count, collected_at: '2026-09-01T17:00:00.123456+08:00' } }), now)?.collectedAt).toBe('2026-09-01T09:00:00.123Z')
  })
  it('rejects fresh images with empty, missing, non-adult or adult tags and uncontracted sfw flags', () => {
    const nonAdult = [{ code: 'adventure' }] as GameV2ListItem['tags']
    const adult = [{ code: 'adult' }] as GameV2ListItem['tags']
    for (const tags of [undefined, [], nonAdult, adult]) {
      const candidate = { ...game({ tags }), sfw: true }
      expect(auditOverviewGameCover(candidate, now, proposedWindow)).toMatchObject({ coverEligible: false, hasImage: true, observation: { freshness: 'fresh' } })
    }
    expect(auditOverviewGameCover(game({ tags: adult }), now).safety).toBe('adult')
    expect(auditOverviewGameCover(game({ header_url: '', capsule_url: '' }), now).hasImage).toBe(false)
  })
  it('preserves deterministic pulse deduplication and does not mutate the shared panel', () => {
    const one = game(), two = game({ id: '92', prices: [{ region: 'US', available: true, is_free: false, currency: 'USD', final_amount: 1, discount_percent: 50 }] as GameV2ListItem['prices'] }), three = game({ id: '93' })
    const panel = { top_online: [game({ id: 'bad', online_count: { ...one.online_count, count: -1 } }), one], highest_discount: [one, two], latest_games: [one, two, three] } as GameV2PanelRecord
    const before = JSON.stringify(panel)
    const result = selectOverviewPulse(panel, now)
    expect([result.players?.id, result.discount?.id, result.latest?.id]).toEqual(['91', '92', '93'])
    expect(JSON.stringify(panel)).toBe(before)
    expect(selectOverviewPulse(null, now).players).toBeNull()
  })
  it('removes only homepage game art while retaining event identity and dates', () => {
    const item: InsightFeedItem = { domain: 'game', type: 'game.windows.added', date: '2026-09-01', occurred_at: null, detail: null, entity: { id: 91, name: 'Game', visual: { kind: 'game_header', asset: 'https://example.test/game.jpg' } } }
    expect(overviewActivityWithoutGameArt(item)).toEqual({ ...item, entity: { id: 91, name: 'Game' } })
    expect(item.entity.visual?.asset).toBeTruthy()
    expect(overviewActivityWithoutGameArt({ ...item, domain: 'site' }).entity.visual).toEqual(item.entity.visual)
  })
})
