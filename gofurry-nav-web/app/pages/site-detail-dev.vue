<template>
  <SiteDetailVisualDevPage />
</template>

<script setup lang="ts">
import SiteDetailVisualDevPage from '@/components/site/SiteDetailVisualDevPage.vue'

definePageMeta({
  path: '/site/:id/:domain/dev',
})
</script>
