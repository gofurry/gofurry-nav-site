import { test as base, expect, type Page } from '@playwright/test'
import { fileURLToPath } from 'node:url'
import { startInsightsFixtureApp } from '../../../scripts/fixtures/insights-app.mjs'
import { mockGameHome } from '../../../scripts/fixtures/insights-overview.mjs'
import { STEAM_SHARED_CDN_GROUP_PREFIXES } from '../../../app/utils/steamAssets'
import { STEAM_PROBE_PATHS } from '../../../app/utils/steamAssetRouting'

interface GameFixtureState {
  legacyNull: boolean
  failure: boolean
  gallery: boolean
  adult: boolean
}

const defaultState = (): GameFixtureState => ({ legacyNull: true, failure: false, gallery: false, adult: false })
const catalogGame = {
  id: '82', appid: 82, name: 'SSR Fixture Game', summary: 'SSR fixture description',
  about_the_game: '<p>SSR fixture introduction</p>',
  site: { view_count: 1, resources: [], groups: [], links: [] },
  platforms: { windows: true }, prices: [], news: [], tags: [], developers: [], publishers: [],
  media: { screenshots: [], movies: [], assets: [] }, requirements: {}, support_info: {}, extra: {},
}

interface GameFixture {
  state: GameFixtureState
  app: Awaited<ReturnType<typeof startInsightsFixtureApp>>
}

export const test = base.extend<{ game: GameFixture }, { gameApp: GameFixture }>({
  // Server ownership is independent of browser/context ownership.
  // eslint-disable-next-line no-empty-pattern -- Playwright requires fixture argument destructuring.
  gameApp: [async ({}, use) => {
    const state = defaultState()
    const app = await startInsightsFixtureApp((url: URL, media: string) => {
      if (url.pathname.endsWith('/game/info')) {
        if (state.failure) return { status: 503 }
        if (url.searchParams.get('id') !== '82') return { status: 404 }
        return { data: state.gallery ? { ...catalogGame,
          tags: state.adult ? [{ id: '777777', code: 'adult', category_code: 'classification', role: 'normal', name: 'Adult' }] : [],
          media: { ...catalogGame.media, screenshots: Array.from({ length: 24 }, (_, i) => ({
            id: i + 1, thumbnail_url: `${media}/shot-${i}.svg`, url: `${media}/shot-${i}.svg`,
          })) },
        } : catalogGame }
      }
      if (url.pathname.endsWith('/game/home')) return { data: mockGameHome(media) }
      if (url.pathname.endsWith('/games/82/insights')) return { data: {
        game: { id: 82, name: catalogGame.name },
        state: { free: null, windows: null, mac: null, linux: null, release: null, as_of: null },
        players: { current: null, peak_30d: null, average_30d: null, as_of: null, observed_days_30d: 0, sample_coverage_30d: null },
        price: null, regional_prices: { as_of: null, regions: state.legacyNull ? null : [] }, recent_changes: [],
      } }
      if (/insights\/(players|prices)$/.test(url.pathname)) return { data: { region: 'CN', as_of: null, available_from: null, points: [] } }
      if (url.pathname.endsWith('/reviews')) return { data: { total: 0, remarks: [] } }
      return { data: [] }
    })
    try { await use({ state, app }) }
    finally { await app.close() }
  }, { scope: 'worker' }],
  game: [async ({ gameApp }, use, testInfo) => {
    Object.assign(gameApp.state, defaultState())
    gameApp.app.requests.length = 0
    try { await use(gameApp) }
    finally {
      if (testInfo.status !== testInfo.expectedStatus) {
        await testInfo.attach('game-fixture.log', { body: gameApp.app.logs(), contentType: 'text/plain' })
      }
      Object.assign(gameApp.state, defaultState())
      gameApp.app.requests.length = 0
    }
  }, { auto: true }],
  baseURL: async ({ game }, use) => { await use(game.app.base) },
  page: async ({ page, game }, use) => {
    // Keep the real routing plugins running, but make their background network
    // probes deterministic. Game Detail is not external CDN acceptance.
    await page.route(url => url.origin === game.app.upstreamUrl && url.pathname === '/system/probes/cdn.bin', route => route.fulfill({
      contentType: 'application/octet-stream', headers: { 'access-control-allow-origin': '*' },
      path: fileURLToPath(new URL('../../fixtures/cdn-probe.bin', import.meta.url)),
    }))
    const steamOrigins = Object.values(STEAM_SHARED_CDN_GROUP_PREFIXES).flat()
    await page.route(url => steamOrigins.includes(url.origin) && STEAM_PROBE_PATHS.includes(url.pathname), route => route.fulfill({
      contentType: 'image/svg+xml', body: '<svg xmlns="http://www.w3.org/2000/svg" width="120" height="45"><rect width="120" height="45" fill="#78644b"/></svg>',
    }))
    await use(page)
  },
})

export { expect }

export async function openGame(page: Page) {
  const response = await page.goto('/games/82', { waitUntil: 'domcontentloaded' })
  expect(response?.status()).toBe(200)
  await page.waitForFunction(() => Boolean((document.querySelector('#__nuxt') as Element & { __vue_app__?: unknown })?.__vue_app__))
}

export async function selectTab(page: Page, tab: string) {
  await page.locator(`[data-game-tab="${tab}"]`).click()
  await expect(page.locator('.game-detail-tab--active')).toHaveAttribute('data-game-tab', tab)
}

export async function expectMissingFacts(page: Page) {
  await expect(page.locator('[data-game-insights]')).toBeVisible()
  await expect(page.locator('[data-current-players]')).toContainText('暂无数据')
  await expect(page.locator('[data-price-kind="missing"]')).toHaveCount(3)
}
