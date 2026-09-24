import { test as base, expect, type Locator } from '@playwright/test'
import { startInsightsFixtureApp } from '../../../../scripts/fixtures/insights-app.mjs'

type FoundationOptions = {
  theme: 'light' | 'dark'
  viewport: 'desktop' | 'mobile'
}

// Layout belongs to this fixture; all control appearance belongs to production
// CSS. The canvas is the sole appearance exception and consumes product tokens.
const foundationCSS = `
  body {
    background: var(--gf-page-background);
    color: var(--gf-text-main);
  }
  [data-visual-foundation] {
    display: grid;
    width: calc(100vw - 48px);
    max-width: 1120px;
    margin: 24px auto;
    padding: 24px;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 20px;
  }
  [data-visual-section] {
    display: grid;
    min-width: 0;
    align-content: start;
    gap: 12px;
  }
  [data-visual-row] {
    display: flex;
    min-width: 0;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
  }
  [data-visual-stack] {
    display: grid;
    min-width: 0;
    gap: 8px;
  }
  [data-visual-card-grid] {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8px;
  }
  @media (max-width: 640px) {
    [data-visual-foundation] {
      width: calc(100vw - 24px);
      margin: 12px auto;
      padding: 16px;
      grid-template-columns: 1fr;
      gap: 16px;
    }
    [data-visual-card-grid] { grid-template-columns: 1fr; }
  }
`

const foundationHTML = `
  <div data-visual-foundation>
    <section data-visual-section="buttons">
      <div class="gf-modal__label">Buttons</div>
      <div data-visual-row>
        <button class="gf-button gf-button--primary">Primary</button>
        <button class="gf-button gf-button--surface">Surface</button>
        <button class="gf-button gf-button--ghost">Ghost</button>
        <button class="gf-button gf-button--danger">Danger</button>
        <button class="gf-button gf-button--surface" disabled>Disabled</button>
      </div>
    </section>
    <section data-visual-section="cards">
      <div class="gf-modal__label">Cards</div>
      <div data-visual-card-grid>
        <div class="gf-card gf-card--compact">Default</div>
        <div class="gf-card gf-card--strong gf-card--compact">Strong</div>
        <div class="gf-card gf-card--flat gf-card--compact">Flat</div>
        <div class="gf-card gf-card--interactive gf-card--compact">Interactive</div>
      </div>
    </section>
    <section data-visual-section="inputs">
      <div class="gf-modal__label">Inputs</div>
      <div data-visual-stack>
        <input class="gf-input" aria-label="Value" value="GoFurry">
        <input class="gf-input" aria-label="Placeholder" placeholder="Placeholder">
        <input class="gf-input" aria-label="Focus" data-visual-focus-input value="Focused input">
      </div>
    </section>
    <section data-visual-section="chips">
      <div class="gf-modal__label">Chips</div>
      <div data-visual-row>
        <span class="gf-chip">Default</span>
        <span class="gf-chip gf-chip--muted">Muted</span>
        <span class="gf-chip gf-chip--active">Active</span>
      </div>
    </section>
    <section data-visual-section="pagination">
      <div class="gf-modal__label">Pagination</div>
      <div class="gf-pagination" data-visual-row>
        <span class="gf-pagination__button">1</span>
        <span class="gf-pagination__button gf-pagination__button--active">2</span>
        <span class="gf-pagination__button">3</span>
        <span class="gf-pagination__total">128 records</span>
      </div>
    </section>
    <section data-visual-section="rating">
      <div class="gf-modal__label">Rating</div>
      <div class="gf-rating rating-star" aria-label="4.6 (128)">
        <div class="rating-star__icons" aria-hidden="true">
          ${[100, 100, 100, 100, 60].map(fill => `
            <span class="rating-star__item">
              <span class="rating-star__empty">★</span>
              <span class="rating-star__fill" style="width: ${fill}%">★</span>
            </span>
          `).join('')}
        </div>
        <span class="rating-star__score">4.6</span>
        <span class="rating-star__count">(128)</span>
      </div>
    </section>
    <section data-visual-section="toggle">
      <div class="gf-modal__label">Preferences Toggle</div>
      <div class="gf-preferences-modal" data-visual-toggle-scope>
        <div data-visual-row>
          <button class="preferences-toggle" aria-label="Off" aria-pressed="false"><span></span></button>
          <button class="preferences-toggle preferences-toggle--on" aria-label="On" aria-pressed="true"><span></span></button>
        </div>
      </div>
    </section>
    <section data-visual-section="modal">
      <div class="gf-modal__label">Generic Modal</div>
      <div class="gf-modal">
        <div class="gf-modal__header">
          <div class="gf-modal__eyebrow">Foundation</div>
          <div class="gf-modal__title">Generic modal</div>
          <p class="gf-modal__desc">Shared modal primitive.</p>
        </div>
        <div class="gf-modal__body">
          <section class="gf-modal__section">
            <div class="gf-modal__label">Section</div>
            <p class="gf-modal__help">Generic modal content.</p>
          </section>
          <div class="gf-modal__notice">
            <div class="gf-modal__notice-title">Notice</div>
            Shared semantic notice.
          </div>
        </div>
        <div class="gf-modal__footer">
          <button class="gf-button gf-button--ghost">Cancel</button>
          <button class="gf-button gf-button--primary">Confirm</button>
        </div>
      </div>
    </section>
  </div>
`

