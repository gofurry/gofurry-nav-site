<template>
  <section v-if="slice" class="insights-section insights-slice-trend" aria-labelledby="insights-slice-trend-title">
    <div class="insights-section__heading">
      <div>
        <p class="insights-eyebrow">{{ $t('insights.dimensions.selectedSlice') }}</p>
        <h2 id="insights-slice-trend-title">{{ label }}</h2>
      </div>
      <p class="insights-slice-trend__range">{{ $t(`insights.ranges.${range}`) }}</p>
    </div>
    <p class="insights-slice-trend__availability">
      {{ $t('insights.dimensions.availability', {
        from: trend?.available_from || '—',
        through: trend?.available_through || '—',
      }) }}
    </p>
    <div class="insights-chart-shell" :aria-busy="loading">
      <div ref="chartRef" class="insights-chart" :class="{ 'insights-chart--visible': points.length >= 2 && !unavailable }" />
      <p v-if="unavailable" class="insights-empty-state insights-chart-state">{{ $t('insights.emptyStates.unavailable') }}</p>
      <p v-else-if="!loading && points.length === 0" class="insights-empty-state insights-chart-state">{{ $t('insights.dimensions.sliceEmpty') }}</p>
      <p v-else-if="!loading && points.length === 1" class="insights-empty-state insights-chart-state">{{ $t('insights.emptyStates.trendOne') }}</p>
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
import type { InsightDimensionTrend, InsightRange } from '@/types/insights'

type EChartsInstance = import('echarts').ECharts

const props = defineProps<{
  slice: string | null
  range: InsightRange
  trend: InsightDimensionTrend | null
  loading?: boolean
  unavailable?: boolean
}>()

const chartRef = ref<HTMLElement | null>(null)
const chart = ref<EChartsInstance | null>(null)
const themeStore = useThemeStore()
const { locale, t } = useI18n()
const isDark = computed(() => themeStore.theme === 'dark')
const points = computed(() => props.trend?.points ?? [])
const label = computed(() => {
  const selected = props.trend?.slice
  if (!selected) return props.slice || ''
  return locale.value === 'en'
    ? selected.label_en || selected.label || selected.value
    : selected.label || selected.label_en || selected.value
})
let resizeObserver: ResizeObserver | null = null
let active = false

async function renderChart() {
  if (!active || !chartRef.value || points.value.length < 2 || props.unavailable) return
  const echarts = await import('echarts')
  if (!active || !chartRef.value) return
  if (!chart.value) chart.value = echarts.init(chartRef.value, undefined, { renderer: 'canvas' })
  const dark = isDark.value
  const series = points.value.map(point => ({
    value: point.metric_value === null ? null : Number((point.metric_value * 100).toFixed(4)),
    date: point.date,
    coverage: point.coverage,
    population: point.population,
    eligible: point.eligible,
    known: point.known,
  }))
  chart.value.setOption({
    animation: false,
    grid: { top: 20, right: 18, bottom: 48, left: 52 },
    tooltip: {
      trigger: 'axis',
      confine: true,
      backgroundColor: dark ? 'rgba(15, 23, 42, .96)' : 'rgba(255, 250, 242, .98)',
      borderColor: dark ? 'rgba(125, 211, 252, .28)' : 'rgba(154, 75, 36, .22)',
      textStyle: { color: dark ? '#e2e8f0' : '#292524' },
      formatter(params: Array<{ data: typeof series[number] }>) {
        const point = params?.[0]?.data
        if (!point) return ''
        const value = point.value === null ? '—' : `${point.value.toFixed(1)}%`
        const coverage = point.coverage === null ? '—' : `${(point.coverage * 100).toFixed(1)}%`
        return `${point.date}<br/>${t('insights.trend.value')}: ${value}<br/>${t('insights.trend.coverage')}: ${coverage}<br/>${t('insights.dimensions.population')}: ${point.population}<br/>${t('insights.dimensions.known')}: ${point.known}`
      },
    },
    xAxis: {
      type: 'category', boundaryGap: false, data: points.value.map(point => point.date),
      axisLine: { lineStyle: { color: dark ? 'rgba(148, 163, 184, .20)' : 'rgba(126, 92, 58, .14)' } },
      axisTick: { show: false }, axisLabel: { color: dark ? '#94a3b8' : '#786f68', hideOverlap: true, margin: 14 },
    },
    yAxis: {
      type: 'value', min: 0, max: 100, axisLabel: { color: dark ? '#94a3b8' : '#786f68', formatter: '{value}%' },
      splitLine: { lineStyle: { color: dark ? 'rgba(148, 163, 184, .20)' : 'rgba(126, 92, 58, .14)' } },
    },
    series: [{
      type: 'line', data: series, connectNulls: false, symbol: 'circle', symbolSize: 6,
      showSymbol: points.value.length <= 31,
      lineStyle: { width: 3, color: dark ? '#c4b5fd' : '#7c3aed' },
      itemStyle: { color: dark ? '#c4b5fd' : '#7c3aed' },
      areaStyle: { color: dark ? 'rgba(196, 181, 253, .16)' : 'rgba(124, 58, 237, .12)' },
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

watch(() => [props.slice, props.range, props.trend, props.unavailable, isDark.value, locale.value], async () => {
  await nextTick()
  await renderChart()
}, { deep: true })

onBeforeUnmount(() => {
  active = false
  resizeObserver?.disconnect()
  chart.value?.dispose()
  chart.value = null
})
</script>
