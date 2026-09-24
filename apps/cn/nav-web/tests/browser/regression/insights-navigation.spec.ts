import { test, expect, openRuntime, keyboardFocus } from '../fixtures/insights-domain'

test('Insights navigation retains its plain hierarchy and real Light/Dark public canvas', async ({ page, request, runtime }) => {
  const html = await openRuntime(page, '/insights/sites?metric=ipv6&range=30d&dimension=country')
  expect(html).toContain('insights-page'); expect(html).not.toContain('insights-hero')
  expect(html).toContain('data-public-background'); expect(html).toContain('data-pattern-status="default"')
  expect(html).toContain('生态观测')
  const pattern = await request.get('/web/background/gofurry-pattern.svg')
  expect(pattern.status()).toBe(200); expect(pattern.headers()['content-type']).toContain('image/svg+xml')
  const background = () => page.locator('.gf-public-background__pattern').evaluate(el => {
    const style = getComputedStyle(el)
    return { color: style.backgroundColor, image: style.maskImage, opacity: Number(style.opacity), repeat: style.maskRepeat, size: style.maskSize }
  })
  expect(await background()).toMatchObject({ color: 'rgb(168, 135, 115)', opacity: .065, repeat: 'repeat', size: '160px 160px' })
  expect((await background()).image).toContain('gofurry-pattern.svg')
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
  await expect.poll(background).toMatchObject({ color: 'rgb(215, 193, 175)', opacity: .045, repeat: 'repeat', size: '160px 160px' })
  expect((await background()).image).toContain('gofurry-pattern.svg')
  runtime.assertQuiet()
})

test('Insights mobile nested route keeps Ecosystem active and keyboard links visible', async ({ page, runtime }) => {
  await page.setViewportSize({ width: 390, height: 844 })
  const html = await openRuntime(page, '/en/insights/games?metric=free&range=90d')
  expect(html).toContain('Ecosystem')
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
