<template>
  <section class="insights-section insights-recent-changes" aria-labelledby="insights-changes-title">
    <div class="insights-section__heading">
      <div>
        <p class="insights-eyebrow">{{ $t('insights.overview.recentDescription') }}</p>
        <h2 id="insights-changes-title">{{ $t('insights.changes.title') }}</h2>
      </div>
    </div>

    <p v-if="unavailable" class="insights-empty-state">
      {{ $t('insights.emptyStates.unavailable') }}
    </p>
    <p v-else-if="items.length === 0" class="insights-empty-state">
      {{ $t('insights.emptyStates.changesEmpty') }}
    </p>
    <div v-else class="insights-change-list">
      <NuxtLink
        v-for="item in items"
        :key="`${item.domain}:${item.entity.id}:${item.type}:${item.occurred_at || item.date}`"
        :to="localePath(entityPath(item))"
        class="insights-change"
        data-change-link
      >
        <span class="insights-change__domain">{{ $t(`insights.changes.${item.domain}`) }}</span>
        <span class="insights-change__body">
          <strong>{{ item.entity.name }}</strong>
          <span>{{ eventLabel(item.type) }}</span>
        </span>
        <time :datetime="item.occurred_at || item.date">{{ formatWhen(item) }}</time>
      </NuxtLink>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { InsightFeedItem } from '@/types/insights'
import { formatInsightChangeWhen, insightChangeI18nKey } from '@/utils/insightChanges'
import { siteDetailPath } from '@/utils/siteRoutes'

defineProps<{
  items: InsightFeedItem[]
  unavailable?: boolean
}>()

const localePath = useLocalePath()
const { locale, t } = useI18n()

function eventLabel(type: string) {
  return t(insightChangeI18nKey(type))
}

function entityPath(item: InsightFeedItem) {
  return item.domain === 'site' ? siteDetailPath(item.entity.id) : `/games/${encodeURIComponent(String(item.entity.id))}`
}

function formatWhen(item: InsightFeedItem) {
  return formatInsightChangeWhen(item, locale.value)
}
</script>
