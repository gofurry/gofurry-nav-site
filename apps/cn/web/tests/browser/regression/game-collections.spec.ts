import { test, expect, openRuntime, settleRuntime, changeCollectionMode, collectionPath } from '../fixtures/game-collections'

for (const locale of ['zh', 'en']) {
  const prefix = locale === 'en' ? '/en' : ''
  test(`Collection Index SSR SFW, preview cardinality, SEO and localized routes (${locale})`, async ({ page, runtime }) => {
    const html = await openRuntime(page, `${prefix}/games/collections`)
    expect(html).not.toContain('Adult preview')
    expect(html).not.toContain('adult-preview.svg')
    expect(runtime.calls).toHaveLength(1)
    expect(runtime.calls[0]!.url.searchParams.get('mode')).toBe('sfw')
    const cards = page.locator('.game-collection-card')
    await expect(cards).toHaveCount(6)
    for (const [i, count] of [3, 2, 1, 0].entries()) await expect(cards.nth(i).locator('img')).toHaveCount(count)
    await expect(cards.first()).toHaveAttribute('href', `${prefix}/games/collections/collection-3`)
    await expect(page).toHaveTitle(locale === 'en' ? 'GoFurry Game Collections - Curated Furry Game Timelines' : 'GoFurry 游戏分区 - 兽人游戏主题合集与发展脉络')
    runtime.assertQuiet()
  })
  test(`Detail Backend chronology, phase boundaries and date precision (${locale})`, async ({ page, runtime }) => {
    const html = await openRuntime(page, `${prefix}/games/collections/collection-3`)
    expect(html).not.toContain('Adult game'); expect(html).not.toContain('adult-game.svg')
    const nodes = page.locator('.game-collection-timeline li')
    expect(await nodes.evaluateAll(elements => elements.map(el => el.getAttribute('data-game-id')))).toEqual(['21', '19', '7', '8', '30', '40', '41', '42', '43', '50', '60'])
    await expect(page.getByTestId('collection-now')).toContainText(locale === 'en' ? 'October 6, 2026' : '2026年10月6日')
    await expect(page.locator('.game-collection-time').first()).toHaveText(locale === 'en' ? 'Approx. 2004' : '约 2004')
    await expect(page.locator('[data-game-id="41"]')).toContainText(locale === 'en' ? 'Q2 2027' : '2027 Q2')
    await expect(page.locator('[data-game-id="40"]')).toContainText(locale === 'en' ? 'still upcoming' : '当前仍未发售')
    for (const id of ['8', '30', '43', '50', '60']) await expect(page.locator(`[data-game-id="${id}"]`)).toHaveAttribute('data-connector', 'none')
    await expect(nodes.first().getByRole('link')).toHaveAttribute('href', `${prefix}/games/21`)
    expect(runtime.calls).toHaveLength(1)
    runtime.assertQuiet()
  })
}

