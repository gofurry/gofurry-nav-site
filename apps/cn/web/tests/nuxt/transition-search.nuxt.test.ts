import { expect, it, vi } from 'vitest'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { useNuxtApp, useRuntimeConfig } from '#app'
import SearchBox from '../../app/components/nav/SearchBox.vue'
import { getSearchSuggestion } from '../../app/services/nav'

vi.mock('../../app/services/nav', () => ({ getSearchSuggestion: vi.fn() }))

it('transition search exposes five upper platforms and never schedules site suggestions', async () => {
  const config = useRuntimeConfig()
  const priorMode = config.public.transitionMode
  const i18n = useNuxtApp().$i18n
  const priorLocale = i18n.locale.value
  config.public.transitionMode = true
  i18n.locale.value = 'zh'
  const media = Object.assign(new EventTarget(), { matches: false }) as MediaQueryList
  vi.spyOn(window, 'matchMedia').mockReturnValue(media)
  const open = vi.spyOn(window, 'open').mockReturnValue(null)
  const view = await mountSuspended(SearchBox)
  try {
    expect(view.find('.search-category-row').exists()).toBe(false)
    expect(view.find('.search-platform-picker').exists()).toBe(false)
    const buttons = view.findAll('[data-transition-platforms] button')
    expect(buttons.map(button => button.text())).toEqual(['必应', 'B站', '小红书', '知乎', '微博'])
    expect(buttons[0]!.attributes('aria-pressed')).toBe('true')
    vi.useFakeTimers()
    await view.find('input').setValue('wolf & 兽人')
    await vi.advanceTimersByTimeAsync(1000)
    expect(getSearchSuggestion).not.toHaveBeenCalled()
    const query = 'wolf%20%26%20%E5%85%BD%E4%BA%BA'
    const urls = [`https://www.bing.com/search?q=${query}`, `https://search.bilibili.com/all?keyword=${query}`,
      `https://www.xiaohongshu.com/search_result?keyword=${query}`, `https://www.zhihu.com/search?type=content&q=${query}`, `https://s.weibo.com/weibo?q=${query}`]
    for (const [index, url] of urls.entries()) {
      await buttons[index]!.trigger('click')
      await view.find('input').trigger('keydown', { key: 'Enter' })
      expect(open).toHaveBeenLastCalledWith(url, '_blank')
    }
    expect(view.find('.search-suggestion-list').exists()).toBe(false)
  } finally {
    view.unmount()
    vi.useRealTimers()
    config.public.transitionMode = priorMode
    i18n.locale.value = priorLocale
    vi.restoreAllMocks()
  }
})
