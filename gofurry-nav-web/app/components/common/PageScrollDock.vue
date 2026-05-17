<template>
  <button
    v-show="isMounted && isVisible"
    class="page-scroll-dock"
    :class="{ 'page-scroll-dock--visible': isVisible }"
    :style="{ '--scroll-progress': `${progressLabel}%` }"
    :title="t('common.scrollStep')"
    :aria-label="t('common.scrollProgress', { percent: progressLabel })"
    type="button"
    @click="scrollUpQuarter"
  >
    <span class="page-scroll-dock__core">{{ progressLabel }}%</span>
  </button>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

const route = useRoute()
const { t } = useI18n()

const scrollProgress = ref(0)
const scrollTop = ref(0)
const maxScroll = ref(0)
const isDesktopViewport = ref(false)
const isMounted = ref(false)
const activeScroller = ref<HTMLElement | null>(null)

let updateTimer: ReturnType<typeof setTimeout> | null = null
let resizeObserver: ResizeObserver | null = null
let observedScroller: HTMLElement | null = null

const progressLabel = computed(() => Math.round(scrollProgress.value))
const isVisible = computed(() => isDesktopViewport.value && maxScroll.value > 24 && scrollTop.value > 72)

function getDocumentScroller() {
  return (document.scrollingElement || document.documentElement || document.body) as HTMLElement | null
}

function getElementMaxScroll(element: HTMLElement) {
  return Math.max(element.scrollHeight - element.clientHeight, 0)
}

function resolveActiveScroller() {
  const documentScroller = getDocumentScroller()
  const documentMaxScroll = documentScroller ? getElementMaxScroll(documentScroller) : 0

  if (documentMaxScroll > 24) {
    return documentScroller
  }

  const candidates = Array.from(document.querySelectorAll<HTMLElement>('body *'))
    .filter((element) => {
      const style = window.getComputedStyle(element)
      const overflowY = style.overflowY
      return (
        element.clientHeight > 120
        && getElementMaxScroll(element) > 24
        && ['auto', 'scroll', 'overlay'].includes(overflowY)
      )
    })
    .sort((a, b) => getElementMaxScroll(b) - getElementMaxScroll(a))

  return candidates[0] || documentScroller
}

function getScrollMetrics() {
  const scroller = activeScroller.value || resolveActiveScroller()
  activeScroller.value = scroller

  if (!scroller) {
    return { top: 0, max: 0 }
  }

  const documentScroller = getDocumentScroller()
  const isDocumentScroller = scroller === documentScroller || scroller === document.documentElement || scroller === document.body
  const top = isDocumentScroller
    ? window.scrollY || scroller.scrollTop || document.documentElement.scrollTop || document.body.scrollTop || 0
    : scroller.scrollTop

  return {
    top,
    max: getElementMaxScroll(scroller),
  }
}

function syncObservedScroller() {
  if (!import.meta.client) {
    return
  }

  const nextScroller = resolveActiveScroller()
  if (nextScroller === observedScroller) {
    activeScroller.value = nextScroller
    return
  }

  observedScroller?.removeEventListener('scroll', updateScrollState)
  observedScroller = nextScroller
  activeScroller.value = nextScroller

  const documentScroller = getDocumentScroller()
  const isDocumentScroller = nextScroller === documentScroller || nextScroller === document.documentElement || nextScroller === document.body
  if (nextScroller && !isDocumentScroller) {
    nextScroller.addEventListener('scroll', updateScrollState, { passive: true })
  }
}

function updateViewportState() {
  if (!import.meta.client) {
    return
  }

  isDesktopViewport.value = window.innerWidth >= 768
}

function updateScrollState() {
  if (!import.meta.client) {
    return
  }

  const metrics = getScrollMetrics()
  scrollTop.value = metrics.top
  maxScroll.value = metrics.max

  if (metrics.max <= 0) {
    scrollProgress.value = 0
    return
  }

  scrollProgress.value = Math.min(100, Math.max(0, (metrics.top / metrics.max) * 100))
}