for (const path of ['/games/collections', '/games/collections/collection-3']) {
  test(`NSFW hydration is one client refresh after SFW SSR: ${path}`, async ({ page, context, runtime }) => {
    await context.addInitScript(() => localStorage.setItem('mode', 'nsfw'))
    const html = await openRuntime(page, path)
    expect(html).not.toContain('Adult game'); expect(html).not.toContain('adult-preview.svg')
    await expect(page.locator('.game-collections-page')).toContainText(path.endsWith('collection-3') ? 'Adult game' : 'Adult collection')
    expect(runtime.calls.map(c => c.url.searchParams.get('mode'))).toEqual(['sfw', 'nsfw'])
    expect(page.url()).not.toContain('mode=')
    runtime.assertQuiet()
  })
  test(`Mode race and refresh failure keep the latest ready content: ${path}`, async ({ page, runtime }) => {
    await openRuntime(page, path)
    const old = runtime.hold(url => url.searchParams.get('mode') === 'nsfw')
    await changeCollectionMode(page, 'nsfw'); await old.wait()
    await changeCollectionMode(page, 'sfw')
    await expect.poll(() => runtime.calls.filter(c => c.completed).length).toBe(2)
    old.release(); await old.done(); await settleRuntime(page)
    await expect(page.locator('.game-collections-page')).not.toContainText('Adult game')
    await expect(page.locator('.game-collections-page')).not.toContainText('Adult collection')
    runtime.state.modeFailure = true
    await changeCollectionMode(page, 'nsfw')
    await expect(page.getByRole('alert')).toContainText('仍为你保留')
    await expect(page.locator('.game-collection-card').first()).toBeVisible()
    runtime.state.modeFailure = false
    await page.getByRole('button', { name: '重试', exact: true }).click()
    await expect(page.locator('.game-collections-page')).toContainText(path.endsWith('collection-3') ? 'Adult game' : 'Adult collection')
    expect(runtime.calls.map(c => c.url.searchParams.get('mode'))).toEqual(['sfw', 'nsfw', 'sfw', 'nsfw', 'nsfw'])
    runtime.assertQuiet()
  })
  test(`Initial 503 keeps a real retry surface: ${path}`, async ({ page, runtime }) => {
    runtime.state.failure = true
    runtime.expectedURLs.add(new URL(path, runtime.app.base).href)
    const response = await page.goto(path)
    expect(response!.status()).toBe(503)
    await expect(page.getByRole('alert')).toContainText('暂时无法加载')
    await settleRuntime(page)
    expect(runtime.calls).toHaveLength(1)
    runtime.state.failure = false
    await page.getByRole('button', { name: '重试', exact: true }).click()
    await expect(page.locator('.game-collection-card').first()).toBeVisible()
    expect(runtime.calls).toHaveLength(2)
    runtime.assertQuiet()
  })
}

test('Index load more retains cards on failure, dedupes, and mode success resets page one', async ({ page, runtime }) => {
  await openRuntime(page, '/games/collections')
  runtime.state.pageFailure = true
  await page.getByRole('button', { name: '加载更多' }).click()
  await expect(page.locator('.game-collections-load')).toContainText('加载失败')
  await expect(page.locator('.game-collection-card')).toHaveCount(6)
  runtime.state.pageFailure = false
  await page.getByRole('button', { name: '重试', exact: true }).click()
  await expect(page.locator('.game-collection-card')).toHaveCount(8)
  await changeCollectionMode(page, 'nsfw')
  await expect(page.locator('.game-collection-card')).toHaveCount(6)
  await expect(page.getByRole('button', { name: '加载更多' })).toBeVisible()
  expect(runtime.calls.map(c => c.url.searchParams.get('page'))).toEqual(['1', '2', '2', '1'])
  runtime.assertQuiet()
})

test('Old load-more response cannot append to a new mode snapshot', async ({ page, runtime }) => {
  await openRuntime(page, '/games/collections')
  const more = runtime.hold(url => url.searchParams.get('page') === '2')
  await page.getByRole('button', { name: '加载更多' }).click(); await more.wait()
  await changeCollectionMode(page, 'nsfw')
  await expect(page.locator('.game-collections-page')).toContainText('Adult collection')
  more.release(); await more.done(); await settleRuntime(page)
  await expect(page.locator('.game-collection-card')).toHaveCount(6)
  runtime.assertQuiet()
})

test('Real Index to Detail and locale navigation use independent route identities', async ({ page, runtime }) => {
  await openRuntime(page, '/games/collections')
  await page.locator('.game-collection-card').first().click()
  await expect(page).toHaveURL(/\/games\/collections\/collection-3$/)
  await expect(page.locator('.game-collection-timeline')).toBeVisible()
  await settleRuntime(page)
  await page.getByRole('button', { name: 'EN', exact: true }).click()
  await expect(page).toHaveURL(/\/en\/games\/collections\/collection-3$/)
  await expect(page.locator('h1')).toHaveText('Forest stories 3')
  await expect(page.getByTestId('collection-now')).toContainText('October 6, 2026')
  await settleRuntime(page)
  expect(runtime.calls.map(c => [c.url.pathname, c.url.searchParams.get('lang'), c.url.searchParams.get('mode')])).toEqual([
    [collectionPath, 'zh', 'sfw'], [collectionPath + '/collection-3', 'zh', 'sfw'], [collectionPath + '/collection-3', 'en', 'sfw'],
  ])
  runtime.assertQuiet()
})

