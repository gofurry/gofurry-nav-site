import type { Page } from '@playwright/test'
import { runtimeTest, expect, openRuntime } from './insights-runtime'

export const patternKey = 'nav/patterns/' + 'a'.repeat(32) + '.svg'
const storageTest = runtimeTest(() => ({}),
  url => url.pathname === '/api/v2/nav/appearance/patterns',
  () => ({ data: { schema_version: 1, patterns: [{ id: '12', name: 'Storage fixture', name_en: 'Storage fixture',
    object_key: patternKey, light_color: '#123456', dark_color: '#abcdef',
    light_opacity: 0.2, dark_opacity: 0.3, default_size_px: 160 }] } }),
)
export const test = storageTest.extend({
  page: async ({ page, runtime }, use) => {
    const mutations: string[] = []
    page.on('request', request => { if (!['GET', 'HEAD'].includes(request.method())) mutations.push(request.method() + ' ' + request.url()) })
    const quiet = runtime.assertQuiet
    runtime.assertQuiet = () => { quiet(); expect(mutations).toEqual([]) }
    await use(page)
  },
})
export async function openBackground(page: Page, navigate = true) {
  if (navigate) await openRuntime(page, '/terms')
  await page.locator('.gf-nav__mode-button').click()
  await expect(page.locator('.gf-modal-backdrop')).toBeVisible()
  const tab = page.getByRole('tab', { name: '页面背景', exact: true })
  await tab.click(); await expect(tab).toHaveAttribute('aria-selected', 'true')
  const panel = page.locator('[data-background-preferences]')
  await expect(panel.getByRole('button', { name: '本地图片', exact: true })).toBeVisible()
  return panel
}
export async function saveBackground(page: Page) {
  await page.locator('.gf-modal__header').getByRole('button', { name: '保存', exact: true }).click()
  await expect(page.locator('.gf-modal-backdrop')).toHaveCount(0)
}
export async function storedBackground(page: Page) {
  return page.evaluate(async () => {
    const record = await new Promise<{ blob: Blob; kind: string; name: string } | undefined>((resolve, reject) => {
      const open = indexedDB.open('gofurry-page-background', 1)
      open.onerror = () => reject(open.error)
      open.onsuccess = () => {
        const db = open.result
        if (!db.objectStoreNames.contains('local-background')) { db.close(); resolve(undefined); return }
        const tx = db.transaction('local-background', 'readonly')
        const get = tx.objectStore('local-background').get('selected')
        tx.oncomplete = () => { db.close(); resolve(get.result) }
        tx.onerror = () => { db.close(); reject(tx.error) }
      }
    })
    const raw = localStorage.getItem('gf_background_preference')
    let decoded: { width: number; height: number } | null = null
    if (record?.kind === 'raster') {
      const bitmap = await createImageBitmap(record.blob)
      decoded = { width: bitmap.width, height: bitmap.height }; bitmap.close()
    }
    return { raw, preference: JSON.parse(raw || 'null'), record: record ? {
      kind: record.kind, name: record.name, size: record.blob.size, text: await record.blob.text(),
      decoded, digest: Array.from(new Uint8Array(await crypto.subtle.digest('SHA-256', await record.blob.arrayBuffer())))
        .map(byte => byte.toString(16).padStart(2, '0')).join(''),
    } : null }
  })
}
export { expect }
