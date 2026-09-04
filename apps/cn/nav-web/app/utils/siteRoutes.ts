export function siteEntityPath(siteId: string | number) {
  const id = encodeURIComponent(String(siteId))
  return `/site/${id}`
}

export function siteTargetPath(siteId: string | number, domain: string) {
  const entityPath = siteEntityPath(siteId)
  const cleanDomain = domain.trim()
  if (!cleanDomain) {
    return entityPath
  }

  return `${entityPath}?domain=${encodeURIComponent(cleanDomain)}`
}