test('Empty visible timeline is 200 and unknown-only omits NOW', async ({ page, runtime }) => {
  await openRuntime(page, '/games/collections/adult-only')
  await expect(page.locator('.game-collections-page')).toContainText('当前暂无可展示作品')
  await expect(page.locator('.game-collections-page')).not.toContainText(/成人|隐藏/)
  await expect(page.getByTestId('collection-now')).toHaveCount(0)
  await openRuntime(page, '/games/collections/unknown-only')
  await expect(page.getByTestId('collection-now')).toHaveCount(0)
  await expect(page.locator('.game-collection-timeline li')).toHaveCount(1)
  runtime.assertQuiet()
})

test('Detail missing is authoritative HTTP 404, never an empty collection', async ({ request, runtime }) => {
  runtime.state.missing = true
  const response = await request.get(runtime.app.base + '/games/collections/missing')
  expect(response.status()).toBe(404)
  expect(await response.text()).not.toContain('当前暂无可展示作品')
  expect(runtime.calls).toHaveLength(1)
  runtime.assertQuiet()
})

test('An authoritative 404 during mode refresh enters the real Nuxt error page', async ({ page, runtime }) => {
  await openRuntime(page, '/games/collections/collection-3')
  runtime.state.missing = true
  await changeCollectionMode(page, 'nsfw')
  await expect(page.locator('.error-page__code')).toHaveText('404')
  await expect(page.locator('.game-collection-timeline')).toHaveCount(0)
  expect(runtime.calls).toHaveLength(2)
  runtime.assertQuiet()
})

for (const width of [390, 768, 1440]) {
  test(`Index grid and visible keyboard focus at ${width}px`, async ({ page, runtime }) => {
    await page.setViewportSize({ width, height: 900 })
    await openRuntime(page, '/games/collections')
    const columns = await page.locator('.game-collections-grid').evaluate(el => getComputedStyle(el).gridTemplateColumns.split(' ').length)
    expect(columns).toBe(width === 390 ? 1 : width === 768 ? 2 : 3)
    const card = page.locator('.game-collection-card').first()
    await page.keyboard.press('Tab'); await card.focus()
    await expect(card).toBeFocused()
    expect(await card.evaluate(el => getComputedStyle(el).outlineStyle)).not.toBe('none')
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
    runtime.assertQuiet()
  })
}

for (const width of [390, 899, 900, 1440]) {
  test(`Timeline geometry and overflow at ${width}px`, async ({ page, runtime }) => {
    await page.setViewportSize({ width, height: 900 })
    await openRuntime(page, '/games/collections/collection-3')
    const nodes = page.locator('[data-phase="past"] li')
    const boxes = await Promise.all([0, 1, 2, 3].map(i => nodes.nth(i).boundingBox()))
    if (width < 900) {
      expect(boxes[1]!.y).toBeGreaterThan(boxes[0]!.y)
      expect(boxes[1]!.x).toBe(boxes[0]!.x)
    } else {
      expect(boxes[1]!.y).toBe(boxes[0]!.y)
      expect(boxes[3]!.x).toBe(boxes[2]!.x)
      expect(boxes[3]!.y).toBeGreaterThan(boxes[2]!.y)
    }
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
    runtime.assertQuiet()
  })
}

test.describe('Date-only presentation in a western timezone', () => {
  test.use({ timezoneId: 'America/Los_Angeles' })
  test('keeps server and browser day labels identical', async ({ page, runtime }) => {
    await openRuntime(page, '/en/games/collections/collection-3')
    await expect(page.locator('[data-game-id="7"] .game-collection-time')).toHaveText('January 2, 2020')
    expect(runtime.count(collectionPath + '/collection-3')).toBe(1)
    runtime.assertQuiet()
  })
})


