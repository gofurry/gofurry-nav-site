import { describe, expect, it } from 'vitest'
import { overviewGameVisuals, overviewSiteVisuals } from '../../app/components/insights/overview/visuals'
import type { InsightFeaturedVisual, InsightOverview, InsightSiteVisual } from '../../app/types/insights'

const game = (id: number): InsightFeaturedVisual => ({ game_id: id, name: ' 作品 ', name_en: '', visual: { kind: 'game_header', asset: `https://shared.steamstatic.com/steam/apps/${id}/header.jpg` } })
const site = (id: number): InsightSiteVisual => ({ site_id: id, name: ' 网站 ', visual: { kind: 'site_icon', asset: `nav/sites/${id}/icon/${'a'.repeat(32)}.png` } })
const legacy: InsightOverview = { generated_at: '2026-10-09T00:00:00Z', entity_count: 0, changes_7d: 0, metrics: [], recent_changes: [] }

describe('Overview dedicated visual projections', () => {
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
