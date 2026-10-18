import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { defineComponent, h, ref } from 'vue'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { AUTO_ADVANCE_MS, useGameShowcaseAutoplay } from '../../app/composables/useGameShowcaseAutoplay'

let intersection: IntersectionObserverCallback
let visibility = 'visible'
const disconnect = vi.fn()
let query: MediaQueryList

beforeEach(() => {
  visibility = 'visible'
  vi.spyOn(document, 'visibilityState', 'get').mockImplementation(() => visibility as DocumentVisibilityState)
  query = Object.assign(new EventTarget(), { matches: false }) as MediaQueryList
  vi.spyOn(window, 'matchMedia').mockReturnValue(query)
  vi.stubGlobal('IntersectionObserver', class {
    constructor(callback: IntersectionObserverCallback) { intersection = callback }
    observe() {}
    disconnect = disconnect
  })
})
afterEach(() => { vi.useRealTimers(); vi.restoreAllMocks(); vi.unstubAllGlobals() })

async function mountAutoplay(count = 4, ready = true) {
  vi.useRealTimers()
  const root = ref<HTMLElement | null>(null), index = ref(0), itemCount = ref(count), artworkReady = ref(ready), snapshotId = ref('snapshot-a')
  let autoplay!: ReturnType<typeof useGameShowcaseAutoplay>
  const advance = vi.fn(() => { artworkReady.value = false; index.value = (index.value + 1) % itemCount.value })
  const wrapper = await mountSuspended(defineComponent({ setup() {
    autoplay = useGameShowcaseAutoplay(root, { index, itemCount, artworkReady, snapshotId }, advance)
    return () => h('section', { ref: root }, [h('button'), h('button')])
  } }))
  vi.useFakeTimers()
  const ratios = (values: number[]) => intersection(values.map(ratio => ({ target: root.value!, isIntersecting: ratio > 0, intersectionRatio: ratio,
    boundingClientRect: new DOMRectReadOnly(), intersectionRect: new DOMRectReadOnly(), rootBounds: null, time: 0 })), {} as IntersectionObserver)
  const tab = (value: string) => { visibility = value; document.dispatchEvent(new Event('visibilitychange')) }
  const hover = (enter: boolean, pointerType = 'mouse') => root.value!.dispatchEvent(new PointerEvent(enter ? 'pointerenter' : 'pointerleave', { pointerType }))
  const focus = (inside: boolean) => root.value!.dispatchEvent(new FocusEvent(inside ? 'focusin' : 'focusout', { relatedTarget: inside ? root.value!.firstChild : document.body }))
  const reduce = (value: boolean) => { Object.assign(query, { matches: value }); query.dispatchEvent(new Event('change')) }
  return { wrapper, root, index, itemCount, artworkReady, snapshotId, autoplay, advance, ratios, tab, hover, focus, reduce }
}

it.each([0, 1])('does not run or show controls for %i items', async count => {
  const view = await mountAutoplay(count)
  try {
    view.ratios([1]); await vi.advanceTimersByTimeAsync(60000)
    expect(view.advance).not.toHaveBeenCalled(); expect(view.autoplay.showControl.value).toBe(false)
  } finally { view.wrapper.unmount() }
})

it('waits a full 6000ms only after current displayed artwork is ready', async () => {
  expect(AUTO_ADVANCE_MS).toBe(6000)
  const view = await mountAutoplay(4, false)
  try {
    view.ratios([1]); await vi.advanceTimersByTimeAsync(60000); expect(view.advance).not.toHaveBeenCalled()
    view.artworkReady.value = true
    await vi.advanceTimersByTimeAsync(5999); expect(view.advance).not.toHaveBeenCalled()
    await vi.advanceTimersByTimeAsync(1); expect(view.advance).toHaveBeenCalledTimes(1)
    await vi.advanceTimersByTimeAsync(4000); expect(view.advance).toHaveBeenCalledTimes(1)
    view.artworkReady.value = true
    await vi.advanceTimersByTimeAsync(5999); expect(view.advance).toHaveBeenCalledTimes(1)
    await vi.advanceTimersByTimeAsync(1); expect(view.advance).toHaveBeenCalledTimes(2)
  } finally { view.wrapper.unmount() }
})

it.each(['intersection', 'hidden', 'hover', 'focus', 'reduced', 'artwork'] as const)('discards elapsed time after %s interruption', async condition => {
  const view = await mountAutoplay()
  const interrupt = (blocked: boolean) => {
    if (condition === 'intersection') view.ratios([blocked ? .49 : .5])
    if (condition === 'hidden') view.tab(blocked ? 'hidden' : 'visible')
    if (condition === 'hover') view.hover(blocked)
    if (condition === 'focus') view.focus(blocked)
    if (condition === 'reduced') view.reduce(blocked)
    if (condition === 'artwork') view.artworkReady.value = !blocked
  }
  try {
    view.ratios([.5]); await vi.advanceTimersByTimeAsync(3000)
    interrupt(true); await vi.advanceTimersByTimeAsync(60000); expect(view.advance).not.toHaveBeenCalled()
    interrupt(false); await vi.advanceTimersByTimeAsync(5999); expect(view.advance).not.toHaveBeenCalled()
    await vi.advanceTimersByTimeAsync(1); expect(view.advance).toHaveBeenCalledTimes(1)
  } finally { view.wrapper.unmount() }
})

