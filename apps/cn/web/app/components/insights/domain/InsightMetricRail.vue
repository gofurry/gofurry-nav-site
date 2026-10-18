<template>
  <div class="insight-metric-rail mb-9 flex max-w-full gap-6 overflow-x-auto overscroll-x-contain" :aria-label="$t('insights.metricsLabel')" data-metric-rail>
    <div v-for="(group, index) in railGroups" :key="index" class="insight-metric-rail__group">
      <p v-if="group.label" class="insight-domain-kicker">{{ $t(group.label) }}</p>
      <div class="insight-metric-rail__items flex justify-between gap-6">
        <button v-for="key in group.keys" :key="key" class="relative flex min-w-max flex-[1_0_128px] flex-col gap-2 text-left" type="button" :data-metric-key="key" :aria-pressed="selectedMetric === key" @click="$emit('select', key)" @focus="reveal">
          <span>{{ $t(`insights.metrics.${key}.name`) }}</span>
          <strong>{{ formatInsightRatio(metricsByKey.get(key)?.value ?? null) }}</strong>
          <small>{{ formatOverviewDelta(metricsByKey.get(key)?.delta_30d) }} <span>/ 30d</span></small>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { InsightMetric, InsightMetricKey } from '@/types/insights'
import { formatInsightRatio } from '@/utils/insightDimensions'
import { formatOverviewDelta } from '@/utils/insightOverview'
const props = defineProps<{
  metrics: readonly InsightMetricKey[]
  metricsByKey: Map<InsightMetricKey, InsightMetric>
  selectedMetric: InsightMetricKey
  groups?: { label: string, keys: readonly InsightMetricKey[] }[]
}>()
defineEmits<{ select: [metric: InsightMetricKey] }>()
const railGroups = computed(() => props.groups ?? [{ label: '', keys: props.metrics }])
function reveal(event: FocusEvent) {
  (event.currentTarget as HTMLElement).scrollIntoView({ block: 'nearest', inline: 'nearest' })
}
</script>
