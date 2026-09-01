<template>
  <section
    class="game-insights-tab"
    data-game-insights
    :data-player-loaded-ranges="playerLoadedRanges"
    :data-price-loaded-ranges="priceLoadedRanges"
  >
    <div class="entity-insights-heading game-insights-heading">
      <div>
        <p class="entity-insights-eyebrow">{{ $t('insights.entity.gameEyebrow') }}</p>
        <h2>{{ $t('insights.entity.gameTitle') }}</h2>
        <p>{{ $t('insights.entity.gameDescription') }}</p>
      </div>
      <NuxtLink :to="localePath('/insights/games')" class="entity-insights-back-link">
        {{ $t('insights.entity.backToGames') }}
      </NuxtLink>
    </div>

    <div v-if="summaryUnavailable || !summary" class="game-insights-summary entity-insights-empty" data-game-summary-unavailable>
      {{ $t('insights.emptyStates.unavailable') }}
    </div>
    <div v-else class="game-insights-summary" data-game-summary>
      <article>
        <span>{{ $t('insights.entity.currentPlayers') }}</span>
        <strong data-current-players>{{ formatPlayers(summary.players.current) }}</strong>
        <small>{{ formatAsOf(summary.players.as_of) }}</small>
      </article>
      <article>
        <span>{{ $t('insights.entity.peak30d') }}</span>
        <strong>{{ formatPlayers(summary.players.peak_30d) }}</strong>
        <small>{{ $t('insights.ranges.30d') }}</small>
      </article>
      <article>
        <span>{{ $t('insights.entity.cnPrice') }}</span>
        <strong :data-price-kind="summaryPrice.kind">{{ summaryPriceText }}</strong>
        <small>{{ formatAsOf(summary.price?.as_of ?? null) }}</small>
      </article>
      <article>
        <span>{{ $t('insights.entity.platforms') }}</span>
        <strong>{{ platformSummary }}</strong>
        <small>{{ formatAsOf(summary.state.as_of) }}</small>
      </article>
      <article>
        <span>{{ $t('insights.entity.freeState') }}</span>
        <strong>{{ freeStateText }}</strong>
        <small>{{ formatAsOf(summary.state.as_of) }}</small>
      </article>
      <article>
        <span>{{ $t('insights.entity.releaseState') }}</span>
        <strong>{{ releaseStateText }}</strong>
        <small>{{ formatAsOf(summary.state.as_of) }}</small>
      </article>
    </div>

    <div class="game-insights-range-row">
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

    <div class="game-insights-history-grid">
      <GamePlayerTrend
        :points="displayedPlayers?.points ?? []"
        :loading="playerLoading[selectedRange]"
        :unavailable="playerFailed[selectedRange]"
        @retry="loadPlayers(selectedRange, true)"
      />
      <GamePriceHistory
        :points="displayedPrices?.points ?? []"
        :loading="priceLoading[selectedRange]"
        :unavailable="priceFailed[selectedRange]"
        @retry="loadPrices(selectedRange, true)"
      />
    </div>

    <InsightsEntityTimeline
      :items="summary?.recent_changes ?? []"
      :unavailable="summaryUnavailable"
    />
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import GamePlayerTrend from '@/components/game/detail/insights/GamePlayerTrend.vue'
import GamePriceHistory from '@/components/game/detail/insights/GamePriceHistory.vue'
import InsightsEntityTimeline from '@/components/insights/InsightsEntityTimeline.vue'
import { getGameInsightPlayers, getGameInsightPrices } from '@/services/game'
import type { GameInsightPlayerHistory, GameInsightPriceHistory, GameInsights, InsightRange } from '@/types/insights'
import { formatCnyMinorAmount, publicPriceDisplay } from '@/utils/insightPrices'

type LoadingState = Record<InsightRange, boolean>
type FailedState = Record<InsightRange, boolean>

const props = defineProps<{
  gameId: string
  summary: GameInsights | null
  summaryUnavailable?: boolean
}>()

