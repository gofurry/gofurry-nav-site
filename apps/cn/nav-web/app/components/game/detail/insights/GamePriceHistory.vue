<template>
  <section class="game-insights-chart-card" data-price-history>
    <h3>{{ $t('insights.entity.priceHistoryRegion', { region: $t(`insights.regions.${region}`) }) }}</h3>

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
import type { GameInsightPricePoint, GameInsightRegion } from '@/types/insights'
import { formatMinorAmount, priceSegmentKey, publicPriceDisplay } from '@/utils/insightPrices'

type EChartsInstance = import('echarts').ECharts

const props = defineProps<{
  points: GameInsightPricePoint[]
  region: GameInsightRegion
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

function pointLines(point: GameInsightPricePoint) {
  const display = publicPriceDisplay(point)
  if (display.kind === 'free') {
    return [`${t('insights.entity.currentPrice')}: ${t('insights.entity.priceFree')}`]
  }
  if (display.kind === 'priced') {
    const current = point.currency ? formatMinorAmount(display.amount, point.currency, locale.value) : t('insights.entity.priceStatusUnknown')
    const lines = [`${t('insights.entity.currentPrice')}: ${current}`]
    if (point.initial_amount !== null && point.currency) {
      lines.push(`${t('insights.entity.originalPriceLabel')}: ${formatMinorAmount(point.initial_amount, point.currency, locale.value)}`)
    }
    if (point.discount_percent !== null && point.discount_percent > 0) {
      lines.push(`${t('insights.entity.discount')}: -${point.discount_percent}%`)
    }
    return lines
  }
  const state = point.state === 'unknown'
    ? t('insights.entity.priceStatusUnknown')
    : t('insights.entity.priceMissingShort')
  return [`${t('insights.entity.priceStatus')}: ${state}`]
}

async function renderChart() {
  if (!active || !chartRef.value || props.points.length < 2 || props.unavailable) return
  const echarts = await import('echarts')
  if (!active || !chartRef.value) return
  if (!chart.value) chart.value = echarts.init(chartRef.value, undefined, { renderer: 'canvas' })

  const dark = isDark.value
  const colors = {
    line: dark ? '#a7f3d0' : '#15803d',
    axis: dark ? '#94a3b8' : '#786f68',
    split: dark ? 'rgba(148, 163, 184, .20)' : 'rgba(126, 92, 58, .14)',
    tooltip: dark ? 'rgba(15, 23, 42, .96)' : 'rgba(255, 250, 242, .98)',
    border: dark ? 'rgba(167, 243, 208, .28)' : 'rgba(21, 128, 61, .22)',
    text: dark ? '#e2e8f0' : '#292524',
  }
  const segments: Array<{ key: string, indexes: number[] }> = []
  let current: { key: string, indexes: number[] } | null = null
  props.points.forEach((point, index) => {
    const key = priceSegmentKey(point)
    const previous = index > 0 ? props.points[index - 1] : null
    const consecutive = previous ? (new Date(`${point.date}T00:00:00Z`).getTime() - new Date(`${previous.date}T00:00:00Z`).getTime()) === 86400000 : true
    if (!key) { current = null; return }
    if (!current || current.key !== key || !consecutive) { current = { key, indexes: [] }; segments.push(current) }
    current.indexes.push(index)
  })

  chart.value.setOption({
    animation: false,
    grid: { top: 20, right: 18, bottom: 48, left: 62 },
    tooltip: {
      trigger: 'axis',
      confine: true,
      backgroundColor: colors.tooltip,
      borderColor: colors.border,
      textStyle: { color: colors.text },
      formatter(params: Array<{ axisValue?: string, data?: { value: number | null, point: GameInsightPricePoint } | null }>) {
        const entry = params.find(item => item.data?.point)?.data
        const axisDate = params.find(item => item.axisValue)?.axisValue
        const point = entry?.point ?? props.points.find(item => item.date === axisDate)
        return point ? [point.date, ...pointLines(point)].join('<br/>') : ''
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
      axisLabel: { color: colors.axis },
      splitLine: { lineStyle: { color: colors.split } },
    },
    series: segments.map(segment => ({
      type: 'line',
      data: props.points.map((point, index) => {
        if (!segment.indexes.includes(index)) return null
        const display = publicPriceDisplay(point)
        return { value: display.kind === 'priced' ? display.amount / 100 : 0, point }
      }),
      connectNulls: false,
      symbol: 'circle',
      symbolSize: 6,
      showSymbol: props.points.length <= 31,
      lineStyle: { width: 3, color: colors.line },
      itemStyle: { color: colors.line },
    })),
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
  () => [props.points, props.unavailable, isDark.value, locale.value],
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
