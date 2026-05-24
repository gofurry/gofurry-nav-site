<template>
  <div
    class="min-h-200 bg-[#f2e3d0]"
    :style="{
      backgroundImage: `url(${bgGrid})`,
      backgroundRepeat: 'repeat'
    }"
  >
    <div class="mx-auto flex w-full max-w-[1700px] gap-4 p-6">
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
import bgGrid from '@/assets/pngs/bg-grid.png'
import GameDetailMain from '@/components/game/detail/GameDetailMain.vue'
import GameDetailSidebar from '@/components/game/detail/GameDetailSidebar.vue'
import { getGameBaseInfo, getGameRemark, getRecommendedGame } from '~/services/game'
import type { GameBaseInfoResponse, RecommendedModel, RemarkResponse } from '~/types/game'

interface GameDetailPageData {
  gameBaseInfo: GameBaseInfoResponse | null
  recommendedGame: RecommendedModel[] | null
  remarkInfo: RemarkResponse | null
}

const route = useRoute()
const { locale, t } = useI18n()

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
const seoTitle = computed(() => {
  const name = gameDetailData.value.gameBaseInfo?.name?.trim()
  return name ? `${name} - GoFurry` : `${t('sidebar.games')} - GoFurry`
})
const seoDescription = computed(() => {
  const description = gameDetailData.value.gameBaseInfo?.info?.trim() ?? ''
  return description.slice(0, 160)
})
const seoImage = computed(() => gameDetailData.value.gameBaseInfo?.cover || '')

useSeoMeta({
  title: () => seoTitle.value,
  description: () => seoDescription.value,
  ogTitle: () => seoTitle.value,
  ogDescription: () => seoDescription.value,
  ogImage: () => seoImage.value,
  twitterCard: 'summary_large_image',
})
</script>
