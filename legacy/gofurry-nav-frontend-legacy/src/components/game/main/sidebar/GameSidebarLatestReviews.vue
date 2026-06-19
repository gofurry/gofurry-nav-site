<template>
  <div class="space-y-3">
    <h3 class="text-sm font-semibold text-gray-700 mb-1">
      最新用户评论
    </h3>

    <!-- loading -->
    <div
        v-if="loading"
        class="text-xs text-gray-400 py-4 text-center"
    >
      加载中…
    </div>

    <!-- empty -->
    <div
        v-else-if="reviews.length === 0"
        class="text-xs text-gray-400 py-4 text-center"
    >
      暂无评论
    </div>

    <!-- 评论列表 -->
    <div
        v-for="(item, index) in reviews"
        :key="index"
        class="flex items-center gap-3 p-3
             rounded-xl
             bg-orange-50 hover:bg-orange-200/50
             transition"
    >
      <!-- 封面和游戏名 -->
      <div
          class="w-[88px] shrink-0
               flex flex-col items-center text-center"
      >
        <img
            :src="item.game_cover"
            class="w-full h-[52px] object-cover rounded-md"
            :alt="item.game_name"
        />
        <p
            class="mt-1 text-xs font-semibold text-gray-800 truncate w-full"
            :title="item.game_name"
        >
          {{ item.game_name }}
        </p>
      </div>

      <!-- 评论内容 -->
      <div class="flex-1 min-w-0 flex flex-col justify-between">
        <!-- 评论正文 -->
        <p
            class="text-sm text-gray-700 leading-snug line-clamp-2"
            :title="item.content"
        >
          {{ item.content }}
        </p>

        <!-- 评论元信息 -->
        <div class="mt-2 text-xs text-gray-400 space-y-0.5">
          <!-- 地区 -->
          <div class="truncate">
            评论地区: {{ item.region }}
          </div>

          <!-- IP + 时间 -->
          <div class="flex items-center justify-between gap-1">
            <span>{{ item.ip }}</span>
            <span class="whitespace-nowrap truncate">
              {{ formatTimeAgo(item.time) }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue"
import { getLatestReview } from "@/utils/api/game.ts"
import type { AnonymousReviewModel } from "@/types/game.ts"

const reviews = ref<AnonymousReviewModel[]>([])
const loading = ref(false)

/**
 * 时间 → xx days ago
 */
function formatTimeAgo(time: string): string {
  const now = Date.now()
  const past = new Date(time.replace(/-/g, "/")).getTime()
  const diff = Math.max(0, now - past)

  const minutes = Math.floor(diff / 60000)
  if (minutes < 60) return `${minutes} min ago`

  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours} hours ago`

  const days = Math.floor(hours / 24)
  return `${days} days ago`
}

async function fetchLatestReviews() {
  try {
    loading.value = true
    reviews.value = await getLatestReview()
  } catch (e) {
    console.error("获取最新评论失败", e)
    reviews.value = []
  } finally {
    loading.value = false
  }
}

onMounted(fetchLatestReviews)
</script>
