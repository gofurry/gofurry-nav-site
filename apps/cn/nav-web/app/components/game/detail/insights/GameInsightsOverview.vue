<template>
  <article class="game-insights-overview" data-game-summary>
    <section class="game-insights-overview__section game-insights-overview__players">
      <h3>{{ $t('insights.entity.playerObservation') }}</h3>
      <div class="game-insights-overview__primary-stat">
        <span>{{ $t('insights.entity.latestPlayerObservation') }}</span>
        <strong data-current-players>{{ formatPlayerCount(summary.players.current) }}</strong>
      </div>
      <dl class="game-insights-overview__metrics">
        <div>
          <dt>{{ $t('insights.entity.peak30d') }}</dt>
          <dd>{{ formatPlayerCount(summary.players.peak_30d) }}</dd>
        </div>
        <div>
          <dt>{{ $t('insights.entity.average30d') }}</dt>
          <dd>{{ formatPlayerCount(summary.players.average_30d) }}</dd>
        </div>
      </dl>
      <div class="game-insights-overview__quality">
        <span>{{ $t('insights.entity.observedDaysValue', { count: summary.players.observed_days_30d }) }}</span>
        <span>{{ $t('insights.entity.sampleCoverageValue', { coverage: formatCoverage(summary.players.sample_coverage_30d) }) }}</span>
      </div>
      <small>{{ formatObservationTime(summary.players.as_of) }}</small>
    </section>

    <section class="game-insights-overview__section game-insights-overview__prices">
      <h3>{{ $t('insights.entity.regionalPrices') }}</h3>
      <ul class="game-insights-price-list">
        <li
          v-for="row in regionalPriceRows"
          :key="row.region"
          :data-price-kind="row.kind"
          :data-price-region-summary="row.region"
        >
          <div class="game-insights-price-list__heading">
            <span>{{ $t(`insights.regions.${row.region}`) }}</span>
            <strong>{{ row.current }}</strong>
            <span v-if="row.discount" class="game-insights-price-list__discount">{{ row.discount }}</span>
          </div>
          <p v-if="row.original">{{ $t('insights.entity.originalPrice', { price: row.original }) }}</p>
          <p v-if="row.observedLow" class="game-insights-price-list__low" data-observed-low>
            {{ row.observedLow }}
          </p>
        </li>
      </ul>
    </section>

    <section class="game-insights-overview__section game-insights-overview__state">
      <h3>{{ $t('insights.entity.gameState') }}</h3>
      <ul class="game-insights-platform-list">
        <li v-for="platform in platforms" :key="platform.name">
          <span>{{ platform.name }}</span>
          <strong :class="`game-insights-platform-list__state--${platform.state}`">
            {{ platform.label }}
          </strong>
        </li>
      </ul>
      <p class="game-insights-state-summary">
        {{ releaseStateText }} <span aria-hidden="true">·</span> {{ freeStateText }}
      </p>
    </section>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { GameInsightRegion, GameInsightRegionalPrice, GameInsights } from '@/types/insights'
import { formatMinorAmount, publicPriceDisplay } from '@/utils/insightPrices'

type PlatformState = 'supported' | 'unsupported' | 'unknown'

const props = defineProps<{
  summary: GameInsights
}>()

const { locale, t } = useI18n()
const regions: GameInsightRegion[] = ['CN', 'US', 'HK']

const regionalPriceRows = computed(() => regions.map((region) => {
  const price = props.summary.regional_prices.regions.find(item => item.region === region) ?? null
  return {
    region,
    kind: priceKind(price),
    current: currentPriceText(price),
    original: originalPriceText(price),
    discount: discountText(price),
    observedLow: observedLowText(price),
  }
}))

const platforms = computed(() => [
  platformItem('Windows', props.summary.state.windows),
  platformItem('macOS', props.summary.state.mac),
  platformItem('Linux', props.summary.state.linux),
])

const freeStateText = computed(() => {
  if (props.summary.state.free === null) return t('insights.entity.booleanUnknown')
  return props.summary.state.free ? t('insights.entity.freeGame') : t('insights.entity.paidGame')
})

const releaseStateText = computed(() => {
  const state = props.summary.state.release
  if (state === 'available' || state === 'upcoming' || state === 'withdrawn') {
    return t(`insights.entity.releaseStates.${state}`)
  }
  return t('insights.entity.booleanUnknown')
})

function formatPlayerCount(value: number | null) {
  if (value === null) return t('insights.entity.dataUnavailable')
  return t('insights.entity.playerCount', { count: new Intl.NumberFormat(locale.value).format(value) })
}

function formatCoverage(value: number | null) {
  if (value === null) return t('insights.entity.dataUnavailable')
  return new Intl.NumberFormat(locale.value, { style: 'percent', maximumFractionDigits: 0 }).format(value)
}

function formatObservationTime(value: string | null) {
  if (!value) return t('insights.entity.latestObservationUnavailable')
  return t('insights.entity.latestObservationAt', { date: value.slice(0, 10) })
}

function priceKind(price: GameInsightRegionalPrice | null) {
  if (!price?.available) return 'missing'
  return price.state ?? 'missing'
}

function currentPriceText(price: GameInsightRegionalPrice | null) {
  if (!price?.available) return t('insights.entity.priceMissingShort')
  const display = publicPriceDisplay(price)
  if (display.kind === 'free') return t('insights.entity.priceFree')
  if (display.kind === 'priced') return formatAmount(display.amount, price.currency)
  if (price.state === 'unknown') return t('insights.entity.priceStatusUnknown')
  if (price.state === 'unpriced') return t('insights.entity.priceMissingShort')
  return t('insights.entity.priceUnavailable')
}

function originalPriceText(price: GameInsightRegionalPrice | null) {
  if (!price?.available || price.state !== 'priced' || price.initial_amount === null || price.final_amount === null) return null
  if (price.initial_amount === price.final_amount) return null
  return formatAmount(price.initial_amount, price.currency)
}

function discountText(price: GameInsightRegionalPrice | null) {
  if (!price?.available || price.discount_percent === null || price.discount_percent <= 0) return null
  return `-${price.discount_percent}%`
}

function observedLowText(price: GameInsightRegionalPrice | null) {
  if (!price?.observed_low) return null
  return t('insights.entity.observedLowValue', {
    price: formatMinorAmount(price.observed_low.amount, price.observed_low.currency, locale.value),
    date: price.observed_low.observed_since,
  })
}

function formatAmount(amount: number, currency: string | null) {
  return currency ? formatMinorAmount(amount, currency, locale.value) : t('insights.entity.priceStatusUnknown')
}

function platformItem(name: string, value: boolean | null) {
  const state: PlatformState = value === null ? 'unknown' : value ? 'supported' : 'unsupported'
  return {
    name,
    state,
    label: t(`insights.entity.platformStates.${state}`),
  }
}
</script>
