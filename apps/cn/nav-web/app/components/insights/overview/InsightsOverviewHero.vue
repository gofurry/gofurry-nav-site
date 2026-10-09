<template>
  <section class="overview-hero mb-14 grid grid-cols-1 gap-6 min-[1200px]:grid-cols-3" :aria-label="$t('insights.overviewHero.label')" data-overview-hero data-hero-mode="data">
    <article class="overview-hero__panel overview-hero__game flex min-w-0 flex-col items-start gap-4 min-[1200px]:col-span-2" :class="{ 'self-start': !gameMetric }" data-hero-domain="game" :data-hero-metric="gameMetric?.key" aria-labelledby="overview-hero-game">
      <p class="overview-kicker">{{ $t('insights.overviewHero.gameKicker') }}</p>
      <div v-if="gameMetric" class="overview-hero__cover grid w-full min-w-0 grid-cols-1 gap-6 min-[1200px]:grid-cols-[minmax(0,13fr)_minmax(0,7fr)]">
        <div class="min-w-0" data-hero-primary>
          <h2 id="overview-hero-game" class="overview-hero__title">{{ $t(`insights.metrics.${gameMetric.key}.name`) }}</h2>
          <p class="overview-hero__value" data-hero-value>{{ formatInsightRatio(gameMetric.value) }}</p>
          <p class="overview-hero__description">{{ $t(`insights.metrics.${gameMetric.key}.description`) }}</p>
          <p class="overview-hero__scope">{{ $t('insights.overviewHero.observedSample') }}</p>
          <progress class="overview-hero__track block w-full" :value="gameMetric.value!" max="1" aria-labelledby="overview-hero-game" />
        </div>
        <dl class="overview-hero__evidence grid min-w-0 grid-cols-2 content-start gap-x-6 gap-y-4 min-[1200px]:grid-cols-1" data-hero-evidence>
          <div data-hero-fact="delta"><dt>{{ $t('insights.overviewHero.change30d') }}</dt><dd data-hero-delta>{{ gameDelta === null ? '—' : $t('insights.overviewHero.percentagePoints', { value: gameDelta }) }}</dd></div>
          <div data-hero-fact="coverage"><dt>{{ $t('insights.overviewHero.coverage') }}</dt><dd>{{ formatInsightRatio(gameMetric.coverage) }}</dd></div>
          <div v-if="overviewSample(gameMetric)" data-hero-fact="sample"><dt>{{ $t('insights.overviewHero.sample') }}</dt><dd>{{ overviewSample(gameMetric) }}</dd></div>
          <div data-hero-fact="date"><dt>{{ $t('insights.overviewHero.factDate') }}</dt><dd><time v-if="gameMetric.as_of" :datetime="gameMetric.as_of">{{ gameMetric.as_of }}</time><span v-else>{{ $t('insights.overviewHero.dateUnavailable') }}</span></dd></div>
        </dl>
      </div>
      <div v-else class="overview-hero__empty">
        <h2 id="overview-hero-game" class="overview-hero__title">{{ $t('insights.overviewHero.gameFallbackTitle') }}</h2>
        <p class="overview-hero__description">{{ $t('insights.overviewHero.gameFallbackDescription') }}</p>
        <p class="overview-hero__scope" data-hero-unavailable>{{ $t('insights.emptyStates.unavailable') }}</p>
      </div>
      <NuxtLink :to="localePath({ path: '/insights/games', query: gameMetric ? { metric: gameMetric.key } : {} })" class="overview-hero__link mt-auto inline-flex items-center gap-3">{{ $t('insights.overviewHero.gameLink') }} <span aria-hidden="true">↗</span></NuxtLink>
    </article>

    <article class="overview-hero__panel overview-hero__site flex min-w-0 flex-col items-start gap-3" :class="{ 'self-start': !siteMetric }" data-hero-domain="site" :data-hero-metric="siteMetric?.key" aria-labelledby="overview-hero-site">
      <p class="overview-kicker">{{ $t('insights.overviewHero.siteKicker') }}</p>
      <template v-if="siteMetric">
        <div class="w-full min-w-0" data-hero-primary>
          <h2 id="overview-hero-site" class="overview-hero__title">{{ $t(`insights.metrics.${siteMetric.key}.name`) }}</h2>
          <p class="overview-hero__value" data-hero-value>{{ formatInsightRatio(siteMetric.value) }}</p>
          <p class="overview-hero__description">{{ $t(`insights.metrics.${siteMetric.key}.description`) }}</p>
          <p class="overview-hero__scope">{{ $t('insights.overviewHero.observedSample') }}</p>
          <progress class="overview-hero__track block w-full" :value="siteMetric.value!" max="1" aria-labelledby="overview-hero-site" />
        </div>
        <dl class="overview-hero__annotations grid w-full grid-cols-1 gap-1" data-hero-evidence>
          <div class="flex flex-wrap justify-between gap-x-4" data-hero-fact="delta"><dt>{{ $t('insights.overviewHero.change30d') }}</dt><dd data-hero-delta>{{ siteDelta === null ? '—' : $t('insights.overviewHero.percentagePoints', { value: siteDelta }) }}</dd></div>
          <div class="flex flex-wrap justify-between gap-x-4" data-hero-fact="coverage"><dt>{{ $t('insights.overviewHero.coverage') }}</dt><dd>{{ formatInsightRatio(siteMetric.coverage) }}</dd></div>
          <div v-if="overviewSample(siteMetric)" class="flex flex-wrap justify-between gap-x-4" data-hero-fact="sample"><dt>{{ $t('insights.overviewHero.sample') }}</dt><dd>{{ overviewSample(siteMetric) }}</dd></div>
          <div class="flex flex-wrap justify-between gap-x-4" data-hero-fact="date"><dt>{{ $t('insights.overviewHero.factDate') }}</dt><dd><time v-if="siteMetric.as_of" :datetime="siteMetric.as_of">{{ siteMetric.as_of }}</time><span v-else>{{ $t('insights.overviewHero.dateUnavailable') }}</span></dd></div>
        </dl>
      </template>
      <div v-else class="overview-hero__empty">
        <h2 id="overview-hero-site" class="overview-hero__title">{{ $t('insights.overviewHero.siteFallbackTitle') }}</h2>
        <p class="overview-hero__description">{{ $t('insights.overviewHero.siteFallbackDescription') }}</p>
        <p class="overview-hero__scope" data-hero-unavailable>{{ $t('insights.emptyStates.unavailable') }}</p>
      </div>
      <NuxtLink :to="localePath({ path: '/insights/sites', query: siteMetric ? { metric: siteMetric.key } : {} })" class="overview-hero__link mt-auto inline-flex items-center gap-3">{{ $t('insights.overviewHero.siteLink') }} <span aria-hidden="true">↗</span></NuxtLink>
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
const gameMetric = computed(() => selectOverviewMetric(props.game, overviewHeroGameKeys))
const siteMetric = computed(() => selectOverviewMetric(props.nav, overviewHeroSiteKeys))
const gameDelta = computed(() => formatOverviewPercentagePoints(gameMetric.value?.delta_30d, locale.value))
const siteDelta = computed(() => formatOverviewPercentagePoints(siteMetric.value?.delta_30d, locale.value))
</script>
