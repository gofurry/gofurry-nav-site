import { test, expect, heroOrigins, lifecycleKey, expectSSRHeroRetained } from '../fixtures/hero-lifecycle'

for (const width of [1440, 390]) {
  test(`${width}px Hero SSR hydrates in place and loads only its viewport`, async ({ page, hero }) => {
    const { ssrHTML } = await hero.open({ width })
    const key = lifecycleKey(width >= 768 ? 'desktop' : 'mobile')
    expect(ssrHTML).toContain(key)
    await expectSSRHeroRetained(page)
    await expect.poll(() => page.locator('.nav-header__background--managed img').evaluate(el => (el as HTMLImageElement).complete && (el as HTMLImageElement).naturalWidth > 0)).toBe(true)
    expect(await hero.currentHero()).toBe(`${heroOrigins.primary}/${key}`)
    expect(hero.heroRequests).toEqual([`${heroOrigins.primary}/${key}`])
    expect(hero.homeCalls).toHaveLength(1)
    expect(hero.heroCalls).toHaveLength(0)
    expect(hero.errors).toEqual([])
  })
}
