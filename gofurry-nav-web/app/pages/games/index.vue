<template>
  <div
    class="games-page flex min-h-full w-full flex-col overflow-clip bg-gray-50 transition-colors duration-500 dark:bg-[#07111f]"
    :class="{ 'games-page--dark': isDarkTheme }"
  >
    <main
      class="relative isolate flex-1 overflow-hidden"
    >
      <GoFurryGridBackground :fixed="false" palette="games" />
      <FallingLeavesCanvas class="z-[1]" mode="viewport" :leaf-count="42" />
      <h1 class="sr-only">{{ gamesPageSeo.heading }}</h1>
      <div class="relative z-10 mx-auto flex w-full max-w-[1700px] gap-4 p-6">
        <section class="w-full xl:w-[75%]">
          <GameInfoPanel
            :initial-raw-data="gamesPageData.mainInfo"
            :initial-panel-data="gamesPageData.panelData"
            :initial-news-record="gamesPageData.latestNews"
          />
        </section>

        <aside class="hidden xl:block xl:w-[25%]">
          <SideBarPanel :initial-reviews="gamesPageData.latestReviews" />
        </aside>
      </div>
    </main>

    <GameToolDock />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import FallingLeavesCanvas from '@/components/common/FallingLeavesCanvas.vue'
import GoFurryGridBackground from '@/components/common/GoFurryGridBackground.vue'
import GameInfoPanel from '@/components/game/main/content/GameInfoPanel.vue'
import GameToolDock from '@/components/game/main/GameToolDock.vue'
import SideBarPanel from '@/components/game/main/sidebar/SideBarPanel.vue'
import { useThemeStore } from '@/stores/theme'
import { getGameHomeData, getLatestReview } from '~/services/game'
import type { AnonymousReviewModel, GameGroupRecord, GamePanelRecord, LatestNewsRecord } from '~/types/game'

interface GamesPageData {
  mainInfo: GameGroupRecord | null
  panelData: GamePanelRecord | null
  latestNews: LatestNewsRecord | null
  latestReviews: AnonymousReviewModel[]
}

const { locale } = useI18n()
const themeStore = useThemeStore()
const lang = computed(() => (locale.value === 'en' ? 'en' : 'zh'))
const isDarkTheme = computed(() => themeStore.theme === 'dark')
const gamesPageSeo = computed(() => locale.value === 'en'
  ? {
      heading: 'GoFurry Furry Games',
      title: 'GoFurry Furry Games - New releases, rankings, discounts, and reviews',
      description: 'Explore furry and anthro games on GoFurry, including recent releases, newly listed titles, free games, popularity rankings, price signals, update news, and community review activity.'
    }
  : {
      heading: 'GoFurry 兽人游戏资料库',
      title: 'GoFurry 兽人游戏资料库 - 新作、排行、折扣与评价',
      description: '在 GoFurry 浏览兽人、拟人与相关题材游戏资料，查看最近发售、最新收录、免费专区、热门排行、价格信号、更新资讯与社区评价动态。'
    }
)

const { data } = await useAsyncData<GamesPageData>(
  () => `games-page:${lang.value}`,
  async () => {
    const [homeData, latestReviews] = await Promise.all([
      getGameHomeData(lang.value).catch(() => null),
      getLatestReview().catch(() => []),
    ])

    return {
      mainInfo: homeData?.mainInfo ?? null,
      panelData: homeData?.panelData ?? null,
      latestNews: homeData?.latestNews ?? null,
      latestReviews,
    }
  },
  {
    watch: [lang],
    default: () => ({
      mainInfo: null,
      panelData: null,
      latestNews: null,
      latestReviews: [],
    }),
  }
)

const gamesPageData = computed(() => data.value!)

useSeoMeta({
  title: () => gamesPageSeo.value.title,
  description: () => gamesPageSeo.value.description,
  ogTitle: () => gamesPageSeo.value.title,
  ogDescription: () => gamesPageSeo.value.description,
})

onMounted(() => {
  themeStore.initTheme()
})
</script>

<style scoped>
:global(.games-page--dark) {
  background: #07111f;
  color: rgba(226, 232, 240, 0.88);
}

