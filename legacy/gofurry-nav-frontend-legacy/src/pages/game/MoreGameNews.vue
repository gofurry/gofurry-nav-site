<template>
  <div
      class="flex flex-col w-full min-h-full bg-[#f2e3d0]"
      :style="{
        backgroundImage: `url(${bgGrid})`,
        backgroundRepeat: 'repeat'
      }"
  >
    <div class="p-6 space-y-6">
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
import { getMoreLatestGameNews } from '@/utils/api/game'
import type { NewsBaseModel } from '@/types/game.ts'

import MoreLatestNewsList from '@/components/game/news/MoreLatestNewsList.vue'
import FixedPagination from '@/components/game/news/FixedPagination.vue'
import bgGrid from '@/assets/pngs/bg-grid.png'
import NewsSearchBall from "@/components/game/news/NewsSearchBall.vue";

const langStore = useLangStore()

const lang = ref(langStore.lang)
const list = ref<NewsBaseModel[]>([])

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
