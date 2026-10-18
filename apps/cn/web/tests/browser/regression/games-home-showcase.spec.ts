import { test, expect, type GamesHomeScene } from '../fixtures/games-home'
import { showcaseOrigins } from '../fixtures/games-home-showcase-data'
import { steamSharedAssetCandidates } from '../../../app/utils/steamAssets'
import type { Page } from '@playwright/test'
import { mkdir } from 'node:fs/promises'
import { join } from 'node:path'

export async function assertGameHomeShowcaseAppearance(scene: GamesHomeScene) {
  const hero = scene.showcase
  await expect(hero).toHaveCSS('backdrop-filter', 'none')
  await expect(hero).toHaveCSS('border-radius', '16px')
  await expect(hero).toHaveCSS('transform', 'none')
  await expect(hero).toHaveCSS('animation-name', 'none')
  await expect(hero).toHaveCSS('background-color', await scene.cards(0).first().evaluate(element => getComputedStyle(element).backgroundColor))
  await expect(hero).toHaveCSS('box-shadow', await scene.cards(0).first().evaluate(element => getComputedStyle(element).boxShadow))
  await expect(hero.locator('.game-home-showcase__title')).toHaveCSS('font-weight', '600')
  await expect(hero.locator('.game-home-showcase__title')).toHaveCSS('-webkit-line-clamp', '2')
  await expect(hero.locator('.game-home-showcase__context')).toHaveCSS('font-weight', '600')
  const image = hero.locator('img:visible').first()
  await expect(image).toHaveCSS('object-fit', 'cover')
  const imageBox = await image.boundingBox(), frameBox = await hero.locator('.game-home-showcase__artwork').boundingBox()
  expect(Math.abs(imageBox!.width - frameBox!.width)).toBeLessThan(1)
  expect(Math.abs(imageBox!.height - frameBox!.height)).toBeLessThan(1)
  const box = await hero.boundingBox()
  await hero.locator('.game-home-showcase__artwork').hover()
  await expect(hero).toHaveCSS('transform', 'none')
  await expect(image).toHaveCSS('transform', 'none')
  expect(await hero.boundingBox()).toEqual(box)
  expect(await hero.locator('.game-home-showcase__context').evaluate(element => getComputedStyle(element).color))
    .not.toBe(await hero.evaluate(element => {
      const probe = document.createElement('span'); probe.style.color = 'var(--gf-danger)'
      element.append(probe); const color = getComputedStyle(probe).color; probe.remove(); return color
    }))
  const hierarchy = await hero.evaluate(element => ['context', 'summary', 'title'].map(part =>
    Number.parseFloat(getComputedStyle(element.querySelector(`.game-home-showcase__${part}`)!).fontSize)))
  expect(hierarchy[0]).toBeLessThan(hierarchy[1]!); expect(hierarchy[1]).toBeLessThan(hierarchy[2]!)
}

test('Empty and failed optional slices keep the original Home content first', async ({ gamesHome }) => {
  const scene = await gamesHome.open({ collections: 'failure' })
  await expect(scene.showcase).toHaveCount(0)
  await expect(scene.root.locator('.game-info-shell > :first-child')).toHaveClass('game-info-group')
  scene.assertQuiet()
})

test('Showcase 503 is optional and leaves catalog, statistics, News and sidebar alive', async ({ gamesHome }) => {
  const scene = await gamesHome.open({ showcase: 'showcase-failure' })
  await expect(scene.showcase).toHaveCount(0)
  await expect(scene.groups).toHaveCount(4)
  await expect(scene.stats).toBeVisible()
  await expect(scene.sidebar).toBeVisible()
  await expect(scene.news).toHaveCount(1)
  scene.assertQuiet()
})

test('Slow Showcase does not block SSR and recovers once after mount', async ({ gamesHome, page }) => {
  const scene = await gamesHome.open({ showcase: 'slow-showcase', dataset: 'news-populated' })
  expect(scene.snapshot.items).toHaveLength(1)
  expect(scene.showcasePending, 'Home rendered before the gated Showcase response').toBe(true)
  expect(scene.rendered).not.toContain('game-home-showcase')
  await expect(scene.showcase).toHaveCount(0)
  await expect(scene.group(0).getByRole('heading', { name: '最近发售', exact: true })).toBeVisible()
  await expect(scene.groups).toHaveCount(4)
  await expect(scene.stats).toBeVisible()
  await expect(scene.stats.locator('.game-stats-row:not(.game-stats-row--placeholder)')).toHaveCount(16)
  await expect(scene.sidebar).toBeVisible()
  await expect(scene.dock).toBeVisible()
  await expect(scene.news.locator('.news-card')).toHaveCount(3)
  await scene.news.getByRole('button', { name: '下一条情报', exact: true }).click()
  await expect(scene.news.locator('.news-pager__count')).toHaveText('2/3')
  await expect(page.getByRole('alert')).toHaveCount(0)
  scene.assertQuiet()

  scene.releaseShowcase()
  await expect.poll(() => scene.showcasePending).toBe(false)
  await expect(scene.showcase).toBeVisible()
  await scene.settle(scene.root)
  await expect(page.getByRole('alert')).toHaveCount(0)
  expect(scene.events).toEqual([])
  scene.assertQuiet()
})

test('Showcase transient failure recovers once without affecting Home', async ({ gamesHome }) => {
  const scene = await gamesHome.open({ showcase: 'showcase-transient' })
  expect(scene.rendered).not.toContain('game-home-showcase')
  await expect(scene.showcase).toBeVisible()
  await expect(scene.group(0)).toBeVisible()
  scene.assertQuiet()
})

