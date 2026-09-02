<template>
  <section class="insights-section insights-dimensions" aria-labelledby="insights-dimensions-title">
    <div class="insights-section__heading insights-dimensions__heading">
      <div>
        <p class="insights-eyebrow">{{ $t('insights.dimensions.eyebrow') }}</p>
        <h2 id="insights-dimensions-title">{{ $t('insights.dimensions.title') }}</h2>
      </div>
      <div class="insights-dimension-options" :aria-label="$t('insights.dimensions.title')">
        <button
          v-for="option in dimensions"
          :key="option"
          type="button"
          :class="{ 'insights-dimension-options__button--active': option === dimension }"
          :aria-pressed="option === dimension"
          :data-dimension="option"
          @click="$emit('dimension', option)"
        >
          {{ $t(`insights.dimensions.names.${option}`) }}
        </button>
      </div>
    </div>

    <p v-if="breakdown?.slice_mode === 'overlapping'" class="insights-dimensions__notice">
      {{ $t('insights.dimensions.overlapping') }}
    </p>
    <p v-if="unavailable" class="insights-empty-state">{{ $t('insights.emptyStates.unavailable') }}</p>
    <p v-else-if="loading" class="insights-empty-state">{{ $t('insights.dimensions.loading') }}</p>
    <p v-else-if="!breakdown?.items.length" class="insights-empty-state">{{ $t('insights.dimensions.empty') }}</p>
    <div v-else class="insights-dimension-table-wrap">
      <table class="insights-dimension-table">
        <thead>
          <tr>
            <th>{{ $t('insights.dimensions.slice') }}</th>
            <th>{{ $t('insights.dimensions.metricValue') }}</th>
            <th>{{ $t('insights.dimensions.coverage') }}</th>
            <th>{{ $t('insights.dimensions.population') }}</th>
            <th>{{ $t('insights.dimensions.eligible') }}</th>
            <th>{{ $t('insights.dimensions.known') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in breakdown.items" :key="item.value">
            <td>
              <button
                type="button"
                :class="{ 'insights-dimension-table__slice--active': selectedSlice === item.value }"
                :aria-pressed="selectedSlice === item.value"
                :data-slice="item.value"
                @click="$emit('slice', item.value)"
              >
                {{ sliceLabel(item) }}
              </button>
            </td>
            <td>{{ percent(item.metric_value) }}</td>
            <td>{{ percent(item.coverage) }}</td>
            <td>{{ integer(item.population) }}</td>
            <td>{{ integer(item.eligible) }}</td>
            <td>{{ integer(item.known) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { InsightDimension, InsightDimensionBreakdown, InsightDimensionSlice } from '@/types/insights'
import { formatInsightRatio } from '@/utils/insightDimensions'

const props = defineProps<{
  dimensions: readonly InsightDimension[]
  dimension: InsightDimension
  selectedSlice: string | null
  breakdown: InsightDimensionBreakdown | null
  loading?: boolean
  unavailable?: boolean
}>()

defineEmits<{
  dimension: [dimension: InsightDimension]
  slice: [value: string]
}>()

const { locale } = useI18n()

function sliceLabel(item: InsightDimensionSlice) {
  if (props.dimension === 'country' && item.value !== 'unknown') {
    try {
      return new Intl.DisplayNames([locale.value === 'en' ? 'en' : 'zh'], { type: 'region' }).of(item.value) || item.value
    } catch {
      return item.value
    }
  }
  return locale.value === 'en'
    ? item.label_en || item.label || item.value
    : item.label || item.label_en || item.value
}

function percent(value: number | null) {
  return formatInsightRatio(value)
}

function integer(value: number) {
  return new Intl.NumberFormat(locale.value === 'en' ? 'en-US' : 'zh-CN').format(value)
}
</script>
