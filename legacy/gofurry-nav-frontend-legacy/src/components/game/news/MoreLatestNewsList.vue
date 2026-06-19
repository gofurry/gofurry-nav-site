<template>
  <div
      ref="containerRef"
      class="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 2xl:grid-cols-5"
  >
    <div
        v-for="(item, idx) in list"
        :key="idx"
        class="cursor-pointer rounded-xl bg-orange-50 p-4 transition hover:bg-orange-100"
        @mouseenter="showPopover(item, $event)"
        @mouseleave="scheduleHide"
        @click="open(item.url)"
    >
      <img
          :src="item.header"
          class="mb-3 rounded-lg object-cover"
          alt=""
      />

      <p class="line-clamp-1 text-sm font-semibold text-orange-900">
        {{ item.name }}
      </p>

      <p class="mt-1 h-10 line-clamp-2 text-xs text-gray-600">
        {{ item.headline }}
      </p>

      <div class="mt-3 flex justify-between text-xs text-gray-500">
        <span>{{ item.author }}</span>
        <span>{{ item.post_time }}</span>
      </div>
    </div>

    <Teleport to="body">
      <Transition>
        <div
            v-if="hoverNews"
            class="news-hover-popover fixed z-50 max-h-[60vh] w-90 overflow-y-auto rounded-lg bg-orange-100 p-4 text-sm shadow-lg backdrop-blur-md sm:w-120"
            :style="popoverStyle"
            @mouseenter="cancelHide"
            @mouseleave="scheduleHide"
        >
          <div v-html="hoverNews.content"></div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref } from 'vue'
import type { NewsBaseModel } from '@/types/game'

defineProps<{ list: NewsBaseModel[] }>()

const containerRef = ref<HTMLElement | null>(null)
const hoverNews = ref<NewsBaseModel | null>(null)
const popoverStyle = ref<Record<string, string>>({})

let currentTarget: HTMLElement | null = null
let hideTimer: number | null = null

function updatePopoverPosition() {
  if (!hoverNews.value || !currentTarget) return

  const targetRect = currentTarget.getBoundingClientRect()
  const cardWidth = currentTarget.offsetWidth
  const cardHeight = currentTarget.offsetHeight
  const popoverWidth = window.innerWidth < 640 ? Math.min(window.innerWidth - 36, 360) : 480
  const marginX = 18
  const marginY = 12

  const popEl = document.querySelector('.news-hover-popover') as HTMLElement | null
  const popoverHeight = popEl?.offsetHeight ?? 200

  let left = targetRect.left + cardWidth / 2 - popoverWidth / 2
  let top = targetRect.top + cardHeight + marginY

  const viewportWidth = window.innerWidth
  const viewportHeight = window.innerHeight

  if (left < marginX) {
    left = marginX
  }

  if (left + popoverWidth > viewportWidth - marginX) {
    left = viewportWidth - popoverWidth - marginX
  }

  if (top + popoverHeight > viewportHeight - marginY) {
    top = targetRect.top - popoverHeight - marginY
  }

  if (top < marginY) {
    top = marginY
  }

  popoverStyle.value = {
    left: `${left}px`,
    top: `${top}px`,
    maxWidth: `${popoverWidth}px`,
    position: 'fixed',
    zIndex: '9999',
    pointerEvents: 'auto',
  }
}

function showPopover(news: NewsBaseModel, event: MouseEvent) {
  cancelHide()
  hoverNews.value = news
  currentTarget = event.currentTarget as HTMLElement
  nextTick(updatePopoverPosition)
}

function scheduleHide() {
  hideTimer = window.setTimeout(() => {
    hoverNews.value = null
    currentTarget = null
  }, 200)
}

function cancelHide() {
  if (hideTimer) {
    clearTimeout(hideTimer)
    hideTimer = null
  }
}

function open(url: string) {
  window.open(url, '_blank', 'noopener')
}

function bindScroll() {
  window.addEventListener('scroll', updatePopoverPosition, { passive: true })
  document.addEventListener('scroll', updatePopoverPosition, {
    passive: true,
    capture: true,
  })
  window.addEventListener('resize', updatePopoverPosition)

  if (containerRef.value) {
    containerRef.value.addEventListener('scroll', updatePopoverPosition, {
      passive: true,
    })
  }
}

function unbindScroll() {
  window.removeEventListener('scroll', updatePopoverPosition)
  document.removeEventListener('scroll', updatePopoverPosition, { capture: true })
  window.removeEventListener('resize', updatePopoverPosition)

  if (containerRef.value) {
    containerRef.value.removeEventListener('scroll', updatePopoverPosition)
  }
}

onMounted(bindScroll)
onUnmounted(unbindScroll)
</script>
