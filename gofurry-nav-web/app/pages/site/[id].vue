<template>
  <NuxtPage v-if="hasChildRoute" />
  <SiteDetailPage v-else />
</template>

<script setup lang="ts">
import SiteDetailPage from '@/components/site/SiteDetailPage.vue'

const route = useRoute()
const hasChildRoute = computed(() => route.path.replace(/\/$/, '').split('/').length > 3)
</script>
