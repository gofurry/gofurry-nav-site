<template>
  <span v-if="direction !== 'none'" class="site-detail-flow-arrow pointer-events-none absolute -bottom-6 left-1/2" :class="direction === 'right' ? 'md:bottom-auto md:left-auto md:-right-6 md:top-1/2' : direction === 'left' ? 'md:bottom-auto md:-left-6 md:top-1/2' : ''" :data-direction="direction" aria-hidden="true">
    <span class="md:hidden"><PhArrowDown class="site-detail-icon" /></span>
    <span class="hidden md:block"><component :is="direction === 'right' ? PhArrowRight : direction === 'left' ? PhArrowLeft : PhArrowDown" class="site-detail-icon" /></span>
  </span>
</template>
<script setup lang="ts">
import { PhArrowDown, PhArrowLeft, PhArrowRight } from '@phosphor-icons/vue'
defineProps<{ direction: string }>()
</script>
