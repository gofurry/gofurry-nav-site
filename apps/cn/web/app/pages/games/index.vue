<template>
  <div
    class="games-page flex min-h-full w-full flex-col overflow-clip"
  >
    <main
      class="relative isolate flex-1 overflow-hidden"
    >
      <h1 class="sr-only">{{ gamesPageSeo.heading }}</h1>
      <div class="relative z-10 mx-auto flex w-full max-w-[1700px] gap-4 p-6">
        <section class="w-full xl:w-[75%]">
          <GameInfoPanel
            :initial-raw-data="gamesPageData.mainInfo"
            :initial-panel-data="gamesPageData.panelData"
            :initial-news-record="gamesPageData.latestNews"
            :showcase="showcase ?? emptyGameShowcase()"
          />
        </section>

        <aside class="hidden xl:block xl:w-[25%]">
          <SideBarPanel :initial-reviews="gamesPageData.latestReviews" :collections="collections" />
        </aside>
      </div>
    </main>

    <GameToolDock />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import GameInfoPanel from '@/components/game/main/content/GameInfoPanel.vue'
import GameToolDock from '@/components/game/main/GameToolDock.vue'
import SideBarPanel from '@/components/game/main/sidebar/SideBarPanel.vue'
import { getGameHomeData, getGameHomeShowcase, getGameCollectionHome, type GameHomeData } from '~/services/game'
import { emptyGameShowcase } from '~/utils/gameShowcasePresentation'
import type { GameShowcaseSnapshot, GameCollectionHome } from '~/types/game'

const { locale } = useI18n()
const lang = computed(() => (locale.value === 'en' ? 'en' : 'zh'))
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

interface GamesPageData { lang: string; home: GameHomeData | null; showcase: GameShowcaseSnapshot | null; collections: GameCollectionHome | null }

const { data } = await useAsyncData<GamesPageData>(
  () => `games-page:${lang.value}`,
  async () => {
    const requestedLang = lang.value
    const [home, showcase, collections] = await Promise.allSettled([
      getGameHomeData(requestedLang),
      getGameHomeShowcase(requestedLang),
      getGameCollectionHome(requestedLang),
    ])
    return {
      lang: requestedLang,
      home: home.status === 'fulfilled' ? home.value : null,
      showcase: showcase.status === 'fulfilled' ? showcase.value : null,
      collections: collections.status === 'fulfilled' ? collections.value : null,
    }
  },
  {
    watch: [lang],
    default: () => ({ lang: '', home: null, showcase: null, collections: null }),
  }
)

const showcase = useGameHomeOptionalRecovery(lang,
  computed(() => data.value?.lang === lang.value ? data.value.showcase : undefined),
  (locale, signal) => getGameHomeShowcase(locale, { timeout: 8000, signal }),
)
const collections = useGameHomeOptionalRecovery(lang,
  computed(() => data.value?.lang === lang.value ? data.value.collections : undefined),
  (locale, signal) => getGameCollectionHome(locale, { timeout: 8000, signal }),
)

const gamesPageData = computed<GameHomeData>(() => data.value?.home ?? {
  mainInfo: nullGameGroups(),
  panelData: nullGamePanel(),
  latestNews: { news_zh: [], news_en: [] },
  latestReviews: [],
})

useSeoMeta({
  title: () => gamesPageSeo.value.title,
  description: () => gamesPageSeo.value.description,
  ogTitle: () => gamesPageSeo.value.title,
  ogDescription: () => gamesPageSeo.value.description,
})

function nullGameGroups() {
  return {
    latest: [],
    recent: [],
    hot: [],
    free: [],
  }
}

function nullGamePanel() {
  return {
    top_count: { one: [], two: [], three: [], four: [] },
    top_discount_vo: [],
    top_price_vo: [],
    bottom_price: { one: [], two: [], three: [], four: [] },
  }
}
</script>
