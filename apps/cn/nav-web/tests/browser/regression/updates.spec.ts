import { test, expect } from '../fixtures/updates'

test('Updates keeps grouped pagination and year toggles through hydration', async ({ page, updates }) => {
  const { ssrHTML } = await updates.open()
  expect(ssrHTML).toContain(updates.items[0]!.title)

  await page.mouse.move(1, 1)
  await page.keyboard.press('Tab')
  await updates.loadMore.focus()
  await expect(updates.loadMore).toBeFocused()
  expect(await updates.loadMore.evaluate(el => el.matches(':focus-visible'))).toBe(true)
  await expect(updates.loadMore).not.toHaveCSS('transform', 'none')
  await updates.loadMore.click()
  await expect(updates.entries).toHaveCount(7)
  await expect(updates.loadMore).toHaveCount(0)

  await updates.yearControl('2026').click()
  await expect(updates.entries).toHaveCount(0)
  await updates.yearControl('2026').click()
  await expect(updates.yearEntries('2026')).toHaveCount(7)
  await expect(updates.entries).toHaveCount(7)
  await expect(updates.loadMore).toHaveCount(0)
  await updates.yearControl('2025').click()
  await expect(updates.yearEntries('2025')).toHaveCount(2)
  await expect(updates.entries).toHaveCount(9)
  await expect(updates.latest).toHaveCount(1)
  updates.assertQuiet()
})

for (const width of [1440, 390]) for (const theme of ['light', 'dark'] as const) {
  test(`English Updates runtime ${width} ${theme}`, async ({ page, updates }) => {
    const { ssrHTML } = await updates.open({ width, height: width === 390 ? 844 : 900, theme, locale: 'en' })
    expect(ssrHTML).toContain('GoFurry Updates')
    await expect(page.locator('[data-public-background]')).toHaveAttribute('data-pattern-status', 'default')
    await page.evaluate(async () => {
      await document.fonts.ready
      await new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve)))
    })
    expect(await page.evaluate(() => Math.max(document.body.scrollWidth, document.documentElement.scrollWidth) - document.documentElement.clientWidth)).toBeLessThanOrEqual(1)
    await updates.loadMore.click()
    await expect(updates.entries).toHaveCount(7)
    await expect(updates.loadMore).toHaveCount(0)
    updates.assertQuiet()
  })
}
