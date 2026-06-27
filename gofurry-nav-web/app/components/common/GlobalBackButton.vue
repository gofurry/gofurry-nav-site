<template>
  <button
    v-if="shouldRender"
    class="global-back-button"
    :class="{ 'global-back-button--visible': isVisible }"
    type="button"
    :aria-label="t('common.back')"
    :title="t('common.back')"
    @click="goBack"
  >
    <span class="global-back-button__icon" aria-hidden="true">
      <img :src="backIcon" alt="" />
    </span>
    <span class="global-back-button__label">{{ t('common.back') }}</span>
  </button>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import backIcon from '@/assets/svgs/back.svg'

const route = useRoute()
const router = useRouter()
const localePath = useLocalePath()
const { t } = useI18n()

const isMounted = ref(false)
const isScrollingDown = ref(false)
const previousInternalPath = ref<string | null>(null)
let lastScrollY = 0
let scrollFrame: number | null = null

const normalizedPath = computed(() => normalizeRoutePath(route.path))
const shouldRender = computed(() => (
  isMounted.value
  && normalizedPath.value !== '/'
  && normalizedPath.value !== '/games'
  && !normalizedPath.value.startsWith('/workshop')
))
const isVisible = computed(() => shouldRender.value && !isScrollingDown.value)

function normalizeRoutePath(path: string) {
  const normalized = path.replace(/^\/(zh|en)(?=\/|$)/, '') || '/'
  return normalized.length > 1 ? normalized.replace(/\/$/, '') : normalized
}

function currentScrollY() {
  const scroller = document.scrollingElement || document.documentElement || document.body
  return window.scrollY || scroller?.scrollTop || 0
}

function updateScrollDirection() {
  scrollFrame = null
  const nextScrollY = currentScrollY()
  const delta = nextScrollY - lastScrollY

  if (nextScrollY < 24 || delta < -6) {
    isScrollingDown.value = false
  } else if (delta > 6) {
    isScrollingDown.value = true
  }

  lastScrollY = Math.max(nextScrollY, 0)
}

function scheduleScrollDirectionUpdate() {
  if (scrollFrame !== null) {
    return
  }

  scrollFrame = window.requestAnimationFrame(updateScrollDirection)
}

function fallbackPath() {
  return normalizedPath.value.startsWith('/games/')
    ? localePath('/games')
    : localePath('/')
}

function hasSameOriginReferrer() {
  if (!document.referrer) {
    return false
  }

  try {
    return new URL(document.referrer).origin === window.location.origin
  } catch {
    return false
  }
}

async function goBack() {
  if (previousInternalPath.value || hasSameOriginReferrer()) {
    router.back()
    return
  }

  await router.push(fallbackPath())
}

watch(
  () => route.fullPath,
  (nextPath, previousPath) => {
    if (previousPath && previousPath !== nextPath) {
      previousInternalPath.value = previousPath
    }
    isScrollingDown.value = false
    lastScrollY = import.meta.client ? currentScrollY() : 0
  }
)

onMounted(() => {
  isMounted.value = true
  lastScrollY = currentScrollY()
  window.addEventListener('scroll', scheduleScrollDirectionUpdate, { passive: true })
  document.addEventListener('scroll', scheduleScrollDirectionUpdate, { passive: true, capture: true })
})

onUnmounted(() => {
  if (scrollFrame !== null) {
    window.cancelAnimationFrame(scrollFrame)
    scrollFrame = null
  }

  window.removeEventListener('scroll', scheduleScrollDirectionUpdate)
  document.removeEventListener('scroll', scheduleScrollDirectionUpdate, { capture: true })
})
</script>

<style scoped>
.global-back-button {
  --global-back-button-bg: #b9aa94;
  --global-back-button-hover-bg: #a99880;

  position: fixed;
  left: calc(3.65rem + env(safe-area-inset-left));
  top: calc(5.05rem + env(safe-area-inset-top));
  z-index: 100;
  display: grid;
  width: 36px;
  height: 34px;
  place-items: center;
  border: 0;
  border-radius: 12px;
  background: var(--global-back-button-bg);
  box-shadow: none;
  color: rgba(248, 250, 252, 0.96);
  cursor: pointer;
  isolation: isolate;
  opacity: 0;
  outline: none;
  overflow: hidden;
  padding: 6px;
  pointer-events: none;
  transform: translate3d(0, -8px, 0);
  appearance: none;
  backface-visibility: hidden;
  transition:
    opacity 220ms ease,
    transform 220ms ease,
    background-color 220ms ease;
  will-change: opacity, transform;
}

html.dark .global-back-button {
  --global-back-button-bg: #475569;
  --global-back-button-hover-bg: #536274;
}

.global-back-button--visible {
  opacity: 0.88;
  pointer-events: auto;
  transform: translate3d(0, 0, 0);
}

.global-back-button:hover {
  opacity: 0.96;
  background: var(--global-back-button-hover-bg);
}

.global-back-button__icon {
  display: block;
  width: 18px;
  height: 18px;
  color: currentColor;
}

.global-back-button__icon img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: contain;
  filter: brightness(0) saturate(100%) invert(98%) sepia(4%) saturate(734%) hue-rotate(180deg) brightness(102%) contrast(97%);
}

html.dark .global-back-button__icon img {
  filter: brightness(0) saturate(100%) invert(98%) sepia(4%) saturate(734%) hue-rotate(180deg) brightness(102%) contrast(97%);
}

.global-back-button__label {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
}

@media (max-width: 639px) {
  .global-back-button {
    left: calc(1.35rem + env(safe-area-inset-left));
    top: calc(4.85rem + env(safe-area-inset-top));
  }
}
</style>
