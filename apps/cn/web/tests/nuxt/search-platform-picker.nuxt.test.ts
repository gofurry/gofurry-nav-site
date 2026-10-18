import { afterEach, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'
import { useNuxtApp } from '#app'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import SearchBox from '../../app/components/nav/SearchBox.vue'
import { isNavPageRevealLocked } from '../../app/utils/navPageReveal'

afterEach(() => vi.restoreAllMocks())

it('resets stable provider slots on live locale changes and releases the picker Reveal lock', async () => {
  const i18n = useNuxtApp().$i18n
  const originalLocale = i18n.locale.value
  await i18n.loadLocaleMessages('en')
  i18n.locale.value = 'zh'
  const media = Object.assign(new EventTarget(), { matches: true }) as MediaQueryList
  const add = vi.spyOn(media, 'addEventListener')
  const remove = vi.spyOn(media, 'removeEventListener')
  vi.spyOn(window, 'matchMedia').mockReturnValue(media)
  const open = vi.spyOn(window, 'open').mockReturnValue(null)
  const view = await mountSuspended(SearchBox)
  const trigger = () => view.find('.search-platform-trigger')
  const options = () => view.findAll('.search-platform-picker button')
  try {
    await view.find('input').setValue('wolf & 兽人')
    await trigger().trigger('click') // Cancels the unchanged suggestion debounce.
    expect(isNavPageRevealLocked()).toBe(true)
    await options()[4]!.trigger('click')
    expect(trigger().text()).toBe('小红书')
    expect(options()[4]!.text()).toBe('必应')
    i18n.locale.value = 'en'
    await nextTick()
    expect(options()).toHaveLength(1)
    expect(trigger().text()).toBe('Bing')
    expect(isNavPageRevealLocked()).toBe(false)
    expect(view.find('input').element.value).toBe('wolf & 兽人')
    await trigger().trigger('click')
    expect(options().map(option => option.text())).toEqual(['Google', 'DuckDuckGo', 'Bilibili', 'Bing', 'Rednote', 'Zhihu', 'Weibo', 'X (Twitter)'])
    await options()[0]!.trigger('click')
    expect(trigger().text()).toBe('Google')
    expect(trigger().attributes('aria-expanded')).toBe('true')
    expect(open).not.toHaveBeenCalled()
    await view.find('.search-icon').trigger('click')
    expect(open).toHaveBeenCalledExactlyOnceWith('https://www.google.com/search?q=wolf%20%26%20%E5%85%BD%E4%BA%BA', '_blank')
    expect(isNavPageRevealLocked()).toBe(true)
  } finally {
    view.unmount()
    i18n.locale.value = originalLocale
  }
  expect(isNavPageRevealLocked()).toBe(false)
  expect(remove.mock.calls).toEqual(add.mock.calls)
})
