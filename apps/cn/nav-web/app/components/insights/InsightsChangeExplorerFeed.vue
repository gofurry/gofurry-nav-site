<template>
  <section class="insights-section insights-change-explorer-feed" aria-labelledby="insights-change-explorer-feed-title">
    <div class="insights-section__heading">
      <div>
        <p class="insights-eyebrow">{{ $t('insights.changeExplorer.feedEyebrow') }}</p>
        <h2 id="insights-change-explorer-feed-title">{{ $t('insights.changeExplorer.feedTitle') }}</h2>
      </div>
    </div>
    <p v-if="unavailable && items.length === 0" class="insights-empty-state">{{ $t('insights.emptyStates.unavailable') }}</p>
    <p v-else-if="!loading && items.length === 0" class="insights-empty-state">{{ $t('insights.changeExplorer.empty') }}</p>
    <div v-else class="insights-change-list">
      <NuxtLink
        v-for="(item, index) in items"
        :key="`${item.domain}:${item.entity.id}:${item.date}:${item.type}:${index}`"
        :to="localePath(entityPath(item))"
        class="insights-change insights-change-explorer-item"
      >
        <span class="insights-change__domain">{{ $t(`insights.changes.${item.domain}`) }}</span>
        <span class="insights-change__body">
          <strong>{{ item.entity.name || `#${item.entity.id}` }}</strong>
          <span>{{ $t(insightChangeI18nKey(item.type)) }}</span>
        </span>
        <span class="insights-change-explorer-item__meta">
          <small>{{ $t(`insights.changeExplorer.categories.${item.domain}.${item.category}`) }}</small>
          <time :datetime="item.occurred_at || item.date">{{ formatInsightChangeWhen(item, locale) }}</time>
        </span>
      </NuxtLink>
    </div>
    <div v-if="nextCursor || loading" class="insights-change-explorer-feed__more">
      <button type="button" data-load-more :disabled="loading" @click="$emit('more')">
        {{ loading ? $t('insights.changeExplorer.loading') : $t('insights.changeExplorer.loadMore') }}
      </button>
    </div>
    <p v-if="unavailable && items.length > 0" class="insights-change-explorer-feed__inline-error">
      {{ $t('insights.changeExplorer.moreUnavailable') }}
    </p>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { InsightExplorerChange } from '@/types/insights'
import { formatInsightChangeWhen, insightChangeI18nKey } from '@/utils/insightChanges'

defineProps<{
  items: InsightExplorerChange[]
  nextCursor: string | null
  loading?: boolean
  unavailable?: boolean
}>()

defineEmits<{ more: [] }>()

const localePath = useLocalePath()
const { locale } = useI18n()

function entityPath(item: InsightExplorerChange) {
  return item.domain === 'site' ? `/site/${item.entity.id}` : `/games/${item.entity.id}`
}
</script>
