<template>
  <div class="insights-page insights-intelligence-page" data-language-intelligence>
    <GoFurryGridBackground :fixed="false" profile="light" />
    <main class="insights-container">
      <InsightsNav />
      <GameIntelligenceNav />
      <header class="insights-hero insights-hero--domain">
        <p class="insights-eyebrow">LANGUAGE INTELLIGENCE</p>
        <h1>{{ $t('insights.languageIntelligence.title') }}</h1>
        <p>{{ $t('insights.languageIntelligence.description') }}</p>
      </header>
      <p v-if="error" class="insights-empty-state">{{ $t('insights.emptyStates.unavailable') }}</p>
      <template v-else-if="overview">
        <div class="intelligence-stats">
          <article><span>{{ $t('insights.languageIntelligence.coverage') }}</span><strong>{{ percent(overview.coverage) }}</strong></article>
          <article><span>{{ $t('insights.languageIntelligence.fresh') }}</span><strong>{{ overview.fresh }}</strong></article>
          <article><span>{{ $t('insights.languageIntelligence.normalization') }}</span><strong>{{ percent(overview.normalization_coverage) }}</strong></article>
        </div>
        <section class="intelligence-panel">
          <h2>{{ $t('insights.languageIntelligence.distribution') }}</h2>
          <p class="insights-data-note">{{ $t('insights.languageIntelligence.overlap') }}</p>
          <div class="intelligence-table-wrap">
            <table class="intelligence-table">
              <thead><tr><th>{{ $t('insights.languageIntelligence.language') }}</th><th>{{ $t('insights.languageIntelligence.supported') }}</th><th>{{ $t('insights.languageIntelligence.share') }}</th><th>{{ $t('insights.languageIntelligence.fullAudio') }}</th></tr></thead>
              <tbody><tr v-for="item in overview.items" :key="item.code"><td>{{ languageName(item.code, item.steam_name) }}</td><td>{{ item.supported_games }}</td><td>{{ percent(item.share) }}</td><td>{{ item.explicit_full_audio_games }} · {{ percent(item.explicit_full_audio_share) }}</td></tr></tbody>
            </table>
          </div>
        </section>
      </template>
      <section class="insights-data-info"><h2>{{ $t('insights.languageIntelligence.aboutTitle') }}</h2><p>{{ $t('insights.languageIntelligence.about') }}</p></section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import GoFurryGridBackground from '@/components/common/GoFurryGridBackground.vue'
import GameIntelligenceNav from '@/components/insights/GameIntelligenceNav.vue'
import InsightsNav from '@/components/insights/InsightsNav.vue'
import { getGameLanguageOverview } from '@/services/game'

const { locale } = useI18n()
const { data: overview, error } = await useAsyncData('language-intelligence', getGameLanguageOverview)

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
