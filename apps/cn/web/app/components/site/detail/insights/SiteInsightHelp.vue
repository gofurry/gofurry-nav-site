<template>
  <span class="inline-flex items-center gap-1">{{ label }}<SiteDetailHelpTooltip :label="label" :text="help" /></span>
</template>
<script setup lang="ts">
import SiteDetailHelpTooltip from '../SiteDetailHelpTooltip.vue'
defineProps<{ label: string; help: string }>()
</script>
