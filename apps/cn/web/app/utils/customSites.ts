import { recordRecentSite, toExternalUrl, type RecentSiteItem } from '@/utils/recentSites'

export interface CustomSiteItem extends RecentSiteItem {}

export const CUSTOM_SITES_STORAGE_KEY = 'navCustomSites'
export const CUSTOM_SITES_EVENT = 'nav-custom-sites-change'
export const MAX_CUSTOM_SITES = 7

function dispatchCustomSitesChange() {
  window.dispatchEvent(new Event(CUSTOM_SITES_EVENT))
}

export function loadCustomSites(): CustomSiteItem[] {
  const saved = localStorage.getItem(CUSTOM_SITES_STORAGE_KEY)
  if (!saved) {
    return []
  }

  try {
    const parsed = JSON.parse(saved)
    if (!Array.isArray(parsed)) {
      return []
    }

    return parsed.filter((item) => item?.id && item?.name && item?.url).slice(0, MAX_CUSTOM_SITES)
  } catch {
    return []
  }
}

export function saveCustomSites(items: CustomSiteItem[]) {
  const nextItems = items
    .filter((item) => item?.id && item?.name && item?.url)
    .slice(0, MAX_CUSTOM_SITES)

  localStorage.setItem(CUSTOM_SITES_STORAGE_KEY, JSON.stringify(nextItems))
  dispatchCustomSitesChange()
}

export function addCustomSite(site: Omit<CustomSiteItem, 'id'>) {
  const normalizedUrl = toExternalUrl(site.url)
  if (!normalizedUrl) {
    return false
  }

  const current = loadCustomSites().filter((item) => toExternalUrl(item.url) !== normalizedUrl)
  if (current.length >= MAX_CUSTOM_SITES) {
    return false
  }

  current.push({
    id: `custom-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    name: site.name.trim(),
    url: normalizedUrl,
  })
  saveCustomSites(current)
  return true
}

export function removeCustomSite(id: string) {
  saveCustomSites(loadCustomSites().filter((item) => item.id !== id))
}

export function reorderCustomSites(ids: string[]) {
  const current = loadCustomSites()
  const orderMap = new Map(current.map((item) => [item.id, item]))
  const ordered = ids
    .map((id) => orderMap.get(id))
    .filter((item): item is CustomSiteItem => Boolean(item))

  const rest = current.filter((item) => !ids.includes(item.id))
  saveCustomSites([...ordered, ...rest])
}

export function visitCustomSite(site: CustomSiteItem) {
  const targetUrl = toExternalUrl(site.url)
  if (!targetUrl) {
    return
  }

  recordRecentSite({
    id: site.id,
    name: site.name,
    url: targetUrl,
  })
  window.open(targetUrl, '_blank')
}
