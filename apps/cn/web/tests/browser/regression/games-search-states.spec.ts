import { test, expect, type SearchScene } from '../fixtures/games-search'

const errorState = (search: SearchScene) => search.page.locator('.game-search-state[data-state="error"]')
const simpleState = (search: SearchScene) => search.page.locator('.search-status-panel')
async function query(search: SearchScene, value: string) {
  await search.openFilter(); await search.keyword.fill(value); await search.apply()
}
async function retry(search: SearchScene) {
  await errorState(search).getByRole('button', { name: '重试', exact: true }).click()
}

test('Initial result failure recovers without locking later filters or pagination', async ({ search }) => {
  search.respond('advanced', '', [503])
  await search.open({ ready: false })
  await expect(errorState(search)).toHaveAttribute('role', 'alert')
  await expect(errorState(search)).toContainText('搜索暂时不可用')
  await expect(search.page.locator('.game-search-pagination')).toHaveCount(0)
  const url = search.page.url()
  const recovery = search.hold('advanced', '')
  await retry(search); await recovery.waitReceived()
  await expect(errorState(search)).toHaveCount(0)
  await expect(search.page.locator('.search-results')).toHaveAttribute('aria-busy', 'true')
  recovery.release(); await recovery.waitCompleted(); await search.waitResults('Result')
  expect(search.page.url()).toBe(url)
  expect(search.calls('advanced')).toHaveLength(2)
  expect(search.calls('tags')).toHaveLength(1)
  await query(search, 'recovered'); await search.waitResults('recovered')
  await search.page.locator('.game-search-page-button').filter({ hasText: /^2$/ }).click()
  await search.waitResults('recovered 5')
  expect(search.calls('advanced')).toHaveLength(4)
  search.assertQuiet()
})

test('Category failure is independent and category retry preserves the open filter draft', async ({ search }) => {
  // Keep this state test independent of the separately audited short-page overlay defect.
  await search.page.setViewportSize({ width: 1440, height: 1100 })
  search.respond('tags', 'zh', [503, 503]) // Two browser attempts, each with the proxy's GET retry.
  await search.open({ ready: false }); await search.waitResults('Result')
  await expect.poll(() => search.calls('tags').length).toBe(4)
  await query(search, 'ready'); await search.waitResults('ready')
  await search.page.locator('.game-search-page-button').filter({ hasText: /^2$/ }).click()
  await search.waitResults('ready 5')
  const count = search.calls('advanced').length
  await search.openFilter(); await search.keyword.fill('unsaved')
  const state = search.filter.locator('.game-search-tag-state')
  await expect(state).toHaveAttribute('data-state', 'error')
  await expect(state).toContainText('其他筛选条件仍可使用')
  const gate = search.hold('tags', 'zh')
  await state.getByRole('button', { name: '重新加载标签' }).click(); await gate.waitReceived()
  await expect(state).toHaveAttribute('data-state', 'pending')
  gate.release(); await gate.waitCompleted()
  await expect(search.leaf('zh Leaf 1-1')).toBeVisible()
  await expect(search.keyword).toHaveValue('unsaved')
  expect(search.calls('advanced')).toHaveLength(count)
  expect(search.calls('tags')).toHaveLength(5)
  await search.cancel(); await search.openFilter()
  await expect(search.keyword).toHaveValue('ready'); await search.cancel()
  search.assertQuiet()
})

test('A failed second page never presents first-page cards as its successful result', async ({ search }) => {
  search.respond('advanced', '', ['success', 503])
  await search.open()
  await search.page.locator('.game-search-page-button').filter({ hasText: /^2$/ }).click()
  await expect(errorState(search)).toBeVisible()
  const url = search.page.url()
  await expect(search.page.locator('.search-result-page-slide:not([aria-hidden="true"]) .search-page-card')).toHaveCount(0)
  await expect(search.page.locator('.game-search-pagination')).toHaveCount(0)
  await retry(search); await search.waitResults('Result 5')
  expect(search.page.url()).toBe(url)
  expect(search.calls('advanced')).toHaveLength(3)
  expect(search.calls('advanced')[2]!.body).toEqual(search.calls('advanced')[1]!.body)
  search.assertQuiet()
})

test('Initial loading, empty results and later success are distinct real states', async ({ search }) => {
  search.respond('advanced', '', ['empty'])
  const gate = search.hold('advanced', '')
  await search.open({ ready: false }); await gate.waitReceived()
  await expect(search.page.locator('.search-results')).toHaveAttribute('aria-busy', 'true')
  await expect(search.page.locator('.search-result-page-slide .search-page-card--skeleton')).toHaveCount(4)
  await expect(search.page.locator('.game-search-pagination')).toHaveCount(0)
  gate.release(); await gate.waitCompleted()
  const empty = search.page.locator('.game-search-state[data-state="empty"]')
  await expect(empty).toContainText('没有找到符合条件的游戏')
  await expect(empty).toContainText('0')
  await expect(errorState(search)).toHaveCount(0)
  await expect(search.page.locator('.game-search-pagination')).toHaveCount(0)
  await query(search, 'found'); await search.waitResults('found')
  await expect(empty).toHaveCount(0)
  search.assertQuiet()
})

