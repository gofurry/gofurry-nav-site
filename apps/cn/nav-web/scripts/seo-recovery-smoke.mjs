#!/usr/bin/env node

const args = parseArgs(process.argv.slice(2))
const baseUrl = normalizeBaseUrl(args['base-url'] || process.env.SEO_BASE_URL || 'http://127.0.0.1:3000')
const siteId = args['site-id'] || process.env.SEO_SITE_ID || '1'
const siteDomain = args['site-domain'] || process.env.SEO_SITE_DOMAIN || 'example.com'
const invalidSiteDomain = args['invalid-site-domain'] || process.env.SEO_INVALID_SITE_DOMAIN || 'seo-invalid-target.invalid'
const gameId = args['game-id'] || process.env.SEO_GAME_ID || '110'
const missingSiteId = args['missing-site-id'] || process.env.SEO_MISSING_SITE_ID || '999999999'
const missingGameId = args['missing-game-id'] || process.env.SEO_MISSING_GAME_ID || '999999999'
const siteGroupId = args['site-group-id'] || process.env.SEO_SITE_GROUP_ID || ''
const failureMode = args.mode === 'failure' || process.env.SEO_EXPECT_UPSTREAM_FAILURES === '1'

if (failureMode) {
  await expectNotFound('/site/abc')
  await expectNotFound('/games/abc')
  await expectStatus(`/site/${siteId}`, 503)
  await expectStatus(`/games/${gameId}`, 503)
  await expectStatus('/sitemap.xml', 503)
  console.log('[seo-recovery] authoritative pages and sitemap fail closed with HTTP 503')
  process.exit(0)
}

await expectRedirect(`/sites/${siteId}`, `/site/${siteId}`)
await expectRedirect(`/en/sites/${siteId}`, `/en/site/${siteId}`)
await expectRedirect(`/site/${siteId}/${encodeURIComponent(siteDomain)}`, `/site/${siteId}`, siteDomain)
await expectRedirect(`/en/site/${siteId}/${encodeURIComponent(siteDomain)}`, `/en/site/${siteId}`, siteDomain)

for (const route of [`/site/${siteId}`, `/site/${siteId}?domain=${encodeURIComponent(siteDomain)}`]) {
  await expectCanonical(route, `/site/${siteId}`)
}
for (const route of [`/en/site/${siteId}`, `/en/site/${siteId}?domain=${encodeURIComponent(siteDomain)}`]) {
  await expectCanonical(route, `/en/site/${siteId}`)
}

await expectStatus(`/games/${gameId}`, 200, 'text/html', 'game-detail-page')
await expectStatus(`/site/${siteId}`, 200, 'text/html', 'site-detail-page')
await expectNotFound('/games/abc')
await expectNotFound('/site/abc')
await expectNotFound(`/site/${siteId}?domain=${encodeURIComponent(invalidSiteDomain)}`)
await expectNotFound(`/games/${missingGameId}`)
await expectNotFound(`/site/${missingSiteId}`)
for (const route of ['/games/prize', '/en/games/prize', '/games/prize/activation', '/en/games/prize/activation']) {
  await expectPrizeNoIndex(route)
}
if (siteGroupId) {
  await expectSiteGroupSeo(`/site-groups/${encodeURIComponent(siteGroupId)}`)
  await expectSiteGroupSeo(`/en/site-groups/${encodeURIComponent(siteGroupId)}`)
}
await expectStatus('/steam', 404)
await expectStatus('/en/steam', 404)

