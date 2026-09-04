<template>
  <div
    class="games-page game-detail-page relative isolate min-h-full w-full overflow-hidden"
  >
    <div class="game-detail-layout relative z-10 mx-auto flex w-full max-w-[1700px] gap-4 p-6">
      <section class="w-full xl:w-[75%]">
        <GameDetailMain
          :game="gameDetailData.gameBaseInfo"
          :remark="gameDetailData.remarkInfo"
          :recommend="gameDetailData.recommendedGame"
          :game-id="gameId"
          :insights="gameInsightsSnapshot.insights"
          :insights-unavailable="gameInsightsSnapshot.unavailable"
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
import { computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import GameDetailMain from '@/components/game/detail/GameDetailMain.vue'
import GameDetailSidebar from '@/components/game/detail/GameDetailSidebar.vue'
import { getGameBaseInfo, getGameInsights, getGameRemark, getRecommendedGame, touchGameView } from '~/services/game'
import type { GameBaseInfoResponse, RecommendedModel, RemarkResponse } from '~/types/game'
import type { GameInsights } from '~/types/insights'
import { authoritativePageStatus } from '~/utils/authoritativePageError'
import { parsePositiveEntityRouteId } from '~/utils/routeIdentity'
import { buildGameDetailSeo } from '~/utils/seo'

definePageMeta({
  validate: route => parsePositiveEntityRouteId(route.params.id) !== null,
})

interface GameDetailPageData {
  gameBaseInfo: GameBaseInfoResponse | null
  recommendedGame: RecommendedModel[] | null
  remarkInfo: RemarkResponse | null
}

interface GameInsightsSnapshot {
  insights: GameInsights | null
  unavailable: boolean
}

const route = useRoute()
const { locale } = useI18n()
const touchedGameIds = new Set<string>()
const touchingGameIds = new Set<string>()

const initialGameId = parsePositiveEntityRouteId(route.params.id)
if (!initialGameId) {
  throw createError({
    statusCode: 404,
    statusMessage: 'Game not found',
  })
}

const gameId = computed(() => String(route.params.id ?? ''))
const lang = computed(() => (locale.value === 'en' ? 'en' : 'zh'))

const detailRequest = useAsyncData<GameDetailPageData>(
  () => `game-detail:${gameId.value}:${lang.value}`,
  async () => {
    const [gameBaseInfo, remarkInfo, recommendedGame] = await Promise.all([
      getGameBaseInfo(gameId.value, lang.value),
      getGameRemark(gameId.value, 1, 5).catch(() => null),
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

const insightsRequest = useAsyncData<GameInsightsSnapshot>(
  () => `game-insights:${gameId.value}`,
  async () => {
    try {
      return { insights: await getGameInsights(gameId.value), unavailable: false }
    } catch {
      return { insights: null, unavailable: true }
    }
  },
  {
    watch: [gameId],
    default: () => ({ insights: null, unavailable: true }),
  },
)

const [detailState, insightsState] = await Promise.all([detailRequest, insightsRequest])
if (detailState.error.value) {
  const statusCode = authoritativePageStatus(detailState.error.value, 'game')
  throw createError({
    statusCode,
    statusMessage: statusCode === 404 ? 'Game not found' : 'Game service temporarily unavailable',
    cause: detailState.error.value,
  })
}
const { data } = detailState

const gameDetailData = computed(() => data.value!)
const gameInsightsSnapshot = computed(() => insightsState.data.value!)
const seo = computed(() => buildGameDetailSeo({
  name: gameDetailData.value.gameBaseInfo?.name,
  description: gameDetailData.value.gameBaseInfo?.info,
  locale: locale.value,
}))
const seoImage = computed(() => gameDetailData.value.gameBaseInfo?.cover || undefined)

useSeoMeta({
  title: () => seo.value.title,
  description: () => seo.value.description,
  ogTitle: () => seo.value.title,
  ogDescription: () => seo.value.description,
  ogImage: () => seoImage.value,
  twitterCard: 'summary_large_image',
})

onMounted(() => {
  watch(
    [gameId, () => data.value?.gameBaseInfo?.appid],
    ([id, appid]) => {
      if (!id || !appid || touchedGameIds.has(id) || touchingGameIds.has(id)) {
        return
      }
      void touchCurrentGameView(id)
    },
    { immediate: true }
  )
})

async function touchCurrentGameView(id: string) {
  touchingGameIds.add(id)

  try {
    const response = await touchGameView(id)
    touchedGameIds.add(id)

    if (gameId.value === id && data.value?.gameBaseInfo && Number.isFinite(response.view_count)) {
      data.value.gameBaseInfo.view_count = response.view_count
    }
  } catch {
    // 浏览量统计是旁路副作用，失败不影响详情页主内容展示。
  } finally {
    touchingGameIds.delete(id)
  }
}
</script>
