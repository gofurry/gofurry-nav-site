<template>
  <div class="space-y-3">
    <h3 class="mb-1 text-sm font-bold text-stone-700/90">
      {{ titleText }}
    </h3>

    <div v-if="loading" class="py-4 text-center text-xs text-gray-400">
      {{ loadingText }}
    </div>

    <div v-else-if="reviews.length === 0" class="py-4 text-center text-xs text-gray-400">
      {{ emptyText }}
    </div>

    <div
      v-for="(item, index) in reviews"
      :key="index"
      class="latest-review-item"
    >
      <div class="flex w-[88px] shrink-0 flex-col items-center text-center">
        <img
          :src="item.game_cover"
          class="h-[52px] w-full rounded-md object-cover"
          :alt="item.game_name"
        />
        <p
          class="mt-1 w-full truncate text-xs font-semibold text-stone-800"
          :title="item.game_name"
        >
          {{ item.game_name }}
        </p>
      </div>

      <div class="flex min-w-0 flex-1 flex-col justify-between">
        <p
          class="line-clamp-2 text-sm leading-snug text-stone-700"
          :title="item.content"
        >
          {{ item.content }}
        </p>

        <div class="mt-2 space-y-0.5 text-xs text-stone-400">
          <div class="truncate">
            {{ regionLabel }}: {{ item.region }}
          </div>

          <div class="flex items-center justify-between gap-1">
            <span>{{ item.ip }}</span>
            <span class="truncate whitespace-nowrap">
              {{ formatTimeAgo(item.time) }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { getLatestReview } from '~/services/game'
import type { AnonymousReviewModel } from '~/types/game'

const props = defineProps<{
  initialReviews?: AnonymousReviewModel[]
}>()

const { locale } = useI18n()
const reviews = ref<AnonymousReviewModel[]>(props.initialReviews ?? [])
const loading = ref(false)

const isEnglish = computed(() => locale.value === 'en')
const titleText = computed(() => (isEnglish.value ? 'Latest Reviews' : '最新评论'))
const loadingText = computed(() => (isEnglish.value ? 'Loading...' : '加载中...'))
const emptyText = computed(() => (isEnglish.value ? 'No reviews yet' : '暂无评论'))
const regionLabel = computed(() => (isEnglish.value ? 'Region' : '评论地区'))

function formatTimeAgo(time: string): string {
  const now = Date.now()
  const past = new Date(time.replace(/-/g, '/')).getTime()
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
  } catch (error) {
    console.error('Failed to load latest reviews:', error)
    reviews.value = []
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  if (!reviews.value.length) {
    fetchLatestReviews()
  }
})
</script>

<style scoped>
.latest-review-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  border: 1px solid rgba(126, 92, 58, 0.17);
  border-radius: 0.82rem;
  background: rgba(255, 250, 242, 0.40);
  box-shadow: 0 5px 14px rgba(91, 62, 28, 0.035);
  padding: 0.68rem;
  transition: background-color 180ms ease, border-color 180ms ease;
}

.latest-review-item:hover {
  border-color: rgba(180, 96, 24, 0.34);
  background: rgba(255, 239, 213, 0.68);
}

:global(.dark) h3 {
  color: rgba(226, 232, 240, 0.86);
}

:global(.dark) .latest-review-item {
  border-color: rgba(226, 232, 240, 0.15);
  background: rgba(226, 232, 240, 0.065);
  box-shadow: none;
}

:global(.dark) .latest-review-item:hover {
  border-color: rgba(148, 163, 184, 0.36);
  background: rgba(148, 163, 184, 0.11);
}

:global(.dark) .latest-review-item .text-stone-800 {
  color: rgba(241, 245, 249, 0.86);
}

:global(.dark) .latest-review-item .text-stone-700 {
  color: rgba(203, 213, 225, 0.72);
}

:global(.dark) .latest-review-item .text-stone-400 {
  color: rgba(148, 163, 184, 0.78);
}

:global(.games-page--dark) .latest-review-item:hover {
  border-color: rgba(148, 163, 184, 0.36) !important;
  background: rgba(148, 163, 184, 0.11) !important;
}
</style>
