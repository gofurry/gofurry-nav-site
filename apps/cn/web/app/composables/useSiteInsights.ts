import { computed, toValue, type MaybeRefOrGetter } from 'vue'
import { getSiteInsights } from '~/services/nav'
import type { SiteInsights } from '~/types/insights'
import { siteInsightsState } from '~/utils/siteInsightsPresentation'

/** Page-owned SSR slice shared by Overview and Insights. Target is not an identity. */
export async function useSiteInsights(siteId: MaybeRefOrGetter<string>) {
  const nuxt = useNuxtApp()
  const request = await useAsyncData(() => `site-insights:${toValue(siteId)}`, async () => {
    const id = toValue(siteId)
    let data: SiteInsights | null = null
    try {
      const response = await nuxt.runWithContext(() => getSiteInsights(id))
      if (String(response.site.id) === id) data = response
    } catch { /* Optional slice failure never escalates to a page error. */ }
    return { id, data }
  }, { default: () => ({ id: '', data: null }) })
  const data = computed(() => request.data.value?.id === toValue(siteId) ? request.data.value.data : null)
  return { data, state: computed(() => siteInsightsState(data.value)), retrying: request.pending,
    retry: () => request.refresh({ dedupe: 'defer' }) }
}
