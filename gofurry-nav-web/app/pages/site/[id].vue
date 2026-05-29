<template>
  <SiteV2DetailDevPage v-if="isDevRoute" />
  <SiteDetailRoutePage v-else />
</template>

<script setup lang="ts">
import SiteDetailRoutePage from '@/components/site/SiteDetailRoutePage.vue'
import SiteV2DetailDevPage from '@/components/site/SiteV2DetailDevPage.vue'

const route = useRoute()
const isDevRoute = computed(() => route.path.replace(/\/$/, '').endsWith('/dev'))
</script>
