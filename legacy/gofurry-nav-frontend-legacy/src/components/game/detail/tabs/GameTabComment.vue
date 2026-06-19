<template>
  <div class="space-y-4 text-sm text-gray-700">

    <!-- 评论列表 -->
    <div v-if="visibleRemarks.length">
      <div
          v-for="(r, index) in visibleRemarks"
          :key="index"
          class="space-y-1 pb-3 mt-1"
      >
        <!-- 顶部信息 -->
        <div class="flex flex-col text-gray-500 text-xs">
          <div class="flex justify-between">
            <span><strong>{{t("game.detail.commenter")}}:</strong> {{ r.name }}</span>
            <span><strong>{{t("game.detail.region")}}:</strong> {{ r.region }}</span>
          </div>
          <div class="flex justify-between">
            <span><strong>IP:</strong> {{ r.ip }}</span>
            <span><strong>{{t("game.detail.time")}}:</strong> {{ r.create_time }}</span>
          </div>
        </div>

        <!-- 星星评分 -->
        <div class="flex items-center gap-1 mt-1">
          <img v-for="i in fullStars(r.score)" :key="'full-' + i + index" :src="starSvg" class="w-4 h-4" />
          <img v-if="hasHalfStar(r.score)" :src="starHalfSvg" class="w-4 h-4" />
          <img v-for="i in emptyStars(r.score)" :key="'empty-' + i + index" :src="starSvg" class="w-4 h-4 opacity-30" />
          <span class="ml-2 text-gray-500">{{ r.score.toFixed(1) }}</span>
        </div>

        <!-- 评论内容 -->
        <div class="text-gray-800 mt-1">{{ r.content }}</div>

        <!-- 分割线 -->
        <div
            v-if="index !== visibleRemarks.length - 1"
            class="w-full my-3 border-t border-dashed border-orange-200/70"
        ></div>
      </div>

      <!-- 加载更多按钮 -->
      <div class="text-center mt-3">
        <button
            v-if="visibleCount < allRemarks.length"
            @click="loadMore"
            class="px-4 py-1 rounded bg-orange-100 text-orange-700 hover:bg-orange-200 transition"
        >
          {{t("common.loadMore")}}
        </button>
        <span v-else class="text-gray-400 text-sm">{{t("game.detail.allCommentsLoaded")}}</span>
      </div>
    </div>

    <!-- 无评论 -->
    <div v-else class="text-center text-gray-400 py-6">
      {{t("game.panel.none")}}
    </div>
  </div>
</template>

<script setup lang="ts">
import type { RemarkResponse } from '@/types/game'
import { i18n } from '@/main.ts'

const { t } = i18n.global

const props = defineProps<{
  remark: RemarkResponse | null
}>()

import { ref, computed, watch } from 'vue'

// 星星资源
import starSvg from '@/assets/svgs/star.svg'
import starHalfSvg from '@/assets/svgs/star-half-alt.svg'

// 星星计算函数
const fullStars = (score: number) => Math.floor(score)
const hasHalfStar = (score: number) => {
  const decimal = score - Math.floor(score)
  return decimal >= 0.45
}
const emptyStars = (score: number) => 5 - fullStars(score) - (hasHalfStar(score) ? 1 : 0)

// 分页加载
const PAGE_SIZE = 5
const visibleCount = ref(PAGE_SIZE)

// 所有评论
const allRemarks = computed(() => props.remark?.remarks ?? [])

// 当前可见评论
const visibleRemarks = computed(() => allRemarks.value.slice(0, visibleCount.value))

// 点击加载更多
function loadMore() {
  visibleCount.value = Math.min(visibleCount.value + PAGE_SIZE, allRemarks.value.length)
}

// 评论发生变化, 重置分页
watch(allRemarks, () => {
  visibleCount.value = PAGE_SIZE
})
</script>

