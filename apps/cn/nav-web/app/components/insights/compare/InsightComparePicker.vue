<template>
  <section class="insight-compare-picker" :aria-label="$t('insights.compare.builderTitle')">
    <div class="insight-compare-picker__search">
      <label :for="inputID">{{ $t('insights.comparePicker.' + domain + 'Label') }}</label>
      <input :id="inputID" ref="input" v-model="keyword" type="search" autocomplete="off" role="combobox"
        :placeholder="$t('insights.comparePicker.' + domain + 'Placeholder')" :aria-expanded="open"
        :aria-controls="listID" :aria-activedescendant="open && active >= 0 ? listID + '-' + active : undefined"
        aria-autocomplete="list" @focus="open = true" @blur="open = false" @keydown="onKey" @input="open = true" />
      <div v-if="open" class="insight-compare-picker__popup">
        <p v-if="loading" role="status">{{ $t('insights.comparePicker.loading') }}</p>
        <p v-else-if="failed" role="status">{{ $t('insights.comparePicker.failed') }}</p>
        <p v-else-if="!results.length" role="status">{{ $t(domain === 'game' && !keyword.trim() ? 'insights.comparePicker.startSearch' : 'insights.comparePicker.noResults') }}</p>
        <ul :id="listID" role="listbox" :aria-label="$t('insights.comparePicker.results')" :aria-busy="loading">
          <li v-for="(item, index) in results" :id="listID + '-' + index" :key="item.id" role="option"
            :aria-selected="active === index" :aria-disabled="selectedIds.length >= 4" :data-picker-result="item.id"
            @mousedown.prevent @click="add(item)" @mousemove="active = index">
            <InsightEntityMedia :domain="domain" :entity="item" />
            <span><strong>{{ item.name }}</strong><small>{{ item.subtitle }}</small></span>
            <span class="insight-compare-picker__add" aria-hidden="true">+</span>
          </li>
        </ul>
      </div>
    </div>
    <p class="insight-compare-picker__count" aria-live="polite">{{ $t('insights.comparePicker.count', { count: selectedIds.length }) }} · {{ $t(selectedIds.length >= 4 ? 'insights.comparePicker.maximum' : selectedIds.length === 1 ? 'insights.compare.oneHint' : selectedIds.length === 0 ? 'insights.compare.emptyHint' : 'insights.compare.readyHint', { count: selectedIds.length }) }}</p>
    <ul class="insight-compare-selected" :aria-label="$t('insights.comparePicker.selected')">
      <li v-for="entity in selected" :key="entity.id" :data-selected-entity="entity.id">
        <InsightEntityMedia :domain="domain" :entity="entity" />
        <span>{{ entity.name }}</span>
        <button type="button" :aria-label="$t('insights.comparePicker.remove', { name: entity.name })" @click="$emit('change', selectedIds.filter(id => id !== entity.id))">×</button>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, useId, watch } from 'vue'
import InsightEntityMedia from '@/components/insights/entity/InsightEntityMedia.vue'
import type { InsightDomain, InsightEntityRef } from '@/types/insights'
import type { ComparePickerEntity } from '@/types/insightCompare'
import { compareSelectedEntities, filterCompareSites } from '@/utils/insightComparePicker'
const props = defineProps<{ domain: InsightDomain; selectedIds: number[]; entities: InsightEntityRef[] }>()
const emit = defineEmits<{ change: [ids: number[]] }>()
const { keyword, items, cache, loading, failed } = useInsightComparePicker(props.domain)
watch(() => props.entities, entities => {
  const merged = new Map(cache.value.map(item => [item.id, item]))
  for (const entity of entities) merged.set(entity.id, entity)
  cache.value = [...merged.values()]
}, { immediate: true })
const inputID = useId(), listID = useId()
const input = ref<HTMLInputElement | null>(null)
const open = ref(false), active = ref(-1)
const selected = computed(() => compareSelectedEntities(props.selectedIds, props.entities, cache.value))
const results = computed(() => props.domain === 'site' ? filterCompareSites(items.value, keyword.value, props.selectedIds)
  : items.value.filter(item => !props.selectedIds.includes(item.id)).slice(0, 10))
watch(results, () => { active.value = -1 })
function add(item: ComparePickerEntity) {
  if (props.selectedIds.length >= 4 || props.selectedIds.includes(item.id)) return
  emit('change', [...props.selectedIds, item.id])
  open.value = false
  input.value?.focus()
  open.value = false
}
async function onKey(event: KeyboardEvent) {
  if (event.key === 'Escape') { open.value = false; event.preventDefault(); return }
  if (event.key === 'Enter' && open.value && active.value >= 0) {
    event.preventDefault()
    const item = results.value[active.value]
    if (item) add(item)
  }
  if (!['ArrowDown', 'ArrowUp'].includes(event.key)) return
  event.preventDefault()
  open.value = true
  if (!results.value.length) return
  active.value = active.value < 0 ? (event.key === 'ArrowDown' ? 0 : results.value.length - 1)
    : (active.value + (event.key === 'ArrowDown' ? 1 : -1) + results.value.length) % results.value.length
  await nextTick()
  document.getElementById(listID + '-' + active.value)?.scrollIntoView({ block: 'nearest' })
}
</script>
