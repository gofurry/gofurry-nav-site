<template>
  <div class="game-detail-sidebar-card p-4">
    <h3 class="game-detail-sidebar-title mb-3 font-semibold">{{ t("game.detail.similarGames") }}</h3>

    <!-- 游戏列表 -->
    <div class="space-y-3">
      <div
          v-for="game in pagedGames"
          :key="game.id"
          class="game-detail-similar-item flex cursor-pointer gap-3 p-2"
          @click="goGameDetail(game.id)"
      >
        <!-- 封面 -->
        <img
            :src="coverOf(game)"
            class="w-12 h-16 rounded-md object-cover flex-shrink-0"
            :alt="game.name"
            @error="loadFallbackCover($event, game.appid)"
        />

        <!-- 游戏信息 -->
        <div class="min-w-0 flex-1 text-sm">
          <div class="game-detail-similar-title truncate font-medium">{{ game.name }}</div>
          <div class="game-detail-similar-summary line-clamp-2 break-words text-xs">{{ game.summary }}</div>
          <div class="game-detail-similar-score mt-1 text-xs">
            {{ t("game.detail.similarity") }}: {{ formatSimilarity(game.display_score) }}
          </div>
          <div v-if="formatReason(game)" class="game-detail-similar-reason truncate text-[11px]">
            {{ formatReason(game) }}
          </div>
        </div>
      </div>

      <!-- 没有数据 -->
      <div v-if="!recommendList.length" class="game-detail-empty py-4 text-center text-sm">
        {{ t("game.panel.none") }}
      </div>
    </div>

    <!-- 分页按钮 -->
    <div
        v-if="recommendList.length > PAGE_SIZE"
        class="flex justify-center gap-2 mt-3"
    >
      <button
          v-for="page in totalPages"
          :key="page"
          @click="currentPage = page"
          :class="[
          'game-detail-page-button px-3 py-1 text-sm font-medium',
          currentPage === page
            ? 'game-detail-page-button--active'
            : 'game-detail-page-button--idle'
        ]"
      >
        {{ page }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue"
import type { RecommendedModel } from "@/types/game"
import { i18n } from '@/main'
import { cdnLibraryCoverUrl, steamLibraryCoverUrl } from '@/utils/gameAssets'

const { t } = i18n.global
const router = useRouter()

const props = defineProps<{
  recommend: RecommendedModel[] | null
}>()

const recommendList = computed(() => props.recommend ?? [])

const PAGE_SIZE = 4
const currentPage = ref(1)

// 总页数
const totalPages = computed(() => {
  const len = recommendList.value.length
  return len === 0 ? 0 : Math.ceil(len / PAGE_SIZE)
})

// 当前页展示的游戏
const pagedGames = computed(() => {
  const list = recommendList.value
  if (list.length === 0) return []
  const start = (currentPage.value - 1) * PAGE_SIZE
  return list.slice(start, start + PAGE_SIZE)
})

// 格式化相似度
function formatSimilarity(sim: number) {
  return `${(sim * 100).toFixed(1)}%`
}

function coverOf(game: RecommendedModel) {
  return steamLibraryCoverUrl(game.appid)
}

function loadFallbackCover(event: Event, appid: string) {
  const image = event.currentTarget as HTMLImageElement | null
  if (!image) return

  if (image.dataset.coverFallback === 'cdn') {
    image.style.visibility = 'hidden'
    return
  }

  const fallback = cdnLibraryCoverUrl(appid)
  if (!fallback) {
    image.style.visibility = 'hidden'
    return
  }

  image.dataset.coverFallback = 'cdn'
  image.src = fallback
}

function formatReason(game: RecommendedModel) {
  const reason = game.reasons?.[0]
  if (!reason) return ''
  return `${reason.label}: ${reason.value}`
}

// 跳转
function goGameDetail(gameId: string) {
  router.push(`/games/${gameId}`)
}

watch(
    totalPages,
    (pages) => {
      if (pages === 0) {
        currentPage.value = 1
      } else if (currentPage.value > pages) {
        currentPage.value = pages
      }
    },
    { immediate: true }
)
</script>

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
