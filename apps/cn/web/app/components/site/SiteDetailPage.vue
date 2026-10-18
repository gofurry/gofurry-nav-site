<template>
  <div data-site-detail :data-site-target="sitePageData.domain" :data-site-tab="routeState.tab" class="site-detail-page relative isolate min-h-full min-w-0">
    <main class="relative mx-auto w-full min-w-0 max-w-[1560px] px-4 pb-10 pt-5 sm:px-8 sm:pt-8 lg:px-10">
      <div data-site-identity-note class="site-detail-identity-note min-w-0">
        <SiteDetailHero
          :site="sitePageData.siteInfo" :name="siteName" :domain="sitePageData.domain"
          :view-count="siteViewCount" :visit-url="targetPresentation.visitUrl"
          :edge-hints="targetPresentation.edgeProviderHints"
        />
        <div class="site-detail-perforation relative mx-5 flex items-center justify-between sm:mx-7" aria-hidden="true"><span /><span /><span /></div>
        <SiteHealthStrip :presentation="targetPresentation" :pending="pending" />
      </div>
      <SitePrimaryTabs :active="routeState.tab" :has-similar="hasSimilar" :aux-active="mobileAuxTab === 'similar'" :desktop="isDesktop" class="mt-6" @select="changeTab" @similar="mobileAuxTab = 'similar'" />
      <div class="mt-5 grid min-w-0 items-start gap-5 xl:grid-cols-[minmax(0,3fr)_minmax(16rem,1fr)] xl:gap-6">
        <div data-site-detail-aside class="site-detail-aside min-w-0 space-y-5 xl:sticky xl:order-2">
          <SiteTargetContext
            :presentation="targetPresentation" :pending="pending"
            :selected="routeState.domain || sitePageData.domain" :site-scope="routeState.tab === 'insights'"
            @select="changeTarget"
          />
          <SiteSimilarSites v-if="visibleSimilar.length || mobileAuxTab" :items="visibleSimilar" :variant="mobileAuxTab ? 'panel' : 'aside'" :class="mobileAuxTab ? '' : 'hidden xl:block'" />
        </div>
        <SiteDetailWorkspace v-show="!mobileAuxTab"
          :data="sitePageData" :active="routeState.tab" :site-id="siteId" :pending="pending"
          :insights="insightsPresentation" :insights-retrying="siteInsights.retrying.value" :trend="insightTrend.current.value"
          :insight-metric="insightMetric" :insight-range="insightRange"
          @insights-retry="siteInsights.retry" @trend-retry="insightTrend.retry" @insight-metric="changeInsightMetric" @insight-range="changeInsightRange"
          :overview="overviewPresentation" :insights-to="insightsTo"
          :observation-view="routeState.tab === 'observation' ? routeState.view : 'overview'"
          :observation="observationPresentation" :history="historyPresentation"
          :security-view="routeState.tab === 'security' ? routeState.view : 'overview'"
          :security="securityPresentation" :raw-headers-to="rawHeadersTo" @security-view="changeSecurityView"
          @observation-view="changeObservationView" @history-sample="observationHistory.selectSample" @history-retry="observationHistory.retry"
        />
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, shallowRef, watch } from 'vue'
import { useSiteRecommendations } from '~/composables/useSiteRecommendations'
import { readDisplayMode, subscribeModeChange, type DisplayMode } from '~/utils/modeStorage'
import SiteSimilarSites from './detail/SiteSimilarSites.vue'
import SiteDetailHero from './SiteDetailHero.vue'
import SiteHealthStrip from './detail/SiteHealthStrip.vue'
import SitePrimaryTabs from './detail/SitePrimaryTabs.vue'
import SiteTargetContext from './detail/SiteTargetContext.vue'
import SiteDetailWorkspace from './detail/SiteDetailWorkspace.vue'
import { useSiteInsights } from '~/composables/useSiteInsights'
import { useSiteInsightTrend } from '~/composables/useSiteInsightTrend'
import { presentSiteInsights } from '~/utils/siteInsightsPresentation'
import { useSiteDetailPage } from '~/composables/useSiteDetailPage'
import { buildSiteDetailSeo } from '~/utils/seo'
import { authoritativePageStatus } from '~/utils/authoritativePageError'
import { buildSiteDetailQuery, selectSiteDetailTab, selectSiteDetailTarget, selectSiteObservationView, selectSiteSecurityView, selectSiteInsightMetric, selectSiteInsightRange, type SiteDetailTab, type SiteObservationView, type SiteSecurityView, type SiteInsightMetric, type SiteInsightRange } from '~/utils/siteDetailRouteState'
import { presentSiteTarget } from '~/utils/siteTargetPresentation'
import { presentSiteOverview } from '~/utils/siteOverviewPresentation'
import { presentSiteObservation } from '~/utils/siteObservationPresentation'
import { presentSiteSecurity } from '~/utils/siteSecurityPresentation'
import { useSiteObservationHistory } from '~/composables/useSiteObservationHistory'

