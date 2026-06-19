<template>
  <div class="rounded-xl flex flex-col gap-3 bg-white/60 backdrop-blur-xs">
    <h3 class="font-bold text-orange-800 p-4">
      {{ title }}
    </h3>

    <div class="space-y-2 rounded-xl max-h-80 overflow-auto text-sm">
      <!-- 表头 -->
      <div
          class="mb-1 mx-1 flex justify-between items-center px-3 py-2
               rounded-lg text-xs font-semibold text-gray-700
               bg-white/60 sticky top-0 z-10"
      >
        <span>{{t("metrics.apiPath")}}</span>
        <span>{{t("metrics.avgResponse")}}</span>
      </div>

      <!-- 数据行 -->
      <div
          v-for="(v, k) in data"
          :key="k"
          class="font-bold m-1 flex justify-between items-center px-3 py-2 rounded-lg hover:bg-white/40"
      >
        <span class="truncate text-gray-700">{{ k }}</span>
        <span class="font-mono text-orange-800">
          {{ formatLatency(v) }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { i18n } from '@/main.ts'

const { t } = i18n.global
defineProps<{
  title: string
  data: Record<string, string>
}>()

function formatLatency(val: string) {
  const s = Number(val)
  if (isNaN(s)) return '-'
  if (s < 1) {
    return `${Math.round(s * 1000)} ms`
  }
  return `${s.toFixed(2)} s`
}
</script>