const localePath = useLocalePath()
const { locale, t } = useI18n()
const ranges: InsightRange[] = ['30d', '90d', 'all']
const selectedRange = ref<InsightRange>('30d')
const playerCache = ref<Partial<Record<InsightRange, GameInsightPlayerHistory>>>({})
const priceCache = ref<Partial<Record<InsightRange, GameInsightPriceHistory>>>({})
const playerLoading = ref<LoadingState>(emptyFlags())
const priceLoading = ref<LoadingState>(emptyFlags())
const playerFailed = ref<FailedState>(emptyFlags())
const priceFailed = ref<FailedState>(emptyFlags())
const lastPlayers = ref<GameInsightPlayerHistory | null>(null)
const lastPrices = ref<GameInsightPriceHistory | null>(null)
let generation = 0

const displayedPlayers = computed(() => playerCache.value[selectedRange.value] ?? lastPlayers.value)
const displayedPrices = computed(() => priceCache.value[selectedRange.value] ?? lastPrices.value)
const playerLoadedRanges = computed(() => ranges.filter(range => playerCache.value[range]).join(','))
const priceLoadedRanges = computed(() => ranges.filter(range => priceCache.value[range]).join(','))
const summaryPrice = computed(() => publicPriceDisplay(props.summary?.price ?? null))
const summaryPriceText = computed(() => {
  if (summaryPrice.value.kind === 'free') return t('insights.entity.priceFree')
  if (summaryPrice.value.kind === 'priced') {
    return `${t('insights.entity.pricePriced')} · ${formatCnyMinorAmount(summaryPrice.value.amount, locale.value)}`
  }
  return t('insights.entity.priceUnavailable')
})
const platformSummary = computed(() => {
  if (!props.summary) return '—'
  return [
    platformLabel('Windows', props.summary.state.windows),
    platformLabel('Linux', props.summary.state.linux),
  ].join(' · ')
})
const freeStateText = computed(() => {
  if (props.summary?.state.free === null || props.summary?.state.free === undefined) return t('insights.entity.booleanUnknown')
  return props.summary.state.free ? t('insights.entity.freeGame') : t('insights.entity.paidGame')
})
const releaseStateText = computed(() => {
  const state = props.summary?.state.release
  if (!state) return t('insights.entity.booleanUnknown')
  const knownStates: Record<string, string> = {
    available: 'available',
    upcoming: 'upcoming',
    withdrawn: 'withdrawn',
  }
  const key = knownStates[state]
  return key ? t(`insights.entity.releaseStates.${key}`) : state
})

function emptyFlags(): LoadingState {
  return { '30d': false, '90d': false, all: false }
}

function formatPlayers(value: number | null) {
  return value === null ? '—' : new Intl.NumberFormat(locale.value).format(value)
}

function formatAsOf(value: string | null) {
  return value ? t('insights.entity.asOf', { date: value.slice(0, 10) }) : t('insights.entity.asOfUnavailable')
}

function platformLabel(name: string, value: boolean | null) {
  if (value === null) return `${name}: ${t('insights.entity.booleanUnknown')}`
  return `${name}: ${value ? t('insights.entity.supported') : t('insights.entity.unsupported')}`
}

function selectRange(range: InsightRange) {
  if (range === selectedRange.value) return
  selectedRange.value = range
  void loadRange(range)
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

async function loadPrices(range: InsightRange, force = false) {
  if ((!force && priceCache.value[range]) || priceLoading.value[range]) return
  const requestGeneration = generation
  const requestGameId = props.gameId
  priceLoading.value[range] = true
  priceFailed.value[range] = false
  try {
    const response = await getGameInsightPrices(requestGameId, range)
    if (requestGeneration !== generation || requestGameId !== props.gameId) return
    priceCache.value = { ...priceCache.value, [range]: response }
    if (selectedRange.value === range) lastPrices.value = response
  } catch {
    if (requestGeneration === generation && requestGameId === props.gameId) priceFailed.value[range] = true
  } finally {
    if (requestGeneration === generation && requestGameId === props.gameId) priceLoading.value[range] = false
  }
}

async function loadRange(range: InsightRange) {
  await Promise.allSettled([loadPlayers(range), loadPrices(range)])
}

function resetForGame() {
  generation += 1
  selectedRange.value = '30d'
  playerCache.value = {}
  priceCache.value = {}
  playerLoading.value = emptyFlags()
  priceLoading.value = emptyFlags()
  playerFailed.value = emptyFlags()
  priceFailed.value = emptyFlags()
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
