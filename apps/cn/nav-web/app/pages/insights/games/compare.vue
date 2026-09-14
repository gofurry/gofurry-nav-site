<template>
  <div class="insights-page insights-workspace-page insights-compare-page" data-game-compare :data-compare-count="selectedIDs.length" :data-compare-region="selectedRegion"
    :data-compare-status="invalidURL ? 'invalid' : error ? 'error' : compare?.status || (ready ? 'loading' : 'builder')">
    <main class="insights-container">
      <EcosystemNavigation context="game" />
      <InsightsWorkspaceHeader :eyebrow="$t('insights.games.title')" :title="$t('insights.gameCompare.title')" :description="$t('insights.gameCompare.description')" />
      <div v-if="invalidURL" class="insight-compare-invalid" role="alert">
        <p>{{ $t('insights.compare.invalid') }}</p><button type="button" @click="updateSelection([])">{{ $t('insights.comparePicker.reset') }}</button>
      </div>
      <InsightComparePicker domain="game" :selected-ids="selectedIDs" :entities="entities" @change="updateSelection" />
      <InsightWorkspaceSelector :items="regions.map(value => ({ value, label: $t('insights.regions.' + value) }))" :selected="selectedRegion" :label="$t('insights.gameCompare.region')" @select="selectRegion" />
      <section class="insight-compare-result" :aria-busy="pending" aria-live="polite">
        <p v-if="invalidURL" class="insights-workspace-note">{{ $t('insights.comparePicker.restart') }}</p>
        <p v-else-if="error" class="insights-empty-state">{{ $t('insights.emptyStates.unavailable') }}</p>
        <p v-else-if="pending && ready" class="insights-empty-state">{{ $t('insights.comparePicker.loadingComparison') }}</p>
        <p v-else-if="compare?.status === 'insufficient_data'" class="insights-empty-state">{{ $t('insights.compare.insufficientData') }}</p>
        <template v-else-if="compare?.status === 'ready'">
          <h2>{{ $t('insights.comparePicker.resultsTitle') }}</h2>
          <div class="insight-compare-horizons">
            <span>{{ $t('insights.gameCompare.stateSnapshot', { date: compare.state_as_of || '—' }) }}</span>
            <span>{{ $t('insights.gameCompare.playerSnapshot', { date: shortTimestamp(compare.player_snapshot_scheduled_for) }) }}</span>
            <span>{{ $t('insights.gameCompare.playerWindow', { date: compare.player_fact_through || '—' }) }}</span>
            <span>{{ $t('insights.gameCompare.region') }}: {{ selectedRegion }}</span>
          </div>
          <InsightCompareMatrix domain="game" :entities="entities" :groups="groups" :label="$t('insights.gameCompare.title')" data-compare-result />
        </template>
      </section>
      <InsightWorkspaceDisclosure :title="$t('insights.compare.aboutTitle')">
        <p>{{ $t('insights.gameCompare.about') }}</p>
        <p>{{ $t('insights.priceIntelligence.about') }}</p>
        <p>{{ $t('insights.playerIntelligence.about') }}</p>
        <p>{{ $t('insights.languageIntelligence.about') }}</p>
      </InsightWorkspaceDisclosure>
    </main>
  </div>
</template>

<script setup lang="ts">
import InsightsWorkspaceHeader from '@/components/insights/workspace/InsightsWorkspaceHeader.vue'
import InsightWorkspaceDisclosure from '@/components/insights/workspace/InsightWorkspaceDisclosure.vue'
import InsightWorkspaceSelector from '@/components/insights/workspace/InsightWorkspaceSelector.vue'
import InsightComparePicker from '@/components/insights/compare/InsightComparePicker.vue'
import InsightCompareMatrix from '@/components/insights/compare/InsightCompareMatrix.vue'
import type { CompareMatrixGroup } from '@/types/insightCompare'
import EcosystemNavigation from '@/components/insights/EcosystemNavigation.vue'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { getGameCompare } from '@/services/game'
import type { GameCompareItem, GameInsightRegion } from '@/types/insights'
import { insightCompareReady, parseInsightCompareIDs } from '@/utils/insightCompare'
import { formatMinorAmount } from '@/utils/insightPrices'

