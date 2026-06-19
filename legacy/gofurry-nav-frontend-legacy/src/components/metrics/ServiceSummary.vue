<template>
  <div class="rounded-xl flex flex-col gap-3 bg-white/60 backdrop-blur-xs">
    <h3 class="font-bold text-orange-800 px-4 pt-4">{{ title }}</h3>

    <div class="grid grid-cols-3 md:grid-cols-6 lg:grid-cols-3 xl:grid-cols-6 gap-4 pb-2">
      <div v-for="item in items" :key="item.label" class="flex flex-col items-center justify-center">
        <div class="text-xs text-gray-600 font-bold">{{ item.label }}</div>
        <div class="text-lg font-semibold text-gray-700">
          {{ item.value }}{{ item.suffix }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  items: {
    label: string
    value: string | number
    suffix?: string
  }[]
}>()
</script>
