import type { Locator } from '@playwright/test'
import { test as base, expect } from '../../fixtures/hero-preferences'
import { ASSET_PROBE_KEY } from '../../../../app/utils/managedAssets'
import { STEAM_PROBE_PATHS } from '../../../../app/utils/steamAssetRouting'

type PreferencesVisualOptions = {
  theme: 'light' | 'dark'
  viewport: 'desktop' | 'mobile'
  page: 'home' | 'background' | 'routing'
}

// 2026-09-20 00:00:00 UTC. Only Routing's browser Date.now/diagnostics use
// this clock; server time, timers and functional scenarios remain real.
const VISUAL_FIXED_NOW = Date.UTC(2026, 8, 20)
type MountPreferencesVisual = (options: PreferencesVisualOptions) => Promise<Locator>

export const test = base.extend<{ mountPreferencesVisual: MountPreferencesVisual }>({
  mountPreferencesVisual: async ({ page, preferences }, use) => {
    const probes: string[] = []
    page.on('request', request => {
      const url = new URL(request.url())
      if (url.pathname === '/' + ASSET_PROBE_KEY || STEAM_PROBE_PATHS.includes(url.pathname)) probes.push(url.href)
    })
    let routing = false
    await use(async options => {
      routing = options.page === 'routing'
      await preferences.open({
        ...(options.viewport === 'desktop' ? { width: 1440, height: 900 } : { width: 390, height: 844 }),
        theme: options.theme,
        ...(routing ? { fixedNow: VISUAL_FIXED_NOW, seedSteamDiagnostics: true } : {}),
      })
      // NavBar must initialize the real Pinia theme from the pre-navigation seed.
      await expect.poll(() => page.locator('html').evaluate(el => el.classList.contains('dark'))).toBe(options.theme === 'dark')
      await preferences.openPreferences()
      const backdrop = page.locator('body > .gf-modal-backdrop')
      const modal = backdrop.locator('.gf-modal.gf-preferences-modal')
      await expect(backdrop).toHaveCount(1)
      await expect(modal).toBeVisible()
      expect(await backdrop.evaluate(el => el.closest('#__nuxt'))).toBeNull()

      const index = ['home', 'background', 'routing'].indexOf(options.page)
      const tab = modal.getByRole('tab').nth(index)
      const panels = modal.locator('.preferences-pages > .preferences-page')
      const panel = panels.nth(index)
      await tab.click()
      await expect(tab).toHaveAttribute('aria-selected', 'true')
      await expect.poll(() => panel.evaluate(el => (el as HTMLElement).inert)).toBe(false)
      await expect.poll(() => modal.locator('.preferences-pages').evaluate((el, index) =>
        Math.abs(el.scrollLeft - index * el.clientWidth), index)).toBeLessThanOrEqual(1)

      // Reach only the default real states, not Catalog/Local/Pattern inventories.
      if (options.page === 'home') {
        await expect(panel.locator('#quick-access-toggle')).toHaveAttribute('aria-pressed', 'true')
        await expect(panel.getByRole('button', { name: '云端随机', exact: true })).toHaveAttribute('aria-pressed', 'true')
      } else if (options.page === 'background') {
        await expect(panel.getByRole('button', { name: '默认图案', exact: true })).toHaveAttribute('aria-pressed', 'true')
        await expect(panel.locator('.background-preferences__preview')).toBeVisible()
      } else {
        await expect(panel.locator('.resource-route__measurements')).toHaveCount(2)
        await expect(panel.locator('.resource-route__measurements').first()).toContainText('10 ms')
        await expect(panel.locator('.resource-route__measurements').first()).toContainText('20 ms')
        await expect(panel.locator('.resource-route__measurements').last()).toContainText('30 ms')
        await expect(panel.locator('.resource-route__measurements').last()).toContainText('60 ms')
        await expect(panel.locator('details[open]')).toHaveCount(0)
        expect(await page.evaluate(() => Date.now())).toBe(VISUAL_FIXED_NOW)
        expect(probes, 'Fresh Managed/Steam diagnostics must prevent automatic probes').toEqual([])
      }

      // Isolate unrelated homepage pixels, never the real Teleport/backdrop or
      // Preferences appearance. All child/scoped styles stay production-owned.
      await page.addStyleTag({ content: `
        #__nuxt { visibility: hidden !important; }
        body { background: var(--gf-page-background); }
      ` })
      await page.evaluate(() => (document.activeElement as HTMLElement | null)?.blur())
      await page.mouse.move(1, 1)
      await modal.evaluate(async el => {
        const finite = el.getAnimations({ subtree: true }).filter(animation => animation.effect?.getComputedTiming().iterations !== Infinity)
        await Promise.all(finite.map(animation => animation.finished.catch(() => {})))
        await document.fonts.ready
        await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())))
      })

      const geometry = await backdrop.evaluate(el => {
        const rect = el.getBoundingClientRect(), style = getComputedStyle(el)
        const modal = el.querySelector('.gf-preferences-modal')!.getBoundingClientRect()
        return {
          x: rect.x, y: rect.y, width: rect.width, height: rect.height,
          viewportWidth: innerWidth, viewportHeight: innerHeight,
          filter: style.backdropFilter, background: style.backgroundColor,
          modal: { left: modal.left, right: modal.right, top: modal.top, bottom: modal.bottom },
        }
      })
      expect(Math.abs(geometry.x)).toBeLessThanOrEqual(1)
      expect(Math.abs(geometry.y)).toBeLessThanOrEqual(1)
      expect(Math.abs(geometry.width - geometry.viewportWidth)).toBeLessThanOrEqual(1)
      expect(Math.abs(geometry.height - geometry.viewportHeight)).toBeLessThanOrEqual(1)
      expect(geometry.filter).not.toBe('none')
      expect(geometry.filter).not.toBe('')
      expect(geometry.background).not.toBe('rgba(0, 0, 0, 0)')
      expect(geometry.modal.left).toBeGreaterThanOrEqual(-1)
      expect(geometry.modal.top).toBeGreaterThanOrEqual(-1)
      expect(geometry.modal.right).toBeLessThanOrEqual(geometry.viewportWidth + 1)
      expect(geometry.modal.bottom).toBeLessThanOrEqual(geometry.viewportHeight + 1)
      // Offscreen Routing legends are clipped .sr-only absolute boxes whose
      // containing block is the Modal (backdrop-filter). Chromium includes them
      // in Modal.scrollWidth, even though every page fits. Check the actual
      // visible frame/header/tabs and active page, not that misleading aggregate.
      for (const target of [modal.locator('.gf-modal__header'), modal.locator('.preferences-tabs'), panel]) {
        const widths = await target.evaluate(el => ({ name: el.className, scroll: el.scrollWidth, client: el.clientWidth }))
        expect(widths.scroll, `${widths.name} must not overflow horizontally`).toBeLessThanOrEqual(widths.client + 1)
      }
      // The carousel intentionally has three pages of scrollable width; all
      // immediate frame boxes still have to fit inside the Modal.
      for (const child of await modal.locator(':scope > *').all()) {
        const box = await child.boundingBox()
        expect(box).not.toBeNull()
        expect(box!.x).toBeGreaterThanOrEqual(geometry.modal.left - 1)
        expect(box!.x + box!.width).toBeLessThanOrEqual(geometry.modal.right + 1)
      }
      expect(preferences.errors).toEqual([])
      if (routing) expect(probes).toEqual([])
      return backdrop
    })
    // Also cover traffic/errors that could arrive during screenshot comparison.
    if (routing) expect(probes).toEqual([])
    expect(preferences.errors).toEqual([])
  },
})

export { expect }
