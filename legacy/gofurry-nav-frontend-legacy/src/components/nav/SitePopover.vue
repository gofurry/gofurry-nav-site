<template>
  <div
      v-if="site"
      ref="popoverRef"
      class="fixed z-9999 bg-orange-50 border border-gray-200 shadow-xl rounded-xl p-4 w-72 text-sm text-gray-700"
      :style="popoverStyle"
      @mouseenter="onMouseEnter"
      @mouseleave="onMouseLeave"
  >
    <h3 class="font-semibold text-base mb-1">
      {{ site.name || '' }}
    </h3>
    <p class="text-gray-600 mb-3 break-words">
      {{ site.info || '' }}
    </p>

    <div v-if="domains.length" class="space-y-1 max-h-44 overflow-y-auto">
      <div
          v-for="domain in domains"
          :key="domain"
          class="flex items-center justify-between text-xs px-2 py-1 rounded hover:bg-orange-100 cursor-pointer"
          @click.stop="goSite(domain)"
      >
        <div class="flex items-center gap-1 truncate max-w-[60%]">
          <img
              :src="pingData[domain]?.status === 'up' ? greenCircle : redCircle"
              class="w-3 h-3"
          />
          <span class="truncate">{{ domain }}</span>
        </div>
        <div class="text-gray-500 shrink-0">
          {{ pingData[domain]?.loss || '-' }}% / {{ pingData[domain]?.delay || '-' }}
        </div>
      </div>
    </div>

    <div v-else class="text-gray-400 text-xs text-center mt-2">
      {{ t("site.siteDnsPanel.none") }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, computed, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import type { Site, Delay } from '@/types/nav'
import greenCircle from '@/assets/svgs/green-circle.svg'
import redCircle from '@/assets/svgs/red-circle.svg'
import { i18n } from "@/main.ts";

const { t } = i18n.global
const router = useRouter()

// Props定义
const props = defineProps<{
  site: Site | null
  targetElement: HTMLElement | null
  visible: boolean
  displayMode: 'sfw' | 'nsfw'
  pingData: Record<string, Delay>
}>()

// 事件定义
const emit = defineEmits<{
  (e: 'mouseenter'): void
  (e: 'mouseleave'): void
  (e: 'get-popover-height', height: number): void
}>()

// 安全获取site
const site = computed(() => props.site || null)

// 悬浮卡片Ref和样式
const popoverRef = ref<HTMLElement | null>(null)
const popoverStyle = ref<Record<string, string>>({
  display: 'block',
  left: '0px',
  top: '0px'
})

// 计算域名列表
const domains = computed(() => {
  if (!site.value) return []
  const siteData = site.value

  if (Array.isArray(siteData.domain)) return siteData.domain

  try {
    const obj = JSON.parse(siteData.domain);
    return Array.isArray(obj?.domain) ? obj.domain : []
  } catch {
    return siteData.domain ? [siteData.domain] : []
  }
})

// 提供全局更新函数给父组件
function setupGlobalUpdate() {
  ;(window as any).sitePopoverUpdate = (position: { left: number; top: number }) => {
    popoverStyle.value = {
      ...popoverStyle.value,
      left: `${position.left}px`,
      top: `${position.top}px`
    }
  }
}

// 获取并传递实际高度给父组件
function sendPopoverHeight() {
  nextTick(() => {
    if (popoverRef.value) {
      const height = popoverRef.value.offsetHeight
      emit('get-popover-height', height)
    }
  })
}

// 跳转站点详情
function goSite(domain: string) {
  if (!site.value) return
  router.push({
    path: `/site/${site.value.id}`,
    query: { domain, mode: props.displayMode }
  })
}

// 事件处理
function onMouseEnter() {
  emit('mouseenter')
}
function onMouseLeave() {
  emit('mouseleave')
}

// 监听props变化
watch([() => props.visible, site], () => {
  if (props.visible && site.value) {
    sendPopoverHeight()
  } else {
    popoverStyle.value.display = 'none'
  }
})

// 生命周期
onMounted(() => {
  setupGlobalUpdate()

  // DOM渲染完成后传递高度
  nextTick(() => {
    sendPopoverHeight()
  })
})

onUnmounted(() => {
  // 清理全局函数
  delete (window as any).sitePopoverUpdate
})
</script>