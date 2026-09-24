<template>
  <div ref="pagination" class="gf-pagination gf-pagination--plain game-search-pagination flex items-center justify-center gap-2 select-none">

    <!-- 页码 -->
    <div class="game-search-pagination-pages">
      <button type="button"
          :aria-current="item.type === 'page' && item.page === currentPage ? 'page' : undefined"
          :aria-label="item.type === 'ellipsis' ? t('game.search.jumpPage') : undefined"
          v-for="(item, idx) in displayPages"
          :key="item.type === 'page' ? `page-${item.page}` : `ellipsis-${idx}`"
          class="gf-pagination__button game-search-page-button"
          :class="item.type === 'page' && item.page === currentPage
            ? 'gf-pagination__button--active game-search-page-button--active'
            : 'game-search-page-button--idle'"
          @click="item.type === 'page' ? changePage(item.page!) : openJump()"
      >
        {{ item.type === 'page' ? item.page : '...' }}
      </button>
    </div>

    <span class="gf-pagination__total game-search-pagination-total ml-2">
      {{ t("common.total") }} {{ total }} {{ t("common.record") }}
    </span>

    <!-- 跳页 -->
    <Teleport v-if="showJump" to="body">
      <div class="games-search-overlay-scope">
    <div
        class="game-search-jump-overlay fixed inset-0 z-50
             flex items-center justify-center"
    >
      <div ref="panel" role="dialog" aria-modal="true" :aria-labelledby="`${id}-title`" tabindex="-1" class="game-search-jump-dialog w-64 space-y-3 p-4">
        <div :id="`${id}-title`" class="game-search-jump-title">{{ t("game.search.jumpPage") }}</div>

        <input
            ref="jumpInput"
            v-model="jumpPage"
            inputmode="numeric"
            :aria-label="t('game.search.jumpPage')"
            :aria-invalid="invalid || undefined"
            @input="clearValidity"
            @keydown.enter.prevent="confirmJump"
            :min="1"
            :max="totalPages"
            class="game-search-jump-input w-full px-2 py-1"
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
    </Teleport>

  </div>
</template>

<script setup lang="ts">
import { computed, ref, useId } from 'vue'
import { i18n } from '@/main'
import { useGameSearchDialog } from '@/composables/useGameSearchDialog'

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
const jumpPage = ref(String(props.currentPage))
const invalid = ref(false)
const id = useId()
const pagination = ref<HTMLElement | null>(null)
const panel = ref<HTMLElement | null>(null)
const jumpInput = ref<HTMLInputElement | null>(null)
useGameSearchDialog(panel, {
  initialFocus: () => jumpInput.value,
  fallbackFocus: () => pagination.value?.querySelector<HTMLElement>('[aria-current="page"]') ?? null,
  dismiss: () => { showJump.value = false },
})
const clearValidity = () => {
  invalid.value = false
  jumpInput.value?.setCustomValidity('')
}
const openJump = () => {
  jumpPage.value = String(props.currentPage)
  invalid.value = false
  showJump.value = true
}
const confirmJump = () => {
  const input = String(jumpPage.value).trim()
  const page = Number(input)
  if (!/^\d+$/.test(input) || !Number.isSafeInteger(page) || page < 1 || page > props.totalPages) {
    invalid.value = true
    jumpInput.value?.setCustomValidity(t('game.search.invalidPage', { max: props.totalPages }))
    jumpInput.value?.reportValidity()
    return
  }
  changePage(page)
  showJump.value = false
}
</script>
