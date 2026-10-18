import { beforeEach, expect, it, vi } from 'vitest'
import { defineComponent, h, nextTick, ref } from 'vue'
import { flushPromises } from '@vue/test-utils'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { clearNuxtData } from '#app'
import { useSiteInsights } from '../../app/composables/useSiteInsights'
const { getInsights } = vi.hoisted(() => ({ getInsights: vi.fn() }))
vi.mock('../../app/services/nav', () => ({ getSiteInsights: getInsights }))
const response = (id = '41', empty = false) => ({ site: { id: Number(id), name: 'Site ' + id }, capabilities: empty ? [] : [{ key: 'ipv6', state: 'supported', as_of: '2026-09-27', ecosystem: { value: .5, coverage: .8 } }], recent_changes: [] })
beforeEach(() => { clearNuxtData(key => key.startsWith('site-insights:')); getInsights.mockReset() })
async function mountInsights() {
  const siteId = ref('41'), domain = ref('a.example')
  let insights!: Awaited<ReturnType<typeof useSiteInsights>>
  const wrapper = await mountSuspended(defineComponent({ async setup() { insights = await useSiteInsights(siteId); return () => h('div', domain.value) } }))
  return { wrapper, siteId, domain, insights }
}
it('shares one Site result, ignores Target, and retries only the failed Site slice', async () => {
  getInsights.mockRejectedValueOnce(new Error('503')).mockResolvedValue(response())
  const { wrapper, insights, domain } = await mountInsights()
  try {
    expect(insights.state.value).toBe('unavailable'); expect(insights.data.value).toBeNull()
    domain.value = 'b.example'; await nextTick(); expect(getInsights).toHaveBeenCalledTimes(1)
    await insights.retry(); expect(insights.state.value).toBe('ready')
    expect(insights.data.value?.site.id).toBe(41); expect(getInsights.mock.calls).toEqual([['41'], ['41']])
  } finally { wrapper.unmount() }
})
it('keeps success-empty distinct from failure and validates Site identity', async () => {
  getInsights.mockResolvedValueOnce(response('41', true)).mockResolvedValueOnce(response('foreign'))
  const { wrapper, insights } = await mountInsights()
  try {
    expect(insights.state.value).toBe('empty'); await insights.retry(); expect(insights.state.value).toBe('unavailable')
  } finally { wrapper.unmount() }
})
it('a pending retry for the old Site cannot overwrite the new Site result', async () => {
  let release!: (value: unknown) => void
  getInsights.mockResolvedValueOnce(response()).mockReturnValueOnce(new Promise(resolve => { release = resolve })).mockResolvedValueOnce(response('42'))
  const { wrapper, siteId, insights } = await mountInsights()
  try {
    const retry = insights.retry(); await nextTick(); expect(insights.retrying.value).toBe(true)
    siteId.value = '42'; await nextTick(); await flushPromises()
    expect(insights.data.value?.site.id).toBe(42)
    release(response('41')); await retry; await flushPromises()
    expect(insights.data.value?.site.id).toBe(42); expect(insights.state.value).toBe('ready')
    expect(getInsights.mock.calls).toEqual([['41'], ['41'], ['42']])
  } finally { wrapper.unmount() }
})
