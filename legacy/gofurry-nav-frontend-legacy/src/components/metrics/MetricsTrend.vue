<template>
  <section class="rounded-xl flex flex-col gap-3 bg-white/60 backdrop-blur-xs">
    <!-- 标题 -->
    <div class="flex items-center justify-between p-4">
      <h3 class="font-bold text-orange-800">{{t("metrics.nodeTrend")}}</h3>

      <div class="flex gap-2 text-xs">
        <button
            v-for="t in timeRanges"
            :key="t.key"
            @click="activeRange = t.key"
            class="px-3 py-1 rounded-lg font-semibold transition"
            :class="activeRange === t.key
            ? 'bg-white/70 text-orange-800'
            : 'text-gray-700 hover:bg-white/50'"
        >
          {{ t.label }}
        </button>
      </div>
    </div>

    <!-- 图标 -->
    <div class="grid grid-cols-1 xl:grid-cols-3 gap-4">
      <div ref="cpuRef" class="h-[260px] rounded-lg" />
      <div ref="memRef" class="h-[260px] rounded-lg" />
      <div ref="connectRef" class="h-[260px] rounded-lg" />
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, nextTick, onBeforeUnmount } from 'vue'
import * as echarts from 'echarts'
import { getPromMetricsHistory } from '@/utils/api/stat'
import type { PromMetricsHistoryModel, MetricsModel } from '@/types/stat'
import { i18n } from '@/main.ts'

const { t } = i18n.global

type RangeKey = 'twenty_minutes' | 'one_hour' | 'twenty_hours'

// 🔹 默认显示 20 分钟
const activeRange = ref<RangeKey>('twenty_minutes')

const timeRanges = [
  { key: 'twenty_minutes', label: '20 '+t("common.minutes") },
  { key: 'one_hour', label: '1 '+t("common.hours") },
  { key: 'twenty_hours', label: '20 '+t("common.hours") }
] as const

const cpuRef = ref<HTMLDivElement | null>(null)
const memRef = ref<HTMLDivElement | null>(null)
const connectRef = ref<HTMLDivElement | null>(null)

let cpuChart: echarts.ECharts | null = null
let memChart: echarts.ECharts | null = null
let connectChart: echarts.ECharts | null = null

const history = ref<PromMetricsHistoryModel | null>(null)

// 初始化
onMounted(async () => {
  history.value = await getPromMetricsHistory()
  await nextTick()
  initCharts()
  renderAll()
  setupResizeObserver()
})

// watch 时间切换
watch(activeRange, () => {
  renderAll()
})

// 初始化图表
function initCharts() {
  if (cpuRef.value) cpuChart = echarts.init(cpuRef.value, undefined, { renderer: 'canvas' })
  if (memRef.value) memChart = echarts.init(memRef.value, undefined, { renderer: 'canvas' })
  if (connectRef.value) connectChart = echarts.init(connectRef.value, undefined, { renderer: 'canvas' })
}

// 渲染所有图表
function renderAll() {
  if (!history.value) return

  renderLineChart(cpuChart, t("metrics.cpuUsage"), history.value.cpu[activeRange.value], true)
  renderLineChart(memChart, t("metrics.memoryUsage"), history.value.memory[activeRange.value], false, true)
  renderLineChart(connectChart, t("metrics.tcpConnections"), history.value.connect[activeRange.value], false)
}

// 数据提取
function extract(data: MetricsModel[]) {
  return {
    times: data.map(i => formatTime(i.time)),
    values: data.map(i => i.usage)
  }
}

// 时间格式化
function formatTime(ts: number) {
  const d = new Date(ts * 1000)
  return `${d.getHours()}:${String(d.getMinutes()).padStart(2, '0')}`
}

// 字节大小转换
function formatBytes(v: number) {
  if (v >= 1024 ** 3) return (v / 1024 ** 3).toFixed(2) + ' GB'
  return (v / 1024 ** 2).toFixed(2) + ' MB'
}

