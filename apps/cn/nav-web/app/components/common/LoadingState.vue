<template>
  <div class="flex min-h-48 items-center justify-center text-sm text-slate-500">
    {{ label }}
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  label?: string
}>(), {
  label: 'Loading...'
})
</script>