const route = useRoute()
const router = useRouter()
const { locale, t } = useI18n()
const requestedSiteId = computed(() => String(route.params.id ?? ''))
const detailRequest = useSiteDetailPage()
const [detailState, siteInsights, recommendations] = await Promise.all([detailRequest, useSiteInsights(requestedSiteId), useSiteRecommendations(requestedSiteId)])
const { data, pending, error, siteId, routeState } = detailState
const displayMode = ref<DisplayMode>('sfw'), mobileAuxTab = ref<'similar' | null>(null), isDesktop = ref(false)
const hasSimilar = computed(() => recommendations.items.value.length > 0)
const visibleSimilar = computed(() => recommendations.items.value.filter(site => displayMode.value === 'nsfw' || site.nsfw !== '1'))
let stopMode: (() => void) | undefined, stopDesktop: (() => void) | undefined
onMounted(() => {
  displayMode.value = readDisplayMode()
  stopMode = subscribeModeChange(({ displayMode: mode }) => { displayMode.value = mode })
  const media = window.matchMedia('(min-width: 1280px)')
  const update = () => {
    isDesktop.value = media.matches
    if (media.matches && mobileAuxTab.value) {
      mobileAuxTab.value = null
      if (document.activeElement?.id === 'site-tab-similar') document.getElementById('site-tab-' + routeState.value.tab)?.focus({ preventScroll: true })
    }
  }
  update(); media.addEventListener('change', update)
  stopDesktop = () => media.removeEventListener('change', update)
})
onBeforeUnmount(() => { stopMode?.(); stopDesktop?.() })
watch([requestedSiteId, locale, hasSimilar], () => { mobileAuxTab.value = null })
watch(() => route.fullPath, () => { mobileAuxTab.value = null })
const insightMetric = computed(() => routeState.value.tab === 'insights' ? routeState.value.metric : 'ipv6')
const insightRange = computed(() => routeState.value.tab === 'insights' ? routeState.value.range : '30d')
const insightsPresentation = computed(() => presentSiteInsights(siteInsights.data.value, siteInsights.state.value, insightMetric.value, t, locale.value))
const insightTrend = useSiteInsightTrend({ active: () => routeState.value.tab === 'insights', metric: insightMetric, range: insightRange })
// Overview keeps the first authoritative Site/language snapshot for this page
// session. A Target detail refresh cannot replace it; reload starts a new one.
const siteSnapshot = shallowRef({ identity: data.value.siteIdentity, summary: data.value.siteHealthSummary })
watch(data, value => {
  if (value?.siteInfo && value.siteIdentity !== siteSnapshot.value.identity) {
    siteSnapshot.value = { identity: value.siteIdentity, summary: value.siteHealthSummary }
  }
})
const overviewPresentation = computed(() => presentSiteOverview(siteSnapshot.value.summary, siteInsights.data.value,
  siteInsights.state.value === 'unavailable', (key, values = {}) => t(key, values), locale.value))
const insightsTo = computed(() => tabLocation('insights'))
const navV2Api = useApi('navV2')
// The reactive async key owns cancellation/stale-result isolation. Preserve the
// last resolved presentation while the next key is pending, without relabeling it.
const lastResolved = shallowRef(data.value!)
watch(data, value => { if (value?.siteInfo) lastResolved.value = value })
const sitePageData = computed(() => data.value?.siteInfo ? data.value : lastResolved.value)
const targetPresentation = computed(() => presentSiteTarget(sitePageData.value, t))
const observationPresentation = computed(() => presentSiteObservation(sitePageData.value, t))
const securityPresentation = computed(() => presentSiteSecurity(sitePageData.value, t))
const rawHeadersTo = computed(() => ({ path: route.path,
  query: buildSiteDetailQuery(selectSiteObservationView(routeState.value, 'http')) }))
const observationHistory = useSiteObservationHistory({ siteId, target: () => sitePageData.value.domain,
  active: () => routeState.value.tab === 'observation' && routeState.value.view === 'performance'
    && !pending.value && !error.value && Boolean(data.value?.siteInfo) })
const historyPresentation = computed(() => ({ state: observationHistory.state.value, rows: observationHistory.rows.value,
  sample: observationHistory.sample.value, total: observationHistory.total.value }))
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
  mobileAuxTab.value = null
  void router.push(tabLocation(tab))
}
function tabLocation(tab: SiteDetailTab) {
  return { path: route.path, query: buildSiteDetailQuery(selectSiteDetailTab(routeState.value, tab)) }
}
function changeTarget(target: string) {
  void router.push({ query: buildSiteDetailQuery(selectSiteDetailTarget(routeState.value, target)) })
}
function changeObservationView(view: SiteObservationView) {
  void router.push({ path: route.path, query: buildSiteDetailQuery(selectSiteObservationView(routeState.value, view)) })
}
function changeSecurityView(view: SiteSecurityView) {
  void router.push({ path: route.path, query: buildSiteDetailQuery(selectSiteSecurityView(routeState.value, view)) })
}
function changeInsightMetric(metric: SiteInsightMetric) {
  void router.push({ path: route.path, query: buildSiteDetailQuery(selectSiteInsightMetric(routeState.value, metric)) })
}
function changeInsightRange(range: SiteInsightRange) {
  void router.push({ path: route.path, query: buildSiteDetailQuery(selectSiteInsightRange(routeState.value, range)) })
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
