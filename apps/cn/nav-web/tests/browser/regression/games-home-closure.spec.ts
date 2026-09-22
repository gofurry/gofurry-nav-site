import { test, expect, gamesHomeClosureTime, assertNewsAppearance, assertLatestReviewsAppearance, type GamesHomeScene } from '../fixtures/games-home'

test.use({ gamesHomeFixedNow: gamesHomeClosureTime, timezoneId: 'UTC' })

async function assertNewsPosition(scene: GamesHomeScene, index: number) {
  const { news } = scene
  await expect(news.locator('.news-pager__count')).toHaveText(`${index + 1}/3`)
  await scene.settle(news)
  const box = await news.evaluate(el => {
    const viewport = el.querySelector<HTMLElement>('.news-viewport')!
    const track = el.querySelector<HTMLElement>('.news-track')!
    const last = el.querySelector<HTMLElement>('.news-card:last-child')!.getBoundingClientRect()
    const bounds = viewport.getBoundingClientRect()
    return { offset: new DOMMatrixReadOnly(getComputedStyle(track).transform).m41,
      max: Math.max(0, track.scrollWidth - viewport.clientWidth), left: bounds.left, right: bounds.right,
      lastLeft: last.left, lastRight: last.right,
      progress: Number.parseFloat((el.querySelector<HTMLElement>('.news-progress-fill')!).style.width) }
  })
  expect(box.offset).toBeLessThanOrEqual(0)
  expect(box.offset).toBeGreaterThanOrEqual(-box.max - 1)
  expect(box.progress).toBeCloseTo((index + 1) / 3 * 100, 4)
  if (index === 0) expect(box.offset).toBe(0)
  if (index === 2) {
    expect(box.offset).toBeCloseTo(-box.max, 0)
    expect(box.lastLeft).toBeGreaterThanOrEqual(box.left - 1)
    expect(box.lastRight).toBeLessThanOrEqual(box.right + 1)
  }
  expect(await scene.page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
}

for (const config of [
  { theme: 'light', locale: 'zh', width: 1440 },
  { theme: 'dark', locale: 'en', width: 390 },
] as const) {
  test(`News SSR, bounded carousel and responsive geometry (${config.theme}/${config.locale})`, async ({ gamesHome }) => {
    const scene = await gamesHome.open({ ...config, dataset: 'news-populated' })
    const { page, news, locale } = scene
    const data = scene.data.latest_news[locale === 'zh' ? 'news_zh' : 'news_en']
    const previous = news.getByRole('button', { name: locale === 'zh' ? '上一条情报' : 'Previous update', exact: true })
    const next = news.getByRole('button', { name: locale === 'zh' ? '下一条情报' : 'Next update', exact: true })
    expect((scene.rendered.match(/class="news-card"/g) || [])).toHaveLength(3)
    expect(scene.rendered).toContain(locale === 'zh' ? '星港漫游：开发日志第二期' : 'Starport: Developer Diary Two')
    expect(scene.rendered).toContain('2026-09-17 12:30')
    await expect(news.locator('.news-card__title')).toHaveText(data.map(item => item.headline))
    await expect(news.locator('.news-card__summary').first()).toHaveText('探索 森林 & 星空。')
    await expect(news.locator('.news-card__title')).not.toContainText(locale === 'zh' ? ['Forest Journey'] : ['秋日更新'])
    await news.scrollIntoViewIfNeeded()
    await expect(news.locator('.news-viewport')).toHaveCSS('overflow-x', 'hidden')
    await expect(news.locator('[class*="news-edge"]')).toHaveCount(0)
    if (config.width === 390) await expect(news.locator('.news-pager__count')).toBeHidden()
    else await expect(news.locator('.news-pager__count')).toBeVisible()
    await assertNewsAppearance(scene)
    await expect(previous).toBeDisabled()
    await expect(next).toBeEnabled()
    await assertNewsPosition(scene, 0)
    await next.click(); await assertNewsPosition(scene, 1)
    await next.click(); await assertNewsPosition(scene, 2)
    await expect(next).toBeDisabled()
    await expect(previous).toBeEnabled()
    await previous.click(); await assertNewsPosition(scene, 1)
    await next.click(); await assertNewsPosition(scene, 2)
    if (config.width === 1440) {
      await page.setViewportSize({ width: 390, height: 900 })
      await assertNewsPosition(scene, 2)
      await expect(news.locator('.news-pager__count')).toBeHidden()
      const url = data[2]!.url
      scene.expectNewsPopup(url)
      const popupPromise = page.waitForEvent('popup')
      await news.locator('.news-card').last().click()
      const popup = await popupPromise
      await expect(popup).toHaveURL(url)
      await popup.waitForLoadState('load')
      await popup.close()
    }
    scene.assertQuiet()
  })
}

test('Empty News and Latest Reviews preserve the real empty surfaces', async ({ gamesHome }) => {
  const scene = await gamesHome.open()
  expect(scene.rendered).not.toContain('class="news-card"')
  await expect(scene.news.getByRole('heading', { name: '更新情报', exact: true })).toBeVisible()
  await expect(scene.news.locator('.news-card')).toHaveCount(0)
  await expect(scene.news.locator('.news-pager')).toHaveCount(0)
  await expect(scene.news.locator('.news-progress-track')).toHaveCount(0)
  await expect(scene.reviews.locator('.latest-review-state')).toHaveText('暂无评论')
  await expect(scene.reviews.locator('.latest-review-item')).toHaveCount(0)
  expect(scene.rendered).toContain('暂无评论')
  scene.assertQuiet()
})

test('A single News item has no pager or progress controls', async ({ gamesHome }) => {
  const scene = await gamesHome.open({ theme: 'dark', dataset: 'news-single' })
  expect((scene.rendered.match(/class="news-card"/g) || [])).toHaveLength(1)
  await scene.news.scrollIntoViewIfNeeded()
  await scene.settle(scene.news)
  await expect(scene.news.locator('.news-card')).toHaveCount(1)
  await expect(scene.news.locator('.news-pager')).toHaveCount(0)
  await expect(scene.news.locator('.news-progress-track')).toHaveCount(0)
  await expect(scene.news.locator('.news-card__title')).toHaveText('森林旅途：秋日更新 & 新章节')
  scene.assertQuiet()
})

for (const config of [{ theme: 'light', locale: 'zh' }, { theme: 'dark', locale: 'en' }] as const) {
  test(`Populated Latest Reviews retain SSR time, content and appearance (${config.theme}/${config.locale})`, async ({ gamesHome }) => {
    const scene = await gamesHome.open({ ...config, dataset: 'reviews-populated' })
    const reviews = scene.reviews.locator('.latest-review-item')
    expect((scene.rendered.match(/class="latest-review-item"/g) || [])).toHaveLength(3)
    for (const [index, item] of scene.data.latest_reviews.entries()) {
      const time = ['5 min ago', '2 hours ago', '3 days ago'][index]!
      for (const text of [item.game_name, item.content, item.region, item.ip, time]) expect(scene.rendered).toContain(text)
      await expect(reviews.nth(index).locator('.latest-review-item__title')).toHaveText(item.game_name)
      await expect(reviews.nth(index).locator('.latest-review-item__body')).toHaveText(item.content)
      await expect(reviews.nth(index).locator('.latest-review-item__meta')).toContainText(time)
      await expect(reviews.nth(index).locator('.latest-review-item__meta')).toContainText(`${config.locale === 'zh' ? '评论地区' : 'Region'}: ${item.region}`)
      await expect(reviews.nth(index).locator('.latest-review-item__meta')).toContainText(item.ip)
    }
    await expect(scene.reviews.locator('.latest-review-state')).toHaveCount(0)
    await assertLatestReviewsAppearance(scene)
    expect(await scene.reviews.evaluate(el => el.scrollWidth <= el.clientWidth)).toBe(true)
    scene.assertQuiet()
  })
}

for (const config of [{ theme: 'light', locale: 'zh' }, { theme: 'dark', locale: 'en' }] as const) {
  for (const [name, widths] of [['narrow', [390, 639, 640, 768]], ['wide', [1023, 1024, 1440]]] as const) {
    test(`Group card/rating/spacer bounds at ${name} breakpoints (${config.theme}/${config.locale})`, async ({ gamesHome }) => {
      const scene = await gamesHome.open({ ...config, width: widths[0], dataset: 'layout-stress' })
      const first = scene.group(0)
      const title = await first.locator('h3').innerText()
      const next = first.getByRole('button', { name: `${title} next page`, exact: true })
      async function assertLayout(firstCount: number, pageNumber: number) {
        await expect(scene.groups.locator('.game-group-pager > span')).toHaveText([`${pageNumber}/3`, '1/3', '1/3', '1/3'])
        await expect(first.locator('.game-group-page-live')).toHaveCount(1)
        await scene.settle(...await scene.groups.all())
        const groups = await scene.groups.evaluateAll(elements => elements.map(group => {
          const shell = group.querySelector<HTMLElement>('.game-group-page-shell')!
          const cards = [...group.querySelectorAll<HTMLElement>('.game-group-page-live .game-card')]
          const bottom = shell.getBoundingClientRect().bottom
          return { count: cards.length, overflow: getComputedStyle(shell).overflow,
            cardClipped: Math.max(...cards.map(card => card.getBoundingClientRect().bottom - bottom)),
            ratingClipped: Math.max(...cards.map(card => card.querySelector('.gf-rating')!.getBoundingClientRect().bottom - bottom)),
            spacerHeight: group.querySelector('.game-card--spacer')!.lastElementChild!.getBoundingClientRect().height,
            ratingHeight: cards[0]!.lastElementChild!.getBoundingClientRect().height }
        }))
        expect(groups).toHaveLength(4)
        for (const [index, group] of groups.entries()) {
          expect(group.count).toBe(index === 0 ? firstCount : 8)
          expect(group.overflow).toBe('hidden')
          expect(group.cardClipped).toBeLessThanOrEqual(1)
          expect(group.ratingClipped).toBeLessThanOrEqual(1)
          expect(Math.abs(group.spacerHeight - group.ratingHeight)).toBeLessThanOrEqual(1)
        }
        expect(await scene.page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
      }
      for (const width of widths) {
        await scene.page.setViewportSize({ width, height: 900 })
        await assertLayout(8, 1)
        for (const [count, pageNumber] of [[8, 2], [1, 3], [8, 1]]) {
          await next.click()
          await assertLayout(count!, pageNumber!)
        }
      }
      scene.assertQuiet()
    })
  }
}
