<template>
  <div class="search-results space-y-4" :class="{ 'search-results--pending': loading, 'search-results--sliding': trackActive }">

    <!-- 游戏列表 -->
    <div class="search-result-grid-shell">
      <div class="search-result-grid search-result-grid--spacer" aria-hidden="true">
        <div
            v-for="index in displayPageSize"
            :key="`spacer-${index}`"
            class="search-page-card search-page-card--spacer"
        >
          <div class="relative mb-2">
            <div class="search-page-cover search-page-skeleton-cover" />
          </div>
          <div class="search-page-skeleton-title" />
          <div class="search-page-skeleton-line" />
          <div class="search-page-skeleton-meta" />
        </div>
      </div>

      <div class="search-result-page-viewport">
        <div
            class="search-result-page-track"
            :class="{
              'search-result-page-track--instant': instantTrack,
              'search-result-page-track--next': trackActive && slideDirection > 0,
              'search-result-page-track--prev': trackActive && slideDirection < 0,
            }"
            :style="{ transform: `translate3d(${trackOffset}%, 0, 0)` }"
        >
          <div
              v-for="slide in renderedSlides"
              :key="slide.key"
              class="search-result-page-slide"
              :aria-hidden="slide.offscreen"
          >
            <div class="search-result-grid">
              <template v-if="slide.loading">
                <div
                    v-for="index in displayPageSize"
                    :key="`loading-${slide.key}-${index}`"
                    class="search-page-card search-page-card--skeleton"
                    aria-hidden="true"
                >
                  <div class="relative mb-2">
                    <div class="search-page-cover search-page-skeleton-cover" />
                  </div>
                  <div class="search-page-skeleton-title" />
                  <div class="search-page-skeleton-line" />
                  <div class="search-page-skeleton-meta" />
                </div>
              </template>

              <template v-else>
                <div
                    v-for="game in slide.games"
                    :key="game.id"
                    class="search-page-card group"
                    @click="goDetail(game.id)"
                >
                  <div class="relative mb-2">
                    <img
                        :src="game.cover"
                        class="search-page-cover"
                        :alt="game.name"
                    />

                    <button
                        class="search-review-button"
                        type="button"
                        :aria-label="`${game.name} review`"
                        @click.stop="openReview(game)"
                    >
                      <svg
                          aria-hidden="true"
                          viewBox="0 0 24 24"
                          class="h-[1.125rem] w-[1.125rem]"
                      >
                        <path
                            fill="currentColor"
                            d="M9 22c-.55 0-1-.45-1-1v-3H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v12c0 1.11-.89 2-2 2h-6.1l-3.7 3.71c-.2.19-.45.29-.7.29H9m1-6v3.08L13.08 16H20V4H4v12h6m5.84-7.8l-1.01 1.01l-2.07-2.03l1.01-1.02c.2-.21.54-.22.78 0l1.29 1.25c.21.21.22.55 0 .79M8 11.91l4.17-4.19l2.07 2.08l-4.16 4.2H8v-2.09Z"
                        />
                      </svg>
                    </button>
                  </div>

                  <div class="search-page-title-row">
                    <div class="search-page-title-wrap">
                      <div class="search-page-title truncate">
                        {{ game.name }}
                      </div>
                      <div v-if="game.primary_tag" class="search-page-tag search-page-tag--primary">
                        {{ game.primary_tag }}
                      </div>
                      <div v-if="game.secondary_tag" class="search-page-tag search-page-tag--secondary">
                        {{ game.secondary_tag }}
                      </div>
                    </div>

                    <a
                        :href="steamPrefix+`${game.appid}`"
                        target="_blank"
                        rel="noopener noreferrer"
                        class="shrink-0"
                        @click.stop
                    >
                      <img
                          src="@/assets/icons/steam.svg"
                          alt="Steam"
                          class="w-4 h-4 opacity-70 hover:opacity-100 transition"
                      />
                    </a>
                  </div>

                  <p class="search-page-desc">
                    {{ game.info }}
                  </p>

                  <div class="search-page-meta">
                    <span class="search-page-score">
                      <img
                          src="@/assets/svgs/star.svg"
                          alt=""
                          class="w-3.5 h-3.5"
                      />
                      <span class="search-page-score-value">
                        {{
                          game.avg_score > 0
                              ? game.avg_score.toFixed(1)
                              : t("game.panel.none")
                        }}
                      </span>
                    </span>

                    <span class="search-page-comment">{{ game.remark_count }} {{t("game.search.comment")}}</span>
                  </div>
                </div>

                <div
                    v-for="index in placeholderCountFor(slide.games)"
                    :key="`placeholder-${slide.key}-${index}`"
                    class="search-page-card search-page-card--placeholder"
                    aria-hidden="true"
                >
                  <div class="relative mb-2">
                    <div class="search-page-cover search-page-skeleton-cover" />
                  </div>
                  <div class="search-page-skeleton-title" />
                  <div class="search-page-skeleton-line" />
                  <div class="search-page-skeleton-meta" />
                </div>
              </template>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 分页 -->
    <GamePagination
        :current-page="currentPage"
        :total-pages="totalPages"
        :total="total"
        @page-change="$emit('page-change', $event)"
    />

    <GameReviewDialog
        :visible="!!reviewGame"
        :game-id="reviewGame?.id ?? ''"
        :game-name="reviewGame?.name ?? ''"
        @close="reviewGame = null"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import GamePagination from '@/components/game/search/GamePagination.vue'
