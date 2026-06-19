<template>
  <div class="flex min-h-screen w-full flex-col bg-gray-50">
    <NavHeader />
    <main
        v-if="isContentRevealed"
        ref="contentRef"
        class="relative z-10 flex-1 bg-[#f2e3d0]"
        :style="{
        backgroundImage: `url(${bgGrid})`,
        backgroundRepeat: 'repeat'
      }"
    >
      <div class="absolute w-full">
        <NavTransitionBar />
      </div>
      <div class="h-10"></div>
      <NavContent />
      <div class="h-10"></div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref } from 'vue'
import NavHeader from '@/components/nav/NavHeader.vue'
import NavTransitionBar from '@/components/nav/NavTransitionBar.vue'
import NavContent from '@/components/nav/NavContent.vue'
import bgGrid from '@/assets/pngs/bg-grid.png'
import { debounce, throttle } from '@/utils/util'
import { dispatchNavPageReveal, isNavPageRevealLocked } from '@/utils/navPageReveal'

const isContentRevealed = ref(false)
const contentRef = ref<HTMLElement | null>(null)

let touchStartY = 0
let mobileMediaQuery: MediaQueryList | null = null

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
  window.addEventListener('wheel', handleWheel, { passive: true })
  window.addEventListener('touchstart', handleTouchStart, { passive: true })
  window.addEventListener('touchmove', handleTouchMove, { passive: true })
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  dispatchNavPageReveal(true)
  mobileMediaQuery?.removeEventListener('change', handleViewportChange)
  window.removeEventListener('scroll', handleScroll)
  window.removeEventListener('wheel', handleWheel)
  window.removeEventListener('touchstart', handleTouchStart)
  window.removeEventListener('touchmove', handleTouchMove)
  window.removeEventListener('keydown', handleKeydown)
})
</script>
