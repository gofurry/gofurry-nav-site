import { test, expect } from '../fixtures/error-experience'
import { mkdir } from 'node:fs/promises'
import { assertNativeReveal, observeThemeReveal } from '../fixtures/theme-reveal'

test('150 error artwork is final in the native theme snapshot', async ({ page, errorExperience }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await errorExperience.open404()
  if (process.env.GOFURRY_FIXTURE_DEV === '1') await page.locator('nuxt-error-overlay').locator('#pip-close').click()
  await expect(errorExperience.root.locator('.error-page__artwork-layer.is-ready')).toHaveCount(2)
  await observeThemeReveal(page)
  const button = page.getByRole('button', { name: '切换明暗主题图标', exact: true })
  for (const target of ['dark', 'light'] as const) {
    const snapshot = await assertNativeReveal(page, button, target)
    expect(snapshot.artOpacity).toBe('1')
    await expect(errorExperience.activeArtwork.locator('img')).toHaveAttribute('src', `/web/404/illustration_001_16_9_${target}.avif`)
  }
  errorExperience.assertQuiet()
})

for (const width of [1440, 390]) for (const theme of ['light', 'dark'] as const) {
  test(`151 error surfaces and 500 variant ${width} ${theme}`, async ({ page, errorExperience }) => {
    await errorExperience.open404({ width, height: width === 390 ? 844 : 900, theme })
    if (process.env.GOFURRY_FIXTURE_DEV === '1') {
      // Nuxt's development-only error preview is expected on a real 404.
      // Dismiss it through its own UI for local artwork review.
      await page.locator('nuxt-error-overlay').locator('#pip-close').click()
    }
    const { root } = errorExperience
    const primary = root.locator('.gf-button--primary'), secondary = root.locator('.gf-button--surface')
    for (const status of [404, 500]) {
      if (status === 500) {
        // Exercise the real Nuxt error boundary's client state, not a fake error component.
        await page.evaluate(() => {
          const app = (document.querySelector('#__nuxt') as Element & {
            __vue_app__: { config: { globalProperties: { $nuxt: { payload: { error: unknown } } } } }
          }).__vue_app__.config.globalProperties.$nuxt
          app.payload.error = { statusCode: 500, statusMessage: 'Fixture server error' }
        })
      }
      await expect(root.locator('.error-page__code')).toHaveText(String(status))
      await expect(root.getByRole('heading')).toHaveText(status === 404 ? '前方只有一片未知' : '前方发生意外')
      await expect(primary).toHaveText(status === 404 ? '回到首页' : '再试一次')
      await expect(secondary).toHaveText(status === 404 ? '原路返回' : '回到首页')
      await expect(root).toHaveCSS('background-color', theme === 'dark' ? 'rgb(25, 27, 30)' : 'rgb(245, 233, 217)')
      await expect(root.locator('.error-page__code')).toHaveCSS('color', theme === 'dark' ? 'rgb(211, 161, 113)' : 'rgb(172, 88, 44)')
      await page.mouse.move(1, 1)
      await expect(primary).toHaveCSS('background-color', theme === 'dark' ? 'rgb(57, 59, 63)' : 'rgb(174, 86, 40)')
      await expect(secondary).toHaveCSS('background-color', 'rgba(0, 0, 0, 0)')
      await primary.hover()
      await expect(primary).toHaveCSS('background-color', theme === 'dark' ? 'rgb(72, 75, 80)' : 'rgb(180, 93, 44)')
      await page.mouse.move(1, 1)
      await primary.focus()
      await primary.press('Tab'); await expect(secondary).toBeFocused()
      await secondary.press('Shift+Tab'); await expect(primary).toBeFocused()
      await expect(primary).toHaveCSS('outline-style', 'solid')
      await expect(primary).toHaveCSS('outline-color', theme === 'dark' ? 'rgb(211, 161, 113)' : 'rgb(172, 88, 44)')
      await primary.press('Tab'); await expect(secondary).toBeFocused()
      expect(await page.evaluate(() => document.documentElement.scrollWidth - innerWidth)).toBeLessThanOrEqual(1)
      if (process.env.GOFURRY_FIXTURE_DEV !== '1') {
        await expect(page.locator('nuxt-error-overlay, vite-error-overlay, #nuxt-error-overlay')).toHaveCount(0)
      }
      if (process.env.GOFURRY_151_ERROR_REVIEW === '1') {
        await mkdir('../../../.cache/151-error-review', { recursive: true })
        await page.screenshot({ path: `../../../.cache/151-error-review/${status}-${width}-${theme}.png` })
      }
    }
    errorExperience.assertQuiet()
  })
}

test.describe('Real Nuxt 404', () => {
  test.use({ locale: 'zh-CN', contextOptions: { reducedMotion: 'no-preference' } })

  test('hydrated keyboard focus safety', async ({ page, errorExperience }) => {
    await errorExperience.open404()
    expect(await page.evaluate(() => matchMedia('(prefers-reduced-motion: reduce)').matches)).toBe(false)
    // Read staged state and focus in one browser turn, before the real delay expires.
    // No animation clock, CSS override or wait for the region to reveal itself.
    const focus = await errorExperience.homeButton.evaluate(button => {
      const actions = button.parentElement!
      const before = getComputedStyle(actions)
      const staged = { opacity: before.opacity, transform: before.transform, delay: before.animationDelay }
      button.focus()
      const after = getComputedStyle(actions), indication = getComputedStyle(button)
      return {
        staged, opacity: after.opacity, transform: after.transform,
        focused: document.activeElement === button, focusVisible: button.matches(':focus-visible'),
        outlineStyle: indication.outlineStyle, outlineWidth: parseFloat(indication.outlineWidth),
      }
    })
    expect(focus.staged.opacity, 'Actions must still be staged hidden before focus').toBe('0')
    expect(focus.staged.transform).not.toBe('none')
    expect(focus.staged.delay).toBe('2.55s')
    expect(focus.focused).toBe(true)
    expect(focus.opacity, 'Focus must reveal the region synchronously').toBe('1')
    expect(focus.transform).toBe('none')
    expect(focus.focusVisible).toBe(true)
    expect(focus.outlineStyle).not.toBe('none')
    expect(focus.outlineWidth).toBeGreaterThan(0)
    await expect(errorExperience.homeButton).toBeFocused()
    await expect(errorExperience.actions).toBeVisible()
    errorExperience.assertQuiet()
  })

  test.describe('No-JS', () => {
    test.use({ javaScriptEnabled: false })

    test('real No-JS fallback', async ({ page, errorExperience }) => {
      await errorExperience.open404()
      expect(await page.evaluate(() => matchMedia('(prefers-reduced-motion: reduce)').matches)).toBe(false)
      const steps = errorExperience.root.locator('.error-page__step')
      await expect(steps).toHaveCount(5)
      for (const step of await steps.all()) {
        await expect(step).toBeVisible()
        await expect(step).toHaveCSS('opacity', '1')
      }
      await expect(errorExperience.activeArtwork).not.toHaveClass(/\bis-ready\b/)
      await expect(errorExperience.activeArtwork).toHaveCSS('opacity', '1')
      errorExperience.assertQuiet()
      // The current noscript contract is readability, not button navigation.
    })
  })
})
