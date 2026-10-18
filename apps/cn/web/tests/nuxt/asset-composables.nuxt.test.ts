import { beforeEach, expect, it } from 'vitest'
import { defineComponent, h, nextTick, ref } from 'vue'
import { mountSuspended } from '@nuxt/test-utils/runtime'
import { useNuxtApp } from '#app'
import { useManagedAsset } from '../../app/composables/useManagedAsset'
import { useSteamAsset } from '../../app/composables/useSteamAsset'
import { STEAM_SHARED_CDN_GROUP_PREFIXES } from '../../app/utils/steamAssets'

const originalKey = 'nav/hero/desktop/' + 'a'.repeat(32) + '.avif'
const source = 'https://shared.akamai.steamstatic.com/store_item_assets/steam/apps/570/hash/header.jpg?v=2#cover'

beforeEach(() => {
  const { $assetCDN: asset, $steamAssetRoute: steam } = useNuxtApp()
  asset.mode.value = 'auto'; asset.recommendation.value = 'primary'
  steam.mode.value = 'auto'; steam.recommendation.value = 'china'
})

async function mountAssets<T>(setup: () => T) {
  let value!: T
  const wrapper = await mountSuspended(defineComponent({
    setup() { value = setup(); return () => h('div') },
  }))
  return { value, wrapper }
}

it('freezes current resources across recommendation/Save, uses new policy on changed keys, and preserves pins after fallback', async () => {
  const { $assetCDN: assetRoute, $steamAssetRoute: steamRoute } = useNuxtApp()
  const key = ref(originalKey), steamSource = ref(source)
  const { value: { managed, steam }, wrapper } = await mountAssets(() => ({
    managed: useManagedAsset(key, '/default.svg'), steam: useSteamAsset(steamSource),
  }))
  try {
    const originalManaged = managed.src.value, originalSteam = steam.src.value
    assetRoute.recommendation.value = 'mirror'; steamRoute.recommendation.value = 'global'
    await nextTick()
    expect(managed.src.value).toBe(originalManaged)
    expect(steam.src.value).toBe(originalSteam)
    assetRoute.mode.value = 'mirror'; steamRoute.mode.value = 'global'
    await nextTick()
    expect(managed.src.value).toBe(originalManaged)
    expect(steam.src.value).toBe(originalSteam)
    key.value = originalKey.replace('a'.repeat(32), 'b'.repeat(32))
    steamSource.value = source.replace('header.jpg', 'capsule.jpg')
    await nextTick()
    expect(managed.src.value).toBe('https://mirror.example/' + key.value)
    expect(steam.src.value.startsWith(STEAM_SHARED_CDN_GROUP_PREFIXES.global[0]!)).toBe(true)
    managed.onError()
    expect(managed.src.value).toBe('https://primary.example/' + key.value)
    managed.onError()
    expect(managed.src.value).toBe('/default.svg')
    expect(assetRoute.mode.value).toBe('mirror')
    for (const prefix of [...STEAM_SHARED_CDN_GROUP_PREFIXES.global, ...STEAM_SHARED_CDN_GROUP_PREFIXES.china].slice(1)) {
      expect(steam.onError()).toBe(true)
      expect(steam.src.value.startsWith(prefix)).toBe(true)
    }
    expect(steam.onError()).toBe(false)
    expect(steamRoute.mode.value).toBe('global')
    expect(assetRoute.reportFailure).toHaveBeenCalled()
    expect(steamRoute.reportFailure).toHaveBeenCalled()
  } finally { wrapper.unmount() }
})

it('retains a snapshot only for the same key and carries failed providers across Hero handoffs', async () => {
  const { $assetCDN: route } = useNuxtApp()
  const first = await mountAssets(() => useManagedAsset(originalKey, '/default.svg'))
  const unmounts = [() => first.wrapper.unmount()]
  try {
    const original = first.value.src.value
    route.recommendation.value = 'mirror'; route.mode.value = 'mirror'
    const handoff = await mountAssets(() => ({
      retained: useManagedAsset(originalKey, '/default.svg', false, first.value.snapshot()),
      fresh: useManagedAsset('nav/hero/mobile/' + 'c'.repeat(32) + '.avif', '', false, first.value.snapshot()),
    }))
    unmounts.push(() => handoff.wrapper.unmount())
    expect(handoff.value.retained.src.value).toBe(original)
    expect(handoff.value.fresh.src.value.startsWith('https://mirror.example/')).toBe(true)
    handoff.value.retained.onError()
    const retry = await mountAssets(() => useManagedAsset(originalKey, '/default.svg', false, handoff.value.retained.snapshot()))
    unmounts.push(() => retry.wrapper.unmount())
    expect(retry.value.src.value).toBe('https://mirror.example/' + originalKey)
    retry.value.onError()
    expect(retry.value.src.value).toBe('/default.svg')
    retry.value.onError()
    const exhausted = await mountAssets(() => useManagedAsset(originalKey, '/default.svg', false, retry.value.snapshot()))
    unmounts.push(() => exhausted.wrapper.unmount())
    expect(exhausted.value.src.value).toBe('')
  } finally { unmounts.reverse().forEach(unmount => unmount()) }
})

it('does not retry the same failed URL when both provider origins are identical', async () => {
  const { $assetCDN: route } = useNuxtApp()
  const mirror = route.origins.mirror
  route.origins.mirror = route.origins.primary
  const { value, wrapper } = await mountAssets(() => useManagedAsset(originalKey, '/default.svg'))
  try {
    value.onError()
    expect(value.src.value).toBe('/default.svg')
    value.onError()
    expect(value.src.value).toBe('')
    expect(route.reportFailure).toHaveBeenCalledTimes(1)
  } finally { wrapper.unmount(); route.origins.mirror = mirror }
})
