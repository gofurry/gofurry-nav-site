import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { defineComponent, h, ref, nextTick } from 'vue'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { useGameShowcaseTracking } from '../../app/composables/useGameShowcaseTracking'
import { submitGameShowcaseEvent } from '../../app/services/game'
import type { GameShowcaseItem } from '../../app/types/game'

vi.mock('../../app/services/game', () => ({ submitGameShowcaseEvent: vi.fn().mockResolvedValue(undefined) }))
const send = vi.mocked(submitGameShowcaseEvent)
const id = '47fde791-293c-4147-8eba-d41cfc03adab'
const item = (key: string) => ({ key, tracking_token: `signed-${key}` }) as GameShowcaseItem
let intersect: IntersectionObserverCallback
const disconnect = vi.fn()
let visibility = 'visible'

beforeEach(() => {
  visibility = 'visible'
  vi.spyOn(document, 'visibilityState', 'get').mockImplementation(() => visibility as DocumentVisibilityState)
  vi.stubGlobal('IntersectionObserver', class {
    constructor(callback: IntersectionObserverCallback) { intersect = callback }
    observe() {}
    disconnect = disconnect
  })
  vi.spyOn(crypto, 'randomUUID').mockReturnValue(id)
})
afterEach(() => { vi.useRealTimers(); vi.restoreAllMocks(); vi.unstubAllGlobals() })

async function mountTracking() {
  vi.useRealTimers()
  const active = ref<GameShowcaseItem | undefined>(item('A')), snapshot = ref('snapshot-a')
  let tracking!: ReturnType<typeof useGameShowcaseTracking>
  const root = ref<HTMLElement | null>(null)
  const wrapper = await mountSuspended(defineComponent({ setup() {
    tracking = useGameShowcaseTracking(root, snapshot, active)
    return () => h('section', { ref: root })
  } }))
  vi.useFakeTimers()
  const ratios = (values: number[]) => intersect(values.map(value => ({ target: root.value!, isIntersecting: value > 0, intersectionRatio: value,
    boundingClientRect: new DOMRectReadOnly(), intersectionRect: new DOMRectReadOnly(), rootBounds: null, time: 0 })), {} as IntersectionObserver)
  const ratio = (value: number) => ratios([value])
  const tab = (state: string) => { visibility = state; document.dispatchEvent(new Event('visibilitychange')) }
  return { wrapper, tracking, active, snapshot, ratio, ratios, tab }
}

it('requires continuous one second at least half visible, dedupes revisits and resets timer on active/snapshot changes', async () => {
  const view = await mountTracking()
  try {
    view.ratio(.49); await vi.advanceTimersByTimeAsync(2000); expect(send).not.toHaveBeenCalled()
    expect(sessionStorage.getItem('gofurry-showcase-session')).toBeNull()
    view.ratio(.5); await vi.advanceTimersByTimeAsync(999); expect(send).not.toHaveBeenCalled()
    view.ratio(.49); view.ratio(.5); await vi.advanceTimersByTimeAsync(1000)
    expect(send).toHaveBeenCalledExactlyOnceWith({ tracking_token: 'signed-A', session_id: id, event: 'impression' })
    view.active.value = item('B'); await nextTick(); await vi.advanceTimersByTimeAsync(500)
    view.active.value = item('C'); await nextTick(); await vi.advanceTimersByTimeAsync(1000)
    expect(send.mock.calls.map(([event]) => event.tracking_token)).toEqual(['signed-A', 'signed-C'])
    view.active.value = item('A'); await vi.advanceTimersByTimeAsync(2000); expect(send).toHaveBeenCalledTimes(2)
    view.snapshot.value = 'snapshot-b'; await vi.advanceTimersByTimeAsync(1000); expect(send).toHaveBeenCalledTimes(3)
    expect(crypto.randomUUID).toHaveBeenCalledTimes(1)
  } finally { view.wrapper.unmount() }
})

