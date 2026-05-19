<template>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
  <ClientOnly v-if="showGlobalScrollDock">
    <PageScrollDock />
  </ClientOnly>
</template>

<script setup lang="ts">
import PageScrollDock from '~/components/common/PageScrollDock.vue'

const route = useRoute()
const normalizedPath = computed(() => route.path.replace(/^\/(zh|en)(?=\/|$)/, '') || '/')
const showGlobalScrollDock = computed(() => normalizedPath.value !== '/archive')

const localeHead = useLocaleHead({
  dir: true,
  lang: true,
  seo: true,
})

useHead(() => ({
  htmlAttrs: {
    ...localeHead.value.htmlAttrs,
  },
  link: [
    ...(localeHead.value.link || []),
  ],
  meta: [
    ...(localeHead.value.meta || []),
  ],
}))
</script>
