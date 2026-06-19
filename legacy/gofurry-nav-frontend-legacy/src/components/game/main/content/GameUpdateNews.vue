<template>
  <div class="mb-8 rounded-2xl bg-orange-50 p-5 backdrop-blur-md">
    <div class="mb-4 flex items-center justify-between">
      <h2 class="text-lg font-bold">{{ t('game.news.title') }}</h2>
      <div class="hidden text-sm text-gray-500 sm:block">{{ t('game.news.desc') }}</div>
    </div>

    <div class="overflow-x-auto">
      <div class="flex min-w-max gap-4">
        <div
            v-for="news in newsList"
            :key="news.id"
            ref="newsRefs"
            class="relative mb-1 w-36 flex-shrink-0 cursor-pointer rounded-lg bg-orange-100/60 p-3 transition hover:bg-orange-200/50 sm:w-72"
            @click="openUrl(news.url)"
        >
          <img
              :src="news.header"
              alt="cover"
              ref="coverRefs"
              class="mb-2 w-full rounded-lg object-cover"
              @mouseenter="onNewsMouseEnter(news, $event)"
              @mouseleave="onNewsMouseLeave"
          />

          <h3 class="truncate font-semibold">{{ htmlToPlainText(news.headline) }}</h3>

          <div class="mt-1 flex justify-between text-xs text-gray-500">
            <span>{{ formatTime(news.post_time) }}</span>
            <span class="hidden sm:block">{{ news.author }}</span>
          </div>

          <p class="mt-2 line-clamp-3 text-sm text-gray-700">
            {{ htmlToPlainText(news.content) }}
          </p>
        </div>
      </div>
    </div>

    <Teleport to="body">
      <Transition
          enter-active-class="transition duration-200 ease-out"
          enter-from-class="opacity-0 scale-95"
          enter-to-class="opacity-100 scale-100"
          leave-active-class="transition duration-150 ease-in"
          leave-from-class="opacity-100 scale-100"
          leave-to-class="opacity-0 scale-95"
      >
        <div
            v-if="hoverNews"
            data-game-update-popover
            class="fixed max-h-[60vh] w-80 overflow-x-hidden overflow-y-auto rounded-lg bg-orange-100 p-4 text-sm text-gray-800 shadow-lg backdrop-blur-md"
            :style="{ left: `${hoverLeft}px`, top: `${hoverTop}px` }"
            @mouseenter="onDetailMouseEnter"
            @mouseleave="onDetailMouseLeave"
        >
          <div v-html="hoverNews.content"></div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from 'vue'
import { i18n } from '@/main.ts'
import { useLangStore } from '@/store/langStore.ts'
import type { LatestNewsRecord, NewsBaseModel } from '@/types/game.ts'
import { getLatestGameNews } from '@/utils/api/game.ts'

const { t } = i18n.global
const langStore = useLangStore()
const lang = ref(langStore.lang)

const newsList = ref<NewsBaseModel[]>([])
const newsRefs = ref<HTMLElement[]>([])
const coverRefs = ref<HTMLElement[]>([])

const hoverNews = ref<NewsBaseModel | null>(null)
const hoveringDetail = ref(false)
const hideTimer: { timer: number | null } = { timer: null }

const hoverTop = ref(0)
const hoverLeft = ref(0)
const HOVER_WIDTH = 320
const HOVER_GAP = 12

function formatTime(postTime: string) {
  const date = new Date(postTime)
  const yyyy = date.getFullYear()
  const mm = String(date.getMonth() + 1).padStart(2, '0')
  const dd = String(date.getDate()).padStart(2, '0')
  const hh = String(date.getHours()).padStart(2, '0')
  const min = String(date.getMinutes()).padStart(2, '0')
  return `${yyyy}-${mm}-${dd} ${hh}:${min}`
}

function htmlToPlainText(html: string) {
  const div = document.createElement('div')
  div.innerHTML = html
  return div.textContent || div.innerText || ''
}

function openUrl(url: string) {
  window.open(url, '_blank')
}

function onNewsMouseEnter(news: NewsBaseModel, event: MouseEvent) {
  if (hideTimer.timer) {
    clearTimeout(hideTimer.timer)
    hideTimer.timer = null
  }

  hoverNews.value = news
  hoveringDetail.value = false

  nextTick(() => {
    const img = event.currentTarget as HTMLElement
    const rect = img.getBoundingClientRect()
    const viewportWidth = window.innerWidth
    const viewportHeight = window.innerHeight
    const popoverEl = document.querySelector('[data-game-update-popover]') as HTMLElement | null
    const popoverHeight = popoverEl?.offsetHeight ?? 220

    let left = rect.left + rect.width / 2 - HOVER_WIDTH / 2
    let top = rect.bottom + HOVER_GAP

    if (left < HOVER_GAP) {
      left = HOVER_GAP
    }

    if (left + HOVER_WIDTH > viewportWidth - HOVER_GAP) {
      left = viewportWidth - HOVER_WIDTH - HOVER_GAP
    }

    if (top + popoverHeight > viewportHeight - HOVER_GAP) {
      top = rect.top - popoverHeight - HOVER_GAP
    }

    if (top < HOVER_GAP) {
      top = HOVER_GAP
    }

    hoverLeft.value = left
    hoverTop.value = top
  })
}

function onNewsMouseLeave() {
  hideTimer.timer = window.setTimeout(() => {
    if (!hoveringDetail.value) {
      hoverNews.value = null
    }
  }, 200)
}

function onDetailMouseEnter() {
  hoveringDetail.value = true
  if (hideTimer.timer) {
    clearTimeout(hideTimer.timer)
    hideTimer.timer = null
  }
}

function onDetailMouseLeave() {
  hoveringDetail.value = false
  hoverNews.value = null
}

async function loadNews() {
  try {
    const res: LatestNewsRecord = await getLatestGameNews()
    newsList.value = lang.value === 'en' ? res.news_en : res.news_zh
    nextTick(() => {
      newsRefs.value = newsRefs.value.slice(0, newsList.value.length)
      coverRefs.value = coverRefs.value.slice(0, newsList.value.length)
    })
  } catch (error) {
    console.error('加载新闻失败', error)
    newsList.value = []
  }
}

onMounted(() => {
  loadNews()
})

watch(
  () => langStore.lang,
  (newLang) => {
    lang.value = newLang
    loadNews()
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

::-webkit-scrollbar {
  height: 8px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: rgba(249, 115, 22, 0.4);
  border-radius: 4px;
  backdrop-filter: blur(4px);
}

::-webkit-scrollbar-thumb:hover {
  background: rgba(249, 115, 22, 0.7);
}
</style>
