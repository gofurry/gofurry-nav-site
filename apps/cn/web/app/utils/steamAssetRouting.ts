import { isSteamGroup, STEAM_SHARED_CDN_GROUP_PREFIXES, type SteamSharedCdnGroup } from './steamAssets.ts'
import { averageProbe, isProbeMeasurement, type ProbeMeasurement } from './resourceRouting.ts'

export const STEAM_LEGACY_KEY = 'gofurry:steam-shared-cdn-preference:v1'
export const STEAM_DIAGNOSTICS_KEY = 'gf_steam_asset_diagnostics_v1'
export const STEAM_PROBE_PATHS = [570, 440, 550, 730].map(id => `/store_item_assets/steam/apps/${id}/capsule_sm_120.jpg`)
export type SteamProbeResult = {
  version: 1
  china: ProbeMeasurement
  global: ProbeMeasurement
  selected: SteamSharedCdnGroup
  checkedAt: number
  sample: string
  lastManualAt?: number
  stale?: boolean
}

export function readSteamDiagnostics(raw: unknown): SteamProbeResult | null {
  if (!raw || typeof raw !== 'object') return null
  const value = raw as SteamProbeResult
  return value.version === 1 && Number.isFinite(value.checkedAt) && isSteamGroup(value.selected)
    && isProbeMeasurement(value.china) && isProbeMeasurement(value.global) ? value : null
}

// The old six-hour record was an automatic result, never an explicit pin.
export function legacySteamRecommendation(raw: unknown, now = Date.now()): SteamSharedCdnGroup | null {
  if (!raw || typeof raw !== 'object') return null
  const record = raw as { group?: unknown; testedAt?: unknown }
  return isSteamGroup(record.group) && typeof record.testedAt === 'number' && record.testedAt <= now
    && now - record.testedAt < 6 * 60 * 60 * 1000 ? record.group : null
}
export function chooseSteamGroup(china: number | null, global: number | null, previous: SteamSharedCdnGroup) {
  if (china === null && global === null) return previous
  if (china === null) return 'global'
  if (global === null) return 'china'
  return china <= global ? 'china' : 'global'
}
export function probeSteamImage(url: string): Promise<ProbeMeasurement> {
  return new Promise(resolve => {
    const started = performance.now()
    const image = new Image()
    const finish = (state: ProbeMeasurement['state']) => {
      clearTimeout(timer)
      image.onload = image.onerror = null
      if (state !== 'success') image.removeAttribute('src')
      resolve({ ms: state === 'success' ? performance.now() - started : null, state })
    }
    const timer = setTimeout(() => finish('timeout'), 2800)
    image.onload = () => finish('success')
    image.onerror = () => finish('failed')
    image.decoding = 'async'
    image.src = url
  })
}
export async function probeSteamGroups(previous: SteamSharedCdnGroup, measure = probeSteamImage, now = Date.now, manual = false): Promise<SteamProbeResult> {
  const bucket = Math.floor(now() / 60_000)
  let result: SteamProbeResult | undefined
  for (const sample of STEAM_PROBE_PATHS) {
    const round = (index: number) => Promise.all((['china', 'global'] as const).map(group =>
      measure(`${STEAM_SHARED_CDN_GROUP_PREFIXES[group][0]}${sample}?gf_probe=${bucket}&round=${index + (manual ? 2 : 0)}`)))
    const first = await round(0)
    const second = await round(1)
    const china = averageProbe([first[0]!, second[0]!]), global = averageProbe([first[1]!, second[1]!])
    result = { version: 1, china, global, selected: chooseSteamGroup(china.ms, global.ms, previous), checkedAt: now(), sample }
    // The sample pool is a fallback, not a request to download all four images.
    if (china.ms !== null || global.ms !== null) break
  }
  return result!
}
