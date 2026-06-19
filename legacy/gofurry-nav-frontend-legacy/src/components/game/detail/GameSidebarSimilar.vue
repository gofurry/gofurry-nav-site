<template>
  <div class="bg-orange-50 rounded-2xl p-4 shadow">
    <h3 class="font-semibold mb-3">{{ t("game.detail.similarGames") }}</h3>

    <!-- 游戏列表 -->
    <div class="space-y-3">
      <div
          v-for="game in pagedGames"
          :key="game.id"
          class="flex gap-3 p-2 rounded-lg hover:bg-orange-200/50 cursor-pointer"
          @click="goGameDetail(game.id)"
      >
        <!-- 封面 -->
        <img
            :src="currentCover[game.appid!]"
            class="w-12 h-16 rounded-md object-cover flex-shrink-0"
            :alt="game.name"
            @error="loadFallbackCover(game.appid!)"
        />

        <!-- 游戏信息 -->
        <div class="flex-1 text-sm min-w-0">
          <div class="font-medium text-gray-800 truncate">{{ game.name }}</div>
          <div class="text-xs text-gray-500 line-clamp-2 break-words">{{ game.info }}</div>
          <div class="text-xs text-orange-600 mt-1">
            {{ t("game.detail.similarity") }}: {{ formatSimilarity(game.similarity) }}
          </div>
        </div>
      </div>

      <!-- 没有数据 -->
      <div v-if="!recommendList.length" class="text-gray-400 text-sm text-center py-4">
        {{ t("game.panel.none") }}
      </div>
    </div>

    <!-- 分页按钮 -->
    <div
        v-if="recommendList.length > PAGE_SIZE"
        class="flex justify-center gap-2 mt-3"
    >
      <button
          v-for="page in totalPages"
          :key="page"
          @click="currentPage = page"
          :class="[
          'px-3 py-1 rounded-md text-sm font-medium',
          currentPage === page
            ? 'bg-orange-400 text-white'
            : 'text-gray-700 hover:bg-orange-200/50'
        ]"
      >
        {{ page }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, reactive } from "vue"
import type { RecommendedModel } from "@/types/game"
import { useRouter } from 'vue-router'
import { i18n } from '@/main.ts'

const { t } = i18n.global
const router = useRouter()

const gamePrefix = import.meta.env.VITE_GAME_PREFIX_URL || ''
const steamPrefix = import.meta.env.VITE_STEAM_COVER_PREFIX_URL || ''

const props = defineProps<{
  recommend: RecommendedModel[] | null
}>()

const recommendList = computed(() => props.recommend ?? [])

const PAGE_SIZE = 4
const currentPage = ref(1)

// 总页数
const totalPages = computed(() => {
  const len = recommendList.value.length
  return len === 0 ? 0 : Math.ceil(len / PAGE_SIZE)
})

// 当前页展示的游戏
const pagedGames = computed(() => {
  const list = recommendList.value
  if (list.length === 0) return []
  const start = (currentPage.value - 1) * PAGE_SIZE
  return list.slice(start, start + PAGE_SIZE)
})

// 格式化相似度
function formatSimilarity(sim: number) {
  return `${(sim * 100).toFixed(1)}%`
}

const fallbackLoaded: Record<string, boolean> = {}
const currentCover = reactive<Record<string, string>>({})

function initCover(appid?: string) {
  if (!appid) return ''
  if (!currentCover[appid]) {
    fallbackLoaded[appid] = false
    // 默认先用 Steam 官方封面
    currentCover[appid] = steamPrefix+`${appid}/library_600x900.jpg`
  }
  return currentCover[appid]
}

function loadFallbackCover(appid: string) {
  if (!fallbackLoaded[appid]) {
    fallbackLoaded[appid] = true
    // Steam 封面失败后, 换成自己的 CDN
    currentCover[appid] = gamePrefix+`${appid}/library_600x900.jpg`
  } else if (fallbackLoaded[appid] &&
      currentCover[appid] && currentCover[appid].includes('library_600x900')) {
    // 如果 CDN library 也失败了, 再用 header.jpg
    currentCover[appid] = gamePrefix+`${appid}/header.jpg`
  }
}

// 当 recommendList 改变时初始化封面
watch(
    () => recommendList.value.map(g => g.appid),
    (list) => {
      list.forEach(appid => initCover(appid))
    },
    { immediate: true }
)

// 跳转
function goGameDetail(gameId: string) {
  router.push(`/games/${gameId}`)
}

watch(
    totalPages,
    (pages) => {
      if (pages === 0) {
        currentPage.value = 1
      } else if (currentPage.value > pages) {
        currentPage.value = pages
      }
    },
    { immediate: true }
)
</script>

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
