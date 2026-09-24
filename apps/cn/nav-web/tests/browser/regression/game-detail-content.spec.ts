import { test, expect } from '../fixtures/game-detail-contract'

test('populated detail is rendered by SSR with one hydration view side effect', async ({ detail }) => {
  await detail.open(); detail.assertInitialReads()
  const body = detail.html.replace(/<script\b[^>]*>[\s\S]*?<\/script>/g, '')
  expect(body).toContain('林间旅途 82'); expect(body).toContain('探索森林中的故事')
  expect(body).toContain('9/18 12:00')
  await expect(detail.page.locator('.game-detail-hero')).toContainText('9/18 12:00')
  expect(detail.html).toMatch(/rel="canonical"[^>]+\/games\/82/)
  expect(detail.html).toContain('hreflang="en-US"')
  // The current shallow SSR snapshot keeps the displayed count; the POST is a
  // side effect, not a promise that the renderer increments that snapshot.
  await expect(detail.page.locator('.game-detail-metric').first()).toContainText('120')
  await expect(detail.page.locator('.game-detail-cover')).toBeVisible()
  await expect(detail.page.locator('.link-tag')).toHaveCount(2)
  await expect(detail.root).toHaveCSS('background-color', 'rgba(0, 0, 0, 0)')
  await expect(detail.page.locator('.game-detail-title')).toHaveCSS('font-size', '24px')
  await expect(detail.page.locator('.game-detail-title')).toHaveCSS('line-height', '31.9999px')
  await expect(detail.page.locator('.game-detail-summary')).toHaveCSS('line-height', '22.75px')
  detail.assertQuiet()
})

test('real recommended and locale navigation reset identity without accepting old pagination', async ({ detail }) => {
  await detail.open(); await detail.tab('comment')
  const held = detail.hold('/reviews', { id: '82', page: '2' })
  await detail.page.locator('.game-detail-comments button').click(); await held.wait()
  await detail.page.locator('.game-detail-similar-viewport .game-detail-similar-item').first().click()
  await expect(detail.page).toHaveURL(/\/games\/83$/); await expect(detail.page.locator('h1')).toHaveText('林间旅途 83')
  held.release(); await expect.poll(() => held.completed).toBe(true)
  await expect(detail.page.locator('.game-detail-tab--active')).toHaveAttribute('data-game-tab', 'intro')
  await detail.tab('comment'); await expect(detail.page.locator('.game-detail-comment')).toHaveCount(5)
  await expect(detail.page.locator('.game-detail-comment-body').first()).toContainText('83')
  await detail.page.locator('.gf-nav__actions').getByRole('button', { name: 'EN', exact: true }).click()
  await expect(detail.page).toHaveURL(/\/en\/games\/83$/); await expect(detail.page.locator('h1')).toHaveText('Forest Journey 83')
  expect(detail.count('/info', { id: '83', lang: 'en' })).toBe(1)
  detail.assertQuiet()
})

test('tags, exact external actions and the existing Review consumer remain live', async ({ detail }) => {
  await detail.open()
  const tags = detail.page.locator('.game-detail-tag:not(.game-detail-tag--more)')
  await expect(tags).toHaveCount(8); await tags.first().hover()
  await expect(tags.first().locator('.game-detail-tag-tip')).toHaveCSS('opacity', '1')
  await detail.page.locator('.game-detail-tag--more').click(); await expect(tags).toHaveCount(10)
  await detail.page.locator('.game-detail-tag--more').click(); await expect(tags).toHaveCount(8)
  await detail.popup(() => detail.page.locator('.game-detail-action--primary').click(), 'https://store.steampowered.com/app/82')
  const link = detail.page.locator('.link-tag').first(); await expect(link).toHaveAttribute('rel', 'noopener noreferrer')
  await detail.popup(() => link.click(), 'https://detail.example.test/demo')
  const shared = `https://t.me/share/url?url=${encodeURIComponent(detail.page.url())}&text=${encodeURIComponent('林间旅途 82')}`
  await detail.popup(() => detail.page.locator('.game-detail-share-button[title="Telegram"]').click(), shared)
  await detail.page.locator('.game-detail-action--secondary').click(); await expect(detail.page.locator('.review-dialog')).toBeVisible()
  await detail.page.locator('.review-dialog__close').click(); await expect(detail.page.locator('.review-dialog')).toHaveCount(0)
  detail.assertQuiet()
})

test('comment pagination has bounded pending, failure recovery and no skipped page', async ({ detail }) => {
  await detail.open(); await detail.tab('comment'); await expect(detail.page.locator('.game-detail-comment')).toHaveCount(5)
  const recover = detail.fail('/reviews', { page: '2' }), held = detail.hold('/reviews', { page: '2' })
  const more = detail.page.locator('.game-detail-comments button')
  await more.click(); await held.wait(); await expect(more).toBeDisabled()
  expect(detail.count('/reviews', { page: '2' })).toBe(1)
  held.release(); await expect(detail.page.locator('[data-detail-reviews-error]')).toBeVisible()
  await expect(detail.page.locator('.game-detail-comment')).toHaveCount(5)
  // Existing GET retry at browser and proxy boundaries: two attempts each.
  const failedAttempts = detail.count('/reviews', { page: '2' }); expect(failedAttempts).toBe(4)
  recover(); await detail.page.locator('[data-detail-reviews-error] button').click()
  await expect(detail.page.locator('.game-detail-comment')).toHaveCount(7)
  await expect(detail.page.locator('.game-detail-comments button')).toHaveCount(0)
  expect(detail.count('/reviews', { page: '2' })).toBe(failedAttempts + 1); expect(detail.count('/reviews', { page: '3' })).toBe(0)
  detail.assertQuiet()
})

