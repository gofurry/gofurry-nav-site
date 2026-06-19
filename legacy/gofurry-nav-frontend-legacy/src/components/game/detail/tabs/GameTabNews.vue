<template>
  <div class="space-y-6 text-gray-700 text-sm">

    <!-- 新闻列表 -->
    <div v-for="(item, index) in displayedNews" :key="index" class="space-y-2 pb-4">
      <!-- 标题 & 原文链接 -->
      <div class="flex items-start justify-between gap-4">
        <h3 class="text-lg font-bold text-gray-800 break-words flex-1">
          {{ item.headline }}
        </h3>

        <div class="flex shrink-0
             items-center justify-center
             px-3 py-1
             text-orange-900 font-bold
             rounded-md underline
             hover:bg-orange-200 hover:text-orange-800">
          <a
              v-if="item.url"
              :href="item.url"
              target="_blank"
          >
            {{t("game.detail.originalLink")}}
          </a>
        </div>
      </div>

      <!-- 作者 & 发布时间 -->
      <div class="flex justify-between text-gray-500 text-xs mb-1">
        <span>{{t("game.detail.author")}}: {{ item.author }}</span>
        <span>{{t("game.detail.time")}}: {{ item.post_time }}</span>
      </div>

      <!-- 新闻内容 -->
      <div v-html="item.content" class="break-words"></div>



      <!-- 分隔线 -->
      <div
          v-if="index !== displayedNews.length - 1"
          class="h-px w-full my-3
           bg-gradient-to-r
           from-transparent via-orange-200 to-transparent
           opacity-70"
      ></div>

    </div>

    <!-- 加载更多 -->
    <div v-if="displayedNews.length < news.length" class="text-center mt-4">
      <button
          class="px-4 py-1 bg-orange-100 hover:bg-orange-200 text-orange-700 rounded"
          @click="loadMore"
      >
        {{t("common.loadMore")}}
      </button>
    </div>

    <!-- 已加载全部 -->
    <div v-else class="text-center text-gray-400 mt-4">
      {{t("game.detail.allNewsLoaded")}}
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { NewsModel } from '@/types/game'
import { i18n } from '@/main.ts'

const { t } = i18n.global

const props = defineProps<{
  news: NewsModel[] | null
}>()

const news = props.news ?? []

// 加载更多逻辑
const pageSize = 5
const displayedCount = ref(Math.min(pageSize, news.length))

const displayedNews = computed(() => news.slice(0, displayedCount.value))

function loadMore() {
  displayedCount.value = Math.min(displayedCount.value + pageSize, news.length)
}
</script>
