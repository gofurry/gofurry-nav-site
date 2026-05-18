<script setup lang="ts">
import type { MetricsRange } from '../types'

defineProps<{ modelValue: MetricsRange }>()
const emit = defineEmits<{ 'update:modelValue': [value: MetricsRange] }>()

const ranges: Array<{ label: string; value: MetricsRange }> = [
  { label: '1h', value: '1h' },
  { label: '6h', value: '6h' },
  { label: '24h', value: '24h' },
]
</script>

<template>
  <div class="segmented-control" aria-label="时间范围">
    <button
      v-for="item in ranges"
      :key="item.value"
      type="button"
      class="segmented-button"
      :class="{ active: item.value === modelValue }"
      @click="emit('update:modelValue', item.value)"
    >
      {{ item.label }}
    </button>
  </div>
</template>
