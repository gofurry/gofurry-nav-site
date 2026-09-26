<template>
  <section
    id="site-workspace"
    data-site-workspace :data-site-workspace-tab="active"
    role="tabpanel" :aria-labelledby="'site-tab-' + active" tabindex="0"
    class="site-detail-workspace min-w-0 xl:order-1"
  >
    <h2 class="site-detail-workspace-title mb-4">{{ t('siteDetail.tabs.' + active) }}</h2>
    <SiteOverviewWorkspace v-if="active === 'overview'" :presentation="overview" :insights-to="insightsTo" />
    <!-- Fetch ownership stays on the Site page, independent of visible panels. -->
    <SiteInsightsPanel v-if="active === 'insights'" :insights="insights" :unavailable="insightsUnavailable" />
    <!-- Target evidence remounts after a resolved switch; history cache stays on the page. -->
    <div :key="data.domain" :aria-busy="pending" class="min-w-0">
      <SiteObservationWorkspace v-if="active === 'observation'" :view="observationView" :presentation="observation" :history="history"
        @select="emit('observationView', $event)" @sample="emit('historySample', $event)" @retry="emit('historyRetry')" />
      <div v-else-if="active === 'security'">
        <h3 class="site-detail-label mb-3">{{ t('siteDetail.securityEvidence') }}</h3>
        <SiteObservationMetricGrid :items="security" />
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import SiteInsightsPanel from '../SiteInsightsPanel.vue'
import SiteOverviewWorkspace from './SiteOverviewWorkspace.vue'
import SiteObservationMetricGrid from '../SiteObservationMetricGrid.vue'
import SiteObservationWorkspace from './observation/SiteObservationWorkspace.vue'
import type { SiteDetailPageData } from '~/composables/useSiteDetailPage'
import type { SiteDetailTab, SiteObservationView } from '~/utils/siteDetailRouteState'
import type { SiteTargetPresentation } from '~/utils/siteTargetPresentation'
import type { SiteInsights } from '~/types/insights'
import type { SiteOverviewPresentation } from '~/utils/siteOverviewPresentation'
import type { RouteLocationRaw } from 'vue-router'
import type { ObservationMetricItem } from '../detailTypes'
import type { SiteObservationPresentation } from '~/utils/siteObservationPresentation'
import type { SiteHistoryPresentation, SiteHistorySample } from '~/composables/useSiteObservationHistory'
const props = defineProps<{
  data: SiteDetailPageData; siteId: string; active: SiteDetailTab; presentation: SiteTargetPresentation
  insights: SiteInsights | null; insightsUnavailable: boolean; pending: boolean
  overview: SiteOverviewPresentation; insightsTo: RouteLocationRaw
  observationView: SiteObservationView; observation: SiteObservationPresentation; history: SiteHistoryPresentation
}>()
const emit = defineEmits<{ observationView: [view: SiteObservationView]; historySample: [sample: SiteHistorySample]; historyRetry: [] }>()
const { t } = useI18n()
const security = computed<ObservationMetricItem[]>(() => [
  { label: 'TLS', value: props.presentation.tlsVersion || '—', tone: 'normal' },
  { label: t('siteDetail.certificate'), value: props.presentation.certificateLabel, tone: props.presentation.certificateVerified === false ? 'warn' : 'normal' },
  { label: t('siteDetail.days'), value: props.presentation.certificateDays === null ? '—' : String(props.presentation.certificateDays), tone: 'normal' },
])
</script>
