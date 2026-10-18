<template>
  <div class="game-collection-timeline flex flex-col gap-8">
    <template v-for="section in sections" :key="section.key">
      <p v-if="section.key === 'future' && showNow" class="game-collection-timeline__now self-center" data-testid="collection-now">
        <strong>{{ t('game.collections.now') }}</strong> · <time :datetime="detail.as_of_date">{{ collectionDate(detail.as_of_date, locale) }}</time>
      </p>
      <section v-if="section.items.length" :data-phase="section.key">
        <h2 class="game-collection-section mb-4">{{ t(`game.collections.sections.${section.key}`) }}</h2>
        <ol class="game-collection-timeline__list grid gap-10" :class="section.snake ? 'game-collection-timeline__snake min-[900px]:grid-cols-3' : 'min-[900px]:grid-cols-3'">
          <li v-for="(item, index) in section.items" :key="item.game_id" class="game-collection-timeline__item relative min-w-0" :data-game-id="item.game_id" :data-connector="section.snake ? connector(index, section.items.length) : 'none'" :style="section.snake ? compactPlacement(index) : undefined">
            <GameCollectionTimelineItem :item="item" />
          </li>
        </ol>
      </section>
    </template>
  </div>
</template>
<script setup lang="ts">
import type { GameCollectionDetail } from '~/types/game'
import { compactPlacement, connector } from '~/utils/serpentineSequence'
import { collectionDate, collectionTimelineSections } from '~/utils/gameCollectionPresentation'
import GameCollectionTimelineItem from './GameCollectionTimelineItem.vue'
const props = defineProps<{ detail: GameCollectionDetail }>()
const { t, locale } = useI18n()
const sections = computed(() => collectionTimelineSections(props.detail.items))
const showNow = computed(() => props.detail.items.some(item => item.phase !== 'unknown'))
</script>
