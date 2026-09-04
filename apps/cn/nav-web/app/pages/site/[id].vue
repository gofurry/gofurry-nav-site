<template>
  <NuxtPage v-if="hasChildRoute" />
  <SiteDetailPage v-else :show-insights="!hasTargetQuery" />
</template>

<script setup lang="ts">
import SiteDetailPage from '@/components/site/SiteDetailPage.vue'

const route = useRoute()
const hasChildRoute = computed(() => {
  const domain = Array.isArray(route.params.domain) ? route.params.domain[0] : route.params.domain
  return typeof domain === 'string' && domain.trim().length > 0
})
const hasTargetQuery = computed(() => {
  const domain = Array.isArray(route.query.domain) ? route.query.domain[0] : route.query.domain
  return typeof domain === 'string' && domain.trim().length > 0
})
</script>