import GameReviewDialog from '@/components/game/common/GameReviewDialog.vue'
import type { SearchPageResponseItem } from '@/types/game'
import { i18n } from '@/main'

const { t } = i18n.global

const router = useRouter()
const localePath = useLocalePath()

const steamPrefix = import.meta.env.VITE_STEAM_APP_PREFIX_URL || ''
const reviewGame = ref<SearchPageResponseItem | null>(null)

const goDetail = (id: number | string) => {
  router.push(localePath(`/games/${id}`))
}

const openReview = (game: SearchPageResponseItem) => {
  reviewGame.value = game
}

const props = defineProps<{
  gameList: SearchPageResponseItem[]
  currentPage: number
  pageSize: number
  pageDirection: 1 | -1
  totalPages: number
  total: number
  loading?: boolean
}>()

defineEmits<{
  (e: 'page-change', page: number): void
}>()

interface SearchSlide {
  key: string
  games: SearchPageResponseItem[]
  loading?: boolean
  offscreen?: boolean
}

const SLIDE_DURATION_MS = 380

const currentGames = ref<SearchPageResponseItem[]>([...props.gameList])
const sourceGames = ref<SearchPageResponseItem[]>([])
const targetGames = ref<SearchPageResponseItem[]>([])
const trackActive = ref(false)
const targetLoading = ref(false)
const trackOffset = ref(0)
const instantTrack = ref(false)
const slideDirection = ref<1 | -1>(1)
const displayPageSize = computed(() => Math.max(1, props.pageSize || 20))
let slideTimer: ReturnType<typeof setTimeout> | null = null
let frameId = 0
let initialized = false

const renderedSlides = computed<SearchSlide[]>(() => {
  if (!trackActive.value) {
    return [{ key: `page-${props.currentPage}`, games: currentGames.value }]
  }

  const sourceSlide: SearchSlide = {
    key: `source-${props.currentPage}`,
    games: sourceGames.value,
    offscreen: true,
  }
  const targetSlide: SearchSlide = {
    key: targetLoading.value ? `loading-${props.currentPage}` : `target-${props.currentPage}`,
    games: targetGames.value,
    loading: targetLoading.value,
    offscreen: false,
  }

  if (props.pageDirection > 0) {
    return [
      sourceSlide,
      targetSlide,
    ]
  }

  return [
    targetSlide,
    sourceSlide,
  ]
})

const placeholderCountFor = (games: SearchPageResponseItem[]) => Math.max(0, displayPageSize.value - games.length)

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

function scheduleCleanup(delay = SLIDE_DURATION_MS + 80) {
  if (slideTimer) {
    clearTimeout(slideTimer)
  }

  slideTimer = setTimeout(() => {
    instantTrack.value = true
    currentGames.value = [...targetGames.value]
    sourceGames.value = []
    targetGames.value = []
    targetLoading.value = false
    trackActive.value = false
    trackOffset.value = 0
    slideTimer = null

    void nextTick().then(() => {
      frameId = requestAnimationFrame(() => {
        instantTrack.value = false
        frameId = 0
      })
    })
  }, delay)
}

function startPageSlide(nextGames: SearchPageResponseItem[], loading: boolean) {
  clearSlideTimer()
  instantTrack.value = false

  sourceGames.value = [...currentGames.value]
  targetGames.value = [...nextGames]
  targetLoading.value = loading
  slideDirection.value = props.pageDirection
  trackActive.value = true
  trackOffset.value = slideDirection.value > 0 ? -100 : 0

  if (!loading) {
    scheduleCleanup()
  }
}

function sameGameList(left: SearchPageResponseItem[], right: SearchPageResponseItem[]) {
  if (left.length !== right.length) {
    return false
  }

  return left.every((game, index) => game.id === right[index]?.id)
}

watch(
  () => props.gameList,
  (nextList) => {
    const nextGames = [...nextList]

    if (!initialized) {
      initialized = true
      currentGames.value = nextGames
      return
    }

    if (trackActive.value && targetLoading.value) {
      targetGames.value = nextGames
      targetLoading.value = false
      scheduleCleanup()
      return
    }

    if (!currentGames.value.length || sameGameList(currentGames.value, nextGames)) {
      currentGames.value = nextGames
      return
    }

    void startPageSlide(nextGames, false)
  },
  { immediate: true }
)

watch(
  () => props.currentPage,
  (page, previousPage) => {
    if (!initialized || page === previousPage || !currentGames.value.length || trackActive.value) {
      return
    }

    void startPageSlide([], true)
  }
)

