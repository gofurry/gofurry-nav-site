<template>
  <section class="insight-dimension-explorer" aria-labelledby="domain-dimension-title" data-dimension-explorer>
    <div class="insight-domain-heading">
      <div><p class="insight-domain-kicker">{{ $t(`insights.metrics.${metricKey}.name`) }}</p><h2 id="domain-dimension-title">{{ $t(`insights.domain.${domain}.dimensions`) }}</h2></div>
      <div class="insight-dimension-options" :aria-label="$t('insights.dimensions.title')">
        <button v-for="option in dimensions" :key="option" type="button" :aria-pressed="option === dimension" :data-dimension="option" @click="$emit('dimension', option)">{{ $t(`insights.dimensions.names.${option}`) }}</button>
      </div>
    </div>
    <p v-if="breakdown?.slice_mode === 'overlapping'" class="insight-dimension-notice" data-overlapping>{{ $t('insights.dimensions.overlapping') }}</p>
    <div :aria-busy="loading">
      <p v-if="unavailable" class="insights-empty-state">{{ $t('insights.emptyStates.unavailable') }}</p>
      <p v-else-if="loading" class="insights-empty-state">{{ $t('insights.dimensions.loading') }}</p>
      <p v-else-if="!bars.length" class="insights-empty-state">{{ $t('insights.dimensions.empty') }}</p>
      <div v-else class="insight-dimension-bars">
        <button v-for="bar in bars" :key="bar.item.value" type="button" :data-slice="bar.item.value" :aria-pressed="selectedSlice === bar.item.value" @click="$emit('slice', bar.item.value)">
          <span class="insight-dimension-bars__label"><strong>{{ sliceLabel(bar.item) }}</strong><b>{{ domain === 'site' ? formatInsightRatio(bar.value) : integer(bar.item.population) }}</b></span>
          <progress v-if="bar.signal !== null" :value="bar.signal" :max="bar.maximum" aria-hidden="true" />
          <span v-else class="insight-dimension-bars__missing" aria-hidden="true" />
          <span class="insight-dimension-bars__context">
            <span v-if="domain === 'site'">{{ $t('insights.dimensions.known') }} {{ integer(bar.item.known) }} / {{ $t('insights.dimensions.population') }} {{ integer(bar.item.population) }}</span>
            <span v-else>{{ $t(`insights.metrics.${metricKey}.name`) }} {{ formatInsightRatio(bar.item.metric_value) }}</span>
            <span>{{ $t('insights.dimensions.coverage') }} {{ formatInsightRatio(bar.item.coverage) }}</span>
          </span>
        </button>
      </div>
    </div>

    <InsightSliceTrend v-if="selectedSlice" :metric-key="metricKey" :slice="selectedSlice" :range="range" :trend="sliceTrend" :loading="sliceLoading" :unavailable="sliceUnavailable" />

    <details v-if="breakdown?.items.length && !loading && !unavailable" class="insight-dimension-raw" data-dimension-raw>
      <summary>{{ $t('insights.domain.fullData') }}</summary>
      <div class="insight-dimension-raw__scroll">
        <table>
          <caption class="sr-only">{{ $t('insights.dimensions.title') }} · {{ $t(`insights.metrics.${metricKey}.name`) }}</caption>
          <thead><tr><th scope="col">{{ $t('insights.dimensions.slice') }}</th><th v-for="key in tableColumns" :key="key" scope="col">{{ $t(`insights.dimensions.${key}`) }}</th></tr></thead>
          <tbody>
            <tr v-for="item in breakdown.items" :key="item.value">
              <th scope="row"><button type="button" :data-table-slice="item.value" :aria-pressed="selectedSlice === item.value" @click="$emit('slice', item.value)">{{ sliceLabel(item) }}</button></th>
              <td>{{ formatInsightRatio(item.metric_value) }}</td><td>{{ formatInsightRatio(item.coverage) }}</td><td>{{ integer(item.population) }}</td><td>{{ integer(item.eligible) }}</td><td>{{ integer(item.known) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </details>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { InsightDimension, InsightDimensionBreakdown, InsightDimensionSlice, InsightDimensionTrend, InsightDomain, InsightMetricKey, InsightRange } from '@/types/insights'
import { formatInsightRatio } from '@/utils/insightDimensions'
import { insightDimensionBars } from '@/utils/insightDomain'
import InsightSliceTrend from './InsightSliceTrend.vue'

const props = defineProps<{
  domain: InsightDomain, metricKey: InsightMetricKey, dimensions: readonly InsightDimension[], dimension: InsightDimension
  selectedSlice: string | null, breakdown: InsightDimensionBreakdown | null, loading: boolean, unavailable: boolean
  range: InsightRange, sliceTrend: InsightDimensionTrend | null, sliceLoading: boolean, sliceUnavailable: boolean
}>()
defineEmits<{ dimension: [dimension: InsightDimension], slice: [value: string] }>()
const { locale } = useI18n()
const bars = computed(() => insightDimensionBars(props.breakdown, props.domain))
const tableColumns = ['metricValue', 'coverage', 'population', 'eligible', 'known'] as const
const integer = (value: number) => new Intl.NumberFormat(locale.value === 'en' ? 'en-US' : 'zh-CN').format(value)
function sliceLabel(item: InsightDimensionSlice) {
  if (props.dimension === 'country' && item.value !== 'unknown') {
    try { return new Intl.DisplayNames([locale.value === 'en' ? 'en' : 'zh'], { type: 'region' }).of(item.value) || item.value }
    catch { return item.value }
  }
  return locale.value === 'en' ? item.label_en || item.label || item.value : item.label || item.label_en || item.value
}
</script>
