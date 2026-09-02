<template>
  <NuxtPage v-if="hasChildRoute" />
  <SiteDetailPage v-else show-insights />
</template>

<script setup lang="ts">
import SiteDetailPage from '@/components/site/SiteDetailPage.vue'

const route = useRoute()
const hasChildRoute = computed(() => {
  const domain = Array.isArray(route.params.domain) ? route.params.domain[0] : route.params.domain
  return typeof domain === 'string' && domain.trim().length > 0
})
</script>