watch(
  () => props.loading,
  (loading) => {
    if (loading) {
      if (!currentGames.value.length || trackActive.value) {
        return
      }

      void startPageSlide([], true)
      return
    }

    if (trackActive.value && targetLoading.value) {
      targetGames.value = [...currentGames.value]
      targetLoading.value = false
      scheduleCleanup()
    }
  }
)

onBeforeUnmount(() => {
  clearSlideTimer()
})
</script>

<style scoped>
.search-page-card {
  container: search-card / inline-size;
  cursor: pointer;
  min-width: 0;
}

.search-result-grid-shell {
  position: relative;
  container: game-search-results / inline-size;
  overflow: hidden;
}

.search-result-grid--spacer {
  pointer-events: none;
  visibility: hidden;
}

.search-result-page-viewport {
  position: absolute;
  inset: 0;
  overflow: hidden;
}

.search-result-page-track {
  display: flex;
  height: 100%;
  width: 100%;
  will-change: transform;
}

.search-result-page-track--instant {
  animation: none;
  transition: none;
}

.search-result-page-track--next {
  animation: search-result-page-next 420ms cubic-bezier(0.22, 1, 0.36, 1) both;
}

.search-result-page-track--prev {
  animation: search-result-page-prev 420ms cubic-bezier(0.22, 1, 0.36, 1) both;
}

.search-result-page-slide {
  width: 100%;
  min-width: 100%;
  flex: 0 0 100%;
  overflow: hidden;
}

.search-result-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1rem;
  min-width: 0;
  transition: opacity 180ms ease, filter 180ms ease;
}

.search-result-page-slide .search-result-grid {
  height: 100%;
  align-content: start;
}

.search-results--pending:not(.search-results--sliding) .search-result-page-slide .search-result-grid {
  opacity: 0.72;
  filter: saturate(0.92);
}

@container game-search-results (min-width: 42rem) {
  .search-result-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@container game-search-results (min-width: 56rem) {
  .search-result-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

@container game-search-results (min-width: 94rem) {
  .search-result-grid {
    grid-template-columns: repeat(5, minmax(0, 1fr));
  }
}

.search-page-title {
  min-width: 0;
  max-width: 100%;
  flex: 1 1 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.search-page-title-row {
  display: flex;
  min-width: 0;
  overflow: hidden;
  align-items: center;
  justify-content: space-between;
  gap: 0.35rem;
  font-size: 0.875rem;
}

.search-page-title-wrap {
  display: flex;
  width: 100%;
  min-width: 0;
  overflow: hidden;
  flex: 1 1 auto;
  align-items: center;
  gap: 0.38rem;
}

.search-page-tag {
  display: inline-flex;
  max-width: min(9.25rem, 48%);
  min-width: 0;
  min-height: 1.35rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@container search-card (max-width: 27rem) {
  .search-page-tag--secondary {
    display: none;
  }
}

@container search-card (max-width: 14.25rem) {
  .search-page-tag--primary,
  .search-page-tag--secondary {
    display: none;
  }
}

.search-page-desc {
  margin-top: 0.25rem;
  display: -webkit-box;
  height: 2.2rem;
  overflow: hidden;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.search-page-meta {
  margin-top: 0.62rem;
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
}

.search-page-score {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.search-page-score-value {
  min-width: 1.45rem;
}

.search-page-card--placeholder {
  pointer-events: none;
  visibility: hidden;
}

.search-page-card--skeleton {
  pointer-events: none;
}

.search-page-skeleton-cover,
.search-page-skeleton-title,
.search-page-skeleton-line,
.search-page-skeleton-meta {
  overflow: hidden;
  border-radius: 0.52rem;
  background: linear-gradient(
    110deg,
    color-mix(in srgb, var(--games-search-surface-strong) 70%, transparent) 0%,
    color-mix(in srgb, var(--games-search-surface-hover) 82%, transparent) 46%,
    color-mix(in srgb, var(--games-search-surface-strong) 70%, transparent) 100%
  );
  background-size: 220% 100%;
  animation: search-card-skeleton 1.35s ease-in-out infinite;
}

.search-page-skeleton-title {
  width: 70%;
  height: 1rem;
}

.search-page-skeleton-line {
  width: 100%;
  height: 2.2rem;
  margin-top: 0.52rem;
}

.search-page-skeleton-meta {
  width: 76%;
  height: 0.86rem;
  margin-top: 0.72rem;
}

@keyframes search-card-skeleton {
  from {
    background-position: 120% 0;
  }
  to {
    background-position: -120% 0;
  }
}

@keyframes search-result-page-next {
  from {
    transform: translate3d(0, 0, 0);
  }
  to {
    transform: translate3d(-100%, 0, 0);
  }
}

@keyframes search-result-page-prev {
  from {
    transform: translate3d(-100%, 0, 0);
  }
  to {
    transform: translate3d(0, 0, 0);
  }
}

</style>
