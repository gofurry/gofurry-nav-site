<template>
  <div class="flex min-h-screen w-full flex-col bg-gray-50">
    <NavHeader
      :desktop-bg-url="navPageData.desktopBgUrl"
      :mobile-bg-url="navPageData.mobileBgUrl"
    />
    <main
        v-show="isContentRevealed"
        ref="contentRef"
        class="relative z-10 flex-1 bg-[#f2e3d0]"
        :style="{
        backgroundImage: `url(${bgGrid})`,
        backgroundRepeat: 'repeat'
      }"
    >
      <div class="absolute w-full">
        <NavTransitionBar :initial-saying="navPageData.saying" />
      </div>
      <div class="h-10"></div>
      <NavContent
        :initial-groups="navPageData.groups"
        :initial-sites="navPageData.sites"
        :initial-ping-data="navPageData.pingData"
      />
      <div class="h-10"></div>
    </main>

    <button
      v-if="showScrollDock"
      class="scroll-dock"
      :style="{ '--scroll-progress': `${scrollProgressLabel}%` }"
      :title="t('navHeader.scrollStep')"
      :aria-label="t('navHeader.scrollStep')"
      type="button"
      @click="scrollUpQuarter()"
    >
      <div class="scroll-progress">
        <span>{{ scrollProgressLabel }}%</span>
      </div>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { getGroups, getImageUrl, getPing, getSaying, getSites } from '~/services/nav'
import type { Delay, Group, SayingModel, Site } from '~/types/nav'
import NavHeader from '@/components/nav/NavHeader.vue'
import NavTransitionBar from '@/components/nav/NavTransitionBar.vue'
import NavContent from '@/components/nav/NavContent.vue'
import bgGrid from '@/assets/pngs/bg-grid.png'
import { debounce, throttle } from '@/utils/util'
import { dispatchNavPageReveal, isNavPageRevealLocked } from '@/utils/navPageReveal'

interface NavPageData {
  desktopBgUrl: string | null
  mobileBgUrl: string | null
  saying: SayingModel | null
  groups: Group[]
  sites: Site[]
  pingData: Record<string, Delay>
}

const isContentRevealed = ref(false)
const contentRef = ref<HTMLElement | null>(null)
const { locale, t } = useI18n()
const scrollProgress = ref(0)
const isDesktopViewport = ref(false)

let touchStartY = 0
let mobileMediaQuery: MediaQueryList | null = null

function parsePingData(data: Record<string, string | undefined>) {
  const result: Record<string, Delay> = {}

  for (const key in data) {
    const value = data[key]
    if (typeof value === 'string') {
      try {
        result[key] = JSON.parse(value) as Delay
      } catch {
        result[key] = { status: 'down', delay: '-', loss: '-', time: '-' }
      }
    } else {
      result[key] = { status: 'down', delay: '-', loss: '-', time: '-' }
    }
  }

  return result
}

const lang = computed(() => (locale.value === 'en' ? 'en' : 'zh'))

const { data } = await useAsyncData<NavPageData>(
  () => `nav-page:${lang.value}`,
  async () => {
    const [groups, sites, ping, saying, desktopBgUrl, mobileBgUrl] = await Promise.all([
      getGroups(lang.value),
      getSites(lang.value),
      getPing(),
      getSaying().catch(() => null),
      getImageUrl('standard').catch(() => null),
      getImageUrl('mobile').catch(() => null),
    ])

    return {
      desktopBgUrl,
      mobileBgUrl,
      saying,
      groups: groups.sort((a, b) => Number(a.priority) - Number(b.priority)),
      sites,
      pingData: parsePingData(ping),
    }
  },
  {
    watch: [lang],
    default: () => ({
      desktopBgUrl: null,
      mobileBgUrl: null,
      saying: null,
      groups: [],
      sites: [],
      pingData: {},
    }),
  }
)

const navPageData = computed(() => data.value!)
const scrollProgressLabel = computed(() => Math.round(scrollProgress.value))
const showScrollDock = computed(() => isDesktopViewport.value && windowScrollY.value > 180)
const windowScrollY = ref(0)

function updateViewportState() {
  if (!import.meta.client) {
    return
  }

  isDesktopViewport.value = window.innerWidth >= 1024
}

function updateScrollProgress() {
  if (!import.meta.client) {
    return
  }

  const root = document.documentElement
  const maxScroll = Math.max(root.scrollHeight - window.innerHeight, 0)
  windowScrollY.value = window.scrollY

  if (maxScroll <= 0) {
    scrollProgress.value = 0
    return
  }

  scrollProgress.value = Math.min(100, Math.max(0, (window.scrollY / maxScroll) * 100))
}

function scrollUpQuarter() {
  if (!import.meta.client) {
    return
  }

  const root = document.documentElement
  const maxScroll = Math.max(root.scrollHeight - window.innerHeight, 0)
  const nextTop = Math.max(0, window.scrollY - maxScroll * 0.25)
  window.scrollTo({ top: nextTop, behavior: 'smooth' })
}

