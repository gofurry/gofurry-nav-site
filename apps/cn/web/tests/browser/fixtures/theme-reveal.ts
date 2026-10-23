import { expect, type Locator, type Page } from '@playwright/test'

type RevealProbe = {
  starts: number
  writes: string[]
  animations: { frames: Keyframe[] | PropertyIndexedKeyframes | null; options: KeyframeAnimationOptions;
    oldAnimation: string; newAnimation: string; oldZ: string; newZ: string; shellColor: string; artOpacity?: string }[]
}
declare global { interface Window { themeRevealProbe: RevealProbe } }

// Observe the real APIs without replacing their clocks, snapshots or animation.
export async function observeThemeReveal(page: Page) {
  await page.evaluate(() => {
    window.themeRevealProbe = { starts: 0, writes: [], animations: [] }
    const write = Storage.prototype.setItem
    Storage.prototype.setItem = function (key, value) {
      if (key === 'theme') window.themeRevealProbe.writes.push(value)
      return write.call(this, key, value)
    }
    const start = document.startViewTransition.bind(document)
    document.startViewTransition = update => { window.themeRevealProbe.starts++; return start(update) }
    const root = document.documentElement, animate = root.animate.bind(root)
    root.animate = (frames, options) => {
      const animation = animate(frames, options)
      if (typeof options === 'object' && options.pseudoElement === '::view-transition-new(root)') {
        const old = getComputedStyle(root, '::view-transition-old(root)'), next = getComputedStyle(root, '::view-transition-new(root)')
        const art = document.querySelector('.error-page__artwork-layer.is-active')
        window.themeRevealProbe.animations.push({ frames, options,
          oldAnimation: old.animationName, newAnimation: next.animationName, oldZ: old.zIndex, newZ: next.zIndex,
          shellColor: getComputedStyle(document.querySelector('.gf-app-shell') ?? root).backgroundColor,
          artOpacity: art ? getComputedStyle(art).opacity : undefined,
        })
        window.dispatchEvent(new Event('theme-reveal-started'))
      }
      return animation
    }
  })
}

export async function assertNativeReveal(page: Page, button: Locator, target: 'light' | 'dark', action: 'click' | 'Enter' | 'Space' = 'click') {
  const before = await page.evaluate(() => ({ calls: window.themeRevealProbe.animations.length,
    starts: window.themeRevealProbe.starts, writes: window.themeRevealProbe.writes.length, scroll: scrollY }))
  const rect = (await button.boundingBox())!, viewport = page.viewportSize()!
  const x = rect.x + rect.width / 2, y = rect.y + rect.height / 2
  if (action === 'click') await button.click()
  else { await button.focus(); await button.press(action) }
  await expect.poll(() => page.evaluate(() => window.themeRevealProbe.animations.length)).toBe(before.calls + 1)
  await expect(page.locator('html')).not.toHaveAttribute('data-gf-theme-reveal')
  expect(await page.evaluate(() => window.themeRevealProbe.starts)).toBe(before.starts + 1)
  expect(await page.evaluate(() => window.themeRevealProbe.writes.slice(-1))).toEqual([target])
  expect(await page.evaluate(() => window.themeRevealProbe.writes.length)).toBe(before.writes + 1)
  expect(await page.evaluate(() => localStorage.getItem('theme'))).toBe(target)
  expect(await page.locator('html').evaluate(el => el.classList.contains('dark'))).toBe(target === 'dark')
  expect(await page.evaluate(() => scrollY)).toBe(before.scroll)
  const result = await page.evaluate(() => window.themeRevealProbe.animations.at(-1)!)
  expect(result.frames).toEqual({ clipPath: [`circle(0px at ${x}px ${y}px)`,
    `circle(${Math.hypot(Math.max(x, viewport.width - x), Math.max(y, viewport.height - y))}px at ${x}px ${y}px)`] })
  expect(result.options).toEqual({ duration: 420, easing: 'cubic-bezier(0.22, 1, 0.36, 1)', pseudoElement: '::view-transition-new(root)', fill: 'both' })
  expect(result).toMatchObject({ oldAnimation: 'none', newAnimation: 'none', oldZ: '1', newZ: '2' })
  return result
}
