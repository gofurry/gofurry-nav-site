<template>
  <div class="mb-8 overflow-hidden rounded-2xl border border-white/35 bg-orange-50/92 p-5 shadow-[0_20px_50px_rgba(120,68,20,0.08)] backdrop-blur-md">
    <div class="mb-4 flex items-center justify-between gap-4">
      <div class="min-w-0">
        <h2 class="text-lg font-bold text-orange-950">{{ t('game.news.title') }}</h2>
        <div class="hidden text-sm text-orange-900/55 sm:block">{{ t('game.news.desc') }}</div>
      </div>

      <div v-if="showControls" class="flex items-center gap-2">
        <span class="hidden text-xs font-medium text-orange-900/45 sm:block">
          {{ activeIndex + 1 }}/{{ newsList.length }}
        </span>
        <button
          class="news-nav-button"
          :class="{ 'news-nav-button--disabled': !canMovePrev }"
          :aria-label="t('game.news.previous')"
          :title="t('game.news.previous')"
          type="button"
          :disabled="!canMovePrev"
          @click="moveByStep(-1)"
        >
          <span aria-hidden="true">‹</span>
        </button>
        <button
          class="news-nav-button"
          :class="{ 'news-nav-button--disabled': !canMoveNext }"
          :aria-label="t('game.news.next')"
          :title="t('game.news.next')"
          type="button"
          :disabled="!canMoveNext"
          @click="moveByStep(1)"
        >
          <span aria-hidden="true">›</span>
        </button>
      </div>
    </div>

    <div class="relative">
      <div class="news-edge news-edge--left" :class="{ 'news-edge--visible': canMovePrev }"></div>
      <div class="news-edge news-edge--right" :class="{ 'news-edge--visible': canMoveNext }"></div>

      <div
        ref="viewportRef"
        class="news-viewport"
      >
        <div
          ref="trackRef"
          class="news-track"
          :style="{ transform: `translate3d(-${trackOffset}px, 0, 0)` }"
        >
          <div
            v-for="(news, index) in newsList"
            :key="newsKey(news, index)"
            :ref="(el) => setCardRef(el, index)"
            class="news-card"
            @click="openUrl(news.url)"
          >
            <img
              :src="news.header"
              :alt="newsImageAlt(news)"
              class="news-card__cover"
              @mouseenter="onNewsMouseEnter(news, $event)"
              @mouseleave="onNewsMouseLeave"
            />

            <div class="news-card__body">
              <h3 class="news-card__title">{{ htmlToPlainText(news.headline) }}</h3>

              <div class="mt-2 flex items-center justify-between gap-3 text-xs text-orange-900/45">
                <span>{{ formatTime(news.post_time) }}</span>
                <span class="truncate text-right">{{ news.author }}</span>
              </div>

              <p class="news-card__summary">
                {{ htmlToPlainText(news.content) }}
              </p>
            </div>
          </div>
        </div>
      </div>

      <div v-if="showControls" class="mt-4">
        <div class="news-progress-track">
          <div class="news-progress-fill" :style="{ width: `${progressWidth}%` }"></div>
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
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import type { ComponentPublicInstance } from 'vue'
import { useI18n } from 'vue-i18n'
import { useLangStore } from '@/store/langStore'
import { getLatestGameNews } from '~/services/game'
import type { LatestNewsRecord, NewsBaseModel } from '~/types/game'

const props = defineProps<{
  initialNewsRecord?: LatestNewsRecord | null
}>()

const { t } = useI18n()
const langStore = useLangStore()
const lang = ref(langStore.lang)

const newsList = ref<NewsBaseModel[]>([])
const viewportRef = ref<HTMLElement | null>(null)
const trackRef = ref<HTMLElement | null>(null)
const cardRefs = ref<Record<number, HTMLElement>>({})
const activeIndex = ref(0)
const trackOffset = ref(0)

const hoverNews = ref<NewsBaseModel | null>(null)
const hoveringDetail = ref(false)
const hideTimer: { timer: number | null } = { timer: null }

const hoverTop = ref(0)
const hoverLeft = ref(0)
const HOVER_WIDTH = 320
const HOVER_GAP = 12

const showControls = computed(() => newsList.value.length > 1)
const canMovePrev = computed(() => activeIndex.value > 0)
const canMoveNext = computed(() => activeIndex.value < newsList.value.length - 1)
const progressWidth = computed(() => {
  if (!newsList.value.length) {
    return 0
  }

  if (newsList.value.length === 1) {
    return 100
  }

  return ((activeIndex.value + 1) / newsList.value.length) * 100
})

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
  return html
    .replace(/<[^>]+>/g, ' ')
    .replace(/&nbsp;/gi, ' ')
    .replace(/&amp;/gi, '&')
    .replace(/&lt;/gi, '<')
    .replace(/&gt;/gi, '>')
    .replace(/\s+/g, ' ')
    .trim()
}

function newsImageAlt(news: NewsBaseModel) {
  return htmlToPlainText(news.headline || news.name || t('game.news.title'))
}

function openUrl(url: string) {
  window.open(url, '_blank')
}

function newsKey(news: NewsBaseModel, index: number) {
  return `${news.id}:${news.post_time}:${news.url}:${index}`
}

function setCardRef(element: Element | ComponentPublicInstance | null, index: number) {
  if (element instanceof HTMLElement) {
    cardRefs.value[index] = element
    return
  }

  delete cardRefs.value[index]
}

function getOrderedCards() {
  return newsList.value
    .map((_, index) => cardRefs.value[index])
    .filter((card): card is HTMLElement => card instanceof HTMLElement)
}

