const STEAM_LIBRARY_COVER_BASE_URL = 'https://shared.steamstatic.com/store_item_assets/steam/apps/'
const CDN_LIBRARY_COVER_BASE_URL = normalizeBaseUrl(
  import.meta.env.VITE_GAME_PREFIX_URL || 'https://qcdn.go-furry.com/game/'
)

function normalizeAppid(appid?: number | string | null) {
  return String(appid ?? '').trim()
}

function normalizeBaseUrl(url: string) {
  return url.endsWith('/') ? url : `${url}/`
}

export function steamLibraryCoverUrl(appid?: number | string | null) {
  const normalizedAppid = normalizeAppid(appid)
  return normalizedAppid ? `${STEAM_LIBRARY_COVER_BASE_URL}${normalizedAppid}/library_600x900.jpg` : ''
}

export function cdnLibraryCoverUrl(appid?: number | string | null) {
  const normalizedAppid = normalizeAppid(appid)
  return normalizedAppid && CDN_LIBRARY_COVER_BASE_URL
    ? `${CDN_LIBRARY_COVER_BASE_URL}${normalizedAppid}/library_600x900.jpg`
    : ''
}

export function gameLibraryCoverSources(appid?: number | string | null) {
  return [
    steamLibraryCoverUrl(appid),
    cdnLibraryCoverUrl(appid),
  ].filter(Boolean)
}
