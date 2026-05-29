<template>
  <NuxtPage v-if="hasChildRoute" />
  <SiteDetailRoutePage v-else />
</template>

<script setup lang="ts">
import SiteDetailRoutePage from '@/components/site/SiteDetailRoutePage.vue'

const route = useRoute()
const hasChildRoute = computed(() => route.path.replace(/\/$/, '').split('/').length > 3)
</script>
