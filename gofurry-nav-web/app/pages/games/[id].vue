<template>
  <div
    class="games-page game-detail-page relative isolate min-h-full w-full overflow-hidden"
  >
    <GoFurryGridBackground :fixed="false" palette="games" />
    <div class="game-detail-layout relative z-10 mx-auto flex w-full max-w-[1700px] gap-4 p-6">
      <section class="w-full xl:w-[75%]">
        <GameDetailMain
          :game="gameDetailData.gameBaseInfo"
          :remark="gameDetailData.remarkInfo"
        />
      </section>

      <aside class="hidden xl:block xl:w-[25%]">
        <GameDetailSidebar
          :game="gameDetailData.gameBaseInfo"
          :recommend="gameDetailData.recommendedGame"
        />
      </aside>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import GoFurryGridBackground from '@/components/common/GoFurryGridBackground.vue'
import GameDetailMain from '@/components/game/detail/GameDetailMain.vue'
import GameDetailSidebar from '@/components/game/detail/GameDetailSidebar.vue'
import { getGameBaseInfo, getGameRemark, getRecommendedGame } from '~/services/game'
import type { GameBaseInfoResponse, RecommendedModel, RemarkResponse } from '~/types/game'
import { steamLibraryCoverUrl } from '~/utils/gameAssets'
import { buildGameDetailSeo } from '~/utils/seo'

interface GameDetailPageData {
  gameBaseInfo: GameBaseInfoResponse | null
  recommendedGame: RecommendedModel[] | null
  remarkInfo: RemarkResponse | null
}

const route = useRoute()
const { locale } = useI18n()

const gameId = computed(() => String(route.params.id ?? ''))
const lang = computed(() => (locale.value === 'en' ? 'en' : 'zh'))

const { data } = await useAsyncData<GameDetailPageData>(
  () => `game-detail:${gameId.value}:${lang.value}`,
  async () => {
    const [gameBaseInfo, remarkInfo, recommendedGame] = await Promise.all([
      getGameBaseInfo(gameId.value, lang.value).catch(() => null),
      getGameRemark(gameId.value).catch(() => null),
      getRecommendedGame(gameId.value, lang.value).catch(() => null),
    ])

    return {
      gameBaseInfo,
      remarkInfo,
      recommendedGame,
    }
  },
  {
    watch: [gameId, lang],
    default: () => ({
      gameBaseInfo: null,
      remarkInfo: null,
      recommendedGame: null,
    }),
  }
)

const gameDetailData = computed(() => data.value!)
const seo = computed(() => buildGameDetailSeo({
  name: gameDetailData.value.gameBaseInfo?.name,
  description: gameDetailData.value.gameBaseInfo?.info,
  locale: locale.value,
}))
const seoImage = computed(() => steamLibraryCoverUrl(gameDetailData.value.gameBaseInfo?.appid))

useSeoMeta({
  title: () => seo.value.title,
  description: () => seo.value.description,
  ogTitle: () => seo.value.title,
  ogDescription: () => seo.value.description,
  ogImage: () => seoImage.value,
  twitterCard: 'summary_large_image',
})
</script>
