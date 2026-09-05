<template>
  <section
    class="game-insights-tab"
    data-game-insights
    :data-player-loaded-ranges="playerLoadedRanges"
    :data-price-loaded-ranges="priceLoadedRanges"
    :data-price-region="selectedRegion"
  >
    <div class="entity-insights-heading game-insights-heading">
      <h2>{{ $t('insights.entity.gameTitle') }}</h2>
      <NuxtLink :to="localePath('/insights/games')" class="entity-insights-back-link">
        {{ $t('insights.entity.backToGames') }}
      </NuxtLink>
    </div>

    <div v-if="summaryUnavailable || !summary" class="game-insights-overview entity-insights-empty" data-game-summary-unavailable>
      {{ $t('insights.emptyStates.unavailable') }}
    </div>
    <GameInsightsOverview v-else :summary="summary" />

    <section class="game-insights-history-section">
      <h3>{{ $t('insights.entity.historyTrends') }}</h3>
      <div class="game-insights-controls">
        <div class="game-insights-control-group">
          <span>{{ $t('insights.entity.priceRegion') }}</span>
          <div class="insights-ranges" :aria-label="$t('insights.entity.priceRegion')">
            <button
              v-for="region in regions"
              :key="region"
              type="button"
              :class="{ 'insights-ranges__button--active': selectedRegion === region }"
              :aria-pressed="selectedRegion === region"
              @click="selectRegion(region)"
            >
              {{ $t(`insights.regions.${region}`) }}
            </button>
          </div>
        </div>
        <div class="game-insights-control-group game-insights-control-group--range">
          <span>{{ $t('insights.entity.historyRange') }}</span>
          <div class="insights-ranges" :aria-label="$t('insights.entity.historyRange')">
            <button
              v-for="option in ranges"
              :key="option"
              type="button"
              :class="{ 'insights-ranges__button--active': option === selectedRange }"
              :aria-pressed="option === selectedRange"
              :data-game-insights-range="option"
              @click="selectRange(option)"
            >
              {{ $t(`insights.ranges.${option}`) }}
            </button>
          </div>
        </div>
      </div>

      <div class="game-insights-history-grid">
        <GamePlayerTrend
          :points="displayedPlayers?.points ?? []"
          :loading="playerLoading[selectedRange]"
          :unavailable="playerFailed[selectedRange]"
          @retry="loadPlayers(selectedRange, true)"
        />
        <GamePriceHistory
          :points="displayedPrices?.points ?? []"
          :region="selectedRegion"
          :loading="priceLoading[priceKey(selectedRegion, selectedRange)]"
          :unavailable="priceFailed[priceKey(selectedRegion, selectedRange)]"
          @retry="loadPrices(selectedRegion, selectedRange, true)"
        />
      </div>
    </section>

    <GameInsightsTimeline
      :items="summary?.recent_changes ?? []"
      :unavailable="summaryUnavailable"
    />
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import GameInsightsOverview from '@/components/game/detail/insights/GameInsightsOverview.vue'
import GameInsightsTimeline from '@/components/game/detail/insights/GameInsightsTimeline.vue'
import GamePlayerTrend from '@/components/game/detail/insights/GamePlayerTrend.vue'
import GamePriceHistory from '@/components/game/detail/insights/GamePriceHistory.vue'
import { getGameInsightPlayers, getGameInsightPrices } from '@/services/game'
import type { GameInsightPlayerHistory, GameInsightPriceHistory, GameInsightRegion, GameInsights, InsightRange } from '@/types/insights'

type LoadingState = Record<InsightRange, boolean>
type FailedState = Record<string, boolean>

const props = defineProps<{
  gameId: string
  summary: GameInsights | null
  summaryUnavailable?: boolean
}>()

