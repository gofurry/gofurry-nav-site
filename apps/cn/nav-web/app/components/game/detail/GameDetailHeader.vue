<template>
  <div class="game-detail-hero flex flex-col gap-5 p-5 sm:flex-row">
    <div class="shrink-0 flex justify-center sm:justify-start">
      <SteamAssetImage
        v-if="currentCover"
        :src="currentCover"
        class="game-detail-cover h-[240px] w-[180px] object-cover"
        :alt="game?.name || 'cover'"
        @error="loadNextCover"
      />
      <div
        v-else
        class="game-detail-cover game-detail-cover--empty flex h-[240px] w-[180px] items-center justify-center text-sm"
      >
        {{ t('game.panel.none') }}
      </div>
    </div>

    <div class="min-w-0 flex-1 flex flex-col gap-3">
      <div class="flex flex-col gap-1 sm:flex-row sm:items-end sm:justify-between">
        <h1 class="game-detail-title break-words text-2xl font-bold">
          {{ game?.name || t('game.panel.none') }}
        </h1>

        <div class="flex items-center gap-2 shrink-0">
          <span
            class="game-detail-metric flex shrink-0 items-center gap-1 text-xs"
          >
            <strong>{{ t('common.visits') }}: </strong>
            <div>{{ (game?.view_count ?? 0).toLocaleString() }}</div>
          </span>

          <span
            v-if="game?.online_count"
            class="game-detail-metric flex shrink-0 items-center gap-1 text-xs"
          >
            <span class="whitespace-nowrap">
              <strong>{{ t('game.detail.onlineNow') }}: </strong>
              <span>{{ game.online_count.toLocaleString() }}</span>
            </span>

            <span
              v-if="game.count_collect_time"
              class="game-detail-time whitespace-nowrap text-[11px]"
            >
              &nbsp;&nbsp;{{ formatTime(game.count_collect_time) }}
            </span>
          </span>
        </div>
      </div>

      <div class="flex flex-wrap gap-2">
        <span
          v-for="tag in displayTags"
          :key="tag.id"
          class="game-detail-tag relative cursor-default px-2 py-0.5 text-xs"
        >
          <span class="relative group">
            {{ tag.name }}

            <div
              v-if="tag.desc"
              class="game-detail-tag-tip pointer-events-none absolute left-1/2 top-full z-10 mt-1 -translate-x-1/2 whitespace-nowrap px-2 py-1 text-xs opacity-0 transition group-hover:opacity-100"
            >
              {{ tag.desc }}
            </div>
          </span>
        </span>

        <span
          v-if="tags.length > 8"
          class="game-detail-tag game-detail-tag--more cursor-pointer px-2 py-0.5 text-xs"
          @click="expanded = !expanded"
        >
          {{ expanded ? t('common.collapse') : t('common.expand') }}
        </span>
      </div>

      <div class="flex items-center gap-2 flex-wrap">
        <div class="flex items-center gap-1">
          <img
            v-for="i in fullStars"
            :key="'full-' + i"
            :src="starSvg"
            alt=""
            class="h-4 w-4"
          />
          <img
            v-if="hasHalfStar"
            :src="starHalfSvg"
            alt=""
            class="h-4 w-4"
          />
          <img
            v-for="i in emptyStars"
            :key="'empty-' + i"
            :src="starSvg"
            alt=""
            class="h-4 w-4 opacity-30"
          />
        </div>

        <span class="game-detail-score font-bold">{{ avgScore.toFixed(1) }}</span>
        <span class="game-detail-score-meta text-sm">
          ( {{ remark?.total ?? 0 }} {{ t('game.detail.commentCountSuffix') }} )
        </span>
      </div>

      <p class="game-detail-summary break-words text-sm leading-relaxed line-clamp-3">
        {{ game?.info || t('game.panel.none') }}
      </p>

      <div class="mt-auto flex items-center gap-3">
        <span class="game-detail-share-label text-sm">{{ t('game.detail.share') }}:</span>

        <button
          v-for="item in shareList"
          :key="item.name"
          class="game-detail-share-button flex h-8 w-8 items-center justify-center"
          @click="share(item.type)"
          :title="item.name"
        >
          <img :src="item.icon" :alt="item.name" class="share-icon h-4 w-4" />
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
import SteamAssetImage from '@/components/common/SteamAssetImage.vue'

import telegramIcon from '@/assets/icons/telegram.svg'
import twitterIcon from '@/assets/icons/twitter.svg'
import weiboIcon from '@/assets/icons/weibo.svg'
import { i18n } from '@/main'

const { t } = i18n.global

const props = defineProps<{
  game: GameBaseInfoResponse | null
  remark: RemarkResponse | null
}>()

const expanded = ref(false)

const shareList = [
  { name: 'Telegram', type: 'telegram', icon: telegramIcon },
  { name: 'Twitter', type: 'twitter', icon: twitterIcon },
  { name: '微博', type: 'weibo', icon: weiboIcon },
]

const coverUrls = computed(() => {
  const urls = [props.game?.cover].filter(Boolean) as string[]
  return [...new Set(urls)]
})
const currentCoverIndex = ref(0)
const coverFailed = ref(false)
const currentCover = computed(() =>
  coverFailed.value ? '' : coverUrls.value[currentCoverIndex.value] || ''
)

function loadNextCover() {
  if (currentCoverIndex.value < coverUrls.value.length - 1) {
    currentCoverIndex.value += 1
    return
  }

  coverFailed.value = true
}

const tags = computed(() => props.game?.tags ?? [])
const displayTags = computed(() => (expanded.value ? tags.value : tags.value.slice(0, 8)))

const avgScore = computed(() => props.remark?.avg_score ?? 0)
const fullStars = computed(() => Math.floor(avgScore.value))
const hasHalfStar = computed(() => avgScore.value - fullStars.value >= 0.45)
const emptyStars = computed(() => 5 - fullStars.value - (hasHalfStar.value ? 1 : 0))

function share(type: string) {
  const url = encodeURIComponent(location.href)
  const title = encodeURIComponent(props.game?.name || '')

  let shareUrl = ''
  switch (type) {
    case 'telegram':
      shareUrl = `https://t.me/share/url?url=${url}&text=${title}`
      break
    case 'twitter':
      shareUrl = `https://twitter.com/intent/tweet?url=${url}&text=${title}`
      break
    case 'weibo':
      shareUrl = `https://service.weibo.com/share/share.php?url=${url}&title=${title}`
      break
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
  return `${month}/${day} ${hour}:${minute}`
}

watch(
  () => props.game?.appid,
  () => {
    currentCoverIndex.value = 0
    coverFailed.value = false
  }
)

</script>

<style scoped>
.line-clamp-3 {
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
