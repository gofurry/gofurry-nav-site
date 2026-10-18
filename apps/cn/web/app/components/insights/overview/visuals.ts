import type { InsightFeaturedVisual, InsightHeroVisual, InsightOverview, InsightSiteVisual } from '@/types/insights'
import { overviewFactDate, overviewHeroGameKeys, overviewRatio, selectOverviewMetric } from '../../../utils/insightOverviewPresentation'

const positiveID = (value: unknown): value is number => typeof value === 'number' && Number.isSafeInteger(value) && value > 0
const record = (value: unknown): value is Record<string, unknown> => typeof value === 'object' && value !== null
const header = /^https:\/\/(?:shared\.steamstatic\.com|shared\.akamai\.steamstatic\.com|cdn\.akamai\.steamstatic\.com|cdn\.cloudflare\.steamstatic\.com|cdn\.steamstatic\.com|steamcdn-a\.akamaihd\.net)\/(?:steam\/apps\/[1-9][0-9]*\/|store_item_assets\/steam\/apps\/[1-9][0-9]*\/(?:[0-9a-f]{40}\/)?)(?:header(?:_alt_assets_[0-9]{1,4})?(?:_schinese)?(?:_2x)?\.jpg)(?:\?t=[0-9]{1,20})?$/
const icon = /^nav\/sites\/([1-9][0-9]*)\/icon\/[a-f0-9]{32}(?:\.[a-z0-9]{1,16})?$/

// These guards validate the dedicated backend projections, never infer image
// eligibility from event artwork, missing adult tags or client-injected flags.
export function overviewGameVisuals(input: unknown): InsightFeaturedVisual[] {
  if (!Array.isArray(input)) return []
  const result: InsightFeaturedVisual[] = [], seen = new Set<number>()
  for (const item of input) {
    if (!record(item) || !positiveID(item.game_id) || seen.has(item.game_id)
      || typeof item.name !== 'string' || typeof item.name_en !== 'string'
      || !record(item.visual) || item.visual.kind !== 'game_header' || typeof item.visual.asset !== 'string'
      || /\s/.test(item.visual.asset) || !header.test(item.visual.asset)) continue
    const name = item.name.trim() || item.name_en.trim()
    if (!name) continue
    seen.add(item.game_id)
    result.push({ game_id: item.game_id, name, name_en: item.name_en.trim(), visual: { kind: 'game_header', asset: item.visual.asset } })
    if (result.length === 3) break
  }
  return result
}

// Ordinary candidates never grant Hero eligibility. The dedicated proof must
// agree with the selected metric's own date, not response generation time.
export function overviewHeroVisual(overview: InsightOverview | null): InsightHeroVisual | null {
  const metric = selectOverviewMetric(overview, overviewHeroGameKeys), proof = overview?.hero_visual
  if (!metric || overviewRatio(metric.value) === null || !record(proof)
    || proof.metric_key !== metric.key || !overviewFactDate(proof.as_of) || proof.as_of !== metric.as_of) return null
  const visual = overviewGameVisuals([proof])[0]
  return visual ? { ...visual, metric_key: proof.metric_key, as_of: proof.as_of } : null
}

export function overviewGamePreviews(overview: InsightOverview | null): InsightFeaturedVisual[] {
  const heroID = overviewHeroVisual(overview)?.game_id
  return overviewGameVisuals(overview?.featured_visuals).filter(item => item.game_id !== heroID).slice(0, 2)
}

export function overviewSiteVisuals(overview: InsightOverview | null): InsightSiteVisual[] {
  if (!Array.isArray(overview?.site_visuals) || !Array.isArray(overview.recent_changes)) return []
  const ids = new Set(overview.recent_changes.map(change => change?.entity?.id).filter(positiveID))
  const eligible = new Map<number, InsightSiteVisual>()
  for (const item of overview.site_visuals) {
    if (!record(item) || !positiveID(item.site_id) || !ids.has(item.site_id)
      || typeof item.name !== 'string' || !item.name.trim() || !record(item.visual)
      || item.visual.kind !== 'site_icon' || typeof item.visual.asset !== 'string'
      || /\s/.test(item.visual.asset)
      || icon.exec(item.visual.asset)?.[1] !== String(item.site_id)) continue
    if (!eligible.has(item.site_id)) eligible.set(item.site_id, { site_id: item.site_id, name: item.name.trim(), visual: { kind: 'site_icon', asset: item.visual.asset } })
  }
  return [...ids].flatMap(id => eligible.has(id) ? [eligible.get(id)!] : []).slice(0, 5)
}
