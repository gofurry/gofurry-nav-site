<template>
  <div class="insights-page insights-intelligence-page" data-regional-price-intelligence>
    <main class="insights-container">
      <EcosystemNavigation context="game" />
      <h1 class="sr-only">{{ $t('insights.priceIntelligence.title') }}</h1>
      <div class="insights-ranges intelligence-selector">
        <button v-for="region in regions" :key="region" :class="{ 'insights-ranges__button--active': selectedRegion === region }" @click="selectRegion(region)">
          {{ $t(`insights.regions.${region}`) }}
        </button>
      </div>
      <p v-if="overviewError" class="insights-empty-state">{{ $t('insights.emptyStates.unavailable') }}</p>
      <section v-else-if="overview" class="intelligence-panel">
        <div class="intelligence-stats intelligence-stats--wide">
          <article v-for="key in overviewKeys" :key="key">
            <span>{{ $t(`insights.priceIntelligence.${key}`) }}</span>
            <strong>{{ key === 'coverage' ? percent(overview.coverage) : overview[key] }}</strong>
          </article>
        </div>
        <small>{{ $t('insights.entity.asOf', { date: overview.as_of ?? '—' }) }}</small>
      </section>
      <section class="intelligence-panel">
        <h2>{{ $t('insights.priceIntelligence.discounts') }}</h2>
        <p v-if="discountError" class="insights-empty-state">{{ $t('insights.emptyStates.unavailable') }}</p>
        <div v-else class="intelligence-table-wrap">
          <table class="intelligence-table">
            <thead><tr><th>{{ $t('insights.playerIntelligence.game') }}</th><th>{{ $t('insights.priceIntelligence.currentPrice') }}</th><th>{{ $t('insights.priceIntelligence.discount') }}</th><th>{{ $t('insights.priceIntelligence.observedLow') }}</th></tr></thead>
            <tbody><tr v-for="item in discounts?.items ?? []" :key="item.game.id"><td><NuxtLink :to="localePath(`/games/${item.game.id}`)">{{ item.game.name }}</NuxtLink></td><td>{{ money(item.final_amount, item.currency) }}</td><td>{{ item.discount_percent }}%</td><td>{{ item.observed_low ? money(item.observed_low.amount, item.observed_low.currency) : '—' }}</td></tr></tbody>
          </table>
        </div>
      </section>
      <section class="insights-data-info"><h2>{{ $t('insights.priceIntelligence.aboutTitle') }}</h2><p>{{ $t('insights.priceIntelligence.about') }}</p></section>
    </main>
  </div>
</template>

<script setup lang="ts">
import EcosystemNavigation from '@/components/insights/EcosystemNavigation.vue'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { getGameDiscounts, getGamePriceOverview } from '@/services/game'
import type { GameInsightRegion, GamePriceOverview } from '@/types/insights'
import { formatMinorAmount } from '@/utils/insightPrices'

const route = useRoute()
const router = useRouter()
const localePath = useLocalePath()
const { locale } = useI18n()
const regions: GameInsightRegion[] = ['CN', 'US', 'HK']
const selectedRegion = computed<GameInsightRegion>(() => regions.includes(route.query.region as GameInsightRegion)
  ? route.query.region as GameInsightRegion
  : 'CN')
const { data: overview, error: overviewError } = await useAsyncData('regional-price-overview', () => getGamePriceOverview(selectedRegion.value), { watch: [selectedRegion] })
const { data: discounts, error: discountError } = await useAsyncData('regional-price-discounts', () => getGameDiscounts(selectedRegion.value), { watch: [selectedRegion] })
const overviewKeys = ['population', 'priced', 'free', 'unpriced', 'unknown', 'unavailable', 'coverage', 'discounted'] as const satisfies readonly (keyof GamePriceOverview)[]

function selectRegion(region: GameInsightRegion) {
  void router.push({ path: route.path, query: { region } })
}

function percent(value: number | null) {
  return value === null ? '—' : new Intl.NumberFormat(locale.value, { style: 'percent', maximumFractionDigits: 1 }).format(value)
}

function money(value: number, currency: string) {
  return formatMinorAmount(value, currency, locale.value)
}
</script>
