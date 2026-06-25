<template>
  <div class="game-news-panel mb-8">
    <div class="mb-4 flex items-center justify-between gap-4">
      <div class="min-w-0">
        <h2 class="game-news-title text-lg font-bold">{{ t('game.news.title') }}</h2>
        <div class="game-news-desc hidden text-sm sm:block">{{ t('game.news.desc') }}</div>
      </div>

      <div v-if="showControls" class="news-pager">
        <span class="news-pager__count hidden text-xs font-semibold sm:block">
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
            <SteamAssetImage
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
import SteamAssetImage from '@/components/common/SteamAssetImage.vue'
import type { LatestNewsRecord, NewsBaseModel } from '~/types/game'

const props = defineProps<{
  initialNewsRecord?: LatestNewsRecord | null
}>()

const { t, locale } = useI18n()
const lang = computed<'zh' | 'en'>(() => locale.value === 'en' ? 'en' : 'zh')

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

function isHtmlElement(element: unknown): element is HTMLElement {
  return typeof HTMLElement !== 'undefined' && element instanceof HTMLElement
}

function setCardRef(element: Element | ComponentPublicInstance | null, index: number) {
  if (isHtmlElement(element)) {
    cardRefs.value[index] = element
    return
  }

  delete cardRefs.value[index]
}

function getOrderedCards() {
  return newsList.value
    .map((_, index) => cardRefs.value[index])
    .filter(isHtmlElement)
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

function applyNewsRecord(record: LatestNewsRecord | null) {
  cardRefs.value = {}
  newsList.value = record
    ? (lang.value === 'en' ? record.news_en : record.news_zh)
    : []
  activeIndex.value = 0
  nextTick(() => {
    updateTrackOffset()
  })
}

onMounted(() => {
  if (newsList.value.length) {
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
  [() => props.initialNewsRecord, lang],
  ([record]) => {
    applyNewsRecord((record as LatestNewsRecord | null | undefined) ?? null)
  },
  { immediate: true }
)
</script>
