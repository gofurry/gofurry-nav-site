type ApiResult<T> = {
  code: number
  data: T
}

type SiteRecord = {
  id: string
  domain: string
}

type GameRecord = {
  id?: string
  game_id?: string
}

type GameListPayload = GameRecord[] | {
  list?: GameRecord[]
  data?: GameRecord[]
  rows?: GameRecord[]
}

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

function firstDomain(rawDomain: string) {
  if (!rawDomain) {
    return ''
  }

  try {
    const parsed = JSON.parse(rawDomain)
    if (Array.isArray(parsed?.domain) && typeof parsed.domain[0] === 'string') {
      return parsed.domain[0]
    }
    if (Array.isArray(parsed) && typeof parsed[0] === 'string') {
      return parsed[0]
    }
  } catch {
    return rawDomain
  }

  return rawDomain
}

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const siteUrl = String(config.public.siteUrl).replace(/\/$/, '')
  const urls = new Set([
    '/',
    '/nav',
    '/games',
    '/updates',
    '/about',
    '/games/news/more',
    '/games/creator'
  ])

  const [sites, games] = await Promise.all([
    $fetch<ApiResult<SiteRecord[]>>('/api/v1/nav/page/site/list', {
      query: { lang: 'zh' }
    }).then((res) => res.code === 1 ? res.data : []).catch(() => []),
    $fetch<ApiResult<GameListPayload>>('/api/v1/game/info/list', {
      query: { num: '9999', lang: 'zh' }
    }).then((res) => res.code === 1 ? res.data : []).catch(() => [])
  ])

  for (const site of sites) {
    if (site?.id != null) {
      const domain = firstDomain(site.domain)
      urls.add(domain
        ? `/site/${encodeURIComponent(String(site.id))}/${encodeURIComponent(domain)}`
        : `/site/${encodeURIComponent(String(site.id))}`)
    }
  }

  const gameList = Array.isArray(games)
    ? games
    : games.list ?? games.data ?? games.rows ?? []

  for (const game of gameList) {
    const gameId = game.id ?? game.game_id
    if (gameId != null && gameId !== '') {
      urls.add(`/games/${String(gameId)}`)
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
