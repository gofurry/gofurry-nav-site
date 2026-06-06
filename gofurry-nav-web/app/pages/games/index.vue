<template>
  <div class="flex min-h-full w-full flex-col overflow-clip bg-gray-50">
    <main
      class="flex-1 bg-[#f2e3d0]"
      :style="{
        backgroundImage: `url(${bgGrid})`,
        backgroundRepeat: 'repeat'
      }"
    >
      <h1 class="sr-only">{{ gamesPageSeo.heading }}</h1>
      <div class="mx-auto flex w-full max-w-[1700px] gap-4 p-6">
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
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import bgGrid from '@/assets/pngs/bg-grid.png'
import GameInfoPanel from '@/components/game/main/content/GameInfoPanel.vue'
import GameToolDock from '@/components/game/main/GameToolDock.vue'
import SideBarPanel from '@/components/game/main/sidebar/SideBarPanel.vue'
import { getGameMainInfo, getGameMainPanel, getLatestGameNews, getLatestReview } from '~/services/game'
import type { AnonymousReviewModel, GameGroupRecord, GamePanelRecord, LatestNewsRecord } from '~/types/game'

interface GamesPageData {
  mainInfo: GameGroupRecord | null
  panelData: GamePanelRecord | null
  latestNews: LatestNewsRecord | null
  latestReviews: AnonymousReviewModel[]
}

const { locale } = useI18n()
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
  'games-page',
  async () => {
    const [mainInfo, panelData, latestNews, latestReviews] = await Promise.all([
      getGameMainInfo().catch(() => null),
      getGameMainPanel().catch(() => null),
      getLatestGameNews().catch(() => null),
      getLatestReview().catch(() => []),
    ])

    return {
      mainInfo,
      panelData,
      latestNews,
      latestReviews,
    }
  },
  {
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
</script>
