<template>
  <div
      v-if="group"
      class="absolute z-50 bg-orange-50 border border-gray-200 shadow-lg rounded-lg p-3 w-64 text-sm text-gray-700"
      :style="popoverStyle"
      @mouseenter="onMouseEnter"
      @mouseleave="onMouseLeave"
  >
    {{ group?.info || '' }}
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, computed } from 'vue'
import type { Group } from '@/types/nav'

// Props定义
const props = defineProps<{
  group: Group | null
  targetElement: HTMLElement | null
  visible: boolean
}>()

// 事件定义
const emit = defineEmits<{
  (e: 'mouseenter'): void
  (e: 'mouseleave'): void
}>()

// 获取group
const group = computed(() => props.group || null)

// 悬浮卡片样式
const popoverStyle = ref<Record<string, string>>({
  left: '0px',
  top: '0px',
  display: 'none'
})

function getElementDocumentOffset(el: HTMLElement | null) {
  if (!el) return { top: 0, left: 0 }

  let top = 0
  let left = 0
  let current: HTMLElement | null = el
  while (current) {
    top += current.offsetTop
    left += current.offsetLeft
    current = current.offsetParent as HTMLElement
  }
  return { top, left }
}

function updatePosition() {
  if (!props.visible || !group.value || !props.targetElement) {
    popoverStyle.value.display = 'none'
    return
  }

  const target = props.targetElement
  const w = 256
  const h = 280

  const docOffset = getElementDocumentOffset(target)
  let left = docOffset.left - w
  let top = docOffset.top - h

  // 边界处理
  const minLeft = 8
  const maxLeft = document.body?.scrollWidth ? (document.body.scrollWidth - w - 8) : minLeft
  left = Math.max(minLeft, Math.min(left, maxLeft))

  popoverStyle.value = {
    left: `${left}px`,
    top: `${top}px`,
    display: 'block',
    position: 'absolute',
    zIndex: '50'
  }
}

// 防抖定时器
let resizeTimer: number | null = null

function handleScrollOrResize() {
  if (resizeTimer) clearTimeout(resizeTimer)
  resizeTimer = window.setTimeout(() => {
    updatePosition()
  }, 16)
}

// 事件处理函数
function onMouseEnter() {
  emit('mouseenter')
}

function onMouseLeave() {
  emit('mouseleave')
}

// 监听props变化更新位置
watch([() => props.visible, () => props.targetElement, group], () => {
  updatePosition()
})

// 生命周期
onMounted(() => {
  if (window) {
    window.addEventListener('scroll', handleScrollOrResize, { passive: true })
    window.addEventListener('resize', handleScrollOrResize)
  }
  updatePosition()
})

onUnmounted(() => {
  if (resizeTimer) clearTimeout(resizeTimer)
  if (window) {
    window.removeEventListener('scroll', handleScrollOrResize)
    window.removeEventListener('resize', handleScrollOrResize)
  }
})
</script>