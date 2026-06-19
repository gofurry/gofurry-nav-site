<template>
  <section class="game-detail-main space-y-4">

    <!-- 顶部信息 -->
    <GameDetailHeader
        :game="game"
        :remark="remark"
    />

    <!-- Tabs -->
    <div class="game-detail-tabs">

      <!-- Tab Header -->
      <div class="game-detail-tab-list flex overflow-x-auto scrollbar-hide">
        <div
            v-for="tab in tabs"
            :key="tab.key"
            @click="activeTab = tab.key"
            class="game-detail-tab flex-shrink-0"
            :class="[
              'px-4 py-3 text-sm cursor-pointer select-none whitespace-nowrap',
              tab.mobileOnly ? 'xl:hidden' : '',
              activeTab === tab.key
                ? 'game-detail-tab--active'
                : 'game-detail-tab--idle'
            ]"
        >
          {{ tab.label }}
        </div>
      </div>

      <!-- Tab Content -->
      <div class="game-detail-tab-panel p-5 text-sm">

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
            :game-id="gameId"
            :remark="remark"
        />

        <div
            v-else-if="activeTab === 'similar' && !isDesktop"
            class="xl:hidden"
        >
          <GameSidebarSimilar :recommend="recommend" />
        </div>

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
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import type { GameBaseInfoResponse, RecommendedModel, RemarkResponse } from '@/types/game'

import GameDetailHeader from '@/components/game/detail/GameDetailHeader.vue'
import GameTabIntro from '@/components/game/detail/tabs/GameTabIntro.vue'
import GameTabGallery from '@/components/game/detail/tabs/GameTabGallery.vue'
import GameTabComment from '@/components/game/detail/tabs/GameTabComment.vue'
import GameTabNews from '@/components/game/detail/tabs/GameTabNews.vue'
import GameTabDetail from '@/components/game/detail/tabs/GameTabDetail.vue'
import GameSidebarSimilar from '@/components/game/detail/GameSidebarSimilar.vue'
import NsfwConfirmModal from '@/components/common/NsfwConfirmModal.vue'
import BlurWrapper from '@/components/common/BlurWrapper.vue'

import { i18n } from '@/main'
import { readMode, subscribeModeChange, writeMode } from '@/utils/modeStorage'

const { t } = i18n.global

const showNsfwModal = ref(false)
const mode = ref(readMode())
let stopModeSubscription: (() => void) | null = null
let desktopMediaQuery: MediaQueryList | null = null
let stopDesktopListener: (() => void) | null = null
const isDesktop = ref(false)

const props = defineProps<{
  game: GameBaseInfoResponse | null
  remark: RemarkResponse | null
  recommend: RecommendedModel[] | null
  gameId: string
}>()

const hasSimilarRecommend = computed(() => (props.recommend?.length ?? 0) > 0)

interface DetailTabItem {
  key: 'intro' | 'gallery' | 'comment' | 'news' | 'similar' | 'detail'
  label: string
  mobileOnly?: boolean
}

// Tabs 配置
const tabs = computed<DetailTabItem[]>(() => ([
  { key: 'intro', label: t('game.detail.introduction') },
  { key: 'gallery', label: t('game.detail.gallery') },
  { key: 'comment', label: t('game.detail.comments') + `(${props.remark?.total ?? 0})` },
  { key: 'news', label: t('game.detail.news') },
  ...(hasSimilarRecommend.value ? [{ key: 'similar', label: t('game.detail.similarGames'), mobileOnly: true } satisfies DetailTabItem] : []),
  { key: 'detail', label: t('game.detail.details') }
]))

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

  desktopMediaQuery = window.matchMedia('(min-width: 1280px)')
  const applyDesktopState = () => {
    isDesktop.value = desktopMediaQuery?.matches ?? false
  }
  applyDesktopState()

  const handler = () => {
    applyDesktopState()
  }
  desktopMediaQuery.addEventListener('change', handler)
  stopDesktopListener = () => desktopMediaQuery?.removeEventListener('change', handler)
})

onUnmounted(() => {
  stopModeSubscription?.()
  stopDesktopListener?.()
})

watch([isDesktop, activeTab], ([desktop, tabKey]) => {
  if (desktop && tabKey === 'similar') {
    activeTab.value = 'intro'
  }
})
</script>
