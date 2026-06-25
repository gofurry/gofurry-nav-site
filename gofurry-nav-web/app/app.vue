<template>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
  <ClientOnly>
    <GlobalBackButton />
    <PageScrollDock />
  </ClientOnly>
</template>

<script setup lang="ts">
import GlobalBackButton from '~/components/common/GlobalBackButton.vue'
import PageScrollDock from '~/components/common/PageScrollDock.vue'

const { locale } = useI18n()
const defaultSeo = computed(() => {
  if (locale.value === 'en') {
    return {
      title: 'GoFurry Navigation - Discover furry culture resources and communities',
      description: 'GoFurry is a bilingual Furry navigation hub for communities, art, fiction, games, tools, and site monitoring.',
    }
  }

  return {
    title: 'GoFurry 兽人控导航站 - 发现兽人文化相关资源与社区',
    description: 'GoFurry 是面向兽人文化爱好者的双语导航站，收录 Furry 社区、艺术、小说、游戏、工具与站点监测资源。',
  }
})

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

useSeoMeta({
  title: () => defaultSeo.value.title,
  description: () => defaultSeo.value.description,
  ogTitle: () => defaultSeo.value.title,
  ogDescription: () => defaultSeo.value.description,
  twitterTitle: () => defaultSeo.value.title,
  twitterDescription: () => defaultSeo.value.description,
})
</script>
