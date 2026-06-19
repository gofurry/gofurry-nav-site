<template>
  <div class="flex justify-center items-center gap-2 select-none">

    <!-- 页码 -->
    <template v-for="(item, idx) in displayPages" :key="idx">
      <span
          v-if="item.type === 'page'"
          class="px-3 py-1 rounded-md text-sm cursor-pointer transition"
          :class="item.page === currentPage
          ? 'bg-orange-400 text-white'
          : 'bg-orange-50 hover:bg-orange-200'"
          @click="changePage(item.page!)"
      >
        {{ item.page }}
      </span>

      <span
          v-else
          class="px-3 py-1 rounded-md text-sm
               bg-orange-50 text-gray-500 cursor-pointer"
          @click="openJump"
      >
        ...
      </span>
    </template>

    <span class="ml-2 text-xs text-gray-500">
      {{ t("common.total") }} {{ total }} {{ t("common.record") }}
    </span>

    <!-- 跳页 -->
    <div
        v-if="showJump"
        class="fixed inset-0 bg-black/20 z-50
             flex items-center justify-center"
    >
      <div class="bg-orange-50 rounded-lg p-4 w-64 space-y-3">
        <div class="text-sm font-semibold">{{ t("game.search.jumpPage") }}</div>

        <input
            v-model.number="jumpPage"
            :min="1"
            :max="totalPages"
            class="w-full px-2 py-1 rounded-md text-sm bg-orange-100 focus:outline-none focus:ring-2 focus:ring-orange-200"
        />

        <div class="flex justify-end gap-2">
          <button class="px-3 py-1 rounded text-sm cursor-pointer hover:bg-orange-100"
                  @click="showJump = false">
            {{ t("common.cancel") }}
          </button>
          <button
              class="px-3 py-1 rounded text-sm bg-orange-400 text-white cursor-pointer hover:bg-orange-300"
              @click="confirmJump"
          >
            {{ t("common.confirm") }}
          </button>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { i18n } from '@/main.ts'

const { t } = i18n.global
const props = defineProps<{
  currentPage: number
  totalPages: number
  total: number
}>()

const emit = defineEmits<{
  (e: 'page-change', page: number): void
}>()

// 计算页码
const displayPages = computed(() => {
  const pages = new Set<number>()

  pages.add(1)
  pages.add(props.totalPages)

  for (let i = props.currentPage - 1; i <= props.currentPage + 1; i++) {
    if (i > 1 && i < props.totalPages) pages.add(i)
  }

  const sorted = [...pages].sort((a, b) => a - b)
  const result: Array<{ type: 'page' | 'ellipsis'; page?: number }> = []

  for (let i = 0; i < sorted.length; i++) {
    if (i > 0 && sorted[i]! - sorted[i - 1]! > 1) {
      result.push({ type: 'ellipsis' })
    }
    result.push({ type: 'page', page: sorted[i]! })
  }

  return result
})

// 切页
const changePage = (page: number) => {
  if (page !== props.currentPage) emit('page-change', page)
}

// 跳页
const showJump = ref(false)
const jumpPage = ref(props.currentPage)

watch(() => props.currentPage, v => (jumpPage.value = v))

const openJump = () => (showJump.value = true)

const confirmJump = () => {
  if (jumpPage.value >= 1 && jumpPage.value <= props.totalPages) {
    emit('page-change', jumpPage.value)
    showJump.value = false
  }
}
</script>