it('restarts after a batched threshold interruption and manual/index/snapshot changes', async () => {
  const view = await mountAutoplay()
  try {
    view.ratios([1]); await vi.advanceTimersByTimeAsync(5999)
    view.ratios([.49, 1]); await vi.advanceTimersByTimeAsync(5999); expect(view.advance).not.toHaveBeenCalled()
    view.index.value = 1; await vi.advanceTimersByTimeAsync(5999); expect(view.advance).not.toHaveBeenCalled()
    view.snapshotId.value = 'snapshot-b'; await vi.advanceTimersByTimeAsync(5999); expect(view.advance).not.toHaveBeenCalled()
    view.autoplay.restart(); await vi.advanceTimersByTimeAsync(5999); expect(view.advance).not.toHaveBeenCalled()
    await vi.advanceTimersByTimeAsync(1); expect(view.advance).toHaveBeenCalledTimes(1)
  } finally { view.wrapper.unmount() }
})

it('user pause survives other conditions and snapshots without storage; Play starts fresh and a new page defaults to enabled', async () => {
  const view = await mountAutoplay()
  const storage = [document.cookie, localStorage.length, sessionStorage.length]
  try {
    view.ratios([1]); await vi.advanceTimersByTimeAsync(4000); view.autoplay.togglePause()
    view.snapshotId.value = 'snapshot-b'; view.hover(true); view.hover(false); view.tab('hidden'); view.tab('visible')
    await vi.advanceTimersByTimeAsync(60000); expect(view.advance).not.toHaveBeenCalled()
    expect(view.autoplay.userPaused.value).toBe(true)
    expect([document.cookie, localStorage.length, sessionStorage.length]).toEqual(storage)
    view.autoplay.togglePause(); await vi.advanceTimersByTimeAsync(5999); expect(view.advance).not.toHaveBeenCalled()
    await vi.advanceTimersByTimeAsync(1); expect(view.advance).toHaveBeenCalledTimes(1)
  } finally { view.wrapper.unmount() }
  const nextPage = await mountAutoplay()
  try { expect(nextPage.autoplay.userPaused.value).toBe(false) }
  finally { nextPage.wrapper.unmount() }
})

it('advances from the last item to the first and gives the wrapped artwork a fresh full window', async () => {
  const view = await mountAutoplay(4)
  try {
    view.index.value = 3; view.ratios([1])
    await vi.advanceTimersByTimeAsync(5999); expect(view.index.value).toBe(3)
    await vi.advanceTimersByTimeAsync(1); expect(view.index.value).toBe(0)
    await vi.advanceTimersByTimeAsync(60000); expect(view.advance).toHaveBeenCalledTimes(1)
    view.artworkReady.value = true
    await vi.advanceTimersByTimeAsync(5999); expect(view.index.value).toBe(0)
    await vi.advanceTimersByTimeAsync(1); expect(view.index.value).toBe(1)
    view.index.value = 3; view.artworkReady.value = true; view.autoplay.togglePause()
    await vi.advanceTimersByTimeAsync(60000); expect(view.index.value).toBe(3)
    view.autoplay.togglePause()
    await vi.advanceTimersByTimeAsync(5999); expect(view.index.value).toBe(3)
    await vi.advanceTimersByTimeAsync(1); expect(view.index.value).toBe(0)
  } finally { view.wrapper.unmount() }
})

it('hides controls for reduced motion, ignores touch hover and keeps focus paused between descendants', async () => {
  const view = await mountAutoplay()
  try {
    view.reduce(true); expect(view.autoplay.showControl.value).toBe(false)
    view.ratios([1]); await vi.advanceTimersByTimeAsync(60000); expect(view.advance).not.toHaveBeenCalled()
    view.reduce(false); expect(view.autoplay.showControl.value).toBe(true)
    view.focus(true)
    view.root.value!.dispatchEvent(new FocusEvent('focusout', { relatedTarget: view.root.value!.lastChild as EventTarget }))
    await vi.advanceTimersByTimeAsync(60000); expect(view.advance).not.toHaveBeenCalled()
    view.focus(false); view.hover(true, 'touch')
    await vi.advanceTimersByTimeAsync(6000); expect(view.advance).toHaveBeenCalledTimes(1)
  } finally { view.wrapper.unmount() }
})

it('cleans its timer, observer and every DOM/media listener on unmount', async () => {
  const view = await mountAutoplay()
  const root = view.root.value!, rootRemove = vi.spyOn(root, 'removeEventListener')
  const documentRemove = vi.spyOn(document, 'removeEventListener'), mediaRemove = vi.spyOn(query, 'removeEventListener')
  view.ratios([1]); await vi.advanceTimersByTimeAsync(3000)
  view.wrapper.unmount()
  await vi.advanceTimersByTimeAsync(60000)
  expect(view.advance).not.toHaveBeenCalled(); expect(disconnect).toHaveBeenCalledTimes(1)
  expect(rootRemove.mock.calls.map(([name]) => name)).toEqual(expect.arrayContaining(['pointerenter', 'pointerleave', 'focusin', 'focusout']))
  expect(documentRemove).toHaveBeenCalledWith('visibilitychange', expect.any(Function))
  expect(mediaRemove).toHaveBeenCalledWith('change', expect.any(Function))
  expect(vi.getTimerCount()).toBe(0)
})

it('rechecks focus after reduced motion removes the focused control without a focusout', async () => {
  const view = await mountAutoplay()
  let active: Element = view.root.value!.firstElementChild!
  vi.spyOn(document, 'activeElement', 'get').mockImplementation(() => active)
  try {
    view.ratios([1]); view.focus(true); view.reduce(true)
    active = document.body
    view.reduce(false)
    await vi.advanceTimersByTimeAsync(5999); expect(view.advance).not.toHaveBeenCalled()
    await vi.advanceTimersByTimeAsync(1); expect(view.advance).toHaveBeenCalledTimes(1)
  } finally { view.wrapper.unmount() }
})
