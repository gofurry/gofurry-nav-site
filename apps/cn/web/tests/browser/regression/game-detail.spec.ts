import { test, expect, openGame, selectTab, expectMissingFacts } from '../fixtures/game-detail'
import { captureBrowserErrors } from '../fixtures/browser-errors'

test.describe('missing historical data', () => {
  test.use({ viewport: { width: 1200, height: 900 } })
  for (const legacyNull of [true, false]) {
    test(`${legacyNull ? 'null' : 'empty'} regions keep missing values through repeated tab switches`, async ({ page, game }) => {
      game.state.legacyNull = legacyNull
      const errors = captureBrowserErrors(page)
      await openGame(page)
      for (const [tab, label] of [['insights', '生态观测'], ['gallery', '画廊'], ['detail', '详情'], ['intro', '介绍'], ['insights', '生态观测']]) {
        await selectTab(page, tab!)
        await expect(page.locator('.game-detail-tab--active')).toHaveText(label!)
        if (tab === 'insights') await expectMissingFacts(page)
        else await expect(page.locator('[data-game-insights]')).not.toBeVisible()
      }
      expect(errors).toEqual([])
    })
  }
})

for (const width of [390, 768, 1440, 1920]) {
  test.describe(`${width}px layout`, () => {
    test.use({ viewport: { width, height: 900 } })
    for (const adult of [false, true]) {
      test(`${adult ? 'adult' : 'ordinary'} gallery stays contained across all tabs`, async ({ page, game }) => {
        Object.assign(game.state, { gallery: true, adult, legacyNull: false })
        const errors = captureBrowserErrors(page)
        await openGame(page)
        let mainWidth: number | undefined
        for (const tab of ['intro', 'gallery', 'insights', 'comment', 'news', 'detail', 'gallery']) {
          await selectTab(page, tab)
          const box = await page.locator('.game-detail-layout').evaluate(layout => {
            const main = layout.querySelector<HTMLElement>(':scope > section')!
            const sidebar = layout.querySelector<HTMLElement>(':scope > aside')!
            const rect = main.getBoundingClientRect()
            return { width: rect.width, right: rect.right, scroll: main.scrollWidth, client: main.clientWidth,
              sidebar: sidebar.getBoundingClientRect().width, sidebarLeft: sidebar.getBoundingClientRect().left,
              layoutRight: layout.getBoundingClientRect().right }
          })
          mainWidth ??= box.width
          expect(Math.abs(box.width - mainWidth), `${tab}: main width changed`).toBeLessThan(1)
          expect(box.right, `${tab}: main escaped viewport`).toBeLessThanOrEqual(width)
          expect(box.scroll, `${tab}: main overflowed`).toBeLessThanOrEqual(box.client + 1)
          if (width >= 1280) {
            expect(Math.abs(box.width / box.sidebar - 3), `${tab}: 75/25 columns changed`).toBeLessThan(0.02)
            expect(box.right).toBeLessThanOrEqual(box.sidebarLeft)
            expect(box.sidebarLeft + box.sidebar).toBeLessThanOrEqual(box.layoutRight)
          }
          if (tab === 'gallery') {
            const gallery = await page.locator('.game-detail-gallery').evaluate(element => {
              const stage = element.querySelector<HTMLElement>('.game-detail-media-stage')!
              const thumbs = element.querySelector<HTMLElement>('.game-detail-thumb-grid')!
              return { width: element.clientWidth, scroll: element.scrollWidth, stage: stage.getBoundingClientRect().width,
                thumbs: thumbs.clientWidth, thumbsScroll: thumbs.scrollWidth, overflow: getComputedStyle(thumbs).overflowX }
            })
            expect(gallery.scroll).toBeLessThanOrEqual(gallery.width + 1)
            expect(gallery.stage).toBeLessThanOrEqual(gallery.width + 1)
            expect(gallery.thumbsScroll).toBeGreaterThan(gallery.thumbs)
            expect(gallery.overflow).toBe('auto')
            if (!adult) {
              await page.locator('.game-detail-thumb').last().click()
              await expect(page.locator('.game-detail-media-image')).toHaveAttribute('src', /shot-23/)
            }
          }
        }
        expect(errors).toEqual([])
      })
    }
  })
}

for (const width of [390, 1440]) {
  test.describe(`${width}px NSFW confirmation`, () => {
    test.use({ viewport: { width, height: 900 } })
    for (const dark of [false, true]) {
      test(`${dark ? 'dark' : 'light'} modal fits and Cancel keeps adult content locked`, async ({ page, game }) => {
        Object.assign(game.state, { gallery: true, adult: true })
        const errors = captureBrowserErrors(page)
        await openGame(page)
        await selectTab(page, 'gallery')
        await page.evaluate(dark => document.documentElement.classList.toggle('dark', dark), dark)
        const unlock = page.locator('.blur-wrapper__unlock').first()
        await unlock.click()
        const modal = page.locator('.gf-modal--compact')
        await expect(modal).toBeVisible()
        const bounds = await modal.evaluate(el => ({
          left: el.getBoundingClientRect().left, right: el.getBoundingClientRect().right,
          scroll: el.scrollWidth, client: el.clientWidth,
        }))
        expect(bounds.left).toBeGreaterThanOrEqual(0)
        expect(bounds.right).toBeLessThanOrEqual(width)
        expect(bounds.scroll).toBeLessThanOrEqual(bounds.client + 1)
        await modal.locator('.gf-button--ghost').click()
        await expect(modal).toHaveCount(0)
        await expect(unlock).toBeVisible()
        expect(errors).toEqual([])
      })
    }
  })
}

for (const path of ['/games/999999', '/en/games/999999', '/games/abc']) {
  test(`authoritative 404: ${path}`, async ({ request }) => {
    expect((await request.get(path)).status()).toBe(404)
  })
}
for (const path of ['/games/82', '/en/games/82']) {
  test(`upstream failure remains 503: ${path}`, async ({ request, game }) => {
    game.state.failure = true
    expect((await request.get(path)).status()).toBe(503)
  })
}
