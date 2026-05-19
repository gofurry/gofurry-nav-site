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
        :initial-display-mode="routeDisplayMode"
      />
      <div class="h-10"></div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { getGroups, getImageUrl, getPing, getSaying, getSites } from '~/services/nav'
import type { Delay, Group, SayingModel, Site } from '~/types/nav'
import NavHeader from '@/components/nav/NavHeader.vue'
import NavTransitionBar from '@/components/nav/NavTransitionBar.vue'
import NavContent from '@/components/nav/NavContent.vue'
import bgGrid from '@/assets/pngs/bg-grid.png'
import { debounce, throttle } from '@/utils/util'
import { dispatchNavPageReveal, isNavPageRevealLocked } from '@/utils/navPageReveal'
import { normalizeDisplayMode, writeMode } from '@/utils/modeStorage'

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
const { locale } = useI18n()
const route = useRoute()

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
const routeDisplayMode = computed(() => {
  return route.query.mode == null ? undefined : normalizeDisplayMode(route.query.mode)
})

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

watch(
  routeDisplayMode,
  (mode) => {
    if (mode) {
      writeMode(mode)
    }
  },
  { immediate: true }
)

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
}

function handleResize() {}

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
  mobileMediaQuery?.removeEventListener('change', handleViewportChange)
  window.removeEventListener('scroll', handleScroll)
  window.removeEventListener('resize', handleResize)
  window.removeEventListener('wheel', handleWheel)
  window.removeEventListener('touchstart', handleTouchStart)
  window.removeEventListener('touchmove', handleTouchMove)
  window.removeEventListener('keydown', handleKeydown)
})
</script>
