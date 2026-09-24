import { test as base, expect, type Locator, type Page } from '@playwright/test'
import { readFile } from 'node:fs/promises'
import { setTimeout as delay } from 'node:timers/promises'
import { startInsightsFixtureApp } from '../../../scripts/fixtures/insights-app.mjs'
import { mockGameHome } from '../../../scripts/fixtures/insights-overview.mjs'
import { STEAM_SHARED_CDN_GROUP_PREFIXES } from '../../../app/utils/steamAssets'

export const origins = { primary: 'https://primary.example', mirror: 'https://mirror.example' }
export const iconKey = 'nav/sites/1/icon/' + 'a'.repeat(32) + '.svg'
export const nextIconKey = iconKey.replace(/a{32}/, 'b'.repeat(32))
export const heroKey = 'nav/hero/desktop/' + 'a'.repeat(32) + '.avif'
const patternKey = 'nav/patterns/' + 'a'.repeat(32) + '.svg'
export const steamSource = 'https://shared.steamstatic.com/store_item_assets/steam/apps/570/header.jpg'
const svg = '<svg xmlns="http://www.w3.org/2000/svg" width="120" height="45"><rect width="120" height="45" fill="#78644b"/></svg>'

interface NetworkState {
  primary: number
  mirror: number
  china: number
  global: number
  failProbes: boolean
  blockedURL: string | null
  probeRequests: string[]
}

interface RoutingScenario {
  state: NetworkState
  expectedNetworkFailures: Set<string>
  releaseProbes(): void
  open(path?: string, options?: {
    cookies?: Record<string, string>
    legacySteamRecommendation?: { group: 'global'; testedAt: number }
  }): Promise<void>
  openResourceRouting(): Promise<void>
  savePreferences(): Promise<void>
  cancelPreferences(): Promise<void>
  settleTab(index: number): Promise<void>
  cookieValue(name: string): Promise<string | undefined>
  waitForDiagnostics(kind: 'managed' | 'steam', selected?: string): Promise<void>
}

type RoutingApp = Awaited<ReturnType<typeof startInsightsFixtureApp>>

