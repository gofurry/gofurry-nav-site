<template>
  <div class="lottery-page relative isolate min-h-[calc(100svh-3.5rem)] overflow-hidden bg-[#11100f] text-stone-100">
    <GoFurryGridBackground :fixed="false" palette="nav-content" />
    <div class="absolute inset-0 z-0 bg-[radial-gradient(circle_at_74%_18%,rgba(244,170,96,0.16),transparent_30%),linear-gradient(115deg,rgba(17,16,15,0.72)_0%,rgba(17,16,15,0.54)_54%,rgba(17,16,15,0.38)_100%)]" aria-hidden="true" />
    <div class="absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-orange-200/60 to-transparent" aria-hidden="true" />

    <div class="relative z-10 mx-auto flex min-h-[calc(100svh-3.5rem)] w-full max-w-6xl flex-col px-5 py-10 sm:px-8 lg:py-14">
      <header class="lottery-hero grid gap-10 py-8 lg:grid-cols-[minmax(0,1fr)_18rem] lg:items-end">
        <div class="max-w-3xl">
          <p class="mb-4 text-xs font-medium uppercase tracking-[0.28em] text-orange-200/70">
            gofurry games
          </p>
          <h1 class="text-4xl font-semibold leading-tight text-white sm:text-6xl">
            {{ t('game.lottery.home.title') }}
          </h1>
          <p class="mt-5 max-w-2xl text-sm leading-7 text-stone-300 sm:text-base">
            {{ t('game.lottery.home.activePool') }} · {{ t('game.lottery.home.winnerAnnouncement') }}
          </p>
        </div>

        <div class="grid grid-cols-2 gap-px overflow-hidden rounded-lg border border-white/10 bg-white/10 backdrop-blur-md">
          <div class="bg-black/[0.24] p-4">
            <div class="text-[11px] uppercase tracking-[0.18em] text-stone-400">
              {{ t('game.lottery.home.activePool') }}
            </div>
            <div class="mt-2 text-3xl font-semibold text-white">{{ activeList.length }}</div>
          </div>
          <div class="bg-black/[0.24] p-4">
            <div class="text-[11px] uppercase tracking-[0.18em] text-stone-400">
              {{ t('game.lottery.home.winnerAnnouncement') }}
            </div>
            <div class="mt-2 text-3xl font-semibold text-white">{{ prizeCount }}</div>
          </div>
        </div>
      </header>

      <div v-if="loading" class="flex flex-1 items-center text-sm text-stone-300">
        {{ t('common.loading') }}
      </div>

      <div v-else class="space-y-16 pb-16">
        <section>
          <div class="mb-5 flex items-end justify-between gap-4">
            <h2 class="text-sm font-medium uppercase tracking-[0.22em] text-stone-300">
              {{ t('game.lottery.home.activePool') }}
            </h2>
            <div class="h-px flex-1 bg-gradient-to-r from-white/14 to-transparent" aria-hidden="true" />
          </div>

          <div v-if="!activeList.length" class="rounded-lg border border-white/10 bg-white/[0.04] px-5 py-8 text-sm text-stone-400">
            {{ t('game.lottery.home.noActiveLottery') }}
          </div>

          <div class="grid gap-4 md:grid-cols-2">
            <button
              v-for="item in activeList"
              :key="item.lottery.id"
              type="button"
              class="lottery-pool group rounded-lg border border-white/10 bg-white/[0.055] p-5 text-left shadow-[0_18px_60px_rgba(0,0,0,0.18)] backdrop-blur-xl transition duration-300 hover:border-orange-200/40 hover:bg-white/[0.08]"
              @click="openLottery(item)"
            >
              <div class="flex items-start justify-between gap-4">
                <h3 class="min-w-0 text-lg font-semibold leading-7 text-white">
                  {{ item.lottery.title }}
                </h3>
                <span class="shrink-0 rounded-full border border-orange-200/20 px-2.5 py-1 text-[11px] text-orange-100/80">
                  {{ t('game.lottery.home.clickToJoin') }}
                </span>
              </div>

              <p class="mt-3 line-clamp-2 min-h-12 text-sm leading-6 text-stone-300">
                {{ item.lottery.desc }}
              </p>

              <div class="mt-5 grid gap-3 text-sm text-stone-300">
                <div class="flex items-center justify-between gap-4">
                  <span class="text-stone-500">{{ t('game.lottery.home.prize') }}</span>
                  <span class="min-w-0 truncate text-right text-stone-100">
                    {{ item.lottery.prize.title }} · {{ item.lottery.prize.platform }}
                  </span>
                </div>
                <div class="flex items-center justify-between gap-4">
                  <span class="text-stone-500">{{ t('game.lottery.home.prizeQuantity') }}</span>
                  <span class="text-stone-100">{{ item.lottery.prize.count }}</span>
                </div>
                <div class="flex items-center justify-between gap-4">
                  <span class="text-stone-500">{{ t('game.lottery.home.participants') }}</span>
                  <span class="text-stone-100">{{ item.count }}</span>
                </div>
              </div>

              <div class="mt-5">
                <div class="h-1 overflow-hidden rounded-full bg-white/10">
                  <div
                    class="h-full rounded-full bg-gradient-to-r from-orange-200 to-orange-400 transition-all duration-500"
                    :style="{ width: calcProgress(item.lottery.start_time, item.lottery.end_time) + '%' }"
                  />
                </div>
                <div class="mt-2 flex justify-between gap-3 text-[11px] text-stone-500">
                  <span>{{ formatDate(item.lottery.start_time) }}</span>
                  <span>{{ formatDate(item.lottery.end_time) }}</span>
                </div>
              </div>
            </button>
          </div>
        </section>

        <section>
          <div class="mb-5 flex items-center justify-between gap-4">
            <h2 class="text-sm font-medium uppercase tracking-[0.22em] text-stone-300">
              {{ t('game.lottery.home.winnerAnnouncement') }}
            </h2>
            <span class="text-xs text-stone-500">{{ t('common.total') }} {{ prizeCount }}</span>
          </div>

          <div v-if="!historyList.length" class="rounded-lg border border-white/10 bg-white/[0.04] px-5 py-8 text-sm text-stone-400">
            {{ t('game.lottery.home.noHistory') }}
          </div>

          <div class="divide-y divide-white/10 overflow-hidden rounded-lg border border-white/10 bg-black/[0.18] backdrop-blur-md">
            <article
              v-for="item in historyList"
              :key="`${item.name}-${item.end_time}`"
              class="grid gap-5 px-5 py-5 transition duration-200 hover:bg-white/[0.035] lg:grid-cols-[minmax(0,1fr)_20rem]"
            >
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-x-3 gap-y-2">
                  <h3 class="text-base font-semibold text-white">
                    {{ item.name }}
                  </h3>
                  <span class="text-xs text-stone-500">
                    {{ t('game.lottery.home.deadline') }} {{ formatDate(item.end_time) }}
                  </span>
                </div>

                <p class="mt-2 line-clamp-2 text-sm leading-6 text-stone-400">
                  {{ item.desc }}
                </p>

                <div v-if="item.winner.length" class="mt-5 flex flex-wrap gap-2">
                  <span
                    v-for="winner in item.winner"
                    :key="winner.email"
                    class="max-w-full truncate rounded-full border border-orange-200/16 bg-orange-200/10 px-2.5 py-1 text-[11px] text-orange-100/90"
                  >
                    {{ winner.name }} · {{ winner.email }}
                  </span>
                </div>

                <div v-else class="mt-5 text-xs text-stone-500">
                  {{ t('game.lottery.home.noWinner') }}
                </div>
              </div>

              <div class="space-y-3 text-sm text-stone-300">
                <div class="flex items-center justify-between gap-3">
                  <span class="shrink-0 whitespace-nowrap text-stone-500">{{ t('game.lottery.home.prize') }}</span>
                  <span class="min-w-0 truncate text-right">
                    {{ item.prize.title }} · {{ item.prize.platform }} × {{ item.prize.count }}
                  </span>
                </div>
                <div class="flex items-center justify-between gap-3">
                  <span class="shrink-0 whitespace-nowrap text-stone-500">{{ t('game.lottery.home.participants') }}</span>
                  <span>{{ item.count }}</span>
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

<style scoped>
.lottery-hero {
  animation: lottery-enter 520ms ease-out both;
}

.lottery-pool {
  animation: lottery-enter 420ms ease-out both;
}

@keyframes lottery-enter {
  from {
    opacity: 0;
    transform: translateY(12px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
