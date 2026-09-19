import { afterEach, beforeEach, vi } from 'vitest'
import { clearNuxtState, refreshCookie } from '#app'

// Nuxt boots its real app/plugins/context. Replace only the two business
// providers so composable tests cannot start external CDN probes at app:mounted.
vi.mock('../../app/plugins/asset-cdn', async () => {
  const { defineNuxtPlugin } = await import('#app')
  const { ref } = await import('vue')
  const { resolveAssetPreferred } = await import('../../app/utils/managedAssets')
  return { default: defineNuxtPlugin(() => {
    const mode = ref<'auto' | 'primary' | 'mirror'>('auto')
    const recommendation = ref<'primary' | 'mirror'>('primary')
    return { provide: { assetCDN: {
      origins: { primary: 'https://primary.example', mirror: 'https://mirror.example' },
      mode, recommendation,
      resolvePreferred: () => resolveAssetPreferred(mode.value, recommendation.value),
      reportFailure: vi.fn(),
    } } }
  }) }
})

vi.mock('../../app/plugins/steam-asset-route', async () => {
  const { defineNuxtPlugin } = await import('#app')
  const { ref } = await import('vue')
  const { resolveSteamPreferred } = await import('../../app/utils/steamAssets')
  return { default: defineNuxtPlugin(() => {
    const mode = ref<'auto' | 'china' | 'global'>('auto')
    const recommendation = ref<'china' | 'global'>('china')
    return { provide: { steamAssetRoute: {
      mode, recommendation,
      resolvePreferred: (locale: string) => resolveSteamPreferred(mode.value, recommendation.value, locale),
      reportFailure: vi.fn(),
    } } }
  }) }
})

// Public Nuxt APIs only: no mutation of payload/_cookies or other internals.
// Each mounted harness is unmounted in its case before this cleanup runs.
function resetNuxtState() {
  for (const name of [
    'gf_hero_mode', 'gf_hero_desktop_id', 'gf_hero_mobile_id',
    'gf_asset_cdn', 'gf_asset_cdn_mode', 'gf_steam_asset_mode', 'gf_steam_asset_group',
  ]) {
    document.cookie = `${name}=; Max-Age=0; Path=/`
    refreshCookie(name)
  }
  clearNuxtState()
  localStorage.clear()
  sessionStorage.clear()
  vi.clearAllMocks()
}
beforeEach(resetNuxtState)
afterEach(resetNuxtState)
