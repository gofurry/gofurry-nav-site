<template>
  <div class="center-panel h-full flex flex-col">
    <!-- 顶部数据条 -->
    <div class="mb-4 p-3 rounded-md bg-gradient-to-r from-[#001c3d] to-[#002e5e] shadow-lg">
      <div>
        <ul class="grid grid-cols-3 gap-2">
          <li class="text-center text-2xl font-bold text-yellow-400 counter">
            {{ commonStat?.site_count ?? '-' }}
          </li>
          <li class="text-center text-2xl font-bold text-yellow-400 counter">
            {{ commonStat?.domain_count ?? '-' }}
          </li>
          <li class="text-center text-2xl font-bold text-yellow-400 counter">
            {{ (commonStat?.site_reach_rate ?? 0).toFixed(2) }}%
          </li>
        </ul>
      </div>
      <div class="mt-2">
        <ul class="grid grid-cols-3 gap-2">
          <li class="text-center text-blue-200 text-sm">{{t("dashboard.siteCount")}}</li>
          <li class="text-center text-blue-200 text-sm">{{t("dashboard.domainCount")}}</li>
          <li class="text-center text-blue-200 text-sm">{{t("dashboard.siteAvailability")}}</li>
        </ul>
      </div>
    </div>

    <!-- 地图容器 -->
    <div class="flex-1 p-3 rounded-md relative overflow-hidden">
      <div class="w-full h-full relative">
        <div class="absolute inset-0 flex items-center justify-center animate-spin-clockwise">
          <img src="https://qcdn.go-furry.com/nav/stat-bg/panel/lbx.png" alt="" class="w-[80%] h-[80%] object-contain">
        </div>

        <div class="absolute inset-0 flex items-center justify-center animate-spin-counter">
          <img src="https://qcdn.go-furry.com/nav/stat-bg/panel/jt.png" alt="" class="w-[70%] h-[70%] object-contain">
        </div>

        <div class="absolute inset-0 flex items-center justify-center opacity-50">
          <img src="https://qcdn.go-furry.com/nav/stat-bg/panel/map.png" alt="" class="w-[60%] h-[60%] object-contain">
        </div>

        <div class="absolute inset-0" ref="mapContainer"></div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import * as echarts from 'echarts'
import chinaMapData from '@/assets/json/china.json'
import { geoCoordMap } from '@/utils/const'
import {i18n} from "@/main.ts";

const { t } = i18n.global

// Props
const props = defineProps({
  commonStat: Object,
  cityStat: Object
})

let mapChart = null
const mapContainer = ref(null)

// 转换格式
const getSeriesData = (cityStat) => {
  if (!cityStat?.region_map) return []
  return Object.entries(cityStat.region_map)
      .map(([name, value]) => {
        const coord = geoCoordMap[name]
        return coord ? { name, value: [...coord, value] } : null
      })
      .filter(Boolean)
}

// 初始化地图
const initMapChart = () => {
  if (!mapContainer.value) return
  mapChart && mapChart.dispose()
  mapChart = echarts.init(mapContainer.value)
  echarts.registerMap('china', chinaMapData)

  const option = {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'item',
      formatter: params => `${params.name}<br/>`+t("dashboard.visits")+`: ${params.value[2] ?? params.value}`
    },
    geo: {
      map: 'china',
      roam: true,
      label: { show: false },
      itemStyle: {
        normal: { areaColor: '#0c213a', borderColor: '#065aab', borderWidth: 1.5 },
        emphasis: { areaColor: '#0958d9' }
      }
    },
    series: [
      {
        name: '城市访问量',
        type: 'effectScatter',
        coordinateSystem: 'geo',
        data: getSeriesData(props.cityStat),
        symbolSize: val => 6,  // 点大小
        showEffectOn: 'render',
        rippleEffect: { period: 6, scale: 3, brushType: 'stroke' }, // 涟漪
        itemStyle: { color: '#ffeb7b', shadowBlur: 10, shadowColor: '#fff' },
        label: { formatter: '{b}', position: 'right', show: true, color: '#ffeb7b' },
        zlevel: 2
      }
    ]
  }

  mapChart.setOption(option)
}

const handleResize = () => mapChart && mapChart.resize()

onMounted(() => {
  setTimeout(initMapChart, 100)
  window.addEventListener('resize', handleResize)
})

watch(() => props.cityStat, () => {
  if (mapChart) initMapChart()
}, { deep: true })

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  mapChart && mapChart.dispose()
})
</script>

<style scoped>
.counter {
  animation: countUp 2s ease-out forwards;
}
@keyframes countUp {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.animate-spin-clockwise { animation: spinClockwise 20s linear infinite; }
.animate-spin-counter { animation: spinCounter 20s linear infinite; }
@keyframes spinClockwise { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
@keyframes spinCounter { from { transform: rotate(0deg); } to { transform: rotate(-360deg); } }
</style>
