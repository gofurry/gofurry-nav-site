<template>
  <div class="info-tab-body changes-panel">
    <div v-if="loading" class="panel-empty">{{ label('变化事件加载中', 'Loading changes') }}</div>
    <div
      v-for="event in events"
      :key="event.key"
      class="change-event-row"
    >
      <div class="change-event-head">
        <span class="protocol-badge">{{ event.protocol }}</span>
        <span class="change-field">{{ event.field }}</span>
        <span class="change-time">{{ event.detectedAt }}</span>
      </div>
      <div class="change-value-grid">
        <div class="change-value-block">
          <p class="change-value-label">{{ label('旧值', 'Old value') }}</p>
          <p class="change-value-text">{{ event.oldValue }}</p>
        </div>
        <div class="change-value-block">
          <p class="change-value-label">{{ label('新值', 'New value') }}</p>
          <p class="change-value-text">{{ event.newValue }}</p>
        </div>
      </div>
    </div>
    <div v-if="!loading && !events.length" class="panel-empty">{{ label('暂无变化事件', 'No change events') }}</div>
  </div>
</template>

<script setup lang="ts">
import { i18n } from '@/main'
import type { ChangeEventItem } from './detailTypes'

defineProps<{
  events: ChangeEventItem[]
  loading: boolean
}>()

function label(zh: string, en: string) {
  return i18n.global.locale.value === 'en' ? en : zh
}
</script>

<style scoped>
.info-tab-body {
  margin-top: 1.25rem;
}

.panel-empty {
  border-top: 1px solid rgba(251, 140, 47, 0.12);
  padding: 0.9rem 0;
  color: #64748b;
  font-size: 0.9rem;
}

:global(.dark .panel-empty){
  border-top-color: rgba(251, 146, 60, 0.16);
  color: #94a3b8;
}

.change-event-row {
  border-top: 1px solid rgba(251, 140, 47, 0.12);
  padding: 0.95rem 0;
}

:global(.dark .change-event-row){
  border-top-color: rgba(251, 146, 60, 0.16);
}

.change-event-row:first-of-type {
  border-top: 0;
  padding-top: 0;
}

.change-event-head {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.55rem;
}

.protocol-badge {
  border-radius: 999px;
  background: rgba(255, 237, 213, 0.76);
  padding: 0.22rem 0.6rem;
  color: #9a4a12;
  font-size: 0.74rem;
  font-weight: 800;
}

:global(.dark .protocol-badge){
  background: rgba(251, 146, 60, 0.14);
  color: #fed7aa;
}

.change-field {
  min-width: 0;
  overflow-wrap: anywhere;
  color: #111827;
  font-size: 0.92rem;
  font-weight: 800;
}

:global(.dark .change-field){
  color: #f8fafc;
}

.change-time {
  margin-left: auto;
  color: #64748b;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.78rem;
}

:global(.dark .change-time){
  color: #94a3b8;
}

.change-value-grid {
  margin-top: 0.8rem;
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.9rem 1.2rem;
}

.change-value-block {
  min-width: 0;
  border-left: 2px solid rgba(251, 140, 47, 0.18);
  padding-left: 0.75rem;
}

:global(.dark .change-value-block){
  border-left-color: rgba(251, 146, 60, 0.24);
}

.change-value-label {
  color: #64748b;
  font-size: 0.76rem;
  font-weight: 800;
}

:global(.dark .change-value-label){
  color: #94a3b8;
}

.change-value-text {
  margin-top: 0.25rem;
  min-width: 0;
  overflow-wrap: anywhere;
  color: #1f2937;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.86rem;
  line-height: 1.55;
}

:global(.dark .change-value-text){
  color: #e2e8f0;
}

@media (max-width: 640px) {
  .change-time {
    margin-left: 0;
    width: 100%;
  }

  .change-value-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}

@media (min-width: 768px) {
  .change-value-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