for (const locale of ['zh', 'en'] as const) {
  test(`Single editorial SSR and hydrated first item, strict locale and no controls (${locale})`, async ({ gamesHome }) => {
    const scene = await gamesHome.open({ showcase: 'single-editorial', locale })
    const first = scene.snapshot.items[0]!
    expect(scene.rendered).toContain(first.title)
    expect(scene.rendered).toContain(first.summary)
    expect(scene.rendered).not.toContain(locale === 'zh' ? 'Echoes of the Wild' : '荒野回声')
    await expect(scene.showcase.locator('h2')).toHaveText(first.title)
    await expect(scene.showcase.locator('.game-home-showcase__context')).toHaveText(locale === 'zh' ? 'GoFurry 精选' : 'GoFurry Pick')
    await expect(scene.showcase.locator('.game-home-showcase__controls')).toHaveCount(0)
    await expect(scene.showcase.locator('[aria-live]')).toHaveText('')
    await expect(scene.showcase).not.toHaveAttribute('tabindex')
    await expect(scene.showcase).not.toHaveAttribute('role', 'application')
    await expect(scene.showcase.locator('.gf-button')).toHaveAttribute('href', `${locale === 'en' ? '/en' : ''}/games/6101`)
    await expect(scene.showcase.locator('.gf-button')).not.toHaveAttribute('target')
    await assertGameHomeShowcaseAppearance(scene)
    await reviewScreenshot(scene, `editorial-${locale}-desktop`)
    expect(await scene.showcase.boundingBox()).not.toBeNull()
    const hero = await scene.showcase.boundingBox(), catalog = await scene.group(0).boundingBox()
    expect(catalog!.y).toBeGreaterThan(hero!.y + hero!.height)
    scene.assertQuiet()
  })
}

test('Four items preserve supplied order, circular navigation, focus and user-only announcements', async ({ gamesHome }) => {
  const scene = await gamesHome.open({ showcase: 'four-items' })
  const { showcase: hero, page, snapshot } = scene
  const next = hero.getByRole('button', { name: '下一项精选' }), prev = hero.getByRole('button', { name: '上一项精选' })
  await expect(hero.locator('.game-home-showcase__count')).toHaveText('01 / 04')
  await expect(prev).toBeEnabled(); await expect(next).toBeEnabled()
  await expect(hero.locator('[aria-live]')).toHaveText('')
  await prev.click()
  await expect(hero.locator('.game-home-showcase__count')).toHaveText('04 / 04')
  await expect(hero.locator('[aria-live]')).toHaveText(`4 / 4 · ${snapshot.items[3]!.title}`)
  await expect(hero.locator('.game-home-showcase__media')).toHaveAttribute('data-direction', 'previous')
  await next.click()
  await expect(hero.locator('.game-home-showcase__count')).toHaveText('01 / 04')
  await expect(hero.locator('[aria-live]')).toHaveText(`1 / 4 · ${snapshot.items[0]!.title}`)
  await expect(hero.locator('.game-home-showcase__media')).toHaveAttribute('data-direction', 'next')
  await next.focus(); await page.keyboard.press('ArrowRight')
  await expect(hero.locator('h2')).toHaveText(snapshot.items[0]!.title)
  for (let i = 1; i < 4; i++) {
    await next.click()
    await expect(hero.locator('h2')).toHaveText(snapshot.items[i]!.title)
    await expect(hero.locator('[aria-live]')).toHaveText(`${i + 1} / 4 · ${snapshot.items[i]!.title}`)
    await expect(hero.locator('.game-home-showcase__count')).toHaveText(`0${i + 1} / 04`)
    if (i < 3) await expect(next).toBeFocused()
  }
  await expect(next).toBeEnabled(); await expect(prev).toBeEnabled()
  await prev.click(); await prev.click(); await prev.click()
  await expect(prev).toBeEnabled()
  await expect(hero.locator('h2')).toHaveText(snapshot.items[0]!.title)
  await expect(hero.locator('[aria-live]')).toHaveText(`1 / 4 · ${snapshot.items[0]!.title}`)
  scene.assertQuiet()
})

for (const width of [1440, 1024, 768, 390]) {
  test(`Showcase responsive composition and bounded long copy (${width})`, async ({ gamesHome }) => {
    const scene = await gamesHome.open({ showcase: 'long-content', width })
    const hero = scene.showcase, media = hero.locator('.game-home-showcase__artwork'), content = hero.locator('.game-home-showcase__content')
    const m = await media.boundingBox(), c = await content.boundingBox(), root = await hero.boundingBox()
    if (width >= 1024) {
      expect(Math.abs(m!.y - c!.y)).toBeLessThan(2)
      expect(m!.width / (m!.width + c!.width)).toBeCloseTo(.64, 2)
      expect(root!.height).toBeLessThan(350)
    } else {
      expect(c!.y).toBeGreaterThanOrEqual(m!.y + m!.height - 1)
      expect(m!.width / m!.height).toBeCloseTo(width < 640 ? 16 / 9 : 2, 1)
      if (width === 390) expect(root!.height).toBeLessThan(450)
    }
    await expect(hero.locator('.game-home-showcase__title')).toHaveCSS('-webkit-line-clamp', '2')
    await expect(hero.locator('.game-home-showcase__summary')).toHaveCSS('-webkit-line-clamp', width < 640 ? '2' : '3')
    await expect(hero.locator('.game-home-showcase__tags li:visible')).toHaveCount(width < 640 ? 2 : 3)
    if (width < 640) await expect(hero.locator('.game-home-showcase__note')).toBeHidden()
    else {
      const note = hero.locator('.game-home-showcase__note')
      await expect(note).toBeVisible()
      const n = await note.boundingBox(), action = await hero.locator('.gf-button').boundingBox()
      expect(n!.y + n!.height).toBeLessThanOrEqual(action!.y)
    }
    expect(await pageOverflow(scene)).toBe(false)
    await expect(hero.locator('.gf-button')).toBeInViewport()
    await reviewScreenshot(scene, `long-content-${width}`)
    scene.assertQuiet()
  })
}

