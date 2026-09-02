<template>
  <div class="insights-page insights-overview-page">
    <GoFurryGridBackground :fixed="false" profile="light" />
    <main class="insights-container">
      <InsightsNav />

      <header class="insights-hero">
        <p class="insights-eyebrow">{{ $t('insights.overview.eyebrow') }}</p>
        <h1>{{ $t('insights.overview.title') }}</h1>
        <p>{{ $t('insights.overview.description') }}</p>
      </header>

      <InsightsStats :items="stats" />

      <section class="insights-previews" aria-label="Insights metrics">
        <article class="insights-preview">
          <div class="insights-preview__heading">
            <h2>{{ $t('insights.overview.sitesPreview') }}</h2>
            <NuxtLink :to="localePath('/insights/sites')">{{ $t('insights.overview.viewSites') }}</NuxtLink>
          </div>
          <p v-if="navUnavailable" class="insights-empty-state">{{ $t('insights.emptyStates.unavailable') }}</p>
          <div v-else class="insights-preview__metrics">
            <div v-for="metric in navPreview" :key="metric.key">
              <span>{{ $t(`insights.metrics.${metric.key}.name`) }}</span>
              <strong>{{ formatPercent(metric.value) }}</strong>
            </div>
          </div>
        </article>

        <article class="insights-preview">
          <div class="insights-preview__heading">
            <h2>{{ $t('insights.overview.gamesPreview') }}</h2>
            <NuxtLink :to="localePath('/insights/games')">{{ $t('insights.overview.viewGames') }}</NuxtLink>
          </div>
          <p v-if="gameUnavailable" class="insights-empty-state">{{ $t('insights.emptyStates.unavailable') }}</p>
          <div v-else class="insights-preview__metrics">
            <div v-for="metric in gamePreview" :key="metric.key">
              <span>{{ $t(`insights.metrics.${metric.key}.name`) }}</span>
              <strong>{{ formatPercent(metric.value) }}</strong>
            </div>
          </div>
        </article>
      </section>

      <InsightsRecentChanges :items="recentChanges" :unavailable="navUnavailable && gameUnavailable" />
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import GoFurryGridBackground from '@/components/common/GoFurryGridBackground.vue'
import InsightsNav from '@/components/insights/InsightsNav.vue'
import InsightsRecentChanges from '@/components/insights/InsightsRecentChanges.vue'
import InsightsStats from '@/components/insights/InsightsStats.vue'
import { getGameInsightsOverview } from '@/services/game'
import { getNavInsightsOverview } from '@/services/nav'
import type { InsightFeedItem, InsightMetric, InsightMetricKey, InsightOverview } from '@/types/insights'
import { buildInsightsSeo } from '@/utils/seo'

interface OverviewSnapshot {
  nav: InsightOverview | null
  game: InsightOverview | null
  navUnavailable: boolean
  gameUnavailable: boolean
}

const { locale, t } = useI18n()
const localePath = useLocalePath()
const { data } = await useAsyncData<OverviewSnapshot>('insights:overview', async () => {
  const [navResult, gameResult] = await Promise.allSettled([
    getNavInsightsOverview(),
    getGameInsightsOverview(),
  ])
  return {
    nav: navResult.status === 'fulfilled' ? navResult.value : null,
    game: gameResult.status === 'fulfilled' ? gameResult.value : null,
    navUnavailable: navResult.status === 'rejected',
    gameUnavailable: gameResult.status === 'rejected',
  }
}, {
  default: () => ({ nav: null, game: null, navUnavailable: true, gameUnavailable: true }),
})

const navUnavailable = computed(() => data.value.navUnavailable)
const gameUnavailable = computed(() => data.value.gameUnavailable)
const stats = computed(() => [
  { label: t('insights.overview.sitesCount'), value: data.value.nav?.entity_count ?? null },
  { label: t('insights.overview.gamesCount'), value: data.value.game?.entity_count ?? null },
  {
    label: t('insights.overview.changesCount'),
    value: data.value.nav && data.value.game
      ? data.value.nav.changes_7d + data.value.game.changes_7d
      : null,
  },
])
const navPreview = computed(() => metricPreview(data.value.nav, ['ipv6', 'tls13', 'security_txt']))
const gamePreview = computed(() => metricPreview(data.value.game, ['free', 'windows', 'linux']))
const recentChanges = computed<InsightFeedItem[]>(() => [
  ...(data.value.nav?.recent_changes ?? []).map(change => ({ ...change, domain: 'site' as const })),
  ...(data.value.game?.recent_changes ?? []).map(change => ({ ...change, domain: 'game' as const })),
].sort((left, right) => eventOrder(right) - eventOrder(left)).slice(0, 8))
const seo = computed(() => buildInsightsSeo('overview', locale.value))

useSeoMeta({
  title: () => seo.value.title,
  description: () => seo.value.description,
  ogTitle: () => seo.value.title,
  ogDescription: () => seo.value.description,
})

function metricPreview(overview: InsightOverview | null, keys: InsightMetricKey[]) {
  const byKey = new Map((overview?.metrics ?? []).map(metric => [metric.key, metric]))
  return keys.map(key => byKey.get(key) ?? { key, value: null }) as Array<Pick<InsightMetric, 'key' | 'value'>>
}

function eventOrder(item: InsightFeedItem) {
  const timestamp = Date.parse(item.occurred_at || `${item.date}T00:00:00Z`)
  return Number.isFinite(timestamp) ? timestamp : 0
}

function formatPercent(value: number | null) {
  return value === null || !Number.isFinite(value) ? '—' : `${(value * 100).toFixed(1)}%`
}
</script>
