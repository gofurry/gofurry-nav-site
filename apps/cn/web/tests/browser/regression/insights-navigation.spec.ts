import { test, expect, openRuntime, keyboardFocus } from '../fixtures/insights-domain'
import { assertNavigationClearance } from '../fixtures/navigation-clearance'
import { assertNativeReveal, observeThemeReveal } from '../fixtures/theme-reveal'

test('150 Insights theme snapshots retain real navigation and data', async ({ page, runtime }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await openRuntime(page, '/insights/sites')
  const content = await page.locator('.insights-page').innerText()
  await observeThemeReveal(page)
  const snapshot = await assertNativeReveal(page, page.getByRole('button', { name: '切换明暗主题图标', exact: true }), 'dark')
  expect(snapshot.shellColor).toBe('rgb(29, 31, 34)')
  await expect(page.locator('.insights-page')).toHaveText(content, { useInnerText: true })
  runtime.assertQuiet()
})

for (const route of ['/insights/sites', '/en/insights/games']) {
  test(`145-C1.1 Insights navigation clearance ${route}`, async ({ page, runtime }) => {
    await openRuntime(page, route)
    for (const width of [390, 768, 900, 1024, 1440]) {
      await page.setViewportSize({ width, height: 900 })
      await assertNavigationClearance(page, '.insights-primary-nav a, .insights-domain-nav a', `insights-${route.includes('games') ? 'games' : 'sites'}-${width}`)
    }
    runtime.assertQuiet()
  })
}

test('Insights navigation retains its plain hierarchy and real Light/Dark public canvas', async ({ page, runtime }) => {
  const html = await openRuntime(page, '/insights/sites?metric=ipv6&range=30d&dimension=country')
  expect(html).toContain('insights-page'); expect(html).not.toContain('insights-hero')
  expect(html).toContain('data-public-background'); expect(html).toContain('data-pattern-status="default"')
  expect(html).toContain('生态观测')
  await expect(page.locator('.gf-nav__links [aria-current="page"]')).toHaveAttribute('href', '/insights')
  expect(await page.locator('.ecosystem-navigation').evaluate(el => el.getBoundingClientRect().top)).toBeGreaterThanOrEqual(88)
  const background = () => page.locator('.gf-public-background__dots--left').evaluate(el => {
    const style = getComputedStyle(el)
    return { mask: style.maskImage, image: style.backgroundImage, opacity: Number(style.opacity), repeat: style.backgroundRepeat, size: style.backgroundSize }
  })
  expect(await background()).toMatchObject({ mask: expect.stringContaining('linear-gradient'), opacity: 1, repeat: 'repeat', size: '26px 26px' })
  expect((await background()).image).toContain('radial-gradient')
  await expect(page.locator('[data-public-background]')).toHaveCSS('background-color', 'rgb(242, 227, 208)')
  const nav = await page.evaluate(() => {
    const primary = document.querySelector('.insights-primary-nav')!, domain = document.querySelector('.insights-domain-nav[data-domain="site"]')!
    const p = primary.getBoundingClientRect(), d = domain.getBoundingClientRect()
    const ps = getComputedStyle(primary.querySelector('a')!), ds = getComputedStyle(domain.querySelector('a')!)
    return { direction: getComputedStyle(document.querySelector('.ecosystem-navigation')!).flexDirection,
      aligned: Math.abs(p.left - d.left) <= 1, stacked: d.top >= p.bottom && d.top - p.bottom <= 12,
      plain: [primary, domain].every(el => { const css = getComputedStyle(el); return css.boxShadow === 'none' && css.backgroundColor === 'rgba(0, 0, 0, 0)' && css.backgroundImage === 'none' && css.backdropFilter === 'none' && css.borderRadius === '0px' }),
      lower: parseFloat(ds.fontSize) < parseFloat(ps.fontSize) && Number(ds.fontWeight) < Number(ps.fontWeight) }
  })
  expect(nav).toEqual({ direction: 'column', aligned: true, stacked: true, plain: true, lower: true })
  for (const selector of ['.insights-primary-nav', '.insights-domain-nav']) await expect(page.locator(selector + ' [aria-current="page"]')).toHaveAttribute('href', '/insights/sites')
  await page.getByRole('button', { name: '切换明暗主题图标', exact: true }).click()
  await expect(page.locator('html')).toHaveClass(/dark/)
  await expect.poll(background).toMatchObject({ mask: expect.stringContaining('linear-gradient'), opacity: 1, repeat: 'repeat', size: '26px 26px' })
  expect((await background()).image).toContain('radial-gradient')
  await expect(page.locator('[data-public-background]')).toHaveCSS('background-color', 'rgb(29, 31, 34)')
  runtime.assertQuiet()
})

