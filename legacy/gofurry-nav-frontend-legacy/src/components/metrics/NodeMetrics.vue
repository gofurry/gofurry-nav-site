<template>
  <section class="rounded-xl flex flex-col gap-3 backdrop-blur-xs">
    <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-4 m2">
      <div
          v-for="item in nodeCards"
          :key="item.label"
          class="rounded-xl bg-white/60 p-4"
      >
        <div class="text-xs font-semibold text-orange-800">
          {{ item.label }}
        </div>
        <div class="mt-2 text-2xl font-bold text-gray-800">
          {{ item.value }}
          <span class="text-sm font-medium">{{ item.suffix }}</span>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
defineProps<{
  nodeCards: {
    label: string
    value: string | number
    suffix?: string
  }[]
}>()
</script>
