<template>
  <NuxtLink :to="localePath(entityPath)" class="insight-activity-item" :class="{ 'insight-activity-item--hero': hero }" :data-domain="item.domain" data-change-link>
    <InsightEntityMedia :domain="item.domain" :entity="item.entity" :eager="hero" />
    <span class="insight-activity-item__body">
      <span v-if="hero" class="insight-activity-item__domain">{{ $t(`insights.changes.${item.domain}`) }}</span>
      <strong>{{ item.entity.name }}</strong>
      <span class="insight-activity-item__event">{{ $t(insightChangeI18nKey(item.type)) }}</span>
      <time :datetime="item.occurred_at || item.date">{{ formatInsightChangeWhen(item, locale) }}</time>
    </span>
    <span class="insight-activity-item__arrow" aria-hidden="true">↗</span>
  </NuxtLink>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { InsightFeedItem } from '@/types/insights'
import { formatInsightChangeWhen, insightChangeI18nKey } from '@/utils/insightChanges'
import { siteEntityPath } from '@/utils/siteRoutes'
import InsightEntityMedia from '../entity/InsightEntityMedia.vue'

const props = defineProps<{ item: InsightFeedItem, hero?: boolean }>()
const { locale } = useI18n()
const localePath = useLocalePath()
const entityPath = computed(() => props.item.domain === 'site' ? siteEntityPath(props.item.entity.id) : `/games/${encodeURIComponent(String(props.item.entity.id))}`)
</script>
