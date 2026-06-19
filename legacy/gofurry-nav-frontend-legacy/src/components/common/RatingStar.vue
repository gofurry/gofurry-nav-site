<template>
  <div class="flex items-center gap-1">
    <!-- 五颗星 -->
    <div class="flex items-center">
      <div
          v-for="n in 5"
          :key="n"
          class="relative w-5 h-5"
      >
        <!-- 灰色空星 -->
        <img
            :src="starFull"
            class="absolute inset-0 w-full h-full opacity-20"
        />

        <!-- 全星 -->
        <img
            v-if="n <= fullStars"
            :src="starFull"
            class="absolute inset-0 w-full h-full"
        />

        <!-- 半星 -->
        <img
            v-else-if="n === fullStars + 1 && hasHalfStar"
            :src="starHalf"
            class="absolute inset-0 w-full h-full"
        />
      </div>
    </div>

    <!-- 评分数 -->
    <span class="text-sm font-medium text-gray-700 ml-1">
      {{ score.toFixed(1) }}
    </span>

    <!-- 评价条数 -->
    <span class="text-xs text-gray-500 ml-1">
      ({{ count }})
    </span>
  </div>
</template>

<script setup lang="ts">
import starFull from "@/assets/svgs/star.svg";
import starHalf from "@/assets/svgs/star-half-alt.svg";

const props = defineProps<{
  score: number; // 0.0 - 5.0
  count: number; // 评分条数
}>();

const fullStars = Math.floor(props.score);
const hasHalfStar = props.score - fullStars >= 0.5;
</script>

<style scoped>
/* 你可根据需求自定义大小 */
</style>