test('Upcoming uses Steam routing and canonical calendar release metadata', async ({ gamesHome }) => {
  const scene = await gamesHome.open({ showcase: 'automatic-upcoming', locale: 'en' })
  await expect(scene.showcase.locator('.game-home-showcase__context')).toHaveText('Upcoming')
  await expect(scene.showcase.locator('.game-home-showcase__meta')).toHaveText('Expected Q2 2027')
  await expect(scene.showcase.locator('img')).toHaveAttribute('src', /shared\.akamai\.steamstatic\.com\/store_item_assets\/steam\/apps\/6101\/showcase\/header\.jpg\?v=2/)
  await expect(scene.showcase.locator('.game-home-showcase__note')).toHaveText('')
  scene.assertQuiet()
})

for (const scenario of ['managed-sponsored-tabletop', 'managed-sponsored-merchandise'] as const) {
  test(`Sponsored context and fixed external CTA (${scenario})`, async ({ gamesHome }) => {
    const scene = await gamesHome.open({ showcase: scenario, theme: 'dark', width: scenario.endsWith('merchandise') ? 390 : 1440, locale: 'en' })
    await expect(scene.showcase.locator('.game-home-showcase__context')).toHaveText(scenario.endsWith('merchandise') ? 'Merchandise · Sponsored' : 'Tabletop · Sponsored')
    await expect(scene.showcase.locator('.game-home-showcase__note')).toHaveText('')
    await expect(scene.showcase.locator('.gf-button')).toHaveText(scenario.endsWith('merchandise') ? 'View Product' : 'View Project')
    for (const link of await scene.showcase.getByRole('link').all()) {
      await expect(link).toHaveAttribute('target', '_blank')
      await expect(link).toHaveAttribute('rel', 'noopener noreferrer')
    }
    await assertGameHomeShowcaseAppearance(scene)
    await reviewScreenshot(scene, scenario)
    scene.assertQuiet()
  })
}

test('Managed Primary failure tries Mirror with focal coordinates; missing mobile key uses desktop', async ({ gamesHome }) => {
  const scene = await gamesHome.open({ showcase: 'primary-failure', width: 390 })
  const media = scene.showcase.locator('img:visible')
  await expect.poll(() => media.evaluate(image => (image as HTMLImageElement).currentSrc)).toBe(`${showcaseOrigins.mirror}/${scene.snapshot.items[0]!.artwork.mobile_object_key}`)
  await expect(media).toHaveCSS('object-position', '25% 75%')
  expect(scene.assets.some(url => url.startsWith(showcaseOrigins.primary))).toBe(true)
  expect(scene.assets.some(url => url.startsWith(showcaseOrigins.mirror))).toBe(true)
  scene.assertQuiet()
})

test('Absent mobile artwork reuses desktop through ManagedAssetImage', async ({ gamesHome }) => {
  const scene = await gamesHome.open({ showcase: 'mobile-missing', width: 390 })
  await expect(scene.showcase.locator('img')).toHaveAttribute('src', `${showcaseOrigins.primary}/${scene.snapshot.items[0]!.artwork.desktop_object_key}`)
  scene.assertQuiet()
})

for (const scenario of ['media-failure', 'steam-failure'] as const) {
  test(`Exhausted media never removes a slide, CTA or geometry (${scenario})`, async ({ gamesHome }) => {
    const scene = await gamesHome.open({ showcase: scenario })
    await expect(scene.showcase.locator('img')).toHaveCount(0)
    await expect(scene.showcase.locator('h2')).toHaveText(scene.snapshot.items[0]!.title)
    await expect(scene.showcase.locator('.gf-button')).toBeVisible()
    const before = await scene.showcase.boundingBox()
    expect(before!.height).toBeGreaterThan(300)
    if (scenario === 'media-failure') {
      await expect(scene.showcase.locator('.game-home-showcase__count')).toHaveText('01 / 04')
      await scene.showcase.getByRole('button', { name: '下一项精选' }).click()
      await expect(scene.showcase.locator('img')).toHaveCount(0)
      await expect(scene.showcase.locator('.game-home-showcase__count')).toHaveText('02 / 04')
      expect(await scene.showcase.boundingBox()).toEqual(before)
    }
    scene.assertQuiet()
  })
}

test('Real intersection sends one impression per snapshot/item and no duplicate on revisit', async ({ gamesHome }) => {
  const scene = await gamesHome.open({ showcase: 'four-items', height: 200 })
  expect(scene.events).toEqual([])
  await scene.page.setViewportSize({ width: 1440, height: 900 })
  await scene.showcase.scrollIntoViewIfNeeded()
  await expect.poll(() => scene.events.filter(event => event.event === 'impression').length).toBe(1)
  await scene.showcase.getByRole('button', { name: '下一项精选' }).click()
  await expect.poll(() => scene.events.filter(event => event.event === 'impression').length).toBe(2)
  await scene.showcase.getByRole('button', { name: '上一项精选' }).click()
  await scene.showcase.getByRole('button', { name: '下一项精选' }).click()
  await scene.showcase.getByRole('button', { name: '下一项精选' }).click()
  await expect.poll(() => scene.events.filter(event => event.event === 'impression').length).toBe(3)
  expect(scene.events.filter(event => event.tracking_token === scene.snapshot.items[0]!.tracking_token)).toHaveLength(1)
  expect(new Set(scene.events.map(event => event.session_id)).size).toBe(1)
  scene.assertQuiet()
})