test('Insights mobile nested route keeps Ecosystem active and keyboard links visible', async ({ page, runtime }) => {
  await page.setViewportSize({ width: 390, height: 844 })
  const html = await openRuntime(page, '/en/insights/games?metric=free&range=90d')
  expect(html).toContain('Ecosystem')
  expect(await page.locator('.insights-page').evaluate(el => el.getBoundingClientRect().top)).toBe(0)
  expect(await page.locator('.ecosystem-navigation').evaluate(el => el.getBoundingClientRect().top)).toBeGreaterThanOrEqual(76)
  await expect(page.locator('.ecosystem-navigation')).toHaveCSS('flex-direction', 'column')
  await expect(page.locator('.insights-container')).not.toContainText(/undefined|NaN|null%/)
  await page.evaluate(() => window.scrollTo({ top: 240, behavior: 'instant' }))
  await expect(page.locator('.mobile-bottom-tabs--visible')).toBeVisible()
  await expect(page.locator('.mobile-bottom-tabs__item')).toHaveCount(4)
  await expect(page.locator('.mobile-bottom-tabs__item--active')).toHaveAttribute('href', '/en/insights')
  for (const selector of ['.insights-primary-nav', '.insights-domain-nav']) {
    const links = page.locator(selector + ' a')
    await keyboardFocus(links.first())
    for (let index = 1; index < await links.count(); index++) await page.keyboard.press('Tab')
    const last = links.last()
    await expect(last).toBeFocused()
    expect(await last.evaluate(el => {
      const box = el.getBoundingClientRect(), nav = el.closest('nav')!.getBoundingClientRect()
      return el.matches(':focus-visible') && getComputedStyle(el).outlineStyle === 'solid' && box.left >= nav.left - 1 && box.right <= nav.right + 1
    })).toBe(true)
  }
  expect(await page.evaluate(() => Math.max(document.documentElement.scrollWidth, document.body.scrollWidth) - innerWidth)).toBeLessThanOrEqual(2)
  runtime.assertQuiet()
})


test('145-A Insights keyboard reveal keeps a partially visible link below the floating shell', async ({ page, runtime }) => {
  await openRuntime(page, '/insights/sites?metric=ipv6&range=30d&dimension=country')
  for (const selector of ['.insights-primary-nav a', '.insights-domain-nav a']) {
    const link = page.locator(selector).first()
    await link.evaluate(el => window.scrollTo({ top: el.getBoundingClientRect().top + scrollY - 24, behavior: 'instant' }))
    await link.focus()
    await expect(link).toBeFocused()
    expect(await link.evaluate(el => el.getBoundingClientRect().top)).toBeGreaterThanOrEqual(88)
  }
  runtime.assertQuiet()
})


test('145-A same-page metric query keeps a completed mobile handoff latched', async ({ page, runtime }) => {
  await page.setViewportSize({ width: 640, height: 844 })
  await openRuntime(page, '/insights/sites?metric=ipv6&range=30d&dimension=country')
  const tabs = page.locator('.mobile-bottom-tabs')
  await page.evaluate(() => window.scrollTo({ top: 180, behavior: 'instant' }))
  await expect(tabs).toHaveAttribute('aria-hidden', 'false')
  await page.evaluate(() => window.scrollTo({ top: 40, behavior: 'instant' }))
  await expect(tabs).toHaveAttribute('aria-hidden', 'false')
  await page.locator('[data-metric-key="http2"]').click()
  await expect(page).toHaveURL(/metric=http2/)
  await expect.poll(() => page.evaluate(() => scrollY)).toBe(40)
  await expect(tabs).toHaveAttribute('aria-hidden', 'false')
  runtime.assertQuiet()
})
