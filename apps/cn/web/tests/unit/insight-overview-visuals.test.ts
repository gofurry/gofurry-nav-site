import { describe, expect, it } from 'vitest'
import { overviewGameVisuals, overviewHeroVisual, overviewGamePreviews, overviewSiteVisuals } from '../../app/components/insights/overview/visuals'
import type { InsightFeaturedVisual, InsightOverview, InsightSiteVisual } from '../../app/types/insights'

const game = (id: number): InsightFeaturedVisual => ({ game_id: id, name: ' 作品 ', name_en: '', visual: { kind: 'game_header', asset: `https://shared.steamstatic.com/steam/apps/${id}/header.jpg` } })
const site = (id: number): InsightSiteVisual => ({ site_id: id, name: ' 网站 ', visual: { kind: 'site_icon', asset: `nav/sites/${id}/icon/${'a'.repeat(32)}.png` } })
const legacy: InsightOverview = { generated_at: '2026-10-09T00:00:00Z', entity_count: 0, changes_7d: 0, metrics: [], recent_changes: [] }

describe('Overview dedicated visual projections', () => {
  const observed = (key: 'free' | 'windows' | 'mac' | 'linux' = 'free'): InsightOverview => ({ ...legacy,
    metrics: [{ key, value: 0, as_of: '2026-10-07', delta_30d: null, coverage: 1, known: 1, eligible: 1, available_from: null }],
    featured_visuals: [game(1), game(2), game(3)],
    hero_visual: { ...game(3), metric_key: key, as_of: '2026-10-07' },
  })
  it.each(['free', 'windows', 'mac', 'linux'] as const)('requires the selected %s metric and its own date, including a real zero', key => {
    const overview = observed(key)
    expect(overviewHeroVisual(overview)?.game_id).toBe(3)
    expect(overviewGamePreviews(overview).map(item => item.game_id)).toEqual([1, 2])
    overview.metrics.unshift({ ...overview.metrics[0]!, key: 'free', value: null })
    expect(overviewHeroVisual(overview)?.metric_key).toBe(key)
    overview.metrics[1]!.as_of = ''
    expect(overviewHeroVisual(overview)).toBeNull()
  })
  it('keeps ordinary previews when old, absent or mismatched proof cannot authorize Hero', () => {
    for (const proof of [undefined, null, {}, { ...observed().hero_visual, metric_key: 'windows' }, { ...observed().hero_visual, as_of: legacy.generated_at.slice(0, 10) }, { ...observed().hero_visual, as_of: '2026-02-30' }, { ...observed().hero_visual, visual: { kind: 'site_icon', asset: game(3).visual.asset } }, { ...observed().hero_visual, visual: { kind: 'game_header', asset: 'https://evil.test/header.jpg' } }]) {
      const overview = { ...observed(), hero_visual: proof } as InsightOverview
      expect(overviewHeroVisual(overview)).toBeNull()
      expect(overviewGamePreviews(overview).map(item => item.game_id)).toEqual([1, 2])
    }
    for (const value of [null, -1, 2, NaN, Infinity]) {
      const overview = observed()
      overview.metrics[0]!.value = value
      expect(overviewHeroVisual(overview)).toBeNull()
    }
    const overview = observed()
    overview.hero_visual = { ...game(1), metric_key: 'free', as_of: '2026-10-07' }
    expect(overviewGamePreviews(overview).map(item => item.game_id)).toEqual([2, 3])
    overview.hero_visual = null
    expect(overviewGamePreviews(overview).map(item => item.game_id)).toEqual([1, 2])
  })
  it('accepts old/empty responses without promoting event art', () => {
    expect(overviewGameVisuals(undefined)).toEqual([])
    expect(overviewSiteVisuals(legacy)).toEqual([])
    expect(overviewGameVisuals([])).toEqual([])
    expect(overviewSiteVisuals({ ...legacy, site_visuals: [] })).toEqual([])
  })
  it('keeps daily order, unique games, names and the three-item bound', () => {
    for (let n = 0; n <= 5; n++) expect(overviewGameVisuals(Array.from({ length: n }, (_, i) => game(i + 1)))).toHaveLength(Math.min(n, 3))
    expect(overviewGameVisuals([game(3), game(3), game(1), game(2)]).map(item => item.game_id)).toEqual([3, 1, 2])
    expect(overviewGameVisuals([game(1)])[0]?.name_en).toBe('')
    expect(overviewGameVisuals([{ ...game(1), name: '', name_en: ' English ' }])[0]?.name).toBe('English')
  })
  it('rejects malformed IDs, names, kinds and untrusted URLs', () => {
    for (const item of [null, {}, { ...game(1), game_id: 0 }, { ...game(1), game_id: 1.5 }, { ...game(1), game_id: Number.MAX_SAFE_INTEGER + 1 }, { ...game(1), name: ' ' }, { ...game(1), visual: { kind: 'site_icon', asset: game(1).visual.asset } }]) expect(overviewGameVisuals([item])).toEqual([])
    for (const asset of ['', 'http://shared.steamstatic.com/steam/apps/1/header.jpg', 'https://shared.steamstatic.com.evil.test/steam/apps/1/header.jpg', 'https://shared.steamstatic.com/header.jpg', game(1).visual.asset + '\n', game(1).visual.asset + '#x']) expect(overviewGameVisuals([{ ...game(1), visual: { kind: 'game_header', asset } }])).toEqual([])
  })
  it('requires site membership, own key, event order and at most five', () => {
    const overview: InsightOverview = { ...legacy, site_visuals: Array.from({ length: 9 }, (_, i) => site(i + 1)), recent_changes: [7, 4, 4, 5, 2, 1, 3].map(id => ({ type: 'site.ipv6.enabled', date: '2026-10-09', occurred_at: null, detail: null, entity: { id, name: 'Site' } })) }
    const before = JSON.stringify(overview)
    expect(overviewSiteVisuals(overview).map(item => item.site_id)).toEqual([7, 4, 5, 2, 1])
    expect(JSON.stringify(overview)).toBe(before)
    for (const item of [site(9), { ...site(7), name: '' }, { ...site(7), visual: site(4).visual }, { ...site(7), visual: { kind: 'site_icon' as const, asset: 'https://example.test/icon.png' } }]) expect(overviewSiteVisuals({ ...overview, site_visuals: [item] })).toEqual([])
  })
})
