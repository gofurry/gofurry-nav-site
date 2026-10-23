import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { createThemeReveal } from '../../app/utils/themeReveal'

function deferred() {
  let resolve!: () => void, reject!: (error?: unknown) => void
  const promise = new Promise<void>((done, fail) => { resolve = done; reject = fail })
  return { promise, resolve, reject }
}
const flush = async () => { for (let i = 0; i < 12; i++) await Promise.resolve() }

function environment() {
  const media = Object.assign(new EventTarget(), { matches: false })
  const ready = deferred(), updated = deferred(), finished = deferred(), animated = deferred()
  const cancelAnimation = vi.fn(() => animated.reject(new Error('cancelled')))
  const animate = vi.fn(() => ({ finished: animated.promise, cancel: cancelAnimation }))
  const attributes = new Set<string>()
  const root = { animate, setAttribute: (name: string) => attributes.add(name), removeAttribute: (name: string) => attributes.delete(name) }
  const color = { transitionProperty: 'background-color', finish: vi.fn() }
  const movement = { transitionProperty: 'transform', finish: vi.fn() }
  const fade = { transitionProperty: 'opacity', finish: vi.fn() }
  const entrance = { finish: vi.fn() }
  let callback!: () => Promise<void>
  const transition = { ready: ready.promise, updateCallbackDone: updated.promise, finished: finished.promise, skipTransition: vi.fn() }
  const start = vi.fn((update: () => Promise<void>) => { callback = update; return transition })
  const doc = Object.assign(new EventTarget(), {
    documentElement: root, hidden: false, startViewTransition: start,
    getAnimations: vi.fn(() => [color, movement, fade, entrance]),
  })
  const win = Object.assign(new EventTarget(), { matchMedia: () => media, CSS: { supports: () => true } })
  const button = { isConnected: true, matches: () => false,
    getBoundingClientRect: () => ({ left: 1200, top: 20, width: 40, height: 40 }) } as unknown as HTMLElement
  vi.stubGlobal('document', doc); vi.stubGlobal('window', win)
  vi.stubGlobal('innerWidth', 1440); vi.stubGlobal('innerHeight', 900)
  const update = vi.fn(async () => {})
  const controller = createThemeReveal()
  return { media, ready, updated, finished, animated, animate, attributes, color, movement, fade, entrance,
    transition, start, doc, win, button, update, controller, callback: () => callback() }
}
let env: ReturnType<typeof environment>
beforeEach(() => { env = environment() })
afterEach(async () => { env.controller.dispose(); await flush(); vi.unstubAllGlobals() })

it('commits once, captures final colors, and reveals the new snapshot from the actual button to the farthest corner', async () => {
  const { controller, button, update, attributes, start, ready, finished, animate } = env
  controller.toggle(button, update); controller.toggle(button, update)
  expect(start).toHaveBeenCalledTimes(1); expect(update).not.toHaveBeenCalled()
  expect(attributes.has('data-gf-theme-reveal')).toBe(true)
  await env.callback(); ready.resolve(); await flush()
  expect(update).toHaveBeenCalledTimes(1)
  expect(env.color.finish).toHaveBeenCalledOnce()
  for (const animation of [env.movement, env.fade, env.entrance]) expect(animation.finish).not.toHaveBeenCalled()
  const radius = Math.hypot(1220, 860)
  expect(animate).toHaveBeenCalledWith({ clipPath: ['circle(0px at 1220px 40px)', `circle(${radius}px at 1220px 40px)`] }, {
    duration: 420, easing: 'cubic-bezier(0.22, 1, 0.36, 1)', pseudoElement: '::view-transition-new(root)', fill: 'both',
  })
  env.animated.resolve(); finished.resolve(); await flush()
  expect(attributes.size).toBe(0)
  controller.toggle(button, update); expect(start).toHaveBeenCalledTimes(2)
})

