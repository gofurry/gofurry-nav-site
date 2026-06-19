<template>
  <div class="bg-orange-50 backdrop-blur-md rounded-xl p-5">
    <div class="flex justify-between items-center mb-4">
      <h2 class="font-bold text-lg">{{ title }}</h2>
      <div class="text-sm text-gray-500 hidden sm:table-cell">
        {{ desc }}
      </div>
    </div>

    <table class="w-full text-sm table-fixed">
      <thead class="text-gray-600">
      <tr>
        <th class="w-28 hidden sm:table-cell"></th>
        <th class="text-left">{{ t("common.game") }}</th>
        <th class="text-right hidden sm:table-cell">{{ t("game.panel.global") }}</th>
        <th class="text-right">{{ t("game.panel.china") }}</th>
        <th class="text-right hidden sm:table-cell">{{ t("game.panel.discount") }}</th>
      </tr>
      </thead>

      <tbody>
      <tr
          v-for="item in listToShow"
          :key="item.id"
          class="group odd:bg-orange-100/50  hover:bg-orange-200/50 h-16"
      >
        <td class="px-2 hidden sm:table-cell">
          <img :src="item.header" class="h-full w-full object-cover" />
        </td>
        <td class="px-2 font-medium truncate">{{ item.name }}</td>
        <td class="text-right px-2 hidden sm:table-cell">
          {{ formatPrice(item.global_price, true) }}
        </td>
        <td class="text-right px-2">
          {{ formatPrice(item.china_price, false) }}
        </td>
        <td
            class="text-right px-2 hidden sm:table-cell font-bold"
            :class="item.discount > 0 ? 'text-red-500' : 'text-gray-400'"
        >
          <span v-if="item.discount > 0">-{{ item.discount }}%</span>
          <span v-else>{{ t("game.panel.none") }}</span>
        </td>
      </tr>
      </tbody>
    </table>

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
import type { PriceRecord } from '@/types/game'
import { i18n } from '@/main'

const { t } = i18n.global

const props = defineProps<{
  title: string
  desc?: string
  list: PriceRecord[]
  expanded: boolean
}>()

const emit = defineEmits<{
  (e: 'toggle'): void
}>()

const listToShow = computed(() =>
    props.expanded ? props.list.slice(0, 15) : props.list.slice(0, 5)
)

function formatPrice(v: number, isGlobal: boolean) {
  if (v === 0) return t('game.panel.none')
  const price = (v / 100).toFixed(2)
  return isGlobal ? `$${price}` : `¥${price}`
}
</script>