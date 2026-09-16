import assert from 'node:assert/strict'
import { startInsightsFixtureApp } from './fixtures/insights-app.mjs'
import { launchPerfBrowser } from './perf/shared.mjs'

const searchBodies = []
const app = await startInsightsFixtureApp((url, _media, body) => {
  if (url.pathname.endsWith('/game/tag-categories')) return { data: [{
    id: '99', code: 'future-category', name: 'Catalog Category', tags: [{
      id: '812345', code: 'sample-tag', name: 'Leaf Tag', category_id: '99',
      category_code: 'future-category', category_name: 'Catalog Category', game_count: 4,
    }],
  }] }
  if (url.pathname.endsWith('/game/search/page')) { searchBodies.push(body); return { data: { total: 0, list: [] } } }
  return { data: [] }
})
let browser
try {
  browser = await launchPerfBrowser()
  for (const width of [390, 1440]) {
    for (const path of ['/games/search', '/en/games/search']) {
      const page = await browser.newPage({ viewport: { width, height: 900 } })
      const errors = []
      page.on('pageerror', error => errors.push(error.message))
      await page.goto(app.base + path, { waitUntil: 'networkidle' })
      await page.locator('.search-filter-button').click()
      await page.waitForSelector('.game-search-filter-chip')
      assert.equal(await page.locator('.game-search-filter-group-title').textContent().then(v => v.trim()), 'Catalog Category')
      const chip = page.locator('.game-search-filter-group-title').locator('..').locator('.game-search-filter-chip')
      assert.equal(await chip.count(), 1, 'category became an assignable chip')
      assert.match(await chip.textContent(), /Leaf Tag/)
      await chip.click()
      await page.waitForSelector('.game-search-filter-chip--active')
      await page.locator('.game-search-filter-action--primary').click()
      await page.waitForURL(url => url.search.length > 0)
      await page.waitForLoadState('networkidle')
      assert(searchBodies.some(body => body.tag_list?.includes(812345)), 'selected leaf ID not sent to search')
      await page.locator('.search-filter-button').click()
      await page.waitForSelector('.game-search-filter-chip--active')
      assert.deepEqual(errors, [])
      await page.close()
      console.log(`[game-tags] ${width}px ${path}: explicit category, leaf-only chip, saved selection PASS`)
    }
  }
  assert(!app.requests.some(url => url.pathname.endsWith('/game/tags')), 'search fetched legacy flat tree input')
} finally {
  await browser?.close()
  await app.close()
}
