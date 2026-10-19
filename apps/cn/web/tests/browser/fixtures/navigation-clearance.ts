import { expect, type Page } from '@playwright/test'
import { mkdir, writeFile } from 'node:fs/promises'

export async function assertNavigationClearance(page: Page, selector: string, label: string) {
  await page.evaluate(() => document.fonts.ready)
  const evidence = await page.locator(selector).evaluateAll(elements => {
    const visibleShell = [...document.querySelectorAll('.gf-nav__actions .gf-nav__shell')]
      .find(el => getComputedStyle(el).display !== 'none')!
    const nav = visibleShell.getBoundingClientRect()
    const tier = document.querySelector('.gf-nav__utility-row')!.getBoundingClientRect()
    return elements.filter(el => el.getBoundingClientRect().width).map(el => {
      const range = document.createRange()
      range.selectNodeContents(el)
      const box = el.matches('h1') ? range.getBoundingClientRect() : el.getBoundingClientRect()
      const overlappingX = box.right > tier.left && box.left < tier.right
      const intercepted = [0.1, 0.5, 0.9].some(x => [0.1, 0.5, 0.9].some(y =>
        !!document.elementFromPoint(box.left + box.width * x, box.top + box.height * y)?.closest('.gf-nav')))
      return { text: el.getAttribute('aria-label') || el.textContent?.trim().slice(0, 60),
        top: box.top, navBottom: nav.bottom, overlappingX, intercepted }
    })
  })
  if (process.env.GOFURRY_145_C1_1_REVIEW) {
    const dir = '../../../.cache/145-c1-1-review/' + process.env.GOFURRY_145_C1_1_REVIEW
    await mkdir(dir, { recursive: true })
    await writeFile(`${dir}/${label}.json`, JSON.stringify(evidence, null, 2))
    await page.screenshot({ path: `${dir}/${label}.png` })
  }
  expect(evidence.length).toBeGreaterThan(0)
  for (const item of evidence) {
    expect(item.intercepted, JSON.stringify(item)).toBe(false)
    if (page.viewportSize()!.width >= 768 && item.overlappingX) expect(item.top - item.navBottom, JSON.stringify(item)).toBeGreaterThanOrEqual(12)
  }
  expect(await page.evaluate(() => document.documentElement.scrollWidth - innerWidth)).toBeLessThanOrEqual(1)
}
