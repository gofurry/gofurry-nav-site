<template>
  <section data-site-health-strip :aria-label="t('siteDetail.health')" :aria-busy="pending" class="site-detail-health grid grid-cols-2 lg:grid-cols-6">
    <div v-for="item in presentation.health" :key="item.key" :data-site-health="item.key" :data-tone="item.tone" class="site-detail-health__item min-w-0">
      <div class="site-detail-label">{{ item.label }}</div>
      <div class="site-detail-health__value break-words">{{ item.value }}</div>
      <div v-if="item.detail" class="site-detail-note break-words">{{ item.detail }}</div>
    </div>
  </section>
</template>

<script setup lang="ts">
import type { SiteTargetPresentation } from '~/utils/siteTargetPresentation'
defineProps<{ presentation: SiteTargetPresentation; pending: boolean }>()
const { t } = useI18n()
</script>
