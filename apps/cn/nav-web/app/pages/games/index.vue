<template>
  <div
    class="games-page flex min-h-full w-full flex-col overflow-clip"
  >
    <main
      class="relative isolate flex-1 overflow-hidden"
    >
      <GoFurryGridBackground :fixed="false" palette="games" />
      <FallingLeavesCanvas
        v-if="ambientReady"
        :key="isDarkTheme ? 'dark-leaves' : 'light-leaves'"
        class="z-[1]"
        mode="viewport"
        :palette="isDarkTheme ? 'bright' : 'warm'"
        :leaf-count="28"
        :mobile-leaf-count="16"
        :fps="24"
      />
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
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import FallingLeavesCanvas from '@/components/common/FallingLeavesCanvas.vue'
import GoFurryGridBackground from '@/components/common/GoFurryGridBackground.vue'
import GameInfoPanel from '@/components/game/main/content/GameInfoPanel.vue'
import GameToolDock from '@/components/game/main/GameToolDock.vue'
import SideBarPanel from '@/components/game/main/sidebar/SideBarPanel.vue'
import { useThemeStore } from '@/stores/theme'
import { getGameHomeData, type GameHomeData } from '~/services/game'

const { locale } = useI18n()
const themeStore = useThemeStore()
const lang = computed(() => (locale.value === 'en' ? 'en' : 'zh'))
const isDarkTheme = computed(() => themeStore.theme === 'dark')
const ambientReady = ref(false)
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

const { data } = await useAsyncData<GameHomeData | null>(
  () => `games-page:${lang.value}`,
  async () => {
    return getGameHomeData(lang.value).catch(() => null)
  },
  {
    watch: [lang],
    default: () => null,
  }
)

const gamesPageData = computed<GameHomeData>(() => data.value ?? {
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

onMounted(() => {
  themeStore.initTheme()
  ambientReady.value = true
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
