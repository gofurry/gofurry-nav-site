import assert from 'node:assert/strict'
import { launchPerfBrowser } from './perf/shared.mjs'
import { startInsightsFixtureApp } from './fixtures/insights-app.mjs'
import { mockGameHome } from './fixtures/insights-overview.mjs'

// A populated catalog with no finalized Facts is a valid development/new-game
// state. Exercise both older null arrays and the corrected empty-array response.
const state = { legacyNull: true, failure: false }
const game = {
  id: '82', appid: 82, name: 'SSR Fixture Game', summary: 'SSR fixture description',
  about_the_game: '<p>SSR fixture introduction</p>',
  site: { view_count: 1, resources: [], groups: [], links: [] },
  platforms: { windows: true }, prices: [], news: [], tags: [], developers: [], publishers: [],
  media: { screenshots: [], movies: [], assets: [] }, requirements: {}, support_info: {}, extra: {},
}
const app = await startInsightsFixtureApp((url, media) => {
  if (url.pathname.endsWith('/game/info')) {
    if (state.failure) return { status: 503 }
    if (url.searchParams.get('id') !== '82') return { status: 404 }
    return { data: game }
  }
  if (url.pathname.endsWith('/game/home')) return { data: mockGameHome(media) }
  if (url.pathname.endsWith('/games/82/insights')) return { data: {
    game: { id: 82, name: game.name },
    state: { free: null, windows: null, mac: null, linux: null, release: null, as_of: null },
    players: { current: null, peak_30d: null, average_30d: null, as_of: null, observed_days_30d: 0, sample_coverage_30d: null },
    price: null, regional_prices: { as_of: null, regions: state.legacyNull ? null : [] }, recent_changes: [],
  } }
  if (/insights\/(players|prices)$/.test(url.pathname)) return { data: { region: 'CN', as_of: null, available_from: null, points: [] } }
  if (url.pathname.endsWith('/reviews')) return { data: { total: 0, remarks: [] } }
  return { data: [] }
})
let browser
try {
  for (const path of ['/games', '/en/games', '/games/82', '/en/games/82']) {
    const response = await fetch(app.base + path)
    const html = await response.text()
    assert.equal(response.status, 200, path)
    assert.match(html, /<title>[^<]+<\/title>/, 'SSR title missing')
    assert.match(html, /<meta[^>]+name="description"[^>]+content="[^"]+"/, 'SSR description missing')
    const canonicals = [...html.matchAll(/<link[^>]+rel="canonical"[^>]+href="([^"]+)"/g)]
    assert.equal(canonicals.length, 1, 'canonical must remain unique')
    assert.equal(new URL(canonicals[0][1]).pathname, path)
    assert.match(html, /hreflang="en-US"/, 'localized alternate missing')
    assert(!/<meta[^>]+name="robots"[^>]+content="[^"]*noindex/.test(html), 'public page became noindex')
    const content = html.replace(/<script\b[^>]*>[\s\S]*?<\/script>/gi, '')
    if (path.endsWith('/82')) {
      assert.match(content, /<h1[^>]*>\s*SSR Fixture Game\s*<\/h1>/, 'detail heading missing from SSR HTML')
      assert(content.includes('<p>SSR fixture introduction</p>'), 'detail introduction missing from SSR HTML')
    } else assert.match(content, />\s*Active game fixture\s*</, 'homepage content missing from SSR HTML')
    console.log(`[game-detail] SSR metadata, canonical, alternates and content ${path} PASS`)
  }
  browser = await launchPerfBrowser()
  for (const legacyNull of [true, false]) {
    state.legacyNull = legacyNull
    const page = await browser.newPage({ viewport: { width: 1200, height: 900 } })
    const errors = []
    page.on('pageerror', error => errors.push(error.message))
    await page.goto(app.base + '/games/82', { waitUntil: 'domcontentloaded' })
    await page.waitForFunction(() => Boolean(document.querySelector('#__nuxt')?.__vue_app__))
    for (const [tab, label] of [['insights', '生态观测'], ['gallery', '画廊'], ['detail', '详情'], ['intro', '介绍'], ['insights', '生态观测']]) {
      await page.locator(`[data-game-tab="${tab}"]`).click()
      await page.waitForFunction(label => document.querySelector('.game-detail-tab--active')?.textContent.trim() === label, label)
      if (tab === 'insights') {
        await page.locator('[data-game-insights]').waitFor({ state: 'visible' })
        assert((await page.locator('[data-current-players]').innerText()).includes('暂无数据'), 'missing players became zero')
        assert.equal(await page.locator('[data-price-kind="missing"]').count(), 3)
      } else assert.equal(await page.locator('[data-game-insights]').isVisible(), false)
    }
    assert.deepEqual(errors, [], 'tab switching raised a rendering exception')
    await page.close()
    console.log(`[game-detail] ${legacyNull ? 'null' : 'empty'} regions, missing facts and repeated tab switching PASS`)
  }
  for (const path of ['/games/999999', '/en/games/999999', '/games/abc']) {
    assert.equal((await fetch(app.base + path)).status, 404, path)
  }
  state.failure = true
  for (const path of ['/games/82', '/en/games/82']) assert.equal((await fetch(app.base + path)).status, 503, path)
  console.log('[game-detail] authoritative 404/503 semantics PASS')
} finally {
  await browser?.close()
  await app.close()
}
