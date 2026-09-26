<template>
  <div data-site-detail :data-site-target="sitePageData.domain" :data-site-tab="routeState.tab" class="site-detail-page relative isolate min-h-full min-w-0">
    <main class="relative mx-auto w-full min-w-0 max-w-[1560px] px-4 pb-10 pt-5 sm:px-8 sm:pt-8 lg:px-10">
      <SiteDetailHero
        :site="sitePageData.siteInfo" :name="siteName" :domain="sitePageData.domain"
        :view-count="siteViewCount" :visit-url="targetPresentation.visitUrl"
      />
      <SiteHealthStrip :presentation="targetPresentation" :pending="pending" class="mt-5 sm:mt-6" />
      <SitePrimaryTabs :active="routeState.tab" class="mt-6" @select="changeTab" />
      <div class="mt-5 grid min-w-0 items-start gap-5 xl:grid-cols-[minmax(0,3fr)_minmax(16rem,1fr)] xl:gap-6">
        <SiteTargetContext
          :presentation="targetPresentation" :pending="pending"
          :selected="routeState.domain || sitePageData.domain" :site-scope="routeState.tab === 'insights'"
          @select="changeTarget"
        />
        <SiteDetailWorkspace
          :data="sitePageData" :site-id="siteId" :active="routeState.tab"
          :presentation="targetPresentation" :insights="siteInsightsSnapshot.insights"
          :insights-unavailable="siteInsightsSnapshot.unavailable" :pending="pending"
        />
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, shallowRef, watch } from 'vue'
import SiteDetailHero from './SiteDetailHero.vue'
import SiteHealthStrip from './detail/SiteHealthStrip.vue'
import SitePrimaryTabs from './detail/SitePrimaryTabs.vue'
import SiteTargetContext from './detail/SiteTargetContext.vue'
import SiteDetailWorkspace from './detail/SiteDetailWorkspace.vue'
import { getSiteInsights } from '@/services/nav'
import type { SiteInsights } from '@/types/insights'
import { useSiteDetailPage } from '~/composables/useSiteDetailPage'
import { buildSiteDetailSeo } from '~/utils/seo'
import { authoritativePageStatus } from '~/utils/authoritativePageError'
import { buildSiteDetailQuery, selectSiteDetailTab, selectSiteDetailTarget, type SiteDetailTab } from '~/utils/siteDetailRouteState'
import { presentSiteTarget } from '~/utils/siteTargetPresentation'

interface SiteInsightsSnapshot {
  insights: SiteInsights | null
  unavailable: boolean
}
const route = useRoute()
const router = useRouter()
const { locale, t } = useI18n()
const requestedSiteId = computed(() => String(route.params.id ?? ''))
const detailRequest = useSiteDetailPage()
const insightsRequest = useAsyncData<SiteInsightsSnapshot>(
  () => `site-insights:${requestedSiteId.value}`,
  async () => {
    try {
      return { insights: await getSiteInsights(requestedSiteId.value), unavailable: false }
    } catch {
      return { insights: null, unavailable: true }
    }
  },
  { default: () => ({ insights: null, unavailable: false }) },
)
const [detailState, insightsState] = await Promise.all([detailRequest, insightsRequest])
const { data, pending, error, siteId, routeState } = detailState
const siteInsightsSnapshot = computed(() => insightsState.data.value)
const navV2Api = useApi('navV2')
// The reactive async key owns cancellation/stale-result isolation. Preserve the
// last resolved presentation while the next key is pending, without relabeling it.
const lastResolved = shallowRef(data.value!)
watch(data, value => { if (value?.siteInfo) lastResolved.value = value })
const sitePageData = computed(() => data.value?.siteInfo ? data.value : lastResolved.value)
const targetPresentation = computed(() => presentSiteTarget(sitePageData.value, t))
const countedView = ref<{ siteId: string; count: number } | null>(null)
const siteViewCount = computed(() => countedView.value?.siteId === siteId.value
  ? countedView.value.count : sitePageData.value.siteInfo?.view_count ?? 0)
const siteName = computed(() => sitePageData.value.siteInfo?.name?.trim() || 'GoFurry')

watch(error, failure => {
  if (!failure) return
  const statusCode = authoritativePageStatus(failure, 'site')
  showError(createError({
    statusCode,
    statusMessage: statusCode === 404 ? 'Site not found' : 'Site service temporarily unavailable',
    cause: failure,
  }))
})
function changeTab(tab: SiteDetailTab) {
  void router.push({ query: buildSiteDetailQuery(selectSiteDetailTab(routeState.value, tab)) })
}
function changeTarget(target: string) {
  void router.push({ query: buildSiteDetailQuery(selectSiteDetailTarget(routeState.value, target)) })
}
const seo = computed(() => buildSiteDetailSeo({
  name: sitePageData.value.siteInfo?.name,
  description: sitePageData.value.siteInfo?.info,
  domain: sitePageData.value.domain,
  locale: locale.value,
}))
useSeoMeta({
  title: () => seo.value.title, description: () => seo.value.description,
  ogTitle: () => seo.value.title, ogDescription: () => seo.value.description,
})
onMounted(() => {
  watch(siteId, value => { void touchSiteView(value) }, { immediate: true })
})
async function touchSiteView(value: string) {
  if (!value) return
  try {
    const response = await navV2Api<{ site_id: number; view_count: number }>(`/nav/sites/${value}/view`, { method: 'POST' })
    if (siteId.value === value && Number.isFinite(response.view_count)) countedView.value = { siteId: value, count: response.view_count }
  } catch {
    // View accounting is an optional side effect, never a page failure.
  }
}
</script>
