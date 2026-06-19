<template>
  <a
      :href="item.value"
      target="_blank"
      rel="noopener noreferrer"
      class="px-3 py-1 rounded-lg bg-orange-100 text-orange-700
           hover:bg-orange-200 transition text-sm"
  >
    {{ item.key }}
  </a>
</template>

<script setup lang="ts">
import type { KvModel } from '@/types/game'

defineProps<{
  item: KvModel
}>()
</script>
