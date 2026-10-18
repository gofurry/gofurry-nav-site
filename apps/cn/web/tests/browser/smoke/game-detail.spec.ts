import { test, expect, openGame, selectTab, expectMissingFacts } from '../fixtures/game-detail'
import { captureBrowserErrors } from '../fixtures/browser-errors'

for (const path of ['/games', '/en/games', '/games/82', '/en/games/82']) {
  test(`SSR metadata and content: ${path}`, async ({ request }) => {
    const response = await request.get(path)
    expect(response.status()).toBe(200)
    const html = await response.text()
    expect(html).toMatch(/<title>[^<]+<\/title>/)
    expect(html).toMatch(/<meta[^>]+name="description"[^>]+content="[^"]+"/)
    const canonicals = [...html.matchAll(/<link[^>]+rel="canonical"[^>]+href="([^"]+)"/g)]
    expect(canonicals).toHaveLength(1)
    expect(new URL(canonicals[0]![1]!).pathname).toBe(path)
    expect(html).toMatch(/hreflang="en-US"/)
    expect(html).not.toMatch(/<meta[^>]+name="robots"[^>]+content="[^"]*noindex/)
    const content = html.replace(/<script\b[^>]*>[\s\S]*?<\/script>/gi, '')
    if (path.endsWith('/82')) {
      expect(content).toMatch(/<h1[^>]*>\s*SSR Fixture Game\s*<\/h1>/)
      expect(content).toContain('<p>SSR fixture introduction</p>')
    } else {
      expect(content).toMatch(/>\s*Active game fixture\s*</)
    }
  })
}

test('hydrates and switches core tabs without turning missing Facts into zero', async ({ page }) => {
  const errors = captureBrowserErrors(page)
  await openGame(page)
  for (const tab of ['intro', 'insights', 'gallery', 'detail', 'intro', 'insights']) {
    await selectTab(page, tab)
    if (tab === 'insights') await expectMissingFacts(page)
    else await expect(page.locator('[data-game-insights]')).not.toBeVisible()
  }
  expect(errors).toEqual([])
})
