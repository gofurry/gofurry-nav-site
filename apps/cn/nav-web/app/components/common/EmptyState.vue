<template>
  <div class="flex min-h-48 flex-col items-center justify-center rounded-lg border border-dashed border-orange-200 bg-white/70 p-8 text-center">
    <p class="text-base font-semibold text-slate-800">{{ title }}</p>
    <p v-if="description" class="mt-2 max-w-md text-sm leading-6 text-slate-500">{{ description }}</p>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>
