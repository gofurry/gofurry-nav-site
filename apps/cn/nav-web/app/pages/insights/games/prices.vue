<template>
  <div class="insights-page insights-workspace-page" data-regional-price-intelligence>
    <main class="insights-container">
      <EcosystemNavigation context="game" />
      <InsightsWorkspaceHeader :eyebrow="$t('insights.games.title')" :title="$t('insights.priceIntelligence.title')" :description="$t('insights.priceIntelligence.description')" />
      <InsightWorkspaceSelector :items="regions.map(value => ({ value, label: $t('insights.regions.' + value) }))" :selected="selectedRegion" :label="$t('insights.workspace.region')" @select="selectRegion" />
      <p v-if="overviewError" class="insights-empty-state" data-price-overview-error>{{ $t('insights.emptyStates.unavailable') }}</p>
      <section v-else-if="overview" class="insights-workspace-summary" :aria-busy="overviewPending">
        <h2>{{ $t('insights.workspace.priceSummary') }}</h2>
        <dl class="insights-workspace-stats">
          <div v-for="key in primaryKeys" :key="key"><dt>{{ $t('insights.priceIntelligence.' + key) }}</dt><dd>{{ key === 'coverage' ? percent(overview.coverage) : overview[key] }}</dd></div>
        </dl>
        <dl class="insights-workspace-secondary">
          <div v-for="key in secondaryKeys" :key="key"><dt>{{ $t('insights.priceIntelligence.' + key) }}</dt><dd>{{ overview[key] }}</dd></div>
        </dl>
        <p class="insights-workspace-meta">{{ $t('insights.entity.asOf', { date: overview.as_of ?? '—' }) }}</p>
      </section>
      <section class="insights-workspace-section" :aria-busy="discountPending">
        <h2>{{ $t('insights.priceIntelligence.discounts') }}</h2>
        <p v-if="discountError" class="insights-empty-state" data-discount-error>{{ $t('insights.emptyStates.unavailable') }}</p>
        <p v-else-if="!discounts?.items.length" class="insights-empty-state">{{ $t('insights.workspace.noItems') }}</p>
        <ul v-else class="insight-discount-list">
          <li v-for="item in discounts.items" :key="item.game.id">
            <NuxtLink :to="localePath('/games/' + item.game.id)" class="insight-discount-row" data-workspace-entity>
              <InsightEntityMedia domain="game" :entity="item.game" />
              <strong class="insight-discount-row__name">{{ item.game.name }}</strong>
              <div class="insight-discount-row__price"><span>{{ $t('insights.priceIntelligence.currentPrice') }}</span><strong>{{ money(item.final_amount, item.currency) }}</strong></div>
              <div><span>{{ $t('insights.priceIntelligence.discount') }}</span><strong>{{ item.discount_percent }}%</strong></div>
              <div><span>{{ $t('insights.priceIntelligence.observedLow') }}</span><strong>{{ item.observed_low ? money(item.observed_low.amount, item.observed_low.currency) : '—' }}</strong></div>
            </NuxtLink>
          </li>
        </ul>
        <p v-if="discounts?.as_of" class="insights-workspace-meta">{{ $t('insights.entity.asOf', { date: discounts.as_of }) }}</p>
      </section>
      <InsightWorkspaceDisclosure :title="$t('insights.priceIntelligence.aboutTitle')">
        <p>{{ $t('insights.priceIntelligence.about') }}</p>
        <p>{{ $t('insights.workspace.discountSubset') }}</p>
        <p>{{ $t('insights.workspace.region') }}: {{ $t('insights.regions.' + selectedRegion) }}</p>
      </InsightWorkspaceDisclosure>
    </main>
  </div>
</template>

<script setup lang="ts">
import InsightsWorkspaceHeader from '@/components/insights/workspace/InsightsWorkspaceHeader.vue'
import InsightWorkspaceDisclosure from '@/components/insights/workspace/InsightWorkspaceDisclosure.vue'

import EcosystemNavigation from '@/components/insights/EcosystemNavigation.vue'
import InsightWorkspaceSelector from '@/components/insights/workspace/InsightWorkspaceSelector.vue'
import InsightEntityMedia from '@/components/insights/entity/InsightEntityMedia.vue'
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
const { data: overview, error: overviewError, pending: overviewPending } = await useAsyncData('regional-price-overview', () => getGamePriceOverview(selectedRegion.value), { watch: [selectedRegion] })
const { data: discounts, error: discountError, pending: discountPending } = await useAsyncData('regional-price-discounts', () => getGameDiscounts(selectedRegion.value), { watch: [selectedRegion] })
const primaryKeys = ['population', 'priced', 'free', 'discounted', 'coverage'] as const satisfies readonly (keyof GamePriceOverview)[]

const secondaryKeys = ['unpriced', 'unknown', 'unavailable'] as const

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
