<template>
  <section class="game-stats-card">
    <header class="relative z-[1]">
      <div class="min-w-0">
        <div class="flex flex-wrap items-center gap-2">
          <h2 class="game-stats-card__title text-lg font-bold">
            {{ title }}
          </h2>
          <span class="game-stats-card__badge px-2 py-0.5 text-xs font-bold">
            {{ listToShow.length }}
          </span>
        </div>
        <p v-if="desc" class="game-stats-card__desc mt-1 truncate text-sm">
          {{ desc }}
        </p>
      </div>
    </header>

    <div v-if="topItem" class="game-stats-feature relative z-[1] mt-3 rounded-xl px-3 py-2">
      <div class="flex min-w-0 items-center gap-3">
        <img
          :src="topItem.header"
          class="stats-feature-cover h-12 w-24 shrink-0 rounded-lg object-cover"
          :alt="topItem.name"
        />
        <div class="min-w-0">
          <div class="game-stats-feature__title truncate text-sm font-bold">
            #1 {{ topItem.name }}
          </div>
          <div class="game-stats-feature__desc mt-0.5 line-clamp-1 text-xs">
            {{ topItem.desc }}
          </div>
          <div class="game-stats-feature__meta mt-0.5 text-xs">
            {{ t('game.panel.global') }} {{ formatPrice(topItem.global_price, true) }} · {{ t('game.panel.discount') }} {{ discountLabel(topItem) }}
          </div>
        </div>
        <div class="game-stats-feature__value shrink-0 text-right text-sm font-bold">
          {{ formatPrice(topItem.china_price, false) }}
        </div>
      </div>
    </div>

    <div class="game-stats-table relative z-[1] mt-4 overflow-hidden rounded-xl">
      <div class="game-stats-table-head grid grid-cols-[2rem_minmax(0,1fr)_5.4rem] items-center gap-2 px-2 py-2 text-xs font-bold sm:grid-cols-[2.2rem_minmax(0,1fr)_5.2rem_5.2rem_4.5rem]">
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
          class="game-stats-row grid min-h-[3.55rem] grid-cols-[2rem_minmax(0,1fr)_5.4rem] items-center gap-2 px-2 py-2 sm:grid-cols-[2.2rem_minmax(0,1fr)_5.2rem_5.2rem_4.5rem]"
          :class="index % 2 === 0 ? 'stats-row--warm' : 'stats-row--clear'"
          :style="{ '--activity': `${activityPercent(item)}%` }"
        >
          <div
            class="game-stats-rank grid h-7 w-7 place-items-center rounded-full text-xs font-extrabold"
            :class="{ 'game-stats-rank--top': index < 3 }"
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
              <div class="game-stats-row__title truncate text-sm font-bold">
                {{ item.name }}
              </div>
              <div class="game-stats-row__meta mt-0.5 text-xs sm:hidden">
                {{ discountLabel(item) }}
              </div>
            </div>
          </div>

          <div class="game-stats-row__value game-stats-row__value--secondary hidden text-right text-sm font-semibold sm:block">
            {{ formatPrice(item.global_price, true) }}
          </div>

          <div class="game-stats-row__value text-right text-sm font-bold">
            {{ formatPrice(item.china_price, false) }}
          </div>

          <div
            class="hidden text-right text-sm font-bold sm:block"
            :class="item.discount > 0 ? 'game-stats-row__discount--deal' : 'game-stats-row__discount--idle'"
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
