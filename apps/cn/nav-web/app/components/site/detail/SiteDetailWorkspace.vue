<template>
  <section
    id="site-workspace"
    data-site-workspace :data-site-workspace-tab="active"
    role="tabpanel" :aria-labelledby="'site-tab-' + active" tabindex="0"
    class="site-detail-workspace min-w-0 xl:order-1"
  >
    <h2 class="site-detail-workspace-title mb-4">{{ t('siteDetail.tabs.' + active) }}</h2>
    <!-- Keep a single Site-owned slice mounted across every Target/workspace. -->
    <SiteInsightsPanel v-show="active === 'overview' || active === 'insights'" :insights="insights" :unavailable="insightsUnavailable" />
    <!-- Transitional panels retain their owners until P3–P6. A resolved Target
         change resets their local history/tab state without remounting the shell. -->
    <div :key="data.domain" :aria-busy="pending" class="min-w-0">
      <div v-if="active === 'overview'" class="mt-5">
        <h3 class="site-detail-label mb-3">{{ t('siteDetail.checks') }}</h3>
        <SiteObservationInfoList :items="checks" :empty-text="t('siteDetail.notObserved')" />
      </div>
      <div v-else-if="active === 'observation'" class="space-y-5">
        <SitePerformancePanel
          v-if="data.siteHttpRecord" :domain="data.domain" :site-id="siteId"
          :http-record="data.siteHttpRecord" :ping-record="data.sitePingRecord" :target-latest-core="data.targetLatestCore"
        />
        <SiteObservationTabs
          :dns-record="data.siteDnsRecord" :http-record="data.siteHttpRecord"
          :ping-record="data.sitePingRecord" :target-latest-core="data.targetLatestCore"
        />
        <SiteMetadataProbePanel
          :http-record="data.siteHttpRecord" :light-probe-state="data.lightProbeState"
          :site-id="siteId" :target="data.domain" :target-latest-core="data.targetLatestCore"
        />
      </div>
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
import SiteObservationInfoList from '../SiteObservationInfoList.vue'
import SiteObservationMetricGrid from '../SiteObservationMetricGrid.vue'
import SitePerformancePanel from '../SitePerformancePanel.vue'
import SiteObservationTabs from '../SiteObservationTabs.vue'
import SiteMetadataProbePanel from '../SiteMetadataProbePanel.vue'
import type { SiteDetailPageData } from '~/composables/useSiteDetailPage'
import type { SiteDetailTab } from '~/utils/siteDetailRouteState'
import type { SiteTargetPresentation } from '~/utils/siteTargetPresentation'
import type { SiteInsights } from '~/types/insights'
import type { ObservationMetricItem } from '../detailTypes'
const props = defineProps<{
  data: SiteDetailPageData; siteId: string; active: SiteDetailTab; presentation: SiteTargetPresentation
  insights: SiteInsights | null; insightsUnavailable: boolean; pending: boolean
}>()
const { t } = useI18n()
const checks = computed(() => props.presentation.protocolStates.map(item => ({
  label: item.label, value: [item.statusLabel, item.duration, item.observedAt].join(' · '),
})))
const security = computed<ObservationMetricItem[]>(() => [
  { label: 'TLS', value: props.presentation.tlsVersion || '—', tone: 'normal' },
  { label: t('siteDetail.certificate'), value: props.presentation.certificateLabel, tone: props.presentation.certificateVerified === false ? 'warn' : 'normal' },
  { label: t('siteDetail.days'), value: props.presentation.certificateDays === null ? '—' : String(props.presentation.certificateDays), tone: 'normal' },
])
</script>