test('Index toolbar, IME debounce and atomic filter draft share the search contract', async ({ page, runtime }) => {
  await openRuntime(page, '/games/collections')
  await page.clock.install()
  const search = page.getByRole('searchbox')
  const filter = page.getByRole('button', { name: '高级筛选', exact: true })
  await expect(page.locator('h1')).toHaveClass('sr-only')
  await expect(page.locator('.game-collections-page > header')).toHaveCount(0)
  const a = await search.boundingBox(), b = await filter.boundingBox()
  expect(Math.abs(a!.y - b!.y)).toBeLessThan(4)
  await search.dispatchEvent('compositionstart')
  await search.fill('sen'); await page.clock.runFor(800)
  await search.fill('森林'); await page.clock.runFor(800)
  expect(runtime.calls).toHaveLength(1)
  await search.dispatchEvent('compositionend'); await page.clock.runFor(350)
  await expect.poll(() => runtime.calls.length).toBe(2)
  expect(runtime.calls[1]!.url.searchParams.get('q')).toBe('森林')
  await expect(page.locator('.game-collection-card').first()).toContainText('森林 · published_desc')
  await filter.click()
  const dialog = page.getByRole('dialog')
  await dialog.getByRole('button', { name: '包含待发行作品', exact: true }).click()
  await dialog.getByRole('button', { name: '作品数量从多到少' }).click()
  await dialog.getByRole('button', { name: '取消', exact: true }).click()
  expect(runtime.calls).toHaveLength(2)
  await expect(filter).toBeFocused()
  await filter.click()
  await expect(dialog.getByRole('button', { name: '全部', exact: true })).toHaveAttribute('aria-pressed', 'true')
  await dialog.getByRole('button', { name: '已发行与待发行', exact: true }).click()
  await dialog.getByRole('button', { name: '名称降序' }).click()
  await dialog.getByRole('button', { name: '应用', exact: true }).click()
  await expect.poll(() => runtime.calls.length).toBe(3)
  expect(Object.fromEntries(runtime.calls[2]!.url.searchParams)).toMatchObject({q:'森林',phase:'mixed',sort:'name_desc',page:'1'})
  await changeCollectionMode(page, 'nsfw')
  await expect.poll(() => runtime.calls.length).toBe(4)
  expect(Object.fromEntries(runtime.calls[3]!.url.searchParams)).toMatchObject({q:'森林',phase:'mixed',sort:'name_desc',mode:'nsfw',page:'1'})
  await expect(page.locator('.game-collection-card').first()).toContainText('Adult collection')
  await page.getByRole('button', {name:'加载更多'}).click()
  await expect.poll(() => runtime.calls.length).toBe(5)
  expect(Object.fromEntries(runtime.calls[4]!.url.searchParams)).toMatchObject({q:'森林',phase:'mixed',sort:'name_desc',mode:'nsfw',page:'2'})
  await filter.click(); await page.keyboard.press('Shift+Tab')
  expect(await dialog.evaluate(el => el.contains(document.activeElement))).toBe(true)
  await page.keyboard.press('Escape'); await expect(dialog).toHaveCount(0)
  await expect(filter).toBeFocused()
  runtime.assertQuiet()
})

test('Index keeps ready cards and ignores stale search and load-more responses', async ({ page, runtime }) => {
  await openRuntime(page, '/games/collections')
  await page.clock.install()
  const search = page.getByRole('searchbox')
  const older = runtime.hold(url => url.searchParams.get('q') === '旧搜索')
  await search.fill('旧搜索'); await page.clock.runFor(350); await older.wait()
  await expect(page.locator('.game-collection-card')).toHaveCount(6)
  await search.fill('新搜索'); await page.clock.runFor(350)
  await expect(page.locator('.game-collection-card').first()).toContainText('新搜索')
  older.release(); await older.done()
  await expect(page.locator('.game-collection-card').first()).toContainText('新搜索')
  const more = runtime.hold(url => url.searchParams.get('page') === '2')
  await page.getByRole('button', {name:'加载更多'}).click(); await more.wait()
  await search.fill('不存在')
  more.release(); await more.done(); await page.clock.runFor(350)
  await expect(page.locator('.game-collection-card')).toHaveCount(0)
  await expect(page.locator('.game-collections-page')).toContainText('暂无')
  runtime.assertQuiet()
})

