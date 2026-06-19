<template>
  <div
      class="bg-orange-50 rounded-2xl p-5 shadow
           flex flex-col sm:flex-row gap-5"
  >

    <!-- 封面图 -->
    <div class="flex justify-center sm:justify-start shrink-0">
      <img
          :src="currentCover"
          class="w-[180px] h-[240px] rounded-xl object-cover"
          :alt="game?.name || 'cover'"
          @error="loadNextCover"
      />
    </div>

    <!-- 右侧信息 -->
    <div class="flex-1 min-w-0 flex flex-col gap-3">

      <!-- 标题 & 在线人数 -->
      <div class="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-1">
        <h1 class="text-2xl font-bold text-gray-800 break-words">
          {{ game?.name || t("game.panel.none") }}
        </h1>

        <div
            v-if="game?.online_count"
            class="text-xs text-orange-500 flex items-center gap-1 shrink-0"
        >
          <span class="whitespace-nowrap">
            <strong>{{t("game.detail.onlineNow")}}: </strong>
            <span>{{ game.online_count.toLocaleString() }}</span>
          </span>

          <span
              v-if="game.count_collect_time"
              class="text-[11px] text-gray-400 whitespace-nowrap"
          >
            &nbsp;&nbsp;{{ formatTime(game.count_collect_time) }}
          </span>
        </div>
      </div>

      <!-- 标签 -->
      <div class="flex flex-wrap gap-2">
        <span
            v-for="tag in displayTags"
            :key="tag.id"
            class="px-2 py-0.5 text-xs rounded-md bg-orange-100 text-orange-700 cursor-default relative"
        >
          <span class="relative group">
            {{ tag.name }}

            <div
                v-if="tag.desc"
                class="absolute z-10 left-1/2 -translate-x-1/2 top-full mt-1
                       bg-gray-800 text-white text-xs rounded px-2 py-1
                       opacity-0 group-hover:opacity-100 transition
                       whitespace-nowrap pointer-events-none"
            >
              {{ tag.desc }}
            </div>
          </span>
        </span>

        <span
            v-if="tags.length > 8"
            class="px-2 py-0.5 text-xs rounded-md cursor-pointer
             bg-orange-200 text-orange-700 hover:bg-orange-300"
            @click="expanded = !expanded"
        >
          {{ expanded ? t("common.collapse") : t("common.expand") }}
        </span>
      </div>

      <!-- 评分 -->
      <div class="flex items-center gap-2 flex-wrap">
        <div class="flex items-center gap-1">
          <img
              v-for="i in fullStars"
              :key="'full-' + i"
              :src="starSvg"
              alt="star"
              class="w-4 h-4"
          />
          <img
              v-if="hasHalfStar"
              :src="starHalfSvg"
              alt="half-star"
              class="w-4 h-4"
          />
          <img
              v-for="i in emptyStars"
              :key="'empty-' + i"
              :src="starSvg"
              alt="empty-star"
              class="w-4 h-4 opacity-30"
          />
        </div>

        <span class="text-orange-500 font-bold">{{ avgScore.toFixed(1) }}</span>
        <span class="text-sm text-gray-500">
          ( {{ remark?.total ?? 0 }} {{ t("game.detail.commentCountSuffix") }} )
        </span>
      </div>

      <!-- 简介 -->
      <p class="text-sm text-gray-700 leading-relaxed break-words line-clamp-3">
        {{ game?.info || t("game.panel.none") }}
      </p>

      <!-- 分享 -->
      <div class="flex items-center gap-3 mt-auto">
        <span class="text-sm text-gray-500">{{ t("game.detail.share") }}:</span>

        <button
            v-for="item in shareList"
            :key="item.name"
            class="w-8 h-8 flex items-center justify-center
                   rounded-full bg-orange-100 hover:bg-orange-200
                   transition"
            @click="share(item.type)"
            :title="item.name"
        >
          <img :src="item.icon" :alt="item.name" class="w-4 h-4 share-icon" />
        </button>
      </div>

    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { GameBaseInfoResponse, RemarkResponse } from '@/types/game'

import starSvg from '@/assets/svgs/star.svg'
import starHalfSvg from '@/assets/svgs/star-half-alt.svg'

import telegramIcon from '@/assets/icons/telegram.svg'
import twitterIcon from '@/assets/icons/twitter.svg'
import weiboIcon from '@/assets/icons/weibo.svg'
import { i18n } from '@/main.ts'

const { t } = i18n.global

const gamePrefix = import.meta.env.VITE_GAME_PREFIX_URL || ''
const steamPrefix = import.meta.env.VITE_STEAM_COVER_PREFIX_URL || ''

const props = defineProps<{
  game: GameBaseInfoResponse | null
  remark: RemarkResponse | null
}>()

const expanded = ref(false)

// 分享
const shareList = [
  { name: 'Telegram', type: 'telegram', icon: telegramIcon },
  { name: 'Twitter', type: 'twitter', icon: twitterIcon },
  { name: '微博', type: 'weibo', icon: weiboIcon }
]

// 封面兜底数组
const coverUrls = computed(() => {
  const appid = props.game?.appid
  if (!appid) return []
  return [
    steamPrefix+`${appid}/library_600x900.jpg`,
    gamePrefix+`${appid}/library_600x900.jpg`,
    gamePrefix+`${appid}/header.jpg`
  ]
})

const currentCoverIndex = ref(0)
const currentCover = ref(coverUrls.value[0])

function loadNextCover() {
  if (currentCoverIndex.value < coverUrls.value.length - 1) {
    currentCoverIndex.value++
    currentCover.value = coverUrls.value[currentCoverIndex.value]
  }
}

// 标签
const tags = computed(() => props.game?.tags ?? [])
const displayTags = computed(() =>
    expanded.value ? tags.value : tags.value.slice(0, 8)
)

// 评分逻辑
const avgScore = computed(() => props.remark?.avg_score ?? 0)
const fullStars = computed(() => Math.floor(avgScore.value))
const hasHalfStar = computed(() => (avgScore.value - fullStars.value) >= 0.45)
const emptyStars = computed(() => 5 - fullStars.value - (hasHalfStar.value ? 1 : 0))

function share(type: string) {
  const url = encodeURIComponent(location.href)
  const title = encodeURIComponent(props.game?.name || '')

  let shareUrl = ''
  switch (type) {
    case 'telegram': shareUrl = `https://t.me/share/url?url=${url}&text=${title}`; break
    case 'twitter': shareUrl = `https://twitter.com/intent/tweet?url=${url}&text=${title}`; break
    case 'weibo': shareUrl = `https://service.weibo.com/share/share.php?url=${url}&title=${title}`; break
  }
  window.open(shareUrl, '_blank')
}

function formatTime(time: string | number) {
  const date = new Date(time)
  if (isNaN(date.getTime())) return ''
  const month = date.getMonth() + 1
  const day = date.getDate()
  const hour = date.getHours().toString().padStart(2, '0')
  const minute = date.getMinutes().toString().padStart(2, '0')
  return `${month}月${day}日 ${hour}:${minute}`
}

// 当 appid 改变时重置封面
watch(
    () => props.game?.appid,
    () => {
      currentCoverIndex.value = 0
      currentCover.value = coverUrls.value[0]
    }
)
</script>

<style scoped>
.share-icon {
  filter:
      invert(42%)
      sepia(96%)
      saturate(1150%)
      hue-rotate(3deg)
      brightness(90%)
      contrast(95%);
}
.line-clamp-3 {
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
