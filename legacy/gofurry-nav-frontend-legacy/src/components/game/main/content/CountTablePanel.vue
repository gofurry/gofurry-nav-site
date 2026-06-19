<template>
  <div class="bg-orange-50 backdrop-blur-md rounded-xl p-5">

    <!-- 标题 -->
    <div class="flex justify-between items-center mb-4">
      <h2 class="font-bold text-lg">{{ title }}</h2>
      <div
          v-if="desc"
          class="text-sm text-gray-500 hidden sm:table-cell"
      >
        {{ desc }}
      </div>
    </div>

    <!-- 表格 -->
    <table class="w-full text-sm table-fixed">
      <thead class="text-gray-600">
      <tr>
        <th class="w-28 text-center hidden sm:table-cell"></th>
        <th class="text-left">{{ t('common.game') }}</th>
        <th class="text-right">{{ t('game.panel.recentOnline') }}</th>
        <th class="text-right hidden sm:table-cell">
          {{ t('game.panel.onlinePeak') }}
        </th>
        <th class="text-right hidden sm:table-cell">
          {{ t('common.time') }}
        </th>
      </tr>
      </thead>

      <tbody>
      <tr
          v-for="item in listToShow"
          :key="item.id"
          class="group odd:bg-orange-100/50
                 hover:bg-orange-200/50 transition-colors duration-150 h-16"
      >
        <!-- 封面 -->
        <td class="px-2 hidden sm:table-cell">
          <img
              :src="item.header"
              class="h-full w-full object-cover rounded"
              alt="cover"
          />
        </td>

        <!-- 名称 -->
        <td class="px-2 font-medium truncate max-w-[200px]">
          {{ item.name }}
        </td>

        <!-- 当前在线 -->
        <td class="text-right px-2 font-semibold">
          {{ item.count_recent }}
        </td>

        <!-- 峰值 -->
        <td class="text-right px-2 font-semibold hidden sm:table-cell">
          {{ item.count_peak }}
        </td>

        <!-- 时间 -->
        <td
            class="text-right px-2 text-gray-500 text-xs truncate hidden sm:table-cell"
        >
          {{ formatTime(item.collect_time) }}
        </td>
      </tr>
      </tbody>
    </table>

    <!-- 展开 -->
    <div v-if="list.length > 5" class="text-center mt-2">
      <div
          class="cursor-pointer text-cyan-900 hover:bg-orange-100 p-2 rounded-md"
          @click="emit('toggle')"
      >
        {{ expanded ? t('common.collapse') : t('common.expand') }}
      </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { i18n } from '@/main'
import type { TopPlayerCountRecord } from '@/types/game'

const { t } = i18n.global

// props

const props = defineProps<{
  title: string
  desc?: string
  list: TopPlayerCountRecord[]
  expanded: boolean
}>()

// 展示逻辑

const emit = defineEmits(['toggle'])

const listToShow = computed(() =>
    props.expanded ? props.list.slice(0, 15) : props.list.slice(0, 5)
)

// 工具方法

function formatTime(t: number) {
  const chinaOffset = 24 * 3600
  const d = new Date((t + chinaOffset) * 1000)

  const hh = String(d.getUTCHours()).padStart(2, '0')
  const mm = String(d.getUTCMinutes()).padStart(2, '0')

  return `${hh}:${mm}`
}
</script>