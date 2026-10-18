import { computed, toValue, type MaybeRefOrGetter } from 'vue'
import { getSiteRecommendations } from '~/services/nav'
import type { SiteRecommendations } from '~/types/nav'

/** Optional SSR slice. Nuxt's reactive identity isolates late Site/language responses. */
export async function useSiteRecommendations(siteId: MaybeRefOrGetter<string>) {
  const { locale } = useI18n(), nuxt = useNuxtApp()
  const lang = computed(() => locale.value === 'en' ? 'en' : 'zh')
  const identity = computed(() => `${toValue(siteId)}:${lang.value}`)
  const request = await useAsyncData(() => `site-recommendations:${identity.value}`, async () => {
    const key = identity.value, id = toValue(siteId), language = lang.value
    let data: SiteRecommendations | null = null
    try {
      const result = await nuxt.runWithContext(() => getSiteRecommendations(id, language))
      if (String(result.site_id) === id && result.state === 'ready' && Array.isArray(result.items)) data = result
    } catch { /* Discovery failure is invisible and cannot fail the Site page. */ }
    return { key, data }
  }, { default: () => ({ key: '', data: null }) })
  const items = computed(() => request.data.value?.key === identity.value
    ? (request.data.value.data?.items ?? []).filter(site => String(site.id) !== toValue(siteId)).slice(0, 8) : [])
  return { items }
}
