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
  padding: 0.95rem 1rem;
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.10);
  transition: background-color 500ms ease, box-shadow 500ms ease;
}

.metric-card:hover {
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.30), 0 0 0 4px rgba(251, 140, 47, 0.08);
}

.metric-card.tone-warm {
  background: rgba(255, 247, 235, 0.76);
}

.metric-card.tone-sky {
  background: rgba(239, 246, 255, 0.72);
}

.metric-card.tone-mint {
  background: rgba(240, 253, 244, 0.72);
}

.metric-card.tone-amber {
  background: rgba(255, 251, 235, 0.76);
}

.metric-card.tone-rose {
  background: rgba(255, 241, 242, 0.70);
}

.metric-card.tone-violet {
  background: rgba(245, 243, 255, 0.68);
}

.metric-card.tone-lime {
  background: rgba(247, 254, 231, 0.70);
}

.metric-card.tone-peach,
.metric-card.is-accent {
  background: rgba(255, 237, 213, 0.72);
}

.metric-card.tone-good {
  background: #e4f7ea;
  box-shadow: inset 0 0 0 1px rgba(22, 163, 74, 0.20);
}

.metric-card.tone-normal {
  background: #fff0c7;
  box-shadow: inset 0 0 0 1px rgba(217, 119, 6, 0.18);
}

.metric-card.tone-warn {
  background: #ffe5df;
  box-shadow: inset 0 0 0 1px rgba(220, 38, 38, 0.16);
}

.metric-card.tone-good:hover {
  background: #d8f2e1;
  box-shadow: inset 0 0 0 1px rgba(22, 163, 74, 0.28), 0 0 0 4px rgba(22, 163, 74, 0.08);
}

.metric-card.tone-normal:hover {
  background: #ffe8ad;
  box-shadow: inset 0 0 0 1px rgba(217, 119, 6, 0.26), 0 0 0 4px rgba(217, 119, 6, 0.08);
}

.metric-card.tone-warn:hover {
  background: #ffd9cf;
  box-shadow: inset 0 0 0 1px rgba(220, 38, 38, 0.24), 0 0 0 4px rgba(220, 38, 38, 0.07);
}

.metric-label {
  margin-bottom: 0.35rem;
  color: #64748b;
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.03em;
  text-transform: uppercase;
}

.metric-value {
  overflow-wrap: anywhere;
  color: #111827;
  font-size: 1rem;
  font-weight: 800;
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