const sitemapResponse = await request('/sitemap.xml')
const sitemap = await sitemapResponse.text()
assert(sitemapResponse.status === 200, `/sitemap.xml expected HTTP 200, received ${sitemapResponse.status}`)
assert(sitemapResponse.headers.get('content-type')?.includes('application/xml'), '/sitemap.xml did not return application/xml')
const sitemapUrls = Array.from(sitemap.matchAll(/<loc>(.*?)<\/loc>/g), match => decodeXml(match[1]))
assert(sitemapUrls.length > 0, 'sitemap contained no URLs')
const sitemapPaths = sitemapUrls.map(value => new URL(value).pathname + new URL(value).search)
assert(sitemapPaths.includes(`/site/${siteId}`), `sitemap omitted /site/${siteId}`)
assert(sitemapPaths.includes(`/games/${gameId}`), `sitemap omitted /games/${gameId}`)
assert(!sitemapPaths.some(path => path.startsWith('/sites/')), 'sitemap contains a legacy /sites/:id URL')
assert(!sitemapPaths.some(path => /^\/(?:en\/)?site\/[^/]+\/.+/.test(path)), 'sitemap contains a Site domain child URL')
assert(!sitemapPaths.some(path => path.includes('?domain=')), 'sitemap contains a Site target query URL')
assert(!/%7b%22|%7B%22|\{&quot;domain&quot;/.test(sitemap), 'sitemap contains encoded JSON domain garbage')
assert(!sitemapPaths.some(path => path === '/games/prize' || path === '/en/games/prize'), 'sitemap contains /games/prize')
assert(!sitemapPaths.some(path => path === '/steam' || path === '/en/steam'), 'sitemap contains the missing /steam route')

console.log(`[seo-recovery] production SSR smoke passed; sitemap URLs: ${sitemapUrls.length}`)

async function expectRedirect(route, expectedPath, expectedDomain = '') {
  const response = await request(route, { redirect: 'manual' })
  assert(response.status === 301, `${route} expected HTTP 301, received ${response.status}`)
  const location = response.headers.get('location')
  assert(location, `${route} did not return a Location header`)
  const target = new URL(location, baseUrl)
  assert(target.pathname === expectedPath, `${route} redirected to ${target.pathname}, expected ${expectedPath}`)
  if (expectedDomain) {
    assert(target.searchParams.get('domain') === expectedDomain, `${route} did not preserve the target domain in the query`)
  } else {
    assert(target.search === '', `${route} unexpectedly added a query string`)
  }
}

async function expectCanonical(route, expectedPath) {
  const response = await request(route)
  const html = await response.text()
  assert(response.status === 200, `${route} expected HTTP 200, received ${response.status}`)
  const links = parseLinks(html)
  const canonicals = links.filter(link => link.rel === 'canonical')
  assert(canonicals.length === 1, `${route} expected exactly one canonical link, received ${canonicals.length}`)
  const canonical = new URL(canonicals[0].href)
  assert(canonical.pathname === expectedPath && canonical.search === '', `${route} canonicalized to ${canonical.pathname}${canonical.search}`)
  const alternates = links.filter(link => link.rel === 'alternate' && link.hreflang)
  assert(alternates.length >= 2, `${route} did not expose localized hreflang links`)
  for (const alternate of alternates) {
    const href = new URL(alternate.href)
    assert(href.search === '', `${route} hreflang ${alternate.hreflang} contains a target query`)
    assert(!href.pathname.includes(`/${encodeURIComponent(siteDomain)}`), `${route} hreflang contains a target domain path`)
  }
}

async function expectNotFound(route) {
  const response = await request(route)
  const html = await response.text()
  assert(response.status === 404, `${route} expected HTTP 404, received ${response.status}`)
  assert(/<meta[^>]+name=["']robots["'][^>]+content=["']noindex, nofollow["']/i.test(html)
    || /<meta[^>]+content=["']noindex, nofollow["'][^>]+name=["']robots["']/i.test(html), `${route} 404 HTML is missing noindex, nofollow`)
}

async function expectPrizeNoIndex(route) {
  const response = await request(route)
  assert(response.status === 200, `${route} expected HTTP 200, received ${response.status}`)
  const robots = response.headers.get('x-robots-tag') || ''
  assert(robots.toLowerCase().includes('noindex') && robots.toLowerCase().includes('follow'), `${route} returned X-Robots-Tag: ${robots || '[missing]'}`)
}

async function expectSiteGroupSeo(route) {
  const response = await request(route)
  const html = await response.text()
  assert(response.status === 200, `${route} expected HTTP 200, received ${response.status}`)
  const title = html.match(/<title>(.*?)<\/title>/i)?.[1] || ''
  const description = parseMeta(html).find(meta => meta.name === 'description')?.content || ''
  assert(title && !title.includes('发现兽人文化相关资源与社区'), `${route} retained the default home title`)
  assert(description && !description.includes('GoFurry is a bilingual Furry navigation hub'), `${route} retained the default site description`)
}

async function expectStatus(route, expectedStatus, expectedContentType = '', bodyMarker = '') {
  const response = await request(route)
  const body = await response.text()
  assert(response.status === expectedStatus, `${route} expected HTTP ${expectedStatus}, received ${response.status}`)
  if (expectedContentType) {
    assert(response.headers.get('content-type')?.includes(expectedContentType), `${route} returned an unexpected content type`)
  }
  if (bodyMarker) {
    assert(body.includes(bodyMarker), `${route} did not SSR ${bodyMarker}`)
  }
}

function request(path, options = {}) {
  return fetch(new URL(path, `${baseUrl}/`), {
    headers: { accept: 'text/html,application/xhtml+xml,application/xml' },
    ...options,
  })
}

function parseLinks(html) {
  return Array.from(html.matchAll(/<link\b[^>]*>/gi), ([tag]) => {
    const attributes = {}
    for (const [, name, value] of tag.matchAll(/([\w:-]+)=["']([^"']*)["']/g)) {
      attributes[name.toLowerCase()] = value
    }
    return attributes
  })
}

function parseMeta(html) {
  return Array.from(html.matchAll(/<meta\b[^>]*>/gi), ([tag]) => {
    const attributes = {}
    for (const [, name, value] of tag.matchAll(/([\w:-]+)=["']([^"']*)["']/g)) {
      attributes[name.toLowerCase()] = value
    }
    return attributes
  })
}

function decodeXml(value) {
  return value
    .replaceAll('&amp;', '&')
    .replaceAll('&lt;', '<')
    .replaceAll('&gt;', '>')
    .replaceAll('&quot;', '"')
    .replaceAll('&apos;', "'")
}

function normalizeBaseUrl(value) {
  return value.replace(/\/$/, '')
}

function parseArgs(values) {
  const parsed = {}
  for (let index = 0; index < values.length; index += 1) {
    const value = values[index]
    if (!value.startsWith('--')) continue
    const [name, inlineValue] = value.slice(2).split('=', 2)
    parsed[name] = inlineValue ?? values[index + 1]
    if (inlineValue === undefined) index += 1
  }
  return parsed
}

function assert(condition, message) {
  if (!condition) throw new Error(message)
}
