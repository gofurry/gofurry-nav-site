<template>
  <section class="overview-ecosystems mb-[var(--insights-section-gap)]" aria-labelledby="overview-ecosystems-title" data-overview-ecosystems>
    <h2 id="overview-ecosystems-title" class="overview-ecosystems__title mb-6">{{ $t('insights.overviewEcosystems.title') }}</h2>
    <div class="grid grid-cols-1 gap-8 min-[960px]:grid-cols-2 min-[1200px]:gap-12">
      <article v-for="side in sides" :key="side.domain" class="overview-ecosystem min-w-0" :aria-labelledby="`overview-${side.domain}-title`" :data-overview-games="side.domain === 'game' ? '' : undefined" :data-overview-sites="side.domain === 'site' ? '' : undefined">
        <div class="flex flex-wrap items-baseline justify-between gap-3">
          <h3 :id="`overview-${side.domain}-title`">{{ $t(`insights.overviewEcosystems.${side.domain}Title`) }}</h3>
          <NuxtLink :to="localePath(side.path)" class="overview-text-link" data-ecosystem-link>{{ $t('insights.editorial.explore') }} <span aria-hidden="true">↗</span></NuxtLink>
        </div>
        <dl class="overview-ecosystem__count flex flex-wrap items-baseline gap-3">
          <dt>{{ $t(`insights.overview.${side.domain === 'game' ? 'gamesCount' : 'sitesCount'}`) }}</dt>
          <dd :data-ecosystem-count="side.domain">{{ side.count === null ? '—' : number(side.count) }}</dd>
        </dl>
        <p class="overview-ecosystem__intro">{{ $t(`insights.overviewEcosystems.${side.domain}Description`) }}</p>
        <p v-if="!side.overview" class="overview-unavailable" data-ecosystem-unavailable>{{ $t('insights.emptyStates.unavailable') }}</p>
        <p v-else-if="!side.metrics.length" class="overview-unavailable" data-ecosystem-empty>{{ $t('insights.overviewEcosystems.metricsUnavailable') }}</p>
        <template v-else>
          <p class="overview-ecosystem__scope">{{ $t('insights.overviewHero.observedSample') }}</p>
          <ul class="m-0 list-none p-0">
            <li v-for="metric in side.metrics" :key="metric.key" class="overview-ecosystem__metric" :data-ecosystem-metric="metric.key">
              <div class="overview-ecosystem__metric-heading flex items-baseline justify-between gap-4">
                <h4 :id="`overview-${side.domain}-metric-${metric.key}`">{{ $t(`insights.metrics.${metric.key}.name`) }}</h4>
                <strong data-ecosystem-value>{{ formatInsightRatio(metric.value) }}</strong>
              </div>
              <p class="overview-ecosystem__definition">{{ $t(`insights.metrics.${metric.key}.description`) }}</p>
              <progress v-if="side.domain === 'site'" class="overview-ecosystem__track block w-full" :value="metric.value!" max="1" :aria-labelledby="`overview-${side.domain}-metric-${metric.key}`" />
              <dl class="overview-ecosystem__facts flex flex-wrap gap-x-5 gap-y-1">
                <div class="flex flex-wrap gap-x-2"><dt>{{ $t('insights.overviewHero.coverage') }}</dt><dd data-ecosystem-coverage>{{ formatInsightRatio(metric.coverage) }}</dd></div>
                <div class="flex flex-wrap gap-x-2"><dt>{{ $t('insights.overviewHero.factDate') }}</dt><dd><time v-if="metric.as_of" :datetime="metric.as_of">{{ metric.as_of }}</time><span v-else>{{ $t('insights.overviewHero.dateUnavailable') }}</span></dd></div>
              </dl>
            </li>
          </ul>
        </template>
        <div v-if="side.domain === 'game' && previews.length" class="overview-ecosystem__visuals mt-6 grid grid-cols-2 gap-4" data-overview-game-previews>
          <InsightsOverviewGameVisual v-for="item in previews" :key="`${item.game_id}:${item.visual.asset}`" :item="item" @failed="failedAssets.add(item.visual.asset)" />
        </div>
        <div v-if="side.domain === 'site' && siteVisuals.length" class="overview-ecosystem__visuals mt-6" data-overview-site-logos>
          <p class="overview-site-visuals__label">{{ $t('insights.overviewVisuals.recentSites') }}</p>
          <ul class="m-0 flex list-none flex-wrap gap-5 p-0">
            <li v-for="item in siteVisuals" :key="`${item.site_id}:${item.visual.asset}`" class="min-w-0">
              <InsightsOverviewSiteVisual :item="item" />
            </li>
          </ul>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import type { InsightOverview } from '@/types/insights'
import { formatInsightRatio } from '@/utils/insightDimensions'
import { overviewEntityCount, selectOverviewEcosystemMetrics } from '@/utils/insightOverviewPresentation'
import { overviewGameVisuals, overviewSiteVisuals } from './visuals'

const props = defineProps<{ nav: InsightOverview | null, game: InsightOverview | null }>()
const { locale } = useI18n()
const localePath = useLocalePath()
// Reserve candidate zero for Hero even when its metric/image is unavailable.
// Only dedicated Overview projections qualify decoration; events never do.
const failedAssets = reactive(new Set<string>())
const previews = computed(() => overviewGameVisuals(props.game?.featured_visuals).slice(1).filter(item => !failedAssets.has(item.visual.asset)))
const siteVisuals = computed(() => overviewSiteVisuals(props.nav))
const sides = computed(() => [
  { domain: 'game' as const, overview: props.game, path: '/insights/games' },
  { domain: 'site' as const, overview: props.nav, path: '/insights/sites' },
].map(side => ({ ...side, count: overviewEntityCount(side.overview?.entity_count), metrics: selectOverviewEcosystemMetrics(side.overview, side.domain) })))
const number = (value: number) => new Intl.NumberFormat(locale.value === 'en' ? 'en-US' : 'zh-CN').format(value)
</script>
