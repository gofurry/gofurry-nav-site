export const STEAM_SHARED_CDN_PREFIXES = [
  'https://shared.steamstatic.com',
  'https://shared.akamai.steamstatic.com',
  'https://shared.cloudflare.steamstatic.com',
  'https://shared.fastly.steamstatic.com',
  'https://shared.cdn.steamchina.queniuam.com',
] as const

const STEAM_SHARED_NATIVE_PREFIX = 'https://shared.steamstatic.com'
const STEAM_SHARED_CHINA_PREFIX = 'https://shared.cdn.steamchina.queniuam.com'

const STEAM_SHARED_FALLBACK_PREFIXES = {
  zh: [
    'https://shared.fastly.steamstatic.com',
    'https://shared.cloudflare.steamstatic.com',
    'https://shared.akamai.steamstatic.com',
    STEAM_SHARED_NATIVE_PREFIX,
    STEAM_SHARED_CHINA_PREFIX,
  ],
  en: [
    'https://shared.akamai.steamstatic.com',
    'https://shared.cloudflare.steamstatic.com',
    'https://shared.fastly.steamstatic.com',
    STEAM_SHARED_NATIVE_PREFIX,
    STEAM_SHARED_CHINA_PREFIX,
  ],
} as const

const steamSharedHosts = new Set(
  STEAM_SHARED_CDN_PREFIXES.map((prefix) => new URL(prefix).hostname)
)

export function steamSharedAssetCandidates(rawUrl?: string | null, locale?: string | null) {
  const source = rawUrl?.trim()
  if (!source) {
    return []
  }

  let parsed: URL
  try {
    parsed = new URL(source)
  } catch {
    return [source]
  }

  if (!['http:', 'https:'].includes(parsed.protocol) || !steamSharedHosts.has(parsed.hostname)) {
    return [source]
  }

  const assetPath = `${parsed.pathname}${parsed.search}${parsed.hash}`
  return [...new Set(steamSharedFallbackPrefixes(locale).map((prefix) => `${prefix}${assetPath}`))]
}

export function preferredSteamSharedAssetUrl(rawUrl?: string | null, locale?: string | null) {
  return steamSharedAssetCandidates(rawUrl, locale)[0] ?? ''
}

function steamSharedFallbackPrefixes(locale?: string | null) {
  return locale === 'en'
    ? STEAM_SHARED_FALLBACK_PREFIXES.en
    : STEAM_SHARED_FALLBACK_PREFIXES.zh
}