type FoundationApp = Awaited<ReturnType<typeof startInsightsFixtureApp>>
type MountFoundation = (options: FoundationOptions) => Promise<Locator>

export const test = base.extend<{ mountFoundation: MountFoundation }, { foundationApp: FoundationApp }>({
  // eslint-disable-next-line no-empty-pattern -- Playwright requires fixture argument destructuring.
  foundationApp: [async ({}, use) => {
    const app = await startInsightsFixtureApp(() => ({ data: [] }))
    try { await use(app) }
    finally { await app.close() }
  }, { scope: 'worker' }],
  baseURL: async ({ foundationApp }, use) => { await use(foundationApp.base) },
  mountFoundation: async ({ page, foundationApp }, use) => {
    const externalRequests: string[] = []
    await page.route('**/*', async route => {
      const url = new URL(route.request().url())
      if (url.origin === foundationApp.base) {
        if (route.request().resourceType() === 'document') {
          const response = await route.fetch()
          const headers = response.headers()
          // Chromium with javaScriptEnabled:false also suppresses style load
          // callbacks and animation frames in Playwright 1.60. CSP excludes ALL
          // product scripts (including inline bootstraps), but permits test-side
          // evaluation/RAF. The real production document/head/CSS stay intact.
          headers['content-security-policy'] = [headers['content-security-policy'], "script-src 'none'"]
            .filter(Boolean).join(', ')
          return route.fulfill({ response, headers })
        }
        return route.continue()
      }
      externalRequests.push(url.href)
      await route.abort('blockedbyclient')
    })

    await use(async ({ theme, viewport }) => {
      await page.setViewportSize(viewport === 'desktop'
        ? { width: 1440, height: 900 }
        : { width: 390, height: 844 })
      const response = await page.goto('/about', { waitUntil: 'load' })
      expect(response?.status()).toBe(200)
      expect(response?.headers()['content-security-policy']).toContain("script-src 'none'")
      expect(await page.evaluate(() => Reflect.has(window, '__NUXT__')), 'Nuxt bootstrap must not execute').toBe(false)
      const head = await page.locator('head').innerHTML()
      expect(await page.locator('head link[rel="stylesheet"]').count()).toBeGreaterThan(0)
      const tokens = await page.evaluate(() => {
        const style = getComputedStyle(document.documentElement)
        return ['--gf-page-background', '--gf-surface', '--gf-accent-fill', '--gf-radius-sm']
          .map(name => [name, style.getPropertyValue(name).trim()])
      })
      for (const [name, value] of tokens) expect(value, `${name} must come from production CSS`).not.toBe('')

      await page.locator('body').evaluate((body, html) => { body.innerHTML = html }, foundationHTML)
      expect(await page.locator('head').innerHTML(), 'Body replacement must retain the production head').toBe(head)
      await page.addStyleTag({ content: foundationCSS })
      await page.evaluate(dark => document.documentElement.classList.toggle('dark', dark), theme === 'dark')

      const root = page.locator('[data-visual-foundation]')
      await expect(root.locator('[data-visual-section]')).toHaveCount(8)
      await expect(page.locator('.gf-preferences-modal')).toHaveCount(1)
      await expect(root.locator('[data-visual-section="toggle"] > [data-visual-toggle-scope].gf-preferences-modal')).toHaveCount(1)
      await expect(root.locator('[data-visual-toggle-scope] .preferences-toggle')).toHaveCount(2)
      await expect(root.locator('.gf-modal')).toHaveCount(1)
      expect(await root.evaluate(element => element.closest('.gf-preferences-modal'))).toBeNull()
      expect(await root.locator('.gf-modal').evaluate(element => element.closest('.gf-preferences-modal'))).toBeNull()
      await expect(root.locator('.gf-preferences-modal .gf-modal, .gf-preferences-modal .gf-input, .gf-preferences-modal .gf-button, .gf-preferences-modal .gf-chip, .gf-modal-backdrop')).toHaveCount(0)

      const focusedInput = page.locator('[data-visual-focus-input]')
      await focusedInput.focus()
      await page.mouse.move(1, 1)
      await page.evaluate(async () => {
        await document.fonts.ready
        await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())))
      })
      await expect(focusedInput).toBeFocused()
      const dimensions = await root.evaluate(element => ({ scroll: element.scrollWidth, client: element.clientWidth }))
      expect(dimensions.scroll, 'Foundation must not overflow horizontally').toBeLessThanOrEqual(dimensions.client + 1)
      expect(externalRequests, 'Foundation must not depend on remote assets or APIs').toEqual([])
      return root
    })
    // The test-scoped context owns route cleanup; there are no fixture gates/tasks.
  },
})

export { expect }
