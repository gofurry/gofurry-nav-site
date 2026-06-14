<template>
  <div class="game-search-pagination flex items-center justify-center gap-2 select-none">

    <!-- 页码 -->
    <template v-for="(item, idx) in displayPages" :key="idx">
      <span
          v-if="item.type === 'page'"
          class="game-search-page-button"
          :class="item.page === currentPage
          ? 'game-search-page-button--active'
          : 'game-search-page-button--idle'"
          @click="changePage(item.page!)"
      >
        {{ item.page }}
      </span>

      <span
          v-else
          class="game-search-page-button game-search-page-button--idle"
          @click="openJump"
      >
        ...
      </span>
    </template>

    <span class="ml-2 text-xs text-stone-500/80 dark:text-slate-300/60">
      {{ t("common.total") }} {{ total }} {{ t("common.record") }}
    </span>

    <!-- 跳页 -->
    <div
        v-if="showJump"
        class="fixed inset-0 bg-black/20 z-50
             flex items-center justify-center"
    >
      <div class="game-search-jump-dialog w-64 space-y-3 rounded-lg p-4">
        <div class="text-sm font-semibold">{{ t("game.search.jumpPage") }}</div>

        <input
            v-model.number="jumpPage"
            :min="1"
            :max="totalPages"
            class="game-search-jump-input w-full rounded-md px-2 py-1 text-sm focus:outline-none"
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

<style scoped>
.game-search-page-button {
  min-width: 2rem;
  border-radius: 999px;
  padding: 0.28rem 0.72rem;
  cursor: pointer;
  text-align: center;
  font-size: 0.84rem;
  transition: background-color 180ms ease, color 180ms ease, border-color 180ms ease;
}

.game-search-page-button--idle {
  border: 1px solid rgba(126, 92, 58, 0.12);
  background: rgba(255, 250, 242, 0.40);
  color: rgba(124, 45, 18, 0.74);
}

.game-search-page-button--idle:hover {
  border-color: rgba(180, 96, 24, 0.32);
  background: rgba(255, 239, 213, 0.68);
  color: rgba(99, 39, 15, 0.96);
}

.game-search-page-button--active {
  border: 1px solid rgba(126, 92, 58, 0.18);
  background: rgba(255, 250, 242, 0.78);
  color: rgba(124, 45, 18, 0.96);
  font-weight: 750;
}

.game-search-jump-dialog {
  border: 1px solid rgba(126, 92, 58, 0.14);
  background: rgba(255, 250, 242, 0.94);
  color: rgba(45, 35, 28, 0.92);
}

.game-search-jump-input {
  background: rgba(255, 239, 213, 0.72);
  color: rgba(45, 35, 28, 0.92);
}

.game-search-jump-input:focus {
  box-shadow: 0 0 0 2px rgba(120, 87, 56, 0.12);
}

.game-search-jump-action {
  cursor: pointer;
  border-radius: 0.48rem;
  padding: 0.28rem 0.72rem;
  font-size: 0.85rem;
  transition: background-color 180ms ease, color 180ms ease;
}

.game-search-jump-action--ghost:hover {
  background: rgba(255, 239, 213, 0.72);
}

.game-search-jump-action--primary {
  background: rgba(124, 45, 18, 0.86);
  color: rgba(255, 250, 242, 0.96);
}

.game-search-jump-action--primary:hover {
  background: rgba(99, 39, 15, 0.96);
}

:global(.games-search-page.games-page--dark) .game-search-page-button--idle {
  border-color: rgba(226, 232, 240, 0.14);
  background: rgba(226, 232, 240, 0.055);
  color: rgba(190, 208, 222, 0.70);
}

:global(.games-search-page.games-page--dark) .game-search-page-button--idle:hover {
  border-color: rgba(203, 213, 225, 0.36);
  background: rgba(226, 232, 240, 0.12);
  color: rgba(226, 232, 240, 0.88);
}

:global(.games-search-page.games-page--dark) .game-search-page-button--active {
  border-color: rgba(203, 213, 225, 0.18);
  background: rgba(226, 232, 240, 0.13);
  color: rgba(226, 232, 240, 0.88);
}

:global(.games-search-page.games-page--dark) .game-search-jump-dialog {
  border-color: rgba(226, 232, 240, 0.14);
  background: rgba(30, 41, 59, 0.96);
  color: rgba(226, 232, 240, 0.88);
}

:global(.games-search-page.games-page--dark) .game-search-jump-input {
  background: rgba(15, 23, 42, 0.50);
  color: rgba(226, 232, 240, 0.88);
}

:global(.games-search-page.games-page--dark) .game-search-jump-input:focus {
  box-shadow: 0 0 0 2px rgba(148, 163, 184, 0.20);
}

:global(.games-search-page.games-page--dark) .game-search-jump-action--ghost:hover {
  background: rgba(226, 232, 240, 0.10);
}

:global(.games-search-page.games-page--dark) .game-search-jump-action--primary {
  background: rgba(203, 213, 225, 0.18);
  color: rgba(241, 245, 249, 0.92);
}

:global(.games-search-page.games-page--dark) .game-search-jump-action--primary:hover {
  background: rgba(203, 213, 225, 0.26);
}
</style>
