<template>
  <section class="insights-change-explorer-feed" aria-labelledby="insights-change-explorer-feed-title" :aria-busy="loading">
    <h2 id="insights-change-explorer-feed-title">{{ $t('insights.changeExplorer.feedTitle') }}</h2>
    <div v-if="unavailable && items.length === 0" class="insights-changes-state" role="status">
      <p>{{ $t('insights.emptyStates.unavailable') }}</p>
      <button type="button" data-retry-changes @click="$emit('retry')">{{ $t('insights.changeExplorer.retry') }}</button>
    </div>
    <p v-else-if="!loading && items.length === 0" class="insights-changes-state" role="status">{{ $t('insights.changeExplorer.empty') }}</p>
    <div v-else class="insights-changes-timeline">
      <section v-for="(group, groupIndex) in groups" :key="`${group.date}:${groupIndex}`" class="insights-changes-day" :aria-labelledby="`changes-date-${groupIndex}`" :data-change-date="group.date">
        <h3 :id="`changes-date-${groupIndex}`"><time :datetime="group.date">{{ formatInsightChangeDate(group.date, locale) }}</time></h3>
        <ol class="insights-changes-events">
          <li v-for="(item, index) in group.items" :key="`${item.domain}:${item.entity.id}:${item.type}:${index}`">
            <NuxtLink :to="localePath(entityPath(item))" class="insights-change-explorer-item" :data-event-type="item.type">
              <InsightEntityMedia :domain="item.domain" :entity="{ ...item.entity, name: item.entity.name || `#${item.entity.id}` }" />
              <span class="insights-change-explorer-item__body">
                <strong>{{ item.entity.name || `#${item.entity.id}` }}</strong>
                <span class="insights-change-explorer-item__event">{{ $t(insightChangeI18nKey(item.type)) }}</span>
                <span class="insights-change-explorer-item__context">
                  {{ $t(`insights.changes.${item.domain}`) }} · {{ $t(`insights.changeExplorer.categories.${item.domain}.${item.category}`) }}
                </span>
              </span>
              <time class="insights-change-explorer-item__time" :datetime="item.occurred_at || item.date">{{ formatInsightChangeWhen(item, locale) }}</time>
            </NuxtLink>
          </li>
        </ol>
      </section>
    </div>
    <p v-if="loading" class="insights-changes-state" role="status">{{ $t('insights.changeExplorer.loading') }}</p>
    <div v-if="nextCursor" class="insights-change-explorer-feed__more">
      <button type="button" data-load-more :disabled="loading" @click="$emit('more')">
        {{ $t('insights.changeExplorer.loadMore') }}
      </button>
    </div>
    <p v-if="unavailable && items.length > 0" class="insights-change-explorer-feed__inline-error" role="status">
      {{ $t('insights.changeExplorer.moreUnavailable') }}
    </p>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import InsightEntityMedia from '@/components/insights/entity/InsightEntityMedia.vue'
import type { InsightExplorerChange } from '@/types/insights'
import { formatInsightChangeWhen, insightChangeI18nKey } from '@/utils/insightChanges'
import { formatInsightChangeDate, groupInsightChangeDates } from '@/utils/insightChangeTimeline'

const props = defineProps<{
  items: InsightExplorerChange[]
  nextCursor: string | null
  loading?: boolean
  unavailable?: boolean
}>()

defineEmits<{ more: []; retry: [] }>()
const groups = computed(() => groupInsightChangeDates(props.items))
const localePath = useLocalePath()
const { locale } = useI18n()

function entityPath(item: InsightExplorerChange) {
  return item.domain === 'site' ? `/site/${item.entity.id}` : `/games/${item.entity.id}`
}
</script>
