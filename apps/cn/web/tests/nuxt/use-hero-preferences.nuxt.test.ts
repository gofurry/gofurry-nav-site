import { expect, it } from 'vitest'
import { defineComponent, h, nextTick } from 'vue'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { clearNuxtState } from '#app'
import { useHeroPreferences } from '../../app/composables/useHeroPreferences'

function seed(name: string, value: string) { document.cookie = `${name}=${encodeURIComponent(value)}; Path=/` }
function readCookie(name: string) {
  const value = document.cookie.split('; ').find(cookie => cookie.startsWith(name + '='))?.slice(name.length + 1)
  return value === undefined ? undefined : decodeURIComponent(value)
}
function unrelatedCookies() { return document.cookie.split('; ').filter(cookie => !cookie.startsWith('gf_hero_')).sort() }
async function mountPreferences() {
  let settings!: ReturnType<typeof useHeroPreferences>
  const wrapper = await mountSuspended(defineComponent({
    setup() { settings = useHeroPreferences(); return () => h('div') },
  }))
  return { wrapper, settings }
}

it('round-trips bigint cookie IDs, retains pins when leaving Fixed, and ignores unrelated Save', async () => {
  seed('gf_hero_mode', 'fixed')
  seed('gf_hero_desktop_id', '9007199254740993')
  seed('gf_hero_mobile_id', '20')
  seed('gf_asset_cdn_mode', 'mirror')
  seed('gf_steam_asset_mode', 'china')
  const untouched = unrelatedCookies()
  const { wrapper, settings } = await mountPreferences()
  try {
    expect(settings.legacyPending.value).toBe(false)
    expect(settings.preference.value).toEqual({ mode: 'fixed', desktopId: '9007199254740993', mobileId: '20' })
    settings.save({ ...settings.preference.value, mode: 'random' })
    await nextTick()
    expect(readCookie('gf_hero_desktop_id')).toBe('9007199254740993')
    expect(readCookie('gf_hero_mobile_id')).toBe('20')
    expect(readCookie('gf_hero_mode')).toBe('random')
    expect(settings.revision.value).toBe(1)
    settings.save(settings.preference.value)
    await nextTick()
    expect(settings.revision.value).toBe(1)
    expect(readCookie('gf_asset_cdn_mode')).toBe('mirror')
    expect(readCookie('gf_steam_asset_mode')).toBe('china')
    expect(document.cookie.split('; ').map(item => item.split('=')[0]!).filter(name => name.startsWith('gf_')).sort()).toEqual([
      'gf_asset_cdn_mode', 'gf_hero_desktop_id', 'gf_hero_mobile_id', 'gf_hero_mode', 'gf_steam_asset_mode',
    ])
    expect(unrelatedCookies()).toEqual(untouched)
  } finally { wrapper.unmount() }

  // Recreate state from the cookies actually encoded by the real composable.
  clearNuxtState(['hero-preference', 'hero-legacy-pending', 'hero-preference-revision'])
  const reloaded = await mountPreferences()
  try {
    expect(reloaded.settings.preference.value).toEqual({ mode: 'random', desktopId: '9007199254740993', mobileId: '20' })
    expect(reloaded.settings.revision.value).toBe(0)
  } finally { reloaded.wrapper.unmount() }
})

it('starts with isolated empty state and clears legacyPending on first Save without reselecting Hero', async () => {
  const { wrapper, settings } = await mountPreferences()
  try {
    expect(settings.preference.value).toEqual({ mode: 'random', desktopId: null, mobileId: null })
    expect(settings.legacyPending.value).toBe(true)
    expect(settings.revision.value).toBe(0)
    settings.save(settings.preference.value)
    await nextTick()
    expect(settings.legacyPending.value).toBe(false)
    expect(settings.revision.value).toBe(0)
    expect(readCookie('gf_hero_mode')).toBe('random')
  } finally { wrapper.unmount() }
})

it('normalizes invalid cookie IDs and updates revision only for real/local changes', async () => {
  seed('gf_hero_mode', 'invalid')
  seed('gf_hero_desktop_id', '9223372036854775808')
  seed('gf_hero_mobile_id', '01')
  const { wrapper, settings } = await mountPreferences()
  try {
    expect(settings.preference.value).toEqual({ mode: 'random', desktopId: null, mobileId: null })
    settings.save({ mode: 'fixed', desktopId: '9223372036854775807', mobileId: '20' })
    expect(settings.revision.value).toBe(1)
    settings.save({ ...settings.preference.value, mobileId: '21' })
    expect(settings.revision.value).toBe(2)
    settings.save(settings.preference.value, true)
    expect(settings.revision.value).toBe(2)
    settings.save({ ...settings.preference.value, mode: 'local' })
    expect(settings.revision.value).toBe(3)
    settings.save(settings.preference.value)
    expect(settings.revision.value).toBe(3)
    settings.save(settings.preference.value, true)
    expect(settings.revision.value).toBe(4)
    await nextTick()
    expect(readCookie('gf_hero_mode')).toBe('local')
    expect(readCookie('gf_hero_desktop_id')).toBe('9223372036854775807')
    expect(readCookie('gf_hero_mobile_id')).toBe('21')
  } finally { wrapper.unmount() }
})
