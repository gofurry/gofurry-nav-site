import { test as base, expect } from '@playwright/test'
import { startInsightsFixtureApp } from '../../../../scripts/fixtures/insights-app.mjs'
import { captureBrowserErrors } from '../../fixtures/browser-errors'
import { STEAM_DIAGNOSTICS_KEY, STEAM_PROBE_PATHS } from '../../../../app/utils/steamAssetRouting'

type FooterApp = Awaited<ReturnType<typeof startInsightsFixtureApp>>
type Clip = { x: number, y: number, width: number, height: number }
type MountFooterShell = (theme: 'light' | 'dark') => Promise<Clip>

export const test = base.extend<{ mountFooterShell: MountFooterShell }, { footerApp: FooterApp }>({
  // eslint-disable-next-line no-empty-pattern -- Playwright requires fixture argument destructuring.
  footerApp: [async ({}, use) => {
    // Terms/Footer need no business data. Every upstream call is unexpected.
    const app = await startInsightsFixtureApp(() => ({ status: 500 }))
    try { await use(app) }
    finally { await app.close() }
  }, { scope: 'worker' }],
  baseURL: async ({ footerApp }, use) => { await use(footerApp.base) },
  mountFooterShell: async ({ page, context, footerApp }, use, testInfo) => {
    const errors = captureBrowserErrors(page)
    const external: string[] = []
    const failed: string[] = []
    const upstreamStart = footerApp.requests.length
    page.on('requestfailed', request => failed.push(`${request.url()}: ${request.failure()?.errorText}`))
    await context.route('**/*', async route => {
      const url = new URL(route.request().url())
      if (url.origin === footerApp.base || ['data:', 'blob:'].includes(url.protocol)) return route.continue()
      external.push(url.href)
      await route.abort('blockedbyclient')
    })
    const assertQuiet = () => {
      expect(errors, 'Footer SSR/hydration must have no browser errors').toEqual([])
      expect(failed, 'No failed resource request is expected').toEqual([])
      expect(external, 'Footer must not fetch external origins').toEqual([])
      expect(footerApp.requests.slice(upstreamStart), 'Terms/Footer must not call business APIs').toEqual([])
    }
    try {
      await use(async theme => {
        await page.setViewportSize({ width: 1440, height: 900 })
        await context.addInitScript(({ origin, theme, steamKey, steamSample }) => {
          if (location.origin !== origin) return
          localStorage.setItem('theme', theme)
          // Real plugins use their normal TTL path; the clock/year is untouched.
          const checkedAt = Date.now()
          localStorage.setItem('gf_asset_cdn_diagnostics', JSON.stringify({
            selected: 'primary', checkedAt, primaryMs: 10, mirrorMs: 20,
            primaryState: 'success', mirrorState: 'success',
          }))
          localStorage.setItem(steamKey, JSON.stringify({
            version: 1, selected: 'china', checkedAt, sample: steamSample,
            china: { ms: 30, state: 'success' }, global: { ms: 60, state: 'success' },
          }))
        }, { origin: footerApp.base, theme, steamKey: STEAM_DIAGNOSTICS_KEY, steamSample: STEAM_PROBE_PATHS[0] })

        const response = await page.goto('/terms', { waitUntil: 'load' })
        expect(response?.status()).toBe(200)
        const ssr = await response!.text()
        expect(ssr).toContain('gf-static-page legal-page')
        expect(ssr).toContain('服务条款')
        expect(ssr).toContain('gf-footer-shell')
        expect(ssr).toContain('gf-footer__section-title')
        await page.waitForFunction(() => Boolean((document.querySelector('#__nuxt') as Element & { __vue_app__?: unknown })?.__vue_app__))
        // Only the production NavBar/Theme Store applies html.dark.
        await expect.poll(() => page.locator('html').evaluate(el => el.classList.contains('dark'))).toBe(theme === 'dark')
        await expect(page.locator('.gf-app-shell > [data-public-background]')).toHaveAttribute('data-pattern-status', 'default')
        await expect(page.locator('[data-public-background]')).toBeVisible()
        const footer = page.locator('.gf-footer')
        await expect(footer).toHaveCount(1)
        await expect(page.locator('.gf-footer-shell')).toHaveCount(1)
        await expect(footer).toBeVisible()

        // Isolate only unrelated fixed tools, never Footer/Shell appearance.
        await page.addStyleTag({ content: '.page-scroll-dock, .mobile-bottom-tabs-root { display: none !important; }' })
        await expect.poll(() => footer.evaluate(el => Array.from(el.querySelectorAll('img'))
          .every(image => image.complete && image.naturalWidth > 0))).toBe(true)
        await page.evaluate(() => document.fonts.ready)
        await footer.evaluate(el => el.scrollIntoView({ block: 'end', behavior: 'instant' }))
        await page.evaluate(() => (document.activeElement as HTMLElement | null)?.blur())
        await page.mouse.move(1, 1)

        const titles = footer.locator('.gf-footer__section-title')
        await expect(titles).toHaveCount(4)
        for (const title of await titles.all()) {
          await expect(title).toHaveCSS('font-size', '12px')
          await expect(title).toHaveCSS('font-weight', '600')
          await expect(title).toHaveCSS('text-transform', 'uppercase')
          await expect(title).toHaveCSS('transition-property', 'color')
          await expect(title).toHaveCSS('transition-duration', '0.5s')
        }
        const headingIcons = titles.locator('svg')
        await expect(headingIcons).toHaveCount(4)
        for (const icon of await headingIcons.all()) await expect(icon).toHaveCSS('opacity', '0.8')
        for (const name of ['网站地图', '开放平台', '反馈', '关于']) {
          await expect(footer.getByRole('heading', { name, exact: true })).toBeVisible()
        }
        for (const name of ['Sitemap', '在线状态', '关于本站', '服务条款']) {
          await expect(footer.getByRole('link', { name, exact: true })).toBeVisible()
        }

        const meta = footer.locator('.gf-footer__meta')
        await expect(meta).toHaveCount(1)
        await expect(meta).toHaveCSS('font-size', '14px')
        await expect(meta).toHaveCSS('transition-property', 'color')
        await expect(meta).toHaveCSS('transition-duration', '0.5s')
        await expect(meta).toHaveCSS('color', theme === 'light' ? 'rgb(148, 163, 184)' : 'rgb(100, 116, 139)')
        const metaLink = meta.locator('.gf-footer__meta-link')
        await expect(metaLink).toHaveCount(1)
        await expect(metaLink).toHaveCSS('color', theme === 'light' ? 'rgb(203, 213, 225)' : 'rgb(148, 163, 184)')
        await expect(metaLink).toHaveCSS('transition-property', 'color')
        await expect(metaLink).toHaveCSS('transition-duration', '0.16s')

        const socialIcons = footer.locator('.gf-footer__social-icon')
        await expect(socialIcons).toHaveCount(4)
        // Accessible identity, not DOM ordering or generated asset URLs, owns brand color.
        for (const [name, color] of Object.entries({
          哔哩哔哩: 'rgb(240, 128, 128)', 微博: 'rgb(255, 69, 0)',
          GitHub: 'rgb(56, 189, 248)', 'X（Twitter）': 'rgb(29, 155, 240)',
        })) {
          const icon = footer.getByRole('link', { name, exact: true }).locator('.gf-footer__social-icon')
          await page.mouse.move(1, 1)
          await expect(icon).toHaveCSS('filter', 'none')
          await icon.hover()
          await expect.poll(() => icon.evaluate(el => getComputedStyle(el).filter)).not.toBe('none')
          await expect.poll(() => icon.evaluate(el => getComputedStyle(el).filter)).toContain(color)
          await page.mouse.move(1, 1)
          await expect.poll(() => icon.evaluate(el => getComputedStyle(el).filter)).toBe('none')
        }
        await page.mouse.move(1, 1)
        for (const icon of await socialIcons.all()) await expect(icon).toHaveCSS('filter', 'none')

        const shell = await page.locator('.gf-app-shell').evaluate(el => {
          const style = getComputedStyle(el)
          return { property: style.transitionProperty, duration: style.transitionDuration, timing: style.transitionTimingFunction }
        })
        expect(shell.property.split(',').map(value => value.trim())).toEqual(expect.arrayContaining([
          'color', 'background-color', 'border-color', 'outline-color', 'text-decoration-color', 'fill', 'stroke',
          '--tw-gradient-from', '--tw-gradient-via', '--tw-gradient-to',
        ]))
        expect(shell.duration.split(',').map(value => value.trim()).every(value => value === '0.5s')).toBe(true)
        expect(shell.timing.split(/,(?![^()]*\))/).map(value => value.trim())
          .every(value => value === 'cubic-bezier(0.4, 0, 0.2, 1)')).toBe(true)

        await page.evaluate(async () => {
          const finite = document.getAnimations().filter(animation => animation.effect?.getComputedTiming().iterations !== Infinity)
          await Promise.all(finite.map(animation => animation.finished.catch(() => {})))
          await document.fonts.ready
          await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())))
        })
        const columns = footer.locator(':scope > div > div')
        await expect(columns).toHaveCount(3)
        await expect(columns.nth(2)).toHaveClass(/\bgf-footer__meta\b/)
        const geometry = await footer.evaluate(el => {
          const box = (node: Element) => {
            const r = node.getBoundingClientRect()
            return { left: r.left, top: r.top, right: r.right, bottom: r.bottom, width: r.width, height: r.height }
          }
          return {
            footer: box(el), columns: Array.from(el.querySelectorAll(':scope > div > div')).map(box),
            meta: box(el.querySelector('.gf-footer__meta')!),
            viewport: { width: innerWidth, height: innerHeight }, documentWidth: document.documentElement.scrollWidth,
          }
        })
        const { footer: bounds, columns: [first, second, third], viewport } = geometry
        expect(bounds.top, 'The complete Footer must fit; never switch to a full-page capture').toBeGreaterThanOrEqual(-1)
        expect(bounds.bottom).toBeLessThanOrEqual(viewport.height + 1)
        expect(bounds.width).toBeGreaterThan(0)
        expect(bounds.height).toBeGreaterThan(0)
        expect(geometry.documentWidth).toBeLessThanOrEqual(viewport.width + 1)
        for (const column of [first, second, third]) {
          expect(column.width).toBeGreaterThan(0)
          expect(column.height).toBeGreaterThan(0)
        }
        expect(second.right).toBeLessThanOrEqual(third.left)
        // Capture the canvas from the Footer edge through the inter-column gap.
        // No dynamic meta/year pixels and no hardcoded crop width.
        const right = Math.min(viewport.width, Math.floor((second.right + third.left) / 2))
        const x = Math.max(0, Math.floor(bounds.left))
        const y = Math.max(0, Math.floor(bounds.top))
        const clip = { x, y, width: right - x, height: Math.min(viewport.height, Math.ceil(bounds.bottom)) - y }
        expect(clip.width).toBeGreaterThan(0)
        expect(clip.height).toBeGreaterThan(0)
        for (const column of [first, second]) {
          expect(column.left).toBeGreaterThanOrEqual(clip.x - 1)
          expect(column.right).toBeLessThanOrEqual(clip.x + clip.width + 1)
          expect(column.top).toBeGreaterThanOrEqual(clip.y - 1)
          expect(column.bottom).toBeLessThanOrEqual(clip.y + clip.height + 1)
        }
        expect(geometry.meta.left, 'Dynamic current-year meta must stay outside the golden').toBeGreaterThanOrEqual(clip.x + clip.width - 1)
        await expect(page.locator('[data-public-background]')).toHaveAttribute('data-pattern-status', 'default')
        assertQuiet()
        await testInfo.attach('footer-shell-contract.json', {
          body: JSON.stringify({ theme, shell, geometry, clip }), contentType: 'application/json',
        })
        return clip
      })
      // Include late prefetch/errors during screenshot comparison.
      assertQuiet()
    } finally {
      await testInfo.attach('footer-shell-requests.json', {
        body: JSON.stringify({ external, failed, upstream: footerApp.requests.slice(upstreamStart).map(url => url.href), errors }),
        contentType: 'application/json',
      })
      if (testInfo.status !== testInfo.expectedStatus) {
        await testInfo.attach('footer-shell-app.log', { body: footerApp.logs(), contentType: 'text/plain' })
      }
      // Playwright disposes the test context/routes. No pending fixture gates exist.
    }
  },
})

export { expect }
