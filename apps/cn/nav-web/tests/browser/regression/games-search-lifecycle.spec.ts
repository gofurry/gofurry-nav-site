import { test, expect, type SearchScene } from '../fixtures/games-search'

const committed = {
  content: 'committed', tagList: '812345',
  pubStartTime: '2026-08-01 00:00:00', pubEndTime: '2026-08-31 23:59:00',
  updateStartTime: '2026-09-01 00:00:00', updateEndTime: '2026-09-20 23:59:00',
}
const committedBody = {
  pageNum: 1, pageSize: 4, content: 'committed', availability: 'available', score: false,
  remark_order: false, time_order: true, tag_list: [812345], lang: 'zh',
  pub_start_time: committed.pubStartTime, pub_end_time: committed.pubEndTime,
  update_start_time: committed.updateStartTime, update_end_time: committed.updateEndTime,
}
async function editDraft(scene: SearchScene, content: string) {
  await scene.keyword.fill(content)
  await scene.pageSize.fill('6')
  await scene.filter.getByRole('radio', { name: '未发售', exact: true }).click()
  // Real datepicker clear controls alter both update bounds; availability clears publish bounds.
  for (const picker of await scene.filter.locator('.game-date-picker').all()) {
    const clear = picker.locator('.dp--clear-btn')
    if (await clear.count()) await clear.click()
  }
  await scene.sort('最高评分').click()
  await scene.sort('最多评论').click()
  await scene.sort('预计发售优先').click()
  await scene.leaf('zh Leaf 1-1').click()
  await scene.leaf('zh Leaf 1-2').click()
}
async function query(scene: SearchScene, content: string) {
  await scene.openFilter(); await scene.keyword.fill(content); await scene.apply()
}

test('Cancel discards every edited filter field and the next page uses only committed criteria', async ({ search }) => {
  await search.open({ query: committed })
  expect(search.calls('advanced')[0]!.body).toEqual(committedBody)
  const originalUrl = search.page.url()
  await search.openFilter()
  const initialDates = await search.filter.locator('.dp__input').evaluateAll(inputs => inputs.map(input => (input as HTMLInputElement).value))
  expect(initialDates).toEqual(['08/01/2026, 00:00', '08/31/2026, 23:59', '09/01/2026, 00:00', '09/20/2026, 23:59'])
  await editDraft(search, 'cancel-draft'); await search.cancel()
  expect(search.page.url()).toBe(originalUrl)
  expect(search.calls('advanced')).toHaveLength(1)
  await search.waitResults('committed')
  await search.openFilter()
  await expect(search.keyword).toHaveValue('committed')
  await expect(search.pageSize).toHaveValue('4')
  await expect(search.filter.getByRole('radio', { name: '已发售', exact: true })).toHaveAttribute('aria-checked', 'true')
  await expect(search.sort('最高评分')).not.toHaveClass(/--active/)
  await expect(search.sort('最多评论')).not.toHaveClass(/--active/)
  await expect(search.sort('最新情报')).toHaveClass(/--active/)
  await expect(search.leaf('zh Leaf 1-1')).toHaveClass(/--active/)
  await expect(search.leaf('zh Leaf 1-2')).not.toHaveClass(/--active/)
  for (const [index, value] of initialDates.entries()) {
    await expect(search.filter.locator('.dp__input').nth(index)).toHaveValue(value)
  }
  await search.cancel()
  await search.page.locator('.game-search-page-button').filter({ hasText: /^2$/ }).click()
  await search.waitResults('committed 5')
  expect(search.calls('advanced')).toHaveLength(2)
  expect(search.calls('advanced')[1]!.body).toEqual({ ...committedBody, pageNum: 2 })
  search.assertQuiet()
})

test('Apply commits one complete snapshot, clears inactive dates and resets page only once', async ({ search }) => {
  await search.open({ query: { ...committed, pageNum: '3' } })
  await search.openFilter(); await editDraft(search, 'applied'); await search.apply()
  await search.waitResults('applied', 6)
  const applied = { pageNum: 1, pageSize: 6, content: 'applied', availability: 'upcoming', score: true,
    remark_order: true, time_order: false, tag_list: [812346], lang: 'zh' }
  expect(search.calls('advanced')).toHaveLength(2)
  expect(search.calls('advanced')[1]!.body).toEqual(applied)
  expect(Object.fromEntries(new URL(search.page.url()).searchParams)).toEqual({ pageSize: '6', content: 'applied',
    availability: 'upcoming', score: 'true', remarkOrder: 'true', timeOrder: 'false', tagList: '812346' })
  await search.openFilter()
  await expect(search.keyword).toHaveValue('applied')
  await expect(search.leaf('zh Leaf 1-2')).toHaveClass(/--active/)
  for (const input of await search.filter.locator('.dp__input').all()) await expect(input).toHaveValue('')
  await search.apply() // Unchanged normalized URL still permits one intentional refresh.
  await expect.poll(() => search.calls('advanced').length).toBe(3)
  await search.waitResults('applied', 6)
  expect(search.calls('advanced')[2]!.body).toEqual(applied)
  await search.openFilter(); await search.keyword.fill('discard-after-apply'); await search.cancel()
  await search.page.locator('.game-search-page-button').filter({ hasText: /^2$/ }).click()
  await search.waitResults('applied 7', 6)
  expect(search.calls('advanced')).toHaveLength(4)
  expect(search.calls('advanced')[3]!.body).toEqual({ ...applied, pageNum: 2 })
  search.assertQuiet()
})

