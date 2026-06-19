<template>
  <div class="flex justify-center items-center gap-2 select-none">

    <span
        v-for="page in 5"
        :key="page"
        class="px-3 py-1 rounded-md text-sm cursor-pointer transition"
        :class="page === currentPage
          ? 'bg-orange-400 text-white'
          : 'bg-orange-50 hover:bg-orange-200'"
        @click="emit('page-change', page)"
    >
      {{ page }}
    </span>

    <span class="ml-2 text-xs text-gray-500">
      {{ t('common.total') }} {{ total }} {{ t('common.record') }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { i18n } from '@/main'

const { t } = i18n.global

defineProps<{
  currentPage: number
  total: number
}>()

const emit = defineEmits<{
  (e: 'page-change', page: number): void
}>()
</script>
