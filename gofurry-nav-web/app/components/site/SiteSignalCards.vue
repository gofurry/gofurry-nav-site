<template>
  <section class="signal-grid">
    <article
      v-for="card in cards"
      :key="card.title"
      class="signal-card duration-500"
      :class="card.tone"
    >
      <p class="text-xs font-semibold text-slate-500 dark:text-slate-400">{{ card.eyebrow }}</p>
      <div class="mt-2 flex items-end justify-between gap-3">
        <h2 class="text-3xl font-black text-slate-900 dark:text-slate-50">{{ card.value }}</h2>
        <span class="text-xs font-semibold text-slate-500 dark:text-slate-400">{{ card.badge }}</span>
      </div>
      <dl class="mt-5 space-y-2 text-xs">
        <div v-for="item in card.items" :key="item.label" class="flex justify-between gap-4">
          <dt class="text-slate-500 dark:text-slate-400">{{ item.label }}</dt>
          <dd class="min-w-0 truncate font-mono font-semibold text-slate-800 dark:text-slate-200">{{ item.value }}</dd>
        </div>
      </dl>
    </article>
  </section>
</template>

<script setup lang="ts">
import type { SiteSignalCard } from './detailTypes'

defineProps<{
  cards: SiteSignalCard[]
}>()
</script>

<style scoped>
.signal-grid {
  margin-top: 1.5rem;
  display: grid;
  grid-template-columns: repeat(1, minmax(0, 1fr));
  gap: 1rem;
}

.signal-card {
  min-height: 13rem;
  border-radius: 24px;
  padding: 1.35rem;
  background: var(--surface);
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.7);
}

.signal-card:hover {
  background: rgba(255, 255, 255, 0.88);
  box-shadow: inset 0 0 0 1px rgba(251, 140, 47, 0.28), 0 0 0 6px rgba(251, 140, 47, 0.07);
}

:global(html.dark .signal-card){
  background: var(--surface);
  box-shadow: inset 0 0 0 1px rgba(148, 163, 184, 0.12);
}

:global(html.dark .signal-card:hover){
  background: rgba(30, 41, 59, 0.86);
  box-shadow: inset 0 0 0 1px rgba(251, 146, 60, 0.22), 0 0 0 6px rgba(251, 146, 60, 0.07);
}

.tone-green {
  border-left: 4px solid rgba(52, 211, 153, 0.86);
}

.tone-amber {
  border-left: 4px solid rgba(251, 191, 36, 0.9);
}

.tone-rose {
  border-left: 4px solid rgba(251, 113, 133, 0.9);
}

@media (min-width: 760px) {
  .signal-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 1280px) {
  .signal-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}
</style>
