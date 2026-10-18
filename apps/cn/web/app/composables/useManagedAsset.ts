import { computed, onBeforeUnmount, onMounted, ref, toValue, watch, type MaybeRefOrGetter } from 'vue'
import { assetCandidate, type AssetCDN } from '~/utils/managedAssets'

export interface ManagedAssetSnapshot {
  key: string | null | undefined
  provider: AssetCDN
  failed: Set<AssetCDN>
  failedFallback: boolean
}

export function useManagedAsset(key: MaybeRefOrGetter<string | null | undefined>, fallback = '', preload: MaybeRefOrGetter<boolean> = false, previous?: ManagedAssetSnapshot) {
  const cdn = useNuxtApp().$assetCDN
  // A Hero handoff can retain the other viewport's unchanged resource. Carry
  // its route/failure snapshot across frames, but never across different keys.
  const retained = previous?.key === toValue(key) ? previous : undefined
  const provider = ref<AssetCDN>(retained?.provider ?? cdn.resolvePreferred())
  const failed = ref(new Set<AssetCDN>(retained?.failed))
  const failedFallback = ref(retained?.failedFallback ?? false)
  const candidate = computed(() => assetCandidate(cdn.origins, provider.value, toValue(key), failed.value, failedFallback.value ? '' : fallback))
  const src = computed(() => candidate.value.url)
  // Route updates apply only to a new key, never to the current resource.
  watch(() => toValue(key), () => { provider.value = cdn.resolvePreferred(); failed.value = new Set(); failedFallback.value = false }, { flush: 'sync' })
  const onError = () => {
    if (candidate.value.provider) { failed.value = new Set([...failed.value, candidate.value.provider]); cdn.reportFailure() }
    else failedFallback.value = true
  }
  let stop: (() => void) | undefined
  onMounted(() => {
    if (preload !== false) stop = watch([src, () => toValue(preload)], ([url, enabled], _old, cleanup) => {
      if (!url || !enabled) return
      const image = new Image()
      // CSS masks fetch anonymously with CORS; probe with the same mode so a
      // missing CORS header triggers fallback instead of an invisible pattern.
      if (toValue(key)?.startsWith('nav/patterns/')) image.crossOrigin = 'anonymous'
      image.onerror = () => { if (src.value === url) onError() }
      image.src = url
      cleanup(() => { image.onerror = null })
    }, { immediate: true })
  })
  onBeforeUnmount(() => stop?.())
  const snapshot = (): ManagedAssetSnapshot => ({ key: toValue(key), provider: provider.value, failed: new Set(failed.value), failedFallback: failedFallback.value })
  return { src, onError, snapshot }
}
