import { parseGameInventory, parseSiteGroupInventory, parseSiteInventory } from '../utils/sitemapInventory'

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

export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event)
  const siteUrl = String(config.public.siteUrl).replace(/\/$/, '')
  const urls = new Set<string>()

  for (const path of [
    '/',
    '/games',
    '/insights',
    '/insights/sites',
    '/insights/sites/certificates',
    '/insights/games',
    '/insights/games/players',
    '/insights/games/prices',
    '/insights/games/languages',
    '/insights/changes',
    '/updates',
    '/about',
    '/privacy',
    '/terms'
  ]) {
    addLocalizedUrls(urls, path)
  }

  let sites
  let siteGroups
  let games
  try {
    const [siteResponse, siteGroupResponse, gameResponse] = await Promise.all([
      $fetch('/api/v2/nav/sites/index'),
      $fetch('/api/v2/nav/site-groups', {
        query: { lang: 'zh' }
      }),
      $fetch('/api/v2/game/list', {
        query: { limit: '5000', lang: 'zh', region: 'CN' }
      })
    ])
    sites = parseSiteInventory(siteResponse)
    siteGroups = parseSiteGroupInventory(siteGroupResponse)
    games = parseGameInventory(gameResponse)
  } catch (error) {
    throw createError({
      statusCode: 503,
      statusMessage: 'Sitemap inventory temporarily unavailable',
      cause: error,
    })
  }

  for (const site of sites) {
    addLocalizedUrls(urls, `/site/${encodeURIComponent(site.id)}`)
  }

  for (const group of siteGroups) {
    addLocalizedUrls(urls, `/site-groups/${encodeURIComponent(group.id)}`)
  }

  for (const game of games) {
    addLocalizedUrls(urls, `/games/${encodeURIComponent(game.id)}`)
  }

  setHeader(event, 'content-type', 'application/xml; charset=utf-8')
  return [
    '<?xml version="1.0" encoding="UTF-8"?>',
    '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">',
    ...Array.from(urls).map((path) => urlEntry(`${siteUrl}${path}`)),
    '</urlset>'
  ].join('')
})
