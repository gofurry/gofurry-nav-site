<template>
  <div class="space-y-2">

    <!-- Tab -->
    <div class="flex flex-col md:flex-row items-center justify-between">

      <!-- 类型切换 -->
      <div class="inline-flex rounded-xl bg-orange-100 p-1">
        <div
            v-for="item in panelTypes"
            :key="item.key"
            class="px-4 py-2 rounded-lg text-sm transition"
            :class="activeType === item.key
              ? 'bg-orange-50 font-bold text-orange-800'
              : 'text-orange-500 hover:bg-orange-200'"
            @click="switchType(item.key)"
        >
          {{ item.label }}
        </div>
      </div>

      <!-- 组内切换 -->
      <div class="inline-flex rounded-xl bg-orange-100 p-1">
        <div
            v-for="(_, idx) in (activeType === 'count' ? countGroups : priceGroups)"
            :key="idx"
            class="px-4 py-2 rounded-lg text-sm transition"
            :class="activeGroup === idx
              ? 'bg-orange-50 font-bold text-orange-800'
              : 'text-orange-500 hover:bg-orange-200/50'"
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
          v-for="panel in countGroups[activeGroup]"
          :key="panel.key"
          :title="panel.title"
          :desc="panel.desc"
          :list="panel.list"
          :expanded="groupExpanded"
          @toggle="groupExpanded = !groupExpanded"
      />

      <!-- 价格 -->
      <PriceTablePanel
          v-if="activeType === 'price'"
          v-for="panel in priceGroups[activeGroup]"
          :key="panel.key"
          :title="panel.title"
          :desc="panel.desc"
          :list="panel.list"
          :expanded="groupExpanded"
          @toggle="groupExpanded = !groupExpanded"
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
const groupExpanded = ref(false)

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
  },
  {
    key: 'count2',
    title: t('game.panel.playerCountTop30'),
    desc: t('game.panel.playerCountDesc2'),
    list: props.topCountList.two,
  },
  {
    key: 'count3',
    title: t('game.panel.playerCountTop45'),
    desc: t('game.panel.playerCountDesc3'),
    list: props.topCountList.three,
  },
  {
    key: 'count4',
    title: t('game.panel.playerCountTop60'),
    desc: t('game.panel.playerCountDesc4'),
    list: props.topCountList.four,
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

const countGroups = computed(() => buildGroups(countPanels.value))
const priceGroups = computed(() => buildGroups(pricePanels.value))

watch([activeGroup, activeType], () => {
  groupExpanded.value = false
})
</script>