<template>
  <div
      v-if="site"
      ref="popoverRef"
      class="fixed z-9999 w-72 rounded-xl border border-orange-100/85 bg-orange-50/95 p-4 text-sm text-gray-700 shadow-[0_16px_40px_rgba(25,35,38,0.16)] backdrop-blur-md transition-[opacity,transform,filter] duration-300 ease-[cubic-bezier(0.22,1,0.36,1)] will-change-[opacity,transform,filter] dark:border-white/10 dark:bg-[rgba(15,23,42,0.94)] dark:text-slate-200 dark:shadow-[0_16px_40px_rgba(2,6,23,0.42)]"
      :class="popoverClasses"
      :style="popoverStyle"
      @mouseenter="onMouseEnter"
      @mouseleave="onMouseLeave"
  >
    <h3 class="font-semibold text-base mb-1 text-stone-900 dark:text-slate-100">
      {{ site.name || '' }}
    </h3>
    <p class="text-gray-600 mb-3 break-words dark:text-slate-400">
      {{ site.info || '' }}
    </p>

    <div v-if="domains.length" class="space-y-1 max-h-44 overflow-y-auto">
      <div
          v-for="domain in domains"
          :key="domain"
          class="flex items-center justify-between text-xs px-2 py-1 rounded hover:bg-orange-100 cursor-pointer dark:hover:bg-white/8"
          @click.stop="goSite(domain)"
      >
        <div class="flex items-center gap-1 truncate max-w-[60%]">
          <img
              :src="pingData[domain]?.status === 'up' ? greenCircle : redCircle"
              class="w-3 h-3"
              alt=""
          />
          <span class="truncate">{{ domain }}</span>
        </div>
        <div class="text-gray-500 shrink-0 dark:text-slate-400">
          {{ pingData[domain]?.loss || '-' }}% / {{ pingData[domain]?.delay || '-' }}
        </div>
      </div>
    </div>

    <div v-else class="text-gray-400 text-xs text-center mt-2 dark:text-slate-500">
      {{ t("site.siteDnsPanel.none") }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import type { Site, Delay } from '@/types/nav'
import greenCircle from '@/assets/svgs/green-circle.svg'
import redCircle from '@/assets/svgs/red-circle.svg'
import { i18n } from "@/main";
import { siteDetailPath } from '@/utils/siteRoutes'

const { t } = i18n.global
const router = useRouter()

// Props定义
const props = defineProps<{
  site: Site | null
  visible: boolean
  position: { left: number; top: number } | null
  placement: 'top' | 'bottom'
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
  left: '0px',
  top: '0px',
})

const isReady = computed(() => props.visible && !!site.value && !!props.position)
const popoverClasses = computed(() => {
  const hiddenTransform = props.placement === 'bottom'
    ? 'translate-y-2 scale-[0.988] blur-[1.5px]'
    : '-translate-y-2 scale-[0.988] blur-[1.5px]'
  return isReady.value
    ? 'pointer-events-auto opacity-100 translate-y-0 scale-100 blur-0'
    : `pointer-events-none opacity-0 ${hiddenTransform}`
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
  router.push(siteDetailPath(site.value.id, domain))
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
  }
})

watch(
  () => props.position,
  (position) => {
    if (!position) {
      return
    }

    popoverStyle.value = {
      left: `${position.left}px`,
      top: `${position.top}px`,
    }
  },
  { immediate: true }
)

// 生命周期
onMounted(() => {
  nextTick(() => {
    sendPopoverHeight()
  })
})
</script>
