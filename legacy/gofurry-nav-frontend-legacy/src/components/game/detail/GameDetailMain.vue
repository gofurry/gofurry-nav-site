<template>
  <section class="space-y-4">

    <!-- 顶部信息 -->
    <GameDetailHeader
        :game="game"
        :remark="remark"
    />

    <!-- Tabs -->
    <div class="bg-orange-50 rounded-2xl shadow">

      <!-- Tab Header -->
      <div class="flex border-b border-orange-100 overflow-x-auto scrollbar-hide">
        <div
            v-for="tab in tabs"
            :key="tab.key"
            @click="activeTab = tab.key"
            class="flex-shrink-0"
            :class="[
              'px-4 py-3 text-sm cursor-pointer select-none whitespace-nowrap',
              activeTab === tab.key
                ? 'text-orange-500 border-b-2 border-orange-400'
                : 'text-gray-500 hover:text-orange-400'
            ]"
        >
          {{ tab.label }}
        </div>
      </div>

      <!-- Tab Content -->
      <div class="p-5 text-sm text-gray-700">

        <!-- Intro -->
        <BlurWrapper
            v-if="activeTab === 'intro'"
            :enable="needBlur"
            :tip='t("common.modal.adultContent")'
            @unlock="openNsfwConfirm"
        >
          <GameTabIntro :game="game" />
        </BlurWrapper>

        <!-- Gallery -->
        <BlurWrapper
            v-else-if="activeTab === 'gallery'"
            :enable="needBlur"
            :tip='t("common.modal.galleryBlur")'
            @unlock="openNsfwConfirm"
        >
          <GameTabGallery
              :movies="game?.movies ?? null"
              :screenshots="game?.screenshots ?? null"
              :blocked="needBlur"
          />
        </BlurWrapper>

        <!-- Comment -->
        <GameTabComment
            v-else-if="activeTab === 'comment'"
            :remark="remark"
        />

        <!-- News -->
        <BlurWrapper
            v-else-if="activeTab === 'news'"
            :enable="needBlur"
            @unlock="openNsfwConfirm"
            :tip='t("common.modal.newsHidden")'
        >
          <GameTabNews :news="game?.news ?? []" />
        </BlurWrapper>

        <!-- Detail -->
        <GameTabDetail
            v-else-if="activeTab === 'detail'"
            :game="game"
        />
      </div>
    </div>

  </section>

  <NsfwConfirmModal
      :show="showNsfwModal"
      @cancel="showNsfwModal = false"
      @confirm="confirmNsfw"
  />
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import type { GameBaseInfoResponse, RemarkResponse } from '@/types/game'

import GameDetailHeader from '@/components/game/detail/GameDetailHeader.vue'
import GameTabIntro from '@/components/game/detail/tabs/GameTabIntro.vue'
import GameTabGallery from '@/components/game/detail/tabs/GameTabGallery.vue'
import GameTabComment from '@/components/game/detail/tabs/GameTabComment.vue'
import GameTabNews from '@/components/game/detail/tabs/GameTabNews.vue'
import GameTabDetail from '@/components/game/detail/tabs/GameTabDetail.vue'
import NsfwConfirmModal from '@/components/common/NsfwConfirmModal.vue'
import BlurWrapper from '@/components/common/BlurWrapper.vue'

import { i18n } from '@/main.ts'
import { readMode, subscribeModeChange, writeMode } from '@/utils/modeStorage'

const { t } = i18n.global

const showNsfwModal = ref(false)
const mode = ref(readMode())
let stopModeSubscription: (() => void) | null = null

const props = defineProps<{
  game: GameBaseInfoResponse | null
  remark: RemarkResponse | null
}>()

// Tabs 配置
const tabs = computed(() => ([
  { key: 'intro', label: t('game.detail.introduction') },
  { key: 'gallery', label: t('game.detail.gallery') },
  { key: 'comment', label: t('game.detail.comments') + `(${props.remark?.total ?? 0})` },
  { key: 'news', label: t('game.detail.news') },
  { key: 'detail', label: t('game.detail.details') }
] as const))

type TabKey = typeof tabs.value[number]['key']
const activeTab = ref<TabKey>('intro')

// ---------- mode 逻辑 ----------

// 从 localStorage 读取 mode
const nsfwEnabled = computed(() => {
  return mode.value === 'nsfw'
})

// 判断是否成人游戏
const isAdultGame = computed<boolean>(() => {
  return props.game?.tags?.some(tag => tag.id === '1014') ?? false
})

// 是否需要模糊处理
const needBlur = computed<boolean>(() => {
  return isAdultGame.value && !nsfwEnabled.value
})

// 打开 NSFW 确认弹窗
const openNsfwConfirm = () => {
  showNsfwModal.value = true
}

// 确认 NSFW 解锁
const confirmNsfw = () => {
  showNsfwModal.value = false

  // 保存到 localStorage
  writeMode('nsfw')
}
onMounted(() => {
  stopModeSubscription = subscribeModeChange(({ mode: nextMode }) => {
    mode.value = nextMode
  })
})

onUnmounted(() => {
  stopModeSubscription?.()
})
</script>
