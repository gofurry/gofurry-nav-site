<template>
  <section class="overview-sites" aria-labelledby="overview-sites-title" data-overview-sites>
    <div class="overview-section-heading">
      <div><p class="overview-kicker">{{ $t('insights.editorial.sitesKicker') }}</p><h2 id="overview-sites-title">{{ $t('insights.editorial.sitesTitle') }}</h2></div>
      <NuxtLink :to="localePath('/insights/sites')" class="overview-text-link">{{ $t('insights.editorial.explore') }} <span aria-hidden="true">↗</span></NuxtLink>
    </div>
    <p class="overview-intro">{{ $t('insights.editorial.sitesDescription') }}</p>
    <p v-if="!overview" class="overview-unavailable">{{ $t('insights.emptyStates.unavailable') }}</p>
    <div v-else class="overview-signals">
      <div v-for="metric in metrics" :key="metric.key" class="overview-signal" :data-metric="metric.key">
        <div class="overview-signal__label"><span :id="`overview-metric-${metric.key}`">{{ $t(`insights.metrics.${metric.key}.name`) }}</span><strong>{{ formatInsightRatio(metric.value) }}</strong></div>
        <progress v-if="metric.signal !== null" :value="metric.signal" max="1" :aria-labelledby="`overview-metric-${metric.key}`" />
        <span v-else class="overview-signal__missing">{{ $t('insights.emptyStates.unavailable') }}</span>
        <p>{{ formatOverviewDelta(metric.delta) }} <span>{{ $t('insights.editorial.deltaWindow') }}</span></p>
      </div>
    </div>
    <div v-if="identities.length" class="overview-site-identities">
      <p class="overview-kicker">{{ $t('insights.editorial.recentSites') }}</p>
      <div>
        <NuxtLink v-for="entity in identities" :key="entity.id" :to="localePath(siteEntityPath(entity.id))" :title="entity.name">
          <InsightEntityMedia domain="site" :entity="entity" />
        </NuxtLink>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { InsightOverview } from '@/types/insights'
import { formatInsightRatio } from '@/utils/insightDimensions'
import { formatOverviewDelta, overviewSignal, overviewSiteIdentities, overviewSiteMetricKeys } from '@/utils/insightOverview'
import { siteEntityPath } from '@/utils/siteRoutes'
import InsightEntityMedia from '../entity/InsightEntityMedia.vue'

const props = defineProps<{ overview: InsightOverview | null }>()
const localePath = useLocalePath()
const identities = computed(() => overviewSiteIdentities(props.overview))
const metrics = computed(() => overviewSiteMetricKeys.map(key => {
  const metric = props.overview?.metrics.find(item => item.key === key)
  return { key, value: metric?.value ?? null, signal: overviewSignal(metric?.value), delta: metric?.delta_30d }
}))
</script>
