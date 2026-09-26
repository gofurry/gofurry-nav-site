<template>
  <dl class="site-observation-facts min-w-0">
    <div v-for="item in items" :key="item.key" :data-site-evidence="item.key" class="grid min-w-0 gap-1 sm:grid-cols-[minmax(8rem,1fr)_minmax(0,3fr)] sm:gap-4">
      <dt class="site-detail-note">{{ item.label }}</dt>
      <dd class="min-w-0 whitespace-pre-wrap break-words">{{ item.value }}</dd>
    </div>
  </dl>
</template>

<script setup lang="ts">
import type { ObservationFact } from '~/utils/siteObservationPresentation'
defineProps<{ items: ObservationFact[] }>()
</script>
