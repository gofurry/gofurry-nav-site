<template>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
  <ClientOnly v-if="showGlobalScrollDock">
    <PageScrollDock />
  </ClientOnly>
</template>

<script setup lang="ts">
const route = useRoute()
const normalizedPath = computed(() => route.path.replace(/^\/(zh|en)(?=\/|$)/, '') || '/')
const showGlobalScrollDock = computed(() => normalizedPath.value !== '/archive')

import PageScrollDock from '~/components/common/PageScrollDock.vue'
</script>
