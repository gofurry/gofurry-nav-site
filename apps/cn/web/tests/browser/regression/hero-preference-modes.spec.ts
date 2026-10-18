import { test, expect, heroOrigins, expectHeroLoaded } from '../fixtures/hero-preferences'
import { expectSSRHeroRetained } from '../fixtures/hero-lifecycle'

for (const width of [1440, 390]) {
  test(`${width}px Fixed SSR resolves independent BigInt pins and ignores saved Local cache`, async ({ page, preferences }) => {
    const mobileId = '9007199254740994'
    const { ssrHTML } = await preferences.open({ width, mode: 'fixed', desktopId: '34', mobileId, seedLocalCache: true })
    const desktop = preferences.desktopCatalog[24]!, mobile = preferences.mobileCatalog[1]!
    expect(ssrHTML).toContain(desktop.object_key)
    expect(ssrHTML).toContain(mobile.object_key)
    expect(ssrHTML).not.toContain(preferences.desktopCatalog[0]!.object_key)
    const expected = `${heroOrigins.primary}/${width >= 768 ? desktop.object_key : mobile.object_key}`
    await expectHeroLoaded(page)
    expect(await preferences.currentHero()).toBe(expected)
    expect(preferences.requestedImages).toEqual([expected])
    await expectSSRHeroRetained(page)
    expect(preferences.homeCalls()).toHaveLength(1)
    expect(preferences.heroCalls()).toHaveLength(0)
    await preferences.reload()
    expect(await preferences.currentHero()).toBe(expected)
    await expectSSRHeroRetained(page)
    expect(await preferences.cookieValue('gf_hero_desktop_id')).toBe('34')
    expect(await preferences.cookieValue('gf_hero_mobile_id')).toBe(mobileId)
    expect(preferences.heroCalls()).toHaveLength(0)
    expect(preferences.errors).toEqual([])
  })
}

test('same persisted Hero ID resolves an updated object key on the next SSR', async ({ preferences }) => {
  const old = preferences.desktopCatalog[24]!.object_key
  await preferences.open({ mode: 'fixed', desktopId: '34' })
  expect(await preferences.currentHero()).toBe(`${heroOrigins.primary}/${old}`)
  const replacement = 'nav/hero/desktop/' + 'f'.repeat(32) + '.avif'
  preferences.replaceDesktopObjectKey('34', replacement)
  const { ssrHTML } = await preferences.reload()
  expect(ssrHTML).toContain(replacement)
  expect(ssrHTML).not.toContain(old)
  expect(await preferences.currentHero()).toBe(`${heroOrigins.primary}/${replacement}`)
  expect(await preferences.cookieValue('gf_hero_desktop_id')).toBe('34')
  expect(preferences.errors).toEqual([])
})

test('invalid Desktop pin falls back without losing the valid Mobile BigInt pin', async ({ page, preferences }) => {
  const { ssrHTML } = await preferences.open({ mode: 'fixed', desktopId: '999', mobileId: '9007199254740994' })
  expect(await preferences.currentHero()).toBe(`${heroOrigins.primary}/${preferences.desktopCatalog[0]!.object_key}`)
  expect(ssrHTML).toContain(preferences.mobileCatalog[1]!.object_key)
  await page.setViewportSize({ width: 390, height: 1000 })
  await expect.poll(() => preferences.currentHero()).toBe(`${heroOrigins.primary}/${preferences.mobileCatalog[1]!.object_key}`)
  expect(await preferences.cookieValue('gf_hero_mobile_id')).toBe('9007199254740994')
  expect(preferences.errors).toEqual([])
})
