<template>
  <section class="game-insights-timeline" data-entity-timeline :data-timeline-mode="mode">
    <div class="game-insights-section-heading">
      <h3>{{ $t('insights.entity.gameTimelineTitle') }}</h3>
      <div class="game-insights-timeline__modes" :aria-label="$t('insights.entity.timelineView')">
        <button
          v-for="option in modes"
          :key="option"
          type="button"
          :class="{ 'is-active': mode === option }"
          :aria-pressed="mode === option"
          @click="mode = option"
        >
          {{ $t(`insights.entity.timelineModes.${option}`) }}
        </button>
      </div>
    </div>

    <p v-if="unavailable" class="entity-insights-empty">
      {{ $t('insights.emptyStates.unavailable') }}
    </p>
    <p v-else-if="orderedItems.length === 0" class="entity-insights-empty">
      {{ $t('insights.emptyStates.changesEmpty') }}
    </p>
    <ol v-else class="game-insights-timeline__list" :class="`game-insights-timeline__list--${mode}`">
      <li
        v-for="(item, index) in orderedItems"
        :key="`${item.type}:${item.occurred_at || item.date}`"
        class="game-insights-timeline__item"
        :data-connector="connector(index, orderedItems.length)"
        :style="compactPlacement(index)"
      >
        <span class="game-insights-timeline__category">{{ categoryLabel(item.type) }}</span>
        <strong>{{ eventLabel(item.type) }}</strong>
        <time :datetime="item.occurred_at || item.date">{{ formatWhen(item) }}</time>
      </li>
    </ol>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, type CSSProperties } from 'vue'
import { useI18n } from 'vue-i18n'
import type { InsightChange } from '@/types/insights'
import { formatInsightChangeWhen, insightChangeI18nKey, insightChangeOrder } from '@/utils/insightChanges'

type TimelineMode = 'compact' | 'list'

const props = defineProps<{
  items: InsightChange[]
  unavailable?: boolean
}>()

const { locale, t } = useI18n()
const modes: TimelineMode[] = ['compact', 'list']
const mode = ref<TimelineMode>('compact')
const orderedItems = computed(() => [...props.items].sort((left, right) => insightChangeOrder(right) - insightChangeOrder(left)))

function eventLabel(type: string) {
  return t(insightChangeI18nKey(type))
}

function formatWhen(item: InsightChange) {
  return formatInsightChangeWhen(item, locale.value)
}

function categoryLabel(type: string) {
  let key = 'other'
  if (type.startsWith('game.price.')) key = 'price'
  else if (type.startsWith('game.discount.')) key = 'discount'
  else if (type.startsWith('game.release.')) key = 'release'
  else if (/^game\.(windows|mac|linux)\./.test(type)) key = 'platform'
  else if (type.startsWith('game.free.')) key = 'pricingModel'
  return t(`insights.entity.changeCategories.${key}`)
}

function compactPlacement(index: number): CSSProperties {
  const position = index % 6
  const column = position < 3 ? position + 1 : 6 - position
  return {
    '--timeline-row': Math.floor(index / 3) + 1,
    '--timeline-column': column,
  } as CSSProperties
}

function connector(index: number, length: number) {
  if (index === length - 1) return 'none'
  const position = index % 6
  if (position === 0 || position === 1) return 'right'
  if (position === 3 || position === 4) return 'left'
  return 'down'
}
</script>
