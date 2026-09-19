import { test, expect, expectLoaded, changeRenderedResourceProp, steamSource } from '../fixtures/resource-routing'
import { captureBrowserErrors } from '../fixtures/browser-errors'

test('loaded Steam image survives automatic/manual/Save; new src uses Global with host fallback', async ({ page, routing }) => {
  const errors = captureBrowserErrors(page, routing.expectedNetworkFailures)
  await routing.open('/games')
  const picture = page.locator('.game-group-page-live img').first()
  await expectLoaded(picture)
  const first = await picture.getAttribute('src')
  expect(first).toMatch(/^https:\/\/shared\.st\.dl\.eccdnx\.com\//)
  routing.releaseProbes()
  await routing.waitForDiagnostics('steam', 'global')
  await expect(picture).toHaveAttribute('src', first!)

  await routing.openResourceRouting()
  const steam = page.locator('[data-resource-routing] .resource-route').nth(1)
  routing.state.china = 5
  routing.state.global = 160
  await steam.getByRole('button', { name: '重新测速', exact: true }).click()
  await routing.waitForDiagnostics('steam', 'china')
  await expect(picture).toHaveAttribute('src', first!)
  await steam.getByRole('radio', { name: '全球线路' }).check()
  await routing.savePreferences()
  expect(await routing.cookieValue('gf_steam_asset_mode')).toBe('global')
  await expect(picture).toHaveAttribute('src', first!)

  await changeRenderedResourceProp(picture, 'src', steamSource + '?changed=1')
  await expect(picture).toHaveAttribute('src', steamSource.replace('shared.steamstatic.com', 'shared.akamai.steamstatic.com') + '?changed=1')
  await expectLoaded(picture)
  routing.state.blockedURL = 'shared.akamai.steamstatic.com'
  await changeRenderedResourceProp(picture, 'src', steamSource + '?changed=2')
  await expect(picture).toHaveAttribute('src', steamSource.replace('shared.steamstatic.com', 'shared.cloudflare.steamstatic.com') + '?changed=2')
  await expectLoaded(picture)
  expect([...routing.expectedNetworkFailures].some(url => url.includes('shared.akamai.steamstatic.com') && url.includes('changed=2'))).toBe(true)
  expect(await routing.cookieValue('gf_steam_asset_mode')).toBe('global')
  expect(errors).toEqual([])
})

test('legacy Steam storage migrates only to recommendation and preserves the hydrated China snapshot', async ({ page, routing }) => {
  const errors = captureBrowserErrors(page)
  await routing.open('/games', { legacySteamRecommendation: { group: 'global', testedAt: Date.now() } })
  await expect.poll(() => routing.cookieValue('gf_steam_asset_group')).toBe('global')
  expect(await routing.cookieValue('gf_steam_asset_mode')).toBeUndefined()
  expect(await page.evaluate(() => localStorage.getItem('gofurry:steam-shared-cdn-preference:v1'))).toBeNull()
  const picture = page.locator('.game-group-page-live img').first()
  await expectLoaded(picture)
  const first = await picture.getAttribute('src')
  expect(first).toMatch(/^https:\/\/shared\.st\.dl\.eccdnx\.com\//)
  await routing.openResourceRouting()
  await expect(page.locator('[data-resource-routing] .resource-route').nth(1).getByRole('radio', { name: '自动优选' })).toBeChecked()
  routing.releaseProbes()
  await routing.waitForDiagnostics('managed')
  await routing.waitForDiagnostics('steam')
  await expect(picture).toHaveAttribute('src', first!)
  expect(await routing.cookieValue('gf_steam_asset_mode')).toBeUndefined()
  expect(errors).toEqual([])
})