for (const source of ['artwork', 'title', 'primary'] as const) {
  test(`Click ${source} precedes impression and navigates to localized game despite analytics 500`, async ({ gamesHome }) => {
    const scene = await gamesHome.open({ showcase: 'tracking-failure', locale: 'en', height: 200 })
    const selector = { artwork: '.game-home-showcase__artwork', title: 'h2 a', primary: '.gf-button' }[source]
    scene.expectGameDestination('6101')
    await scene.showcase.locator(selector).click()
    await expect(scene.page).toHaveURL(/\/en\/games\/6101$/)
    await expect(scene.page.getByRole('heading', { name: 'Showcase destination', exact: true })).toBeVisible()
    await expect.poll(() => scene.events.some(event => event.event === 'click' && event.source === source)).toBe(true)
    expect(scene.events[0]?.event).toBe('click')
    await expect.poll(() => scene.page.locator('.game-detail-page').count()).toBe(1)
    scene.assertQuiet()
  })
}

test('Secondary external popup still opens when analytics returns 500', async ({ gamesHome }) => {
  // This case owns click failure. Keep the Hero below qualified visibility so
  // a concurrent impression cannot race its request/response accounting.
  const scene = await gamesHome.open({ showcase: 'tracking-failure', height: 120 })
  const url = scene.snapshot.items[0]!.secondary_action!.target!
  scene.expectShowcasePopup(url)
  const popup = scene.page.waitForEvent('popup')
  await scene.showcase.locator('.game-home-showcase__secondary').click()
  const destination = await popup
  await expect(destination).toHaveURL(url)
  await destination.waitForLoadState('load')
  await expect.poll(() => scene.events.some(event => event.event === 'click' && event.source === 'secondary')).toBe(true)
  await expect(scene.showcase).toBeVisible()
  await expect(scene.page.getByRole('alert')).toHaveCount(0)
  await destination.close()
  scene.assertQuiet()
})

async function pageOverflow(scene: GamesHomeScene) {
  return scene.page.evaluate(() => document.documentElement.scrollWidth > innerWidth)
}

// Optional review artifacts only, never accepted Visual snapshots or a new runner.
async function reviewScreenshot(scene: GamesHomeScene, name: string) {
  const directory = process.env.GOFURRY_SHOWCASE_REVIEW_DIR
  if (!directory) return
  await mkdir(directory, { recursive: true })
  await scene.page.mouse.move(1, 1)
  await scene.settle(scene.showcase)
  await scene.page.screenshot({ path: join(directory, `${name}.png`) })
}

for (const width of [1440, 375]) {
  test(`Responsive artwork loads one variant and retains decoded nodes with cache disabled (${width})`, async ({ gamesHome, page }) => {
    const cdp = await page.context().newCDPSession(page)
    await cdp.send('Network.enable')
    await cdp.send('Network.setCacheDisabled', { cacheDisabled: true })
    const scene = await gamesHome.open({ showcase: 'two-managed', width })
    const hero = scene.showcase, variant = width < 1024 ? 'mobile_object_key' : 'desktop_object_key'
    const firstURL = `${showcaseOrigins.primary}/${scene.snapshot.items[0]!.artwork[variant]}`
    const secondURL = `${showcaseOrigins.primary}/${scene.snapshot.items[1]!.artwork[variant]}`
    const current = hero.locator('.game-home-showcase__frame[data-active="true"]')
    await expect.poll(() => current.locator('img').evaluate(image => (image as HTMLImageElement).currentSrc)).toBe(firstURL)
    const firstNode = await current.locator('img').elementHandle()
    expect(scene.assets.filter(url => url.includes('/game/showcase/'))).toEqual([firstURL])
    await hero.getByRole('button', { name: '下一项精选' }).click()
    await expect(current).toHaveAttribute('data-key', scene.snapshot.items[1]!.key)
    await expect.poll(() => current.locator('img').evaluate(image => (image as HTMLImageElement).currentSrc)).toBe(secondURL)
    for (let index = 0; index < 3; index++) {
      await hero.getByRole('button', { name: '上一项精选' }).click()
      await expect(current).toHaveAttribute('data-key', scene.snapshot.items[0]!.key)
      expect(await current.locator('img').evaluate((image, original) => image === original, firstNode)).toBe(true)
      await hero.getByRole('button', { name: '下一项精选' }).click()
      await expect(current).toHaveAttribute('data-key', scene.snapshot.items[1]!.key)
    }
    expect(scene.assets.filter(url => url.includes('/game/showcase/'))).toEqual([firstURL, secondURL])
    if (width < 640) {
      const action = await hero.locator('.gf-button').boundingBox(), controls = await hero.locator('.game-home-showcase__controls').boundingBox()
      expect(controls!.y).toBeGreaterThanOrEqual(action!.y + action!.height)
      for (const button of await hero.locator('.game-home-showcase__controls button').all()) {
        const box = await button.boundingBox(); expect(box!.width).toBeGreaterThanOrEqual(44); expect(box!.height).toBeGreaterThanOrEqual(44)
      }
      expect(await pageOverflow(scene)).toBe(false)
    }
    await reviewScreenshot(scene, `refined-two-items-${width}`)
    scene.assertQuiet()
  })
}

