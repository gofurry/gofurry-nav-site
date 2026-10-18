import { beforeEach, expect, it, vi } from 'vitest'
import { defineComponent, h, nextTick, ref } from 'vue'
import { flushPromises } from '@vue/test-utils'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { useSiteInsightTrend } from '../../app/composables/useSiteInsightTrend'
import type { InsightRange, NavInsightMetricKey } from '../../app/types/insights'
const { getTrend } = vi.hoisted(() => ({ getTrend: vi.fn() }))
vi.mock('../../app/services/nav', () => ({ getNavInsightsTrend: getTrend }))
const response = (key = 'ipv6', requested_range = '30d', count = 3) => ({ key, requested_range, points: Array.from({ length: count }, (_, index) => ({ date: '2026-09-2' + index, value: .5, coverage: .8 })) })
beforeEach(() => { getTrend.mockReset() })
async function mountTrend(enabled = false) {
  const active = ref(enabled), metric = ref<NavInsightMetricKey>('ipv6'), range = ref<InsightRange>('30d'), domain = ref('a.example')
  let trend!: ReturnType<typeof useSiteInsightTrend>
  const wrapper = await mountSuspended(defineComponent({ setup() { trend = useSiteInsightTrend({ active, metric, range }); return () => h('div', domain.value) } }))
  return { wrapper, active, metric, range, domain, trend }
}
it('loads only after activation, reuses metric/range cache and ignores domain changes', async () => {
  getTrend.mockImplementation((key, range) => Promise.resolve(response(key, range)))
  const { wrapper, active, metric, range, domain, trend } = await mountTrend()
  try {
    expect(getTrend).not.toHaveBeenCalled(); expect(trend.current.value.state).toBe('loading')
    active.value = true; await nextTick(); await flushPromises()
    expect(getTrend).toHaveBeenCalledExactlyOnceWith('ipv6', '30d'); expect(trend.current.value.state).toBe('ready')
    domain.value = 'b.example'; await nextTick(); expect(getTrend).toHaveBeenCalledTimes(1)
    metric.value = 'tls13'; await nextTick(); await flushPromises(); expect(getTrend).toHaveBeenLastCalledWith('tls13', '30d')
    metric.value = 'ipv6'; await nextTick(); await flushPromises(); expect(getTrend).toHaveBeenCalledTimes(2)
    range.value = '90d'; await nextTick(); await flushPromises(); expect(getTrend).toHaveBeenLastCalledWith('ipv6', '90d')
    range.value = 'all'; await nextTick(); await flushPromises(); expect(getTrend).toHaveBeenLastCalledWith('ipv6', 'all')
    active.value = false; await nextTick(); active.value = true; await nextTick(); expect(getTrend).toHaveBeenCalledTimes(4)
  } finally { wrapper.unmount() }
})
it('classifies an active unresolved identity as loading and caches successful empty', async () => {
  let release!: (value: unknown) => void
  getTrend.mockReturnValue(new Promise(resolve => { release = resolve }))
  const { wrapper, active, trend } = await mountTrend(true)
  try {
    expect(trend.current.value).toEqual({ state: 'loading', points: [] })
    await trend.retry(); expect(getTrend).toHaveBeenCalledTimes(1)
    release(response('ipv6', '30d', 0)); await flushPromises(); expect(trend.current.value.state).toBe('empty')
    active.value = false; await nextTick(); active.value = true; await nextTick(); expect(getTrend).toHaveBeenCalledTimes(1)
  } finally { wrapper.unmount() }
})
it('isolates failure until explicit retry, without treating it as success-empty', async () => {
  getTrend.mockRejectedValueOnce(new Error('503')).mockResolvedValue(response())
  const { wrapper, active, trend } = await mountTrend(true)
  try {
    await flushPromises(); expect(trend.current.value.state).toBe('unavailable')
    active.value = false; await nextTick(); active.value = true; await nextTick(); expect(getTrend).toHaveBeenCalledTimes(1)
    await trend.retry(); expect(trend.current.value.state).toBe('ready'); expect(getTrend).toHaveBeenCalledTimes(2)
  } finally { wrapper.unmount() }
})
it('late responses populate only their own cache, never the currently selected identity', async () => {
  let release!: (value: unknown) => void
  getTrend.mockReturnValueOnce(new Promise(resolve => { release = resolve })).mockResolvedValueOnce(response('tls13', '90d', 2))
  const { wrapper, metric, range, trend } = await mountTrend(true)
  try {
    metric.value = 'tls13'; range.value = '90d'; await nextTick(); await flushPromises()
    expect(trend.current.value.points).toHaveLength(2)
    release(response('ipv6', '30d', 7)); await flushPromises(); expect(trend.current.value.points).toHaveLength(2)
    metric.value = 'ipv6'; range.value = '30d'; await nextTick(); await flushPromises()
    expect(trend.current.value.points).toHaveLength(7); expect(getTrend).toHaveBeenCalledTimes(2)
  } finally { wrapper.unmount() }
})
it('rejects a trend response carrying a different identity', async () => {
  getTrend.mockResolvedValue(response('tls13'))
  const { wrapper, trend } = await mountTrend(true)
  try { await flushPromises(); expect(trend.current.value.state).toBe('unavailable') } finally { wrapper.unmount() }
})
it('an unmounted request cannot publish into a later page session', async () => {
  let release!: (value: unknown) => void
  getTrend.mockReturnValue(new Promise(resolve => { release = resolve }))
  const { wrapper, trend } = await mountTrend(true)
  wrapper.unmount(); release(response()); await flushPromises()
  expect(trend.current.value.state).toBe('loading')
})
