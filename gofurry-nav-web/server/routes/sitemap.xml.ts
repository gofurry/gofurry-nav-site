type ApiResult<T> = {
  code: number
  data: T
}

type SiteRecord = {
  id: number | string
  domains: string[]
}

type SiteGroupRecord = {
  id: number | string
}

type GameRecord = {
  id?: string
  game_id?: string
}

type GameListPayload = GameRecord[]

function escapeXml(value: string) {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&apos;')
}

function urlEntry(loc: string) {
  return `<url><loc>${escapeXml(loc)}</loc></url>`
}

function localizedPaths(path: string) {
  if (path === '/') {
    return ['/', '/en']
  }
  return [path, `/en${path}`]
}

function addLocalizedUrls(urls: Set<string>, path: string) {
  for (const localizedPath of localizedPaths(path)) {
    urls.add(localizedPath)
  }
}

function firstDomain(domains: string[] | undefined) {
  if (!domains?.length) {
    return ''
  }
  return domains[0] ?? ''
}

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const siteUrl = String(config.public.siteUrl).replace(/\/$/, '')
  const urls = new Set<string>()

  for (const path of [
    '/',
    '/games',
    '/updates',
    '/about',
    '/privacy',
    '/terms',
    '/games/prize'
  ]) {
    addLocalizedUrls(urls, path)
  }

  const [sites, siteGroups, games] = await Promise.all([
    $fetch<ApiResult<{ items: SiteRecord[] }>>('/api/v2/nav/sites/index').then((res) => res.code === 1 ? res.data.items : []).catch(() => []),
    $fetch<ApiResult<SiteGroupRecord[]>>('/api/v2/nav/sync/site-groups', {
      query: { lang: 'zh' }
    }).then((res) => res.code === 1 ? res.data : []).catch(() => []),
    $fetch<ApiResult<GameListPayload>>('/api/v2/game/list', {
      query: { limit: '5000', lang: 'zh', region: 'CN' }
    }).then((res) => res.code === 1 ? res.data : []).catch(() => [])
  ])

  for (const site of sites) {
    if (site?.id != null) {
      const domain = firstDomain(site.domains)
      addLocalizedUrls(urls, domain
        ? `/site/${encodeURIComponent(String(site.id))}/${encodeURIComponent(domain)}`
        : `/site/${encodeURIComponent(String(site.id))}`)
    }
  }

  for (const group of siteGroups) {
    if (group?.id != null && group.id !== '') {
      addLocalizedUrls(urls, `/site-groups/${encodeURIComponent(String(group.id))}`)
    }
  }

  for (const game of games) {
    const gameId = game.id ?? game.game_id
    if (gameId != null && gameId !== '') {
      addLocalizedUrls(urls, `/games/${encodeURIComponent(String(gameId))}`)
    }
  }

  setHeader(event, 'content-type', 'application/xml; charset=utf-8')
  return [
    '<?xml version="1.0" encoding="UTF-8"?>',
    '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">',
    ...Array.from(urls).map((path) => urlEntry(`${siteUrl}${path}`)),
    '</urlset>'
  ].join('')
})
