<template>
  <div class="game-detail-sidebar-card p-4">
    <h3 class="game-detail-sidebar-title mb-3 font-semibold">{{ t("game.detail.similarGames") }}</h3>

    <!-- 游戏列表 -->
    <div class="game-detail-similar-list-shell" :class="{ 'game-detail-similar-list-shell--paged': totalPages > 1 }">
      <div
          v-if="recommendList.length"
          class="game-detail-similar-list game-detail-similar-list--spacer space-y-3"
          aria-hidden="true"
      >
        <div
            v-for="index in PAGE_SIZE"
            :key="`spacer-${index}`"
            class="game-detail-similar-item flex gap-3 p-2"
        >
          <div class="h-16 w-12 flex-shrink-0 rounded-md" />
          <div class="game-detail-similar-body min-w-0 flex-1 text-sm">
            <div class="game-detail-similar-title truncate font-medium">placeholder</div>
            <div class="game-detail-similar-summary line-clamp-2 text-xs">placeholder</div>
            <div class="game-detail-similar-score mt-1 text-xs">placeholder</div>
            <div class="game-detail-similar-reason truncate text-[11px]">placeholder</div>
          </div>
        </div>
      </div>

      <div v-if="recommendList.length" class="game-detail-similar-viewport">
        <div
            :key="similarTrackKey"
            class="game-detail-similar-track"
            :class="{
              'game-detail-similar-track--instant': instantTrack,
              'game-detail-similar-track--next': isSliding && pageDirection > 0,
              'game-detail-similar-track--prev': isSliding && pageDirection < 0,
            }"
        >
          <div
              v-for="slide in displayedSlides"
              :key="slide.key"
              class="game-detail-similar-slide"
              :aria-hidden="slide.page !== currentPage"
          >
            <div class="game-detail-similar-list space-y-3">
          <div
              v-for="game in slide.games"
              :key="game.id"
              class="game-detail-similar-item flex cursor-pointer gap-3 p-2"
              @click="goGameDetail(game.id)"
          >
            <!-- 封面 -->
            <SteamAssetImage
                v-if="coverOf(game)"
                :src="coverOf(game)"
                class="w-12 h-16 rounded-md object-cover flex-shrink-0"
                :alt="game.name"
                @error="hideBrokenCover"
            />
            <div v-else class="game-detail-similar-cover-empty w-12 h-16 rounded-md flex-shrink-0" />

            <!-- 游戏信息 -->
            <div class="game-detail-similar-body min-w-0 flex-1 text-sm">
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

          <div
              v-for="index in placeholderCountFor(slide.games)"
              :key="`placeholder-${slide.page}-${index}`"
              class="game-detail-similar-item game-detail-similar-item--placeholder flex gap-3 p-2"
              aria-hidden="true"
          >
            <div class="w-12 h-16 rounded-md flex-shrink-0" />
            <div class="game-detail-similar-body min-w-0 flex-1 text-sm">
              <div class="game-detail-similar-title truncate font-medium">placeholder</div>
              <div class="game-detail-similar-summary line-clamp-2 text-xs">placeholder</div>
              <div class="game-detail-similar-score mt-1 text-xs">placeholder</div>
              <div class="game-detail-similar-reason truncate text-[11px]">placeholder</div>
            </div>
          </div>
            </div>
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
          @click="changePage(page)"
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
import { computed, nextTick, onBeforeUnmount, ref, watch } from "vue"
import type { RecommendedModel } from "@/types/game"
import SteamAssetImage from '@/components/common/SteamAssetImage.vue'
import { i18n } from '@/main'

const { t } = i18n.global
const router = useRouter()
const localePath = useLocalePath()

const props = defineProps<{
  recommend: RecommendedModel[] | null
}>()

const recommendList = computed(() => props.recommend ?? [])

const PAGE_SIZE = 4
const currentPage = ref(1)
const renderedPage = ref(1)
const sourcePage = ref(1)
const targetPage = ref(1)
const pageDirection = ref<1 | -1>(1)
const isSliding = ref(false)
const instantTrack = ref(false)
const slideNonce = ref(0)
let slideTimer: ReturnType<typeof setTimeout> | null = null
let frameId = 0

// 总页数
const totalPages = computed(() => {
  const len = recommendList.value.length
  return len === 0 ? 0 : Math.ceil(len / PAGE_SIZE)
})

const gamePages = computed(() => {
  const list = recommendList.value
  const pages: RecommendedModel[][] = []

  for (let index = 0; index < list.length; index += PAGE_SIZE) {
    pages.push(list.slice(index, index + PAGE_SIZE))
  }

  return pages
})

const gamesForPage = (page: number) => {
  if (page <= 0) {
    return []
  }

  return gamePages.value[page - 1] ?? []
}

