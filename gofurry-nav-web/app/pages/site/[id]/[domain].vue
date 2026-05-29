<template>
  <NuxtPage v-if="isDevRoute" />
  <SiteDetailRoutePage v-else />
</template>

<script setup lang="ts">
import SiteDetailRoutePage from '@/components/site/SiteDetailRoutePage.vue'

const route = useRoute()
const isDevRoute = computed(() => route.path.replace(/\/$/, '').endsWith('/dev'))
</script>
