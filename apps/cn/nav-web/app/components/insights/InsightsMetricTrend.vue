<template>
  <section class="insights-section insights-trend" aria-labelledby="insights-trend-title">
    <div class="insights-section__heading insights-trend__heading">
      <div>
        <p class="insights-eyebrow">{{ $t(`insights.metrics.${metricKey}.name`) }}</p>
        <h2 id="insights-trend-title">{{ $t('insights.trend.title') }}</h2>
      </div>
      <div class="insights-ranges" :aria-label="$t('insights.trend.title')">
        <button
          v-for="option in ranges"
          :key="option"
          type="button"
          :class="{ 'insights-ranges__button--active': option === range }"
          :aria-pressed="option === range"
          :data-range="option"
          @click="$emit('range', option)"
        >
          {{ $t(`insights.ranges.${option}`) }}
        </button>
      </div>
    </div>

    <div class="insights-chart-shell" :aria-busy="loading">
      <div ref="chartRef" class="insights-chart" :class="{ 'insights-chart--visible': points.length >= 2 && !unavailable }" />
      <p v-if="unavailable" class="insights-empty-state insights-chart-state">
        {{ $t('insights.emptyStates.unavailable') }}
      </p>
      <p v-else-if="!loading && points.length === 0" class="insights-empty-state insights-chart-state">
        {{ $t('insights.emptyStates.trendEmpty') }}
      </p>
      <p v-else-if="!loading && points.length === 1" class="insights-empty-state insights-chart-state">
        {{ $t('insights.emptyStates.trendOne') }}
      </p>
      <div v-if="loading" class="insights-chart-loading" role="status">
        <span class="insights-chart-loading__dot" aria-hidden="true" />
        {{ $t('insights.trend.loading') }}
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { InsightMetricKey, InsightRange, InsightTrendPoint } from '@/types/insights'

type EChartsInstance = import('echarts').ECharts

const props = defineProps<{
  metricKey: InsightMetricKey
  range: InsightRange
  points: InsightTrendPoint[]
  loading?: boolean
  unavailable?: boolean
}>()

defineEmits<{
  range: [range: InsightRange]
}>()

const ranges: InsightRange[] = ['30d', '90d', 'all']
const chartRef = ref<HTMLElement | null>(null)
const chart = ref<EChartsInstance | null>(null)
const themeStore = useThemeStore()
const { locale, t } = useI18n()
const isDark = computed(() => themeStore.theme === 'dark')
let resizeObserver: ResizeObserver | null = null
let active = false

async function renderChart() {
  if (!active || !chartRef.value || props.points.length < 2 || props.unavailable) return
  const echarts = await import('echarts')
  if (!active || !chartRef.value) return
  if (!chart.value) chart.value = echarts.init(chartRef.value, undefined, { renderer: 'canvas' })

  const dark = isDark.value
  const colors = {
    line: dark ? '#7dd3fc' : '#9a4b24',
    axis: dark ? '#94a3b8' : '#786f68',
    split: dark ? 'rgba(148, 163, 184, .20)' : 'rgba(126, 92, 58, .14)',
    tooltip: dark ? 'rgba(15, 23, 42, .96)' : 'rgba(255, 250, 242, .98)',
    tooltipBorder: dark ? 'rgba(125, 211, 252, .28)' : 'rgba(154, 75, 36, .22)',
    text: dark ? '#e2e8f0' : '#292524',
    area: dark ? 'rgba(125, 211, 252, .18)' : 'rgba(154, 75, 36, .16)',
  }
  const series = props.points.map(point => ({
    value: point.value === null ? null : Number((point.value * 100).toFixed(4)),
    date: point.date,
    coverage: point.coverage,
  }))

  chart.value.setOption({
    animation: false,
    grid: { top: 20, right: 18, bottom: 48, left: 52 },
    tooltip: {
      trigger: 'axis',
      confine: true,
      backgroundColor: colors.tooltip,
      borderColor: colors.tooltipBorder,
      textStyle: { color: colors.text },
      formatter(params: Array<{ data: { value: number | null, date: string, coverage: number | null } }>) {
        const point = params?.[0]?.data
        if (!point) return ''
        const value = point.value === null ? '—' : `${point.value.toFixed(1)}%`
        const coverage = point.coverage === null ? '—' : `${(point.coverage * 100).toFixed(1)}%`
        return `${point.date}<br/>${t('insights.trend.value')}: ${value}<br/>${t('insights.trend.coverage')}: ${coverage}`
      },
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: props.points.map(point => point.date),
      axisLine: { lineStyle: { color: colors.split } },
      axisTick: { show: false },
      axisLabel: { color: colors.axis, hideOverlap: true, margin: 14 },
    },
    yAxis: {
      type: 'value',
      min: 0,
      max: 100,
      axisLabel: { color: colors.axis, formatter: '{value}%' },
      splitLine: { lineStyle: { color: colors.split } },
    },
    series: [{
      type: 'line',
      data: series,
      connectNulls: false,
      symbol: 'circle',
      symbolSize: 6,
      showSymbol: props.points.length <= 31,
      lineStyle: { width: 3, color: colors.line },
      itemStyle: { color: colors.line },
      areaStyle: { color: colors.area },
    }],
  }, true)
}

onMounted(async () => {
  active = true
  resizeObserver = new ResizeObserver(() => chart.value?.resize())
  if (chartRef.value) resizeObserver.observe(chartRef.value)
  await nextTick()
  await renderChart()
})

watch(
  () => [props.metricKey, props.range, props.points, props.unavailable, isDark.value, locale.value],
  async () => {
    await nextTick()
    await renderChart()
  },
  { deep: true },
)

onBeforeUnmount(() => {
  active = false
  resizeObserver?.disconnect()
  chart.value?.dispose()
  chart.value = null
})
</script>
