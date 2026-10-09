<template>
  <div class="insights-page insights-overview-page">
    <main class="insights-container">
      <EcosystemNavigation />

      <header class="overview-header mb-6 min-[1200px]:mb-7" data-overview-header>
        <p class="overview-kicker">GOFURRY / INSIGHTS</p>
        <h1>{{ $t('insights.overviewHero.title') }}</h1>
        <p class="overview-header__intro max-w-[680px]">{{ $t('insights.overviewHero.description') }}</p>
      </header>

      <InsightsOverviewHero :nav="data.nav" :game="data.game" />

      <InsightsOverviewEcosystems :nav="data.nav" :game="data.game" />

      <InsightsOverviewActivity :items="recentChanges" :summary="activitySummary" />

      <InsightsOverviewDirectory />
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import EcosystemNavigation from '@/components/insights/EcosystemNavigation.vue'
import InsightsOverviewActivity from '@/components/insights/activity/InsightsOverviewActivity.vue'
import InsightsOverviewEcosystems from '@/components/insights/overview/InsightsOverviewEcosystems.vue'
import InsightsOverviewHero from '@/components/insights/overview/InsightsOverviewHero.vue'
import InsightsOverviewDirectory from '@/components/insights/overview/InsightsOverviewDirectory.vue'
import { getGameInsightsOverview } from '@/services/game'
import { getNavInsightsOverview } from '@/services/nav'
import type { InsightOverview } from '@/types/insights'
import { overviewActivity } from '@/utils/insightOverview'
import { overviewActivitySummary } from '@/utils/insightOverviewPresentation'
import { buildInsightsSeo } from '@/utils/seo'

interface OverviewSnapshot {
  nav: InsightOverview | null
  game: InsightOverview | null
}

const { locale } = useI18n()
const { data } = await useAsyncData<OverviewSnapshot>(() => `insights:overview:${locale.value}`, async () => {
  const [navResult, gameResult] = await Promise.allSettled([
    getNavInsightsOverview(),
    getGameInsightsOverview(),
  ])
  return {
    nav: navResult.status === 'fulfilled' ? navResult.value : null,
    game: gameResult.status === 'fulfilled' ? gameResult.value : null,
  }
}, {
  default: () => ({ nav: null, game: null }),
})

const activitySummary = computed(() => overviewActivitySummary(data.value.nav, data.value.game))
const recentChanges = computed(() => overviewActivity(data.value.nav, data.value.game))
const seo = computed(() => buildInsightsSeo('overview', locale.value))

useSeoMeta({
  title: () => seo.value.title,
  description: () => seo.value.description,
  ogTitle: () => seo.value.title,
  ogDescription: () => seo.value.description,
})
</script>
