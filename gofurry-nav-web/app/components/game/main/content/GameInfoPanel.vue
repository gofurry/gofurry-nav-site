<template>
  <div class="game-info-shell mb-8 p-5">
    <GameInfoGroup
      v-if="firstGroup"
      :key="firstGroup.title"
      :group="firstGroup"
      @more="handleMore"
    />

    <GameStatsPanels
      v-if="panelData"
      :top-price-list="panelData.top_price_vo"
      :discount-list="panelData.top_discount_vo"
      :top-count-list="panelData.top_count"
      :bottom-price-list="panelData.bottom_price"
    />

    <GameInfoGroup
      v-for="group in middleGroups"
      :key="group.title"
      :group="group"
      @more="handleMore"
    />

    <GameUpdateNews :initial-news-record="initialNewsRecord" />

    <GameInfoGroup
      v-if="lastGroup"
      :key="lastGroup.title"
      :group="lastGroup"
      @more="handleMore"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import GameInfoGroup from '@/components/game/main/content/GameInfoGroup.vue'
import GameStatsPanels from '@/components/game/main/content/GameStatsPanels.vue'
import GameUpdateNews from '@/components/game/main/content/GameUpdateNews.vue'
import type { BaseGameInfoRecord, GameGroupRecord, GamePanelRecord, LatestNewsRecord } from '~/types/game'

interface GameItem {
  id: string
  name: string
  cover: string
  desc: string
  score: number
  scoreCount: number
}

interface GameGroupViewModel {
  title: string
  games: GameItem[]
}

const props = defineProps<{
  initialRawData?: GameGroupRecord | null
  initialPanelData?: GamePanelRecord | null
  initialNewsRecord?: LatestNewsRecord | null
}>()

const { locale } = useI18n()
const lang = computed(() => locale.value === 'en' ? 'en' : 'zh')
const panelData = computed(() => props.initialPanelData ?? null)

function mapGames(list: BaseGameInfoRecord[], currentLang: string): GameItem[] {
  return list.map((game) => ({
    id: game.game_id,
    name: currentLang === 'en' ? game.name_en : game.name,
    cover: game.header,
    desc: currentLang === 'en' ? game.info_en : game.info,
    score: game.avg_score,
    scoreCount: game.comment_count,
  }))
}

const groups = computed<GameGroupViewModel[]>(() => {
  const rawData = props.initialRawData
  if (!rawData) {
    return []
  }

  const currentLang = lang.value

  return [
    {
      title: currentLang === 'en' ? 'Latest Release' : '最近发售',
      games: mapGames(rawData.latest, currentLang),
    },
    {
      title: currentLang === 'en' ? 'Recently Added' : '最近收录',
      games: mapGames(rawData.recent, currentLang),
    },
    {
      title: currentLang === 'en' ? 'Free to Play' : '免费专区',
      games: mapGames(rawData.free, currentLang),
    },
    {
      title: currentLang === 'en' ? 'Hot Ranking' : '热门排行',
      games: mapGames(rawData.hot, currentLang),
    },
  ]
})

const firstGroup = computed(() => groups.value[0] || null)
const middleGroups = computed(() => groups.value.slice(1, groups.value.length - 1))
const lastGroup = computed(() => (groups.value.length > 1 ? groups.value[groups.value.length - 1] : null))

function handleMore(group: GameGroupViewModel) {
  console.log('show more:', group.title)
}
</script>
