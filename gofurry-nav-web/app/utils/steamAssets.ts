export const STEAM_SHARED_CDN_PREFIXES = [
  'https://shared.steamstatic.com',
  'https://shared.akamai.steamstatic.com',
  'https://shared.cloudflare.steamstatic.com',
  'https://shared.fastly.steamstatic.com',
  'https://shared.st.dl.eccdnx.com',
  'https://shared.cdn.steamchina.queniuam.com',
] as const

const STEAM_SHARED_NATIVE_PREFIX = 'https://shared.steamstatic.com'
const STEAM_SHARED_CHINA_PREFIX = 'https://shared.cdn.steamchina.queniuam.com'
const STEAM_SHARED_CHINA_BAISHAN_PREFIX = 'https://shared.st.dl.eccdnx.com'

type SteamSharedCdnGroup = 'china' | 'global'

const STEAM_SHARED_CDN_GROUP_PREFIXES: Record<SteamSharedCdnGroup, readonly string[]> = {
  china: [
    STEAM_SHARED_CHINA_BAISHAN_PREFIX,
    STEAM_SHARED_CHINA_PREFIX,
  ],
  global: [
    'https://shared.akamai.steamstatic.com',
    'https://shared.cloudflare.steamstatic.com',
    'https://shared.fastly.steamstatic.com',
    STEAM_SHARED_NATIVE_PREFIX,
  ],
} as const

const STEAM_SHARED_CDN_PREFERENCE_KEY = 'gofurry:steam-shared-cdn-preference:v1'
const STEAM_SHARED_CDN_PREFERENCE_TTL_MS = 6 * 60 * 60 * 1000
const STEAM_SHARED_CDN_PROBE_TIMEOUT_MS = 2800
const STEAM_SHARED_CDN_PREFERENCE_EVENT = 'gofurry:steam-shared-cdn-preference-updated'

const steamSharedHosts = new Set(
  STEAM_SHARED_CDN_PREFIXES.map((prefix) => new URL(prefix).hostname)
)
let activeProbe: Promise<void> | null = null

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

  const assetPath = `${toSteamShared1xPath(parsed.pathname)}${parsed.search}${parsed.hash}`
  const candidates = steamSharedFallbackPrefixes(locale).map((prefix) => `${prefix}${assetPath}`)
  candidates.push(source)
  return [...new Set(candidates)]
}

export function preferredSteamSharedAssetUrl(rawUrl?: string | null, locale?: string | null) {
  return steamSharedAssetCandidates(rawUrl, locale)[0] ?? ''
}

export function ensureSteamSharedCdnPreference(rawUrl?: string | null, locale?: string | null) {
  if (!isBrowser() || readSteamSharedCdnPreference()) {
    return
  }

  const parsed = parseSteamSharedUrl(rawUrl)
  if (!parsed || activeProbe) {
    return
  }

  const assetPath = `${toSteamShared1xPath(parsed.pathname)}${parsed.search}${parsed.hash}`
  const defaultGroup = defaultSteamSharedCdnGroup(locale)
  const fallbackGroup = defaultGroup === 'china' ? 'global' : 'china'

  activeProbe = Promise
    .allSettled([
      probeSteamSharedCdnGroup(defaultGroup, assetPath),
      probeSteamSharedCdnGroup(fallbackGroup, assetPath),
    ])
    .then((results) => {
      const winner = results
        .map((result) => result.status === 'fulfilled' ? result.value : null)
        .filter((result): result is { group: SteamSharedCdnGroup, duration: number } => Boolean(result))
        .sort((a, b) => a.duration - b.duration)[0]

      if (winner) {
        writeSteamSharedCdnPreference(winner.group)
      }
    })
    .finally(() => {
      activeProbe = null
    })
}

function steamSharedFallbackPrefixes(locale?: string | null) {
  const preferredGroup = readSteamSharedCdnPreference() || defaultSteamSharedCdnGroup(locale)
  const fallbackGroup = preferredGroup === 'china' ? 'global' : 'china'

  return [
    ...STEAM_SHARED_CDN_GROUP_PREFIXES[preferredGroup],
    ...STEAM_SHARED_CDN_GROUP_PREFIXES[fallbackGroup],
  ]
}

