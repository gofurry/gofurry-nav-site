<template>
  <div class="mb-8 p-5">
    <div class="mb-4 flex flex-col items-start gap-3 sm:flex-row sm:items-center sm:justify-between">
      <h3 class="text-2xl font-bold text-gray-800">
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
              <img
                :src="item.cover"
                class="mb-2 h-32 w-full rounded-md object-cover"
                :alt="item.name"
                @click.stop="goGameDetail(item.id)"
              />

              <button
                class="review-float-button review-float-button--light"
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

            <p class="line-clamp-1 text-sm font-semibold text-gray-900">
              {{ item.name }}
            </p>

            <p class="mt-1 h-[2rem] overflow-hidden text-xs text-gray-600">
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
.game-card {
  position: relative;
  cursor: pointer;
  border-radius: 0.75rem;
  padding: 0.5rem;
  transition: background-color 220ms ease;
}

.game-card:hover,
.game-card:focus-within {
  background: rgba(251, 146, 60, 0.16);
}

.review-float-button {
  position: absolute;
  right: 0.55rem;
  top: 0.55rem;
  display: grid;
  width: 2.15rem;
  height: 2.15rem;
  place-items: center;
  border-radius: 999px;
  opacity: 0;
  transform: translateY(0.25rem) scale(0.96);
  transition: opacity 220ms ease, transform 220ms cubic-bezier(0.22, 1, 0.36, 1), background-color 180ms ease, color 180ms ease;
}

.game-card:hover .review-float-button,
.game-card:focus-within .review-float-button {
  opacity: 1;
  transform: translateY(0) scale(1);
}

.review-float-button--light {
  border: 1px solid rgba(255, 255, 255, 0.48);
  background: rgba(255, 250, 242, 0.84);
  color: rgba(124, 45, 18, 0.92);
  box-shadow: 0 8px 22px rgba(45, 28, 12, 0.14);
  backdrop-filter: blur(8px);
}

.review-float-button--light:hover {
  background: rgba(255, 244, 228, 0.96);
  color: rgba(99, 39, 15, 1);
}

.review-float-button--dark {
  border: 1px solid rgba(255, 255, 255, 0.18);
  background: rgba(28, 25, 23, 0.72);
  color: rgba(255, 247, 237, 0.94);
  box-shadow: 0 8px 22px rgba(0, 0, 0, 0.20);
  backdrop-filter: blur(8px);
}

.review-float-button--dark:hover {
  background: rgba(28, 25, 23, 0.86);
}

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

.game-card--spacer {
  background: transparent;
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

.game-group-pager {
  display: inline-flex;
  align-items: center;
  gap: 0.48rem;
}

.game-group-pager button {
  display: grid;
  width: 1.65rem;
  height: 1.65rem;
  place-items: center;
  border: 0;
  border-radius: 999px;
  background: transparent;
  color: rgba(154, 52, 18, 0.74);
  font-size: 1.1rem;
  line-height: 1;
  transition:
    background 500ms ease,
    color 500ms ease;
}

.game-group-pager button:hover {
  background: rgba(154, 52, 18, 0.12);
  color: rgba(124, 45, 18, 0.98);
}

.game-group-pager span {
  min-width: 2.25rem;
  color: rgba(87, 83, 78, 0.70);
  font-size: 0.76rem;
  font-weight: 650;
  text-align: center;
}

.game-group-more {
  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  gap: 0.24rem;
  border-bottom: 1px solid rgba(154, 52, 18, 0.24);
  color: rgba(124, 45, 18, 0.88);
  font-size: 0.9rem;
  font-weight: 650;
  line-height: 1.5;
  white-space: nowrap;
  transition:
    border-color 220ms ease,
    color 220ms ease;
}

.game-group-more:hover {
  border-color: rgba(154, 52, 18, 0.70);
  color: rgba(124, 45, 18, 1);
}

:global(.dark) h3 {
  color: rgba(241, 245, 249, 0.88);
}

:global(.dark) .game-card:hover,
:global(.dark) .game-card:focus-within {
  background: rgba(148, 163, 184, 0.08);
}

:global(.dark) .game-card p:first-of-type {
  color: rgba(241, 245, 249, 0.88);
}

:global(.dark) .game-card p:nth-of-type(2) {
  color: rgba(203, 213, 225, 0.66);
}

:global(.games-page--dark) .game-card:hover,
:global(.games-page--dark) .game-card:focus-within {
  background: rgba(148, 163, 184, 0.09);
}

:global(.games-page--dark) .game-card p:first-of-type {
  color: rgba(248, 250, 252, 0.96);
}

:global(.games-page--dark) .game-card p:nth-of-type(2) {
  color: rgba(226, 232, 240, 0.76);
}

:global(.dark) .game-group-pager button {
  color: rgba(180, 213, 226, 0.62);
}

:global(.dark) .game-group-pager button:hover {
  background: rgba(148, 163, 184, 0.10);
  color: rgba(226, 232, 240, 0.80);
}

:global(.dark) .game-group-pager span {
  color: rgba(203, 213, 225, 0.68);
}

:global(.dark) .game-group-more {
  border-color: rgba(148, 163, 184, 0.22);
  color: rgba(180, 213, 226, 0.68);
}

:global(.dark) .game-group-more:hover {
  border-color: rgba(148, 163, 184, 0.42);
  color: rgba(226, 232, 240, 0.84);
}

:global(.games-page--dark) .game-group-pager button {
  color: rgba(180, 213, 226, 0.62) !important;
}

:global(.games-page--dark) .game-group-pager button:hover {
  background: rgba(148, 163, 184, 0.12) !important;
  color: rgba(226, 232, 240, 0.80) !important;
}

:global(.games-page--dark) .game-group-more {
  border-color: rgba(148, 163, 184, 0.22) !important;
  color: rgba(180, 213, 226, 0.68) !important;
}

:global(.games-page--dark) .game-group-more:hover {
  border-color: rgba(148, 163, 184, 0.42) !important;
  color: rgba(226, 232, 240, 0.84) !important;
}

:global(.dark) .review-float-button--light {
  border-color: rgba(255, 255, 255, 0.14);
  background: rgba(15, 23, 42, 0.76);
  color: rgba(226, 232, 240, 0.86);
  box-shadow: 0 8px 22px rgba(0, 0, 0, 0.22);
}

:global(.dark) .review-float-button--light:hover {
  background: rgba(30, 41, 59, 0.92);
  color: rgba(241, 245, 249, 0.94);
}

:global(.games-page--dark) .review-float-button--light {
  border-color: rgba(255, 255, 255, 0.14);
  background: rgba(15, 23, 42, 0.76);
  color: rgba(226, 232, 240, 0.88);
  box-shadow: 0 8px 22px rgba(0, 0, 0, 0.22);
}

:global(.games-page--dark) .review-float-button--light:hover {
  background: rgba(30, 41, 59, 0.92);
  color: rgba(248, 250, 252, 0.96);
}
</style>
