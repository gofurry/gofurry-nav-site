<template>
  <div
      v-if="site"
      ref="popoverRef"
      class="site-popover"
      :class="popoverClasses"
      :style="popoverStyle"
      @mouseenter="onMouseEnter"
      @mouseleave="onMouseLeave"
  >
    <h3 class="site-popover__title">
      {{ site.name || '' }}
    </h3>
    <p class="site-popover__desc">
      {{ site.info || '' }}
    </p>

    <div v-if="domains.length" class="site-popover__domains space-y-1">
      <div
          v-for="domain in domains"
          :key="domain"
          class="site-popover__domain"
          @click.stop="goSite(domain)"
      >
        <div class="site-popover__domain-name">
          <img
              :src="pingData[domain]?.status === 'up' ? greenCircle : redCircle"
              class="w-3 h-3"
              alt=""
          />
          <span class="truncate">{{ domain }}</span>
        </div>
        <div class="site-popover__domain-metrics">
          {{ pingData[domain]?.loss || '-' }}% / {{ pingData[domain]?.delay || '-' }}
        </div>
      </div>
    </div>

    <div v-else class="site-popover__empty">
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
    ? 'site-popover--hidden-bottom'
    : 'site-popover--hidden-top'
  return isReady.value
    ? 'site-popover--visible'
    : hiddenTransform
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
