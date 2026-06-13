<template>
  <section class="game-stats-card p-5">
    <header class="relative z-[1]">
      <div class="min-w-0">
        <div class="flex flex-wrap items-center gap-2">
          <h2 class="text-lg font-bold text-gray-950">
            {{ title }}
          </h2>
          <span class="rounded-full border border-orange-300/30 bg-orange-50/50 px-2 py-0.5 text-xs font-bold text-orange-800/80">
            {{ listToShow.length }}
          </span>
        </div>
        <p v-if="desc" class="mt-1 truncate text-sm text-gray-500">
          {{ desc }}
        </p>
      </div>
    </header>

    <div v-if="topItem" class="relative z-[1] mt-3 rounded-xl border border-white/35 bg-white/22 px-3 py-2">
      <div class="flex min-w-0 items-center gap-3">
        <img
          :src="topItem.header"
          class="stats-feature-cover h-12 w-24 shrink-0 rounded-lg object-cover"
          :alt="topItem.name"
        />
        <div class="min-w-0">
          <div class="truncate text-sm font-bold text-gray-950">
            #1 {{ topItem.name }}
          </div>
          <div class="mt-0.5 line-clamp-1 text-xs text-gray-500">
            {{ topItem.desc }}
          </div>
          <div class="mt-0.5 text-xs text-gray-500">
            {{ t('game.panel.global') }} {{ formatPrice(topItem.global_price, true) }} · {{ t('game.panel.discount') }} {{ discountLabel(topItem) }}
          </div>
        </div>
        <div class="shrink-0 text-right text-sm font-bold text-gray-950">
          {{ formatPrice(topItem.china_price, false) }}
        </div>
      </div>
    </div>

    <div class="relative z-[1] mt-4 overflow-hidden rounded-xl border border-white/35">
      <div class="grid grid-cols-[2rem_minmax(0,1fr)_5.4rem] items-center gap-2 px-2 py-2 text-xs font-bold text-gray-500 sm:grid-cols-[2.2rem_minmax(0,1fr)_5.2rem_5.2rem_4.5rem]">
        <span class="text-center">#</span>
        <span class="text-left">{{ t('common.game') }}</span>
        <span class="hidden text-right sm:block">{{ t('game.panel.global') }}</span>
        <span class="text-right">{{ t('game.panel.china') }}</span>
        <span class="hidden text-right sm:block">{{ t('game.panel.discount') }}</span>
      </div>

      <div>
        <article
          v-for="(item, index) in listToShow"
          :key="item.id"
          class="game-stats-row grid min-h-[3.55rem] grid-cols-[2rem_minmax(0,1fr)_5.4rem] items-center gap-2 border-t border-white/35 px-2 py-2 transition hover:bg-orange-200/45 sm:grid-cols-[2.2rem_minmax(0,1fr)_5.2rem_5.2rem_4.5rem]"
          :class="index % 2 === 0 ? 'stats-row--warm' : 'stats-row--clear'"
          :style="{ '--activity': `${activityPercent(item)}%` }"
        >
          <div
            class="grid h-7 w-7 place-items-center rounded-full bg-white/45 text-xs font-extrabold text-stone-600"
            :class="{ 'bg-orange-100/70 text-orange-800': index < 3 }"
          >
            {{ index + 1 }}
          </div>

          <div class="flex min-w-0 items-center gap-3">
            <img
              :src="item.header"
              class="stats-row-cover h-11 w-20 rounded-md object-cover"
              :alt="item.name"
            />
            <div class="min-w-0">
              <div class="truncate text-sm font-bold text-gray-950">
                {{ item.name }}
              </div>
              <div class="mt-0.5 text-xs text-gray-500 sm:hidden">
                {{ discountLabel(item) }}
              </div>
            </div>
          </div>

          <div class="hidden text-right text-sm font-semibold text-gray-700 sm:block">
            {{ formatPrice(item.global_price, true) }}
          </div>

          <div class="text-right text-sm font-bold text-gray-950">
            {{ formatPrice(item.china_price, false) }}
          </div>

          <div
            class="hidden text-right text-sm font-bold sm:block"
            :class="item.discount > 0 ? 'text-red-500' : 'text-gray-400'"
          >
            {{ discountLabel(item) }}
          </div>
        </article>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { PriceRecord } from '@/types/game'
import { i18n } from '@/main'

const { t } = i18n.global

const props = defineProps<{
  title: string
  desc?: string
  list: PriceRecord[]
}>()

const listToShow = computed(() =>
  props.list.slice(0, 15)
)

const topItem = computed(() => props.list[0] || null)

