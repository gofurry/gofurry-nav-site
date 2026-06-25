<template>
  <div class="mb-8 p-5">
    <div class="mb-4 flex flex-col items-start gap-3 sm:flex-row sm:items-center sm:justify-between">
      <h3 class="game-group-title text-2xl font-bold">
        {{ group.title }}
      </h3>

      <div class="flex w-full items-center justify-between gap-2 sm:w-auto sm:justify-end">
        <div v-if="totalPages > 1" class="game-group-pager" :aria-label="`${group.title} pagination`">
          <button
            type="button"
            :aria-label="`${group.title} previous page`"
            @click="changePage(-1)"
          >
            ‹
          </button>
          <span>{{ currentPage + 1 }}/{{ totalPages }}</span>
          <button
            type="button"
            :aria-label="`${group.title} next page`"
            @click="changePage(1)"
          >
            ›
          </button>
        </div>

        <NuxtLink
          :to="localePath('/games/search')"
          class="game-group-more"
        >
          <span>{{ t('common.showMore') }}</span>
          <span aria-hidden="true">›</span>
        </NuxtLink>
      </div>
    </div>

    <div class="game-group-page-shell">
      <div class="game-group-page-grid game-group-page-spacer" aria-hidden="true">
        <div
          v-for="index in PAGE_SIZE"
          :key="index"
          class="game-card game-card--spacer"
        >
          <div class="mb-2 h-32 w-full rounded-md"></div>
          <p class="h-[1.25rem]"></p>
          <p class="mt-1 h-[2rem]"></p>
          <div class="mt-2 h-[1.25rem]"></div>
        </div>
      </div>

      <Transition :name="pageTransitionName">
        <div
          :key="currentPage"
          class="game-group-page-grid game-group-page-live"
        >
          <div
            v-for="item in visibleGames"
            :key="item.id"
            class="game-card group"
          >
            <div class="relative">
              <SteamAssetImage
                :src="item.cover"
                class="mb-2 h-32 w-full rounded-md object-cover"
                :alt="item.name"
                @click.stop="goGameDetail(item.id)"
              />

              <button
                class="review-float-button"
                type="button"
                :aria-label="`${item.name} review`"
                @click.stop="openReview(item)"
              >
                <svg
                  aria-hidden="true"
                  viewBox="0 0 24 24"
                  class="h-5 w-5"
                >
                  <path
                    fill="currentColor"
                    d="M9 22c-.55 0-1-.45-1-1v-3H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v12c0 1.11-.89 2-2 2h-6.1l-3.7 3.71c-.2.19-.45.29-.7.29H9m1-6v3.08L13.08 16H20V4H4v12h6m5.84-7.8l-1.01 1.01l-2.07-2.03l1.01-1.02c.2-.21.54-.22.78 0l1.29 1.25c.21.21.22.55 0 .79M8 11.91l4.17-4.19l2.07 2.08l-4.16 4.2H8v-2.09Z"
                  />
                </svg>
              </button>
            </div>

            <p class="game-card__title line-clamp-1 text-sm font-semibold">
              {{ item.name }}
            </p>

            <p class="game-card__desc mt-1 h-[2rem] overflow-hidden text-xs">
              {{ item.desc }}
            </p>

            <div class="mt-2">
              <RatingStar :score="item.score" :count="item.scoreCount" />
            </div>
          </div>
        </div>
      </Transition>
    </div>
    <GameReviewDialog
      :visible="!!reviewGame"
      :game-id="reviewGame?.id ?? ''"
      :game-name="reviewGame?.name ?? ''"
      @close="reviewGame = null"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import RatingStar from '@/components/common/RatingStar.vue'
import SteamAssetImage from '@/components/common/SteamAssetImage.vue'
import GameReviewDialog from '@/components/game/common/GameReviewDialog.vue'

const { t } = useI18n()
const router = useRouter()
const localePath = useLocalePath()

interface GameItem {
  id: string
  name: string
  cover: string
  desc: string
  score: number
  scoreCount: number
}

interface GameGroup {
  title: string
  games: GameItem[]
}

const props = defineProps<{
  group: GameGroup
}>()

defineEmits<{
  (e: 'more', group: GameGroup): void
}>()

function goGameDetail(id: string) {
  router.push(localePath(`/games/${id}`))
}

const PAGE_SIZE = 8
const currentPage = ref(0)
const pageDirection = ref<1 | -1>(1)
const reviewGame = ref<GameItem | null>(null)

const visibleGames = computed(() => {
  const start = currentPage.value * PAGE_SIZE
  return props.group.games.slice(start, start + PAGE_SIZE)
})

const totalPages = computed(() => Math.max(1, Math.ceil(props.group.games.length / PAGE_SIZE)))
const pageTransitionName = computed(() => pageDirection.value > 0 ? 'game-page-next' : 'game-page-prev')

watch([() => props.group.title, () => props.group.games.length], () => {
  currentPage.value = Math.min(currentPage.value, totalPages.value - 1)
})

function changePage(delta: number) {
  if (totalPages.value <= 1) {
    return
  }
  pageDirection.value = delta > 0 ? 1 : -1
  currentPage.value = (currentPage.value + delta + totalPages.value) % totalPages.value
}

function openReview(item: GameItem) {
  reviewGame.value = item
}
</script>

<style scoped>
.game-group-page-shell {
  position: relative;
  overflow: hidden;
}

.game-group-page-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 1rem;
}

.game-group-page-spacer {
  pointer-events: none;
  visibility: hidden;
}

.game-group-page-live {
  position: absolute;
  inset: 0;
  width: 100%;
}

.game-page-next-enter-active,
.game-page-next-leave-active,
.game-page-prev-enter-active,
.game-page-prev-leave-active {
  transition:
    opacity 360ms ease,
    transform 420ms cubic-bezier(0.22, 1, 0.36, 1);
  will-change: transform, opacity;
}

.game-page-next-leave-active,
.game-page-prev-leave-active {
  position: absolute;
  inset: 0;
  width: 100%;
}

.game-page-next-enter-from {
  opacity: 0;
  transform: translateX(2.75rem);
}

.game-page-next-leave-to {
  opacity: 0;
  transform: translateX(-2.75rem);
}

.game-page-prev-enter-from {
  opacity: 0;
  transform: translateX(-2.75rem);
}

.game-page-prev-leave-to {
  opacity: 0;
  transform: translateX(2.75rem);
}

@media (min-width: 640px) {
  .game-group-page-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 768px) {
  .game-group-page-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (min-width: 1024px) {
  .game-group-page-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

</style>
