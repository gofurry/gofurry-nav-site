import { test, expect } from '../fixtures/seo-recovery'
import { test as transitionTest } from '../fixtures/transition-mode'
import { settleRuntime } from '../fixtures/insights-runtime'

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
    for (const suffix of ['', '?domain=target.example', '?domain=alt.example&tab=observation&view=dns',
      '?tab=security&view=tls', '?domain=alt.example&tab=insights&metric=certificate_verified&range=90d',
      '?tab=banana&view=banana&metric=banana&range=banana']) {
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
  expect(paths).toEqual(expect.arrayContaining(['/site/41', '/en/site/41', '/games/82', '/en/games/82', '/site-groups/12', '/en/site-groups/12', '/updates/109', '/en/updates/109']))
  expect(new Set(paths).size).toBe(paths.length)
  for (const path of paths) {
    if (/^\/(en\/)?site\//.test(path)) expect(new URL(path, runtime.app.base).search).toBe('')
    expect(path).not.toMatch(/^\/(en\/)?sites\/|^\/(en\/)?site\/[^/]+\/.+|\?domain=/)
    expect(path).not.toMatch(/^\/(en\/)?(?:games\/(?:search|prize(?:\/activation)?)|steam|insights\/(?:sites|games)\/compare)$/)
  }
  expect(xml).not.toMatch(/%7b%22|%7B%22|\{&quot;domain&quot;/)
  expect(runtime.calls.map(call => call.url.pathname).sort()).toEqual(['/api/v2/game/list', '/api/v2/nav/site-groups', '/api/v2/nav/sites/index', '/api/v2/nav/updates'])
  expect(runtime.calls.find(call => call.url.pathname.endsWith('/updates'))!.url.search).toBe('?lang=zh')
  runtime.assertQuiet()
})
for (const failure of ['site', 'game', 'sitemap', 'updates'] as const) test('SEO fails closed on ' + failure + ' upstream failure', async ({ request, runtime }) => {
  runtime.state.failure = failure
  const response = await request.get(['sitemap', 'updates'].includes(failure) ? '/sitemap.xml' : failure === 'site' ? '/site/41' : '/games/82')
  expect(response.status()).toBe(503)
  for (const invalid of ['/site/abc', '/games/abc']) expect((await request.get(invalid)).status()).toBe(404)
  runtime.assertQuiet()
})
for (const prefix of ['', '/en']) {
  transitionTest('transition direct access, aliases and infrastructure ' + (prefix || 'zh'), async ({ request, runtime }) => {
    for (const path of ['/games', '/games/search', '/games/prize', '/games/collections', '/updates', '/updates/109', '/about', '/about/faolan', '/privacy', '/terms', '/sites', '/nav', '/insights', '/site-groups/12', '/unknown']) {
      const response = await request.get(prefix + path, { maxRedirects: 0 })
      expect(response.status(), path).toBe(503)
      expect(response.headers()['retry-after']).toBe('3600')
      expect(response.headers()['cache-control']).toBe('no-store')
      const html = await response.text()
      expect(html).toContain('网站目前处于建设中，部分内容与服务暂未开放。')
      expect(html).not.toContain('__NUXT_DATA__')
      expect(response.headers().location).toBeUndefined()
    }
    for (const path of ['/sites/41', '/site/41/target.example']) {
      const response = await request.get(prefix + path, { maxRedirects: 0 })
      expect(response.status()).toBe(301)
      expect(response.headers().location).toContain(prefix + '/site/41')
    }
    expect(await (await request.get('/healthz')).text()).toBe('ok')
    expect((await request.get('/defaultLogo.svg')).status()).toBe(200)
    expect((await request.get('/robots.txt')).status()).toBe(200)
    expect(runtime.calls).toHaveLength(0)
    runtime.assertQuiet()
  })
  for (const width of [1440, 390]) {
    transitionTest(`transition Home and base details ${prefix || 'zh'} ${width}`, async ({ page, runtime }, testInfo) => {
      await page.setViewportSize({ width, height: 900 })
      const response = await page.goto(prefix || '/')
      expect(response?.status()).toBe(200)
      const html = await response!.text()
      expect(html).toContain('data-transition-home')
      expect(html).not.toContain('nav-content-shell')
      await page.waitForFunction(() => Boolean((document.querySelector('#__nuxt') as Element & { __vue_app__?: unknown })?.__vue_app__))
      await settleRuntime(page)
      const platforms = page.locator('[data-transition-platforms]')
      await expect(platforms.locator('button')).toHaveCount(5)
      await expect(platforms.locator('button').first()).toHaveText(prefix ? 'Bing' : '必应')
      const platformBox = await platforms.boundingBox(), inputBox = await page.locator('.search-input').boundingBox()
      expect(platformBox!.y + platformBox!.height).toBeLessThan(inputBox!.y)
      await expect(page.locator('.search-category-row, .search-platform-picker, .nav-header__quick-access, .nav-header__scroll-hint')).toHaveCount(0)
      await expect(page.locator('.nav-content-shell, .gf-navbar, .mobile-bottom-tab-bar, .page-scroll-dock')).toHaveCount(0)
      await page.locator('.search-input').fill('wolf')
      await expect(page.locator('.search-suggestion-list')).toHaveCount(0)
      const footer = page.locator('[data-transition-footer]')
      await expect(footer).toContainText('© 2026 成都远岸开物科技有限公司（官网建设中，暂无跳转） · 蜀ICP备2025158691号-2')
      await expect(footer.locator('a')).toHaveCount(1)
      await expect(footer.locator('a')).toHaveAttribute('href', 'https://beian.miit.gov.cn/')
      const box = await footer.boundingBox()
      expect(box!.y + box!.height).toBeLessThanOrEqual(901)
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
      if (!prefix) await page.screenshot({ path: testInfo.outputPath('transition-home.png'), fullPage: true })
      for (const [path, name] of [['/site/41?tab=insights', 'Site fixture 41'], ['/games/82', 'Game fixture 82']]) {
        const detailResponse = await page.goto(prefix + path)
        expect(detailResponse?.status()).toBe(200)
        const detailHTML = await detailResponse!.text()
        expect(detailHTML).not.toContain('Other profile sentinel')
        await expect(page.locator('[data-transition-detail] h1')).toHaveText(name!)
        await page.waitForFunction(() => Boolean((document.querySelector('#__nuxt') as Element & { __vue_app__?: unknown })?.__vue_app__))
        await settleRuntime(page)
        await expect(page.locator('[data-transition-footer]')).toBeVisible()
        await expect(page.locator('[data-site-detail], .game-detail-main, [data-game-tab]')).toHaveCount(0)
      }
      expect(runtime.calls.map(call => call.url.pathname)).toEqual(['/api/v2/nav/home/hero', '/api/v2/nav/sites/directory', '/api/v2/game/info'])
      runtime.assertQuiet()
    })
  }
}

transitionTest('transition sitemap includes only open localized pages', async ({ request, runtime }) => {
  const response = await request.get('/sitemap.xml')
  expect(response.status()).toBe(200)
  const xml = await response.text()
  const paths = [...xml.matchAll(/<loc>(.*?)<\/loc>/g)].map(match => new URL(match[1]!).pathname)
  expect(paths.sort()).toEqual(['/', '/en', '/site/41', '/en/site/41', '/games/82', '/en/games/82'].sort())
  expect(runtime.calls.map(call => call.url.pathname).sort()).toEqual(['/api/v2/game/list', '/api/v2/nav/sites/index'])
  runtime.assertQuiet()
})

transitionTest('transition client navigation cannot mount a paused page', async ({ page, runtime }) => {
  await page.goto('/')
  await page.waitForFunction(() => Boolean((document.querySelector('#__nuxt') as Element & { __vue_app__?: unknown })?.__vue_app__))
  await settleRuntime(page)
  const failure = await page.evaluate(async () => {
    const app = (document.querySelector('#__nuxt') as Element & { __vue_app__: {
      config: { globalProperties: { $nuxt: { $router: { push: (to: string) => Promise<unknown> } } } }
    } }).__vue_app__.config.globalProperties.$nuxt
    try { await app.$router.push('/games'); return null }
    catch (error) {
      const failure = error as { statusCode: number; message: string }
      return { statusCode: failure.statusCode, message: failure.message }
    }
  })
  expect(failure).toEqual({ statusCode: 503, message: '网站目前处于建设中，部分内容与服务暂未开放。' })
  await expect(page.locator('main')).toContainText('网站目前处于建设中，部分内容与服务暂未开放。')
  await expect(page.locator('[data-transition-home]')).toHaveCount(0)
  await expect(page.locator('.game-home-page, .game-detail-main')).toHaveCount(0)
  expect(runtime.calls.map(call => call.url.pathname)).toEqual(['/api/v2/nav/home/hero'])
  runtime.assertQuiet()
})
