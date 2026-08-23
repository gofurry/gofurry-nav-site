<template>
  <div class="lottery-page relative isolate min-h-[calc(100svh-3.5rem)] overflow-hidden">
    <GoFurryGridBackground :fixed="false" palette="nav-content" />
    <div class="lottery-page__wash absolute inset-0 z-0" aria-hidden="true" />
    <div class="lottery-page__top-line absolute inset-x-0 top-0 h-px" aria-hidden="true" />

    <div class="relative z-10 mx-auto flex min-h-[calc(100svh-3.5rem)] w-full max-w-6xl flex-col px-5 py-10 sm:px-8 lg:py-14">
      <header class="lottery-hero grid gap-10 py-8 lg:grid-cols-[minmax(0,1fr)_18rem] lg:items-end">
        <div class="max-w-3xl">
          <p class="lottery-page__eyebrow mb-4 text-xs font-medium uppercase tracking-[0.28em]">
            gofurry games
          </p>
          <h1 class="lottery-page__title text-4xl font-semibold leading-tight sm:text-6xl">
            {{ t('game.lottery.home.title') }}
          </h1>
          <p class="lottery-page__subtitle mt-5 max-w-2xl text-sm leading-7 sm:text-base">
            {{ t('game.lottery.home.activePool') }} · {{ t('game.lottery.home.winnerAnnouncement') }}
          </p>
        </div>

        <div class="lottery-summary grid grid-cols-2 gap-px overflow-hidden rounded-lg backdrop-blur-md">
          <div class="lottery-summary__item p-4">
            <div class="lottery-summary__label text-[11px] uppercase tracking-[0.18em]">
              {{ t('game.lottery.home.activePool') }}
            </div>
            <div class="lottery-summary__value mt-2 text-3xl font-semibold">{{ activeList.length }}</div>
          </div>
          <div class="lottery-summary__item p-4">
            <div class="lottery-summary__label text-[11px] uppercase tracking-[0.18em]">
              {{ t('game.lottery.home.winnerAnnouncement') }}
            </div>
            <div class="lottery-summary__value mt-2 text-3xl font-semibold">{{ prizeCount }}</div>
          </div>
        </div>
      </header>

      <div v-if="loading" class="lottery-page__loading flex flex-1 items-center text-sm">
        {{ t('common.loading') }}
      </div>

      <div v-else class="space-y-16 pb-16">
        <section class="lottery-section">
          <div class="mb-5 flex items-end justify-between gap-4">
            <h2 class="lottery-section__title text-sm font-medium uppercase tracking-[0.22em]">
              {{ t('game.lottery.home.activePool') }}
            </h2>
            <div class="lottery-section__divider h-px flex-1" aria-hidden="true" />
          </div>

          <div v-if="!activeList.length" class="lottery-empty rounded-lg px-5 py-8 text-sm">
            {{ t('game.lottery.home.noActiveLottery') }}
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <button
              v-for="item in activeList"
              :key="item.lottery.id"
              type="button"
              class="lottery-pool group rounded-lg p-5 text-left backdrop-blur-xl transition duration-300"
              @click="openLottery(item)"
            >
              <div class="flex items-start justify-between gap-4">
                <h3 class="lottery-pool__title min-w-0 text-lg font-semibold leading-7">
                  {{ item.lottery.title }}
                </h3>
                <span class="lottery-pool__action shrink-0 rounded-full px-2.5 py-1 text-[11px]">
                  {{ t('game.lottery.home.clickToJoin') }}
                </span>
              </div>

              <p class="lottery-pool__desc mt-3 line-clamp-2 min-h-12 text-sm leading-6">
                {{ item.lottery.desc }}
              </p>

              <div class="lottery-meta mt-5 grid gap-3 text-sm">
                <div class="flex items-center justify-between gap-4">
                  <span class="lottery-meta__label">{{ t('game.lottery.home.prize') }}</span>
                  <span class="lottery-meta__value min-w-0 truncate text-right">
                    {{ item.lottery.prize.title }} · {{ item.lottery.prize.platform }}
                  </span>
                </div>
                <div class="flex items-center justify-between gap-4">
                  <span class="lottery-meta__label">{{ t('game.lottery.home.prizeQuantity') }}</span>
                  <span class="lottery-meta__value">{{ item.lottery.prize.count }}</span>
                </div>
                <div class="flex items-center justify-between gap-4">
                  <span class="lottery-meta__label">{{ t('game.lottery.home.participants') }}</span>
                  <span class="lottery-meta__value">{{ item.count }}</span>
                </div>
              </div>

              <div class="mt-5">
                <div class="lottery-progress h-1 overflow-hidden rounded-full">
                  <div
                    class="lottery-progress__bar h-full rounded-full transition-all duration-500"
                    :style="{ width: calcProgress(item.lottery.start_time, item.lottery.end_time) + '%' }"
                  />
                </div>
                <div class="lottery-progress__time mt-2 flex justify-between gap-3 text-[11px]">
                  <span>{{ formatDate(item.lottery.start_time) }}</span>
                  <span>{{ formatDate(item.lottery.end_time) }}</span>
                </div>
              </div>
            </button>
          </div>
        </section>

        <section class="lottery-section">
          <div class="mb-5 flex items-center justify-between gap-4">
            <h2 class="lottery-section__title text-sm font-medium uppercase tracking-[0.22em]">
              {{ t('game.lottery.home.winnerAnnouncement') }}
            </h2>
            <span class="lottery-section__count text-xs">{{ t('common.total') }} {{ prizeCount }}</span>
          </div>

          <div v-if="!historyList.length" class="lottery-empty rounded-lg px-5 py-8 text-sm">
            {{ t('game.lottery.home.noHistory') }}
          </div>

          <div class="lottery-history overflow-hidden rounded-lg backdrop-blur-md">
            <article
              v-for="item in historyList"
              :key="`${item.name}-${item.end_time}`"
              class="lottery-history__row grid gap-5 px-5 py-5 transition duration-200 lg:grid-cols-[minmax(0,1fr)_20rem]"
            >
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-x-3 gap-y-2">
                  <h3 class="lottery-history__title text-base font-semibold">
                    {{ item.name }}
                  </h3>
                  <span class="lottery-history__deadline text-xs">
                    {{ t('game.lottery.home.deadline') }} {{ formatDate(item.end_time) }}
                  </span>
                </div>

                <p class="lottery-history__desc mt-2 line-clamp-2 text-sm leading-6">
                  {{ item.desc }}
                </p>

                <div v-if="item.winner.length" class="mt-5 flex flex-wrap gap-2">
                  <span
                    v-for="winner in item.winner"
                    :key="winner.email"
                    class="lottery-winner max-w-full truncate rounded-full px-2.5 py-1 text-[11px]"
                  >
                    {{ winner.name }} · {{ winner.email }}
                  </span>
                </div>

                <div v-else class="lottery-history__empty mt-5 text-xs">
                  {{ t('game.lottery.home.noWinner') }}
                </div>
              </div>

              <div class="lottery-meta space-y-3 text-sm">
                <div class="flex items-center justify-between gap-3">
                  <span class="lottery-meta__label shrink-0 whitespace-nowrap">{{ t('game.lottery.home.prize') }}</span>
                  <span class="lottery-meta__value min-w-0 truncate text-right">
                    {{ item.prize.title }} · {{ item.prize.platform }} × {{ item.prize.count }}
                  </span>
                </div>
                <div class="flex items-center justify-between gap-3">
                  <span class="lottery-meta__label shrink-0 whitespace-nowrap">{{ t('game.lottery.home.participants') }}</span>
                  <span class="lottery-meta__value">{{ item.count }}</span>
                </div>
              </div>
            </article>
          </div>
        </section>
      </div>
    </div>
  </div>

  <LotteryJoinModal
    v-if="showModal && currentLottery"
    :lottery="currentLottery"
    @close="showModal = false"
  />
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue"
import { getLottery } from "@/utils/api/game"
import { i18n } from '@/main'
import GoFurryGridBackground from '@/components/common/GoFurryGridBackground.vue'
import LotteryJoinModal from "@/components/game/lottery/LotteryJoinModal.vue"

