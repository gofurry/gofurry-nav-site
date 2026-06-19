<template>
  <div class="min-h-full flex flex-col w-full bg-orange-50">
    <div
        class="top-0 w-full flex items-center justify-between p-4 bg-orange-100 backdrop-blur-sm shadow-sm"
    >
      <h2 class="h-4 text-lg font-semibold text-gray-800 flex items-center gap-2">
        {{ t('game.lottery.home.title') }}
      </h2>
    </div>

    <div class="flex flex-col px-6 py-10 max-w-6xl mx-auto">

      <div v-if="loading" class="text-center py-10 text-gray-400">
        {{ t('common.loading') }}
      </div>

      <div v-else>

        <section class="mb-16">
          <h3 class="text-xl font-bold text-gray-800 mb-4">
            {{ t('game.lottery.home.activePool') }}
          </h3>

          <div v-if="!activeList.length" class="text-gray-400">
            {{ t('game.lottery.home.noActiveLottery') }}
          </div>

          <div class="grid md:grid-cols-2 gap-6">
            <div
                v-for="item in activeList"
                :key="item.lottery.id"
                @click="openLottery(item)"
                class="bg-orange-100 border-2 border-orange-200 rounded-xl
                 hover:border-[#E49C69] px-4 py-2 transition duration-300"
            >
              <div class="flex justify-between items-center">
                <div class="text-lg font-bold text-orange-800 mb-2">
                  {{ item.lottery.title }}
                </div>

                <div class="text-xs text-gray-400">{{ t('game.lottery.home.clickToJoin') }}</div>
              </div>

              <p class="text-gray-600 mb-4 line-clamp-2 h-12 overflow-hidden">
                {{ item.lottery.desc }}
              </p>

              <div class="text-sm text-gray-500 mb-2 flex">
                <div class="font-bold text-gray-600">{{ t('game.lottery.home.prize') }}:&nbsp;</div>
                <div class="truncate">{{ item.lottery.prize.title }} ({{ item.lottery.prize.platform }})</div>
              </div>

              <div class="flex justify-between items-center">
                <div class="text-sm text-gray-500 mb-2 flex">
                  <div class="font-bold text-gray-600">{{ t('game.lottery.home.prizeQuantity') }}:&nbsp;</div>
                  <div>{{ item.lottery.prize.count }}</div>
                </div>

                <div class="text-sm text-gray-500 mb-3 flex">
                  <div class="font-bold text-gray-600">{{ t('game.lottery.home.participants') }}:&nbsp;</div>
                  <div>{{ item.count }}</div>
                </div>
              </div>

              <div class="mb-2">
                <div class="w-full bg-orange-50 h-2 rounded-full">
                  <div
                      class="bg-orange-500 h-2 rounded-full transition-all"
                      :style="{ width: calcProgress(item.lottery.start_time, item.lottery.end_time) + '%' }"
                  ></div>
                </div>
                <div class="flex justify-between items-center mt-0.5">
                  <div class="text-xs text-gray-400">
                    {{ formatDate(item.lottery.start_time) }}
                  </div>
                  <div class="text-xs text-gray-400">
                    {{ formatDate(item.lottery.end_time) }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section>
          <h2 class="text-xl font-semibold mb-6 text-gray-800">
            {{ t('game.lottery.home.winnerAnnouncement') }} - {{ t('common.total') }} {{ prizeCount }}
          </h2>

          <div v-if="!historyList.length" class="text-gray-400">
            {{ t('game.lottery.home.noHistory') }}
          </div>

          <div class="space-y-6">
            <div
                v-for="item in historyList"
                :key="item.name"
                class="bg-orange-100 border-2 border-orange-200 rounded-lg px-4 py-2 transition"
            >
              <div class="flex justify-between items-center mb-2">
                <h3 class="font-bold text-orange-800">
                  {{ item.name }}
                </h3>
                <span class="text-xs text-gray-400">
                  {{ t('game.lottery.home.deadline') }}:&nbsp;{{ formatDate(item.end_time) }}
                </span>
              </div>

              <p class="text-gray-600 mb-3">
                {{ item.desc }}
              </p>

              <div class="flex justify-between items-center mb-2">
                <div class="text-sm text-gray-500 flex">
                  <div class="font-bold text-gray-600">{{ t('game.lottery.home.prize') }}:&nbsp;</div>
                  <div>
                    {{ item.prize.title }} ({{ item.prize.platform }})
                    * {{ item.prize.count }}
                  </div>
                </div>

                <div class="text-sm text-gray-500 flex">
                  <div class="font-bold text-gray-600">{{ t('game.lottery.home.participants') }}:&nbsp;</div>
                  <div>
                    {{ item.count }}
                  </div>
                </div>
              </div>

              <div
                  v-if="item.winner.length"
                  class="flex flex-wrap gap-2"
              >
                <span
                    v-for="winner in item.winner"
                    :key="winner.email"
                    class="bg-[#343131] text-orange-100 text-xs px-3 py-1 rounded-full"
                >
                  {{ winner.name }} - {{ winner.email }}
                </span>
              </div>

              <div
                  v-else
                  class="text-xs text-gray-400 mt-2"
              >
                {{ t('game.lottery.home.noWinner') }}
              </div>
            </div>
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
import { onMounted, ref } from "vue"
import { getLottery } from "@/utils/api/game.ts"
import { i18n } from '@/main.ts'
import LotteryJoinModal from "@/components/game/lottery/LotteryJoinModal.vue"

import type {
  LotteryActiveModel,
  HistoryPrizeModel
} from "@/types/game.ts"

const { t } = i18n.global

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