:global(.games-page.games-page--dark .game-info-shell),
:global(.games-page.games-page--dark .game-sidebar-shell) {
  border-color: rgba(226, 232, 240, 0.18);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.10), rgba(226, 232, 240, 0.045)),
    rgba(226, 232, 240, 0.055);
}

:global(.games-page.games-page--dark .game-news-panel),
:global(.games-page.games-page--dark .game-stats-card) {
  border-color: rgba(226, 232, 240, 0.12);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.052), rgba(226, 232, 240, 0.024)),
    rgba(226, 232, 240, 0.030);
}

:global(.games-page--dark h2),
:global(.games-page--dark h3),
:global(.games-page--dark .game-card p:first-of-type),
:global(.games-page--dark .game-stats-card .text-gray-950),
:global(.games-page--dark .game-stats-row .text-gray-950),
:global(.games-page--dark .game-site-item .text-gray-800),
:global(.games-page--dark .latest-review-item .text-stone-800),
:global(.games-page--dark .news-card__title) {
  color: rgba(241, 245, 249, 0.88);
}

:global(.games-page--dark .game-card p:nth-of-type(2)),
:global(.games-page--dark .game-stats-card p),
:global(.games-page--dark .game-stats-card .text-gray-500),
:global(.games-page--dark .game-stats-row .text-gray-500),
:global(.games-page--dark .game-stats-row .text-gray-700),
:global(.games-page--dark .game-site-item .text-gray-500),
:global(.games-page--dark .latest-review-item .text-stone-700),
:global(.games-page--dark .latest-review-item .text-stone-400),
:global(.games-page--dark .sidebar-section-title),
:global(.games-page--dark .news-card__summary),
:global(.games-page--dark .news-card__meta) {
  color: rgba(203, 213, 225, 0.66);
}

:global(.games-page--dark .game-card:hover),
:global(.games-page--dark .game-card:focus-within) {
  background: rgba(148, 163, 184, 0.08);
}

:global(.games-page--dark .game-stats-row::before) {
  background:
    linear-gradient(
      90deg,
      rgba(203, 213, 225, 0.30),
      rgba(148, 163, 184, 0.16) 54%,
      rgba(100, 116, 139, 0.035)
    ) !important;
}

:global(.games-page--dark .stats-row--warm) {
  background: rgba(30, 41, 59, 0.24);
}

:global(.games-page--dark .stats-row--clear) {
  background: rgba(15, 23, 42, 0.18);
}

:global(.games-page.games-page--dark .game-stats-row:hover) {
  background: rgba(51, 65, 85, 0.28);
}

:global(.games-page:not(.games-page--dark) .game-stats-row:hover) {
  background: rgba(255, 229, 194, 0.64) !important;
}

:global(.games-page--dark .game-stats-card header span) {
  border-color: rgba(148, 163, 184, 0.20);
  background: rgba(148, 163, 184, 0.08);
  color: rgba(203, 213, 225, 0.76);
}

:global(.games-page.games-page--dark .game-site-item),
:global(.games-page.games-page--dark .latest-review-item),
:global(.games-page.games-page--dark .sidebar-action-button),
:global(.games-page.games-page--dark .search-results-panel),
:global(.games-page.games-page--dark .search-result-card),
:global(.games-page.games-page--dark .news-card) {
  border-color: rgba(226, 232, 240, 0.15);
  background: rgba(226, 232, 240, 0.065);
}

:global(.games-page.games-page--dark .game-site-item:hover),
:global(.games-page.games-page--dark .latest-review-item:hover),
:global(.games-page.games-page--dark .sidebar-action-button:hover),
:global(.games-page.games-page--dark .search-result-card:hover),
:global(.games-page.games-page--dark .news-card:hover) {
  border-color: rgba(203, 213, 225, 0.38) !important;
  background: rgba(226, 232, 240, 0.13) !important;
}

:global(.games-page--dark .sidebar-action-button),
:global(.games-page--dark .game-group-more),
:global(.games-page--dark .game-group-pager button),
:global(.games-page--dark .news-nav-button),
:global(.games-page--dark .sidebar-expand-button) {
  color: rgba(190, 208, 222, 0.72) !important;
}

