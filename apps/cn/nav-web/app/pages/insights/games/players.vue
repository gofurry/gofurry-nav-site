<template>
  <div class="insights-page insights-workspace-page" data-player-intelligence>
    <main class="insights-container">
      <EcosystemNavigation context="game" />
      <InsightsWorkspaceHeader :eyebrow="$t('insights.games.title')" :title="$t('insights.playerIntelligence.title')" :description="$t('insights.playerIntelligence.description')" />
      <InsightWorkspaceSelector :items="metrics.map(value => ({ value, label: $t('insights.playerIntelligence.metrics.' + value) }))" :selected="selectedMetric" :label="$t('insights.workspace.rankingMetric')" @select="selectMetric" />
      <p v-if="error" class="insights-empty-state">{{ $t('insights.emptyStates.unavailable') }}</p>
      <template v-else-if="ranking">
        <section class="insights-workspace-summary" :aria-label="$t('insights.workspace.observationSummary')">
          <dl class="insights-workspace-stats">
            <div><dt>{{ $t('insights.playerIntelligence.population') }}</dt><dd>{{ ranking.population }}</dd></div>
            <div><dt>{{ $t('insights.playerIntelligence.ranked') }}</dt><dd>{{ ranking.ranked }}</dd></div>
            <div><dt>{{ $t('insights.playerIntelligence.coverage') }}</dt><dd>{{ percent(ranking.entity_coverage) }}</dd></div>
          </dl>
          <p class="insights-workspace-meta">{{ horizonText }}</p>
        </section>
        <section class="insights-workspace-section" :aria-busy="pending">
          <h2>{{ $t('insights.workspace.ranking') }}</h2>
          <InsightRankingList :items="ranking.items" :metric="ranking.metric" />
        </section>
      </template>
      <InsightWorkspaceDisclosure :title="$t('insights.playerIntelligence.aboutTitle')">
        <p>{{ $t('insights.playerIntelligence.about') }}</p>
        <p>{{ $t('insights.workspace.playerBasis') }}</p>
        <p>{{ horizonText }}</p>
        <p v-if="ranking?.observed_from || ranking?.observed_through">{{ $t('insights.workspace.observationWindow', { from: formatWorkspaceTimestamp(ranking.observed_from, locale), through: formatWorkspaceTimestamp(ranking.observed_through, locale) }) }}</p>
      </InsightWorkspaceDisclosure>
    </main>
  </div>
</template>

<script setup lang="ts">
import InsightsWorkspaceHeader from '@/components/insights/workspace/InsightsWorkspaceHeader.vue'
import InsightWorkspaceDisclosure from '@/components/insights/workspace/InsightWorkspaceDisclosure.vue'

import EcosystemNavigation from '@/components/insights/EcosystemNavigation.vue'
import InsightWorkspaceSelector from '@/components/insights/workspace/InsightWorkspaceSelector.vue'
import InsightRankingList from '@/components/insights/workspace/InsightRankingList.vue'
import { formatWorkspaceTimestamp } from '@/utils/insightWorkspace'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { getGamePlayerRanking } from '@/services/game'
import type { GamePlayerRankingMetric } from '@/types/insights'

const route = useRoute()
const router = useRouter()
const { locale, t } = useI18n()
const metrics: GamePlayerRankingMetric[] = ['latest_observed', 'peak_30d', 'average_30d']
const selectedMetric = computed<GamePlayerRankingMetric>(() => metrics.includes(route.query.metric as GamePlayerRankingMetric)
  ? route.query.metric as GamePlayerRankingMetric
  : 'latest_observed')
const { data: ranking, error, pending } = await useAsyncData('game-player-ranking', () => getGamePlayerRanking(selectedMetric.value), { watch: [selectedMetric] })

function selectMetric(metric: GamePlayerRankingMetric) {
  void router.push({ path: route.path, query: { metric } })
}

function percent(value: number | null) {
  return value === null ? '—' : new Intl.NumberFormat(locale.value, { style: 'percent', maximumFractionDigits: 1 }).format(value)
}

const horizonText = computed(() => selectedMetric.value === 'latest_observed'
  ? t('insights.playerIntelligence.snapshot', { date: formatWorkspaceTimestamp(ranking.value?.snapshot_scheduled_for ?? null, locale.value) })
  : t('insights.playerIntelligence.window', { from: ranking.value?.window_from ?? '—', through: ranking.value?.window_through ?? '—' }))
</script>
