import { afterEach, expect, it, vi } from 'vitest'
import { useNuxtApp } from '#app'
import { computed, defineComponent, h, nextTick, ref } from 'vue'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { useGameHomeOptionalRecovery } from '../../app/composables/useGameHomeOptionalRecovery'
import { getGameHomeShowcase, getGameCollectionHome } from '../../app/services/game'

afterEach(() => vi.unstubAllGlobals())
it('keeps the short SSR budget while giving one client recovery its own bounded window', async () => {
  const fetcher = vi.fn().mockResolvedValue({ code: 1, data: {} }), signal = new AbortController().signal
  vi.stubGlobal('$fetch', fetcher)
  for (const read of [getGameHomeShowcase, getGameCollectionHome]) {
    await useNuxtApp().runWithContext(() => read('zh'))
    expect(fetcher).toHaveBeenLastCalledWith(expect.any(String), expect.objectContaining({ timeout: 1000, retry: 0 }))
    await useNuxtApp().runWithContext(() => read('en', { timeout: 8000, signal }))
    expect(fetcher).toHaveBeenLastCalledWith(expect.any(String), expect.objectContaining({ timeout: 8000, retry: 0, signal }))
  }
})

function gate() { let resolve!: (value: string) => void; const promise = new Promise<string>(r => { resolve = r }); return { promise, resolve } }
async function mountRecovery(initial: string | null, load: (lang: string, signal: AbortSignal) => Promise<string>) {
  const identity = ref('zh'), data = ref({ lang: 'zh', value: initial })
  let value!: ReturnType<typeof useGameHomeOptionalRecovery<string>>
  const wrapper = await mountSuspended(defineComponent({ setup() {
    value = useGameHomeOptionalRecovery(identity, computed(() => data.value.lang === identity.value ? data.value.value : undefined), load)
    return () => h('p', value.value ?? 'absent')
  } }))
  return { wrapper, value, identity, data }
}
it.each(['ready', ''])('never recovers a successful SSR payload, including valid empty (%s)', async initial => {
  const load = vi.fn().mockResolvedValue('unexpected')
  const view = await mountRecovery(initial, load)
  expect(load).not.toHaveBeenCalled(); expect(view.value.value).toBe(initial)
  view.wrapper.unmount()
})
it('recovers once with the final locale and commits the successful payload', async () => {
  const pending = gate(), load = vi.fn().mockReturnValue(pending.promise)
  const view = await mountRecovery(null, load)
  expect(load).toHaveBeenCalledExactlyOnceWith('zh', expect.any(AbortSignal))
  expect(view.value.value).toBeNull()
  pending.resolve('recovered'); await nextTick(); await nextTick()
  expect(view.value.value).toBe('recovered')
  view.data.value = { lang: 'zh', value: null }; await nextTick()
  expect(load).toHaveBeenCalledTimes(1); view.wrapper.unmount()
})
it('silently retains absence after failure without a retry loop', async () => {
  const load = vi.fn().mockRejectedValue(new Error('503'))
  const view = await mountRecovery(null, load)
  await nextTick(); view.data.value = { lang: 'zh', value: null }; await nextTick()
  expect(load).toHaveBeenCalledTimes(1); expect(view.value.value).toBeNull()
  view.wrapper.unmount()
})
it('waits for the new language read and ignores late recovery from the old identity', async () => {
  const old = gate(), current = gate()
  const load = vi.fn().mockReturnValueOnce(old.promise).mockReturnValueOnce(current.promise)
  const view = await mountRecovery(null, load)
  view.identity.value = 'en'; await nextTick()
  expect(load.mock.calls[0]![1].aborted).toBe(true)
  expect(load).toHaveBeenCalledTimes(1)
  view.data.value = { lang: 'en', value: null }; await nextTick()
  expect(load).toHaveBeenLastCalledWith('en', expect.any(AbortSignal))
  current.resolve('EN'); await nextTick(); await nextTick()
  old.resolve('stale ZH'); await nextTick(); await nextTick()
  expect(view.value.value).toBe('EN'); view.wrapper.unmount()
})
it('aborts on unmount and ignores completion even if the loader ignores abort', async () => {
  const pending = gate(), load = vi.fn().mockReturnValue(pending.promise)
  const view = await mountRecovery(null, load)
  view.wrapper.unmount(); expect(load.mock.calls[0]![1].aborted).toBe(true)
  pending.resolve('late'); await nextTick(); await nextTick()
  expect(view.value.value).toBeNull()
})
