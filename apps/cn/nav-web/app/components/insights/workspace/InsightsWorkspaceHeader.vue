<template>
  <header class="insights-workspace-header">
    <p class="insights-workspace-eyebrow">{{ eyebrow }}</p>
    <h1>{{ title }}</h1>
    <p class="insights-workspace-description">{{ description }}</p>
  </header>
</template>

<script setup lang="ts">
defineProps<{ eyebrow: string; title: string; description: string }>()
</script>
