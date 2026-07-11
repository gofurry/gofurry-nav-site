<script setup lang="ts">
import { LineChart } from 'echarts/charts'
import { GridComponent, LegendComponent, TooltipComponent } from 'echarts/components'
import * as echarts from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { MetricSeries } from '../types'

echarts.use([LineChart, GridComponent, LegendComponent, TooltipComponent, CanvasRenderer])

const props = withDefaults(
  defineProps<{
    series: MetricSeries[]
    height?: number
    smooth?: boolean
  }>(),
  {
    height: 220,
    smooth: true,
  },
)

const chartEl = ref<HTMLDivElement | null>(null)
let chart: echarts.ECharts | undefined

const hasData = computed(() => props.series.some((item) => item.points.length > 0))

function render() {
  if (!chartEl.value || !hasData.value) return
  if (!chart) chart = echarts.init(chartEl.value, undefined, { renderer: 'canvas' })
  chart.setOption(
    {
      animationDuration: 260,
      backgroundColor: 'transparent',
      color: ['#d7e0e8', '#98bfa9', '#d0ad72', '#d18888', '#9fb5c8', '#b9aa92'],
      grid: { top: 28, right: 12, bottom: 28, left: 42 },
      legend: {
        top: 0,
        right: 0,
        icon: 'roundRect',
        itemWidth: 10,
        itemHeight: 6,
        textStyle: { color: '#8b96a2', fontSize: 11 },
      },
      tooltip: {
        trigger: 'axis',
        backgroundColor: 'rgba(9, 11, 14, 0.96)',
        borderColor: 'rgba(176, 193, 206, 0.22)',
        borderWidth: 1,
        textStyle: { color: '#d8e1e8' },
        axisPointer: { lineStyle: { color: 'rgba(216, 225, 232, 0.25)' } },
        valueFormatter: (value: number) => {
          const unit = props.series.find((item) => item.points.some((point) => point.value === value))?.unit || ''
          return `${Number(value).toFixed(unit === 'ms' ? 0 : 1)}${unit}`
        },
      },
      xAxis: {
        type: 'time',
        axisLabel: { color: '#687481', fontSize: 11 },
        axisLine: { lineStyle: { color: 'rgba(176, 193, 206, 0.14)' } },
        axisTick: { show: false },
        splitLine: { show: false },
      },
      yAxis: {
        type: 'value',
        axisLabel: { color: '#687481', fontSize: 11 },
        axisLine: { show: false },
        axisTick: { show: false },
        splitLine: { lineStyle: { color: 'rgba(176, 193, 206, 0.09)' } },
      },
      series: props.series.map((item) => ({
        name: item.name,
        type: 'line',
        showSymbol: false,
        smooth: props.smooth,
        lineStyle: { width: 2 },
        areaStyle: { opacity: 0.08 },
        data: item.points.map((point) => [point.timestamp, point.value]),
      })),
    },
    true,
  )
}

function resize() {
  chart?.resize()
}

onMounted(() => {
  void nextTick(render)
  window.addEventListener('resize', resize)
})

watch(
  () => props.series,
  () => void nextTick(render),
  { deep: true },
)

onBeforeUnmount(() => {
  window.removeEventListener('resize', resize)
  chart?.dispose()
  chart = undefined
})
</script>

<template>
  <div class="chart-shell" :style="{ height: `${height}px` }">
    <div v-if="!hasData" class="chart-empty">暂无趋势数据</div>
    <div v-show="hasData" ref="chartEl" class="h-full w-full"></div>
  </div>
</template>
