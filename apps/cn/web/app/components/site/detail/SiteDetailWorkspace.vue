<template>
  <section
    id="site-workspace"
    data-site-workspace :data-site-workspace-tab="active"
    role="tabpanel" :aria-labelledby="'site-tab-' + active" tabindex="0"
    class="site-detail-workspace min-w-0 xl:order-1"
  >
    <h2 v-if="active === 'overview'" class="site-detail-workspace-title mb-4">{{ t('siteDetail.tabs.' + active) }}</h2>
    <SiteOverviewWorkspace v-if="active === 'overview'" :presentation="overview" :insights-to="insightsTo" :retrying="insightsRetrying" @retry="emit('insightsRetry')" />
    <!-- Fetch ownership stays on the Site page, independent of visible panels. -->
    <SiteInsightsWorkspace v-if="active === 'insights'" :site-id="siteId" :presentation="insights" :retrying="insightsRetrying" :metric="insightMetric" :range="insightRange" :trend="trend"
      @metric="emit('insightMetric', $event)" @range="emit('insightRange', $event)" @retry="emit('insightsRetry')" @trend-retry="emit('trendRetry')" />
    <!-- Target evidence remounts after a resolved switch; history cache stays on the page. -->
    <div :key="data.domain" :aria-busy="pending" class="min-w-0">
      <SiteObservationWorkspace v-if="active === 'observation'" :view="observationView" :presentation="observation" :history="history"
        @select="emit('observationView', $event)" @sample="emit('historySample', $event)" @retry="emit('historyRetry')" />
      <SiteSecurityWorkspace v-else-if="active === 'security'" :view="securityView" :presentation="security" :raw-headers-to="rawHeadersTo"
        @select="emit('securityView', $event)" />
    </div>
  </section>
</template>

<script setup lang="ts">
import SiteInsightsWorkspace from './insights/SiteInsightsWorkspace.vue'
import SiteOverviewWorkspace from './SiteOverviewWorkspace.vue'
import SiteSecurityWorkspace from './security/SiteSecurityWorkspace.vue'
import SiteObservationWorkspace from './observation/SiteObservationWorkspace.vue'
import type { SiteDetailPageData } from '~/composables/useSiteDetailPage'
import type { SiteDetailTab, SiteObservationView, SiteSecurityView, SiteInsightMetric, SiteInsightRange } from '~/utils/siteDetailRouteState'
import type { SiteInsightsPresentation } from '~/utils/siteInsightsPresentation'
import type { SiteInsightTrendPresentation } from '~/composables/useSiteInsightTrend'
import type { SiteOverviewPresentation } from '~/utils/siteOverviewPresentation'
import type { RouteLocationRaw } from 'vue-router'
import type { SiteSecurityPresentation } from '~/utils/siteSecurityPresentation'
import type { SiteObservationPresentation } from '~/utils/siteObservationPresentation'
import type { SiteHistoryPresentation, SiteHistorySample } from '~/composables/useSiteObservationHistory'
defineProps<{
  data: SiteDetailPageData; active: SiteDetailTab; siteId: string
  insights: SiteInsightsPresentation; insightsRetrying: boolean; pending: boolean
  insightMetric: SiteInsightMetric; insightRange: SiteInsightRange; trend: SiteInsightTrendPresentation
  overview: SiteOverviewPresentation; insightsTo: RouteLocationRaw
  observationView: SiteObservationView; observation: SiteObservationPresentation; history: SiteHistoryPresentation
  securityView: SiteSecurityView; security: SiteSecurityPresentation; rawHeadersTo: RouteLocationRaw
}>()
const emit = defineEmits<{ observationView: [view: SiteObservationView]; securityView: [view: SiteSecurityView]; historySample: [sample: SiteHistorySample]; historyRetry: [];
  insightMetric: [metric: SiteInsightMetric]; insightRange: [range: SiteInsightRange]; insightsRetry: []; trendRetry: [] }>()
const { t } = useI18n()
</script>
