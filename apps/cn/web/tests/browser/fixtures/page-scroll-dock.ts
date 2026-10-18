import { test as base, expect, type Locator } from '@playwright/test'
import { startInsightsFixtureApp } from '../../../scripts/fixtures/insights-app.mjs'
import { STEAM_DIAGNOSTICS_KEY, STEAM_PROBE_PATHS } from '../../../app/utils/steamAssetRouting'
import { captureBrowserErrors } from './browser-errors'

type DockApp = Awaited<ReturnType<typeof startInsightsFixtureApp>>
type DockScenario = {
  dock: Locator
  openTerms(): Promise<void>
  maxScroll(): Promise<number>
  scrollProgress(): Promise<number>
  scrollToProgress(percent: number): Promise<void>
  assertQuiet(): void
}

export const test = base.extend<{ scrollDock: DockScenario }, { dockApp: DockApp }>({
  // eslint-disable-next-line no-empty-pattern -- Playwright requires fixture argument destructuring.
  dockApp: [async ({}, use) => {
    // Real Legal content needs no business data; unexpected requests fail.
    const app = await startInsightsFixtureApp(() => ({ status: 500 }))
    try { await use(app) }
    finally { await app.close() }
  }, { scope: 'worker' }],
  baseURL: async ({ dockApp }, use) => { await use(dockApp.base) },
  scrollDock: async ({ page, context, dockApp }, use, testInfo) => {
    const errors = captureBrowserErrors(page)
    const external: string[] = [], failed: string[] = []
    const upstreamStart = dockApp.requests.length
    page.on('requestfailed', request => failed.push(`${request.url()}: ${request.failure()?.errorText}`))
    await context.route('**/*', async route => {
      const url = new URL(route.request().url())
      if (url.origin === dockApp.base || ['data:', 'blob:'].includes(url.protocol)) return route.continue()
      external.push(url.href)
      await route.abort('blockedbyclient')
    })
    const dock = page.locator('.page-scroll-dock')
    const metrics = () => page.evaluate(() => {
      const scroller = document.scrollingElement || document.documentElement || document.body
      const max = scroller.scrollHeight - scroller.clientHeight
      return { max, top: scroller.scrollTop, progress: max > 0 ? scroller.scrollTop / max * 100 : 0 }
    })
    const assertQuiet = () => {
      expect(dockApp.requests.slice(upstreamStart), 'Dock/Legal rendering must not call business APIs').toEqual([])
      expect(external, 'Only bundled same-origin resources are allowed').toEqual([])
      expect(failed, 'No resource failures are expected').toEqual([])
      expect(errors, 'No application or hydration errors are allowed').toEqual([])
    }
    try {
      await use({
        dock, assertQuiet,
        maxScroll: async () => (await metrics()).max,
        scrollProgress: async () => (await metrics()).progress,
        async openTerms() {
          await page.setViewportSize({ width: 1440, height: 900 })
          await context.addInitScript(({ origin, steamDiagnosticsKey, steamSample }) => {
            if (location.origin !== origin) return
            localStorage.setItem('theme', 'light')
            // The real global plugins take their quiet TTL path, with no route pin.
            const checkedAt = Date.now()
            localStorage.setItem('gf_asset_cdn_diagnostics', JSON.stringify({
              selected: 'primary', checkedAt, primaryMs: 10, mirrorMs: 20,
              primaryState: 'success', mirrorState: 'success',
            }))
            localStorage.setItem(steamDiagnosticsKey, JSON.stringify({
              version: 1, selected: 'china', checkedAt, sample: steamSample,
              china: { ms: 30, state: 'success' }, global: { ms: 60, state: 'success' },
            }))
          }, { origin: dockApp.base, steamDiagnosticsKey: STEAM_DIAGNOSTICS_KEY, steamSample: STEAM_PROBE_PATHS[0] })
          const response = await page.goto('/terms', { waitUntil: 'load' })
          expect(response?.status()).toBe(200)
          expect(await response!.text()).toContain('服务条款')
          await page.waitForFunction(() => Boolean((document.querySelector('#__nuxt') as Element & { __vue_app__?: unknown })?.__vue_app__))
          await expect(page.locator('.gf-static-page.legal-page')).toBeVisible()
          await expect(page.locator('html')).not.toHaveClass(/\bdark\b/)
          await expect.poll(() => page.locator('img').evaluateAll(images => images.every(image => image.complete && image.naturalWidth > 0))).toBe(true)
          await page.evaluate(async () => { await document.fonts.ready })
          const initial = await metrics()
          expect(initial.max, 'The real Terms page must naturally exceed the default Dock threshold').toBeGreaterThan(320)
          expect(initial.top).toBe(0)
          await expect(dock).toHaveCount(1)
          assertQuiet()
          await testInfo.attach('dock-document-metrics.json', { body: JSON.stringify(initial), contentType: 'application/json' })
        },
        async scrollToProgress(percent) {
          await page.evaluate(value => {
            const scroller = document.scrollingElement || document.documentElement || document.body
            window.scrollTo({ top: (scroller.scrollHeight - scroller.clientHeight) * value / 100, behavior: 'instant' })
          }, percent)
          // Let the production scroll listener and RAF update observable state.
          await expect.poll(() => dock.textContent()).toBe(`${percent}%`)
          await expect(dock).toHaveAttribute('aria-label', `当前位置 ${percent}%`)
          const actual = await metrics()
          expect(Math.abs(actual.progress - percent)).toBeLessThanOrEqual(100 / actual.max)
        },
      })
      assertQuiet()
    } finally {
      await testInfo.attach('dock-requests.json', {
        body: JSON.stringify({ upstream: dockApp.requests.slice(upstreamStart).map(url => url.href), external, failed, errors }),
        contentType: 'application/json',
      })
      if (testInfo.status !== testInfo.expectedStatus) {
        await testInfo.attach('dock-app.log', { body: dockApp.logs(), contentType: 'text/plain' })
      }
      // No gates/tasks; Playwright owns disposal of test-scoped context/routes.
    }
  },
})

export { expect }