test('Home simple search exposes current-query failure, retry and empty states', async ({ search }) => {
  search.respond('simple', 'fox', [503])
  search.respond('simple', 'none', ['empty'])
  await search.open({ home: true })
  await search.input.fill('wolf')
  await expect(search.page.locator('.search-result-title').first()).toHaveText('wolf 1')
  await search.input.fill('fox')
  await expect(simpleState(search)).toHaveAttribute('data-state', 'error')
  await expect(search.page.locator('.search-result-title')).toHaveCount(0)
  const gate = search.hold('simple', 'fox')
  await simpleState(search).getByRole('button', { name: '重试', exact: true }).click()
  await gate.waitReceived()
  await expect(search.input).toBeFocused()
  await expect(simpleState(search)).toHaveAttribute('data-state', 'pending')
  await expect(simpleState(search).getByRole('button')).toHaveCount(0)
  gate.release(); await gate.waitCompleted()
  await expect(search.page.locator('.search-result-title').first()).toHaveText('fox 1')
  await search.input.fill('none')
  await expect(simpleState(search)).toHaveAttribute('data-state', 'empty')
  await expect(simpleState(search)).toContainText('没有找到相关游戏')
  expect(search.calls('simple')).toHaveLength(4)
  search.assertQuiet()
})

test('Search simple response after blur stays closed, and clear discards an old error', async ({ search }) => {
  search.respond('simple', 'fox', [503])
  await search.open()
  await search.input.fill('wolf')
  await expect(search.page.locator('.search-result-title').first()).toHaveText('wolf 1')
  const gate = search.hold('simple', 'fox')
  await search.input.fill('fox'); await gate.waitReceived()
  await search.page.locator('.game-search-pagination-total').click()
  await expect(simpleState(search)).toHaveCount(0)
  gate.release(); await gate.waitCompleted()
  await expect(simpleState(search)).toHaveCount(0)
  await search.input.focus()
  await expect(simpleState(search)).toHaveAttribute('data-state', 'error')
  await simpleState(search).getByRole('button', { name: '重试', exact: true }).click()
  await expect(search.page.locator('.search-result-title').first()).toHaveText('fox 1')
  await search.input.focus(); await search.input.press('Escape')
  await expect(search.page.locator('.search-results-panel')).toHaveCount(0)
  await search.page.locator('.game-search-pagination-total').click(); await search.input.focus()
  await expect(search.page.locator('.search-result-title').first()).toHaveText('fox 1')
  await search.input.fill('')
  await expect(simpleState(search)).toHaveCount(0)
  await expect(search.page.locator('.search-results-panel')).toHaveCount(0)
  expect(search.calls('simple')).toHaveLength(3)
  search.assertQuiet()
})

test('A canceled old failure cannot replace the new pending state or successful result', async ({ search }) => {
  await search.open()
  search.respond('advanced', 'old-failure', [503])
  const old = search.hold('advanced', 'old-failure'), fresh = search.hold('advanced', 'fresh')
  await query(search, 'old-failure'); await old.waitReceived(); search.expectAbort(old)
  await query(search, 'fresh'); await fresh.waitReceived(); await search.waitAborted(old)
  old.release(); await old.waitCompleted()
  await expect(search.page.locator('.search-results')).toHaveAttribute('aria-busy', 'true')
  await expect(errorState(search)).toHaveCount(0)
  fresh.release(); await fresh.waitCompleted(); await search.waitResults('fresh')
  expect(search.calls('advanced')).toHaveLength(3)
  search.assertQuiet()
})

test('English controlled rejection is visible and retry recovers without translation keys', async ({ search }) => {
  search.respond('advanced', '', ['rejected'])
  await search.open({ locale: 'en', ready: false })
  await expect(errorState(search)).toContainText('Search is temporarily unavailable')
  await expect(errorState(search)).not.toContainText('game.search.')
  const url = search.page.url()
  await errorState(search).getByRole('button', { name: 'Retry', exact: true }).click()
  await search.waitResults('Result')
  expect(search.page.url()).toBe(url)
  expect(search.calls('advanced')).toHaveLength(2)
  expect(search.calls('tags')).toHaveLength(1)
  search.assertQuiet()
})

for (const width of [390, 1440]) for (const theme of ['light', 'dark'] as const) {
  test(`Filter remains clickable above Footer on short error and empty pages (${width}px ${theme})`, async ({ search }) => {
    await search.page.setViewportSize({ width, height: 900 })
    search.respond('advanced', '', [503])
    search.respond('advanced', 'empty', ['empty'])
    await search.open({ theme, ready: false })
    await expect(errorState(search)).toBeVisible()

    for (const state of ['error', 'empty']) {
      await expect(search.page.locator(`.game-search-state[data-state="${state}"]`)).toBeVisible()
      await search.openFilter()
      await expect.poll(() => search.filter.evaluate(panel => {
        const overlay = panel.closest('.game-search-filter-overlay')!
        const footer = document.querySelector('.gf-footer-shell')!.getBoundingClientRect()
        const box = panel.getBoundingClientRect()
        const panelHit = [.1, .5, .95].every(ratio => panel.contains(document.elementFromPoint(
          box.x + box.width / 2, box.y + box.height * ratio,
        )))
        // The footer is genuinely in this viewport; its canvas must not intercept
        // either the dialog or the backdrop above it.
        return footer.top < innerHeight && panelHit && [4, innerWidth - 4].every(x =>
          overlay.contains(document.elementFromPoint(x, Math.min(innerHeight - 4, footer.top + 20))),
        )
      })).toBe(true)
      const leaf = search.leaf('zh Leaf 1-1')
      await leaf.click()
      await expect(leaf).toHaveClass(state === 'error' ? /game-search-filter-chip--active/ : /game-search-filter-chip--idle/)
      if (state === 'error') {
        await search.keyword.fill('empty'); await search.apply()
      } else {
        await search.cancel()
        await expect(search.page.locator('.game-search-filter-overlay')).toHaveCount(0)
      }
    }
    search.assertQuiet()
  })
}
