<template>
  <div
    v-if="group"
    class="group-popover"
    :style="popoverStyle"
    @mouseenter="onMouseEnter"
    @mouseleave="onMouseLeave"
  >
    {{ group?.info || group?.name || '' }}
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
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

const popoverStyle = ref<Record<string, string>>({
  left: '0px',
  top: '0px',
  display: 'none'
})

function updatePosition() {
  if (!props.visible || !group.value || !props.targetElement) {
    popoverStyle.value.display = 'none'
    return
  }

  const target = props.targetElement
  const w = 256
  const gap = 8
  const safeInset = 12
  const targetRect = target.getBoundingClientRect()
  const popoverHeight = 96

  let left = targetRect.left
  let top = targetRect.bottom + gap

  if (left + w > window.innerWidth - safeInset) {
    left = window.innerWidth - w - safeInset
  }

  left = Math.max(safeInset, left)

  if (top + popoverHeight > window.innerHeight - safeInset) {
    top = Math.max(safeInset, targetRect.top - popoverHeight - gap)
  }

  popoverStyle.value = {
    left: `${left}px`,
    top: `${top}px`,
    display: 'block',
    position: 'fixed',
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

<style scoped>
.group-popover {
  z-index: 95;
  width: 16rem;
  border-radius: 0.875rem;
  border: 1px solid rgba(255, 255, 255, 0.45);
  background: rgba(255, 247, 237, 0.94);
  padding: 0.75rem 0.85rem;
  color: rgba(86, 47, 14, 0.86);
  font-size: 0.875rem;
  line-height: 1.5;
  box-shadow:
    0 18px 42px rgba(87, 48, 10, 0.16),
    inset 0 1px 0 rgba(255, 255, 255, 0.58);
  backdrop-filter: blur(12px);
}

:global(.dark .group-popover) {
  border-color: rgba(255, 255, 255, 0.10);
  background: rgba(15, 23, 42, 0.92);
  color: rgba(226, 232, 240, 0.92);
  box-shadow:
    0 18px 42px rgba(2, 6, 23, 0.42),
    inset 0 1px 0 rgba(255, 255, 255, 0.04);
}
</style>
