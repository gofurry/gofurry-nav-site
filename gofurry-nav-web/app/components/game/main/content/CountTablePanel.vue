<template>
  <section class="game-stats-card">
    <header class="relative z-[1]">
      <div class="min-w-0">
        <div class="flex flex-wrap items-center gap-2">
          <h2 class="game-stats-card__title text-lg font-bold">
            {{ title }}
          </h2>
          <span class="game-stats-card__badge px-2 py-0.5 text-xs font-bold">
            {{ rankRange }}
          </span>
        </div>
        <p v-if="desc" class="game-stats-card__desc mt-1 truncate text-sm">
          {{ desc }}
        </p>
      </div>
    </header>

    <div v-if="topItem" class="game-stats-feature relative z-[1] mt-3 rounded-xl px-3 py-2">
      <div class="flex min-w-0 items-center gap-3">
        <SteamAssetImage
          :src="topItem.header"
          class="stats-feature-cover h-12 w-24 shrink-0 rounded-lg object-cover"
          :alt="topItem.name"
        />
        <div class="min-w-0">
          <div class="game-stats-feature__title truncate text-sm font-bold">
            #{{ rankStart }} {{ topItem.name }}
          </div>
          <div class="game-stats-feature__desc mt-0.5 line-clamp-1 text-xs">
            {{ topItem.desc }}
          </div>
          <div class="game-stats-feature__meta mt-0.5 text-xs">
            {{ t('game.panel.onlinePeak') }} {{ formatNumber(topItem.count_peak) }} · {{ formatTime(topItem.collect_time) }}
          </div>
        </div>
        <div class="game-stats-feature__value shrink-0 text-right text-sm font-bold">
          {{ formatNumber(topItem.count_recent) }}
        </div>
      </div>
    </div>
    <div
      v-else
      class="game-stats-feature game-stats-feature--placeholder relative z-[1] mt-3 rounded-xl px-3 py-2"
      aria-hidden="true"
    />

    <div class="game-stats-table relative z-[1] mt-4 overflow-hidden rounded-xl">
      <div class="game-stats-table-head grid grid-cols-[2rem_minmax(0,1fr)_5.2rem] items-center gap-2 px-2 py-2 text-xs font-bold sm:grid-cols-[2.2rem_minmax(0,1fr)_5.4rem_5.4rem_3.4rem]">
        <span class="text-center">#</span>
        <span class="text-left">{{ t('common.game') }}</span>
        <span class="text-right">{{ t('game.panel.recentOnline') }}</span>
        <span class="hidden text-right sm:block">{{ t('game.panel.onlinePeak') }}</span>
        <span class="hidden text-right sm:block">{{ t('common.time') }}</span>
      </div>

      <div>
        <article
          v-for="(item, index) in listToShow"
          :key="item.id"
          class="game-stats-row grid min-h-[3.55rem] grid-cols-[2rem_minmax(0,1fr)_5.2rem] items-center gap-2 px-2 py-2 sm:grid-cols-[2.2rem_minmax(0,1fr)_5.4rem_5.4rem_3.4rem]"
          :class="index % 2 === 0 ? 'stats-row--warm' : 'stats-row--clear'"
          :style="{ '--activity': `${activityPercent(item)}%` }"
        >
          <div
            class="game-stats-rank grid h-7 w-7 place-items-center rounded-full text-xs font-extrabold"
            :class="{ 'game-stats-rank--top': rankStart + index <= 3 }"
          >
            {{ rankStart + index }}
          </div>

          <div class="flex min-w-0 items-center gap-3">
            <SteamAssetImage
              :src="item.header"
              class="stats-row-cover h-11 w-20 rounded-md object-cover"
              :alt="item.name"
            />
            <div class="min-w-0">
              <div class="game-stats-row__title truncate text-sm font-bold">
                {{ item.name }}
              </div>
              <div class="game-stats-row__meta mt-0.5 text-xs sm:hidden">
                {{ formatTime(item.collect_time) }}
              </div>
            </div>
          </div>

          <div class="game-stats-row__value text-right text-sm font-bold">
            {{ formatNumber(item.count_recent) }}
          </div>

          <div class="game-stats-row__value game-stats-row__value--secondary hidden text-right text-sm font-semibold sm:block">
            {{ formatNumber(item.count_peak) }}
          </div>

          <div class="game-stats-row__muted hidden text-right text-xs sm:block">
            {{ formatTime(item.collect_time) }}
          </div>
        </article>
        <article
          v-for="index in placeholderRows"
          :key="`placeholder-${index}`"
          class="game-stats-row game-stats-row--placeholder grid min-h-[3.55rem] grid-cols-[2rem_minmax(0,1fr)_5.2rem] items-center gap-2 px-2 py-2 sm:grid-cols-[2.2rem_minmax(0,1fr)_5.4rem_5.4rem_3.4rem]"
          :class="(listToShow.length + index - 1) % 2 === 0 ? 'stats-row--warm' : 'stats-row--clear'"
          aria-hidden="true"
        />
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { i18n } from '@/main'
import SteamAssetImage from '@/components/common/SteamAssetImage.vue'
import type { TopPlayerCountRecord } from '@/types/game'

const { t } = i18n.global

const props = withDefaults(defineProps<{
  title: string
  desc?: string
  list: TopPlayerCountRecord[]
  rankStart?: number
}>(), {
  rankStart: 1,
})

const listToShow = computed(() => props.list.slice(0, 15))

const placeholderRows = computed(() => Math.max(0, 15 - listToShow.value.length))

const topItem = computed(() => props.list[0] || null)

const maxRecent = computed(() => Math.max(1, ...listToShow.value.map(item => item.count_recent)))

const rankRange = computed(() => {
  if (!listToShow.value.length) {
    return '#-'
  }
  const start = props.rankStart
  const end = props.rankStart + listToShow.value.length - 1
  return `#${start}-${end}`
})

function activityPercent(item: TopPlayerCountRecord) {
  return Math.max(6, Math.min(100, Math.round((item.count_recent / maxRecent.value) * 100)))
}

function formatNumber(value: number) {
  return new Intl.NumberFormat().format(value || 0)
}

function formatTime(value: number) {
  if (!value) {
    return '-'
  }
  const d = new Date(value * 1000)
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')

  return `${hh}:${mm}`
}
</script>
