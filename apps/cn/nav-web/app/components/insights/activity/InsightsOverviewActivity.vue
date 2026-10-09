<template>
  <section class="overview-activity mb-[var(--insights-section-gap)] grid grid-cols-1 gap-8 min-[1200px]:grid-cols-4 min-[1200px]:gap-12" aria-labelledby="overview-activity-title" :data-activity-state="summary.availability" data-overview-activity>
    <header class="min-w-0" data-activity-heading>
      <h2 id="overview-activity-title" class="overview-heading">{{ $t('insights.editorial.activityTitle') }}</h2>
      <p class="overview-activity__intro">{{ $t('insights.overviewActivity.description') }}</p>
      <dl class="overview-activity__total flex flex-wrap items-baseline gap-3">
        <dt>{{ $t('insights.overview.changesCount') }}</dt>
        <dd data-activity-total>{{ summary.total === null ? '—' : number(summary.total) }}</dd>
      </dl>
    </header>
    <div class="min-w-0 min-[1200px]:col-span-3" data-activity-content>
      <p v-if="summary.availability === 'site-only' || summary.availability === 'game-only'" class="overview-activity__notice" data-activity-partial>{{ $t(`insights.overviewActivity.${summary.availability === 'site-only' ? 'gameUnavailable' : 'siteUnavailable'}`) }}</p>
      <p v-if="summary.availability === 'unavailable'" class="overview-activity__notice" data-activity-unavailable>{{ $t('insights.emptyStates.unavailable') }}</p>
      <p v-else-if="!items.length" class="overview-activity__notice" data-activity-empty>{{ $t(summary.availability === 'complete' ? 'insights.emptyStates.changesEmpty' : 'insights.overviewActivity.availableEmpty') }}</p>
      <ol v-else class="overview-activity__list m-0 list-none p-0" data-activity-list>
        <li v-for="(item, index) in items.slice(0, 5)" :key="`${item.domain}:${item.entity.id}:${item.type}:${index}`">
          <InsightsOverviewActivityItem :item="overviewActivityWithoutGameArt(item)" />
        </li>
      </ol>
      <NuxtLink :to="localePath(overviewChangesPath)" class="overview-text-link mt-6 inline-flex items-center gap-2" data-activity-all>{{ $t('insights.overviewActivity.allChanges') }} <span aria-hidden="true">↗</span></NuxtLink>
    </div>
  </section>
</template>

<script setup lang="ts">
import type { InsightFeedItem } from '@/types/insights'
import { useI18n } from 'vue-i18n'
import { overviewChangesPath } from '@/utils/insightOverview'
import { overviewActivityWithoutGameArt, type overviewActivitySummary } from '@/utils/insightOverviewPresentation'
import InsightsOverviewActivityItem from '../overview/InsightsOverviewActivityItem.vue'
defineProps<{ items: InsightFeedItem[], summary: ReturnType<typeof overviewActivitySummary> }>()
const { locale } = useI18n()
const localePath = useLocalePath()
const number = (value: number) => new Intl.NumberFormat(locale.value === 'en' ? 'en-US' : 'zh-CN').format(value)
</script>
