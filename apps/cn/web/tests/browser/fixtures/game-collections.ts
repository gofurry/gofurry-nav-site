import { runtimeTest, openRuntime, settleRuntime, expect } from './insights-runtime'
import { collectionDetail, collectionIndex } from './game-collections-data'
import type { Page } from '@playwright/test'

export const collectionPath = '/api/v2/game/collections'
export const test = runtimeTest(
  () => ({ failure: false, modeFailure: false, pageFailure: false, missing: false }),
  url => url.pathname === collectionPath || url.pathname.startsWith(collectionPath + '/'),
  (url, media, _body, state) => {
    const mode = url.searchParams.get('mode')!, lang = url.searchParams.get('lang')!, page = Number(url.searchParams.get('page') || 1)
    expect(['sfw', 'nsfw']).toContain(mode)
    expect(['zh', 'en']).toContain(lang)
    if (state.missing) return { status: 404 }
    if (state.failure || (state.modeFailure && mode === 'nsfw') || (state.pageFailure && page > 1)) return { status: 503 }
    if (url.pathname === collectionPath) {
      expect(Object.keys(Object.fromEntries(url.searchParams)).sort()).toEqual(['lang', 'mode', 'page', 'page_size', 'phase', 'q', 'sort'])
      expect(url.searchParams.get('page_size')).toBe('24')
      const data = collectionIndex(media, lang, mode, page)
      const q = url.searchParams.get('q')!, phase = url.searchParams.get('phase')!, sort = url.searchParams.get('sort')!
      expect(['all', 'released', 'upcoming', 'mixed']).toContain(phase)
      expect(['published_desc', 'count_desc', 'count_asc', 'name_asc', 'name_desc']).toContain(sort)
      // Explicit response plans exercise client ownership, not a second SQL filter.
      if (q === '不存在' || (q === '隐藏作品' && mode === 'sfw')) {
        data.items = []; data.total = 0; data.has_more = false
      } else if (q || phase !== 'all' || sort !== 'published_desc') {
        data.items = data.items.map(item => ({ ...item, name: `${item.name} · ${q || phase} · ${sort}` }))
      }
      return { data }
    }
    expect(Object.keys(Object.fromEntries(url.searchParams)).sort()).toEqual(['lang', 'mode'])
    return { data: collectionDetail(media, lang, mode, url.pathname.split('/').at(-1)) }
  }, {}, { expectedStatuses: [404, 503] },
)
export async function changeCollectionMode(page: Page, mode: 'sfw' | 'nsfw') {
  await page.evaluate(value => {
    localStorage.setItem('mode', value)
    window.dispatchEvent(new CustomEvent('mode-change', { detail: { mode: value, displayMode: value } }))
  }, mode)
}
export { openRuntime, settleRuntime, expect }