test('Index accepts only mode-visible search results and Games card material', async ({ page, runtime }) => {
  await openRuntime(page, '/games/collections')
  const card = page.locator('.game-collection-index-card').first()
  const tokens = await card.evaluate(el => {
    const s = getComputedStyle(el)
    return {border:s.borderColor, radius:s.borderRadius, padding:s.padding,
      background:s.backgroundColor, shadow:s.boxShadow}
  })
  const material = await card.evaluate(el => {
    const probe = document.createElement('span')
    probe.style.backgroundColor = 'var(--games-home-card-bg)'
    probe.style.boxShadow = 'var(--games-home-card-shadow)'
    el.append(probe)
    const style = getComputedStyle(probe)
    const background = style.backgroundColor, shadow = style.boxShadow
    probe.style.backgroundColor = 'var(--games-home-card-hover-bg)'
    const hover = getComputedStyle(probe).backgroundColor
    probe.remove()
    return { background, shadow, hover }
  })
  expect(tokens.background).toBe(material.background); expect(tokens.shadow).toBe(material.shadow)
  expect(tokens.border).toBe('rgba(0, 0, 0, 0)'); expect(tokens.radius).toBe('12px'); expect(tokens.padding).toBe('8px')
  await expect(card.locator('.game-collection-card__collage')).toHaveCSS('border-radius', '6px')
  await card.hover(); await expect(card).toHaveCSS('border-color', 'rgba(0, 0, 0, 0)')
  await expect(card).toHaveCSS('background-color', material.hover)
  await page.clock.install()
  await page.getByRole('searchbox').fill('隐藏作品'); await page.clock.runFor(350)
  await expect(page.locator('.game-collection-card')).toHaveCount(0)
  await changeCollectionMode(page, 'nsfw')
  await expect(page.locator('.game-collection-card').first()).toContainText('Adult collection')
  expect(runtime.calls.at(-1)!.url.searchParams.get('q')).toBe('隐藏作品')
  runtime.assertQuiet()
})

for (const locale of ['zh', 'en']) {
  test(`Detail compact header, localized decorations and Games material (${locale})`, async ({ page }) => {
    await openRuntime(page, `${locale === 'en' ? '/en' : ''}/games/collections/collection-3`)
    const header = page.locator('.game-collection-detail-header')
    const title = await header.locator('h1').boundingBox(), count = await header.locator('.game-collection-note').boundingBox(), back = await header.locator('.game-collection-back').boundingBox()
    expect(title && count && back).toBeTruthy()
    expect(count!.x).toBeGreaterThan(title!.x + title!.width)
    expect(Math.abs(count!.y + count!.height - title!.y - title!.height)).toBeLessThan(12)
    expect(back!.x).toBeGreaterThan(count!.x + count!.width)
    await expect(header.locator('.game-collection-copy')).not.toBeEmpty()
    const card = page.locator('.game-collection-timeline-card').first()
    await expect(card.locator('.game-collection-timeline-tag--primary')).toHaveText(locale === 'en' ? 'Story' : '剧情')
    await expect(card.locator('.game-collection-timeline-tag').last()).toHaveText(locale === 'en' ? 'Visual novel' : '视觉小说')
    const metrics = card.locator('.game-collection-timeline-metrics')
    await expect(metrics).toContainText('4.6 · 28')
    await expect(metrics).toContainText('1,234')
    await expect(metrics.locator('[aria-label]').first()).toHaveAttribute('aria-label', locale === 'en' ? 'Rating 4.6 from 28 reviews' : '评分 4.6，共 28 条评价')
    const online = metrics.getByRole('img', { name: locale === 'en' ? /^Latest observed players: 1,234, collected at/ : /^最近观测在线人数 1,234，采集于/ })
    await expect(online).toHaveAttribute('title', await online.getAttribute('aria-label') ?? '')
    for (const part of ['2026', '08:00', 'UTC']) await expect(online).toHaveAttribute('aria-label', new RegExp(part))
    await expect(metrics.locator('svg[aria-hidden="true"]')).toHaveCount(3)
    await expect(metrics.getByRole('img', { name:locale === 'en' ? 'Rating 4.6 from 28 reviews' : '评分 4.6，共 28 条评价', exact:true })).toHaveCount(1)
    await expect(page.locator('[data-game-id="8"] .game-collection-timeline-metrics')).toBeEmpty()
    await expect(page.locator('[data-game-id="7"] .game-collection-timeline-metrics')).toContainText('0')
    const material = await card.evaluate(el => {
      const s = getComputedStyle(el), probe = document.createElement('span')
      probe.style.backgroundColor = 'var(--games-home-card-bg)'; probe.style.boxShadow = 'var(--games-home-card-shadow)'; el.append(probe)
      const expectedBg = getComputedStyle(probe).backgroundColor, expectedShadow = getComputedStyle(probe).boxShadow
      probe.style.backgroundColor = 'var(--games-home-card-hover-bg)'; const hover = getComputedStyle(probe).backgroundColor; probe.remove()
      return { border:s.borderColor, bg:s.backgroundColor, radius:s.borderRadius, shadow:s.boxShadow, expectedBg, expectedShadow, hover,
        title:getComputedStyle(el.querySelector('h3')!).minHeight, summary:getComputedStyle(el.querySelector('.game-collection-copy')!).minHeight }
    })
    expect(material.border).toBe('rgba(0, 0, 0, 0)'); expect(material.radius).toBe('12px')
    expect(material.bg).toBe(material.expectedBg); expect(material.shadow).toBe(material.expectedShadow)
    expect(parseFloat(material.title)).toBeGreaterThan(40); expect(parseFloat(material.summary)).toBeGreaterThan(40)
    const now = page.getByTestId('collection-now')
    await expect(now.locator('time')).toHaveAttribute('datetime','2026-10-06')
    expect(await now.evaluate(el => ({ tag:el.tagName, border:getComputedStyle(el).borderColor, bg:getComputedStyle(el).backgroundColor }))).toMatchObject({ tag:'P', border:'rgba(0, 0, 0, 0)',bg:material.bg })
    await expect(now.getByRole('button')).toHaveCount(0)
    await card.hover(); await expect(card).toHaveCSS('background-color',material.hover); await expect(card).toHaveCSS('border-color','rgba(0, 0, 0, 0)')
  })
}

