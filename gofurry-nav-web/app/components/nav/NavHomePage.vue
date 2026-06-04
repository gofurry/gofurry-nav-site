<template>
  <div class="flex min-h-screen w-full flex-col bg-gray-50">
    <NavHeader
      :desktop-bg-url="navPageData.desktopBgUrl"
      :mobile-bg-url="navPageData.mobileBgUrl"
    />
    <NavToolDock :items="toolDockSites" />
    <main
        v-show="isContentRevealed"
        ref="contentRef"
        class="relative z-10 flex-1 overflow-hidden"
    >
      <GoFurryGridBackground :fixed="false" palette="nav-content" />
      <div class="absolute z-30 w-full">
        <NavTransitionBar :initial-saying="navPageData.saying" />
      </div>
      <div class="relative z-10">
        <div class="h-10"></div>
        <NavContent
          :initial-groups="navPageData.groups"
          :initial-ping-data="navPageData.pingData"
        />
        <div class="h-10"></div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { getNavHome } from '~/services/nav'
import type { Delay, Group, SayingModel, Site } from '~/types/nav'
import NavHeader from '@/components/nav/NavHeader.vue'
import NavToolDock from '@/components/nav/NavToolDock.vue'
import NavTransitionBar from '@/components/nav/NavTransitionBar.vue'
import NavContent from '@/components/nav/NavContent.vue'
import GoFurryGridBackground from '@/components/common/GoFurryGridBackground.vue'
import { debounce, throttle } from '@/utils/util'
import { dispatchNavPageReveal, isNavPageRevealLocked } from '@/utils/navPageReveal'
import { readDisplayMode, subscribeModeChange, type DisplayMode } from '@/utils/modeStorage'

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
const displayMode = ref<DisplayMode>(readDisplayMode())

let touchStartY = 0
let mobileMediaQuery: MediaQueryList | null = null
let stopModeSubscription: (() => void) | null = null

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
const toolDockSites = computed(() => {
  return navPageData.value.sites.filter(site => displayMode.value === 'nsfw' || String(site.nsfw) !== '1')
})

const { data } = await useAsyncData<NavPageData>(
  () => `nav-page:${lang.value}`,
  async () => {
    const home = await getNavHome(lang.value)

    return {
      desktopBgUrl: home.backgrounds.desktop || null,
      mobileBgUrl: home.backgrounds.mobile || null,
      saying: home.saying,
      groups: home.groups.sort((a, b) => Number(a.priority) - Number(b.priority)),
      sites: home.sites,
      pingData: parsePingData(home.ping),
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
  displayMode.value = readDisplayMode()
  stopModeSubscription = subscribeModeChange(({ displayMode: nextMode }) => {
    displayMode.value = nextMode
  })

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
  stopModeSubscription?.()
  mobileMediaQuery?.removeEventListener('change', handleViewportChange)
  window.removeEventListener('scroll', handleScroll)
  window.removeEventListener('resize', handleResize)
  window.removeEventListener('wheel', handleWheel)
  window.removeEventListener('touchstart', handleTouchStart)
  window.removeEventListener('touchmove', handleTouchMove)
  window.removeEventListener('keydown', handleKeydown)
})
</script>
