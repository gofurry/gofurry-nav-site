import assert from 'node:assert/strict'
import { mkdir } from 'node:fs/promises'
import { join } from 'node:path'
import { launchPerfBrowser, reportsDir } from './perf/shared.mjs'
import { startInsightsFixtureApp } from './fixtures/insights-app.mjs'
import { mockGameHome } from './fixtures/insights-overview.mjs'

const app = await startInsightsFixtureApp((url, media) => {
  if (!url.pathname.endsWith('/game/home')) return { data: [] }
  const home = mockGameHome(media)
  const games = Array.from({ length: 17 }, (_, i) => ({
    ...home.panel.latest_games[0], id: String(100 + i),
    name: `布局检查 ${i + 1} — a long localized game title`,
    name_en: `Layout fixture ${i + 1} — a long game title`,
    summary: '测试多行简介不会使最后一排的评分和卡片圆角被分页容器裁掉。 A longer description for responsive layout.',
    avg_score: 4.2, comment_count: i + 1,
  }))
  for (const key of ['latest_games', 'updated_games', 'free_games', 'popular_games']) home.panel[key] = games
  return { data: home }
})
let browser
try {
  browser = await launchPerfBrowser()
  const output = join(reportsDir, 'game-group-layout')
  await mkdir(output, { recursive: true })
  for (const theme of ['light', 'dark']) {
    const page = await browser.newPage({ viewport: { width: 390, height: 900 }, colorScheme: theme })
    await page.addInitScript(theme => localStorage.setItem('theme', theme), theme)
    await page.goto(app.base + (theme === 'dark' ? '/en/games' : '/games'), { waitUntil: 'networkidle' })
    await page.waitForFunction(() => Boolean(document.querySelector('#__nuxt')?.__vue_app__))
    const check = async (width, expectedCards) => {
      const groups = await page.locator('.game-info-group').evaluateAll(groups => groups.map(group => {
        const shell = group.querySelector('.game-group-page-shell')
        const cards = [...group.querySelectorAll('.game-group-page-live .game-card')]
        const bottom = shell.getBoundingClientRect().bottom
        return {
          count: cards.length, overflow: getComputedStyle(shell).overflow,
          cardClipped: Math.max(...cards.map(card => card.getBoundingClientRect().bottom - bottom)),
          ratingClipped: Math.max(...cards.map(card => card.querySelector('.gf-rating').getBoundingClientRect().bottom - bottom)),
          placeholderRatingHeight: group.querySelector('.game-card--spacer').lastElementChild.getBoundingClientRect().height,
          ratingHeight: cards[0].lastElementChild.getBoundingClientRect().height,
        }
      }))
      assert.equal(groups.length, 4)
      for (const [index, group] of groups.entries()) {
        assert.equal(group.count, index === 0 ? expectedCards : 8)
        assert.equal(group.overflow, 'hidden', 'pagination must still clip horizontal transitions')
        assert(group.cardClipped <= 1, `${theme} ${width}px group ${index}: card clipped by ${group.cardClipped}px`)
        assert(group.ratingClipped <= 1, `${theme} ${width}px group ${index}: rating clipped by ${group.ratingClipped}px`)
        assert(Math.abs(group.placeholderRatingHeight - group.ratingHeight) <= 1, 'placeholder and real rating heights diverged')
      }
      console.log(`[game-groups] ${theme} ${width}px: four groups, page size ${expectedCards}, card/rating bounds PASS`)
    }
    for (const width of [390, 639, 640, 768, 1023, 1024, 1440]) {
      await page.setViewportSize({ width, height: 900 })
      await page.waitForTimeout(100)
      await check(width, 8)
      const next = page.locator('.game-info-group').first().locator('.game-group-pager button').last()
      for (const count of [8, 1, 8]) {
        await next.click()
        await page.waitForTimeout(460)
        await check(width, count)
      }
      if ([390, 768, 1440].includes(width)) {
        await page.locator('.game-info-group').first().locator('.game-group-page-live .game-card').last().scrollIntoViewIfNeeded()
        await page.screenshot({ path: join(output, `${theme}-${width}-last-row.png`) })
      }
    }
    await page.close()
  }
} finally {
  await browser?.close()
  await app.close()
}