test('Slow artwork keeps the decoded frame, skips pending impressions and ignores stale completion after rapid navigation', async ({ gamesHome, page }) => {
  const scene = await gamesHome.open({ showcase: 'two-managed' })
  const hero = scene.showcase, first = scene.snapshot.items[0]!, second = scene.snapshot.items[1]!
  await expect(hero.locator('.game-home-showcase__frame[data-active="true"] img')).toBeVisible()
  await page.clock.install()
  const url = `${showcaseOrigins.primary}/${second.artwork.desktop_object_key}`
  const gate = scene.holdArtwork(url)
  try {
    await hero.getByRole('button', { name: '下一项精选' }).click()
    await gate.requested
    await expect(hero.locator('h2')).toHaveText(second.title)
    await expect(hero.locator('.game-home-showcase__count')).toHaveText('02 / 02')
    const visible = hero.locator('.game-home-showcase__frame[data-active="true"]')
    await expect(visible).toHaveAttribute('data-key', first.key)
    expect(await visible.locator('img').evaluate(image => (image as HTMLImageElement).naturalWidth)).toBeGreaterThan(0)
    await expect(hero.locator('.game-home-showcase__artwork')).toHaveAttribute('href', '/games/6101')
    await page.clock.runFor(1500)
    expect(scene.events.filter(event => event.event === 'impression' && event.tracking_token === second.tracking_token)).toEqual([])
    await hero.getByRole('button', { name: '上一项精选' }).click()
    const finished = page.waitForResponse(url)
    gate.release()
    await finished
    await expect(hero.locator('.game-home-showcase__frame[data-active="false"] img')).toHaveJSProperty('complete', true)
    await page.clock.runFor(300)
    await expect(visible).toHaveAttribute('data-key', first.key)
    await hero.getByRole('button', { name: '下一项精选' }).click()
    await expect(visible).toHaveAttribute('data-key', second.key)
    await page.clock.runFor(999)
    expect(scene.events.filter(event => event.event === 'impression' && event.tracking_token === second.tracking_token)).toEqual([])
    await page.clock.runFor(1)
    await expect.poll(() => scene.events.filter(event => event.event === 'impression' && event.tracking_token === second.tracking_token).length).toBe(1)
    expect(scene.assets.filter(value => value === url)).toHaveLength(1)
    scene.assertQuiet()
  } finally { gate.release() }
})

