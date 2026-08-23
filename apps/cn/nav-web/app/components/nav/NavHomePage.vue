<template>
  <div class="nav-home-page flex w-full flex-col">
    <NavHeader
      :desktop-bg-url="navPageData.desktopBgUrl"
      :mobile-bg-url="navPageData.mobileBgUrl"
    />
    <NavToolDock v-if="isContentRevealed" />
    <main
        v-if="isContentMounted"
        ref="contentRef"
        class="nav-content-shell relative z-10 flex-1 overflow-hidden"
    >
      <GoFurryGridBackground :fixed="false" palette="nav-content" />
      <div class="absolute z-30 w-full">
        <NavTransitionBar :initial-saying="navPageData.saying" />
      </div>
      <div class="relative z-10">
        <div class="mx-auto w-full max-w-[2080px] px-4 sm:px-6 xl:px-8">
          <div class="h-10"></div>
          <NavContent
            :initial-groups="navPageData.groups"
            :initial-spotlight="navPageData.spotlight"
            :initial-ping-data="navPageData.pingData"
          />
          <div class="h-10"></div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { getNavHome } from '~/services/nav'
import type { Delay, Group, NavHomeSpotlight, SayingModel } from '~/types/nav'
import NavHeader from '@/components/nav/NavHeader.vue'
import NavToolDock from '@/components/nav/NavToolDock.vue'
import NavTransitionBar from '@/components/nav/NavTransitionBar.vue'
import NavContent from '@/components/nav/NavContent.vue'
import GoFurryGridBackground from '@/components/common/GoFurryGridBackground.vue'
import { debounce, throttle } from '@/utils/util'
import { dispatchNavPageReveal, isNavPageRevealLocked } from '@/utils/navPageReveal'

interface NavPageData {
  desktopBgUrl: string | null
  mobileBgUrl: string | null
  saying: SayingModel | null
  groups: Group[]
  spotlight: NavHomeSpotlight
  pingData: Record<string, Delay>
}

const isContentMounted = ref(false)
const isContentRevealed = ref(false)
const contentRef = ref<HTMLElement | null>(null)
const { locale } = useI18n()

let touchStartY = 0
let mobileMediaQuery: MediaQueryList | null = null
let prewarmTimer: ReturnType<typeof setTimeout> | null = null
let idlePrewarmId: number | null = null
let revealInProgress = false

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
    const home = await getNavHome(lang.value)

    return {
      desktopBgUrl: home.backgrounds.desktop || null,
      mobileBgUrl: home.backgrounds.mobile || null,
      saying: home.saying,
      groups: home.groups.sort((a, b) => Number(a.priority) - Number(b.priority)),
      spotlight: home.spotlight,
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
      spotlight: { page_size: 6, featured: [], popular: [], latest: [], random: [] },
      pingData: {},
    }),
  }
)

const navPageData = computed(() => data.value!)

function mountContent() {
  if (isContentMounted.value) {
    return
  }

  isContentMounted.value = true
}

function cancelScheduledPrewarm() {
  if (prewarmTimer) {
    clearTimeout(prewarmTimer)
    prewarmTimer = null
  }

  if (idlePrewarmId !== null && window.cancelIdleCallback) {
    window.cancelIdleCallback(idlePrewarmId)
    idlePrewarmId = null
  }
}

function scheduleContentPrewarm() {
  if (isContentMounted.value || !import.meta.client) {
    return
  }

  cancelScheduledPrewarm()

  const prewarm = () => {
    prewarmTimer = null
    idlePrewarmId = null
    mountContent()
  }

  prewarmTimer = setTimeout(() => {
    prewarmTimer = null
    if (isContentMounted.value) {
      return
    }

    if (window.requestIdleCallback) {
      idlePrewarmId = window.requestIdleCallback(prewarm, { timeout: 1400 })
      return
    }

    prewarm()
  }, 700)
}

function waitForFrames(count = 2) {
  return new Promise<void>((resolve) => {
    const step = () => {
      count -= 1
      if (count <= 0) {
        resolve()
        return
      }
      requestAnimationFrame(step)
    }

    requestAnimationFrame(step)
  })
}

async function revealContent(shouldScroll = true, force = false) {
  if (!force && isNavPageRevealLocked()) {
    return
  }

  if (isContentRevealed.value || revealInProgress) {
    return
  }

  revealInProgress = true
  try {
    cancelScheduledPrewarm()
    mountContent()
    await nextTick()
    await waitForFrames()

    isContentRevealed.value = true
    dispatchNavPageReveal(true)

    if (!shouldScroll) {
      return
    }

    contentRef.value?.scrollIntoView({
      behavior: 'smooth',
      block: 'start',
    })
  } finally {
    revealInProgress = false
  }
}

function syncRevealByViewport() {
  if (window.innerWidth < 768) {
    void revealContent(false, true)
  }
}

function handleViewportChange(event: MediaQueryListEvent) {
  if (event.matches) {
    void revealContent(false, true)
  }
}

const handleRevealScroll = throttle(() => {
  if (window.scrollY > 24) {
    void revealContent()
  }
}, 120)

const handleRevealScrollEnd = debounce(() => {
  if (window.scrollY > 24) {
    void revealContent()
  }
}, 180)

function handleScroll() {
  handleRevealScroll()
  handleRevealScrollEnd()
}

function handleResize() {}

function handleWheel(event: WheelEvent) {
  if (event.deltaY > 8) {
    void revealContent()
  }
}

function handleTouchStart(event: TouchEvent) {
  touchStartY = event.touches[0]?.clientY ?? 0
}

function handleTouchMove(event: TouchEvent) {
  const currentY = event.touches[0]?.clientY ?? touchStartY
  if (touchStartY - currentY > 12) {
    void revealContent()
  }
}

function handleKeydown(event: KeyboardEvent) {
  if (['ArrowDown', 'PageDown', 'Space', 'End'].includes(event.code)) {
    void revealContent()
  }
}

onMounted(() => {
  dispatchNavPageReveal(false)
  syncRevealByViewport()
  scheduleContentPrewarm()

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
  cancelScheduledPrewarm()
  mobileMediaQuery?.removeEventListener('change', handleViewportChange)
  window.removeEventListener('scroll', handleScroll)
  window.removeEventListener('resize', handleResize)
  window.removeEventListener('wheel', handleWheel)
  window.removeEventListener('touchstart', handleTouchStart)
  window.removeEventListener('touchmove', handleTouchMove)
  window.removeEventListener('keydown', handleKeydown)
})
</script>
