<template>
  <div
      class="relative isolate flex min-h-full w-full flex-col overflow-hidden"
  >
    <GoFurryGridBackground :fixed="false" palette="nav-content" />
    <div class="relative z-10 p-6 space-y-6">
      <h1 class="sr-only">{{ newsPageSeo.heading }}</h1>
      <MoreLatestNewsList
          :list="pageList"
      />

      <FixedPagination
          :current-page="currentPage"
          :total="total"
          @page-change="onPageChange"
      />

      <NewsSearchBall
          :news-list="list"
          @select="goNewsDetail"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useLangStore } from '@/store/langStore'
import { getMoreLatestGameNews } from '~/services/game'
import type { NewsBaseModel } from '@/types/game'

import MoreLatestNewsList from '@/components/game/news/MoreLatestNewsList.vue'
import FixedPagination from '@/components/game/news/FixedPagination.vue'
import GoFurryGridBackground from '@/components/common/GoFurryGridBackground.vue'
import NewsSearchBall from "@/components/game/news/NewsSearchBall.vue";

const langStore = useLangStore()

const lang = ref(langStore.lang)
const list = ref<NewsBaseModel[]>([])
const newsPageSeo = computed(() => lang.value === 'en'
  ? {
      heading: 'GoFurry Furry Game News',
      title: 'GoFurry Furry Game News - Latest updates from furry and anthro games',
      description: 'Read the latest furry and anthro game news collected by GoFurry, including update posts, release notes, creator activity, community signals, and related game discovery links.'
    }
  : {
      heading: 'GoFurry 兽人游戏资讯',
      title: 'GoFurry 兽人游戏资讯 - 兽人和拟人游戏最新动态',
      description: '查看 GoFurry 收集的兽人、拟人与相关题材游戏资讯，包括更新公告、发布动态、创作者消息、社区信号与相关游戏发现入口。'
    }
)

useSeoMeta({
  title: () => newsPageSeo.value.title,
  description: () => newsPageSeo.value.description,
  ogTitle: () => newsPageSeo.value.title,
  ogDescription: () => newsPageSeo.value.description,
})

const pageSize = 20
const totalPages = 5
const currentPage = ref(1)

const total = computed(() => list.value.length)

const pageList = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return list.value.slice(start, start + pageSize)
})

const goNewsDetail = (news :NewsBaseModel) => {
  window.open(news.url, '_blank', 'noopener')
}

const fetchData = async () => {
  const res = await getMoreLatestGameNews(lang.value)

  list.value = (res ?? [])
      .slice(0, 100)
      .sort((a, b) => {
        return new Date(b.post_time).getTime()
            - new Date(a.post_time).getTime()
      })

  currentPage.value = 1
}

const onPageChange = (page: number) => {
  if (page >= 1 && page <= totalPages) {
    currentPage.value = page
  }
}

watch(
    () => langStore.lang,
    async (v) => {
      lang.value = v
      await fetchData()
    }
)

onMounted(fetchData)
</script>
