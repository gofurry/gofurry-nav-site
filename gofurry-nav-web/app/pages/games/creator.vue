<template>
  <main
      class="flex-1 bg-[#f2e3d0]"
      :style="{
      backgroundImage: `url(${bgGrid})`,
      backgroundRepeat: 'repeat'
    }"
  >
    <h1 class="sr-only">{{ creatorPageSeo.heading }}</h1>
    <CreatorGallery />
  </main>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import CreatorGallery from '@/components/game/creators/CreatorGallery.vue'
import bgGrid from '@/assets/pngs/bg-grid.png'

const { locale } = useI18n()
const creatorPageSeo = computed(() => locale.value === 'en'
  ? {
      heading: 'GoFurry Game Creators',
      title: 'GoFurry Game Creators - Curators, developers, publishers, and community contributors',
      description: 'Browse furry game creators and community contributors on GoFurry, including curators, bloggers, developers, publishers, translators, and content creators connected to game discovery.'
    }
  : {
      heading: 'GoFurry 兽人游戏创作者',
      title: 'GoFurry 兽人游戏创作者 - 策展、开发者、发行与社区贡献者',
      description: '在 GoFurry 浏览兽人游戏相关创作者与社区贡献者，包括策展人、博主、开发者、发行商、翻译者和内容创作者，继续发现相关游戏资源。'
    }
)

useSeoMeta({
  title: () => creatorPageSeo.value.title,
  description: () => creatorPageSeo.value.description,
  ogTitle: () => creatorPageSeo.value.title,
  ogDescription: () => creatorPageSeo.value.description,
})
</script>
