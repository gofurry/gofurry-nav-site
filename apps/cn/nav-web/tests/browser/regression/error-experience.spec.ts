import { test, expect } from '../fixtures/error-experience'

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
