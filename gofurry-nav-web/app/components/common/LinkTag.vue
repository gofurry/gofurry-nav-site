<template>
  <a
      :href="item.value"
      target="_blank"
      rel="noopener noreferrer"
      class="link-tag px-3 py-1 text-sm"
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
