<template>
  <section class="game-insights-chart-card" data-player-history>
    <h3>{{ $t('insights.entity.playerHistory') }}</h3>

    <div class="game-insights-chart-shell" :aria-busy="loading">
      <div ref="chartRef" class="game-insights-chart" :class="{ 'game-insights-chart--visible': points.length >= 2 && !unavailable }" />
      <div v-if="unavailable" class="entity-insights-empty game-insights-chart-state">
        <p>{{ $t('insights.emptyStates.unavailable') }}</p>
        <button type="button" class="entity-insights-retry" @click="$emit('retry')">
          {{ $t('insights.entity.retry') }}
        </button>
      </div>
      <p v-else-if="!loading && points.length === 0" class="entity-insights-empty game-insights-chart-state">
        {{ $t('insights.emptyStates.trendEmpty') }}
      </p>
      <p v-else-if="!loading && points.length === 1" class="entity-insights-empty game-insights-chart-state">
        {{ $t('insights.emptyStates.trendOne') }}
      </p>
      <div v-if="loading" class="game-insights-chart-loading" role="status">
        <span class="insights-chart-loading__dot" aria-hidden="true" />
        {{ $t('insights.trend.loading') }}
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { GameDetailInsightRange, GameInsightPlayerPoint } from '@/types/insights'
import { formatGameInsightAxisDate } from '@/utils/insightHistoryRanges'

type EChartsInstance = import('echarts').ECharts

const props = defineProps<{
  points: GameInsightPlayerPoint[]
  range: GameDetailInsightRange
  loading?: boolean
  unavailable?: boolean
}>()

defineEmits<{ retry: [] }>()

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
    peak: dark ? '#7dd3fc' : '#2563eb',
    average: dark ? '#94a3b8' : '#64748b',
    axis: dark ? '#94a3b8' : '#786f68',
    split: dark ? 'rgba(148, 163, 184, .20)' : 'rgba(126, 92, 58, .14)',
    tooltip: dark ? 'rgba(15, 23, 42, .96)' : 'rgba(255, 250, 242, .98)',
    border: dark ? 'rgba(125, 211, 252, .28)' : 'rgba(217, 119, 6, .22)',
    text: dark ? '#e2e8f0' : '#292524',
    area: dark ? 'rgba(125, 211, 252, .12)' : 'rgba(37, 99, 235, .10)',
  }
  const peakSeries = props.points.map(point => ({ value: point.max, point }))
  const averageSeries = props.points.map(point => ({ value: point.avg, point }))

  chart.value.setOption({
    animation: false,
    grid: { top: 48, right: 18, bottom: 48, left: 58 },
    legend: {
      top: 0,
      left: 0,
      itemWidth: 18,
      itemHeight: 3,
      textStyle: { color: colors.axis },
      data: [t('insights.entity.dailyPeak'), t('insights.entity.dailyAverage')],
    },
    tooltip: {
      trigger: 'axis',
      confine: true,
      backgroundColor: colors.tooltip,
      borderColor: colors.border,
      textStyle: { color: colors.text },
      formatter(params: { axisValue?: string, data?: { value: number | null, point: GameInsightPlayerPoint } } | Array<{ axisValue?: string, data?: { value: number | null, point: GameInsightPlayerPoint } }>) {
        const entries = Array.isArray(params) ? params : [params]
        const axisDate = entries.find(item => item.axisValue)?.axisValue
        const point = entries.find(item => item.data?.point)?.data?.point
          ?? props.points.find(item => item.date === axisDate)
        if (!point) return ''
        const average = point.avg === null ? t('insights.entity.dataUnavailable') : new Intl.NumberFormat(locale.value).format(point.avg)
        return `${point.date}<br/>${t('insights.entity.playerTooltipPeak')}: ${new Intl.NumberFormat(locale.value).format(point.max)}<br/>${t('insights.entity.playerTooltipAverage')}: ${average}`
      },
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: props.points.map(point => point.date),
      axisLine: { lineStyle: { color: colors.split } },
      axisTick: { show: false },
      axisLabel: {
        color: colors.axis,
        hideOverlap: true,
        margin: 14,
        formatter: (value: string) => formatGameInsightAxisDate(value, props.range),
      },
    },
    yAxis: {
      type: 'value',
      min: 0,
      minInterval: 1,
      axisLabel: { color: colors.axis },
      splitLine: { lineStyle: { color: colors.split } },
    },
    series: [
      {
        name: t('insights.entity.dailyPeak'),
        type: 'line',
        data: peakSeries,
        connectNulls: false,
        symbol: 'circle',
        symbolSize: 6,
        showSymbol: props.points.length <= 31,
        lineStyle: { width: 3, color: colors.peak },
        itemStyle: { color: colors.peak },
        areaStyle: { color: colors.area },
      },
      {
        name: t('insights.entity.dailyAverage'),
        type: 'line',
        data: averageSeries,
        connectNulls: false,
        symbol: 'circle',
        symbolSize: 5,
        showSymbol: props.points.length <= 31,
        lineStyle: { width: 2, color: colors.average },
        itemStyle: { color: colors.average },
      },
    ],
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
  () => [props.points, props.range, props.unavailable, isDark.value, locale.value],
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
