<template>
  <details class="insight-workspace-disclosure">
    <summary>{{ title }}</summary>
    <div><slot /></div>
  </details>
</template>

<script setup lang="ts">
defineProps<{ title: string }>()
</script>
