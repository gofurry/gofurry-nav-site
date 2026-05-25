export function siteDetailPath(siteId: string | number, domain?: string) {
  const id = encodeURIComponent(String(siteId))
  const cleanDomain = domain?.trim()

  if (!cleanDomain) {
    return `/site/${id}`
  }

  return `/site/${id}/${encodeURIComponent(cleanDomain)}`
}
