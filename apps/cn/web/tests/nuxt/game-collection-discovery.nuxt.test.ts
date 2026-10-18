import { afterEach, expect, it, vi } from 'vitest'
import { defineComponent, h, nextTick } from 'vue'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { useGameCollectionDiscovery } from '../../app/composables/useGameCollectionDiscovery'
import { collectionIndex } from '../browser/fixtures/game-collections-data'
import type { GameCollectionIndex } from '../../app/types/game'

afterEach(() => { vi.useRealTimers(); localStorage.removeItem('mode') })
const initial = () => collectionIndex('local', 'zh', 'sfw', 1)
function gate() { let resolve!: (value: GameCollectionIndex) => void; const promise = new Promise<GameCollectionIndex>(r => { resolve = r }); return { promise, resolve } }
const change = (mode: string) => window.dispatchEvent(new CustomEvent('mode-change', { detail: { mode } }))
async function setup(load = vi.fn().mockResolvedValue(initial())) {
  let state!: ReturnType<typeof useGameCollectionDiscovery>
  const wrapper = await mountSuspended(defineComponent({ setup() {
    state = useGameCollectionDiscovery({ initial: initial(), load })
    return () => h('input', { value: state.inputValue.value, onInput: state.onInput,
      onCompositionstart: state.onCompositionStart, onCompositionend: state.onCompositionEnd })
  } }))
  return { wrapper, state, load, input: wrapper.find('input') }
}
it('commits final IME text once after 350ms, preserving SSR cards and canceling on disposal', async () => {
  const { wrapper, state, load, input } = await setup()
  vi.useFakeTimers()
  expect(load).not.toHaveBeenCalled()
  await input.trigger('compositionstart')
  await input.setValue('sen'); await vi.advanceTimersByTimeAsync(700)
  await input.setValue('森林'); await vi.advanceTimersByTimeAsync(700)
  expect(load).not.toHaveBeenCalled(); expect(state.criteria.value.q).toBe('')
  await input.trigger('compositionend'); await input.trigger('input')
  await vi.advanceTimersByTimeAsync(349); expect(load).not.toHaveBeenCalled()
  expect(state.snapshot.value).toEqual(initial())
  await vi.advanceTimersByTimeAsync(1)
  expect(load).toHaveBeenCalledExactlyOnceWith('sfw', 1, { q: '森林', phase: 'all', sort: 'published_desc' })
  await input.setValue('other'); wrapper.unmount(); await vi.runAllTimersAsync()
  change('nsfw'); expect(load).toHaveBeenCalledTimes(1)
})
it('shares latest generation across keyword, filter, mode and stale pagination', async () => {
  const older = gate(), latest = gate(), more = gate()
  const load = vi.fn().mockReturnValueOnce(older.promise).mockReturnValueOnce(latest.promise).mockReturnValueOnce(more.promise).mockResolvedValue(initial())
  const { wrapper, state, input } = await setup(load)
  vi.useFakeTimers()
  await input.setValue('森林'); await vi.advanceTimersByTimeAsync(350)
  state.applyFilter({ phase: 'mixed', sort: 'count_desc' })
  latest.resolve(initial()); await nextTick(); await nextTick()
  older.resolve({ ...initial(), items: [] }); await nextTick(); await nextTick()
  expect(state.snapshot.value?.items).toHaveLength(6)
  void state.loadMore()
  change('nsfw'); await nextTick(); await nextTick()
  more.resolve(collectionIndex('local','zh','sfw',2)); await nextTick(); await nextTick()
  expect(state.snapshot.value?.items).toHaveLength(6)
  expect(load.mock.calls).toEqual([
    ['sfw',1,{q:'森林',phase:'all',sort:'published_desc'}],
    ['sfw',1,{q:'森林',phase:'mixed',sort:'count_desc'}],
    ['sfw',2,{q:'森林',phase:'mixed',sort:'count_desc'}],
    ['nsfw',1,{q:'森林',phase:'mixed',sort:'count_desc'}],
  ])
  wrapper.unmount()
})
it('retains ready cards on failure, retries and dedupes pagination', async () => {
  const load = vi.fn().mockRejectedValueOnce(new Error('503')).mockResolvedValueOnce(initial()).mockRejectedValueOnce(new Error('503')).mockResolvedValueOnce(collectionIndex('local','zh','sfw',2))
  const { wrapper, state } = await setup(load)
  state.applyFilter({ phase:'released', sort:'count_asc' }); await nextTick(); await nextTick()
  expect(state.error.value).toBeTruthy(); expect(state.snapshot.value?.items).toHaveLength(6)
  await state.refresh(); expect(state.error.value).toBeNull()
  await state.loadMore(); expect(state.loadError.value).toBe(true); expect(state.snapshot.value?.items).toHaveLength(6)
  await state.loadMore(); expect(state.snapshot.value?.items).toHaveLength(8)
  wrapper.unmount()
})
