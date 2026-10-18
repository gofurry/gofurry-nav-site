import { test, expect, openRuntime, settleRuntime, assertRuntimeSurface, indexPath, updatesNow } from '../fixtures/updates'

// Eight representative compositions. State/security matrices remain Functional/Unit.
for (const detail of [false, true]) for (const device of ['desktop', 'mobile'] as const) for (const theme of ['light', 'dark'] as const) {
  test(`${detail ? 'Release article' : 'Release index'} ${theme} ${device}`, async ({ page, context, runtime }) => {
    await page.setViewportSize({ width: device === 'desktop' ? 1440 : 390, height: 900 })
    await page.clock.setFixedTime(new Date(updatesNow))
    await context.addInitScript(theme => localStorage.setItem('theme', theme), theme)
    await openRuntime(page, detail ? '/updates/108' : '/updates')
    const selector = detail ? '[data-update-detail]' : '[data-updates-index]'
    const root = page.locator(selector)
    await expect(page.locator('[data-public-background]')).toHaveAttribute('data-pattern-status', 'default')
    // Preserve the real shell/canvas; unrelated fixed controls own separate goldens.
    await page.addStyleTag({ content: '.page-scroll-dock, .mobile-bottom-tabs-root { display: none !important; }' })
    for (const image of await root.locator('img').all()) {
      await image.scrollIntoViewIfNeeded()
      await expect.poll(() => image.evaluate(el => (el as HTMLImageElement).complete && (el as HTMLImageElement).naturalWidth > 0)).toBe(true)
    }
    await page.evaluate(() => { (document.activeElement as HTMLElement | null)?.blur(); window.scrollTo({ top: 0, behavior: 'instant' }) })
    await page.mouse.move(0, 0)
    await settleRuntime(page)
    await page.evaluate(async () => {
      await document.fonts.ready
      await Promise.all(document.getAnimations().filter(animation => animation.effect?.getComputedTiming().iterations !== Infinity).map(animation => animation.finished.catch(() => {})))
      await new Promise<void>(resolve => requestAnimationFrame(() => requestAnimationFrame(() => resolve())))
    })
    await assertRuntimeSurface(page, selector, theme)
    expect(runtime.calls.map(call => call.url.pathname)).toEqual([indexPath + (detail ? '/108' : '')])
    runtime.assertQuiet()
    await expect(root).toHaveScreenshot(`${detail ? 'update-detail' : 'updates'}-${theme}-${device}.png`)
    runtime.assertQuiet()
  })
}
