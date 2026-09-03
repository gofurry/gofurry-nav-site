<template>
  <div
    class="insights-page insights-compare-page"
    data-game-compare
    :data-compare-count="selectedIDs.length"
    :data-compare-region="selectedRegion"
    :data-compare-status="compare?.status || (ready ? 'loading' : 'builder')"
  >
    <main class="insights-container">
      <EcosystemNavigation context="game" />
      <h1 class="sr-only">{{ $t('insights.gameCompare.title') }}</h1>

      <section class="compare-builder">
        <h2>{{ $t('insights.compare.builderTitle') }}</h2>
        <p>{{ builderHint }}</p>
        <form class="compare-builder__form" @submit.prevent="applySelection">
          <label>
            <span>{{ $t('insights.compare.idsLabel') }}</span>
            <input v-model="input" inputmode="numeric" autocomplete="off" :placeholder="$t('insights.compare.idsPlaceholder')" />
          </label>
          <button type="submit">{{ $t('insights.compare.apply') }}</button>
        </form>
        <div class="insights-ranges intelligence-selector compare-regions" :aria-label="$t('insights.gameCompare.region')">
          <button v-for="region in regions" :key="region" :class="{ 'insights-ranges__button--active': selectedRegion === region }" @click="selectRegion(region)">
            {{ $t(`insights.regions.${region}`) }}
          </button>
        </div>
        <p v-if="inputError || invalidURL" class="compare-builder__error">{{ $t('insights.compare.invalid') }}</p>
      </section>

      <p v-if="error" class="insights-empty-state">{{ $t('insights.emptyStates.unavailable') }}</p>
      <p v-else-if="compare?.status === 'insufficient_data'" class="insights-empty-state">{{ $t('insights.compare.insufficientData') }}</p>

      <section v-else-if="compare?.status === 'ready'" class="compare-result" data-compare-result>
        <div class="compare-horizons">
          <span>{{ $t('insights.gameCompare.stateSnapshot', { date: compare.state_as_of || '—' }) }}</span>
          <span>{{ $t('insights.gameCompare.playerSnapshot', { date: shortTimestamp(compare.player_snapshot_scheduled_for) }) }}</span>
          <span>{{ $t('insights.gameCompare.playerWindow', { date: compare.player_fact_through || '—' }) }}</span>
        </div>
        <div class="compare-table-wrap">
          <table class="compare-table">
            <thead>
              <tr>
                <th>{{ $t('insights.compare.fact') }}</th>
                <th v-for="item in compare.games" :key="item.game.id" :data-compare-entity-id="item.game.id">
                  <NuxtLink :to="localePath(`/games/${item.game.id}`)">{{ item.game.name || `#${item.game.id}` }}</NuxtLink>
                  <small>#{{ item.game.id }}</small>
                </th>
              </tr>
            </thead>
            <tbody>
              <tr><th>{{ $t('insights.entity.releaseState') }}</th><td v-for="item in compare.games" :key="item.game.id">{{ release(item) }}</td></tr>
              <tr><th>{{ $t('insights.entity.freeState') }}</th><td v-for="item in compare.games" :key="item.game.id">{{ booleanValue(item.state.free, 'free') }}</td></tr>
              <tr v-for="platform in platforms" :key="platform"><th>{{ platformLabel(platform) }}</th><td v-for="item in compare.games" :key="item.game.id">{{ booleanValue(item.state[platform]) }}</td></tr>
              <tr><th>{{ $t('insights.gameCompare.currentPlayers') }}</th><td v-for="item in compare.games" :key="item.game.id" :data-current-player-available="item.players.current_available">{{ currentPlayers(item) }}</td></tr>
              <tr><th>{{ $t('insights.gameCompare.peak30d') }}</th><td v-for="item in compare.games" :key="item.game.id">{{ numberOrDash(item.players.peak_30d) }}</td></tr>
              <tr><th>{{ $t('insights.gameCompare.average30d') }}</th><td v-for="item in compare.games" :key="item.game.id">{{ numberOrDash(item.players.average_30d) }}</td></tr>
              <tr><th>{{ $t('insights.gameCompare.playerQuality') }}</th><td v-for="item in compare.games" :key="item.game.id">{{ playerQuality(item) }}</td></tr>
              <tr><th>{{ $t('insights.gameCompare.currentPrice', { region: selectedRegion }) }}</th><td v-for="item in compare.games" :key="item.game.id" :data-price-state="item.price.state || 'unavailable'">{{ price(item) }}</td></tr>
              <tr><th>{{ $t('insights.entity.observedLow') }}</th><td v-for="item in compare.games" :key="item.game.id">{{ observedLow(item) }}</td></tr>
              <tr><th>{{ $t('insights.gameCompare.languageEvidence') }}</th><td v-for="item in compare.games" :key="item.game.id" :data-language-evidence="item.languages.evidence">{{ $t(`insights.gameCompare.languageStates.${item.languages.evidence}`) }}</td></tr>
              <tr><th>{{ $t('insights.gameCompare.supportedLanguages') }}</th><td v-for="item in compare.games" :key="item.game.id">{{ list(item.languages.supported) }}</td></tr>
              <tr><th>{{ $t('insights.gameCompare.fullAudio') }}</th><td v-for="item in compare.games" :key="item.game.id">{{ list(item.languages.explicit_full_audio) }}</td></tr>
              <tr><th>{{ $t('insights.gameCompare.unmappedLanguages') }}</th><td v-for="item in compare.games" :key="item.game.id">{{ list(item.languages.unknown_names) }}</td></tr>
            </tbody>
          </table>
        </div>
      </section>

      <section class="insights-data-info">
        <h2>{{ $t('insights.compare.aboutTitle') }}</h2>
        <p>{{ $t('insights.gameCompare.about') }}</p>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import EcosystemNavigation from '@/components/insights/EcosystemNavigation.vue'
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { getGameCompare } from '@/services/game'
import type { GameCompareItem, GameInsightRegion } from '@/types/insights'
import { insightCompareReady, parseInsightCompareIDs } from '@/utils/insightCompare'
import { formatMinorAmount } from '@/utils/insightPrices'

