<template>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
  <ClientOnly>
    <PageScrollDock />
  </ClientOnly>
</template>

<script setup lang="ts">
import PageScrollDock from '~/components/common/PageScrollDock.vue'
</script>
