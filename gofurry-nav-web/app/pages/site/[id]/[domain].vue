<template>
  <SiteV2DetailDevPage v-if="isDevRoute" />
  <SiteDetailRoutePage v-else-if="isV2Route" enable-domain-switcher domain-route-suffix="/v2" />
  <SiteDetailRoutePage v-else />
</template>

<script setup lang="ts">
import SiteDetailRoutePage from '@/components/site/SiteDetailRoutePage.vue'
import SiteV2DetailDevPage from '@/components/site/SiteV2DetailDevPage.vue'

const route = useRoute()
const normalizedPath = computed(() => route.path.replace(/\/$/, ''))
const isDevRoute = computed(() => normalizedPath.value.endsWith('/dev'))
const isV2Route = computed(() => normalizedPath.value.endsWith('/v2'))
</script>