const route = useRoute()
const router = useRouter()
const localePath = useLocalePath()
const { locale, t } = useI18n()
const regions = ['CN', 'US', 'HK'] as const satisfies readonly GameInsightRegion[]
const platforms = ['windows', 'mac', 'linux'] as const
const parsedIDs = computed(() => parseInsightCompareIDs(route.query.ids))
const selectedIDs = computed(() => parsedIDs.value ?? [])
const invalidURL = computed(() => parsedIDs.value === null)
const ready = computed(() => insightCompareReady(selectedIDs.value))
const selectedRegion = computed<GameInsightRegion>(() => regions.includes(route.query.region as GameInsightRegion) ? route.query.region as GameInsightRegion : 'CN')
const requestKey = computed(() => `${selectedIDs.value.join(',')}:${selectedRegion.value}`)
const input = ref(typeof route.query.ids === 'string' ? route.query.ids : '')
const inputError = ref(false)

watch(() => route.query.ids, value => { input.value = typeof value === 'string' ? value : '' })

const { data: compare, error } = await useAsyncData('game-compare', async () => {
  if (!ready.value) return null
  return getGameCompare(selectedIDs.value, selectedRegion.value)
}, { watch: [requestKey] })

const builderHint = computed(() => selectedIDs.value.length === 0
  ? t('insights.compare.emptyHint')
  : selectedIDs.value.length === 1
    ? t('insights.compare.oneHint')
    : t('insights.compare.readyHint', { count: selectedIDs.value.length }))

function applySelection() {
  const ids = parseInsightCompareIDs(input.value)
  inputError.value = ids === null
  if (ids === null) return
  void router.push({ path: route.path, query: { ...(ids.length ? { ids: ids.join(',') } : {}), region: selectedRegion.value } })
}

function selectRegion(region: GameInsightRegion) {
  void router.push({ path: route.path, query: { ...(selectedIDs.value.length ? { ids: selectedIDs.value.join(',') } : {}), region } })
}

function booleanValue(value: boolean | null, kind: 'support' | 'free' = 'support') {
  if (value === null) return t('insights.entity.booleanUnknown')
  if (kind === 'free') return value ? t('insights.entity.freeGame') : t('insights.entity.paidGame')
  return value ? t('insights.entity.supported') : t('insights.entity.unsupported')
}

function platformLabel(platform: typeof platforms[number]) {
  return platform === 'windows' ? 'Windows' : platform === 'mac' ? 'macOS' : 'Linux'
}

function release(item: GameCompareItem) {
  return item.state.release ? t(`insights.entity.releaseStates.${item.state.release}`) : t('insights.entity.booleanUnknown')
}

function currentPlayers(item: GameCompareItem) {
  return item.players.current_available && item.players.current !== null ? numberOrDash(item.players.current) : '—'
}

function numberOrDash(value: number | null) {
  return value === null ? '—' : new Intl.NumberFormat(locale.value, { maximumFractionDigits: 1 }).format(value)
}

function percent(value: number | null) {
  return value === null ? '—' : new Intl.NumberFormat(locale.value, { style: 'percent', maximumFractionDigits: 1 }).format(value)
}

function playerQuality(item: GameCompareItem) {
  return t('insights.gameCompare.qualityValue', {
    days: item.players.observed_days_30d,
    samples: item.players.successful_samples_30d,
    coverage: percent(item.players.sample_coverage_30d),
  })
}

function price(item: GameCompareItem) {
  if (!item.price.available || !item.price.state) return t('insights.entity.priceMissing')
  if (item.price.state === 'free') return t('insights.entity.priceFree')
  if (item.price.state === 'priced' && item.price.final_amount !== null && item.price.currency) {
    return formatMinorAmount(item.price.final_amount, item.price.currency, locale.value)
  }
  return t(`insights.entity.priceStates.${item.price.state}`)
}

function observedLow(item: GameCompareItem) {
  const low = item.price.observed_low
  return low ? formatMinorAmount(low.amount, low.currency, locale.value) : '—'
}

function list(values: string[]) { return values.length ? values.join(', ') : '—' }
function shortTimestamp(value: string | null) { return value?.slice(0, 16).replace('T', ' ') || '—' }

useSeoMeta({
  title: () => `${t('insights.gameCompare.title')} | GoFurry`,
  description: () => t('insights.gameCompare.description'),
  robots: 'noindex, follow',
})
</script>
