<template>
  <div ref="containerRef" class="w-full min-h-screen">
    <div class="w-full border-b border-orange-200 bg-orange-100">
      <div class="grid grid-cols-3 text-sm sm:mx-4 sm:flex sm:gap-3">
        <div
            v-for="item in CREATOR_TYPES"
            :key="item.type"
            class="relative cursor-pointer px-1 py-3 text-center text-gray-600 transition hover:text-orange-500"
            :class="activeType === item.type && 'font-semibold text-orange-800'"
            @click="activeType = item.type"
        >
          <span class="flex-1">{{ item.label }}</span>
          <span
              v-if="activeType === item.type"
              class="absolute bottom-px left-1/2 h-0.5 w-full -translate-x-1/2 rounded-full bg-orange-400"
          />
        </div>
      </div>
    </div>

    <div
        class="grid grid-cols-2 gap-4 p-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 xl:grid-cols-7 2xl:grid-cols-10"
    >
      <div
          v-for="creator in filteredCreators"
          :key="creator.id"
          class="flex cursor-pointer flex-col items-center gap-2 rounded-lg bg-orange-50 p-3 transition hover:bg-orange-100"
          @mouseenter="showPopover($event, creator)"
          @mouseleave="scheduleHide"
          @click="goCreator(creator)"
      >
        <img
            :src="creator.avatar"
            class="h-20 w-20 rounded-sm object-cover"
            alt=""
        />
        <p class="w-full truncate text-center text-sm">
          {{ creator.name }}
        </p>
      </div>
    </div>

    <Teleport to="body">
      <CreatorPopover
          v-if="activeCreator"
          :creator="activeCreator"
          :style="popoverStyle"
          @mouseenter="cancelHide"
          @mouseleave="scheduleHide"
      />
    </Teleport>

    <SearchBubble
        :creators="creators"
        @select="goCreator"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { i18n } from '@/main.ts'
import CreatorPopover from '@/components/game/creators/CreatorPopover.vue'
import SearchBubble from '@/components/game/creators/SearchBubble.vue'
import { useLangStore } from '@/store/langStore'
import type { CreatorResponse } from '@/types/game'
import { getGameCreator } from '@/utils/api/game'

const { t } = i18n.global

interface CreatorType {
  type: number
  label: string
}

const containerRef = ref<HTMLDivElement | null>(null)
const activeType = ref(1)
const creators = ref<CreatorResponse[]>([])
const activeCreator = ref<CreatorResponse | null>(null)
const popoverStyle = ref<Record<string, string>>({})
const langStore = useLangStore()
const lang = ref(langStore.lang)

let hideTimer: number | null = null
let currentTarget: HTMLElement | null = null

const CREATOR_TYPES = computed<CreatorType[]>(() => [
  { type: 1, label: t('game.creator.curator') },
  { type: 2, label: t('game.creator.blogger') },
  { type: 3, label: t('game.creator.developer') },
  { type: 4, label: t('game.creator.publisher') },
  { type: 5, label: t('game.creator.translator') },
  { type: 6, label: t('game.creator.contentCreator') },
])

const updatePopoverPosition = () => {
  if (!activeCreator.value || !currentTarget) return

  const targetRect = currentTarget.getBoundingClientRect()
  const cardWidth = currentTarget.offsetWidth
  const cardHeight = currentTarget.offsetHeight
  const popoverWidth = 320
  const margin = 12

  let popoverHeight = 140
  const popoverEl = document.querySelector('body .creator-popover') as HTMLElement | null
  if (popoverEl) {
    popoverHeight = popoverEl.offsetHeight
  }

  let left = targetRect.left + cardWidth / 2 - popoverWidth / 2
  let top = targetRect.top + cardHeight + margin

  const viewportWidth = window.innerWidth
  const viewportHeight = window.innerHeight

  if (left < margin) {
    left = margin
  }

  if (left + popoverWidth > viewportWidth - margin) {
    left = viewportWidth - popoverWidth - margin
  }

  if (top + popoverHeight > viewportHeight - margin) {
    top = targetRect.top - popoverHeight - margin
  }

  if (top < margin) {
    top = margin
  }

  popoverStyle.value = {
    left: `${left}px`,
    top: `${top}px`,
    position: 'fixed',
    zIndex: '9999',
    pointerEvents: 'auto',
  }
}

const bindAllScrollEvents = () => {
  window.addEventListener('scroll', updatePopoverPosition, { passive: true })
  document.addEventListener('scroll', updatePopoverPosition, { passive: true, capture: true })
  if (containerRef.value) {
    containerRef.value.addEventListener('scroll', updatePopoverPosition, { passive: true })
  }
  window.addEventListener('resize', updatePopoverPosition)
}

const unbindAllScrollEvents = () => {
  window.removeEventListener('scroll', updatePopoverPosition)
  document.removeEventListener('scroll', updatePopoverPosition, { capture: true })
  if (containerRef.value) {
    containerRef.value.removeEventListener('scroll', updatePopoverPosition)
  }
  window.removeEventListener('resize', updatePopoverPosition)
}

const loadCreators = async () => {
  try {
    const res = await getGameCreator(lang.value)
    creators.value = Array.isArray(res) ? res : []
  } catch (err) {
    console.error('get creator list fail', err)
    creators.value = []
  }
}

const filteredCreators = computed(() =>
  creators.value
    .filter((creator) => creator.type === activeType.value)
    .sort((a, b) => Number(a.id) - Number(b.id))
)

const goCreator = (creator: CreatorResponse) => {
  if (!creator.url) return
  window.open(creator.url, '_blank')
}

const showPopover = (event: MouseEvent, creator: CreatorResponse) => {
  cancelHide()
  activeCreator.value = creator
  currentTarget = event.currentTarget as HTMLElement
  nextTick(() => {
    updatePopoverPosition()
  })
}

const scheduleHide = () => {
  hideTimer = window.setTimeout(() => {
    activeCreator.value = null
    currentTarget = null
  }, 200)
}

const cancelHide = () => {
  if (hideTimer) {
    clearTimeout(hideTimer)
    hideTimer = null
  }
}

onMounted(async () => {
  await loadCreators()
  await nextTick()
  bindAllScrollEvents()
})

onUnmounted(() => {
  unbindAllScrollEvents()
})

watch(
  () => langStore.lang,
  async (newLang) => {
    lang.value = newLang
    await loadCreators()
  }
)
</script>
