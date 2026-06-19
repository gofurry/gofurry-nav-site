<template>
  <div class="h-full">
    <div class="flex h-full gap-4">
      <!-- 左侧环形图 -->
      <div class="w-1/2 bg-[#001c3d] rounded-md overflow-hidden lg:block hidden">
        <div class="p-3 text-center font-bold text-white text-lg">
          {{t("dashboard.regionDistribution")}}
        </div>
        <div class="h-1 bg-gradient-to-r from-blue-500 to-purple-500"></div>

        <div class="p-4 h-[calc(100%-60px)] grid grid-cols-3 gap-4">
          <div id="bt01" class="h-full"></div>
          <div id="bt02" class="h-full"></div>
          <div id="bt03" class="h-full"></div>
        </div>

        <div class="h-1 bg-gradient-to-r from-blue-500 to-purple-500"></div>
      </div>

      <!-- 右侧折线图 -->
      <div class="w-full lg:w-1/2 bg-[#001c3d] rounded-md overflow-hidden">
        <div class="p-3 text-center font-bold text-white text-lg">
          {{t("dashboard.visit7Days")}}
        </div>
        <div class="h-1 bg-gradient-to-r from-blue-500 to-purple-500"></div>

        <div class="p-2 h-[calc(100%-60px)]" id="echart4"></div>

        <div class="h-1 bg-gradient-to-r from-blue-500 to-purple-500"></div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted, watch, nextTick } from 'vue'
import * as echarts from 'echarts'
import {i18n} from "@/main.ts";

const { t } = i18n.global

const props = defineProps({
  viewsCount: { type: Object, required: true },
  cityStat: { type: Object, required: true },
  countryStat: { type: Object, required: true },
  provinceStat: { type: Object, required: true }
})

let chartInstances = {}

const initPieCharts = () => {
  const createPieOption = (dataMap, title) => {
    const data = Object.entries(dataMap)
        .map(([name, value]) => ({ name, value }))
        .sort((a, b) => b.value - a.value)
        .slice(0, 12) // 前12个
    return {
      title: {
        text: title,
        left: 'center',
        top: '2%',
        textStyle: { color: '#fff', fontSize: 14, fontWeight: 'bold' }
      },
      tooltip: { trigger: 'item', formatter: '{b}<br/>{c} '+t("common.times")+' ({d}%)' },
      legend: { show: false },
      series: [
        {
          name: title,
          type: 'pie',
          radius: ['45%', '70%'],
          center: ['50%', '55%'],
          avoidLabelOverlap: true,
          minAngle: 12,
          label: {
            show: true,
            position: 'outside',
            color: '#fff',
            fontSize: 11,
            formatter: '{b|{b}}\n{per|{d}%}',
            rich: {
              b: { color: '#fff', fontSize: 11, lineHeight: 16 },
              per: { color: '#00eaff', fontSize: 10 }
            }
          },
          labelLine: {
            show: true,
            length: 10,
            length2: 15,
            smooth: true
          },
          itemStyle: {
            borderColor: '#001c3d',
            borderWidth: 2
          },
          data
        }
      ]
    }
  }

  // 初始化三个环形图
  const charts = [
    { id: 'bt01', data: props.cityStat.region_map, title: t("dashboard.cityDistribution") },
    { id: 'bt02', data: props.provinceStat.region_map, title: t("dashboard.provinceDistribution") },
    { id: 'bt03', data: props.countryStat.region_map, title: t("dashboard.countryDistribution") }
  ]

  charts.forEach(({ id, data, title }) => {
    const dom = document.getElementById(id)
    if (!dom || !data) return
    if (chartInstances[id]) chartInstances[id].dispose()

    const instance = echarts.init(dom)
    instance.setOption(createPieOption(data, title))
    chartInstances[id] = instance
  })
}

// 折线图初始化保持不变
const initLineChart = () => {
  const chartDom = document.getElementById('echart4')
  if (!chartDom) return

  const { date, count } = props.viewsCount || { date: [], count: [] }

  if (chartInstances.echart4) chartInstances.echart4.dispose()
  const instance = echarts.init(chartDom)

  const option = {
    tooltip: { trigger: 'axis' },
    grid: { left: '3%', right: '5%', bottom: '8%', top: '10%', containLabel: true },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: date,
      axisLine: { lineStyle: { color: 'rgba(255,255,255,0.3)' } },
      axisLabel: { color: 'rgba(255,255,255,0.7)', fontSize: 10 }
    },
    yAxis: {
      type: 'value',
      axisLine: { show: false },
      splitLine: { lineStyle: { color: 'rgba(255,255,255,0.1)' } },
      axisLabel: { color: 'rgba(255,255,255,0.6)', fontSize: 10 }
    },
    series: [
      {
        name: t("dashboard.visits"),
        type: 'line',
        smooth: true,
        symbol: 'circle',
        symbolSize: 6,
        itemStyle: { color: '#00d2ff' },
        lineStyle: { color: '#00d2ff', width: 2 },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(0,210,255,0.5)' },
            { offset: 1, color: 'rgba(0,210,255,0)' }
          ])
        },
        data: count
      }
    ]
  }

  instance.setOption(option)
  chartInstances.echart4 = instance
}

const handleResize = () => {
  Object.values(chartInstances).forEach(c => c.resize())
}

onMounted(async () => {
  await nextTick()
  initPieCharts()
  initLineChart()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  Object.values(chartInstances).forEach(c => c.dispose())
})

watch(
    () => [props.cityStat, props.countryStat, props.provinceStat, props.viewsCount],
    async () => {
      await nextTick()
      initPieCharts()
      initLineChart()
    },
    { deep: true }
)
</script>
