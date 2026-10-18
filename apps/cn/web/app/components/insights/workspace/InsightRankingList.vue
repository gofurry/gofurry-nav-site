<template>
  <p v-if="!items.length" class="insights-empty-state">{{ $t('insights.workspace.noItems') }}</p>
  <ol v-else class="insight-ranking-list" :aria-label="$t('insights.workspace.ranking')">
    <li v-for="item in items" :key="item.game.id" :value="item.rank" :data-rank="item.rank">
      <NuxtLink :to="localePath('/games/' + item.game.id)" class="insight-ranking-row" :class="{ 'insight-ranking-row--lead': item.rank <= 3 }" data-workspace-entity>
        <span class="insight-ranking-row__rank" aria-hidden="true">{{ String(item.rank).padStart(2, '0') }}</span>
        <InsightEntityMedia domain="game" :entity="item.game" />
        <strong class="insight-ranking-row__name">{{ item.game.name || '#' + item.game.id }}</strong>
        <div class="insight-ranking-row__value"><strong>{{ number(item.value) }}</strong><span>{{ $t('insights.playerIntelligence.value') }}</span></div>
        <div class="insight-ranking-row__quality">
          <template v-if="metric === 'latest_observed'">
            <span>{{ $t('insights.playerIntelligence.observed') }}</span>
            <time :datetime="item.observed_at || undefined">{{ formatWorkspaceTimestamp(item.observed_at, locale) }}</time>
          </template>
          <template v-else>
            <span>{{ $t('insights.playerIntelligence.observedDays', { count: item.observed_days ?? '—' }) }} · {{ $t('insights.playerIntelligence.successfulSamples', { count: item.successful_samples ?? '—' }) }}</span>
            <span>{{ $t('insights.workspace.sampleCoverage') }} {{ formatInsightRatio(item.sample_coverage) }}</span>
          </template>
        </div>
      </NuxtLink>
    </li>
  </ol>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import InsightEntityMedia from '@/components/insights/entity/InsightEntityMedia.vue'
import type { GamePlayerRankingItem, GamePlayerRankingMetric } from '@/types/insights'
import { formatInsightRatio } from '@/utils/insightDimensions'
import { formatWorkspaceTimestamp } from '@/utils/insightWorkspace'
defineProps<{ items: GamePlayerRankingItem[]; metric: GamePlayerRankingMetric }>()
const { locale } = useI18n()
const localePath = useLocalePath()
const number = (value: number) => new Intl.NumberFormat(locale.value, { maximumFractionDigits: 1 }).format(value)
</script>