function refreshDockState() {
  updateViewportState()
  syncObservedScroller()
  updateScrollState()
}

function scheduleRefresh() {
  if (!import.meta.client) {
    return
  }

  if (updateTimer) {
    clearTimeout(updateTimer)
  }

  updateTimer = setTimeout(() => {
    refreshDockState()
  }, 120)
}

function scrollUpQuarter() {
  if (!import.meta.client) {
    return
  }

  const metrics = getScrollMetrics()
  const nextTop = Math.max(0, metrics.top - metrics.max * 0.25)
  const scroller = activeScroller.value || getDocumentScroller()
  const documentScroller = getDocumentScroller()
  const isDocumentScroller = scroller === documentScroller || scroller === document.documentElement || scroller === document.body

  if (isDocumentScroller) {
    window.scrollTo({ top: nextTop, behavior: 'smooth' })
    return
  }

  scroller?.scrollTo({ top: nextTop, behavior: 'smooth' })
}

watch(
  () => route.fullPath,
  async () => {
    if (!import.meta.client) {
      return
    }

    await nextTick()
    refreshDockState()
    scheduleRefresh()
  }
)

onMounted(() => {
  isMounted.value = true
  refreshDockState()
  requestAnimationFrame(() => {
    refreshDockState()
  })
  resizeObserver = new ResizeObserver(() => {
    scheduleRefresh()
  })
  resizeObserver.observe(document.documentElement)
  resizeObserver.observe(document.body)
  window.addEventListener('scroll', updateScrollState, { passive: true })
  window.addEventListener('resize', refreshDockState)
  window.addEventListener('load', scheduleRefresh)
})

onUnmounted(() => {
  if (updateTimer) {
    clearTimeout(updateTimer)
  }

  resizeObserver?.disconnect()
  resizeObserver = null
  observedScroller?.removeEventListener('scroll', updateScrollState)
  observedScroller = null

  window.removeEventListener('scroll', updateScrollState)
  window.removeEventListener('resize', refreshDockState)
  window.removeEventListener('load', scheduleRefresh)
})
</script>

<style scoped>
.page-scroll-dock {
  position: fixed;
  right: 22px;
  bottom: 22px;
  z-index: 85;
  display: grid;
  place-items: center;
  width: 46px;
  height: 46px;
  border: none;
  border-radius: 999px;
  background:
    conic-gradient(from 220deg, rgba(122, 228, 255, 0.88) var(--scroll-progress), rgba(255, 255, 255, 0.08) 0),
    radial-gradient(circle at 32% 28%, rgba(255, 255, 255, 0.14), transparent 48%);
  box-shadow:
    0 12px 32px rgba(8, 14, 20, 0.2),
    inset 0 0 0 1px rgba(255, 255, 255, 0.08);
  opacity: 0;
  pointer-events: none;
  transition: opacity 180ms ease, box-shadow 180ms ease, filter 180ms ease;
}

.page-scroll-dock--visible {
  opacity: 0.78;
  pointer-events: auto;
}

.page-scroll-dock::before {
  content: '';
  position: absolute;
  inset: 4px;
  border-radius: inherit;
  background: linear-gradient(180deg, rgba(13, 22, 30, 0.6), rgba(17, 28, 36, 0.46));
  backdrop-filter: blur(10px);
}

.page-scroll-dock::after {
  content: '';
  position: absolute;
  inset: 11px;
  border-radius: inherit;
  border: 1px solid rgba(132, 226, 255, 0.16);
  opacity: 0.7;
}

.page-scroll-dock:hover {
  opacity: 0.93;
  box-shadow:
    0 14px 34px rgba(8, 14, 20, 0.24),
    inset 0 0 0 1px rgba(255, 255, 255, 0.1);
  filter: saturate(1.06);
}

.page-scroll-dock__core {
  position: relative;
  z-index: 1;
  font-size: 0.64rem;
  font-weight: 600;
  color: rgba(235, 248, 255, 0.88);
  letter-spacing: 0;
}

@media (max-width: 767px) {
  .page-scroll-dock {
    display: none;
  }
}
</style>
