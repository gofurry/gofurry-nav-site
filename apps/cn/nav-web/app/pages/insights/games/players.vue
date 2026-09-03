<template>
  <div class="insights-page insights-intelligence-page" data-player-intelligence>
    <main class="insights-container">
      <EcosystemNavigation context="game" />
      <h1 class="sr-only">{{ $t('insights.playerIntelligence.title') }}</h1>
      <div class="insights-ranges intelligence-selector">
        <button v-for="metric in metrics" :key="metric" :class="{ 'insights-ranges__button--active': selectedMetric === metric }" @click="selectMetric(metric)">
          {{ $t(`insights.playerIntelligence.metrics.${metric}`) }}
        </button>
      </div>
      <p v-if="error" class="insights-empty-state">{{ $t('insights.emptyStates.unavailable') }}</p>
      <section v-else-if="ranking" class="intelligence-panel">
        <div class="intelligence-stats">
          <article><span>{{ $t('insights.playerIntelligence.population') }}</span><strong>{{ ranking.population }}</strong></article>
          <article><span>{{ $t('insights.playerIntelligence.ranked') }}</span><strong>{{ ranking.ranked }}</strong></article>
          <article><span>{{ $t('insights.playerIntelligence.coverage') }}</span><strong>{{ percent(ranking.entity_coverage) }}</strong></article>
        </div>
        <p class="insights-data-note">{{ horizonText }}</p>
        <div class="intelligence-table-wrap">
          <table class="intelligence-table"><thead><tr><th>#</th><th>{{ $t('insights.playerIntelligence.game') }}</th><th>{{ $t('insights.playerIntelligence.value') }}</th><th>{{ $t('insights.playerIntelligence.quality') }}</th></tr></thead>
            <tbody><tr v-for="item in ranking.items" :key="item.game.id"><td>{{ item.rank }}</td><td><NuxtLink :to="localePath(`/games/${item.game.id}`)">{{ item.game.name || `#${item.game.id}` }}</NuxtLink></td><td>{{ number(item.value) }}</td><td>{{ quality(item) }}</td></tr></tbody>
          </table>
        </div>
      </section>
      <section class="insights-data-info"><h2>{{ $t('insights.playerIntelligence.aboutTitle') }}</h2><p>{{ $t('insights.playerIntelligence.about') }}</p></section>
    </main>
  </div>
</template>

<script setup lang="ts">
import EcosystemNavigation from '@/components/insights/EcosystemNavigation.vue'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { getGamePlayerRanking } from '@/services/game'
import type { GamePlayerRankingItem, GamePlayerRankingMetric } from '@/types/insights'

const route = useRoute()
const router = useRouter()
const localePath = useLocalePath()
const { locale, t } = useI18n()
const metrics: GamePlayerRankingMetric[] = ['latest_observed', 'peak_30d', 'average_30d']
const selectedMetric = computed<GamePlayerRankingMetric>(() => metrics.includes(route.query.metric as GamePlayerRankingMetric)
  ? route.query.metric as GamePlayerRankingMetric
  : 'latest_observed')
const { data: ranking, error } = await useAsyncData('game-player-ranking', () => getGamePlayerRanking(selectedMetric.value), { watch: [selectedMetric] })

function selectMetric(metric: GamePlayerRankingMetric) {
  void router.push({ path: route.path, query: { metric } })
}

function number(value: number) {
  return new Intl.NumberFormat(locale.value, { maximumFractionDigits: 1 }).format(value)
}

function percent(value: number | null) {
  return value === null ? '—' : new Intl.NumberFormat(locale.value, { style: 'percent', maximumFractionDigits: 1 }).format(value)
}

function quality(item: GamePlayerRankingItem) {
  if (selectedMetric.value === 'latest_observed') return item.observed_at ? t('insights.playerIntelligence.observed') : '—'
  const parts = [
    t('insights.playerIntelligence.observedDays', { count: item.observed_days ?? 0 }),
    t('insights.playerIntelligence.successfulSamples', { count: item.successful_samples ?? 0 }),
  ]
  if (item.sample_coverage !== null) parts.push(percent(item.sample_coverage))
  return parts.join(' · ')
}

const horizonText = computed(() => selectedMetric.value === 'latest_observed'
  ? t('insights.playerIntelligence.snapshot', { date: ranking.value?.snapshot_scheduled_for?.slice(0, 16) ?? '—' })
  : t('insights.playerIntelligence.window', { from: ranking.value?.window_from ?? '—', through: ranking.value?.window_through ?? '—' }))
</script>