for (const width of [390, 768, 1440]) {
  test(`Detail flowing connectors, reduced motion and fixed card geometry at ${width}`, async ({ page }) => {
    await page.setViewportSize({ width, height:900 })
    await page.emulateMedia({ reducedMotion:'no-preference' })
    await openRuntime(page, '/games/collections/flow-long')
    const nodes = page.locator('.game-collection-timeline__item')
    expect(await nodes.evaluateAll(els => els.map(el => el.getAttribute('data-game-id')))).toEqual(['200','201','202','203','204','205','206'])
    expect(await nodes.evaluateAll(els => els.map(el => el.getAttribute('data-connector')))).toEqual(['right','right','down','left','left','down','none'])
    for (const [index, direction] of [[0,'right'],[2,'down'],[3,'left']] as const) {
      const css = await nodes.nth(index).evaluate(el => { const s=getComputedStyle(el,'::after'); return { width:s.width,height:s.height,image:s.backgroundImage,animation:s.animationName,direction:s.animationDirection,pointer:s.pointerEvents } })
      const horizontal = width >= 900 && direction !== 'down'
      expect(css.width).toBe(horizontal ? '40px' : '2px'); expect(css.height).toBe(horizontal ? '2px' : '40px')
      expect(css.image).toContain('repeating-linear-gradient'); expect(css.animation).toBe(horizontal ? 'collection-flow-right' : 'collection-flow-down')
      expect(css.direction).toBe(horizontal && direction==='left' ? 'reverse' : 'normal'); expect(css.pointer).toBe('none')
    }
    await page.emulateMedia({ reducedMotion:'reduce' })
    expect(await nodes.first().evaluate(el => getComputedStyle(el,'::after').animationName)).toBe('none')
    expect(await nodes.first().evaluate(el => getComputedStyle(el,'::after').backgroundImage)).toContain('repeating-linear-gradient')
    const heights = await page.locator('.game-collection-timeline-card').evaluateAll(els => els.map(el => el.getBoundingClientRect().height))
    expect(Math.max(...heights)-Math.min(...heights)).toBeLessThanOrEqual(1)
    for (const part of ['.game-collection-title','.game-collection-copy']) await expect(nodes.nth(1).locator(part)).toHaveCSS('-webkit-line-clamp','2')
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
    await openRuntime(page, '/games/collections/collection-3')
    for (const phase of ['releasedUnknown','tba','unknown']) {
      expect(await page.locator(`[data-phase="${phase}"] li`).evaluateAll(els => els.every(el => getComputedStyle(el,'::after').content === 'none'))).toBe(true)
    }
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
  })
}
