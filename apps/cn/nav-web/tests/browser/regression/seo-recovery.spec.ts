import { test, expect } from '../fixtures/seo-recovery'

function attributes(tag: string) {
  return Object.fromEntries([...tag.matchAll(/([\w:-]+)=(?:"([^"]*)"|'([^']*)')/g)]
    .map(([, key, doubleQuoted, singleQuoted]) => [key!.toLowerCase(), doubleQuoted ?? singleQuoted!]))
}
function links(html: string) { return [...html.matchAll(/<link\b[^>]*>/gi)].map(([tag]) => attributes(tag)) }
for (const prefix of ['', '/en']) {
  test('SEO noindex, canonical and legacy redirects ' + (prefix || 'zh'), async ({ request, runtime }) => {
    for (const suffix of ['/games/search', '/games/prize', '/games/prize/activation']) {
      const response = await request.get(prefix + suffix, { maxRedirects: 0 })
      expect(response.status()).toBe(200)
      expect(response.headers()['x-robots-tag']?.split(',').map(value => value.trim())).toEqual(expect.arrayContaining(['noindex', 'follow']))
    }
    const home = await request.get(prefix || '/'), homeLinks = links(await home.text())
    expect(home.status()).toBe(200)
    const homeCanonical = homeLinks.filter(link => link.rel === 'canonical')
    expect(homeCanonical).toHaveLength(1)
    expect(new URL(homeCanonical[0]!.href!).href).toBe('https://go-furry.com' + (prefix || '/'))
    expect(homeLinks.filter(link => link.rel === 'alternate' && link.hreflang).length).toBeGreaterThanOrEqual(2)
    const group = await request.get(prefix + '/site-groups/12'), groupHTML = await group.text()
    expect(group.status()).toBe(200)
    expect(groupHTML).toContain('site-group-page')
    expect(groupHTML.match(/<title>(.*?)<\/title>/i)?.[1]).toContain('Fixture community')
    const description = [...groupHTML.matchAll(/<meta\b[^>]*>/gi)]
      .map(([tag]) => attributes(tag))
      .find(meta => meta.name === 'description')?.content
    expect(description).toContain('Isolated group summary')
    expect(description).toContain('Fixture community')
    for (const [path, domain] of [['/sites/41', ''], ['/site/41/target.example', 'target.example']]) {
      const response = await request.get(prefix + path, { maxRedirects: 0 })
      expect(response.status()).toBe(301)
      const target = new URL(response.headers().location!, runtime.app.base)
      expect(target.pathname).toBe(prefix + '/site/41')
      expect(target.searchParams.get('domain') || '').toBe(domain)
      if (!domain) expect(target.search).toBe('')
    }
    for (const suffix of ['', '?domain=target.example']) {
      const response = await request.get(prefix + '/site/41' + suffix), html = await response.text()
      expect(response.status()).toBe(200); expect(html).toContain('site-detail-page')
      const pageLinks = links(html), canonical = pageLinks.filter(link => link.rel === 'canonical')
      expect(canonical).toHaveLength(1)
      expect(new URL(canonical[0]!.href!).pathname).toBe(prefix + '/site/41')
      expect(new URL(canonical[0]!.href!).search).toBe('')
      const alternates = pageLinks.filter(link => link.rel === 'alternate' && link.hreflang)
      expect(alternates.length).toBeGreaterThanOrEqual(2)
      for (const link of alternates) { expect(new URL(link.href!).search).toBe(''); expect(link.href).not.toContain('target.example') }
    }
    const game = await request.get(prefix + '/games/82')
    expect(game.status()).toBe(200); expect(await game.text()).toContain('game-detail-page')
    runtime.assertQuiet()
  })
  test('SEO authoritative missing and removed Workshop routes ' + (prefix || 'zh'), async ({ request, runtime }) => {
    for (const route of ['/site/abc', '/games/abc', '/site/999999999', '/games/999999999', '/site/41?domain=invalid.example', '/steam',
      '/workshop', '/workshop/tools', '/workshop/developer', '/workshop/discussion']) {
      const response = await request.get(prefix + route, { maxRedirects: 0 }), html = await response.text()
      expect(response.status(), route).toBe(404)
      expect(response.headers().location).toBeUndefined()
      expect(html).toMatch(/<meta[^>]+(?:name=["']robots["'][^>]+content=["']noindex, nofollow["']|content=["']noindex, nofollow["'][^>]+name=["']robots["'])/i)
    }
    runtime.assertQuiet()
  })
}
test('SEO sitemap serializes only canonical localized public inventory', async ({ request, runtime }) => {
  const response = await request.get('/sitemap.xml'), xml = await response.text()
  expect(response.status()).toBe(200); expect(response.headers()['content-type']).toContain('application/xml')
  const urls = [...xml.matchAll(/<loc>(.*?)<\/loc>/g)].map(match => new URL(match[1]!.replaceAll('&amp;', '&')))
  const paths = urls.map(url => url.pathname + url.search)
  expect(paths).toEqual(expect.arrayContaining(['/site/41', '/en/site/41', '/games/82', '/en/games/82', '/site-groups/12', '/en/site-groups/12']))
  expect(new Set(paths).size).toBe(paths.length)
  for (const path of paths) {
    expect(path).not.toMatch(/^\/(en\/)?sites\/|^\/(en\/)?site\/[^/]+\/.+|\?domain=/)
    expect(path).not.toMatch(/^\/(en\/)?(?:games\/(?:search|prize(?:\/activation)?)|steam|insights\/(?:sites|games)\/compare)$/)
  }
  expect(xml).not.toMatch(/%7b%22|%7B%22|\{&quot;domain&quot;/)
  expect(runtime.calls.map(call => call.url.pathname).sort()).toEqual(['/api/v2/game/list', '/api/v2/nav/site-groups', '/api/v2/nav/sites/index'])
  runtime.assertQuiet()
})
for (const failure of ['site', 'game', 'sitemap'] as const) test('SEO fails closed on ' + failure + ' upstream failure', async ({ request, runtime }) => {
  runtime.state.failure = failure
  const response = await request.get(failure === 'sitemap' ? '/sitemap.xml' : failure === 'site' ? '/site/41' : '/games/82')
  expect(response.status()).toBe(503)
  for (const invalid of ['/site/abc', '/games/abc']) expect((await request.get(invalid)).status()).toBe(404)
  runtime.assertQuiet()
})
