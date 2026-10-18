export interface RecentSiteItem {
  id: string
  name: string
  url: string
}

export const RECENT_SITES_STORAGE_KEY = 'recentSites'
export const RECENT_SITES_EVENT = 'recent-sites-change'
const MAX_RECENT_SITES = 12

export function toExternalUrl(url: string): string {
  const trimmed = url.trim()
  if (!trimmed) {
    return ''
  }

  return /^https?:\/\//i.test(trimmed) ? trimmed : `https://${trimmed}`
}

export function loadRecentSites(): RecentSiteItem[] {
  const saved = localStorage.getItem(RECENT_SITES_STORAGE_KEY)
  if (!saved) {
    return []
  }

  try {
    const parsed = JSON.parse(saved)
    if (!Array.isArray(parsed)) {
      return []
    }

    return parsed.filter(item => item?.id && item?.name && item?.url)
  } catch {
    return []
  }
}

export function recordRecentSite(site: RecentSiteItem) {
  const normalizedUrl = toExternalUrl(site.url)
  if (!normalizedUrl) {
    return
  }

  const nextItem: RecentSiteItem = {
    id: site.id,
    name: site.name,
    url: normalizedUrl,
  }

  const deduped = loadRecentSites().filter(item => item.url !== nextItem.url)
  const nextSites = [nextItem, ...deduped].slice(0, MAX_RECENT_SITES)
  localStorage.setItem(RECENT_SITES_STORAGE_KEY, JSON.stringify(nextSites))
  window.dispatchEvent(new Event(RECENT_SITES_EVENT))
}
