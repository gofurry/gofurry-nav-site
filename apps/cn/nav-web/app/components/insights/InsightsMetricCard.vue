<template>
  <button
    type="button"
    class="insights-metric-card"
    :class="{ 'insights-metric-card--selected': selected }"
    :aria-pressed="selected"
    :data-metric-key="metricKey"
    @click="$emit('select', metricKey)"
  >
    <span class="insights-metric-card__label">{{ $t(`insights.metrics.${metricKey}.name`) }}</span>
    <strong>{{ formatPercent(metric?.value ?? null) }}</strong>
    <span class="insights-metric-card__meta">
      <span>30d</span>
      <span>{{ formatDelta(metric?.delta_30d ?? null) }}</span>
    </span>
    <span class="insights-metric-card__date">
      {{ metric?.as_of ? $t('insights.metrics.asOf', { date: metric.as_of }) : $t('insights.dataInfo.unavailable') }}
    </span>
  </button>
</template>

<script setup lang="ts">
import type { InsightMetric, InsightMetricKey } from '@/types/insights'

defineProps<{
  metricKey: InsightMetricKey
  metric: InsightMetric | null
  selected: boolean
}>()

defineEmits<{
  select: [metricKey: InsightMetricKey]
}>()

function formatPercent(value: number | null) {
  if (value === null || !Number.isFinite(value)) return '—'
  return `${(value * 100).toFixed(1)}%`
}

function formatDelta(value: number | null) {
  if (value === null || !Number.isFinite(value)) return '—'
  if (value === 0) return '0.0%'
  return `${value > 0 ? '↑' : '↓'} ${Math.abs(value * 100).toFixed(1)}%`
}
</script>
