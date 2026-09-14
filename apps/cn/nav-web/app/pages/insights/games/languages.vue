<template>
  <div class="insights-page insights-workspace-page" data-language-intelligence>
    <main class="insights-container">
      <EcosystemNavigation context="game" />
      <InsightsWorkspaceHeader :eyebrow="$t('insights.games.title')" :title="$t('insights.languageIntelligence.title')" :description="$t('insights.languageIntelligence.description')" />
      <p v-if="error" class="insights-empty-state">{{ $t('insights.emptyStates.unavailable') }}</p>
      <template v-else-if="overview">
        <section class="insights-workspace-summary">
          <h2>{{ $t('insights.workspace.observationSummary') }}</h2>
          <dl class="insights-workspace-stats">
            <div><dt>{{ $t('insights.languageIntelligence.coverage') }}</dt><dd>{{ percent(overview.coverage) }}</dd></div>
            <div><dt>{{ $t('insights.languageIntelligence.fresh') }}</dt><dd>{{ overview.fresh }}</dd></div>
            <div><dt>{{ $t('insights.languageIntelligence.normalization') }}</dt><dd>{{ percent(overview.normalization_coverage) }}</dd></div>
          </dl>
          <dl class="insights-workspace-secondary">
            <div v-for="key in qualityKeys" :key="key"><dt>{{ $t('insights.workspace.' + key) }}</dt><dd>{{ overview[key] }}</dd></div>
          </dl>
          <p class="insights-workspace-meta">{{ $t('insights.entity.asOf', { date: overview.as_of ?? '—' }) }}</p>
        </section>
        <section class="insights-workspace-section">
          <h2>{{ $t('insights.languageIntelligence.distribution') }}</h2>
          <p class="insights-workspace-note" data-language-overlap>{{ $t('insights.languageIntelligence.overlap') }}</p>
          <p v-if="!overview.items.length" class="insights-empty-state">{{ $t('insights.workspace.noItems') }}</p>
          <ol v-else class="insight-language-distribution">
            <li v-for="item in overview.items.slice(0, 12)" :key="item.code" :data-language="item.code">
              <div><strong>{{ languageName(item.code, item.steam_name) }}</strong><span>{{ $t('insights.languageIntelligence.supported') }} {{ item.supported_games }} <b>{{ percent(item.share) }}</b></span></div>
              <progress v-if="item.share !== null" :value="overviewSignal(item.share) ?? 0" max="1" :aria-label="languageName(item.code, item.steam_name)" />
              <span v-else class="insights-workspace-note">—</span>
              <p>{{ $t('insights.languageIntelligence.fullAudio') }} {{ item.explicit_full_audio_games }} · {{ percent(item.explicit_full_audio_share) }}</p>
            </li>
          </ol>
          <details class="insights-workspace-raw" data-workspace-raw>
            <summary>{{ $t('insights.workspace.rawData') }}</summary>
            <div class="insights-workspace-table-scroll">
              <table>
                <caption class="sr-only">{{ $t('insights.languageIntelligence.distribution') }}</caption>
                <thead><tr><th scope="col">{{ $t('insights.languageIntelligence.language') }}</th><th scope="col">{{ $t('insights.languageIntelligence.supported') }}</th><th scope="col">{{ $t('insights.languageIntelligence.share') }}</th><th scope="col">{{ $t('insights.languageIntelligence.fullAudio') }}</th></tr></thead>
                <tbody><tr v-for="item in overview.items" :key="item.code"><th scope="row">{{ languageName(item.code, item.steam_name) }}</th><td>{{ item.supported_games }}</td><td>{{ percent(item.share) }}</td><td>{{ item.explicit_full_audio_games }} · {{ percent(item.explicit_full_audio_share) }}</td></tr></tbody>
              </table>
            </div>
          </details>
        </section>
      </template>
      <InsightWorkspaceDisclosure :title="$t('insights.languageIntelligence.aboutTitle')">
        <p>{{ $t('insights.languageIntelligence.about') }}</p>
        <p>{{ $t('insights.languageIntelligence.overlap') }}</p>
        <p>{{ $t('insights.workspace.normalizationBasis') }}</p>
      </InsightWorkspaceDisclosure>
    </main>
  </div>
</template>

<script setup lang="ts">
import InsightsWorkspaceHeader from '@/components/insights/workspace/InsightsWorkspaceHeader.vue'
import InsightWorkspaceDisclosure from '@/components/insights/workspace/InsightWorkspaceDisclosure.vue'

import EcosystemNavigation from '@/components/insights/EcosystemNavigation.vue'
import { overviewSignal } from '@/utils/insightOverview'
import { useI18n } from 'vue-i18n'
import { getGameLanguageOverview } from '@/services/game'

const { locale } = useI18n()
const { data: overview, error } = await useAsyncData('language-intelligence', getGameLanguageOverview)

const qualityKeys = ['stale', 'unobserved', 'unmapped_games', 'unmapped_entries'] as const

function percent(value: number | null) {
  return value === null ? '—' : new Intl.NumberFormat(locale.value, { style: 'percent', maximumFractionDigits: 1 }).format(value)
}

function languageName(code: string, fallback: string) {
  try {
    return new Intl.DisplayNames([locale.value], { type: 'language' }).of(code) || fallback || code
  } catch {
    return fallback || code
  }
}
</script>
