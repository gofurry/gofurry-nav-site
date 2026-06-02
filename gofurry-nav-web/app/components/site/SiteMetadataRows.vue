<template>
  <div class="metadata-panel">
    <div v-if="items.length" class="metadata-list">
      <div
        v-for="item in items"
        :key="item.label"
        class="metadata-row"
      >
        <div class="metadata-label">{{ item.label }}</div>
        <div class="metadata-value">{{ displayValue(item.value) }}</div>
      </div>
    </div>
    <div v-else class="metadata-empty">{{ emptyText }}</div>
  </div>
</template>

<script setup lang="ts">
import type { DetailInfoItem } from './detailTypes'

defineProps<{
  emptyText: string
  items: DetailInfoItem[]
}>()

function displayValue(value: string | string[]) {
  return Array.isArray(value) ? value.join(', ') : value
}
</script>

<style scoped>
.metadata-panel {
  margin-top: 1.15rem;
}

.metadata-list {
  display: grid;
  gap: 0.74rem;
}

.metadata-row {
  display: grid;
  grid-template-columns: minmax(8rem, 13rem) minmax(0, 1fr);
  gap: clamp(1rem, 4vw, 2.4rem);
  align-items: start;
  min-width: 0;
}

.metadata-label {
  min-width: 0;
  color: #64748b;
  font-size: 0.92rem;
  font-weight: 800;
  line-height: 1.55;
}

:global(.dark .metadata-label){
  color: #94a3b8;
}

.metadata-value {
  min-width: 0;
  overflow-wrap: anywhere;
  color: #111827;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.9rem;
  line-height: 1.55;
}

:global(.dark .metadata-value){
  color: #e2e8f0;
}

.metadata-empty {
  color: #64748b;
  font-size: 0.9rem;
}

:global(.dark .metadata-empty){
  color: #94a3b8;
}

@media (max-width: 640px) {
  .metadata-row {
    grid-template-columns: minmax(0, 1fr);
    gap: 0.18rem;
  }
}
</style>
