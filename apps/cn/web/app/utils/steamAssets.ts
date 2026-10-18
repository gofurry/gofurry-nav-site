export type SteamSharedCdnGroup = 'china' | 'global'
export type SteamAssetMode = 'auto' | SteamSharedCdnGroup
export const isSteamGroup = (value: unknown): value is SteamSharedCdnGroup => value === 'china' || value === 'global'
export const normalizeSteamMode = (value: unknown): SteamAssetMode => isSteamGroup(value) ? value : 'auto'

export const STEAM_SHARED_CDN_GROUP_PREFIXES: Record<SteamSharedCdnGroup, readonly string[]> = {
  china: ['https://shared.st.dl.eccdnx.com', 'https://shared.cdn.steamchina.queniuam.com'],
  global: ['https://shared.akamai.steamstatic.com', 'https://shared.cloudflare.steamstatic.com', 'https://shared.fastly.steamstatic.com', 'https://shared.steamstatic.com'],
}
export const STEAM_SHARED_CDN_PREFIXES = [...STEAM_SHARED_CDN_GROUP_PREFIXES.china, ...STEAM_SHARED_CDN_GROUP_PREFIXES.global]
const steamSharedHosts = new Set(STEAM_SHARED_CDN_PREFIXES.map(prefix => new URL(prefix).hostname))

export function defaultSteamSharedCdnGroup(locale?: string | null): SteamSharedCdnGroup {
  return locale?.toLowerCase().startsWith('en') ? 'global' : 'china'
}
export function resolveSteamPreferred(mode: SteamAssetMode, recommendation: SteamSharedCdnGroup | null, locale?: string | null) {
  return mode === 'auto' ? recommendation || defaultSteamSharedCdnGroup(locale) : mode
}

export function steamSharedAssetCandidates(rawUrl?: string | null, preferred: SteamSharedCdnGroup = 'china') {
  const source = rawUrl?.trim()
  if (!source) return []
  let parsed: URL
  try { parsed = new URL(source) } catch { return [source] }
  if (!['http:', 'https:'].includes(parsed.protocol) || !steamSharedHosts.has(parsed.hostname)) return [source]
  const assetPath = `${parsed.pathname}${parsed.search}${parsed.hash}`
  const other = preferred === 'china' ? 'global' : 'china'
  const prefixes = [...STEAM_SHARED_CDN_GROUP_PREFIXES[preferred], ...STEAM_SHARED_CDN_GROUP_PREFIXES[other]]
  return [...new Set([...prefixes.map(prefix => `${prefix}${assetPath}`), source])]
}
