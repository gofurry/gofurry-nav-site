<template>
  <div class="insight-trend-workspace" :data-trend-domain="domain">
    <InsightMetricTrend :metric-key="metricKey" :range="range" :points="points" :loading="loading" :unavailable="unavailable" @range="$emit('range', $event)" />
    <aside class="insight-trend-context" :aria-label="$t('insights.domain.current')">
      <p class="insight-domain-kicker">{{ $t('insights.domain.current') }}</p>
      <strong class="insight-trend-context__value">{{ formatInsightRatio(metric?.value ?? null) }}</strong>
      <p class="insight-trend-context__delta">{{ formatOverviewDelta(metric?.delta_30d) }} <span>{{ $t('insights.editorial.deltaWindow') }}</span></p>
      <dl>
        <div><dt>{{ $t('insights.trend.coverage') }}</dt><dd>{{ formatInsightRatio(metric?.coverage ?? null) }}</dd></div>
        <template v-if="domain === 'site'">
          <div><dt>{{ $t('insights.dimensions.known') }}</dt><dd>{{ metric?.known ?? '—' }}</dd></div>
          <div><dt>{{ $t('insights.dimensions.eligible') }}</dt><dd>{{ metric?.eligible ?? '—' }}</dd></div>
        </template>
        <div><dt>{{ $t('insights.dataInfo.asOf') }}</dt><dd>{{ metric?.as_of || '—' }}</dd></div>
        <div v-if="domain === 'site'"><dt>{{ $t('insights.dataInfo.availableFrom') }}</dt><dd>{{ metric?.available_from || '—' }}</dd></div>
      </dl>
    </aside>
  </div>
</template>

<script setup lang="ts">
import type { InsightDomain, InsightMetric, InsightMetricKey, InsightRange, InsightTrendPoint } from '@/types/insights'
import { formatInsightRatio } from '@/utils/insightDimensions'
import { formatOverviewDelta } from '@/utils/insightOverview'
import InsightMetricTrend from './InsightMetricTrend.vue'
defineProps<{ domain: InsightDomain, metricKey: InsightMetricKey, metric: InsightMetric | null, range: InsightRange, points: InsightTrendPoint[], loading: boolean, unavailable: boolean }>()
defineEmits<{ range: [range: InsightRange] }>()
</script>
