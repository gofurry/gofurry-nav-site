<template>
  <section class="insight-domain-activity" aria-labelledby="domain-activity-title" data-domain-activity>
    <div class="insight-domain-heading"><h2 id="domain-activity-title">{{ $t(`insights.domain.${domain}.activity`) }}</h2></div>
    <p v-if="!items.length" class="insights-empty-state">{{ $t(unavailable ? 'insights.emptyStates.unavailable' : 'insights.emptyStates.changesEmpty') }}</p>
    <div v-else class="insight-domain-activity__items"><InsightActivityItem v-for="item in items" :key="`${item.entity.id}:${item.type}`" :item="item" /></div>
  </section>
</template>

<script setup lang="ts">
import type { InsightDomain, InsightFeedItem } from '@/types/insights'
import InsightActivityItem from '../activity/InsightActivityItem.vue'
defineProps<{ domain: InsightDomain, items: InsightFeedItem[], unavailable: boolean }>()
</script>
