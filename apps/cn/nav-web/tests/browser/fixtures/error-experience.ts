import { test as base, expect, type Locator } from '@playwright/test'
import { startInsightsFixtureApp } from '../../../scripts/fixtures/insights-app.mjs'
import { STEAM_DIAGNOSTICS_KEY, STEAM_PROBE_PATHS } from '../../../app/utils/steamAssetRouting'
import { captureBrowserErrors } from './browser-errors'

const missingPath = '/__gofurry_error_contract_missing__'
type ErrorTheme = 'light' | 'dark'
type ErrorApp = Awaited<ReturnType<typeof startInsightsFixtureApp>>
type ErrorScenario = {
  root: Locator
  actions: Locator
  homeButton: Locator
  backButton: Locator
  activeArtwork: Locator
  open404(options?: { theme?: ErrorTheme; width?: number; height?: number }): Promise<{ ssrHTML: string }>
  assertQuiet(): void
}

export const test = base.extend<{ errorExperience: ErrorScenario }, { errorApp: ErrorApp }>({
  // eslint-disable-next-line no-empty-pattern -- Playwright requires fixture argument destructuring.
  errorApp: [async ({}, use) => {
    // The real error boundary needs no business data. Unexpected calls fail.
    const app = await startInsightsFixtureApp(() => ({ status: 500 }))
    try { await use(app) }
    finally { await app.close() }
  }, { scope: 'worker' }],
  baseURL: async ({ errorApp }, use) => { await use(errorApp.base) },
  errorExperience: async ({ page, context, javaScriptEnabled, errorApp }, use, testInfo) => {
    const errors = captureBrowserErrors(page)
    const external: string[] = [], failed: string[] = [], badResponses: string[] = []
    const documentDiagnostics: string[] = [], disabledScripts: string[] = []
    let ssrHTML = ''
    const upstreamStart = errorApp.requests.length
    const documentURL = errorApp.base + missingPath
    // Keep the generic collector intact. Only Chromium's exact diagnostic for
    // this deliberately missing main document is accounted for separately.
    page.on('console', message => {
      if (message.type() === 'error' && message.location().url === documentURL
        && message.text() === `Failed to load resource: the server responded with a status of 404 (Page not found: ${missingPath})`) {
        documentDiagnostics.push(`error: ${message.text()}`)
      }
    })
    page.on('requestfailed', request => {
      const url = new URL(request.url())
      if (!javaScriptEnabled && request.resourceType() === 'script' && request.failure()?.errorText === 'csp'
        && url.origin === errorApp.base && url.pathname.startsWith('/_nuxt/') && url.pathname.endsWith('.js')) {
        disabledScripts.push(url.pathname)
        return
      }
      failed.push(`${request.url()}: ${request.failure()?.errorText}`)
    })
    page.on('response', response => {
      const request = response.request()
      if (response.status() < 400) return
      if (response.url() === documentURL && response.status() === 404
        && request.isNavigationRequest() && request.frame() === page.mainFrame()) return
      badResponses.push(`${response.status()} ${response.url()}`)
    })
    await context.route('**/*', async route => {
      const url = new URL(route.request().url())
      if (url.origin === errorApp.base || ['data:', 'blob:'].includes(url.protocol)) return route.continue()
      external.push(url.href)
      await route.abort('blockedbyclient')
    })

    const root = page.locator('.error-page')
    const actions = root.locator('.error-page__actions')
    const homeButton = root.getByRole('button', { name: '回到首页', exact: true })
    const backButton = root.getByRole('button', { name: '原路返回', exact: true })
    const activeArtwork = root.locator('.error-page__artwork-layer.is-active')
    const assertQuiet = () => {
      expect(errorApp.requests.slice(upstreamStart), 'Error rendering must not call business APIs').toEqual([])
      expect(external, 'Error rendering must use only bundled local assets').toEqual([])
      expect(failed, 'No failed resource requests are expected').toEqual([])
      expect(badResponses, 'Only the actual missing main document may return 404').toEqual([])
      // No-JS blocks SSR module preloads by design. Verify each was actually
      // advertised by this document, never excuse a failed image/CSS/request.
      for (const script of disabledScripts) expect(ssrHTML).toContain(`href="${script}"`)
      expect(documentDiagnostics, 'Exactly the expected main-document 404 diagnostic').toHaveLength(1)
      expect(errors, 'No application, hydration or other browser errors beyond the expected HTTP 404').toEqual(documentDiagnostics)
    }
    try {
      await use({
        root, actions, homeButton, backButton, activeArtwork, assertQuiet,
        async open404({ theme = 'light', width = 1440, height = 900 } = {}) {
          await page.setViewportSize({ width, height })
          if (javaScriptEnabled) {
            await context.addInitScript(({ origin, theme, steamDiagnosticsKey, steamSample }) => {
              if (location.origin !== origin) return
              localStorage.setItem('theme', theme)
              // Keep real plugins active; only their existing TTL history is seeded.
              const checkedAt = Date.now()
              localStorage.setItem('gf_asset_cdn_diagnostics', JSON.stringify({
                selected: 'primary', checkedAt, primaryMs: 10, mirrorMs: 20,
                primaryState: 'success', mirrorState: 'success',
              }))
              localStorage.setItem(steamDiagnosticsKey, JSON.stringify({
                version: 1, selected: 'china', checkedAt, sample: steamSample,
                china: { ms: 30, state: 'success' }, global: { ms: 60, state: 'success' },
              }))
            }, { origin: errorApp.base, theme, steamDiagnosticsKey: STEAM_DIAGNOSTICS_KEY, steamSample: STEAM_PROBE_PATHS[0] })
          } else {
            expect(theme, 'The real No-JS fallback only supports default Light').toBe('light')
          }
          const response = await page.goto(missingPath, { waitUntil: 'load' })
          expect(response?.status()).toBe(404)
          ssrHTML = await response!.text()
          for (const evidence of ['404', '前方只有一片未知', 'error-page']) expect(ssrHTML).toContain(evidence)
          await expect(root).toHaveCount(1)
          await expect(root).toBeVisible()
          await expect(root.locator('.error-page__code')).toHaveText('404')
          await expect(root.getByRole('heading', { name: '前方只有一片未知', exact: true })).toBeVisible()
          await expect(root.getByText('回头吧，旅行者', { exact: true })).toBeVisible()
          await expect(root.getByText('还有更多美好等待你探索', { exact: true })).toBeVisible()
          await expect(homeButton).toBeVisible()
          await expect(backButton).toBeVisible()
          await expect(page.locator('meta[name="robots"]')).toHaveAttribute('content', 'noindex, nofollow')
          if (javaScriptEnabled) {
            await page.waitForFunction(() => Boolean((document.querySelector('#__nuxt') as Element & { __vue_app__?: unknown })?.__vue_app__))
            // NavBar calls the real Theme Store; the fixture never writes html.dark.
            await expect.poll(() => page.locator('html').evaluate(el => el.classList.contains('dark'))).toBe(theme === 'dark')
            await expect(root).toHaveClass(/\bis-enhanced\b/)
            await expect(root.locator('.error-page__artwork-layer.is-active.is-ready')).toHaveCount(1)
          } else {
            await expect(page.locator('html')).not.toHaveClass(/\bdark\b/)
            await expect(root).not.toHaveClass(/\bis-enhanced\b/)
            expect(await page.evaluate(() => Boolean((document.querySelector('#__nuxt') as Element & { __vue_app__?: unknown })?.__vue_app__))).toBe(false)
          }
          await expect(activeArtwork).toHaveCount(1)
          await expect(activeArtwork.locator('img')).toHaveAttribute('src', `/web/404/illustration_001_16_9_${theme}.avif`)
          await expect.poll(() => activeArtwork.locator('img').evaluate((image: HTMLImageElement) => image.complete && image.naturalWidth > 0)).toBe(true)
          if (javaScriptEnabled) await expect(root).toHaveClass(/\bhas-entered\b/)
          assertQuiet()
          return { ssrHTML }
        },
      })
      assertQuiet()
    } finally {
      await testInfo.attach('error-experience-requests.json', {
        body: JSON.stringify({ upstream: errorApp.requests.slice(upstreamStart).map(url => url.href), external, failed, badResponses, errors, documentDiagnostics, disabledScripts }),
        contentType: 'application/json',
      })
      if (testInfo.status !== testInfo.expectedStatus) {
        await testInfo.attach('error-experience-app.log', { body: errorApp.logs(), contentType: 'text/plain' })
      }
      // No gates or outstanding tasks; Playwright disposes context/routes.
    }
  },
})

export { expect }
