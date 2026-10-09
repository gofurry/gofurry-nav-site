<template>
  <section class="overview-activity" aria-labelledby="overview-activity-title" data-overview-activity>
    <div class="overview-section-heading">
      <div><p class="overview-kicker">{{ $t('insights.editorial.activityKicker') }}</p><h2 id="overview-activity-title">{{ $t('insights.editorial.activityTitle') }}</h2></div>
      <NuxtLink :to="localePath(overviewChangesPath)" class="overview-text-link">{{ $t('insights.editorial.allChanges') }} <span aria-hidden="true">↗</span></NuxtLink>
    </div>
    <p v-if="!items.length" class="overview-unavailable">{{ $t(unavailable ? 'insights.emptyStates.unavailable' : 'insights.emptyStates.changesEmpty') }}</p>
    <div v-else class="overview-activity__list grid grid-cols-1 min-[768px]:grid-cols-2 gap-x-6">
      <InsightActivityItem v-for="(item, index) in items.slice(0, 5)" :key="`${item.domain}:${item.entity.id}:${item.type}:${index}`" :item="overviewActivityWithoutGameArt(item)" />
    </div>
  </section>
</template>

<script setup lang="ts">
import type { InsightFeedItem } from '@/types/insights'
import { overviewChangesPath } from '@/utils/insightOverview'
import { overviewActivityWithoutGameArt } from '@/utils/insightOverviewPresentation'
import InsightActivityItem from './InsightActivityItem.vue'
defineProps<{ items: InsightFeedItem[], unavailable: boolean }>()
const localePath = useLocalePath()
</script>