function defaultSteamSharedCdnGroup(locale?: string | null): SteamSharedCdnGroup {
  return locale?.toLowerCase().startsWith('en') ? 'global' : 'china'
}

function toSteamShared1xPath(pathname: string) {
  return pathname
    .replace(/header_2x(\.[a-z0-9]+)$/i, 'header$1')
    .replace(/library_capsule_2x(\.[a-z0-9]+)$/i, 'library_capsule$1')
    .replace(/capsule_main_2x(\.[a-z0-9]+)$/i, 'capsule_main$1')
    .replace(/hero_capsule_2x(\.[a-z0-9]+)$/i, 'hero_capsule$1')
    .replace(/capsule_small_2x(\.[a-z0-9]+)$/i, 'capsule_small$1')
    .replace(/library_hero_2x(\.[a-z0-9]+)$/i, 'library_hero$1')
    .replace(/library_logo_2x(\.[a-z0-9]+)$/i, 'library_logo$1')
}

function parseSteamSharedUrl(rawUrl?: string | null) {
  const source = rawUrl?.trim()
  if (!source) {
    return null
  }

  try {
    const parsed = new URL(source)
    if (!['http:', 'https:'].includes(parsed.protocol) || !steamSharedHosts.has(parsed.hostname)) {
      return null
    }
    return parsed
  } catch {
    return null
  }
}

function readSteamSharedCdnPreference(): SteamSharedCdnGroup | null {
  if (!isBrowser()) {
    return null
  }

  try {
    const raw = window.localStorage.getItem(STEAM_SHARED_CDN_PREFERENCE_KEY)
    if (!raw) {
      return null
    }
    const parsed = JSON.parse(raw) as { group?: SteamSharedCdnGroup, testedAt?: number }
    if ((parsed.group !== 'china' && parsed.group !== 'global') || typeof parsed.testedAt !== 'number') {
      return null
    }
    if (Date.now() - parsed.testedAt > STEAM_SHARED_CDN_PREFERENCE_TTL_MS) {
      window.localStorage.removeItem(STEAM_SHARED_CDN_PREFERENCE_KEY)
      return null
    }
    return parsed.group
  } catch {
    return null
  }
}

function writeSteamSharedCdnPreference(group: SteamSharedCdnGroup) {
  if (!isBrowser()) {
    return
  }

  try {
    window.localStorage.setItem(STEAM_SHARED_CDN_PREFERENCE_KEY, JSON.stringify({
      group,
      testedAt: Date.now(),
    }))
    window.dispatchEvent(new CustomEvent(STEAM_SHARED_CDN_PREFERENCE_EVENT, { detail: { group } }))
  } catch {
    // Ignore storage failures; image fallback still works without persisted preference.
  }
}

function probeSteamSharedCdnGroup(group: SteamSharedCdnGroup, assetPath: string) {
  const primaryPrefix = STEAM_SHARED_CDN_GROUP_PREFIXES[group][0]
  return probeSteamSharedImage(`${primaryPrefix}${assetPath}`).then((duration) => ({ group, duration }))
}

function probeSteamSharedImage(url: string) {
  return new Promise<number>((resolve, reject) => {
    const startedAt = performance.now()
    const image = new Image()
    const timer = window.setTimeout(() => {
      cleanup()
      reject(new Error('steam shared cdn probe timeout'))
    }, STEAM_SHARED_CDN_PROBE_TIMEOUT_MS)

    function cleanup() {
      window.clearTimeout(timer)
      image.onload = null
      image.onerror = null
    }

    image.onload = () => {
      const duration = performance.now() - startedAt
      cleanup()
      resolve(duration)
    }
    image.onerror = () => {
      cleanup()
      reject(new Error('steam shared cdn probe failed'))
    }
    image.decoding = 'async'
    image.src = url
  })
}

function isBrowser() {
  return typeof window !== 'undefined'
}
