<template>
  <details class="insights-data-info">
    <summary>{{ $t('insights.dataInfo.title') }}</summary>
    <div class="insights-data-info__content">
      <p>{{ $t(`insights.metrics.${metricKey}.description`) }}</p>
      <dl>
        <div>
          <dt>{{ $t('insights.dataInfo.asOf') }}</dt>
          <dd>{{ metric?.as_of || unavailable }}</dd>
        </div>
        <div>
          <dt>{{ $t('insights.dataInfo.coverage') }}</dt>
          <dd>{{ coverage }}</dd>
        </div>
        <div>
          <dt>{{ $t('insights.dataInfo.availableFrom') }}</dt>
          <dd>{{ metric?.available_from || unavailable }}</dd>
        </div>
      </dl>
      <p class="insights-data-info__caveat">{{ $t('insights.dataInfo.caveat') }}</p>
    </div>
  </details>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { InsightMetric, InsightMetricKey } from '@/types/insights'

const props = defineProps<{
  metricKey: InsightMetricKey
  metric: InsightMetric | null
}>()

const { t } = useI18n()
const unavailable = computed(() => t('insights.dataInfo.unavailable'))
const coverage = computed(() => {
  if (!props.metric) return unavailable.value
  return `${props.metric.known} / ${props.metric.eligible}`
})
</script>