it.each(['unsupported', 'no-waapi', 'no-clip', 'reduced', 'hidden', 'disconnected', 'invalid-position', 'missing-button'])(
  'switches directly without snapshot state when %s', async reason => {
    const { controller, button, update, doc, media, win, attributes } = env
    if (reason === 'unsupported') Object.assign(doc, { startViewTransition: undefined })
    if (reason === 'no-waapi') Object.assign(doc.documentElement, { animate: undefined })
    if (reason === 'no-clip') win.CSS.supports = () => false
    if (reason === 'reduced') media.matches = true
    if (reason === 'hidden') doc.hidden = true
    if (reason === 'disconnected') Object.assign(button, { isConnected: false })
    if (reason === 'invalid-position') button.getBoundingClientRect = () => ({ left: NaN } as DOMRect)
    controller.toggle(reason === 'missing-button' ? null : button, update)
    await flush()
    expect(update).toHaveBeenCalledOnce(); expect(env.start).not.toHaveBeenCalled(); expect(attributes.size).toBe(0)
    expect(env.color.finish).toHaveBeenCalledTimes(reason === 'reduced' ? 1 : 0)
  },
)

it.each(['reduce', 'hidden', 'resize', 'pagehide', 'pageswap', 'route', 'unmount'])(
  'settles %s before a deferred update, ignores late ready work and releases every listener', async reason => {
    const { controller, button, update, doc, win, media, attributes, ready } = env
    const listeners = [doc, win, media].map(target => ({ add: vi.spyOn(target, 'addEventListener'), remove: vi.spyOn(target, 'removeEventListener') }))
    controller.toggle(button, update)
    if (reason === 'reduce') { media.matches = true; media.dispatchEvent(new Event('change')) }
    else if (reason === 'hidden') { doc.hidden = true; doc.dispatchEvent(new Event('visibilitychange')) }
    else if (reason === 'unmount') controller.dispose()
    else if (reason === 'route') controller.cancel()
    else win.dispatchEvent(new Event(reason))
    await flush()
    await env.callback(); ready.resolve(); await flush()
    expect(update).toHaveBeenCalledOnce(); expect(env.animate).not.toHaveBeenCalled()
    expect(env.transition.skipTransition).toHaveBeenCalledOnce(); expect(attributes.size).toBe(0)
    for (const { add, remove } of listeners) expect(remove.mock.calls).toEqual(add.mock.calls)
  },
)

it.each(['start', 'ready', 'update', 'animate', 'animation', 'finished'])(
  'applies the target once and clears snapshot state after %s fails', async failure => {
    if (failure === 'start') env.start.mockImplementation(() => { throw new Error('snapshot unavailable') })
    env.controller.toggle(env.button, env.update)
    if (failure === 'ready') env.ready.reject(new Error('snapshot failed'))
    if (failure === 'update') env.updated.reject(new Error('update failed'))
    if (failure === 'finished') env.finished.reject(new Error('transition failed'))
    if (failure === 'animate' || failure === 'animation') {
      await env.callback()
      if (failure === 'animate') env.animate.mockImplementation(() => { throw new Error('pseudo element unsupported') })
      env.ready.resolve(); await flush()
      if (failure === 'animation') env.animated.reject(new Error('animation failed'))
    }
    await flush()
    expect(env.update).toHaveBeenCalledOnce(); expect(env.attributes.size).toBe(0)
  },
)

it('can resume after a live Reduced interruption without an old callback cancelling the new reveal', async () => {
  env.controller.toggle(env.button, env.update)
  await env.callback(); env.ready.resolve(); await flush()
  env.media.matches = true; env.media.dispatchEvent(new Event('change')); await flush()
  expect(env.attributes.size).toBe(0)
  env.media.matches = false; env.media.dispatchEvent(new Event('change'))
  const nextReady = deferred(), nextFinished = deferred()
  env.start.mockReturnValue({ ...env.transition, ready: nextReady.promise, finished: nextFinished.promise })
  env.controller.toggle(env.button, env.update)
  env.finished.resolve(); await flush()
  expect(env.attributes.has('data-gf-theme-reveal')).toBe(true)
  expect(env.start).toHaveBeenCalledTimes(2)
})
