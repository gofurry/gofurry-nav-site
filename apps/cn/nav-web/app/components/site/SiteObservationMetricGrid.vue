<template>
  <div class="metric-grid">
    <div
      v-for="(item, index) in items"
      :key="item.label"
      class="metric-card"
      :class="`tone-${item.tone ?? metricToneClasses[index % metricToneClasses.length]}`"
    >
      <div class="metric-label">{{ item.label }}</div>
      <div class="metric-value">{{ item.value || '-' }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ObservationMetricItem, ObservationTone } from './detailTypes'

defineProps<{
  items: ObservationMetricItem[]
}>()

const metricToneClasses: ObservationTone[] = ['normal', 'good', 'normal', 'warn']
</script>

<style scoped>
.metric-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.8rem;
}

.metric-card {
  min-width: 0;
  border-radius: 8px;
  background: rgba(255, 250, 242, 0.42);
  padding: 0.95rem 1rem;
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.10);
  transition: background-color 500ms ease, box-shadow 500ms ease;
}

.metric-card:hover {
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.30), 0 0 0 4px rgba(251, 140, 47, 0.08);
}

:global(html.dark .metric-card){
  background: rgba(15, 23, 42, 0.52);
  box-shadow: inset 0 0 0 1px rgba(148, 163, 184, 0.12);
}

:global(html.dark .metric-card:hover){
  box-shadow: inset 0 0 0 1px rgba(251, 146, 60, 0.24), 0 0 0 4px rgba(251, 146, 60, 0.07);
}

.metric-card.tone-warm {
  background: rgba(255, 250, 242, 0.42);
}

:global(html.dark .metric-card.tone-warm){
  background: rgba(15, 23, 42, 0.52);
}

.metric-card.tone-sky {
  background: rgba(255, 250, 242, 0.42);
}

:global(html.dark .metric-card.tone-sky){
  background: rgba(15, 23, 42, 0.52);
}

.metric-card.tone-mint {
  background: rgba(255, 250, 242, 0.42);
}

:global(html.dark .metric-card.tone-mint){
  background: rgba(15, 23, 42, 0.52);
}

.metric-card.tone-amber {
  background: rgba(255, 250, 242, 0.42);
}

:global(html.dark .metric-card.tone-amber){
  background: rgba(15, 23, 42, 0.52);
}

.metric-card.tone-rose {
  background: rgba(255, 250, 242, 0.42);
}

:global(html.dark .metric-card.tone-rose){
  background: rgba(15, 23, 42, 0.52);
}

.metric-card.tone-violet {
  background: rgba(255, 250, 242, 0.42);
}

:global(html.dark .metric-card.tone-violet){
  background: rgba(15, 23, 42, 0.52);
}

.metric-card.tone-lime {
  background: rgba(255, 250, 242, 0.42);
}

:global(html.dark .metric-card.tone-lime){
  background: rgba(15, 23, 42, 0.52);
}

.metric-card.tone-peach,
.metric-card.is-accent {
  background: rgba(255, 250, 242, 0.42);
}

:global(html.dark .metric-card.tone-peach),
:global(html.dark .metric-card.is-accent){
  background: rgba(15, 23, 42, 0.52);
}

.metric-card.tone-good {
  background: rgba(255, 250, 242, 0.42);
  box-shadow: inset 0 0 0 1px rgba(22, 163, 74, 0.20);
}

:global(html.dark .metric-card.tone-good){
  background: rgba(15, 23, 42, 0.52);
  box-shadow: inset 0 0 0 1px rgba(34, 197, 94, 0.16);
}

.metric-card.tone-normal {
  background: rgba(255, 250, 242, 0.42);
  box-shadow: inset 0 0 0 1px rgba(217, 119, 6, 0.18);
}

:global(html.dark .metric-card.tone-normal){
  background: rgba(15, 23, 42, 0.52);
  box-shadow: inset 0 0 0 1px rgba(245, 158, 11, 0.16);
}

.metric-card.tone-warn {
  background: rgba(255, 250, 242, 0.42);
  box-shadow: inset 0 0 0 1px rgba(220, 38, 38, 0.16);
}

:global(html.dark .metric-card.tone-warn){
  background: rgba(15, 23, 42, 0.52);
  box-shadow: inset 0 0 0 1px rgba(248, 113, 113, 0.16);
}

.metric-card.tone-good:hover {
  background: #d8f2e1;
  box-shadow: inset 0 0 0 1px rgba(22, 163, 74, 0.28), 0 0 0 4px rgba(22, 163, 74, 0.08);
}

:global(html.dark .metric-card.tone-good:hover){
  background: rgba(22, 101, 52, 0.34);
  box-shadow: inset 0 0 0 1px rgba(34, 197, 94, 0.24), 0 0 0 4px rgba(34, 197, 94, 0.08);
}

.metric-card.tone-normal:hover {
  background: #ffe8ad;
  box-shadow: inset 0 0 0 1px rgba(217, 119, 6, 0.26), 0 0 0 4px rgba(217, 119, 6, 0.08);
}

:global(html.dark .metric-card.tone-normal:hover){
  background: rgba(146, 64, 14, 0.34);
  box-shadow: inset 0 0 0 1px rgba(245, 158, 11, 0.24), 0 0 0 4px rgba(245, 158, 11, 0.08);
}

.metric-card.tone-warn:hover {
  background: #ffd9cf;
  box-shadow: inset 0 0 0 1px rgba(220, 38, 38, 0.24), 0 0 0 4px rgba(220, 38, 38, 0.07);
}

:global(html.dark .metric-card.tone-warn:hover){
  background: rgba(153, 27, 27, 0.34);
  box-shadow: inset 0 0 0 1px rgba(248, 113, 113, 0.24), 0 0 0 4px rgba(248, 113, 113, 0.07);
}

.metric-label {
  margin-bottom: 0.35rem;
  color: #64748b;
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.03em;
  text-transform: uppercase;
}

:global(html.dark .metric-label){
  color: #94a3b8;
}

.metric-value {
  overflow-wrap: anywhere;
  color: #111827;
  font-size: 1rem;
  font-weight: 800;
}

:global(html.dark .metric-value){
  color: #f8fafc;
}

@media (min-width: 640px) {
  .metric-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 1024px) {
  .metric-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}
</style>
