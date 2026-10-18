<template>
  <div class="insights-page insights-domain-page insights-site-domain" data-domain="site" :data-selected-metric="selectedMetric" :data-selected-dimension="selectedDimension" :data-selected-slice="selectedSlice || ''">
    <main class="insights-container">
      <EcosystemNavigation context="site" />
      <InsightsDomainHeader domain="site" :entity-count="overview?.entity_count ?? null" :generated-at="overview?.generated_at ?? null" />
      <p v-if="overviewUnavailable" class="insights-empty-state">{{ $t('insights.emptyStates.unavailable') }}</p>

      <InsightMetricRail :metrics="metrics" :metrics-by-key="metricsByKey" :selected-metric="selectedMetric" @select="selectMetric" />

      <InsightTrendWorkspace domain="site" :metric-key="selectedMetric" :metric="selectedMetricData" :range="selectedRange" :points="trend?.points ?? []" :loading="trendLoading" :unavailable="trendUnavailable" @range="selectRange" />

      <InsightDimensionExplorer domain="site" :metric-key="selectedMetric" :dimensions="dimensions" :dimension="selectedDimension" :selected-slice="selectedSlice" :breakdown="breakdown" :loading="breakdownLoading" :unavailable="breakdownUnavailable" :range="selectedRange" :slice-trend="sliceTrend" :slice-loading="sliceTrendLoading" :slice-unavailable="sliceTrendUnavailable" @dimension="selectDimension" @slice="selectSlice" />

      <InsightDomainActivity domain="site" :items="recentChanges" :unavailable="overviewUnavailable" />

      <section class="insight-domain-continue" aria-labelledby="domain-continue-title">
        <div class="insight-domain-heading"><h2 id="domain-continue-title">{{ $t('insights.domain.continue') }}</h2></div>
        <div>
          <NuxtLink v-for="item in destinations" :key="item.path" :to="localePath(item.path)"><span><strong>{{ $t(`insights.editorial.links.${item.key}.title`) }}</strong><span>{{ $t(`insights.editorial.links.${item.key}.description`) }}</span></span><span aria-hidden="true">↗</span></NuxtLink>
        </div>
      </section>
      <InsightsDataInfo :metric-key="selectedMetric" :metric="selectedMetricData" />
    </main>
  </div>
</template>

<script setup lang="ts">
import EcosystemNavigation from '@/components/insights/EcosystemNavigation.vue'
import { computed } from 'vue'
import InsightsDomainHeader from '@/components/insights/domain/InsightsDomainHeader.vue'
import InsightMetricRail from '@/components/insights/domain/InsightMetricRail.vue'
import InsightTrendWorkspace from '@/components/insights/domain/InsightTrendWorkspace.vue'
import InsightDimensionExplorer from '@/components/insights/domain/InsightDimensionExplorer.vue'
import InsightDomainActivity from '@/components/insights/domain/InsightDomainActivity.vue'
import { overviewExploreGroups } from '@/utils/insightOverview'
import { useI18n } from 'vue-i18n'
import InsightsDataInfo from '@/components/insights/InsightsDataInfo.vue'
import { useInsightsDomain } from '@/composables/useInsightsDomain'
import { useInsightsDimensions } from '@/composables/useInsightsDimensions'
import { getNavInsightsBreakdown, getNavInsightsOverview, getNavInsightsSliceTrend, getNavInsightsTrend } from '@/services/nav'
import type { NavInsightMetricKey, SiteInsightDimension } from '@/types/insights'
import { buildInsightsSeo } from '@/utils/seo'

const navMetrics = ['ipv6', 'tls13', 'http2', 'hsts', 'csp', 'security_txt', 'certificate_verified'] as const satisfies readonly NavInsightMetricKey[]
const siteDimensions = ['country', 'group', 'nsfw', 'public_interest'] as const satisfies readonly SiteInsightDimension[]
const { locale } = useI18n()
const localePath = useLocalePath()
const destinations = overviewExploreGroups.site.slice(1)
const {
  metrics,
  overview,
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
