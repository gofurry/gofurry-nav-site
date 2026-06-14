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
import { computed, onMounted, ref, watch } from 'vue'
import GameInfoGroup from '@/components/game/main/content/GameInfoGroup.vue'
import GameStatsPanels from '@/components/game/main/content/GameStatsPanels.vue'
import GameUpdateNews from '@/components/game/main/content/GameUpdateNews.vue'
import { useLangStore } from '@/store/langStore'
import { getGameHomeData } from '~/services/game'
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

const langStore = useLangStore()
const lang = computed(() => langStore.lang)

const rawData = ref<GameGroupRecord | null>(props.initialRawData ?? null)
const panelData = ref<GamePanelRecord | null>(props.initialPanelData ?? null)
const groups = ref<GameGroupViewModel[]>([])

const firstGroup = computed(() => groups.value[0] || null)
const middleGroups = computed(() => groups.value.slice(1, groups.value.length - 1))
const lastGroup = computed(() => (groups.value.length > 1 ? groups.value[groups.value.length - 1] : null))

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

function updateGroups() {
  if (!rawData.value) {
    groups.value = []
    return
  }

  const currentLang = lang.value

  groups.value = [
    {
      title: currentLang === 'en' ? 'Latest Release' : '最近发售',
      games: mapGames(rawData.value.latest, currentLang),
    },
    {
      title: currentLang === 'en' ? 'Recently Added' : '最近收录',
      games: mapGames(rawData.value.recent, currentLang),
    },
    {
      title: currentLang === 'en' ? 'Free to Play' : '免费专区',
      games: mapGames(rawData.value.free, currentLang),
    },
    {
      title: currentLang === 'en' ? 'Hot Ranking' : '热门排行',
      games: mapGames(rawData.value.hot, currentLang),
    },
  ]
}

async function loadGameInfoPanel() {
  try {
    const homeData = await getGameHomeData(lang.value)

    rawData.value = homeData.mainInfo
    panelData.value = homeData.panelData
    updateGroups()
  } catch (error) {
    console.error('Failed to load games page content:', error)
  }
}

watch(
  () => langStore.lang,
  () => {
    loadGameInfoPanel()
  }
)

if (rawData.value) {
  updateGroups()
}

onMounted(() => {
  if (!rawData.value || !panelData.value) {
    loadGameInfoPanel()
  }
})

function handleMore(group: GameGroupViewModel) {
  console.log('show more:', group.title)
}
</script>
