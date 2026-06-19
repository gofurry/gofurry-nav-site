<template>
  <div
      v-if="showBubble"
      ref="bubbleRef"
      class="fixed top-10 right-6 z-50 group"
      style="touch-action: none;"
  >
    <div
        class="relative w-20 h-20 rounded-full bg-gray-300 hover:cursor-pointer
             border-4 border-gray-900/20 hover:border-gray-500/70
             overflow-hidden
             flex items-center justify-center
             transition-all duration-300
             group-hover:scale-110
             select-none"
    >
      <img
          src="https://qcdn.go-furry.com/game/background/steam.jpg"
          class="w-full h-full object-cover pointer-events-none"
          draggable="false"
      />

      <div
          class="absolute z-10 text-lg font-semibold text-stroke
               opacity-0 translate-y-2
               transition-all duration-300
               group-hover:opacity-100
               group-hover:translate-y-0
               pointer-events-none truncate"
      >
        {{ t('game.lottery.bubble.lottery') }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from "vue-router"
import { ref, onMounted, onBeforeUnmount, nextTick, watch } from "vue"
import { i18n } from '@/main.ts'

const { t } = i18n.global
const router = useRouter()
const bubbleRef = ref<HTMLElement | null>(null)
const showBubble = ref(false)

const STORAGE_KEY = "prize-bubble-position"
const BUBBLE_SIZE = 80 // 气泡宽高 80px
const MIN_LEFT = 100    // 左侧最小距离
const MIN_TOP = 10     // top-10

let isDragging = false
let startX = 0
let startY = 0
let offsetX = 0
let offsetY = 0
let moved = false
let listenersBound = false

function syncBubbleVisibility() {
  const saved = localStorage.getItem('showBubble')
  showBubble.value = saved !== 'false'
}

function goPrize() {
  router.push("/games/prize")
}

// 保存位置
function savePosition() {
  localStorage.setItem(STORAGE_KEY, JSON.stringify({ offsetX, offsetY }))
}

// 加载位置
function loadPosition() {
  const saved = localStorage.getItem(STORAGE_KEY)
  if (!saved) return

  const pos = JSON.parse(saved)
  offsetX = pos.offsetX
  offsetY = pos.offsetY
}

// 应用 transform
function applyTransform() {
  if (!bubbleRef.value) return
  bubbleRef.value.style.transform = `translate(${offsetX}px, ${offsetY}px)`
}

// clamp 工具
function clamp(value: number, min: number, max: number) {
  return Math.min(Math.max(value, min), max)
}

// 限制气泡在屏幕内
function adjustWithinScreen() {
  const maxTop = window.innerHeight - BUBBLE_SIZE - MIN_TOP

  // 因为 offsetX 是从右向左的负值，左侧限制 MIN_LEFT
  const minX = -(window.innerWidth - BUBBLE_SIZE - MIN_LEFT)
  offsetX = clamp(offsetX, minX, 0)
  offsetY = clamp(offsetY, 0, maxTop)

  applyTransform()
}

function onPointerDown(e: PointerEvent) {
  isDragging = true
  moved = false

  startX = e.clientX - offsetX
  startY = e.clientY - offsetY

  bubbleRef.value?.setPointerCapture(e.pointerId)
}

function onPointerMove(e: PointerEvent) {
  if (!isDragging) return

  let x = e.clientX - startX
  let y = e.clientY - startY

  const minX = -(window.innerWidth - BUBBLE_SIZE - MIN_LEFT)
  const maxTop = window.innerHeight - BUBBLE_SIZE - MIN_TOP

  x = clamp(x, minX, 0)
  y = clamp(y, 0, maxTop)

  if (Math.abs(x - offsetX) > 0 || Math.abs(y - offsetY) > 0) moved = true

  offsetX = x
  offsetY = y

  applyTransform()
}

function onPointerUp(e: PointerEvent) {
  isDragging = false
  bubbleRef.value?.releasePointerCapture(e.pointerId)

  if (!moved) {
    goPrize()
    return
  }

  savePosition()
}

function onResize() {
  adjustWithinScreen()
  savePosition()
}

function bindBubbleEvents() {
  const el = bubbleRef.value
  if (!el || listenersBound) return

  el.addEventListener("pointerdown", onPointerDown)
  el.addEventListener("pointermove", onPointerMove)
  el.addEventListener("pointerup", onPointerUp)
  listenersBound = true
}

function unbindBubbleEvents() {
  const el = bubbleRef.value
  if (!el || !listenersBound) return

  el.removeEventListener("pointerdown", onPointerDown)
  el.removeEventListener("pointermove", onPointerMove)
  el.removeEventListener("pointerup", onPointerUp)
  listenersBound = false
}

watch(showBubble, async (visible) => {
  if (visible) {
    await nextTick()
    adjustWithinScreen()
    bindBubbleEvents()
    return
  }

  unbindBubbleEvents()
})

onMounted(async () => {
  syncBubbleVisibility()
  loadPosition()
  window.addEventListener("resize", onResize)
  window.addEventListener("show-bubble-change", syncBubbleVisibility)

  if (showBubble.value) {
    await nextTick()
    adjustWithinScreen()
    bindBubbleEvents()
  }
})

onBeforeUnmount(() => {
  window.removeEventListener("show-bubble-change", syncBubbleVisibility)
  window.removeEventListener("resize", onResize)
  unbindBubbleEvents()
})
</script>

<style scoped>
.text-stroke {
  -webkit-text-stroke: 1px black;
  color: white;
}
</style>
