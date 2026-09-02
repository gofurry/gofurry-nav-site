<template>
  <dl class="insights-stats">
    <div v-for="item in items" :key="item.label" class="insights-stat">
      <dt>{{ item.label }}</dt>
      <dd>{{ formatValue(item.value) }}</dd>
    </div>
  </dl>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

defineProps<{
  items: Array<{ label: string, value: number | null }>
}>()

const { locale } = useI18n()

function formatValue(value: number | null) {
  if (value === null || !Number.isFinite(value)) return '—'
  return new Intl.NumberFormat(locale.value === 'en' ? 'en-US' : 'zh-CN').format(value)
}
</script>
