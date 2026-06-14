<template>
  <div class="space-y-4">

    <!-- Tab -->
    <div class="flex flex-col items-center justify-between gap-3 md:flex-row">

      <!-- 类型切换 -->
      <div class="stats-type-tabs inline-flex items-center gap-5">
        <div
            v-for="item in panelTypes"
            :key="item.key"
            class="stats-type-tab relative cursor-pointer px-1 pb-2 text-sm font-semibold"
            :class="activeType === item.key
              ? 'stats-type-tab--active'
              : 'stats-type-tab--idle'"
            @click="switchType(item.key)"
        >
          {{ item.label }}
        </div>
      </div>

      <!-- 组内切换 -->
      <div class="inline-flex items-center gap-2">
        <div
            v-for="(_, idx) in (activeType === 'count' ? visibleCountGroups : priceGroups)"
            :key="idx"
            class="stats-page-tab grid h-7 min-w-7 cursor-pointer place-items-center rounded-full px-2 text-sm font-semibold"
            :class="activeGroup === idx
              ? 'stats-page-tab--active'
              : 'stats-page-tab--idle'"
            @click="activeGroup = idx"
        >
          {{ idx + 1 }}
        </div>
      </div>

    </div>

    <!-- 面板内容 -->
    <div class="grid lg:grid-cols-2 gap-6">

      <!-- 在线人数 -->
      <CountTablePanel
          v-if="activeType === 'count'"
          v-for="panel in visibleCountGroups[activeGroup]"
          :key="panel.key"
          :title="panel.title"
          :desc="panel.desc"
          :list="panel.list"
          :rank-start="panel.rankStart"
      />

      <!-- 价格 -->
      <PriceTablePanel
          v-if="activeType === 'price'"
          v-for="panel in priceGroups[activeGroup]"
          :key="panel.key"
          :title="panel.title"
          :desc="panel.desc"
          :list="panel.list"
      />
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { i18n } from '@/main'
import type { PriceRecord, TopCountVo, BottomPriceVo } from '@/types/game'
import PriceTablePanel from './PriceTablePanel.vue'
import CountTablePanel from './CountTablePanel.vue'

const { t } = i18n.global

// props

const props = defineProps<{
  topPriceList: PriceRecord[]
  discountList: PriceRecord[]
  topCountList: TopCountVo
  bottomPriceList: BottomPriceVo
}>()

// Tab 状态

type PanelType = 'count' | 'price'

const panelTypes = [
  { key: 'count' as PanelType, label: t('game.panel.playerCount') },
  { key: 'price' as PanelType, label: t('game.panel.price') },
]

const activeType = ref<PanelType>('count')
const activeGroup = ref(0)

function switchType(type: PanelType) {
  activeType.value = type
  activeGroup.value = 0
}

// 在线人数面板

const countPanels = computed(() => [
  {
    key: 'count1',
    title: t('game.panel.playerCountTop15'),
    desc: t('game.panel.playerCountDesc1'),
    list: props.topCountList.one,
    rankStart: 1,
  },
  {
    key: 'count2',
    title: t('game.panel.playerCountTop30'),
    desc: t('game.panel.playerCountDesc2'),
    list: props.topCountList.two,
    rankStart: 16,
  },
  {
    key: 'count3',
    title: t('game.panel.playerCountTop45'),
    desc: t('game.panel.playerCountDesc3'),
    list: props.topCountList.three,
    rankStart: 31,
  },
  {
    key: 'count4',
    title: t('game.panel.playerCountTop60'),
    desc: t('game.panel.playerCountDesc4'),
    list: props.topCountList.four,
    rankStart: 46,
  },
])

// 价格面板

const pricePanels = computed(() => [
  {
    key: 'topPrice',
    title: t('game.panel.topPrice'),
    desc: t('game.panel.topPriceDesc'),
    list: props.topPriceList,
  },
  {
    key: 'discount',
    title: t('game.panel.discountTop'),
    desc: t('game.panel.discountTopDesc'),
    list: props.discountList,
  },
  {
    key: 'bottom1',
    title: t('game.panel.priceZone10'),
    desc: t('game.panel.priceZone10Desc'),
    list: props.bottomPriceList.one,
  },
  {
    key: 'bottom2',
    title: t('game.panel.priceZone15'),
    desc: t('game.panel.priceZone15Desc'),
    list: props.bottomPriceList.two,
  },
  {
    key: 'bottom3',
    title: t('game.panel.priceZone20'),
    desc: t('game.panel.priceZone20Desc'),
    list: props.bottomPriceList.three,
  },
  {
    key: 'bottom4',
    title: t('game.panel.priceZone25'),
    desc: t('game.panel.priceZone25Desc'),
    list: props.bottomPriceList.four,
  },
])

// 分组逻辑

function buildGroups<T>(list: T[]) {
  const groups: T[][] = []
  for (let i = 0; i < list.length; i += 2) {
    groups.push(list.slice(i, i + 2))
  }
  return groups
}

const priceGroups = computed(() => buildGroups(pricePanels.value))
const visibleCountGroups = computed(() => buildGroups(countPanels.value.filter(panel => panel.list.length > 0)))

watch([activeType, visibleCountGroups, priceGroups], () => {
  const groupLength = activeType.value === 'count'
    ? visibleCountGroups.value.length
    : priceGroups.value.length
  if (activeGroup.value >= groupLength) {
    activeGroup.value = Math.max(0, groupLength - 1)
  }
})
</script>
