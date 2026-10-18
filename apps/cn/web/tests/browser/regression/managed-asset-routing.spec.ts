import { test, expect, expectLoaded, readManagedResources, changeRenderedResourceProp, origins, iconKey, nextIconKey } from '../fixtures/resource-routing'
import { captureBrowserErrors } from '../fixtures/browser-errors'

test('loaded icon/Hero/Pattern survive automatic/manual/Save; new keys use the pin and real failures fall back', async ({ page, routing }) => {
  const errors = captureBrowserErrors(page, routing.expectedNetworkFailures)
  await routing.open()
  const icon = page.locator('.nav-site-card__logo img').first()
  await icon.scrollIntoViewIfNeeded()
  await expectLoaded(icon)
  await expectLoaded(page.locator('.nav-header__background--managed img'))
  await expect(page.locator('[data-pattern-status=server]')).toHaveCount(1)
  const before = await readManagedResources(page)
  expect(Object.values(before).every(url => url.includes('primary.example'))).toBe(true)
  await icon.evaluate(el => {
    const mutations: string[] = []
    Object.assign(window, { assetSrcMutations: mutations })
    new MutationObserver(records => mutations.push(...records.map(() => (el as HTMLImageElement).src)))
      .observe(el, { attributes: true, attributeFilter: ['src'] })
  })

  routing.releaseProbes()
  await routing.waitForDiagnostics('managed', 'mirror')
  expect(await routing.cookieValue('gf_asset_cdn')).toBe('mirror')
  expect(await readManagedResources(page)).toEqual(before)

  await routing.openResourceRouting()
  const assets = page.locator('[data-resource-routing] .resource-route').nth(0)
  routing.state.primary = 10
  routing.state.mirror = 170
  await assets.getByRole('button', { name: '重新测速', exact: true }).click()
  await routing.waitForDiagnostics('managed', 'primary')
  expect(await readManagedResources(page)).toEqual(before)
  await assets.getByRole('radio', { name: 'Cloudflare' }).check()
  await routing.savePreferences()
  expect(await routing.cookieValue('gf_asset_cdn_mode')).toBe('mirror')
  expect(await readManagedResources(page)).toEqual(before)
  expect(await page.evaluate(() => (window as Window & { assetSrcMutations?: string[] }).assetSrcMutations)).toEqual([])

  await changeRenderedResourceProp(icon, 'objectKey', nextIconKey)
  await expect(icon).toHaveAttribute('src', `${origins.mirror}/${nextIconKey}`)
  await expectLoaded(icon)
  routing.state.blockedURL = `${origins.mirror}/${iconKey}`
  await changeRenderedResourceProp(icon, 'objectKey', iconKey)
  await expect(icon).toHaveAttribute('src', `${origins.primary}/${iconKey}`)
  await expectLoaded(icon)
  expect(routing.expectedNetworkFailures.has(`${origins.mirror}/${iconKey}`)).toBe(true)
  expect(await routing.cookieValue('gf_asset_cdn_mode')).toBe('mirror')
  expect(errors).toEqual([])
})
