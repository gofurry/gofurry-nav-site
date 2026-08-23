<template>
  <div class="rounded-lg border border-red-200 bg-red-50 p-5 text-red-800">
    <p class="font-semibold">{{ title }}</p>
    <p v-if="description" class="mt-2 text-sm leading-6">{{ description }}</p>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>
