import { beforeEach, expect, it, vi } from 'vitest'
import { defineComponent, h, nextTick, ref } from 'vue'
import { flushPromises } from '@vue/test-utils'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { useSiteObservationHistory } from '../../app/composables/useSiteObservationHistory'
const { api } = vi.hoisted(() => ({ api: vi.fn() }))
vi.mock('../../app/composables/useApi', () => ({ useApi: () => api }))
beforeEach(() => { api.mockReset() })
const response = (target = 'a.example', count = 100) => ({ target, protocol: 'ping', items: Array.from({ length: count }, (_, index) => ({
  target, protocol: 'ping', status: 'success', observed_at: '2026-09-26T12:00:00Z', payload: { avg_rtt_ms: index, loss_rate: 0 },
})) })
async function mountHistory(activeValue = false) {
  const siteId = ref('41'), target = ref('a.example'), active = ref(activeValue)
  let history!: ReturnType<typeof useSiteObservationHistory>
  const wrapper = await mountSuspended(defineComponent({ setup() {
    history = useSiteObservationHistory({ siteId, target, active }); return () => h('div')
  } }))
  return { wrapper, siteId, target, active, history }
}
it('loads only after activation, slices locally and caches across views and Targets', async () => {
  api.mockImplementation((path: string) => Promise.resolve(response(path.includes('b.example') ? 'b.example' : 'a.example')))
  const { wrapper, active, target, history } = await mountHistory()
  try {
    expect(api).not.toHaveBeenCalled()
    active.value = true; await nextTick(); await flushPromises()
    expect(api).toHaveBeenCalledTimes(1)
    expect(api.mock.calls[0]).toEqual(['/nav/sites/41/targets/a.example/observations', { query: { protocol: 'ping', limit: 100, payload_mode: 'preview' }, retry: 0 }])
    expect(history.state.value).toBe('ready'); expect(history.rows.value).toHaveLength(20)
    for (const sample of [60, 100] as const) { history.selectSample(sample); expect(history.rows.value).toHaveLength(sample) }
    active.value = false; await nextTick(); active.value = true; await nextTick()
    expect(api).toHaveBeenCalledTimes(1)
    target.value = 'b.example'; await nextTick(); await flushPromises()
    expect(history.sample.value).toBe(20); expect(api).toHaveBeenCalledTimes(2)
    target.value = 'a.example'; await nextTick(); await flushPromises()
    expect(history.state.value).toBe('ready'); expect(api).toHaveBeenCalledTimes(2)
  } finally { wrapper.unmount() }
})
it('classifies pending then empty and caches empty without blank/implicit refetch', async () => {
  let release!: (value: unknown) => void
  api.mockReturnValue(new Promise(resolve => { release = resolve }))
  const { wrapper, active, history } = await mountHistory(true)
  try {
    expect(history.sample.value).toBe(20); expect(history.state.value).toBe('loading')
    expect(api).toHaveBeenCalledTimes(1)
    release(response('a.example', 0)); await flushPromises()
    expect(history.state.value).toBe('empty')
    active.value = false; await nextTick(); active.value = true; await nextTick()
    expect(api).toHaveBeenCalledTimes(1)
  } finally { wrapper.unmount() }
})
it('isolates failure, retries explicitly and never treats it as empty', async () => {
  api.mockRejectedValueOnce(new Error('503')).mockResolvedValue(response())
  const { wrapper, history } = await mountHistory(true)
  try {
    await flushPromises(); expect(history.state.value).toBe('unavailable')
    await history.retry(); expect(history.state.value).toBe('ready'); expect(api).toHaveBeenCalledTimes(2)
  } finally { wrapper.unmount() }
})
it('an older Target response can populate its cache but cannot replace the selected Target', async () => {
  let release!: (value: unknown) => void
  api.mockReturnValueOnce(new Promise(resolve => { release = resolve })).mockResolvedValueOnce(response('b.example', 3))
  const { wrapper, target, history } = await mountHistory(true)
  try {
    target.value = 'b.example'; await nextTick(); await flushPromises()
    expect(history.total.value).toBe(3)
    release(response('a.example', 50)); await flushPromises()
    expect(history.total.value).toBe(3)
    target.value = 'a.example'; await nextTick(); await flushPromises()
    expect(history.total.value).toBe(50); expect(api).toHaveBeenCalledTimes(2)
  } finally { wrapper.unmount() }
})
it('separates Site IDs and rejects a foreign response identity', async () => {
  api.mockResolvedValueOnce(response()).mockResolvedValueOnce(response('foreign'))
  const { wrapper, siteId, history } = await mountHistory(true)
  try {
    await flushPromises(); expect(history.state.value).toBe('ready')
    siteId.value = '42'; await nextTick(); await flushPromises()
    expect(history.state.value).toBe('unavailable'); expect(api).toHaveBeenCalledTimes(2)
  } finally { wrapper.unmount() }
})
