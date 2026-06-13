<template>
  <div class="rating-star" :aria-label="`${scoreLabel} (${count})`">
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

<style scoped>
.rating-star {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  min-width: 0;
}

.rating-star__icons {
  display: inline-flex;
  align-items: center;
  gap: 0.05rem;
}

.rating-star__item {
  position: relative;
  display: inline-grid;
  width: 1.05rem;
  height: 1.05rem;
  place-items: center;
  font-size: 1.05rem;
  line-height: 1;
}

.rating-star__empty,
.rating-star__fill {
  position: absolute;
  inset: 0;
  overflow: hidden;
  font-family: Arial, sans-serif;
  line-height: 1.05rem;
}

.rating-star__empty {
  color: rgba(154, 52, 18, 0.18);
  -webkit-text-stroke: 0.6px rgba(154, 52, 18, 0.18);
}

.rating-star__fill {
  color: #f59e0b;
  text-shadow: none;
  white-space: nowrap;
}

.rating-star__score {
  color: #334155;
  font-size: 0.84rem;
  font-weight: 650;
}

.rating-star__count {
  color: rgba(71, 85, 105, 0.68);
  font-size: 0.72rem;
}

:global(.dark) .rating-star__empty {
  color: rgba(148, 163, 184, 0.30);
  -webkit-text-stroke: 0;
}

:global(.dark) .rating-star__fill {
  color: #f59e0b;
  text-shadow: none;
}

:global(.dark) .rating-star__score {
  color: rgba(226, 232, 240, 0.92);
}

:global(.dark) .rating-star__count {
  color: rgba(203, 213, 225, 0.82);
}
</style>