test('Simple search supersedes its exact old request without reporting wrapped aborts', async ({ search }) => {
  await search.open()
  const old = search.hold('simple', 'slow'), fresh = search.hold('simple', 'fresh')
  await search.input.fill('slow'); await old.waitReceived(); search.expectAbort(old)
  await search.input.fill('fresh'); await fresh.waitReceived(); await search.waitAborted(old)
  old.release(); await old.waitCompleted()
  await expect(search.page.locator('.search-results-panel')).toHaveCount(0)
  fresh.release(); await fresh.waitCompleted()
  await expect(search.page.locator('.search-result-title')).toHaveText(['fresh 1', 'fresh 2', 'fresh 3', 'fresh 4'])
  expect(search.calls('simple').map(call => call.body)).toEqual([{ txt: 'slow', lang: 'zh' }, { txt: 'fresh', lang: 'zh' }])
  expect(search.calls('advanced')).toHaveLength(1)
  search.assertQuiet()
})

test('Home simple search stays cleared after the pending request is canceled and released', async ({ search }) => {
  await search.open({ home: true })
  const old = search.hold('simple', 'clear-me')
  await search.input.fill('clear-me'); await old.waitReceived(); search.expectAbort(old)
  await search.input.fill(''); await search.waitAborted(old)
  old.release(); await old.waitCompleted()
  await expect(search.input).toHaveValue('')
  await expect(search.page.locator('.search-results-panel')).toHaveCount(0)
  await search.input.focus()
  await expect(search.page.locator('.search-results-panel')).toHaveCount(0)
  expect(search.calls('simple')).toHaveLength(1)
  expect(search.calls('home')).toHaveLength(1)
  search.assertQuiet()
})

test('Simple search unmount cancels its request while the destination Home remains clean', async ({ search }) => {
  await search.open()
  const old = search.hold('simple', 'leaving')
  await search.input.fill('leaving'); await old.waitReceived(); search.expectAbort(old)
  await search.leaveForHome(); await search.waitAborted(old)
  old.release(); await old.waitCompleted()
  await expect(search.input).toHaveValue('')
  await expect(search.page.locator('.search-results-panel')).toHaveCount(0)
  expect(search.calls('simple')).toHaveLength(1)
  expect(search.calls('home')).toHaveLength(1)
  search.assertQuiet()
})

test('Advanced search supersession cannot clear the latest pending state or overwrite its result', async ({ search }) => {
  await search.open()
  const old = search.hold('advanced', 'slow'), fresh = search.hold('advanced', 'fresh')
  await query(search, 'slow'); await old.waitReceived(); search.expectAbort(old)
  await query(search, 'fresh'); await fresh.waitReceived(); await search.waitAborted(old)
  old.release(); await old.waitCompleted()
  await expect(search.page.locator('.search-results')).toHaveClass(/search-results--pending/)
  fresh.release(); await fresh.waitCompleted(); await search.waitResults('fresh')
  expect(search.calls('advanced').map(call => call.body?.content)).toEqual(['', 'slow', 'fresh'])
  search.assertQuiet()
})

test('Advanced search unmount releases only the old page request and leaves Home usable', async ({ search }) => {
  await search.open()
  const old = search.hold('advanced', 'leaving')
  await query(search, 'leaving'); await old.waitReceived(); search.expectAbort(old)
  await search.leaveForHome(); await search.waitAborted(old)
  old.release(); await old.waitCompleted()
  await expect(search.page.locator('.games-search-page')).toHaveCount(0)
  expect(search.calls('advanced')).toHaveLength(2)
  expect(search.calls('home')).toHaveLength(1)
  search.assertQuiet()
})

test('Locale navigation cancels old categories and late metadata preserves current draft selection', async ({ search }) => {
  const old = search.hold('tags', 'zh')
  // The departing page's locale watcher runs before the new locale page mounts.
  // Preserve that existing lifecycle, accounting for each exact canceled request.
  const departing = search.hold('tags', 'en'), fresh = search.hold('tags', 'en')
  await search.open({ query: { content: 'committed', tagList: '812345' } })
  const departingResult = search.hold('advanced', 'committed'), freshResult = search.hold('advanced', 'committed')
  await old.waitReceived(); search.expectAbort(old)
  await search.page.locator('.gf-nav').getByRole('button', { name: 'EN', exact: true }).click()
  await expect(search.page).toHaveURL(/\/en\/games\/search/)
  await departing.waitReceived(); search.expectAbort(departing)
  await departingResult.waitReceived(); search.expectAbort(departingResult)
  await fresh.waitReceived(); await search.waitAborted(old)
  await search.waitAborted(departing)
  await freshResult.waitReceived(); await search.waitAborted(departingResult)
  freshResult.release(); await freshResult.waitCompleted(); await search.waitResults('committed')
  await search.openFilter(); await search.keyword.fill('locale-draft')
  old.release(); departing.release(); departingResult.release()
  await old.waitCompleted(); await departing.waitCompleted(); await departingResult.waitCompleted()
  await expect(search.filter.locator('.game-search-filter-group-title')).toHaveCount(0)
  fresh.release(); await fresh.waitCompleted()
  await expect(search.filter.locator('.game-search-filter-group-title')).toHaveText(['en Category 1', 'en Category 2'])
  await expect(search.keyword).toHaveValue('locale-draft')
  await expect(search.leaf('en Leaf 1-1')).toHaveClass(/--active/)
  await search.leaf('en Leaf 1-2').click()
  await search.cancel(); await search.openFilter()
  await expect(search.keyword).toHaveValue('committed')
  await expect(search.leaf('en Leaf 1-1')).toHaveClass(/--active/)
  await expect(search.leaf('en Leaf 1-2')).not.toHaveClass(/--active/)
  await search.cancel()
  expect(search.calls('tags').map(call => call.query.lang)).toEqual(['zh', 'en', 'en'])
  expect(search.calls('advanced').map(call => call.body?.lang)).toEqual(['zh', 'en', 'en'])
  search.assertQuiet()
})
