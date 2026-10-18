<template>
  <dl class="site-observation-evidence-list min-w-0 space-y-1.5">
    <div v-for="item in items" :key="item.key" :data-site-evidence="item.key" class="grid min-w-0 gap-x-4 gap-y-1 sm:grid-cols-[minmax(0,1fr)_minmax(0,3fr)]">
      <dt class="site-observation-caption break-words">{{ item.label }}</dt>
      <dd :data-tone="item.tone" class="site-observation-evidence-value min-w-0 whitespace-pre-wrap">{{ item.value }}</dd>
    </div>
  </dl>
</template>

<script setup lang="ts">
import type { ObservationFact } from '~/utils/siteObservationPresentation'
defineProps<{ items: ObservationFact[] }>()
</script>
