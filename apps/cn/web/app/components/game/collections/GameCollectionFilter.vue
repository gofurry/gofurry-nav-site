<template>
  <div class="game-search-filter-overlay fixed inset-0 z-50 flex items-center justify-center px-4">
    <div ref="panel" role="dialog" aria-modal="true" :aria-labelledby="`${id}-title`" tabindex="-1" class="game-search-filter-panel w-full max-w-2xl overflow-hidden p-6">
      <div class="max-h-[calc(80vh-3rem)] space-y-6 overflow-y-auto">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <h2 :id="`${id}-title`" class="game-search-filter-title">{{ t('game.collections.advancedFilter') }}</h2>
          <div class="flex gap-2">
            <button ref="cancelButton" type="button" class="game-search-filter-action game-search-filter-action--ghost" @click="emit('close')">{{ t('common.cancel') }}</button>
            <button type="button" class="game-search-filter-action game-search-filter-action--primary" @click="emit('apply', { ...draft })">{{ t('game.collections.apply') }}</button>
          </div>
        </div>
        <fieldset>
          <legend class="game-search-filter-label">{{ t('game.collections.phaseLabel') }}</legend>
          <div class="mt-2 flex flex-wrap gap-2">
            <button v-for="phase in phases" :key="phase" type="button" :aria-pressed="draft.phase === phase" class="game-search-filter-chip" :class="draft.phase === phase ? 'game-search-filter-chip--active' : 'game-search-filter-chip--idle'" @click="draft.phase = phase">{{ t(`game.collections.phases.${phase}`) }}</button>
          </div>
        </fieldset>
        <fieldset>
          <legend class="game-search-filter-label">{{ t('game.collections.sortLabel') }}</legend>
          <div class="mt-2 flex flex-wrap gap-2">
            <button v-for="sort in sorts" :key="sort" type="button" :aria-pressed="draft.sort === sort" class="game-search-filter-chip" :class="draft.sort === sort ? 'game-search-filter-chip--active' : 'game-search-filter-chip--idle'" @click="draft.sort = sort">{{ t(`game.collections.sorts.${sort}`) }}</button>
          </div>
        </fieldset>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { reactive, ref, useId } from 'vue'
import type { GameCollectionCriteria, GameCollectionPhaseFilter, GameCollectionSort } from '~/types/game'
import { useGameSearchDialog } from '~/composables/useGameSearchDialog'
const props = defineProps<{ criteria: GameCollectionCriteria }>()
const emit = defineEmits<{ close: []; apply: [filter: Pick<GameCollectionCriteria, 'phase' | 'sort'>] }>()
const { t } = useI18n()
const id = useId()
const panel = ref<HTMLElement | null>(null)
const cancelButton = ref<HTMLElement | null>(null)
const draft = reactive({ phase: props.criteria.phase, sort: props.criteria.sort })
const phases: GameCollectionPhaseFilter[] = ['all', 'released', 'upcoming', 'mixed']
const sorts: GameCollectionSort[] = ['published_desc', 'count_desc', 'count_asc', 'name_asc', 'name_desc']
useGameSearchDialog(panel, { dismiss: () => emit('close'), initialFocus: () => cancelButton.value })
</script>
