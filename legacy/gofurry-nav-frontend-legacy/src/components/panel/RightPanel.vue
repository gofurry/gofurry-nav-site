<template>
  <div class="h-full flex flex-col gap-4">
    <!-- 上部图表 -->
    <div class="flex-1 bg-[#001c3d] rounded-md overflow-hidden flex flex-col">
      <div class="p-3 text-center text-white font-bold text-lg">
        {{t("dashboard.categoryStats")}}
      </div>
      <div class="h-1 bg-gradient-to-r from-blue-500 to-purple-500"></div>
      <div class="flex-1 flex items-stretch">
        <div class="p-2 w-full h-full" id="echart3"></div>
      </div>
      <div class="h-1 bg-gradient-to-r from-blue-500 to-purple-500"></div>
    </div>

    <!-- 下部滚动列表 -->
    <div class="flex-1 bg-[#001c3d] rounded-md overflow-hidden flex flex-col">
      <div class="p-3 text-center text-white font-bold text-lg">{{t("dashboard.latestCollected")}}</div>
      <div class="h-1 bg-gradient-to-r from-blue-500 to-purple-500"></div>

      <div class="wrap flex-1 overflow-hidden">
        <ul ref="marqueeList" class="flex-col">
          <li
              v-for="(site, index) in props.latestSites"
              :key="site.name + site.create_time"
              class="p-2 border-b border-[rgba(66,153,225,0.1)]"
          >
            <p
                class="grid gap-2 text-sm text-white"
                style="grid-template-columns: 30px 100px 60px 140px;"
            >
              <span class="text-gray-400 text-center">{{ index + 1 }}</span>
              <span class="truncate">{{ site.name }}</span>
              <span class="text-center">{{ site.country }}</span>
              <span class="text-green-400 text-right">{{ site.create_time }}</span>
            </p>
          </li>
        </ul>
      </div>

      <div class="h-1 bg-gradient-to-r from-blue-500 to-purple-500"></div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted, ref, watch } from 'vue'
import * as echarts from 'echarts'
import {i18n} from "@/main.ts";

const { t } = i18n.global

const props = defineProps({
  groupCount: { type: Array, default: () => [] },
  latestSites: { type: Array, default: () => [] }
})

let echart3 = null
let marqueeInterval = null

// 初始化图表
const initEchart3 = () => {
  const chartDom = document.getElementById('echart3')
  if (!chartDom || !props.groupCount.length) return

  echart3 = echarts.init(chartDom)

  const option = {
    grid: { containLabel: true, top: 20, left: 0, right: 15, bottom: 20 },
    tooltip: {
      show: true,
      formatter: (params) => {
        return t('dashboard.siteGroupCount', {
          count: params.value ?? params.data ?? params.c
        })
      }
    },
    xAxis: {
      type: 'value',
      axisLine: { show: false },
      axisLabel: { color: 'rgba(255,255,255,0.6)' },
      splitLine: { show: false }
    },
    yAxis: {
      type: 'category',
      data: props.groupCount.map(item => item.name),
      axisLine: { show: false },
      axisLabel: { color: 'rgba(255,255,255,0.8)' },
      axisTick: { show: false }
    },
    series: [{
      type: 'bar',
      data: props.groupCount.map(item => item.count),
      itemStyle: { color: '#58c485' },
      label: { show: true, position: 'insideRight', color: '#fff' }
    }]
  }

  echart3.setOption(option)
}

// 滚动列表逻辑
const marqueeList = ref(null)
const initMarquee = () => {
  if (!marqueeList.value) return
  const container = marqueeList.value.parentElement
  const content = marqueeList.value
  const clone = content.cloneNode(true)
  container.appendChild(clone)

  let top = 0
  const speed = 0.3
  const height = content.offsetHeight

  marqueeInterval = setInterval(() => {
    top += speed
    if (top >= height) top = 0
    content.style.transform = `translateY(-${top}px)`
    clone.style.transform = `translateY(-${top}px)`
  }, 16)
}

const handleResize = () => {
  echart3 && echart3.resize()
}

onMounted(() => {
  initEchart3()
  initMarquee()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  echart3 && echart3.dispose()
  marqueeInterval && clearInterval(marqueeInterval)
})

watch(() => props.groupCount, () => {
  initEchart3()
})
</script>