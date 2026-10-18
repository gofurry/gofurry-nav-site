import { beforeEach, expect, it, vi } from 'vitest'
import { clearNuxtData } from '#app'
import { defineComponent, h, nextTick, ref } from 'vue'
import { flushPromises } from '@vue/test-utils'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { useUpdateIndex } from '../../app/composables/useUpdateIndex'
import type { NavUpdatesResponse } from '../../app/types/nav'

const { getUpdates } = vi.hoisted(() => ({ getUpdates: vi.fn() }))
vi.mock('../../app/services/nav', () => ({ getNavUpdates: getUpdates }))
function response(page = 1, lang = 'zh', ids = page === 1 ? [4, 3] : [2, 1]): NavUpdatesResponse {
  return { schema_version: 1, state: ids.length ? 'ready' : 'empty', generated_at: '2026-09-28T08:00:00Z',
    page, page_size: 21, total: 4, has_more: page === 1,
    items: ids.map(id => ({ id, title: `${lang} ${id}`, summary: '', version: null, commit_sha: null, published_at: '2026-09-28T08:00:00Z' })) }
}
beforeEach(() => { clearNuxtData(); getUpdates.mockReset(); getUpdates.mockImplementation((lang, { page }) => Promise.resolve(response(page, lang))) })
async function mountIndex() {
  const lang = ref<'zh' | 'en'>('zh')
  let index!: Awaited<ReturnType<typeof useUpdateIndex>>
  const wrapper = await mountSuspended(defineComponent({ async setup() { index = await useUpdateIndex(lang); return () => h('div') } }))
  return { wrapper, lang, index }
}
it('requests a bounded first page, then only explicit pages, and deduplicates overlapping releases', async () => {
  const { wrapper, index } = await mountIndex()
  try {
    expect(getUpdates).toHaveBeenCalledExactlyOnceWith('zh', { page: 1, page_size: 21 })
    expect(index.total.value).toBe(4); expect(index.items.value.map(item => item.id)).toEqual([4, 3])
    getUpdates.mockResolvedValueOnce(response(2, 'zh', [3, 2, 1]))
    await index.loadMore()
    expect(index.items.value.map(item => item.id)).toEqual([4, 3, 2, 1])
    expect(index.hasMore.value).toBe(false)
    await index.loadMore(); expect(getUpdates).toHaveBeenCalledTimes(2)
  } finally { wrapper.unmount() }
})
it('blocks concurrent clicks, retains visible content after failure, and retries the same page', async () => {
  const { wrapper, index } = await mountIndex()
  try {
    let reject!: (error: Error) => void
    getUpdates.mockReturnValueOnce(new Promise((_, fail) => { reject = fail }))
    const pending = index.loadMore(); await index.loadMore()
    expect(index.loadingMore.value).toBe(true); expect(getUpdates).toHaveBeenCalledTimes(2)
    reject(new Error('offline')); await pending
    expect(index.moreError.value).toBe(true); expect(index.state.value).toBe('ready')
    expect(index.items.value).toHaveLength(2)
    await index.loadMore()
    expect(getUpdates.mock.calls.slice(1).map(call => call[1])).toEqual([{ page: 2, page_size: 21 }, { page: 2, page_size: 21 }])
    expect(index.moreError.value).toBe(false); expect(index.items.value).toHaveLength(4)
  } finally { wrapper.unmount() }
})
it('language changes reset appended pages and reject an old in-flight language response', async () => {
  const { wrapper, lang, index } = await mountIndex()
  try {
    let release!: (value: NavUpdatesResponse) => void
    getUpdates.mockReturnValueOnce(new Promise(resolve => { release = resolve }))
    const pending = index.loadMore()
    const signal = getUpdates.mock.calls.at(-1)![2] as AbortSignal
    lang.value = 'en'; await nextTick(); await flushPromises()
    expect(signal.aborted).toBe(true)
    expect(index.items.value.map(item => item.title)).toEqual(['en 4', 'en 3'])
    release(response(2)); await pending
    expect(index.items.value.map(item => item.title)).toEqual(['en 4', 'en 3'])
    await index.loadMore(); expect(index.items.value.map(item => item.title)).toEqual(['en 4', 'en 3', 'en 2', 'en 1'])
  } finally { wrapper.unmount() }
})
it('unmount cancels additional pages without publishing a late result', async () => {
  const { wrapper, index } = await mountIndex()
  let release!: (value: NavUpdatesResponse) => void
  getUpdates.mockReturnValueOnce(new Promise(resolve => { release = resolve }))
  const pending = index.loadMore(), signal = getUpdates.mock.calls.at(-1)![2] as AbortSignal
  wrapper.unmount(); expect(signal.aborted).toBe(true)
  // Nuxt releases its first-page async-data cache on the next tick.
  await flushPromises()
  const afterUnmount = index.items.value.map(item => item.id)
  release(response(2)); await pending
  expect(index.items.value.map(item => item.id)).toEqual(afterUnmount)
})
