import { computed, onMounted, ref, shallowRef, toValue, watch, type MaybeRefOrGetter } from 'vue'
import { useI18n } from 'vue-i18n'
import { steamSharedAssetCandidates } from '~/utils/steamAssets'

export function useSteamAsset(source: MaybeRefOrGetter<string | null | undefined>, preload = false) {
  const route = useNuxtApp().$steamAssetRoute
  const { locale } = useI18n()
  const candidates = shallowRef<string[]>([])
  const index = ref(0)
  watch(() => toValue(source), value => {
    index.value = 0
    candidates.value = steamSharedAssetCandidates(value, route.resolvePreferred(locale.value))
  }, { immediate: true, flush: 'sync' })
  const src = computed(() => candidates.value[index.value] || '')
  const onError = () => {
    if (candidates.value.length > 1) route.reportFailure()
    if (index.value >= candidates.value.length - 1) return false
    index.value++
    return true
  }
  // Video poster attributes have no separate error event. Probe that same URL
  // as an image so posters use the same fallback contract as SteamAssetImage.
  onMounted(() => {
    if (preload) watch(src, (url, _old, cleanup) => {
      if (!url) return
      const image = new Image()
      image.onerror = () => { if (src.value === url) onError() }
      image.src = url
      cleanup(() => { image.onerror = null })
    }, { immediate: true })
  })
  return { src, onError }
}
