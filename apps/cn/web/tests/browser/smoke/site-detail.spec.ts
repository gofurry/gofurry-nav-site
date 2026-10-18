import { test, expect, openRuntime, settleRuntime } from '../fixtures/site-detail'

test('Site Detail serves and hydrates the default Performance workspace', async ({ page, runtime }) => {
  const html = await openRuntime(page, '/site/41') // Shared owner asserts SSR 200 and hydration.
  expect(html).toContain('data-site-hero')
  expect(html).toContain('data-site-observation-view="performance"')
  await expect(page.locator('[data-site-hero]')).toBeVisible()
  await expect(page.locator('[data-site-target-context]:visible')).toBeVisible()
  await expect(page.locator('[data-site-observation-view]')).toHaveAttribute('data-site-observation-view', 'performance')
  await expect(page.locator('[data-site-performance]')).toBeVisible()
  await expect(page.locator('[data-site-performance-chart]')).toHaveAttribute('data-site-chart-ready', 'true')
  await settleRuntime(page)
  runtime.assertQuiet()
})
