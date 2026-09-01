<template>
  <div
    class="insights-page insights-domain-page"
    :data-selected-metric="selectedMetric"
    :data-selected-dimension="selectedDimension"
    :data-selected-slice="selectedSlice || ''"
  >
    <GoFurryGridBackground :fixed="false" profile="light" />
    <main class="insights-container">
      <InsightsNav />

      <header class="insights-hero insights-hero--domain">
        <p class="insights-eyebrow">{{ $t('insights.sites.eyebrow') }}</p>
        <h1>{{ $t('insights.sites.title') }}</h1>
        <p>{{ $t('insights.sites.description') }}</p>
      </header>

      <p v-if="overviewUnavailable" class="insights-empty-state insights-overview-error">
        {{ $t('insights.emptyStates.unavailable') }}
      </p>

      <div class="insights-metric-strip" aria-label="Website metrics">
        <InsightsMetricCard
          v-for="metricKey in metrics"
          :key="metricKey"
          :metric-key="metricKey"
          :metric="metricsByKey.get(metricKey) ?? null"
          :selected="selectedMetric === metricKey"
          @select="selectMetric"
        />
      </div>

      <InsightsMetricTrend
        :metric-key="selectedMetric"
        :range="selectedRange"
        :points="trend?.points ?? []"
        :loading="trendLoading"
        :unavailable="trendUnavailable"
        @range="selectRange"
      />

      <InsightsDimensionBreakdown
        :dimensions="dimensions"
        :dimension="selectedDimension"
        :selected-slice="selectedSlice"
        :breakdown="breakdown"
        :loading="breakdownLoading"
        :unavailable="breakdownUnavailable"
        @dimension="selectDimension"
        @slice="selectSlice"
      />

      <InsightsSliceTrend
        :slice="selectedSlice"
        :range="selectedRange"
        :trend="sliceTrend"
        :loading="sliceTrendLoading"
        :unavailable="sliceTrendUnavailable"
      />

      <InsightsDataInfo :metric-key="selectedMetric" :metric="selectedMetricData" />
      <InsightsRecentChanges :items="recentChanges" :unavailable="overviewUnavailable" />
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import GoFurryGridBackground from '@/components/common/GoFurryGridBackground.vue'
import InsightsDataInfo from '@/components/insights/InsightsDataInfo.vue'
import InsightsDimensionBreakdown from '@/components/insights/InsightsDimensionBreakdown.vue'
import InsightsMetricCard from '@/components/insights/InsightsMetricCard.vue'
import InsightsMetricTrend from '@/components/insights/InsightsMetricTrend.vue'
import InsightsNav from '@/components/insights/InsightsNav.vue'
import InsightsRecentChanges from '@/components/insights/InsightsRecentChanges.vue'
import InsightsSliceTrend from '@/components/insights/InsightsSliceTrend.vue'
import { useInsightsDomain } from '@/composables/useInsightsDomain'
import { useInsightsDimensions } from '@/composables/useInsightsDimensions'
import { getNavInsightsBreakdown, getNavInsightsOverview, getNavInsightsSliceTrend, getNavInsightsTrend } from '@/services/nav'
import type { NavInsightMetricKey, SiteInsightDimension } from '@/types/insights'
import { buildInsightsSeo } from '@/utils/seo'

const navMetrics = ['ipv6', 'tls13', 'security_txt'] as const satisfies readonly NavInsightMetricKey[]
const siteDimensions = ['country', 'group', 'nsfw', 'public_interest'] as const satisfies readonly SiteInsightDimension[]
const { locale } = useI18n()
const {
  metrics,
  overviewUnavailable,
  trend,
  trendUnavailable,
  trendLoading,
  selectedMetric,
  selectedRange,
  selectedMetricData,
  metricsByKey,
  recentChanges,
  selectMetric,
  selectRange,
} = await useInsightsDomain({
  domain: 'site',
  defaultMetric: 'ipv6',
  metrics: navMetrics,
  getOverview: getNavInsightsOverview,
  getTrend: getNavInsightsTrend,
})
const {
  dimensions,
  selectedDimension,
  selectedSlice,
  breakdown,
  breakdownUnavailable,
  breakdownLoading,
  sliceTrend,
  sliceTrendUnavailable,
  sliceTrendLoading,
  selectDimension,
  selectSlice,
} = await useInsightsDimensions({
  domain: 'site',
  metric: selectedMetric,
  range: selectedRange,
  defaultDimension: 'country',
  dimensions: siteDimensions,
  getBreakdown: getNavInsightsBreakdown,
  getSliceTrend: getNavInsightsSliceTrend,
})
const seo = computed(() => buildInsightsSeo('sites', locale.value))

useSeoMeta({
  title: () => seo.value.title,
  description: () => seo.value.description,
  ogTitle: () => seo.value.title,
  ogDescription: () => seo.value.description,
})
</script>