export const test = base.extend<{ routing: RoutingScenario }, { routingApp: RoutingApp }>({
  // Only stable API payloads and the production servers belong to the worker.
  // eslint-disable-next-line no-empty-pattern -- Playwright requires fixture argument destructuring.
  routingApp: [async ({}, use) => {
    const app = await startInsightsFixtureApp((url: URL) => {
      if (url.pathname === '/api/v2/nav/home') return { data: {
        schema_version: 4, generated_at: '2026-09-01T12:00:00Z', cache_state: {},
        groups: [{ id: '1', name: 'Resource fixtures', info: '', priority: 1, sites: [{ id: '1', name: 'Fixture site', domain: 'example.com', info: 'Routing', icon: iconKey, nsfw: '0', welfare: '0', view_count: 0 }] }],
        spotlight: { page_size: 6, featured: [], popular: [], latest: [], random: [] },
        saying: null, ping: {}, hero: { desktop: { id: '1', object_key: heroKey }, mobile: null },
      } }
      if (url.pathname === '/api/v2/nav/appearance/patterns') return { data: {
        schema_version: 1, patterns: [{ id: '1', name: 'Fixture pattern', name_en: 'Fixture pattern', object_key: patternKey, light_color: '#123456', dark_color: '#abcdef', light_opacity: .12, dark_opacity: .08, default_size_px: 120 }],
      } }
      if (url.pathname.endsWith('/game/home')) {
        const home = mockGameHome()
        for (const game of home.panel.latest_games) game.header_url = steamSource + '?id=' + game.id
        return { data: home }
      }
      return { data: [] }
    }, { NUXT_PUBLIC_ASSET_PRIMARY_BASE: origins.primary, NUXT_PUBLIC_ASSET_MIRROR_BASE: origins.mirror })
    try { await use(app) }
    finally { await app.close() }
  }, { scope: 'worker' }],
  baseURL: async ({ routingApp }, use) => { await use(routingApp.base) },
  routing: async ({ routingApp, page, context }, use, testInfo) => {
    // Every test receives fresh controls, gate, request evidence and browser storage.
    const state: NetworkState = { primary: 160, mirror: 15, china: 180, global: 15, failProbes: false, blockedURL: null, probeRequests: [] }
    const expectedNetworkFailures = new Set<string>()
    let releaseProbes!: () => void
    const gate = new Promise<void>(resolve => { releaseProbes = resolve })
    const probeBody = await readFile(new URL('../../fixtures/cdn-probe.bin', import.meta.url))
    const settleTab = async (index: number) => {
      await page.waitForFunction(index => {
        const pages = document.querySelector('.preferences-pages')
        return pages && Math.abs(pages.scrollLeft - pages.clientWidth * index) < 2
          && document.querySelectorAll('.preferences-tabs [role=tab]')[index]?.getAttribute('aria-selected') === 'true'
      }, index)
    }
    const finishPreferences = async (button: string) => {
      await page.locator(`.gf-modal__header-actions .gf-button--${button}`).click()
      await expect(page.locator('[data-resource-routing]')).toHaveCount(0)
    }
    try {
      await use({
        state, expectedNetworkFailures, releaseProbes, settleTab,
        async open(path = '/', options = {}) {
          await context.addCookies(Object.entries(options.cookies || {}).map(([name, value]) => ({ name, value, url: routingApp.base })))
          await context.addInitScript(({ baseURL, legacy }) => {
            if (location.origin !== baseURL) return
            localStorage.setItem('gf_background_preference', JSON.stringify({ version: 1, source: 'server', pattern_id: '1', overrides: {} }))
            if (legacy) localStorage.setItem('gofurry:steam-shared-cdn-preference:v1', JSON.stringify(legacy))
          }, { baseURL: routingApp.base, legacy: options.legacySteamRecommendation })
          await context.route('**/*', async route => {
            const url = new URL(route.request().url())
            if (url.origin === routingApp.base || ['data:', 'blob:'].includes(url.protocol)) return route.continue()
            const managed = Object.values(origins).includes(url.origin)
            const steamGroup = (['china', 'global'] as const).find(group => STEAM_SHARED_CDN_GROUP_PREFIXES[group].includes(url.origin))
            if (!managed && !steamGroup) return route.abort()
            const probe = managed ? url.pathname === '/system/probes/cdn.bin' : url.searchParams.has('gf_probe')
            try {
              if (probe) {
                state.probeRequests.push(url.href)
                await gate
                const group = managed ? (url.origin === origins.primary ? 'primary' : 'mirror') : steamGroup!
                await delay(state[group])
              }
              if ((probe && state.failProbes) || (!probe && state.blockedURL && url.href.includes(state.blockedURL))) {
                expectedNetworkFailures.add(url.href)
                return await route.abort('failed')
              }
              await route.fulfill({
                status: 200, contentType: probe && managed ? 'application/octet-stream' : 'image/svg+xml',
                headers: { 'Access-Control-Allow-Origin': '*' }, body: probe && managed ? probeBody : svg,
              })
            } catch (error) {
              // A timed-out Image can cancel a gated request before it is released.
              if (!String(error).includes('Route is already handled')) throw error
            }
          })
          const response = await page.goto(path, { waitUntil: 'domcontentloaded' })
          expect(response?.status()).toBe(200)
          await page.waitForFunction(() => Boolean((document.querySelector('#__nuxt') as Element & { __vue_app__?: unknown })?.__vue_app__))
        },
        async openResourceRouting() {
          await page.locator('.gf-nav__mode-button').click()
          await page.locator('.preferences-tabs').getByRole('tab').last().click()
          await settleTab(2)
        },
        savePreferences: () => finishPreferences('primary'),
        cancelPreferences: () => finishPreferences('ghost'),
        cookieValue: async name => (await context.cookies()).find(cookie => cookie.name === name)?.value,
        async waitForDiagnostics(kind, selected) {
          await page.waitForFunction(({ key, selected }) => {
            const diagnostics = JSON.parse(localStorage.getItem(key) || 'null')
            return diagnostics && (!selected || diagnostics.selected === selected)
          }, { key: kind === 'managed' ? 'gf_asset_cdn_diagnostics' : 'gf_steam_asset_diagnostics_v1', selected })
        },
      })
    } finally {
      // Release probes even after a failed assertion. Playwright owns this test's
      // context and routes: unrouteAll(wait) can stall on an already-handled probe
      // and prevent the runner from reaching context.close().
      releaseProbes()
      if (testInfo.status !== testInfo.expectedStatus) {
        await testInfo.attach('resource-routing-fixture.log', { body: routingApp.logs(), contentType: 'text/plain' })
        await testInfo.attach('resource-routing-network.json', { body: JSON.stringify({ ...state, expectedNetworkFailures: [...expectedNetworkFailures] }, null, 2), contentType: 'application/json' })
      }
    }
  },
})

export { expect }

export async function expectLoaded(image: Locator) {
  await expect.poll(() => image.evaluate(el => el instanceof HTMLImageElement && el.complete && el.naturalWidth > 0)).toBe(true)
}

export function readManagedResources(page: Page) {
  return page.evaluate(() => ({
    icon: document.querySelector<HTMLImageElement>('.nav-site-card__logo img')!.src,
    hero: document.querySelector<HTMLImageElement>('.nav-header__background--managed img')!.currentSrc,
    pattern: getComputedStyle(document.querySelector('.gf-public-background__pattern')!).maskImage,
  }))
}

// Test-only bridge for simulating a new prop on the same mounted resource
// component. Not a general Vue introspection API.
export async function changeRenderedResourceProp(locator: Locator, name: 'objectKey' | 'src', value: string) {
  await locator.evaluate((element, { name, value }) => {
    interface Instance { subTree: VNode; props: Record<string, unknown> }
    interface VNode { el?: Element; component?: Instance; suspense?: { activeBranch?: VNode }; children?: VNode[] }
    const app = (document.querySelector('#__nuxt') as Element & { __vue_app__: { _container: { _vnode: VNode } } }).__vue_app__
    const find = (vnode?: VNode): Instance | null => {
      if (!vnode || typeof vnode !== 'object') return null
      const instance = vnode.component
      if (instance?.subTree.el === element && name in instance.props) return instance
      if (instance) { const found = find(instance.subTree); if (found) return found }
      if (vnode.suspense) { const found = find(vnode.suspense.activeBranch); if (found) return found }
      for (const child of Array.isArray(vnode.children) ? vnode.children : []) { const found = find(child); if (found) return found }
      return null
    }
    const instance = find(app._container._vnode)
    if (!instance) throw new Error('Resource component was not found')
    instance.props[name] = value
  }, { name, value })
}