test('First render stays still; manual switching retains existing motion and reduced motion is instant', async ({ gamesHome, page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  const scene = await gamesHome.open({ showcase: 'two-managed' })
  const hero = scene.showcase
  await expect(hero.locator('.game-home-showcase__copy')).toHaveCSS('animation-name', 'none')
  await hero.getByRole('button', { name: '下一项精选' }).click()
  await expect(hero.locator('.game-home-showcase__copy')).toHaveCSS('animation-name', 'games-home-showcase-copy')
  await expect(hero.locator('.game-home-showcase__copy')).toHaveCSS('animation-duration', '0.2s')
  await expect(hero.getByRole('button', { name: '上一项精选' })).toBeEnabled()
  await hero.getByRole('button', { name: '上一项精选' }).click()
  await expect(hero.locator('h2')).toHaveText(scene.snapshot.items[0]!.title)
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await hero.getByRole('button', { name: '下一项精选' }).click()
  await expect(hero.locator('.game-home-showcase__copy')).toHaveCSS('animation-name', 'none')
  await expect(hero.locator('.game-home-showcase__frame[data-active="true"]')).toHaveCSS('transition-duration', '0s')
  scene.assertQuiet()
})

// Install before navigation, while the normal reduced-motion fixture keeps
// autoplay disabled. Freeze after its RAF/font readiness, then enable motion
// outside the viewport. Native IntersectionObserver establishes the exact start.
async function autoplayScene(page: Page, open: () => Promise<GamesHomeScene>) {
  await page.clock.install({ time: new Date('2026-10-04T04:00:00Z') })
  const scene = await open()
  await page.clock.pauseAt(await page.evaluate(() => Date.now() + 100))
  await page.emulateMedia({ reducedMotion: 'no-preference' })
  await expect(scene.showcase.getByRole('button', { name: '暂停自动轮播' })).toBeVisible()
  await setAutoplayViewport(scene, true)
  return scene
}

async function setAutoplayViewport(scene: GamesHomeScene, visible: boolean) {
  await scene.page.setViewportSize({ width: 1440, height: visible ? 900 : 200 })
  await scene.showcase.evaluate((element, expected) => new Promise<void>(resolve => {
    const observer = new IntersectionObserver(entries => {
      if (entries.some(entry => (entry.isIntersecting && entry.intersectionRatio >= .5) === expected)) {
        observer.disconnect(); resolve()
      }
    }, { threshold: [0, .5] })
    observer.observe(element)
  }), visible)
}
async function leaveShowcase(scene: GamesHomeScene) {
  await scene.page.mouse.move(1, 1)
  await scene.page.evaluate(() => (document.activeElement as HTMLElement | null)?.blur())
}
async function documentVisibility(scene: GamesHomeScene, visible: boolean) {
  // Controlled platform boundary: both real composables receive the same
  // visibilitychange, with their real observers/timers/network senders intact.
  await scene.page.evaluate(value => {
    Object.defineProperty(document, 'visibilityState', { configurable: true, value: value ? 'visible' : 'hidden' })
    document.dispatchEvent(new Event('visibilitychange'))
  }, visible)
}
const activeArtwork = (scene: GamesHomeScene) => scene.showcase.locator('.game-home-showcase__frame[data-active="true"]')
const countLabel = (scene: GamesHomeScene) => scene.showcase.locator('.game-home-showcase__count')

test('Autoplay waits 6000ms, stays silent, and does not replace the separate one-second impression rule', async ({ gamesHome, page }) => {
  const scene = await autoplayScene(page, () => gamesHome.open({ showcase: 'four-items', height: 200 }))
  expect(scene.events).toEqual([])
  expect(scene.assets.filter(url => url.includes('/game/showcase/'))).toHaveLength(1)
  await page.clock.runFor(999); expect(scene.events).toEqual([])
  await page.clock.runFor(1); await expect.poll(() => scene.events.length).toBe(1)
  await page.clock.runFor(4999); await expect(countLabel(scene)).toHaveText('01 / 04')
  await page.clock.runFor(1); await expect(countLabel(scene)).toHaveText('02 / 04')
  await expect(activeArtwork(scene)).toHaveAttribute('data-key', scene.snapshot.items[1]!.key)
  await expect(scene.showcase.locator('[aria-live]')).toHaveText('')
  expect(scene.events).toHaveLength(1)
  await page.clock.runFor(999); expect(scene.events).toHaveLength(1)
  await page.clock.runFor(1); await expect.poll(() => scene.events.length).toBe(2)
  expect(scene.events.map(event => [event.event, event.tracking_token])).toEqual(scene.snapshot.items.slice(0, 2).map(item => ['impression', item.tracking_token]))
  scene.assertQuiet()
})

test('Autoplay retains slow artwork and starts the next full window only after decoded handoff', async ({ gamesHome, page }) => {
  const scene = await autoplayScene(page, () => gamesHome.open({ showcase: 'four-items', height: 200 }))
  const second = scene.snapshot.items[1]!
  const url = steamSharedAssetCandidates(second.artwork.url, 'china')[0]!
  const gate = scene.holdArtwork(url)
  try {
    await page.clock.runFor(6000); await gate.requested
    await expect(countLabel(scene)).toHaveText('02 / 04')
    await expect(activeArtwork(scene)).toHaveAttribute('data-key', scene.snapshot.items[0]!.key)
    expect(await activeArtwork(scene).locator('img').evaluate(image => (image as HTMLImageElement).naturalWidth)).toBeGreaterThan(0)
    await page.clock.runFor(20000)
    await expect(countLabel(scene)).toHaveText('02 / 04')
    expect(scene.events.some(event => event.tracking_token === second.tracking_token)).toBe(false)
    const response = page.waitForResponse(url); gate.release(); await response
    await expect(activeArtwork(scene)).toHaveAttribute('data-key', second.key)
    await page.clock.runFor(5999); await expect(countLabel(scene)).toHaveText('02 / 04')
    await page.clock.runFor(1); await expect(countLabel(scene)).toHaveText('03 / 04')
    await expect(scene.showcase.locator('[aria-live]')).toHaveText('')
    await expect.poll(() => scene.events.filter(event => event.tracking_token === second.tracking_token).length).toBe(1)
    expect(scene.assets.filter(value => value === url)).toHaveLength(1)
    scene.assertQuiet()
  } finally { gate.release() }
})

for (const reason of ['hover', 'focus', 'intersection', 'visibility', 'reduced-motion'] as const) {
  test(`Autoplay restarts a full window after ${reason} interruption`, async ({ gamesHome, page }) => {
    const scene = await autoplayScene(page, () => gamesHome.open({ showcase: 'two-managed', height: 200 }))
    await page.clock.runFor(3000)
    if (reason === 'hover') await scene.showcase.hover()
    if (reason === 'focus') await scene.showcase.getByRole('button', { name: '下一项精选' }).focus()
    if (reason === 'intersection') await setAutoplayViewport(scene, false)
    if (reason === 'visibility') await documentVisibility(scene, false)
    if (reason === 'reduced-motion') {
      await page.emulateMedia({ reducedMotion: 'reduce' })
      await expect(scene.showcase.getByRole('button', { name: '暂停自动轮播' })).toHaveCount(0)
      await expect(scene.showcase.getByRole('button', { name: '下一项精选' })).toBeEnabled()
    }
    await page.clock.runFor(20000); await expect(countLabel(scene)).toHaveText('01 / 02')
    if (reason === 'hover' || reason === 'focus') await leaveShowcase(scene)
    if (reason === 'intersection') await setAutoplayViewport(scene, true)
    if (reason === 'visibility') await documentVisibility(scene, true)
    if (reason === 'reduced-motion') {
      await page.emulateMedia({ reducedMotion: 'no-preference' })
      await expect(scene.showcase.getByRole('button', { name: '暂停自动轮播' })).toBeVisible()
    }
    await page.clock.runFor(5999); await expect(countLabel(scene)).toHaveText('01 / 02')
    await page.clock.runFor(1); await expect(countLabel(scene)).toHaveText('02 / 02')
    await expect.poll(() => scene.events.filter(event => event.event === 'impression').length).toBe(1)
    scene.assertQuiet()
  })
}

test('Manual Next resets autoplay, announces once, and subsequent auto switching never changes aria-live', async ({ gamesHome, page }) => {
  const scene = await autoplayScene(page, () => gamesHome.open({ showcase: 'four-items', height: 200 }))
  await page.clock.runFor(3000)
  await scene.showcase.getByRole('button', { name: '下一项精选' }).click()
  await expect(activeArtwork(scene)).toHaveAttribute('data-key', scene.snapshot.items[1]!.key)
  const announcement = `2 / 4 · ${scene.snapshot.items[1]!.title}`
  await expect(scene.showcase.locator('[aria-live]')).toHaveText(announcement)
  await page.clock.runFor(20000); await expect(countLabel(scene)).toHaveText('02 / 04')
  await leaveShowcase(scene)
  await page.clock.runFor(5999); await expect(countLabel(scene)).toHaveText('02 / 04')
  await page.clock.runFor(1); await expect(countLabel(scene)).toHaveText('03 / 04')
  await expect(scene.showcase.locator('[aria-live]')).toHaveText(announcement)
  await expect.poll(() => scene.events.filter(event => event.event === 'impression').length).toBe(2)
  scene.assertQuiet()
})

test('Pause persists across loops and Play always waits a fresh 6000ms including at the last item', async ({ gamesHome, page }) => {
  const scene = await autoplayScene(page, () => gamesHome.open({ showcase: 'two-managed', height: 200 }))
  await page.clock.runFor(3000)
  await scene.showcase.getByRole('button', { name: '暂停自动轮播' }).click()
  await leaveShowcase(scene)
  await page.clock.runFor(60000); await expect(countLabel(scene)).toHaveText('01 / 02')
  await scene.showcase.getByRole('button', { name: '继续自动轮播' }).click()
  await leaveShowcase(scene)
  await page.clock.runFor(5999); await expect(countLabel(scene)).toHaveText('01 / 02')
  await page.clock.runFor(1); await expect(countLabel(scene)).toHaveText('02 / 02')
  await expect(activeArtwork(scene)).toHaveAttribute('data-key', scene.snapshot.items[1]!.key)
  await expect(scene.showcase.getByRole('button', { name: '下一项精选' })).toBeEnabled()
  await scene.showcase.getByRole('button', { name: '暂停自动轮播' }).click()
  await leaveShowcase(scene)
  await page.clock.runFor(60000); await expect(countLabel(scene)).toHaveText('02 / 02')
  await scene.showcase.getByRole('button', { name: '继续自动轮播' }).click()
  await leaveShowcase(scene)
  await page.clock.runFor(5999); await expect(countLabel(scene)).toHaveText('02 / 02')
  await page.clock.runFor(1); await expect(countLabel(scene)).toHaveText('01 / 02')
  await expect(activeArtwork(scene)).toHaveAttribute('data-key', scene.snapshot.items[0]!.key)
  await page.clock.runFor(6000); await expect(countLabel(scene)).toHaveText('02 / 02')
  await expect(scene.showcase.locator('[aria-live]')).toHaveText('')
  await expect.poll(() => scene.events.filter(event => event.event === 'impression').length).toBe(2)
  scene.assertQuiet()
})

test('Reduced-motion default disables autoplay entirely but both manual boundaries loop and announce', async ({ gamesHome, page }) => {
  await page.clock.install()
  const scene = await gamesHome.open({ showcase: 'four-items' })
  await page.clock.pauseAt(await page.evaluate(() => Date.now() + 100))
  await expect(scene.showcase.getByRole('button', { name: /自动轮播/ })).toHaveCount(0)
  await page.clock.runFor(60000); await expect(countLabel(scene)).toHaveText('01 / 04')
  await scene.showcase.getByRole('button', { name: '上一项精选' }).click()
  await expect(activeArtwork(scene)).toHaveAttribute('data-key', scene.snapshot.items[3]!.key)
  await expect(scene.showcase.locator('[aria-live]')).toHaveText(`4 / 4 · ${scene.snapshot.items[3]!.title}`)
  await leaveShowcase(scene)
  await page.clock.runFor(60000); await expect(countLabel(scene)).toHaveText('04 / 04')
  await scene.showcase.getByRole('button', { name: '下一项精选' }).click()
  await expect(countLabel(scene)).toHaveText('01 / 04')
  await expect(scene.showcase.locator('[aria-live]')).toHaveText(`1 / 4 · ${scene.snapshot.items[0]!.title}`)
  await leaveShowcase(scene); await page.clock.runFor(60000)
  await expect(countLabel(scene)).toHaveText('01 / 04')
  await expect.poll(() => scene.events.filter(event => event.event === 'impression').length).toBe(2)
  scene.assertQuiet()
})

test('Two complete autoplay rounds wrap 4 to 1 silently and dedupe impressions and decoded artwork', async ({ gamesHome, page }) => {
  const scene = await autoplayScene(page, () => gamesHome.open({ showcase: 'four-items', height: 200 }))
  const firstNode = await activeArtwork(scene).locator('img').elementHandle()
  let firstRoundRequests = 0
  for (let step = 1; step <= 8; step++) {
    await page.clock.runFor(5999)
    await expect(countLabel(scene)).toHaveText(`0${(step - 1) % 4 + 1} / 04`)
    await page.clock.runFor(1)
    const index = step % 4
    await expect(countLabel(scene)).toHaveText(`0${index + 1} / 04`)
    await expect(activeArtwork(scene)).toHaveAttribute('data-key', scene.snapshot.items[index]!.key)
    await expect(scene.showcase.locator('[aria-live]')).toHaveText('')
    await expect(scene.showcase.locator('.game-home-showcase__media')).toHaveAttribute('data-direction', 'next')
    if (step === 4) firstRoundRequests = scene.assets.length
  }
  await page.clock.runFor(1000)
  await expect.poll(() => scene.events.filter(event => event.event === 'impression').length).toBe(4)
  for (const item of scene.snapshot.items) expect(scene.events.filter(event => event.tracking_token === item.tracking_token && event.event === 'impression')).toHaveLength(1)
  expect(await activeArtwork(scene).locator('img').evaluate((image, first) => image === first, firstNode)).toBe(true)
  expect(scene.assets).toHaveLength(firstRoundRequests)
  await expect(scene.showcase.locator('.game-home-showcase__frame')).toHaveCount(4)
  scene.assertQuiet()
})

test('A last-to-first auto wrap retains the last decoded artwork while the first responsive variant is slow', async ({ gamesHome, page }) => {
  const scene = await autoplayScene(page, () => gamesHome.open({ showcase: 'four-items', height: 200 }))
  const first = scene.snapshot.items[0]!, last = scene.snapshot.items[3]!
  await scene.showcase.getByRole('button', { name: '上一项精选' }).click()
  await expect(activeArtwork(scene)).toHaveAttribute('data-key', last.key)
  const url = `${showcaseOrigins.primary}/${first.artwork.mobile_object_key}`, gate = scene.holdArtwork(url)
  try {
    await page.setViewportSize({ width: 375, height: 900 }); await gate.requested
    expect(await activeArtwork(scene).locator('img').evaluate(image => (image as HTMLImageElement).naturalWidth)).toBeGreaterThan(0)
    await leaveShowcase(scene)
    await page.clock.runFor(6000); await expect(countLabel(scene)).toHaveText('01 / 04')
    await expect(activeArtwork(scene)).toHaveAttribute('data-key', last.key)
    const announcement = `4 / 4 · ${last.title}`
    await expect(scene.showcase.locator('[aria-live]')).toHaveText(announcement)
    await page.clock.runFor(20000); await expect(countLabel(scene)).toHaveText('01 / 04')
    const response = page.waitForResponse(url); gate.release(); await response
    await expect(activeArtwork(scene)).toHaveAttribute('data-key', first.key)
    await page.clock.runFor(5999); await expect(countLabel(scene)).toHaveText('01 / 04')
    await page.clock.runFor(1); await expect(countLabel(scene)).toHaveText('02 / 04')
    await expect(scene.showcase.locator('[aria-live]')).toHaveText(announcement)
    scene.assertQuiet()
  } finally { gate.release() }
})

for (const width of [1440, 1024, 768, 375]) {
  test(`Autoplay control uses existing appearance and English labels without overflowing (${width})`, async ({ gamesHome, page }) => {
    await page.emulateMedia({ reducedMotion: 'no-preference' })
    const scene = await gamesHome.open({ showcase: 'two-managed', locale: 'en', width })
    const hero = scene.showcase
    await hero.getByRole('button', { name: 'Pause autoplay' }).click()
    await leaveShowcase(scene)
    const pause = hero.getByRole('button', { name: 'Resume autoplay' }), next = hero.getByRole('button', { name: 'Next showcase item' })
    await expect(pause).toBeVisible()
    for (const property of ['width', 'height', 'background-color', 'color', 'border-radius', 'font-size']) {
      await expect(pause).toHaveCSS(property, await next.evaluate((element, key) => getComputedStyle(element).getPropertyValue(key), property))
    }
    expect(await pageOverflow(scene)).toBe(false)
    const root = await hero.boundingBox(), cta = await hero.locator('.gf-button').boundingBox(), controls = await hero.locator('.game-home-showcase__controls').boundingBox()
    expect(cta!.x + cta!.width).toBeLessThanOrEqual(root!.x + root!.width)
    expect(controls!.x + controls!.width).toBeLessThanOrEqual(root!.x + root!.width)
    if (width >= 1024) expect(root!.height).toBeLessThan(350)
    await reviewScreenshot(scene, `autoplay-control-${width}`)
    scene.assertQuiet()
  })
}


for (const width of [390, 768, 1024, 1440]) {
  test(`Showcase reserves geometry across sparse/editorial/sponsored slides at ${width}`, async ({ gamesHome, page }) => {
    const scene = await gamesHome.open({ showcase: 'varied-content', width })
    const hero = scene.showcase
    const heights: number[] = [], artworkHeights: number[] = []
    for (let index = 0; index < 3; index++) {
      await expect(hero.locator('h2')).toHaveText(scene.snapshot.items[index]!.title)
      heights.push((await hero.boundingBox())!.height)
      artworkHeights.push((await hero.locator('.game-home-showcase__artwork-frame').boundingBox())!.height)
      const primary = hero.locator('.game-home-showcase__primary')
      await expect(primary).toBeVisible()
      await expect(primary).toHaveCSS('border-color', 'rgba(0, 0, 0, 0)')
      await primary.hover(); await expect(primary).toHaveCSS('border-color', 'rgba(0, 0, 0, 0)')
      await expect(hero.locator('h2')).toHaveCSS('-webkit-line-clamp', '2')
      await expect(hero.locator('.game-home-showcase__summary')).toHaveCSS('-webkit-line-clamp', width < 640 ? '2' : '3')
      const rootBox = (await hero.boundingBox())!
      for (const selector of ['.game-home-showcase__actions', '.game-home-showcase__controls']) {
        const box = (await hero.locator(selector).boundingBox())!
        expect(box.y + box.height).toBeLessThanOrEqual(rootBox.y + rootBox.height + 1)
        expect(box.x + box.width).toBeLessThanOrEqual(rootBox.x + rootBox.width + 1)
      }
      expect(await hero.evaluate(el => el.scrollWidth <= el.clientWidth)).toBe(true)
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
      await reviewScreenshot(scene, `varied-${width}-${index}`)
      await hero.getByRole('button', {name:'下一项精选'}).click()
    }
    expect(Math.max(...heights) - Math.min(...heights)).toBeLessThanOrEqual(1)
    expect(Math.max(...artworkHeights) - Math.min(...artworkHeights)).toBeLessThanOrEqual(1)
    if (width >= 1024) await expect(hero.locator(':scope > div').first()).toHaveCSS('height', '320px')
    scene.assertQuiet()
  })
}
