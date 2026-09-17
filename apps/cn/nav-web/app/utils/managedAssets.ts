import { averageProbe, isProbeMeasurement, type ProbeMeasurement } from './resourceRouting.ts'

export type AssetCDN = 'primary' | 'mirror'
export type AssetMode = 'auto' | AssetCDN
export const isAssetCDN = (value: unknown): value is AssetCDN => value === 'primary' || value === 'mirror'
export const normalizeAssetMode = (value: unknown): AssetMode => isAssetCDN(value) ? value : 'auto'
export function resolveAssetPreferred(mode: AssetMode, recommendation: AssetCDN | null): AssetCDN {
  return mode === 'auto' ? recommendation || 'primary' : mode
}
export type AssetOrigins = Record<AssetCDN, string>
export const ASSET_CDN_TTL_SECONDS = 12 * 60 * 60
export const ASSET_PROBE_KEY = 'system/probes/cdn.bin'
export const ASSET_PROBE_SHA256 = 'a2dacdf8cbd7f21e14efa73f83610bc1f40225cae90e691788000ca82cd8004e'
const managedKey = /^(?:nav\/sites\/[1-9]\d*\/icon\/[a-f0-9]{32}(?:\.[a-z0-9]{1,16})?|nav\/hero\/(?:desktop|mobile)\/[a-f0-9]{32}\.avif|nav\/patterns\/[a-f0-9]{32}\.svg)$/

export function assetURL(origins: AssetOrigins, provider: AssetCDN, key: string | null | undefined) {
  if (!key || !managedKey.test(key)) return ''
  const base = origins[provider].replace(/\/+$/, '')
  if (!/^https?:\/\/[^/]+$/.test(base)) return ''
  return `${base}/${key}`
}

export function chooseAssetCDN(primaryMs: number | null, mirrorMs: number | null): AssetCDN {
  if (mirrorMs === null) return 'primary'
  if (primaryMs === null) return 'mirror'
  const gain = primaryMs - mirrorMs
  return gain >= primaryMs * 0.2 || gain >= 50 ? 'mirror' : 'primary'
}

export type AssetProbeResult = { primaryMs: number | null; mirrorMs: number | null; primaryState: ProbeMeasurement['state']; mirrorState: ProbeMeasurement['state']; selected: AssetCDN; checkedAt: number; lastManualAt?: number; stale?: boolean }
export function readAssetDiagnostics(raw: unknown): AssetProbeResult | null {
  if (!raw || typeof raw !== 'object') return null
  const value = raw as AssetProbeResult
  // Existing diagnostics predate explicit success/timeout/failed states.
  const primaryState = value.primaryState || (value.primaryMs === null ? 'failed' : 'success')
  const mirrorState = value.mirrorState || (value.mirrorMs === null ? 'failed' : 'success')
  return isAssetCDN(value.selected) && Number.isFinite(value.checkedAt)
    && isProbeMeasurement({ state: primaryState, ms: value.primaryMs }) && isProbeMeasurement({ state: mirrorState, ms: value.mirrorMs })
    ? { ...value, primaryState, mirrorState } : null
}
export async function probeAssetCDNs(origins: AssetOrigins, fetcher: typeof fetch = fetch, now = () => performance.now(), previous: AssetCDN = 'primary'): Promise<AssetProbeResult> {
  const measure = async (provider: AssetCDN): Promise<ProbeMeasurement> => {
    const controller = new AbortController()
    const timer = setTimeout(() => controller.abort(), 3500)
    try {
      const start = now()
      const response = await fetcher(`${origins[provider].replace(/\/+$/, '')}/${ASSET_PROBE_KEY}?probe=${crypto.randomUUID()}`, { signal: controller.signal, mode: 'cors', credentials: 'omit', cache: 'no-store' })
      if (!response.ok) return { ms: null, state: 'failed' }
      const bytes = await response.arrayBuffer()
      const elapsed = now() - start
      if (bytes.byteLength !== 8192) return { ms: null, state: 'failed' }
      const digest = await crypto.subtle.digest('SHA-256', bytes)
      const hash = [...new Uint8Array(digest)].map((byte) => byte.toString(16).padStart(2, '0')).join('')
      return hash === ASSET_PROBE_SHA256 ? { ms: elapsed, state: 'success' } : { ms: null, state: 'failed' }
    } catch { return { ms: null, state: controller.signal.aborted ? 'timeout' : 'failed' } } finally { clearTimeout(timer) }
  }
  const first = await Promise.all([measure('primary'), measure('mirror')])
  const second = await Promise.all([measure('primary'), measure('mirror')])
  const primary = averageProbe([first[0]!, second[0]!]), mirror = averageProbe([first[1]!, second[1]!])
  const primaryMs = primary.ms, mirrorMs = mirror.ms
  return { primaryMs, mirrorMs, primaryState: primary.state, mirrorState: mirror.state, selected: primaryMs === null && mirrorMs === null ? previous : chooseAssetCDN(primaryMs, mirrorMs), checkedAt: Date.now() }
}

export function assetCandidate(origins: AssetOrigins, preferred: AssetCDN, key: string | null | undefined, failed: ReadonlySet<AssetCDN>, fallback: string) {
  const candidates: AssetCDN[] = [preferred, preferred === 'primary' ? 'mirror' : 'primary']
  const failedURLs = new Set([...failed].map(provider => assetURL(origins, provider, key)))
  for (const provider of candidates) { if (!failed.has(provider)) { const url = assetURL(origins, provider, key); if (url && !failedURLs.has(url)) return { url, provider } } }
  return { url: fallback, provider: null }
}
