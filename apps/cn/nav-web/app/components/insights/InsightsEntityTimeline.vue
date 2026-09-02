<template>
  <section class="entity-insights-timeline" data-entity-timeline>
    <div class="entity-insights-heading">
      <div>
        <p class="entity-insights-eyebrow">{{ $t('insights.entity.timelineEyebrow') }}</p>
        <h3>{{ $t('insights.entity.timelineTitle') }}</h3>
      </div>
    </div>

    <p v-if="unavailable" class="entity-insights-empty">
      {{ $t('insights.emptyStates.unavailable') }}
    </p>
    <p v-else-if="orderedItems.length === 0" class="entity-insights-empty">
      {{ $t('insights.emptyStates.changesEmpty') }}
    </p>
    <ol v-else class="entity-insights-timeline__list">
      <li
        v-for="item in orderedItems"
        :key="`${item.type}:${item.occurred_at || item.date}`"
        class="entity-insights-timeline__item"
      >
        <span aria-hidden="true" class="entity-insights-timeline__dot" />
        <div>
          <strong>{{ eventLabel(item.type) }}</strong>
          <time :datetime="item.occurred_at || item.date">{{ formatWhen(item) }}</time>
        </div>
      </li>
    </ol>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { InsightChange } from '@/types/insights'
import { formatInsightChangeWhen, insightChangeI18nKey, insightChangeOrder } from '@/utils/insightChanges'

const props = defineProps<{
  items: InsightChange[]
  unavailable?: boolean
}>()

const { locale, t } = useI18n()
const orderedItems = computed(() => [...props.items].sort((left, right) => insightChangeOrder(right) - insightChangeOrder(left)))

function eventLabel(type: string) {
  return t(insightChangeI18nKey(type))
}

function formatWhen(item: InsightChange) {
  return formatInsightChangeWhen(item, locale.value)
}
</script>