const maxPriceValue = computed(() => Math.max(1, ...listToShow.value.map(priceMetricValue)))

function priceMetricValue(item: PriceRecord) {
  if (item.discount > 0) {
    return item.discount
  }
  return Math.max(item.china_price || 0, item.global_price || 0)
}

function activityPercent(item: PriceRecord) {
  return Math.max(6, Math.min(100, Math.round((priceMetricValue(item) / maxPriceValue.value) * 100)))
}

function discountLabel(item: PriceRecord) {
  return item.discount > 0 ? `-${item.discount}%` : t('game.panel.none')
}

function formatPrice(value: number, isGlobal: boolean) {
  if (value === 0) {
    return t('game.panel.none')
  }
  const price = (value / 100).toFixed(2)
  return isGlobal ? `$${price}` : `¥${price}`
}
</script>

<style scoped>
.game-stats-card {
  container: game-stats-card / inline-size;
  position: relative;
  overflow: hidden;
  border: 1px solid rgba(126, 92, 58, 0.16);
  border-radius: 1.05rem;
  background: rgba(255, 250, 242, 0.20);
  box-shadow: none;
  backdrop-filter: blur(1px);
}

.game-stats-card::before {
  content: none;
}

.stats-feature-cover,
.stats-row-cover {
  display: none;
}

@container game-stats-card (min-width: 380px) {
  .stats-feature-cover {
    display: block;
  }
}

@container game-stats-card (min-width: 380px) {
  .stats-row-cover {
    display: block;
  }
}

.game-stats-row {
  isolation: isolate;
  position: relative;
  overflow: hidden;
}

.game-stats-row::before {
  content: "";
  position: absolute;
  inset: 0 auto 0 0;
  z-index: -1;
  width: var(--activity);
  background: linear-gradient(90deg, rgba(251, 146, 60, 0.14), rgba(251, 146, 60, 0.02));
}

.stats-row--warm {
  background: rgba(255, 237, 213, 0.30);
}

.stats-row--clear {
  background: rgba(255, 255, 255, 0.22);
}

:global(.dark) .game-stats-card {
  border-color: rgba(226, 232, 240, 0.12);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.052), rgba(226, 232, 240, 0.024)),
    rgba(226, 232, 240, 0.03);
}

:global(.dark) .game-stats-card h2,
:global(.dark) .game-stats-row .text-gray-950 {
  color: rgba(241, 245, 249, 0.88);
}

:global(.dark) .game-stats-card p,
:global(.dark) .game-stats-row .text-gray-500 {
  color: rgba(203, 213, 225, 0.62);
}

:global(.dark) .game-stats-card header span {
  border-color: rgba(148, 163, 184, 0.20);
  background: rgba(148, 163, 184, 0.08);
  color: rgba(203, 213, 225, 0.76);
}

:global(.dark) .game-stats-card .border-white\/35 {
  border-color: rgba(148, 163, 184, 0.13);
}

:global(.dark) .game-stats-row::before {
  background: linear-gradient(90deg, rgba(148, 163, 184, 0.20), rgba(100, 116, 139, 0.04));
}

:global(.dark) .stats-row--warm {
  background: rgba(30, 41, 59, 0.24);
}

:global(.dark) .stats-row--clear {
  background: rgba(15, 23, 42, 0.18);
}

:global(.dark) .game-stats-row:hover {
  background: rgba(51, 65, 85, 0.30) !important;
}

:global(.dark) .game-stats-row .bg-white\/45 {
  background: rgba(255, 255, 255, 0.10);
  color: rgba(226, 232, 240, 0.88);
}

:global(.dark) .game-stats-row .bg-orange-100\/70 {
  background: rgba(148, 163, 184, 0.18) !important;
  color: rgba(226, 232, 240, 0.86) !important;
}

:global(.dark) .game-stats-row .text-gray-700 {
  color: rgba(226, 232, 240, 0.78);
}

:global(.dark) .game-stats-row .text-gray-400 {
  color: rgba(148, 163, 184, 0.72);
}

:global(.games-page--dark) .game-stats-row::before {
  background: linear-gradient(90deg, rgba(148, 163, 184, 0.22), rgba(100, 116, 139, 0.045)) !important;
}

:global(.games-page--dark) .game-stats-card header span {
  border-color: rgba(148, 163, 184, 0.20) !important;
  background: rgba(148, 163, 184, 0.08) !important;
  color: rgba(203, 213, 225, 0.76) !important;
}

:global(.games-page--dark) .game-stats-row .bg-orange-100\/70 {
  background: rgba(148, 163, 184, 0.18) !important;
  color: rgba(226, 232, 240, 0.86) !important;
}
</style>