it('restarts the full interval when an observer batch contains exit and re-entry', async () => {
  const view = await mountTracking()
  try {
    view.ratio(1); await vi.advanceTimersByTimeAsync(500)
    view.ratios([.49, .8]); await vi.advanceTimersByTimeAsync(999)
    expect(send).not.toHaveBeenCalled()
    await vi.advanceTimersByTimeAsync(1); expect(send).toHaveBeenCalledTimes(1)
  } finally { view.wrapper.unmount() }
})

it('cancels on hidden tab and unmount, restarting the full interval when visible', async () => {
  const view = await mountTracking()
  view.ratio(1); await vi.advanceTimersByTimeAsync(500)
  view.tab('hidden'); await vi.advanceTimersByTimeAsync(5000); expect(send).not.toHaveBeenCalled()
  view.tab('visible'); await vi.advanceTimersByTimeAsync(999); expect(send).not.toHaveBeenCalled()
  await vi.advanceTimersByTimeAsync(1); expect(send).toHaveBeenCalledTimes(1)
  view.active.value = item('B'); await vi.advanceTimersByTimeAsync(500)
  view.wrapper.unmount(); await vi.advanceTimersByTimeAsync(2000)
  view.tab('visible'); view.ratio(1); await vi.advanceTimersByTimeAsync(2000)
  expect(send).toHaveBeenCalledTimes(1); expect(disconnect).toHaveBeenCalled()
})

it('clicks send immediately with each source and no fabricated impression', async () => {
  sessionStorage.setItem('gofurry-showcase-session', id)
  const view = await mountTracking()
  try {
    for (const source of ['artwork', 'title', 'primary', 'secondary'] as const) view.tracking.click(source)
    expect(send.mock.calls.map(([event]) => event)).toEqual(['artwork', 'title', 'primary', 'secondary'].map(source => ({
      tracking_token: 'signed-A', session_id: id, event: 'click', source,
    })))
    expect(crypto.randomUUID).not.toHaveBeenCalled()
  } finally { view.wrapper.unmount() }
})

it('replaces invalid stored identity; blocked storage uses only one page-memory UUID', async () => {
  sessionStorage.setItem('gofurry-showcase-session', 'not-v4')
  const first = await mountTracking()
  try { first.tracking.click('title'); expect(sessionStorage.getItem('gofurry-showcase-session')).toBe(id) }
  finally { first.wrapper.unmount() }
  vi.stubGlobal('sessionStorage', {
    getItem() { throw new Error('blocked') },
    setItem() { throw new Error('blocked') },
    clear() {},
  })
  const second = await mountTracking()
  try {
    second.tracking.click('title'); second.tracking.click('primary')
    expect(send).toHaveBeenCalledTimes(3)
    expect(crypto.randomUUID).toHaveBeenCalledTimes(2)
  } finally { second.wrapper.unmount() }
})

it('disables side effects without crypto.randomUUID, even with stored identity', async () => {
  sessionStorage.setItem('gofurry-showcase-session', id)
  const view = await mountTracking()
  vi.stubGlobal('crypto', {})
  try {
    view.ratio(1); await vi.advanceTimersByTimeAsync(1500); view.tracking.click('primary')
    expect(send).not.toHaveBeenCalled()
  } finally { view.wrapper.unmount() }
})

it('keeps pending media out of impressions while attributing artwork clicks to the displayed item', async () => {
  const view = await mountTracking()
  try {
    view.ratio(1); await vi.advanceTimersByTimeAsync(500)
    view.active.value = undefined
    await vi.advanceTimersByTimeAsync(1500)
    expect(send).not.toHaveBeenCalled()
    view.tracking.click('artwork', item('A'))
    view.tracking.click('primary', item('B'))
    expect(send.mock.calls.map(([event]) => [event.event, event.tracking_token])).toEqual([['click', 'signed-A'], ['click', 'signed-B']])
    view.active.value = item('B')
    await vi.advanceTimersByTimeAsync(999); expect(send).toHaveBeenCalledTimes(2)
    await vi.advanceTimersByTimeAsync(1)
    expect(send).toHaveBeenLastCalledWith({ tracking_token: 'signed-B', session_id: id, event: 'impression' })
  } finally { view.wrapper.unmount() }
})
