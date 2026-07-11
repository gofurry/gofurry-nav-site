<template>
  <div class="game-detail-news space-y-6 text-sm">

    <!-- 新闻列表 -->
    <div v-for="(item, index) in displayedNews" :key="index" class="game-detail-news-item space-y-2 pb-4">
      <!-- 标题 & 原文链接 -->
      <div class="flex items-start justify-between gap-4">
        <h3 class="game-detail-news-title flex-1 break-words text-lg font-bold">
          {{ item.headline }}
        </h3>

        <div class="game-detail-source-link flex shrink-0 items-center justify-center px-3 py-1 font-bold">
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
      <div class="game-detail-news-meta mb-1 flex justify-between text-xs">
        <span>{{t("game.detail.author")}}: {{ item.author }}</span>
        <span>{{t("game.detail.time")}}: {{ formatNewsTime(item.post_time) }}</span>
      </div>

      <!-- 新闻内容 -->
      <div v-html="item.content" class="game-detail-prose break-words"></div>



      <!-- 分隔线 -->
      <div
          v-if="index !== displayedNews.length - 1"
          class="game-detail-divider my-3 h-px w-full"
      ></div>

    </div>

    <!-- 加载更多 -->
    <div v-if="displayedNews.length < news.length" class="text-center mt-4">
      <button
          class="game-detail-load-more px-4 py-1"
          @click="loadMore"
      >
        {{t("common.loadMore")}}
      </button>
    </div>

    <!-- 已加载全部 -->
    <div v-else class="game-detail-empty mt-4 text-center">
      {{t("game.detail.allNewsLoaded")}}
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { NewsModel } from '@/types/game'
import { i18n } from '@/main'

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

function formatNewsTime(value?: string) {
  const raw = String(value || '').trim()
  if (!raw) {
    return ''
  }

  const isoMatch = raw.match(/^(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2}:\d{2})/)
  if (isoMatch) {
    return `${isoMatch[1]} ${isoMatch[2]}`
  }

  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) {
    return raw
  }

  const pad = (num: number) => String(num).padStart(2, '0')
  return [
    date.getFullYear(),
    pad(date.getMonth() + 1),
    pad(date.getDate()),
  ].join('-') + ` ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}
</script>
