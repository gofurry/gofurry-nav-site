<template>
  <div v-if="items.length" class="info-list">
    <div
      v-for="(item, index) in items"
      :key="`${item.label}:${index}`"
      class="info-row"
      :class="`tone-${item.tone ?? headerToneClasses[index % headerToneClasses.length]}`"
    >
      <span class="info-label">{{ item.label }}</span>
      <span class="info-value">{{ item.value }}</span>
    </div>
  </div>
  <div v-else class="empty-state">{{ emptyText }}</div>
</template>

<script setup lang="ts">
import type { ObservationInfoItem, ObservationTone } from './detailTypes'

defineProps<{
  items: ObservationInfoItem[]
  emptyText: string
}>()

const headerToneClasses: ObservationTone[] = ['normal', 'good', 'normal', 'warn']
</script>

<style scoped>
.info-list {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0 1rem;
}

.info-row {
  display: grid;
  grid-template-columns: minmax(7rem, 0.32fr) minmax(0, 1fr);
  gap: 0.8rem;
  min-width: 0;
  border-bottom: 1px solid rgba(251, 140, 47, 0.10);
  border-left: 2px solid transparent;
  border-radius: 0;
  padding: 0.58rem 0.85rem 0.58rem 0.75rem;
  font-size: 0.9rem;
  transition: background-color 500ms ease, border-color 500ms ease, color 500ms ease;
}

:global(.dark .info-row){
  border-bottom-color: rgba(148, 163, 184, 0.12);
}

.info-row:hover {
  background: rgba(255, 237, 213, 0.68);
  border-left-color: rgba(251, 140, 47, 0.58);
}

:global(.dark .info-row:hover){
  background: rgba(251, 146, 60, 0.12);
  border-left-color: rgba(251, 146, 60, 0.54);
}

.info-row.tone-good,
.info-row.tone-normal,
.info-row.tone-warn,
.info-row.tone-warm,
.info-row.tone-mint,
.info-row.tone-sky,
.info-row.tone-amber {
  background: transparent;
  box-shadow: none;
}

.info-label {
  color: #64748b;
  font-weight: 800;
}

:global(.dark .info-label){
  color: #94a3b8;
}

.info-row:hover .info-label {
  color: #9a4a12;
}

:global(.dark .info-row:hover .info-label){
  color: #fdba74;
}

.info-value {
  overflow-wrap: anywhere;
  color: #1f2937;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
}

:global(.dark .info-value){
  color: #e2e8f0;
}

.info-row:hover .info-value {
  color: #111827;
}

:global(.dark .info-row:hover .info-value){
  color: #f8fafc;
}

.empty-state {
  border-radius: 8px;
  background: rgba(255, 250, 242, 0.54);
  padding: 1rem;
  color: #64748b;
  font-size: 0.875rem;
}

:global(.dark .empty-state){
  background: rgba(15, 23, 42, 0.54);
  color: #94a3b8;
}

@media (min-width: 1024px) {
  .info-list {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .info-row {
    grid-template-columns: minmax(0, 1fr);
    gap: 0.25rem;
  }
}
</style>