import type {
  LotteryActiveModel,
  HistoryPrizeModel
} from "@/types/game"

const { t } = i18n.global
const isZh = computed(() => i18n.global.locale.value === 'zh')
const pageSeo = computed(() => (
  isZh.value
    ? {
        title: 'GoFurry 兽人游戏抽奖 - 活动奖池与获奖记录',
        description: '查看 GoFurry 兽人游戏抽奖活动、当前可参与奖池、奖品信息、参与人数与历史获奖记录，发现与兽人游戏社区相关的福利动态。'
      }
    : {
        title: 'GoFurry Game Lottery - Active prize pools and winner records',
        description: 'View GoFurry game lottery events, active prize pools, prize details, participant counts, winner announcements, and community reward activity around furry games.'
      }
))

useSeoMeta({
  title: () => pageSeo.value.title,
  description: () => pageSeo.value.description,
  ogTitle: () => pageSeo.value.title,
  ogDescription: () => pageSeo.value.description,
})

const loading = ref(true)
const activeList = ref<LotteryActiveModel[]>([])
const historyList = ref<HistoryPrizeModel[]>([])
const prizeCount = ref(0)

const showModal = ref(false)
const currentLottery = ref<LotteryActiveModel | null>(null)

function openLottery(item: LotteryActiveModel) {
  currentLottery.value = item
  showModal.value = true
}

onMounted(async () => {
  try {
    const res = await getLottery()
    activeList.value = res.active || []
    historyList.value = res.history.prize || []
    prizeCount.value = res.history.prize_count
  } finally {
    loading.value = false
  }
})

function formatDate(time: string) {
  return time.replace(" ", "  ")
}

function calcProgress(start: string, end: string) {
  const now = Date.now()
  const s = new Date(start).getTime()
  const e = new Date(end).getTime()
  if (now <= s) return 0
  if (now >= e) return 100
  return Math.floor(((now - s) / (e - s)) * 100)
}
</script>