function getTargetOffset(targetIndex: number) {
  const viewport = viewportRef.value
  const track = trackRef.value
  const cards = getOrderedCards()
  const target = cards[targetIndex]

  if (!viewport || !track || !target) {
    return 0
  }

  const maxOffset = Math.max(track.scrollWidth - viewport.clientWidth, 0)
  return Math.min(target.offsetLeft, maxOffset)
}

function updateTrackOffset() {
  if (!newsList.value.length) {
    activeIndex.value = 0
    trackOffset.value = 0
    return
  }

  activeIndex.value = Math.min(Math.max(activeIndex.value, 0), newsList.value.length - 1)
  trackOffset.value = getTargetOffset(activeIndex.value)
}

function moveByStep(direction: -1 | 1) {
  const nextIndex = Math.min(
    Math.max(activeIndex.value + direction, 0),
    newsList.value.length - 1
  )

  if (nextIndex === activeIndex.value) {
    return
  }

  activeIndex.value = nextIndex
  trackOffset.value = getTargetOffset(nextIndex)
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

function applyNewsRecord(record: LatestNewsRecord) {
  cardRefs.value = {}
  newsList.value = lang.value === 'en' ? record.news_en : record.news_zh
  activeIndex.value = 0
  nextTick(() => {
    updateTrackOffset()
  })
}

async function loadNews() {
  try {
    const record = props.initialNewsRecord ?? await getLatestGameNews()
    applyNewsRecord(record)
  } catch (error) {
    console.error('Failed to load latest game news:', error)
    newsList.value = []
  }
}

if (props.initialNewsRecord) {
  applyNewsRecord(props.initialNewsRecord)
}

onMounted(() => {
  if (!newsList.value.length) {
    loadNews()
  } else {
    nextTick(() => {
      updateTrackOffset()
    })
  }

  window.addEventListener('resize', updateTrackOffset)
})

onUnmounted(() => {
  window.removeEventListener('resize', updateTrackOffset)
})

watch(
  () => langStore.lang,
  (newLang) => {
    lang.value = newLang

    if (props.initialNewsRecord) {
      applyNewsRecord(props.initialNewsRecord)
      return
    }

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

.news-viewport {
  overflow: hidden;
  padding: 0.1rem 0 0.35rem;
}

.news-track {
  display: flex;
  gap: 1rem;
  will-change: transform;
  transition: transform 340ms cubic-bezier(0.22, 1, 0.36, 1);
}

.news-card {
  position: relative;
  width: 15.5rem;
  flex-shrink: 0;
  cursor: pointer;
  overflow: hidden;
  border-radius: 1rem;
  border: 1px solid rgba(255, 255, 255, 0.38);
  background:
    linear-gradient(180deg, rgba(255, 251, 245, 0.95), rgba(254, 239, 217, 0.92));
  box-shadow:
    0 14px 28px rgba(128, 77, 24, 0.08),
    inset 0 1px 0 rgba(255, 255, 255, 0.36);
  transition: background-color 180ms ease, box-shadow 180ms ease, border-color 180ms ease;
}

.news-card:hover {
  border-color: rgba(255, 255, 255, 0.52);
  background:
    linear-gradient(180deg, rgba(255, 250, 242, 0.98), rgba(252, 236, 214, 0.96));
  box-shadow:
    0 20px 34px rgba(128, 77, 24, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.4);
}

.news-card__cover {
  aspect-ratio: 16 / 9;
  width: 100%;
  object-fit: cover;
}

.news-card__body {
  padding: 0.95rem 0.95rem 1rem;
}

.news-card__title {
  display: -webkit-box;
  min-height: 2.8rem;
  overflow: hidden;
  color: rgba(89, 45, 10, 0.96);
  font-weight: 700;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.news-card__summary {
  display: -webkit-box;
  margin-top: 0.7rem;
  min-height: 4rem;
  overflow: hidden;
  color: rgba(87, 56, 28, 0.78);
  font-size: 0.925rem;
  line-height: 1.45;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
}

.news-nav-button {
  display: grid;
  place-items: center;
  width: 2rem;
  height: 2rem;
  border: 1px solid rgba(180, 96, 24, 0.16);
  border-radius: 999px;
  background: rgba(255, 250, 244, 0.8);
  color: rgba(139, 72, 16, 0.88);
  backdrop-filter: blur(10px);
  transition: background-color 180ms ease, border-color 180ms ease, color 180ms ease, opacity 180ms ease;
}

.news-nav-button:hover {
  background: rgba(255, 245, 233, 0.95);
  border-color: rgba(180, 96, 24, 0.24);
}

.news-nav-button--disabled {
  opacity: 0.36;
}

.news-edge {
  pointer-events: none;
  position: absolute;
  top: 0;
  bottom: 1.6rem;
  z-index: 1;
  width: 2.5rem;
  opacity: 0;
  transition: opacity 180ms ease;
}

.news-edge--visible {
  opacity: 1;
}

.news-edge--left {
  left: -0.25rem;
  background: linear-gradient(90deg, rgba(252, 243, 231, 0.92), rgba(252, 243, 231, 0));
}

.news-edge--right {
  right: -0.25rem;
  background: linear-gradient(270deg, rgba(252, 243, 231, 0.92), rgba(252, 243, 231, 0));
}

.news-progress-track {
  height: 0.25rem;
  overflow: hidden;
  border-radius: 999px;
  background: rgba(181, 101, 37, 0.12);
}

.news-progress-fill {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, rgba(245, 158, 11, 0.9), rgba(251, 191, 36, 0.96));
  transition: width 220ms ease;
}

@media (min-width: 640px) {
  .news-card {
    width: 19.5rem;
  }
}

@media (min-width: 1024px) {
  .news-card {
    width: 21rem;
  }
}
</style>
