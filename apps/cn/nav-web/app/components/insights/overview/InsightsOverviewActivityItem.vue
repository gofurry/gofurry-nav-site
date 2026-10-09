<template>
  <NuxtLink :to="localePath(entityPath)" class="overview-event flex items-start gap-4" :data-domain="item.domain" data-change-link>
    <span class="overview-event__icon flex shrink-0 items-center justify-center" aria-hidden="true" data-activity-icon><component :is="item.domain === 'game' ? PhGameController : PhGlobe" /></span>
    <span class="min-w-0 flex-1">
      <span class="overview-event__meta flex flex-wrap items-baseline gap-x-4 gap-y-1">
        <span>{{ $t(`insights.changes.${item.domain}`) }}</span>
        <time :datetime="hasExactTime ? item.occurred_at! : item.date">{{ formatInsightChangeWhen(item, locale, 'UTC') }}{{ hasExactTime ? ' UTC' : '' }}</time>
      </span>
      <strong class="block">{{ item.entity.name }}</strong>
      <span class="overview-event__description block">{{ $t(insightChangeI18nKey(item.type)) }}</span>
    </span>
    <span class="overview-event__arrow shrink-0" aria-hidden="true">↗</span>
  </NuxtLink>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { PhGameController, PhGlobe } from '@phosphor-icons/vue'
import type { InsightFeedItem } from '@/types/insights'
import { formatInsightChangeWhen, insightChangeI18nKey } from '@/utils/insightChanges'
import { siteEntityPath } from '@/utils/siteRoutes'

const props = defineProps<{ item: InsightFeedItem }>()
const { locale } = useI18n()
const localePath = useLocalePath()
const entityPath = computed(() => props.item.domain === 'site' ? siteEntityPath(props.item.entity.id) : `/games/${encodeURIComponent(String(props.item.entity.id))}`)
// Explicit UTC keeps exact timestamps identical on the server and browser;
// day-precision events remain dates and never gain a fabricated time.
const hasExactTime = computed(() => Boolean(props.item.occurred_at && Number.isFinite(Date.parse(props.item.occurred_at))))
</script>
