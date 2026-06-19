<template>
  <div class="left-panel h-full flex flex-col gap-4">
    <!-- 上部数据卡片 -->
    <div class="flex-1 bg-[#001c3d] rounded-md overflow-hidden flex flex-col">
      <div class="p-3 text-center font-bold text-lg text-white">{{t("dashboard.collectResult")}}</div>
      <div class="h-1 bg-gradient-to-r from-blue-500 to-purple-500"></div>
      <div class="p-4">
        <ul class="grid grid-cols-3 gap-4 text-center">
          <li>
            <h2 class="text-xl font-bold text-yellow-400">{{ commonStat?.dns_count ?? '-' }}</h2>
            <span class="text-blue-200 text-xs">DNS {{t("dashboard.logCount")}}</span>
          </li>
          <li>
            <h2 class="text-xl font-bold text-yellow-400">{{ commonStat?.http_count ?? '-' }}</h2>
            <span class="text-blue-200 text-xs">HTTP {{t("dashboard.logCount")}}</span>
          </li>
          <li>
            <h2 class="text-xl font-bold text-yellow-400">{{ commonStat?.ping_count ?? '-' }}</h2>
            <span class="text-blue-200 text-xs">Ping {{t("dashboard.logCount")}}</span>
          </li>
        </ul>
      </div>
      <div class="h-1 bg-gradient-to-r from-blue-500 to-purple-500"></div>
    </div>

    <!-- 下部数据卡片 -->
    <div class="flex-1 bg-[#001c3d] rounded-md overflow-hidden flex flex-col">
      <div class="p-3 text-center font-bold text-lg text-white">{{t("dashboard.visitStats")}}</div>
      <div class="h-1 bg-gradient-to-r from-blue-500 to-purple-500"></div>
      <div class="p-4">
        <ul class="grid grid-cols-3 gap-4 text-center">
          <li>
            <h2 class="text-xl font-bold text-yellow-400">{{ viewsCount?.total ?? '-' }}</h2>
            <span class="text-blue-200 text-xs">{{t("dashboard.totalVisits")}}</span>
          </li>
          <li>
            <h2 class="text-xl font-bold text-yellow-400">{{ viewsCount?.year_count ?? '-' }}</h2>
            <span class="text-blue-200 text-xs">{{t("dashboard.yearVisits")}}</span>
          </li>
          <li>
            <h2 class="text-xl font-bold text-yellow-400">{{ viewsCount?.month_count ?? '-' }}</h2>
            <span class="text-blue-200 text-xs">{{t("dashboard.monthVisits")}}</span>
          </li>
          <li>
            <h2 class="text-xl font-bold text-yellow-400">{{ lastDayCount }}</h2>
            <span class="text-blue-200 text-xs">{{t("dashboard.dayVisits")}}</span>
          </li>
          <li>
            <h2 class="text-xl font-bold text-yellow-400">{{ ((commonStat?.non_profit_business_ratio ?? 0) * 100).toFixed(2) }}%</h2>
            <span class="text-blue-200 text-xs">{{t("dashboard.publicBenefitRate")}}</span>
          </li>
          <li>
            <h2 class="text-xl font-bold text-yellow-400">{{ ((commonStat?.sfw_nsfw_ratio ?? 0) * 100).toFixed(2) }}%</h2>
            <span class="text-blue-200 text-xs">{{t("dashboard.nsfwRate")}}</span>
          </li>
        </ul>
      </div>
      <div class="h-1 bg-gradient-to-r from-blue-500 to-purple-500"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {i18n} from "@/main.ts";

const { t } = i18n.global

const props = defineProps({
  commonStat: Object,
  viewsCount: Object
})

const lastDayCount = computed(() => {
  const counts = props.viewsCount?.count
  return counts && counts.length ? counts[counts.length - 1] : '-'
})
</script>

<style scoped>
.left-panel {
  overflow-y: auto;
  padding: 4px;
}
</style>
