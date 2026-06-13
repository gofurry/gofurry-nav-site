<template>
  <div class="game-news-panel mb-8">
    <div class="mb-4 flex items-center justify-between gap-4">
      <div class="min-w-0">
        <h2 class="text-lg font-bold text-orange-950">{{ t('game.news.title') }}</h2>
        <div class="hidden text-sm text-orange-900/55 sm:block">{{ t('game.news.desc') }}</div>
      </div>

      <div v-if="showControls" class="news-pager">
        <span class="hidden text-xs font-semibold text-stone-500/80 sm:block">
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
            />

            <div class="news-card__body">
              <div class="news-card__meta">
                <span>{{ formatTime(news.post_time) }}</span>
                <span v-if="news.name" class="truncate">{{ news.name }}</span>
              </div>

              <h3 class="news-card__title">{{ htmlToPlainText(news.headline) }}</h3>

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

.game-news-panel {
  position: relative;
  overflow: hidden;
  border: 1px solid rgba(126, 92, 58, 0.16);
  border-radius: 1.05rem;
  padding: 1.1rem;
  background: rgba(255, 250, 242, 0.20);
  box-shadow: none;
  backdrop-filter: blur(1px);
}

.news-viewport {
  overflow: hidden;
  padding: 0.1rem 0 0.2rem;
}

.news-track {
  display: flex;
  gap: 0.9rem;
  will-change: transform;
  transition: transform 340ms cubic-bezier(0.22, 1, 0.36, 1);
}

.news-card {
  position: relative;
  width: 15.5rem;
  flex-shrink: 0;
  cursor: pointer;
  overflow: hidden;
  border-radius: 0.78rem;
  border: 1px solid rgba(126, 92, 58, 0.12);
  background: rgba(255, 250, 242, 0.22);
  box-shadow: none;
  backdrop-filter: blur(1px);
  transition: background-color 180ms ease, box-shadow 180ms ease, border-color 180ms ease;
}

.news-card:hover {
  border-color: rgba(180, 96, 24, 0.18);
  background: rgba(255, 244, 228, 0.38);
}

.news-card__cover {
  aspect-ratio: 16 / 9;
  width: 100%;
  max-height: 9.8rem;
  object-fit: cover;
}

.news-card__body {
  padding: 0.82rem 0.9rem 0.95rem;
}

.news-card__meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  color: rgba(120, 83, 53, 0.58);
  font-size: 0.72rem;
  line-height: 1.2;
}

.news-card__title {
  display: -webkit-box;
  min-height: 2.55rem;
  overflow: hidden;
  margin-top: 0.45rem;
  color: rgba(71, 42, 20, 0.92);
  font-weight: 750;
  line-height: 1.32;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.news-card__summary {
  display: -webkit-box;
  margin-top: 0.55rem;
  min-height: 3.65rem;
  overflow: hidden;
  color: rgba(87, 56, 28, 0.68);
  font-size: 0.875rem;
  line-height: 1.45;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
}

.news-pager {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
}

.news-nav-button {
  display: grid;
  place-items: center;
  width: 1.75rem;
  height: 1.75rem;
  border: 0;
  border-radius: 999px;
  background: transparent;
  color: rgba(154, 52, 18, 0.72);
  transition: background-color 180ms ease, border-color 180ms ease, color 180ms ease, opacity 180ms ease;
}

.news-nav-button:hover {
  background: rgba(251, 146, 60, 0.10);
  color: rgba(124, 45, 18, 0.88);
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
  background: linear-gradient(90deg, rgba(255, 250, 242, 0.42), rgba(255, 250, 242, 0));
}

.news-edge--right {
  right: -0.25rem;
  background: linear-gradient(270deg, rgba(255, 250, 242, 0.42), rgba(255, 250, 242, 0));
}

.news-progress-track {
  height: 0.12rem;
  overflow: hidden;
  border-radius: 999px;
  background: rgba(126, 92, 58, 0.10);
}

.news-progress-fill {
  height: 100%;
  border-radius: inherit;
  background: rgba(180, 83, 9, 0.42);
  transition: width 220ms ease;
}

@media (min-width: 640px) {
  .news-card {
    width: 19.5rem;
  }
}

@media (min-width: 1024px) {
  .news-card {
    width: 20.5rem;
  }
}

:global(.dark) .game-news-panel {
  border-color: rgba(226, 232, 240, 0.12);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.052), rgba(226, 232, 240, 0.024)),
    rgba(226, 232, 240, 0.03);
}

:global(.dark) .game-news-panel h2 {
  color: rgba(241, 245, 249, 0.88);
}

:global(.dark) .game-news-panel .text-orange-900\/55,
:global(.dark) .game-news-panel .text-stone-500\/80 {
  color: rgba(203, 213, 225, 0.62);
}

:global(.dark) .news-card {
  border-color: rgba(148, 163, 184, 0.14);
  background: rgba(15, 23, 42, 0.34);
}

:global(.dark) .news-card:hover {
  border-color: rgba(148, 163, 184, 0.28);
  background: rgba(30, 41, 59, 0.52);
}

:global(.dark) .news-card__meta {
  color: rgba(203, 213, 225, 0.58);
}

:global(.dark) .news-card__title {
  color: rgba(226, 232, 240, 0.86);
}

:global(.dark) .news-card__summary {
  color: rgba(203, 213, 225, 0.66);
}

:global(.dark) .news-nav-button {
  color: rgba(180, 213, 226, 0.62);
}

:global(.dark) .news-nav-button:hover {
  background: rgba(148, 163, 184, 0.10);
  color: rgba(226, 232, 240, 0.80);
}

:global(.games-page--dark) .news-nav-button {
  color: rgba(180, 213, 226, 0.62) !important;
}

:global(.games-page--dark) .news-nav-button:hover {
  background: rgba(148, 163, 184, 0.10) !important;
  color: rgba(226, 232, 240, 0.80) !important;
}

:global(.dark) .news-edge--left {
  background: linear-gradient(90deg, rgba(10, 21, 36, 0.58), rgba(10, 21, 36, 0));
}

:global(.dark) .news-edge--right {
  background: linear-gradient(270deg, rgba(10, 21, 36, 0.58), rgba(10, 21, 36, 0));
}

:global(.dark) .news-progress-track {
  background: rgba(148, 163, 184, 0.13);
}

:global(.dark) .news-progress-fill {
  background: rgba(148, 163, 184, 0.46);
}

:global(.games-page--dark) .news-progress-track {
  background: rgba(148, 163, 184, 0.13);
}

:global(.games-page--dark) .news-progress-fill {
  background: rgba(148, 163, 184, 0.46) !important;
}
</style>
