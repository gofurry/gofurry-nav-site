import { afterEach, expect, it, vi } from 'vitest'
import { useNuxtApp, useRuntimeConfig } from '#app'
import { getGameHomeShowcase, submitGameShowcaseEvent } from '../../app/services/game'
import { emptyGameShowcase } from '../../app/utils/gameShowcasePresentation'

afterEach(() => vi.unstubAllGlobals())

it('reads an independent localized slice with bounded optional failure and leaves the DTO untouched', async () => {
  const snapshot = emptyGameShowcase()
  const fetcher = vi.fn().mockResolvedValue({ code: 1, data: snapshot })
  vi.stubGlobal('$fetch', fetcher)
  for (const [input, lang] of [['en', 'en'], ['zh', 'zh'], ['bad', 'zh']]) {
    const result = await useNuxtApp().runWithContext(() => getGameHomeShowcase(input))
    expect(result).toBe(snapshot)
    expect(fetcher).toHaveBeenLastCalledWith('/game/home/showcase', expect.objectContaining({ query: { lang, region: 'CN' }, retry: 0, timeout: 1000 }))
  }
})

it('accepts an empty 204 directly and silently settles transport errors', async () => {
  const fetcher = vi.fn().mockResolvedValue(undefined)
  vi.stubGlobal('$fetch', fetcher)
  const body = { tracking_token: 'opaque', session_id: '47fde791-293c-4147-8eba-d41cfc03adab', event: 'click' as const, source: 'title' as const }
  await expect(useNuxtApp().runWithContext(() => submitGameShowcaseEvent(body))).resolves.toBeUndefined()
  expect(fetcher).toHaveBeenCalledExactlyOnceWith('/game/home/showcase/events', {
    baseURL: useRuntimeConfig().public.gameV2ApiBase, method: 'POST', body,
    credentials: 'include', keepalive: true, retry: 0, timeout: 5000,
    onResponse: expect.any(Function),
  })
  const emptyResponse = new Response(null, { status: 204 })
  const read = vi.spyOn(emptyResponse, 'text')
  await fetcher.mock.calls[0]![1].onResponse({ response: emptyResponse })
  expect(read).toHaveBeenCalledTimes(1)
  fetcher.mockRejectedValue(new Error('analytics unavailable'))
  await expect(useNuxtApp().runWithContext(() => submitGameShowcaseEvent(body))).resolves.toBeUndefined()
})
