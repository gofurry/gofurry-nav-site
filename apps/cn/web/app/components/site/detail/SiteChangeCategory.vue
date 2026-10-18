<template>
  <span class="site-detail-note inline-flex items-center gap-1">
    <component :is="icons[category] || PhClockCounterClockwise" class="site-detail-icon shrink-0" aria-hidden="true" />{{ label }}
  </span>
</template>
<script setup lang="ts">
import { PhPulse, PhArrowsLeftRight, PhCertificate, PhClockCounterClockwise } from '@phosphor-icons/vue'
import type { Component } from 'vue'
defineProps<{ category: string; label: string }>()
const icons: Record<string, Component> = { capability: PhPulse, target: PhArrowsLeftRight, certificate: PhCertificate }
</script>