const route = useRoute()
const router = useRouter()
const { locale, t } = useI18n()
const regions = ['CN', 'US', 'HK'] as const satisfies readonly GameInsightRegion[]
const platforms = ['windows', 'mac', 'linux'] as const
const parsedIDs = computed(() => parseInsightCompareIDs(route.query.ids))
const selectedIDs = computed(() => parsedIDs.value ?? [])
const invalidURL = computed(() => parsedIDs.value === null)
const ready = computed(() => insightCompareReady(selectedIDs.value))
const selectedRegion = computed<GameInsightRegion>(() => regions.includes(route.query.region as GameInsightRegion) ? route.query.region as GameInsightRegion : 'CN')
const requestKey = computed(() => `${selectedIDs.value.join(',')}:${selectedRegion.value}`)
const { data: snapshot, error, pending } = await useAsyncData('game-compare', async () => {
  const selection = requestKey.value
  const result = ready.value ? await getGameCompare(selectedIDs.value, selectedRegion.value) : null
  return { selection, result }
}, { watch: [requestKey] })
const compare = computed(() => snapshot.value?.selection === requestKey.value ? snapshot.value.result : null)
const orderedItems = computed(() => selectedIDs.value.flatMap(id => {
  const item = compare.value?.games.find(item => item.game.id === id)
  return item ? [item] : []
}))
const entities = computed(() => orderedItems.value.map(item => item.game))

function updateSelection(ids: number[]) {
  if (parseInsightCompareIDs(ids.join(',')) === null) return
  void router.push({ path: route.path, query: { ...route.query, ids: ids.length ? ids.join(',') : undefined, region: selectedRegion.value } })
}

function selectRegion(region: GameInsightRegion) {
  void router.push({ path: route.path, query: { ...route.query, region } })
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

const groups = computed<CompareMatrixGroup[]>(() => {
  const row = (key: string, label: string, value: (item: GameCompareItem) => string, attributes?: (item: GameCompareItem) => Record<string, string | boolean>) => ({
    key, label, cells: orderedItems.value.map(item => ({ text: value(item), attributes: attributes?.(item) })),
  })
  return [
    { key: 'basic', label: t('insights.comparePicker.groups.basic'), rows: [
      row('release', t('insights.entity.releaseState'), release),
      row('free', t('insights.entity.freeState'), item => booleanValue(item.state.free, 'free')),
    ] },
    { key: 'platform', label: t('insights.comparePicker.groups.platform'), rows: platforms.map(key => row(key, platformLabel(key), item => booleanValue(item.state[key]))) },
    { key: 'activity', label: t('insights.comparePicker.groups.activity'), rows: [
      row('current', t('insights.gameCompare.currentPlayers'), currentPlayers, item => ({ 'data-current-player-available': item.players.current_available })),
      row('peak', t('insights.gameCompare.peak30d'), item => numberOrDash(item.players.peak_30d)),
      row('average', t('insights.gameCompare.average30d'), item => numberOrDash(item.players.average_30d)),
      row('quality', t('insights.gameCompare.playerQuality'), playerQuality),
    ] },
    { key: 'price', label: t('insights.comparePicker.groups.price'), rows: [
      row('currentPrice', t('insights.gameCompare.currentPrice', { region: selectedRegion.value }), price, item => ({ 'data-price-state': item.price.state || 'unavailable' })),
      row('observedLow', t('insights.entity.observedLow'), observedLow),
    ] },
    { key: 'language', label: t('insights.comparePicker.groups.language'), rows: [
      row('evidence', t('insights.gameCompare.languageEvidence'), item => t('insights.gameCompare.languageStates.' + item.languages.evidence), item => ({ 'data-language-evidence': item.languages.evidence })),
      row('supported', t('insights.gameCompare.supportedLanguages'), item => list(item.languages.supported)),
      row('audio', t('insights.gameCompare.fullAudio'), item => list(item.languages.explicit_full_audio)),
      row('unmapped', t('insights.gameCompare.unmappedLanguages'), item => list(item.languages.unknown_names)),
    ] },
  ]
})

useSeoMeta({
  title: () => `${t('insights.gameCompare.title')} | GoFurry`,
  description: () => t('insights.gameCompare.description'),
  robots: 'noindex, follow',
})
</script>
