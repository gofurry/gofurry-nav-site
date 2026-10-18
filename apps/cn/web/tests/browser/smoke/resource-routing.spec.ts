import { test, expect, origins, heroKey } from '../fixtures/resource-routing'
import { captureBrowserErrors } from '../fixtures/browser-errors'

test('Managed SSR explicit pin takes precedence over recommendation', async ({ request }) => {
  const response = await request.get('/', { headers: { Cookie: 'gf_asset_cdn=primary; gf_asset_cdn_mode=mirror' } })
  expect(response.status()).toBe(200)
  expect(await response.text()).toContain(`${origins.mirror}/${heroKey}`)
})

test('Steam SSR explicit pin takes precedence over recommendation', async ({ request }) => {
  const response = await request.get('/games', { headers: { Cookie: 'gf_steam_asset_mode=global; gf_steam_asset_group=china' } })
  expect(response.status()).toBe(200)
  expect(await response.text()).toContain('https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/570/header.jpg')
})

test('Resource Routing opens after hydration with both route sections', async ({ page, routing }) => {
  const errors = captureBrowserErrors(page)
  await routing.open()
  await routing.openResourceRouting()
  await expect(page.locator('[data-resource-routing] .resource-route')).toHaveCount(2)
  await expect(page.locator('[data-resource-routing]')).toBeVisible()
  routing.releaseProbes()
  await routing.waitForDiagnostics('managed')
  await routing.waitForDiagnostics('steam')
  expect(errors).toEqual([])
})
