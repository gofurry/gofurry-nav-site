<template>
  <div
    class="games-page game-detail-page relative isolate min-h-full w-full overflow-hidden"
  >
    <div class="game-detail-layout relative z-10 mx-auto flex w-full max-w-[1700px] gap-4 p-6">
      <section class="min-w-0 w-full max-w-full xl:w-[75%]">
        <GameDetailMain
          :game="gameDetailData.gameBaseInfo"
          :remark="gameDetailData.remarkInfo"
          :remark-unavailable="gameDetailData.remarkUnavailable"
          :remark-loading="retrying.reviews"
          :recommend="gameDetailData.recommendedGame"
          :recommend-unavailable="gameDetailData.recommendUnavailable"
          :recommend-loading="retrying.recommendations"
          @retry-reviews="retryReviews"
          @retry-recommendations="retryRecommendations"
          :game-id="gameId"
          :insights="gameInsightsSnapshot.insights"
          :insights-unavailable="gameInsightsSnapshot.unavailable"
        />
      </section>

      <aside class="hidden min-w-0 xl:block xl:w-[25%]">
        <GameDetailSidebar
          :game="gameDetailData.gameBaseInfo"
          :recommend="gameDetailData.recommendedGame"
          :recommend-unavailable="gameDetailData.recommendUnavailable"
          :recommend-loading="retrying.recommendations"
          @retry-recommendations="retryRecommendations"
        />
      </aside>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount, reactive, watch } from 'vue'
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
  remarkUnavailable: boolean
  recommendUnavailable: boolean
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
      remarkUnavailable: remarkInfo === null,
      recommendUnavailable: recommendedGame === null,
    }
  },
  {
    // Locale navigation can mount a second consumer of the same pending identity.
    dedupe: 'defer',
    watch: [gameId, lang],
    default: () => ({
      gameBaseInfo: null,
      remarkInfo: null,
      recommendedGame: null,
      remarkUnavailable: false,
      recommendUnavailable: false,
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
const retrying = reactive({ reviews: false, recommendations: false })
let retryGeneration = 0
watch([gameId, lang], () => {
  retryGeneration += 1
  retrying.reviews = false; retrying.recommendations = false
}, { flush: 'sync' })
onBeforeUnmount(() => { retryGeneration += 1 })

async function retryReviews() {
  if (retrying.reviews) return
  const generation = retryGeneration, id = gameId.value, requestLang = lang.value
  retrying.reviews = true
  try {
    const response = await getGameRemark(id, 1, 5)
    if (generation === retryGeneration && id === gameId.value && requestLang === lang.value && data.value) {
      data.value = { ...data.value, remarkInfo: response, remarkUnavailable: false }
    }
  } catch {
    // The unavailable slice remains visible and retryable; the main page stays intact.
  } finally {
    if (generation === retryGeneration) retrying.reviews = false
  }
}

async function retryRecommendations() {
  if (retrying.recommendations) return
  const generation = retryGeneration, id = gameId.value, requestLang = lang.value
  retrying.recommendations = true
  try {
    const response = await getRecommendedGame(id, requestLang)
    if (generation === retryGeneration && id === gameId.value && requestLang === lang.value && data.value) {
      data.value = { ...data.value, recommendedGame: response, recommendUnavailable: false }
    }
  } catch {
    // Do not turn an unavailable recommendation slice into a successful empty list.
  } finally {
    if (generation === retryGeneration) retrying.recommendations = false
  }
}

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
