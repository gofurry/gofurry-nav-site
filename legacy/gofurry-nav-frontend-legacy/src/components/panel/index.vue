<template>
  <div
      class="panel-container rounded-xl w-full h-screen bg-[#000b18] text-white overflow-hidden
      bg-no-repeat bg-cover bg-center"
      :style="{ backgroundImage: `url(https://qcdn.go-furry.com/nav/stat-bg/panel/bg.jpg)` }"
  >
    <Header />

    <!-- 加载动画 -->
    <Loading v-if="isLoading" />

    <!-- 主体内容 -->
    <div v-if="!isLoading" class="p-4">
      <div class="flex h-[65vh] gap-4">
        <div class="w-1/4 lg:block hidden">
          <LeftPanel
              :commonStat="commonStat || {}"
              :viewsCount="viewsCount || {}"
          />
        </div>
        <div class="w-full lg:w-2/4">
          <CenterPanel
              :commonStat="commonStat"
              :cityStat="cityStat"
          />
        </div>

        <div class="w-1/4 lg:block hidden">
          <RightPanel
              :groupCount="groupCount"
              :latestSites="latestSites"
          />
        </div>
      </div>

      <div class="h-[23vh] mt-4">
        <BottomPanel
            :viewsCount="viewsCount"
            :cityStat="cityStat"
            :countryStat="countryStat"
            :provinceStat="provinceStat"
        />
      </div>
    </div>
  </div>
</template>


<script setup lang="ts">
import { ref, onMounted } from 'vue'
import Loading from '@/components/panel/Loading.vue'
import Header from '@/components/panel/Header.vue'
import LeftPanel from '@/components/panel/LeftPanel.vue'
import CenterPanel from '@/components/panel/CenterPanel.vue'
import RightPanel from '@/components/panel/RightPanel.vue'
import BottomPanel from '@/components/panel/BottomPanel.vue'
import { useLangStore } from '@/store/langStore.ts'

// API
import {
  getGroupCount,
  getViewsCount,
  getCityStat,
  getCountryStat,
  getProvinceStat,
  getSiteCommonStat,
  getLatestSiteList,
  getLatestPingList
} from '@/utils/api/stat'

// 引入类型
import type {
  GroupCount,
  ViewsCount,
  RegionStat,
  CommonStat,
  SiteModel,
  PingModel
} from '@/types/stat'

// 加载状态
const isLoading = ref(true)

const langStore = useLangStore()

// 定义响应式数据
const groupCount = ref<GroupCount[]>([])
const viewsCount = ref<ViewsCount | null>(null)
const cityStat = ref<RegionStat | null>(null)
const countryStat = ref<RegionStat | null>(null)
const provinceStat = ref<RegionStat | null>(null)
const commonStat = ref<CommonStat | null>(null)
const latestSites = ref<SiteModel[]>([])
const latestPings = ref<PingModel[]>([])

onMounted(async () => {
  try {
    isLoading.value = true

    // 并行请求所有接口
    const lang = langStore.lang
    const [
      groupRes,
      viewsRes,
      cityRes,
      countryRes,
      provinceRes,
      commonRes,
      latestSitesRes,
      latestPingsRes
    ] = await Promise.all([
      getGroupCount(lang),
      getViewsCount(),
      getCityStat(),
      getCountryStat(),
      getProvinceStat(),
      getSiteCommonStat(),
      getLatestSiteList(lang),
      getLatestPingList()
    ])

    // 存入响应式变量
    groupCount.value = groupRes
    viewsCount.value = viewsRes
    cityStat.value = cityRes
    countryStat.value = countryRes
    provinceStat.value = provinceRes
    commonStat.value = commonRes
    latestSites.value = latestSitesRes
    latestPings.value = latestPingsRes


  } catch (error) {
    console.error('数据请求失败：', error)
  } finally {
    isLoading.value = false
  }
})
</script>

<style scoped>
</style>