test('similar games preserve four-slot paging and the mobile-only consumer', async ({ detail }) => {
  await detail.open(); const sidebar = detail.page.locator('aside .game-detail-similar-list-shell')
  const height = (await sidebar.boundingBox())!.height
  await detail.page.locator('aside .game-detail-page-button').last().click()
  await expect(detail.page.locator('aside .game-detail-similar-track')).toHaveClass(/--next/)
  await expect(detail.page.locator('aside .game-detail-similar-track')).not.toHaveClass(/--next/)
  await expect(detail.page.locator('aside .game-detail-similar-viewport .game-detail-similar-item--placeholder')).toHaveCount(2)
  expect((await sidebar.boundingBox())!.height).toBe(height)
  await detail.page.setViewportSize({ width: 390, height: 844 })
  await detail.page.locator('[data-game-tab="similar"]').focus()
  await detail.page.setViewportSize({ width: 1440, height: 900 })
  await expect(detail.page.locator('[data-game-tab="intro"]')).toHaveAttribute('tabindex', '0')
  await expect(detail.page.locator('[data-game-tab="similar"]')).toHaveAttribute('tabindex', '-1')
  await detail.page.setViewportSize({ width: 390, height: 844 }); await detail.tab('similar')
  await expect(detail.page.locator('.game-detail-tab-panel .game-detail-similar-viewport')).toBeVisible()
  await detail.page.setViewportSize({ width: 1440, height: 900 })
  await expect(detail.page.locator('.game-detail-tab--active')).toHaveAttribute('data-game-tab', 'intro')
  detail.assertQuiet()
})

test('News pagination is local and Detail metadata and links remain meaningful', async ({ detail }) => {
  await detail.open(); const reads = detail.reads.length
  await detail.tab('news'); await expect(detail.page.locator('.game-detail-news-item')).toHaveCount(5)
  await detail.page.locator('.game-detail-news .game-detail-load-more').click()
  await expect(detail.page.locator('.game-detail-news-item')).toHaveCount(7)
  expect(detail.reads).toHaveLength(reads)
  await detail.tab('detail')
  await expect(detail.page.locator('.game-detail-info')).toContainText('Forest Studio')
  await expect(detail.page.locator('.game-detail-info')).toContainText('¥32.00')
  await expect(detail.page.locator('.game-detail-requirement-card')).toHaveCount(2)
  await expect(detail.page.locator('.game-detail-rating-card')).toContainText('ESRB')
  await detail.popup(() => detail.page.locator('a[href="https://detail.example.test/support"]').click(), 'https://detail.example.test/support')
  detail.assertQuiet()
})

test('successful empty content stays empty instead of becoming an unavailable slice', async ({ detail }) => {
  await detail.open({ empty: true })
  await expect(detail.page.locator('.game-detail-intro-card')).toHaveCount(0)
  await expect(detail.page.locator('.game-detail-score')).toHaveText('0.0')
  for (const key of ['gallery', 'news', 'comment']) { await detail.tab(key); await expect(detail.page.locator('.game-detail-tab-panel .game-detail-empty').first()).toBeVisible() }
  await expect(detail.page.locator('[data-detail-reviews-error], [data-detail-recommend-error]')).toHaveCount(0)
  detail.assertQuiet()
})

test('review first-read failure is distinct from empty and retries only reviews', async ({ detail }) => {
  const recover = detail.fail('/reviews', { page: '1' })
  await detail.open(); await detail.tab('comment')
  await expect(detail.page.locator('[data-detail-reviews-error]')).toBeVisible()
  await expect(detail.page.locator('.game-detail-score-meta')).not.toContainText('0')
  const before = detail.reads.length; recover()
  await detail.page.locator('[data-detail-reviews-error] button').click()
  await expect(detail.page.locator('.game-detail-comment')).toHaveCount(5)
  expect(detail.reads.slice(before).map(call => call.path)).toEqual(['/reviews'])
  detail.assertQuiet()
})

test('recommendation failure preserves the main detail and can retry locally', async ({ detail }) => {
  const recover = detail.fail('/recommend/similar')
  await detail.open(); await expect(detail.page.locator('h1')).toHaveText('林间旅途 82')
  await expect(detail.page.locator('[data-detail-recommend-error]')).toBeVisible()
  recover(); const before = detail.reads.length
  await detail.page.locator('[data-detail-recommend-error] button').click()
  await expect(detail.page.locator('.game-detail-similar-viewport .game-detail-similar-item:not([aria-hidden])')).toHaveCount(4)
  expect(detail.reads.slice(before).map(call => call.path)).toEqual(['/recommend/similar'])
  detail.assertQuiet()
})