// 绘制折线图
function renderLineChart(
    chart: echarts.ECharts | null,
    title: string,
    data: MetricsModel[],
    isPercent: boolean,
    isMemory = false
) {
  if (!chart) return
  chart.clear()

  const { times, values } = extract(data)

  chart.setOption({
    title: {
      text: title,
      left: 'center',
      textStyle: { color: '#364153', fontSize: 13 }
    },
    tooltip: {
      trigger: 'item',
      axisPointer: { type: 'none' },
      confine: true,
      backgroundColor: 'rgba(58, 56, 56, 0.85)',
      borderWidth: 0,
      borderColor: 'transparent',
      padding: [8, 12],
      textStyle: { color: '#ccc', fontSize: 12, lineHeight: 18 },
      formatter: (params: any) => {
        const d = params.data
        if (!d) return t("game.panel.none")
        return `
        <div style="line-height:1.5">
          <div><strong>`+t("common.time")+`:</strong> ${times[params.dataIndex]}</div>
          <div><strong>`+t("common.result")+`:</strong> ${
            isMemory ? formatBytes(d) : isPercent ? d.toFixed(2)+'%' : d
        }</div>
        </div>
      `
      }
    },
    grid: { left: '3%', right: '3%', top: '15%', bottom: '10%', containLabel: true },
    xAxis: {
      type: 'category',
      data: times,
      axisLabel: { color: '#364153', fontSize: 12 },
      axisLine: { show: false },
      axisTick: { show: false }
    },
    yAxis: {
      type: 'value',
      axisLabel: {
        color: '#364153',
        formatter: (v: number) => isMemory ? formatBytes(v) : isPercent ? v.toFixed(2)+'%' : v
      },
      splitLine: { lineStyle: { type: 'dashed', color: '#888' } }
    },
    series: [{
      type: 'line',
      data: values,
      smooth: true,
      symbol: 'circle',
      symbolSize: 5,
      itemStyle: { color: '#c2410c' },
      lineStyle: { width: 2, color: '#c2410c' },
      areaStyle: {
        opacity: 0.2,
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: 'rgba(249,115,22,0.4)' },
          { offset: 1, color: 'rgba(249,115,22,0.05)' }
        ])
      },
      emphasis: { itemStyle: { symbolSize: 8, color: '#ea580c' } }
    }]
  })

  setupHover(chart, values)
}

// hover 吸附最近点
function setupHover(chart: echarts.ECharts, values: number[]) {
  chart.getZr().off('mousemove')
  chart.getZr().on('mousemove', (event) => {
    const pointInPixel: [number, number] = [event.offsetX ?? 0, event.offsetY ?? 0]
    let nearestIndex = -1
    let minDist = Infinity

    values.forEach((v, i) => {
      const pixel = chart.convertToPixel({ seriesIndex: 0 }, [i, v])
      if (!Array.isArray(pixel) || pixel.length < 2) return

      const [px, py] = pixel as [number, number]
      const dist = Math.hypot(px - pointInPixel[0], py - pointInPixel[1])

      if (dist < minDist && dist <= 10) {
        minDist = dist
        nearestIndex = i
      }
    })

    if (nearestIndex !== -1) {
      chart.dispatchAction({ type: 'highlight', seriesIndex: 0, dataIndex: nearestIndex })
      chart.dispatchAction({ type: 'showTip', seriesIndex: 0, dataIndex: nearestIndex })
    } else {
      chart.dispatchAction({ type: 'downplay', seriesIndex: 0 })
      chart.dispatchAction({ type: 'hideTip' })
    }
  })
}

// ResizeObserver 自动适配
let resizeObserver: ResizeObserver | null = null
function setupResizeObserver() {
  resizeObserver = new ResizeObserver(() => {
    cpuChart?.resize()
    memChart?.resize()
    connectChart?.resize()
  })
  if (cpuRef.value) resizeObserver.observe(cpuRef.value)
  if (memRef.value) resizeObserver.observe(memRef.value)
  if (connectRef.value) resizeObserver.observe(connectRef.value)
}

onBeforeUnmount(() => {
  if (resizeObserver) {
    if (cpuRef.value) resizeObserver.unobserve(cpuRef.value)
    if (memRef.value) resizeObserver.unobserve(memRef.value)
    if (connectRef.value) resizeObserver.unobserve(connectRef.value)
  }
  resizeObserver = null
  cpuChart?.dispose()
  memChart?.dispose()
  connectChart?.dispose()
})
</script>
