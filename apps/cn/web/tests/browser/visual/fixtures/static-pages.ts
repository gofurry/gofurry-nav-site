import { test as base, expect } from '@playwright/test'
import { startInsightsFixtureApp } from '../../../../scripts/fixtures/insights-app.mjs'
import { captureBrowserErrors } from '../../fixtures/browser-errors'
import { STEAM_DIAGNOSTICS_KEY, STEAM_PROBE_PATHS } from '../../../../app/utils/steamAssetRouting'

type StaticPageOptions = {
  page: 'about' | 'terms'
  theme: 'light' | 'dark'
  viewport: 'desktop' | 'mobile'
}
type StaticApp = Awaited<ReturnType<typeof startInsightsFixtureApp>>
type MountStaticPage = (options: StaticPageOptions) => Promise<void>

export const test = base.extend<{ mountStaticPage: MountStaticPage }, { staticApp: StaticApp }>({
  // eslint-disable-next-line no-empty-pattern -- Playwright requires fixture argument destructuring.
  staticApp: [async ({}, use) => {
    // Static content needs no business fixtures. Unexpected calls are recorded
    // by the shared server and must fail the per-scenario zero-request contract.
    const app = await startInsightsFixtureApp(() => ({ status: 500 }))
    try { await use(app) }
    finally { await app.close() }
  }, { scope: 'worker' }],
  baseURL: async ({ staticApp }, use) => { await use(staticApp.base) },
  mountStaticPage: async ({ page, context, staticApp }, use, testInfo) => {
    const errors = captureBrowserErrors(page)
    const externalRequests: string[] = []
    const failedRequests: string[] = []
    const upstreamStart = staticApp.requests.length
    page.on('requestfailed', request => failedRequests.push(`${request.url()}: ${request.failure()?.errorText}`))
    await context.route('**/*', async route => {
      const url = new URL(route.request().url())
      if (url.origin === staticApp.base || ['data:', 'blob:'].includes(url.protocol)) return route.continue()
      externalRequests.push(url.href)
      await route.abort('blockedbyclient')
    })
    const assertQuiet = () => {
      expect(errors, 'Static hydration and rendering must have no browser errors').toEqual([])
      expect(failedRequests, 'No failed resource request is expected').toEqual([])
      expect(externalRequests, 'Static rendering must not fetch external origins').toEqual([])
      expect(staticApp.requests.slice(upstreamStart), 'Static rendering must not call upstream APIs').toEqual([])
    }
    try {
      await use(async options => {
        await page.setViewportSize(options.viewport === 'desktop'
          ? { width: 1440, height: 900 } : { width: 390, height: 844 })
        await context.addInitScript(({ origin, theme, steamDiagnosticsKey, steamSample }) => {
          if (location.origin !== origin) return
          localStorage.setItem('theme', theme)
          // Both global plugins probe even on Static pages. Seed valid history
          // so their real TTL path stays quiet; modes remain Auto, JS stays live.
          const checkedAt = Date.now()
          localStorage.setItem('gf_asset_cdn_diagnostics', JSON.stringify({
            selected: 'primary', checkedAt, primaryMs: 10, mirrorMs: 20,
            primaryState: 'success', mirrorState: 'success',
          }))
          localStorage.setItem(steamDiagnosticsKey, JSON.stringify({
            version: 1, selected: 'china', checkedAt, sample: steamSample,
            china: { ms: 30, state: 'success' }, global: { ms: 60, state: 'success' },
          }))
        }, { origin: staticApp.base, theme: options.theme, steamDiagnosticsKey: STEAM_DIAGNOSTICS_KEY, steamSample: STEAM_PROBE_PATHS[0] })

        const response = await page.goto(`/${options.page}`, { waitUntil: 'load' })
        expect(response?.status()).toBe(200)
        const ssr = await response!.text()
        expect(ssr).toContain(`gf-static-page ${options.page === 'about' ? 'about-page' : 'legal-page'}`)
        expect(ssr).toContain('data-public-background')
        await page.waitForFunction(() => Boolean((document.querySelector('#__nuxt') as Element & { __vue_app__?: unknown })?.__vue_app__))
        // NavBar initializes the real Theme Store from the seed; never set html.dark here.
        await expect.poll(() => page.locator('html').evaluate(el => el.classList.contains('dark'))).toBe(options.theme === 'dark')
        const root = page.locator('.gf-static-page')
        await expect(root).toHaveCount(1)
        await expect(root).toBeVisible()
        await expect(page.locator('.gf-app-shell > [data-public-background]')).toHaveAttribute('data-pattern-status', 'default')
        await expect(page.locator('[data-public-background]')).toBeVisible()
        await expect(root).toHaveCSS('background-color', 'rgba(0, 0, 0, 0)')

        // Only unrelated fixed tools are isolated. Canvas, panels, content and
        // the real product runtime remain intact throughout viewport capture.
        await page.addStyleTag({ content: '.page-scroll-dock, .mobile-bottom-tabs-root { display: none !important; }' })
        await expect.poll(() => root.evaluate(el => Array.from(el.querySelectorAll('img'))
          .every(image => image.complete && image.naturalWidth > 0))).toBe(true)
        await root.evaluate(el => el.scrollIntoView({ block: 'start', behavior: 'instant' }))
        await page.evaluate(() => (document.activeElement as HTMLElement | null)?.blur())
        await page.mouse.move(1, 1)
        await page.evaluate(async () => {
          const finite = document.getAnimations().filter(animation => animation.effect?.getComputedTiming().iterations !== Infinity)
          await Promise.all(finite.map(animation => animation.finished.catch(() => {})))
          await document.fonts.ready
          await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())))
        })

        const geometry = await root.evaluate(el => {
          const rect = el.getBoundingClientRect(), style = getComputedStyle(el)
          return {
            top: rect.top, left: rect.left, right: rect.right, width: rect.width, height: rect.height,
            scrollWidth: el.scrollWidth, clientWidth: el.clientWidth,
            documentWidth: document.documentElement.scrollWidth, viewportWidth: innerWidth,
            background: style.backgroundColor,
            transition: { property: style.transitionProperty, duration: style.transitionDuration, timing: style.transitionTimingFunction },
          }
        })
        expect(Math.abs(geometry.top), 'Capture must start at the Static root').toBeLessThanOrEqual(1)
        expect(geometry.width).toBeGreaterThan(0)
        expect(geometry.height).toBeGreaterThan(0)
        expect(geometry.left).toBeGreaterThanOrEqual(-1)
        expect(geometry.right).toBeLessThanOrEqual(geometry.viewportWidth + 1)
        expect(geometry.scrollWidth).toBeLessThanOrEqual(geometry.clientWidth + 1)
        expect(geometry.documentWidth).toBeLessThanOrEqual(geometry.viewportWidth + 1)
        expect(geometry.background).toBe('rgba(0, 0, 0, 0)')
        expect(geometry.transition.duration.split(',').map(value => value.trim())).toContain('0.5s')
        expect(geometry.transition.property.split(',').map(value => value.trim())).toEqual(
          expect.arrayContaining(['color', 'background-color', 'border-color', 'outline-color', 'text-decoration-color', 'fill', 'stroke']))
        expect(geometry.transition.timing).not.toBe('')

        const panels = root.locator(options.page === 'about' ? '.about-panel' : '.legal-panel')
        if (options.page === 'about') {
          await expect(root).toHaveClass(/\babout-page\b/)
          expect(await panels.count()).toBeGreaterThanOrEqual(2)
          await expect(root.locator('.about-link').first()).toBeVisible()
          await expect(root.locator('.about-avatar')).toHaveCount(1)
        } else {
          await expect(root).toHaveClass(/\blegal-page\b/)
          await expect(root.locator('.gf-static-section-list')).toBeVisible()
          expect(await root.locator('.gf-static-section').count()).toBeGreaterThan(1)
        }
        for (const panel of await panels.all()) {
          await expect(panel).toBeVisible()
          const appearance = await panel.evaluate(el => {
            const style = getComputedStyle(el)
            return { shadow: style.boxShadow, blur: style.backdropFilter }
          })
          expect(appearance.shadow).not.toBe('none')
          expect(appearance.blur).not.toBe('none')
        }
        await expect(root.locator('.gf-static-page__top-veil')).not.toHaveCSS('background-image', 'none')
        await expect(page.locator('[data-public-background]')).toHaveAttribute('data-pattern-status', 'default')
        assertQuiet()
        await testInfo.attach('static-transition.json', {
          body: JSON.stringify({ ...options, ...geometry.transition }), contentType: 'application/json',
        })
      })
      // Include errors/traffic arriving while Playwright compares the screenshot.
      assertQuiet()
    } finally {
      await testInfo.attach('static-requests.json', {
        body: JSON.stringify({ external: externalRequests, failed: failedRequests, upstream: staticApp.requests.slice(upstreamStart).map(url => url.href) }),
        contentType: 'application/json',
      })
      if (testInfo.status !== testInfo.expectedStatus) {
        await testInfo.attach('static-app.log', { body: staticApp.logs(), contentType: 'text/plain' })
      }
      // Playwright disposes the test-scoped context/routes; no gates or tasks wait here.
    }
  },
})

export { expect }