const localePath = useLocalePath()
const ranges: InsightRange[] = ['30d', '90d', 'all']
const regions: GameInsightRegion[] = ['CN', 'US', 'HK']
const selectedRange = ref<InsightRange>('30d')
const selectedRegion = ref<GameInsightRegion>('CN')
const playerCache = ref<Partial<Record<InsightRange, GameInsightPlayerHistory>>>({})
const priceCache = ref<Record<string, GameInsightPriceHistory>>({})
const playerLoading = ref<LoadingState>(emptyFlags())
const priceLoading = ref<FailedState>(emptyPriceFlags())
const playerFailed = ref<FailedState>(emptyFlags())
const priceFailed = ref<FailedState>(emptyPriceFlags())
const lastPlayers = ref<GameInsightPlayerHistory | null>(null)
const lastPrices = ref<GameInsightPriceHistory | null>(null)
let generation = 0

const displayedPlayers = computed(() => playerCache.value[selectedRange.value] ?? lastPlayers.value)
const displayedPrices = computed(() => priceCache.value[priceKey(selectedRegion.value, selectedRange.value)] ?? lastPrices.value)
const playerLoadedRanges = computed(() => ranges.filter(range => playerCache.value[range]).join(','))
const priceLoadedRanges = computed(() => Object.keys(priceCache.value).join(','))

function emptyFlags(): LoadingState {
  return { '30d': false, '90d': false, all: false }
}

function emptyPriceFlags(): FailedState { return {} }
function priceKey(region: GameInsightRegion, range: InsightRange) { return `${region}:${range}` }

function selectRange(range: InsightRange) {
  if (range === selectedRange.value) return
  selectedRange.value = range
  void loadRange(range)
}

function selectRegion(region: GameInsightRegion) {
  if (region === selectedRegion.value) return
  selectedRegion.value = region
  void loadPrices(region, selectedRange.value)
}

async function loadPlayers(range: InsightRange, force = false) {
  if ((!force && playerCache.value[range]) || playerLoading.value[range]) return
  const requestGeneration = generation
  const requestGameId = props.gameId
  playerLoading.value[range] = true
  playerFailed.value[range] = false
  try {
    const response = await getGameInsightPlayers(requestGameId, range)
    if (requestGeneration !== generation || requestGameId !== props.gameId) return
    playerCache.value = { ...playerCache.value, [range]: response }
    if (selectedRange.value === range) lastPlayers.value = response
  } catch {
    if (requestGeneration === generation && requestGameId === props.gameId) playerFailed.value[range] = true
  } finally {
    if (requestGeneration === generation && requestGameId === props.gameId) playerLoading.value[range] = false
  }
}

async function loadPrices(region: GameInsightRegion, range: InsightRange, force = false) {
  const key = priceKey(region, range)
  if ((!force && priceCache.value[key]) || priceLoading.value[key]) return
  const requestGeneration = generation
  const requestGameId = props.gameId
  priceLoading.value[key] = true
  priceFailed.value[key] = false
  try {
    const response = await getGameInsightPrices(requestGameId, region, range)
    if (requestGeneration !== generation || requestGameId !== props.gameId) return
    priceCache.value = { ...priceCache.value, [key]: response }
    if (selectedRange.value === range && selectedRegion.value === region) lastPrices.value = response
  } catch {
    if (requestGeneration === generation && requestGameId === props.gameId) priceFailed.value[key] = true
  } finally {
    if (requestGeneration === generation && requestGameId === props.gameId) priceLoading.value[key] = false
  }
}

async function loadRange(range: InsightRange) {
  await Promise.allSettled([loadPlayers(range), loadPrices(selectedRegion.value, range)])
}

function resetForGame() {
  generation += 1
  selectedRange.value = '30d'
  selectedRegion.value = 'CN'
  playerCache.value = {}
  priceCache.value = {}
  playerLoading.value = emptyFlags()
  priceLoading.value = emptyPriceFlags()
  playerFailed.value = emptyFlags()
  priceFailed.value = emptyPriceFlags()
  lastPlayers.value = null
  lastPrices.value = null
}

onMounted(() => {
  void loadRange(selectedRange.value)
})

watch(() => props.gameId, () => {
  resetForGame()
  void loadRange('30d')
})
</script>
