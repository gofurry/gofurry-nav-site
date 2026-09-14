<template>
  <div class="insights-page insights-overview-page">
    <main class="insights-container">
      <EcosystemNavigation />

      <header class="overview-header" data-overview-header>
        <h1>{{ $t('insights.overview.title') }}</h1>
        <p class="overview-header__intro">{{ $t('insights.editorial.description') }}</p>
        <div class="overview-header__facts">
          <dl class="overview-stats">
            <div v-for="stat in stats" :key="stat.label"><dt>{{ stat.label }}</dt><dd>{{ stat.value === null ? '—' : number(stat.value) }}</dd></div>
          </dl>
          <p class="overview-freshness">{{ $t('insights.editorial.snapshot') }}<time v-if="generatedAt" :datetime="generatedAt">{{ formatOverviewSnapshot(generatedAt, locale) }}</time><span v-else>—</span></p>
        </div>
      </header>

      <InsightsOverviewActivity :items="recentChanges" :unavailable="!data.nav && !data.game" />

      <div class="overview-ecosystems">
        <InsightsOverviewSites :overview="data.nav" />
        <InsightsOverviewGamePulse :panel="data.panel" />
      </div>

      <section class="overview-explore" aria-labelledby="overview-explore-title" data-overview-explore>
        <div class="overview-section-heading"><div><p class="overview-kicker">{{ $t('insights.editorial.exploreKicker') }}</p><h2 id="overview-explore-title">{{ $t('insights.editorial.exploreTitle') }}</h2></div></div>
        <div class="overview-explore__groups">
          <div v-for="(items, domain) in overviewExploreGroups" :key="domain">
            <h3>{{ $t(`insights.editorial.${domain === 'site' ? 'sitesTitle' : 'gamesTitle'}`) }}</h3>
            <NuxtLink v-for="item in items" :key="item.path" :to="localePath(item.path)" class="overview-explore__link">
              <span><strong>{{ $t(`insights.editorial.links.${item.key}.title`) }}</strong><span>{{ $t(`insights.editorial.links.${item.key}.description`) }}</span></span><span aria-hidden="true">↗</span>
            </NuxtLink>
          </div>
        </div>
        <NuxtLink :to="localePath(overviewChangesPath)" class="overview-explore__all"><span>{{ $t('insights.editorial.allChanges') }}</span><span aria-hidden="true">↗</span></NuxtLink>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import EcosystemNavigation from '@/components/insights/EcosystemNavigation.vue'
import InsightsOverviewActivity from '@/components/insights/activity/InsightsOverviewActivity.vue'
import InsightsOverviewSites from '@/components/insights/overview/InsightsOverviewSites.vue'
import InsightsOverviewGamePulse from '@/components/insights/overview/InsightsOverviewGamePulse.vue'
import { getGameInsightsOverview, getGameHomePanel } from '@/services/game'
import { getNavInsightsOverview } from '@/services/nav'
import type { GameV2PanelRecord } from '@/types/game'
import type { InsightOverview } from '@/types/insights'
import { formatOverviewSnapshot, overviewActivity, overviewChangesPath, overviewExploreGroups, overviewGeneratedAt } from '@/utils/insightOverview'
import { buildInsightsSeo } from '@/utils/seo'

interface OverviewSnapshot {
  nav: InsightOverview | null
  game: InsightOverview | null
  panel: GameV2PanelRecord | null
}

const { locale, t } = useI18n()
const localePath = useLocalePath()
const { data } = await useAsyncData<OverviewSnapshot>(() => `insights:overview:${locale.value}`, async () => {
  const [navResult, gameResult, panelResult] = await Promise.allSettled([
    getNavInsightsOverview(),
    getGameInsightsOverview(),
    getGameHomePanel(locale.value),
  ])
  return {
    nav: navResult.status === 'fulfilled' ? navResult.value : null,
    game: gameResult.status === 'fulfilled' ? gameResult.value : null,
    panel: panelResult.status === 'fulfilled' ? panelResult.value : null,
  }
}, {
  default: () => ({ nav: null, game: null, panel: null }),
})

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
const number = (value: number) => new Intl.NumberFormat(locale.value === 'en' ? 'en-US' : 'zh-CN').format(value)
const generatedAt = computed(() => overviewGeneratedAt(data.value.nav, data.value.game))
const recentChanges = computed(() => overviewActivity(data.value.nav, data.value.game))
const seo = computed(() => buildInsightsSeo('overview', locale.value))

useSeoMeta({
  title: () => seo.value.title,
  description: () => seo.value.description,
  ogTitle: () => seo.value.title,
  ogDescription: () => seo.value.description,
})
</script>
