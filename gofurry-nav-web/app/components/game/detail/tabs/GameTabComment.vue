<template>
  <div class="game-detail-comments space-y-4 text-sm">

    <!-- 评论列表 -->
    <div v-if="visibleRemarks.length">
      <div
          v-for="(r, index) in visibleRemarks"
          :key="index"
          class="game-detail-comment mt-1 space-y-1 pb-3"
      >
        <!-- 顶部信息 -->
        <div class="game-detail-comment-meta flex flex-col text-xs">
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
          <img v-for="i in fullStars(r.score)" :key="'full-' + i + index" :src="starSvg" class="w-4 h-4" alt="" />
          <img v-if="hasHalfStar(r.score)" :src="starHalfSvg" class="w-4 h-4" alt="" />
          <img v-for="i in emptyStars(r.score)" :key="'empty-' + i + index" :src="starSvg" class="w-4 h-4 opacity-30" alt="" />
          <span class="game-detail-comment-score ml-2">{{ r.score.toFixed(1) }}</span>
        </div>

        <!-- 评论内容 -->
        <div class="game-detail-comment-body mt-1">{{ r.content }}</div>

        <!-- 分割线 -->
        <div
            v-if="index !== visibleRemarks.length - 1"
            class="game-detail-divider my-3 w-full"
        ></div>
      </div>

      <!-- 加载更多按钮 -->
      <div class="text-center mt-3">
        <button
            v-if="visibleCount < allRemarks.length"
            @click="loadMore"
            class="game-detail-load-more px-4 py-1"
        >
          {{t("common.loadMore")}}
        </button>
        <span v-else class="game-detail-empty text-sm">{{t("game.detail.allCommentsLoaded")}}</span>
      </div>
    </div>

    <!-- 无评论 -->
    <div v-else class="game-detail-empty py-6 text-center">
      {{t("game.panel.none")}}
    </div>
  </div>
</template>

<script setup lang="ts">
import type { RemarkResponse } from '@/types/game'
import { i18n } from '@/main'

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
