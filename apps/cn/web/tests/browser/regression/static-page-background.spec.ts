import { createHash } from 'node:crypto'
import { mkdir } from 'node:fs/promises'
import { test, expect } from '../fixtures/nav-shell'

for (const width of [1440, 390]) for (const theme of ['light', 'dark'] as const) {
  test(`static page background ignores retired preferences: ${width}px ${theme}`, async ({ page, context, navShell }) => {
    const source = theme === 'light' ? 'server' : 'local'
    await context.addInitScript(source => {
      localStorage.setItem('gf_background_preference', JSON.stringify({ version: 1, source, pattern_id: '1', overrides: { opacity: 1, size_px: 1, color: '#ff0000' } }))
      const opened: string[] = []
      Object.assign(window, { backgroundDatabaseReads: opened })
      const open = indexedDB.open.bind(indexedDB)
      indexedDB.open = (name, version) => { opened.push(name); return open(name, version) }
    }, source)
    await navShell.open({ width, theme })
    const html = await (await page.request.get('/terms')).text()
    expect(html).toContain('data-public-background')
    expect(html).toContain('data-pattern-status="default"')
    const response = await page.request.get('/web/background/background.webp')
    expect(response.status()).toBe(200)
    expect(response.headers()['content-type']).toContain('image/webp')
    expect(createHash('sha256').update(await response.body()).digest('hex'))
      .toBe('f2f59f8c44bcc77986f7757fde4b32aa127e5d5606089c962b57cca5557042c1')
    const background = page.locator('.gf-public-background__pattern')
    const expectedOpacity = theme === 'light' ? '0.055' : '0.075'
    await expect(background).toHaveCSS('background-image', /\/web\/background\/background\.webp/)
    await expect(background).toHaveCSS('mask-image', 'none')
    await expect(background).toHaveCSS('background-repeat', 'repeat')
    await expect(background).toHaveCSS('background-size', '512px 512px')
    await expect(background).toHaveCSS('opacity', expectedOpacity)
    await expect(page.locator('[data-public-background]')).toHaveCSS('opacity', '1')
    await expect(page.locator('[data-public-background]')).toHaveCSS('pointer-events', 'none')
    await page.evaluate(() => {
      window.dispatchEvent(new Event('gf-background-change'))
      window.dispatchEvent(new Event('focus'))
      window.dispatchEvent(new StorageEvent('storage', { key: 'gf_background_preference' }))
    })
    await expect(background).toHaveCSS('opacity', expectedOpacity)
    expect(await page.evaluate(() => (window as Window & { backgroundDatabaseReads?: string[] }).backgroundDatabaseReads)).toEqual([])
    if (process.env.GOFURRY_149_REVIEW === '1') {
      await mkdir('../../../.cache/149-review', { recursive: true })
      await page.screenshot({ path: `../../../.cache/149-review/background-${width}-${theme}.png` })
    }
    navShell.assertQuiet()
  })
}
