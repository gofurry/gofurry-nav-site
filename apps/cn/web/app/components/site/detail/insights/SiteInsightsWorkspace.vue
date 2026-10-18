<template>
  <div data-site-insights data-site-insights-workspace :data-site-insights-state="presentation.state" class="site-intelligence min-w-0 space-y-7">
    <header data-site-insights-header>
      <div class="flex flex-wrap items-center justify-between gap-3">
        <h2 class="site-detail-workspace-title">{{ t('siteDetail.tabs.insights') }}</h2>
        <div class="flex flex-wrap gap-x-5 gap-y-2">
          <NuxtLink data-site-insights-ecosystem :to="localePath('/insights/sites')" class="site-detail-text-link inline-flex items-center gap-1">{{ t('siteIntelligence.ecosystem') }}<PhArrowSquareOut class="site-detail-icon" aria-hidden="true" /></NuxtLink>
          <NuxtLink data-site-insights-compare :to="localePath({ path: '/insights/sites/compare', query: { ids: siteId } })" class="site-detail-text-link inline-flex items-center gap-1">{{ t('siteIntelligence.compare') }}<PhArrowsLeftRight class="site-detail-icon" aria-hidden="true" /></NuxtLink>
        </div>
      </div>
      <div v-if="presentation.state === 'unavailable'" data-site-insights-unavailable class="mt-3 flex flex-wrap items-baseline gap-3" :aria-busy="retrying">
        <p role="status" class="site-detail-note">{{ t('siteIntelligence.unavailable') }}</p>
        <button data-site-insights-retry type="button" :disabled="retrying" class="site-detail-text-link" @click="emit('retry')">{{ t(retrying ? 'siteIntelligence.loading' : 'siteIntelligence.retry') }}</button>
      </div>
    </header>
    <SiteCapabilityMatrix :groups="presentation.groups" :selected="metric" @select="emit('metric', $event)" />
    <SiteCapabilityDetail :capability="presentation.selected">
      <SiteInsightTrend :presentation="trend" :metric="metric" :range="range" @range="emit('range', $event)" @retry="emit('trendRetry')" />
    </SiteCapabilityDetail>
    <SiteInsightChanges :items="presentation.changes" :state="presentation.changesState" />
  </div>
</template>
<script setup lang="ts">
import type { SiteInsightsPresentation } from '~/utils/siteInsightsPresentation'
import { PhArrowSquareOut, PhArrowsLeftRight } from '@phosphor-icons/vue'
import type { SiteInsightMetric, SiteInsightRange } from '~/utils/siteDetailRouteState'
import type { SiteInsightTrendPresentation } from '~/composables/useSiteInsightTrend'
import SiteCapabilityMatrix from './SiteCapabilityMatrix.vue'
import SiteCapabilityDetail from './SiteCapabilityDetail.vue'
import SiteInsightTrend from './SiteInsightTrend.vue'
import SiteInsightChanges from './SiteInsightChanges.vue'
defineProps<{ siteId: string; presentation: SiteInsightsPresentation; metric: SiteInsightMetric; range: SiteInsightRange; trend: SiteInsightTrendPresentation; retrying: boolean }>()
const emit = defineEmits<{ metric: [value: SiteInsightMetric]; range: [value: SiteInsightRange]; retry: []; trendRetry: [] }>()
const { t } = useI18n()
const localePath = useLocalePath()
</script>
