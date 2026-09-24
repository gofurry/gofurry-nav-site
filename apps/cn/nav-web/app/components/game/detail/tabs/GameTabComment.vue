<template>
  <div class="game-detail-comments space-y-4 text-sm">
    <div v-if="unavailable || pageFailed" data-detail-reviews-error role="status" :aria-busy="loading || isLoading">
      <p>{{ t('game.detail.reviewsUnavailable') }}</p>
      <button class="game-detail-load-more px-4 py-1" :disabled="loading || isLoading" @click="unavailable ? emit('retry') : loadMore()">
        {{ loading || isLoading ? t('common.loading') : t('game.detail.retry') }}
      </button>
    </div>

    <div v-if="remarks.length" class="space-y-3">
      <div
        v-for="(remark, index) in remarks"
        :key="`${remark.create_time}-${remark.ip}-${index}`"
        class="game-detail-comment mt-1 space-y-1 p-4"
      >
        <div class="game-detail-comment-meta flex flex-col text-xs">
          <div class="flex justify-between">
            <span><strong>{{ t("game.detail.commenter") }}:</strong> {{ remark.name }}</span>
            <span><strong>{{ t("game.detail.region") }}:</strong> {{ remark.region }}</span>
          </div>
          <div class="flex justify-between">
            <span><strong>IP:</strong> {{ remark.ip }}</span>
            <span><strong>{{ t("game.detail.time") }}:</strong> {{ remark.create_time }}</span>
          </div>
        </div>

        <div class="mt-1 flex items-center gap-1">
          <img v-for="i in fullStars(remark.score)" :key="`full-${i}-${index}`" :src="starSvg" class="h-4 w-4" alt="" />
          <img v-if="hasHalfStar(remark.score)" :src="starHalfSvg" class="h-4 w-4" alt="" />
          <img v-for="i in emptyStars(remark.score)" :key="`empty-${i}-${index}`" :src="starSvg" class="h-4 w-4 opacity-30" alt="" />
          <span class="game-detail-comment-score ml-2">{{ remark.score.toFixed(1) }}</span>
        </div>

        <div class="game-detail-comment-body mt-1">{{ remark.content }}</div>

        <div
          v-if="index !== remarks.length - 1"
          class="game-detail-divider my-3 w-full"
        />
      </div>

      <div class="mt-3 text-center">
        <button
          v-if="hasMore && !pageFailed"
          :disabled="isLoading"
          @click="loadMore"
          class="game-detail-load-more px-4 py-1 disabled:opacity-60"
        >
          {{ isLoading ? t("common.loading") : t("common.loadMore") }}
        </button>
        <span v-else-if="!hasMore" class="game-detail-empty text-sm">{{ t("game.detail.allCommentsLoaded") }}</span>
      </div>
    </div>

    <div v-else-if="!unavailable" class="game-detail-empty py-6 text-center">
      {{ isLoading ? t("common.loading") : t("game.panel.none") }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, onBeforeUnmount } from 'vue'
import type { RemarkModel, RemarkResponse } from '@/types/game'
import { getGameRemark } from '~/services/game'
import { i18n } from '@/main'

import starSvg from '@/assets/svgs/star.svg'
import starHalfSvg from '@/assets/svgs/star-half-alt.svg'

const { t } = i18n.global

const props = defineProps<{
  gameId: string
  remark: RemarkResponse | null
  unavailable?: boolean
  loading?: boolean
}>()
const emit = defineEmits<{ retry: [] }>()

const PAGE_SIZE = 5

const remarks = ref<RemarkModel[]>([])
const total = ref(0)
const currentPage = ref(1)
const isLoading = ref(false)
const pageFailed = ref(false)
let generation = 0
onBeforeUnmount(() => { generation += 1 })

const hasMore = computed(() => remarks.value.length < total.value)

const fullStars = (score: number) => Math.floor(score)
const hasHalfStar = (score: number) => score - Math.floor(score) >= 0.45
const emptyStars = (score: number) => 5 - fullStars(score) - (hasHalfStar(score) ? 1 : 0)

function resetRemarks() {
  generation += 1
  isLoading.value = false
  pageFailed.value = false
  remarks.value = props.remark?.remarks ? [...props.remark.remarks] : []
  total.value = props.remark?.total ?? 0
  currentPage.value = remarks.value.length > 0 ? 1 : 0
}

async function loadMore() {
  if (!props.gameId || isLoading.value || !hasMore.value) {
    return
  }

  isLoading.value = true
  pageFailed.value = false
  const requestGameId = props.gameId
  const requestGeneration = generation
  try {
    const nextPage = currentPage.value + 1
    const response = await getGameRemark(requestGameId, nextPage, PAGE_SIZE)
    if (generation !== requestGeneration || props.gameId !== requestGameId) {
      return
    }

    remarks.value = [...remarks.value, ...(response.remarks ?? [])]
    total.value = response.total ?? total.value
    currentPage.value = response.page_num || nextPage
  } catch {
    if (generation === requestGeneration) pageFailed.value = true
  } finally {
    if (generation === requestGeneration) isLoading.value = false
  }
}

watch(
  () => [props.gameId, props.remark] as const,
  () => {
    resetRemarks()
  },
  { immediate: true }
)
</script>
