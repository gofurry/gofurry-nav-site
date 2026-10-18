<template>
  <header class="insights-workspace-header" :class="visuallyHidden ? 'm-0' : 'mb-5 max-w-[900px] min-[768px]:mb-6'">
    <h1 :class="{ 'sr-only': visuallyHidden }">{{ title }}</h1>
  </header>
</template>

<script setup lang="ts">
defineProps<{ title: string; visuallyHidden?: boolean }>()
</script>