const displayedSlides = computed(() => {
  if (!isSliding.value) {
    return [{
      key: `similar-page-${renderedPage.value}`,
      page: renderedPage.value,
      games: gamesForPage(renderedPage.value),
    }]
  }

  const source = {
    key: `similar-source-${sourcePage.value}`,
    page: sourcePage.value,
    games: gamesForPage(sourcePage.value),
  }
  const target = {
    key: `similar-target-${targetPage.value}`,
    page: targetPage.value,
    games: gamesForPage(targetPage.value),
  }

  return pageDirection.value > 0 ? [source, target] : [target, source]
})

const similarTrackKey = computed(() => {
  if (isSliding.value) {
    return `similar-motion-${sourcePage.value}-${targetPage.value}-${slideNonce.value}`
  }

  return `similar-stable-${renderedPage.value}-${slideNonce.value}`
})

const placeholderCountFor = (games: RecommendedModel[]) => Math.max(0, PAGE_SIZE - games.length)

function clearSlideTimer() {
  if (slideTimer) {
    clearTimeout(slideTimer)
    slideTimer = null
  }

  if (frameId) {
    cancelAnimationFrame(frameId)
    frameId = 0
  }
}

function finishSlide() {
  instantTrack.value = true
  renderedPage.value = targetPage.value
  isSliding.value = false
  slideTimer = null

  void nextTick().then(() => {
    frameId = requestAnimationFrame(() => {
      instantTrack.value = false
      frameId = 0
    })
  })
}

function changePage(page: number) {
  if (page === currentPage.value || page < 1 || page > totalPages.value) {
    return
  }

  clearSlideTimer()
  instantTrack.value = false
  pageDirection.value = page > renderedPage.value ? 1 : -1
  sourcePage.value = renderedPage.value
  targetPage.value = page
  currentPage.value = page
  isSliding.value = true
  slideNonce.value += 1

  slideTimer = setTimeout(finishSlide, 560)
}

// 格式化相似度
function formatSimilarity(sim: number) {
  return `${(sim * 100).toFixed(1)}%`
}

function coverOf(game: RecommendedModel) {
  return game.library_cover_url || game.library_cover_2x_url || ''
}

function hideBrokenCover(event: Event) {
  const image = event.currentTarget as HTMLImageElement | null
  if (!image) return

  image.style.visibility = 'hidden'
}

function formatReason(game: RecommendedModel) {
  const reason = game.reasons?.[0]
  if (!reason) return ''
  return `${reason.label}: ${reason.value}`
}

// 跳转
function goGameDetail(gameId: string) {
  router.push(localePath(`/games/${gameId}`))
}

watch(
    totalPages,
    (pages) => {
      if (pages === 0) {
        currentPage.value = 1
        renderedPage.value = 1
        sourcePage.value = 1
        targetPage.value = 1
      } else if (currentPage.value > pages) {
        currentPage.value = pages
        renderedPage.value = pages
        sourcePage.value = pages
        targetPage.value = pages
      }
    },
    { immediate: true }
)

onBeforeUnmount(() => {
  clearSlideTimer()
})
</script>

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.game-detail-similar-list-shell {
  position: relative;
  min-width: 0;
  overflow: hidden;
  contain: layout paint;
}

.game-detail-similar-list,
.game-detail-similar-item {
  width: 100%;
  min-width: 0;
}

.game-detail-similar-list--spacer {
  pointer-events: none;
  visibility: hidden;
}

.game-detail-similar-viewport {
  position: absolute;
  inset: 0;
  overflow: hidden;
}

.game-detail-similar-track {
  display: flex;
  height: 100%;
  width: 100%;
  min-width: 0;
  will-change: transform;
}

.game-detail-similar-track--instant {
  animation: none;
  transition: none;
}

.game-detail-similar-track--next {
  animation: game-detail-similar-next 520ms cubic-bezier(0.22, 1, 0.36, 1) both;
}

.game-detail-similar-track--prev {
  animation: game-detail-similar-prev 520ms cubic-bezier(0.22, 1, 0.36, 1) both;
}

.game-detail-similar-slide {
  width: 100%;
  min-width: 100%;
  flex: 0 0 100%;
  overflow: hidden;
}

.game-detail-similar-item {
  height: 6.625rem;
  max-height: 6.625rem;
  min-height: 6.625rem;
  overflow: hidden;
}

.game-detail-similar-body {
  overflow: hidden;
}

.game-detail-similar-cover-empty {
  background: rgba(148, 163, 184, 0.16);
}

.game-detail-similar-title,
.game-detail-similar-reason {
  max-width: 100%;
  white-space: nowrap;
}

.game-detail-similar-summary {
  max-height: 2rem;
  line-height: 1rem;
}

.game-detail-similar-item--placeholder {
  visibility: hidden;
  pointer-events: none;
}

@keyframes game-detail-similar-next {
  from {
    transform: translate3d(0, 0, 0);
  }
  to {
    transform: translate3d(-100%, 0, 0);
  }
}

@keyframes game-detail-similar-prev {
  from {
    transform: translate3d(-100%, 0, 0);
  }
  to {
    transform: translate3d(0, 0, 0);
  }
}

</style>