function togglePageScrollbar(hidden: boolean) {
  if (!import.meta.client) {
    return
  }

  document.documentElement.classList.toggle('nav-page-scrollbar-hidden', hidden)
  document.body.classList.toggle('nav-page-scrollbar-hidden', hidden)
}

function revealContent(shouldScroll = true, force = false) {
  if (!force && isNavPageRevealLocked()) {
    return
  }

  if (isContentRevealed.value) {
    return
  }

  isContentRevealed.value = true
  dispatchNavPageReveal(true)

  if (!shouldScroll) {
    return
  }

  nextTick(() => {
    contentRef.value?.scrollIntoView({
      behavior: 'smooth',
      block: 'start',
    })
  })
}

function syncRevealByViewport() {
  if (window.innerWidth < 768) {
    revealContent(false, true)
  }

  updateViewportState()
  updateScrollProgress()
}

function handleViewportChange(event: MediaQueryListEvent) {
  if (event.matches) {
    revealContent(false, true)
  }
}

const handleRevealScroll = throttle(() => {
  if (window.scrollY > 24) {
    revealContent()
  }
}, 120)

const handleRevealScrollEnd = debounce(() => {
  if (window.scrollY > 24) {
    revealContent()
  }
}, 180)

function handleScroll() {
  handleRevealScroll()
  handleRevealScrollEnd()
  updateScrollProgress()
}

function handleResize() {
  updateViewportState()
  updateScrollProgress()
}

function handleWheel(event: WheelEvent) {
  if (event.deltaY > 8) {
    revealContent()
  }
}

function handleTouchStart(event: TouchEvent) {
  touchStartY = event.touches[0]?.clientY ?? 0
}

function handleTouchMove(event: TouchEvent) {
  const currentY = event.touches[0]?.clientY ?? touchStartY
  if (touchStartY - currentY > 12) {
    revealContent()
  }
}

function handleKeydown(event: KeyboardEvent) {
  if (['ArrowDown', 'PageDown', 'Space', 'End'].includes(event.code)) {
    revealContent()
  }
}

onMounted(() => {
  dispatchNavPageReveal(false)
  togglePageScrollbar(true)
  syncRevealByViewport()

  mobileMediaQuery = window.matchMedia('(max-width: 767px)')
  mobileMediaQuery.addEventListener('change', handleViewportChange)
  window.addEventListener('scroll', handleScroll, { passive: true })
  window.addEventListener('resize', handleResize)
  window.addEventListener('wheel', handleWheel, { passive: true })
  window.addEventListener('touchstart', handleTouchStart, { passive: true })
  window.addEventListener('touchmove', handleTouchMove, { passive: true })
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  dispatchNavPageReveal(true)
  togglePageScrollbar(false)
  mobileMediaQuery?.removeEventListener('change', handleViewportChange)
  window.removeEventListener('scroll', handleScroll)
  window.removeEventListener('resize', handleResize)
  window.removeEventListener('wheel', handleWheel)
  window.removeEventListener('touchstart', handleTouchStart)
  window.removeEventListener('touchmove', handleTouchMove)
  window.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
:global(html.nav-page-scrollbar-hidden),
:global(body.nav-page-scrollbar-hidden) {
  -ms-overflow-style: none;
  scrollbar-width: none;
}

:global(html.nav-page-scrollbar-hidden::-webkit-scrollbar),
:global(body.nav-page-scrollbar-hidden::-webkit-scrollbar) {
  display: none;
}

.scroll-dock {
  position: fixed;
  right: 24px;
  bottom: 24px;
  z-index: 70;
  display: grid;
  place-items: center;
  width: 52px;
  height: 52px;
  background: transparent;
  opacity: 0.82;
  transition: opacity 180ms ease, filter 180ms ease;
}

.scroll-dock:hover {
  opacity: 0.94;
  filter: drop-shadow(0 10px 24px rgba(10, 20, 28, 0.16));
}

.scroll-progress {
  position: relative;
  display: grid;
  place-items: center;
  width: 52px;
  height: 52px;
  border-radius: 999px;
  background:
    conic-gradient(from 210deg, rgba(132, 219, 255, 0.88) var(--scroll-progress), rgba(255, 255, 255, 0.08) 0),
    radial-gradient(circle at 30% 25%, rgba(255, 255, 255, 0.12), transparent 42%);
}

.scroll-progress::before {
  content: '';
  position: absolute;
  inset: 4px;
  border-radius: inherit;
  background:
    linear-gradient(180deg, rgba(10, 17, 21, 0.64), rgba(18, 28, 33, 0.56));
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
}

.scroll-progress span {
  position: relative;
  z-index: 1;
  font-size: 0.68rem;
  font-weight: 600;
  color: rgba(232, 248, 255, 0.86);
  letter-spacing: 0;
}

@media (max-width: 1023px) {
  .scroll-dock {
    display: none;
  }
}
</style>
