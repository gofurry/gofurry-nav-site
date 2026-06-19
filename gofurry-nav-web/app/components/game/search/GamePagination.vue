<template>
  <div class="gf-pagination game-search-pagination flex items-center justify-center gap-2 select-none">

    <!-- 页码 -->
    <div class="game-search-pagination-pages">
      <span
          v-for="(item, idx) in displayPages"
          :key="item.type === 'page' ? `page-${item.page}` : `ellipsis-${idx}`"
          class="gf-pagination__button game-search-page-button"
          :class="item.type === 'page' && item.page === currentPage
            ? 'gf-pagination__button--active game-search-page-button--active'
            : 'game-search-page-button--idle'"
          @click="item.type === 'page' ? changePage(item.page!) : openJump()"
      >
        {{ item.type === 'page' ? item.page : '...' }}
      </span>
    </div>

    <span class="gf-pagination__total game-search-pagination-total ml-2">
      {{ t("common.total") }} {{ total }} {{ t("common.record") }}
    </span>

    <!-- 跳页 -->
    <div
        v-if="showJump"
        class="game-search-jump-overlay fixed inset-0 z-50
             flex items-center justify-center"
    >
      <div class="game-search-jump-dialog w-64 space-y-3 p-4">
        <div class="game-search-jump-title text-sm font-semibold">{{ t("game.search.jumpPage") }}</div>

        <input
            v-model.number="jumpPage"
            :min="1"
            :max="totalPages"
            class="game-search-jump-input w-full px-2 py-1 text-sm focus:outline-none"
        />

        <div class="flex justify-end gap-2">
          <button class="game-search-jump-action game-search-jump-action--ghost"
                  @click="showJump = false">
            {{ t("common.cancel") }}
          </button>
          <button
              class="game-search-jump-action game-search-jump-action--primary"
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
import { i18n } from '@/main'

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
