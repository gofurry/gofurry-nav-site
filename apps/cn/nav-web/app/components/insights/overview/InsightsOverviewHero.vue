<template>
  <section class="overview-hero mb-6 grid grid-cols-1 gap-6 min-[1200px]:grid-cols-3" :aria-label="$t('insights.overviewHero.label')" data-overview-hero data-hero-mode="data">
    <article v-for="side in sides" :key="side.domain" class="overview-hero__panel flex min-w-0 flex-col items-start" :class="{ 'min-[1200px]:col-span-2': side.domain === 'game' }" :data-hero-domain="side.domain" :data-hero-metric="side.metric?.key" :aria-labelledby="`overview-hero-${side.domain}`">
      <p class="overview-kicker">{{ $t(`insights.overviewHero.${side.domain}Kicker`) }}</p>
      <template v-if="side.metric">
        <h2 :id="`overview-hero-${side.domain}`" class="overview-hero__title">{{ $t(`insights.metrics.${side.metric.key}.name`) }}</h2>
        <p class="overview-hero__value">{{ formatInsightRatio(side.metric.value) }}</p>
        <p class="overview-hero__description">{{ $t(`insights.metrics.${side.metric.key}.description`) }}</p>
        <p class="overview-hero__scope">{{ $t('insights.overviewHero.observedSample') }}</p>
        <progress class="overview-hero__track w-full" :value="side.metric.value!" max="1" :aria-labelledby="`overview-hero-${side.domain}`" />
        <dl class="overview-hero__facts grid w-full grid-cols-1 gap-1">
          <div class="flex flex-wrap justify-between gap-x-4"><dt>{{ $t('insights.overviewHero.coverage') }}</dt><dd>{{ formatInsightRatio(side.metric.coverage) }}</dd></div>
          <div v-if="overviewSample(side.metric)" class="flex flex-wrap justify-between gap-x-4"><dt>{{ $t('insights.overviewHero.sample') }}</dt><dd>{{ overviewSample(side.metric) }}</dd></div>
          <div class="flex flex-wrap justify-between gap-x-4"><dt>{{ $t('insights.overviewHero.change30d') }}</dt><dd data-hero-delta>{{ side.delta === null ? '—' : $t('insights.overviewHero.percentagePoints', { value: side.delta }) }}</dd></div>
          <div class="flex flex-wrap justify-between gap-x-4"><dt>{{ $t('insights.overviewHero.factDate') }}</dt><dd><time v-if="side.metric.as_of" :datetime="side.metric.as_of">{{ side.metric.as_of }}</time><span v-else>{{ $t('insights.overviewHero.dateUnavailable') }}</span></dd></div>
        </dl>
      </template>
      <template v-else>
        <h2 :id="`overview-hero-${side.domain}`" class="overview-hero__title">{{ $t(`insights.overviewHero.${side.domain}FallbackTitle`) }}</h2>
        <p class="overview-hero__description">{{ $t(`insights.overviewHero.${side.domain}FallbackDescription`) }}</p>
        <p class="overview-hero__scope" data-hero-unavailable>{{ $t('insights.emptyStates.unavailable') }}</p>
      </template>
      <NuxtLink :to="localePath({ path: `/insights/${side.domain === 'game' ? 'games' : 'sites'}`, query: side.metric ? { metric: side.metric.key } : {} })" class="overview-hero__link mt-auto inline-flex items-center gap-3">{{ $t(`insights.overviewHero.${side.domain}Link`) }} <span aria-hidden="true">↗</span></NuxtLink>
    </article>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { InsightOverview } from '@/types/insights'
import { formatInsightRatio } from '@/utils/insightDimensions'
import { formatOverviewPercentagePoints, overviewHeroGameKeys, overviewHeroSiteKeys, overviewSample, selectOverviewMetric } from '@/utils/insightOverviewPresentation'

const props = defineProps<{ nav: InsightOverview | null, game: InsightOverview | null }>()
const { locale } = useI18n()
const localePath = useLocalePath()
const sides = computed(() => [
  { domain: 'game', metric: selectOverviewMetric(props.game, overviewHeroGameKeys) },
  { domain: 'site', metric: selectOverviewMetric(props.nav, overviewHeroSiteKeys) },
].map(side => ({ ...side, delta: formatOverviewPercentagePoints(side.metric?.delta_30d, locale.value) })))
</script>
