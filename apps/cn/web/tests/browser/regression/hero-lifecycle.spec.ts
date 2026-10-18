import { test, expect, heroOrigins, lifecycleKey, expectSSRHeroRetained } from '../fixtures/hero-lifecycle'

for (const width of [1440, 390]) {
  test(`${width}px displayed Hero owns failure authority and only explicit refresh changes identity`, async ({ page, hero }) => {
    await hero.open({ width })
    const variant = width >= 768 ? 'desktop' : 'mobile'
    const initial = `${heroOrigins.primary}/${lifecycleKey(variant)}`
    expect(await hero.currentHero()).toBe(initial)
    await expectSSRHeroRetained(page)
    const before = await hero.paintSample()
    expect(await hero.auxiliaryImageCount()).toBe(0)
    await hero.failAuxiliaryImages()
    // Flush any queued fallback work before comparing same-run painted bytes.
    await page.waitForTimeout(250)
    expect(await hero.currentHero()).toBe(initial)
    expect(await hero.paintSample()).toEqual(before)
    expect(await hero.auxiliaryImageCount()).toBe(0)

    await hero.probeAndPinMirror()
    expect(await page.evaluate(() => JSON.parse(localStorage.getItem('gf_asset_cdn_diagnostics')!).selected)).toBe('mirror')
    expect(await hero.currentHero()).toBe(initial)
    await page.evaluate(() => { window.dispatchEvent(new Event('focus')); window.dispatchEvent(new Event('resize')) })
    // The real prewarm timer mounts content; no fake timer or refetch hook.
    await expect(page.locator('.nav-content-shell')).toHaveCount(1)
    await page.waitForTimeout(900)
    expect(hero.homeCalls).toHaveLength(1)
    expect(hero.heroCalls).toHaveLength(0)
    expect(await hero.currentHero()).toBe(initial)
    expect(hero.heroRequests).toEqual([initial])

    await hero.refreshHome()
    await expect.poll(() => hero.currentHero()).toBe(`${heroOrigins.mirror}/${lifecycleKey(variant, 'b')}`)
    expect(hero.homeCalls).toHaveLength(2)
    expect(hero.heroCalls).toHaveLength(0)
    expect(hero.errors).toEqual([])
  })
}

for (const terminal of [false, true]) {
  test(`renderer failure before hydration ${terminal ? 'exhausts both providers to terminal fallback' : 'falls back Primary to Mirror'}`, async ({ page, hero }) => {
    hero.blockedHeroOrigins.add(heroOrigins.primary)
    if (terminal) hero.blockedHeroOrigins.add(heroOrigins.mirror)
    await hero.open()
    await expectSSRHeroRetained(page)
    await expect.poll(() => page.locator('.nav-header__background--managed img').evaluate(el => (el as HTMLImageElement).complete && (el as HTMLImageElement).naturalWidth > 0)).toBe(true)
    const primary = `${heroOrigins.primary}/${lifecycleKey('desktop')}`
    const mirror = `${heroOrigins.mirror}/${lifecycleKey('desktop')}`
    expect(hero.failuresBeforeHydration).toEqual([primary])
    expect(hero.heroRequests).toEqual([primary, mirror])
    expect([...hero.expectedNetworkFailures]).toEqual(terminal ? [primary, mirror] : [primary])
    if (terminal) expect(await hero.currentHero()).toMatch(/^data:/)
    else expect(await hero.currentHero()).toBe(mirror)
    expect(hero.errors).toEqual([])
  })
}

test('empty Mobile pool never borrows Desktop Hero', async ({ hero }) => {
  hero.emptyMobile = true
  await hero.open({ width: 390 })
  expect(hero.heroRequests).toEqual([])
  expect(await hero.currentHero()).toMatch(/^data:/)
  expect(hero.homeCalls).toHaveLength(1)
  expect(hero.heroCalls).toHaveLength(0)
  expect(hero.errors).toEqual([])
})
