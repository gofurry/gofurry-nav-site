<template>
  <div class="insights-page insights-domain-page insights-game-domain" data-domain="game" :data-selected-metric="selectedMetric" :data-selected-dimension="selectedDimension" :data-selected-slice="selectedSlice || ''">
    <main class="insights-container">
      <EcosystemNavigation context="game" />
      <InsightsDomainHeader domain="game" :entity-count="overview?.entity_count ?? null" :generated-at="overview?.generated_at ?? null" />
      <p v-if="overviewUnavailable" class="insights-empty-state">{{ $t('insights.emptyStates.unavailable') }}</p>

      <section class="insight-domain-pulse" aria-labelledby="domain-pulse-title" data-domain-pulse>
        <div class="insight-domain-heading"><div><p class="insight-domain-kicker">{{ $t('insights.editorial.gamesKicker') }}</p><h2 id="domain-pulse-title">{{ $t('insights.domain.pulse') }}</h2></div></div>
        <InsightsGamePulse :panel="panelData.panel" />
      </section>

      <InsightMetricRail :metrics="metrics" :metrics-by-key="metricsByKey" :selected-metric="selectedMetric" :groups="metricGroups" @select="selectMetric" />

      <InsightTrendWorkspace domain="game" :metric-key="selectedMetric" :metric="selectedMetricData" :range="selectedRange" :points="trend?.points ?? []" :loading="trendLoading" :unavailable="trendUnavailable" @range="selectRange" />

      <InsightDimensionExplorer domain="game" :metric-key="selectedMetric" :dimensions="dimensions" :dimension="selectedDimension" :selected-slice="selectedSlice" :breakdown="breakdown" :loading="breakdownLoading" :unavailable="breakdownUnavailable" :range="selectedRange" :slice-trend="sliceTrend" :slice-loading="sliceTrendLoading" :slice-unavailable="sliceTrendUnavailable" @dimension="selectDimension" @slice="selectSlice" />

      <InsightDomainActivity domain="game" :items="recentChanges" :unavailable="overviewUnavailable" />

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
import InsightsGamePulse from '@/components/insights/domain/InsightsGamePulse.vue'
import type { GameV2PanelRecord } from '@/types/game'
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
import { getGameHomePanel, getGameInsightsBreakdown, getGameInsightsOverview, getGameInsightsSliceTrend, getGameInsightsTrend } from '@/services/game'
import type { GameInsightDimension, GameInsightMetricKey } from '@/types/insights'
import { buildInsightsSeo } from '@/utils/seo'

const gameMetrics = ['free', 'windows', 'mac', 'linux'] as const satisfies readonly GameInsightMetricKey[]
const gameDimensions = ['primary_tag', 'tag'] as const satisfies readonly GameInsightDimension[]
const { locale } = useI18n()
const localePath = useLocalePath()
const destinations = overviewExploreGroups.game.slice(1)
const panelSnapshot = useAsyncData(() => `insights:game:panel:${locale.value}`, async () => {
  const [result] = await Promise.allSettled([getGameHomePanel(locale.value)])
  return { panel: result.status === 'fulfilled' ? result.value : null }
}, { default: () => ({ panel: null as GameV2PanelRecord | null }) })
const metricGroups = [
  { label: 'insights.domain.businessModel', keys: ['free'] as const },
  { label: 'insights.domain.platforms', keys: ['windows', 'mac', 'linux'] as const },
]
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
  domain: 'game',
  defaultMetric: 'free',
  metrics: gameMetrics,
  getOverview: getGameInsightsOverview,
  getTrend: getGameInsightsTrend,
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
  domain: 'game',
  metric: selectedMetric,
  range: selectedRange,
  defaultDimension: 'primary_tag',
  dimensions: gameDimensions,
  getBreakdown: getGameInsightsBreakdown,
  getSliceTrend: getGameInsightsSliceTrend,
})
const { data: panelData } = await panelSnapshot
const seo = computed(() => buildInsightsSeo('games', locale.value))

useSeoMeta({
  title: () => seo.value.title,
  description: () => seo.value.description,
  ogTitle: () => seo.value.title,
  ogDescription: () => seo.value.description,
})
</script>
