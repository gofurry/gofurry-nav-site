<template>
  <div
      class="min-h-200 bg-[#f2e3d0]"
      :style="{
      backgroundImage: `url(${bgGrid})`,
      backgroundRepeat: 'repeat'
    }"
  >
    <div class="w-full max-w-[1700px] mx-auto flex gap-4 p-6">

      <section class="w-full xl:w-[75%]">
        <GameDetailMain
            :game="gameBaseInfo"
            :remark="remarkInfo"
        />
      </section>

      <aside class="hidden xl:block xl:w-[25%]">
        <GameDetailSidebar
            :game="gameBaseInfo"
            :recommend="recommendedGame"
        />
      </aside>

    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'

import bgGrid from '@/assets/pngs/bg-grid.png'

import GameDetailMain from '@/components/game/detail/GameDetailMain.vue'
import GameDetailSidebar from '@/components/game/detail/GameDetailSidebar.vue'

import {getGameBaseInfo, getGameRemark, getRecommendedGame} from '@/utils/api/game'
import type {GameBaseInfoResponse, RecommendedModel, RemarkResponse} from '@/types/game'

import { useLangStore } from '@/store/langStore'

const route = useRoute()
const langStore = useLangStore()

const lang = ref(langStore.lang)

const gameBaseInfo = ref<GameBaseInfoResponse | null>(null)
const recommendedGame = ref<RecommendedModel[] | null>(null)
const remarkInfo = ref<RemarkResponse | null>(null)

async function fetchGameBaseInfo() {
  try {
    const gameId = route.params.id as string
    gameBaseInfo.value = await getGameBaseInfo(gameId, lang.value)
  } catch (err) {
    console.error('[GameDetail] fetchGameBaseInfo failed:', err)
  }
}

async function fetchRecommendedGame() {
  try {
    const gameId = route.params.id as string
    recommendedGame.value = await getRecommendedGame(gameId, lang.value)
  } catch (err) {
    console.error('[GameDetail] fetchGameBaseInfo failed:', err)
  }
}

async function fetchGameRemark() {
  try {
    const gameId = route.params.id as string
    remarkInfo.value = await getGameRemark(gameId)
  } catch (err) {
    console.error('[GameDetail] fetchGameRemark failed:', err)
  }
}

function clearDetail() {
  gameBaseInfo.value = null
  recommendedGame.value = null
  remarkInfo.value = null
}

watch(
    () => langStore.lang,
    async (val) => {
      lang.value = val
      await fetchGameBaseInfo()
      await fetchRecommendedGame()
    }
)

watch(
    () => route.params.id,
    async (id) => {
      if (id) {
        clearDetail()
        await fetchGameBaseInfo()
        await fetchGameRemark()
        await fetchRecommendedGame()
      }
    }
)

onMounted(async () => {
  await fetchGameBaseInfo()
  await fetchGameRemark()
  await fetchRecommendedGame()
})
</script>