:global(.games-page--dark .game-group-pager span),
:global(.games-page--dark .text-orange-900\/55),
:global(.games-page--dark .text-stone-500\/80) {
  color: rgba(203, 213, 225, 0.62);
}

:global(.games-page.games-page--dark .rating-star__score) {
  color: rgba(226, 232, 240, 0.92);
}

:global(.games-page.games-page--dark .rating-star__count) {
  color: rgba(203, 213, 225, 0.82);
}

:global(.games-page.games-page--dark .rating-star__empty) {
  color: rgba(148, 163, 184, 0.30);
  -webkit-text-stroke: 0;
}

:global(.games-page.games-page--dark .rating-star__fill) {
  color: #f59e0b;
  text-shadow: none;
}

:global(.games-page.games-page--dark .news-progress-track) {
  background: rgba(148, 163, 184, 0.16) !important;
}

:global(.games-page.games-page--dark .news-progress-fill) {
  background: rgba(203, 213, 225, 0.55) !important;
  background-color: rgba(203, 213, 225, 0.55) !important;
}

:global(.games-page.games-page--dark .game-group-more) {
  border-color: rgba(148, 163, 184, 0.30) !important;
  border-bottom-color: rgba(148, 163, 184, 0.30) !important;
  color: rgba(190, 208, 222, 0.72) !important;
}

:global(.games-page.games-page--dark .game-group-more:hover) {
  border-color: rgba(203, 213, 225, 0.56) !important;
  border-bottom-color: rgba(203, 213, 225, 0.56) !important;
  color: rgba(226, 232, 240, 0.90) !important;
}

:global(.games-page.games-page--dark .game-group-pager button:hover),
:global(.games-page.games-page--dark .news-nav-button:hover) {
  background: rgba(226, 232, 240, 0.11) !important;
  color: rgba(226, 232, 240, 0.88) !important;
}

:global(.games-page.games-page--dark .sidebar-expand-button:hover) {
  background: transparent !important;
  color: rgba(226, 232, 240, 0.88) !important;
}

:global(.games-page.games-page--dark .game-sidebar-search-input:focus) {
  border-color: rgba(203, 213, 225, 0.38) !important;
  box-shadow: 0 0 0 1px rgba(148, 163, 184, 0.18) !important;
}

:global(.games-page:not(.games-page--dark) .game-group-pager button:hover),
:global(.games-page:not(.games-page--dark) .news-nav-button:hover) {
  background: rgba(154, 52, 18, 0.12) !important;
  color: rgba(124, 45, 18, 1) !important;
}

:global(.games-page:not(.games-page--dark) .game-group-more:hover) {
  border-color: rgba(154, 52, 18, 0.70) !important;
  color: rgba(124, 45, 18, 1) !important;
}

:global(.games-page:not(.games-page--dark) .sidebar-action-button:hover),
:global(.games-page:not(.games-page--dark) .game-site-item:hover),
:global(.games-page:not(.games-page--dark) .latest-review-item:hover),
:global(.games-page:not(.games-page--dark) .search-result-card:hover) {
  border-color: rgba(180, 96, 24, 0.34) !important;
  background: rgba(255, 239, 213, 0.68) !important;
}

:global(.games-page.games-page--dark .sidebar-action-button:hover),
:global(.games-page.games-page--dark .game-site-item:hover),
:global(.games-page.games-page--dark .latest-review-item:hover) {
  color: rgba(226, 232, 240, 0.90) !important;
}

:global(.games-page.games-page--dark .game-site-item:hover .text-gray-800),
:global(.games-page.games-page--dark .latest-review-item:hover .text-stone-800) {
  color: rgba(241, 245, 249, 0.92) !important;
}

:global(.games-page.games-page--dark .stats-type-tab--active),
:global(.games-page.games-page--dark .stats-page-tab--active) {
  color: rgba(226, 232, 240, 0.86) !important;
}

:global(.games-page.games-page--dark .stats-type-tab--active::after) {
  background: rgba(203, 213, 225, 0.58) !important;
}
</style>
