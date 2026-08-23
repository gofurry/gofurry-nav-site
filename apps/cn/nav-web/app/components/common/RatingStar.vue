<template>
  <div class="gf-rating rating-star" :aria-label="`${scoreLabel} (${count})`">
    <div class="rating-star__icons" aria-hidden="true">
      <span
        v-for="n in 5"
        :key="n"
        class="rating-star__item"
      >
        <span class="rating-star__empty">★</span>
        <span
          class="rating-star__fill"
          :style="{ width: `${starFillPercent(n)}%` }"
        >
          ★
        </span>
      </span>
    </div>

    <span class="rating-star__score">
      {{ scoreLabel }}
    </span>

    <span class="rating-star__count">
      ({{ count }})
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  score: number
  count: number
}>()

const normalizedScore = computed(() => Math.min(5, Math.max(0, Number(props.score) || 0)))
const scoreLabel = computed(() => normalizedScore.value.toFixed(1))

function starFillPercent(index: number) {
  const remaining = normalizedScore.value - (index - 1)
  if (remaining >= 1) {
    return 100
  }
  if (remaining <= 0) {
    return 0
  }
  return remaining * 100
}
</script>
