<template>
  <section class="overview-activity" aria-labelledby="overview-activity-title" data-overview-activity>
    <div class="overview-section-heading">
      <div><p class="overview-kicker">{{ $t('insights.editorial.activityKicker') }}</p><h2 id="overview-activity-title">{{ $t('insights.editorial.activityTitle') }}</h2></div>
      <NuxtLink :to="localePath(overviewChangesPath)" class="overview-text-link">{{ $t('insights.editorial.allChanges') }} <span aria-hidden="true">↗</span></NuxtLink>
    </div>
    <p v-if="!items.length" class="overview-unavailable">{{ $t(unavailable ? 'insights.emptyStates.unavailable' : 'insights.emptyStates.changesEmpty') }}</p>
    <div v-else class="overview-activity__layout" :class="{ 'overview-activity__layout--single': items.length === 1 }">
      <InsightActivityItem v-if="items[0]" :item="items[0]" hero />
      <div v-if="items.length > 1" class="overview-activity__list">
        <InsightActivityItem v-for="item in items.slice(1, 5)" :key="`${item.domain}:${item.entity.id}:${item.type}`" :item="item" />
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import type { InsightFeedItem } from '@/types/insights'
import { overviewChangesPath } from '@/utils/insightOverview'
import InsightActivityItem from './InsightActivityItem.vue'
defineProps<{ items: InsightFeedItem[], unavailable: boolean }>()
const localePath = useLocalePath()
</script>
