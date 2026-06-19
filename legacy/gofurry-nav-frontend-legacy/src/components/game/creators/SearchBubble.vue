<template>
  <!-- 固定悬浮搜索 -->
  <div class="fixed top-40 sm:top-25 right-4 z-[10000]">
    <!-- 搜索主体 -->
    <div
        class="relative flex items-center
               bg-orange-900 hover:bg-orange-800
               text-white shadow-lg
               transition-all duration-300 ease-out
               rounded-full overflow-hidden"
        :class="open ? 'w-72 px-4 h-10' : 'w-10 h-10 justify-center cursor-pointer'"
        @click="toggle"
    >
      <!-- 搜索 icon -->
      <img src="@/assets/svgs/search-white.svg" alt="api" class="w-5 h-5" />

      <!-- 输入框 -->
      <input
          v-if="open"
          v-model="keyword"
          @click.stop
          type="text"
          placeholder="Search..."
          class="ml-3 w-full bg-transparent
                 placeholder-white/70
                 outline-none text-sm"
      />
    </div>

    <!-- 搜索结果 -->
    <div
        v-if="open && keyword && results.length"
        class="mt-2 w-72
               bg-orange-50 rounded-xl shadow-xl
               border border-orange-100
               max-h-64 overflow-auto
               relative"
    >
      <!-- 搜索结果项 -->
      <div
          v-for="item in results"
          :key="item.id"
          class="flex gap-3 px-3 py-2 cursor-pointer
                 hover:bg-orange-100 transition"
          @mouseenter="showHover(item, $event)"
          @mouseleave="scheduleHide"
          @click="select(item)"
      >
        <!-- 头像 -->
        <img
            :src="item.avatar"
            class="w-9 h-9 rounded object-cover shrink-0
                   bg-orange-100"
            alt=""
        />

        <!-- 文本信息 -->
        <div class="flex-1 min-w-0">
          <!-- 名称 & 类型 -->
          <div class="flex items-center justify-between gap-2">
            <p class="text-sm font-medium text-gray-800 truncate">
              {{ item.name }}
            </p>
            <span class="text-xs text-orange-800 shrink-0">
              {{ typeMap[item.type] }}
            </span>
          </div>

          <!-- 简介 -->
          <p class="text-xs text-gray-500 truncate">
            {{ item.info }}
          </p>
        </div>
      </div>
    </div>

    <Teleport to="body">
      <div
          v-if="hoverCreator"
          class="fixed z-[10001] w-64 bg-orange-50 rounded-xl shadow-xl
                 border border-orange-100 p-3 text-xs"
          :style="hoverStyle"
          @mouseenter="cancelHide"
          @mouseleave="scheduleHide"
      >
        <!-- Links -->
        <div v-if="hoverCreator.links?.length" class="mb-3 max-h-24 overflow-hidden">
          <p class="font-semibold text-gray-700 mb-1">
            {{ t('game.creator.link') }}
          </p>
          <SiteIconList :items="hoverCreator.links" />
        </div>

        <!-- Contact -->
        <div v-if="hoverCreator.contact?.length">
          <p class="font-semibold text-gray-700 mb-1">
            {{ t('game.creator.contact') }}
          </p>

          <ul class="space-y-1">
            <li
                v-for="c in hoverCreator.contact.slice(0, 3)"
                :key="c.key"
                class="text-gray-600 truncate"
                :title="`${c.key}：${c.value}`"
            >
              {{ c.key }}：{{ c.value }}
            </li>
          </ul>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import type { CreatorResponse } from '@/types/game'
import SiteIconList from "@/components/common/SiteIconList.vue";
import { i18n } from '@/main'

const { t } = i18n.global

const props = defineProps<{
  creators: CreatorResponse[]
}>()

const emit = defineEmits<{
  (e: 'select', creator: CreatorResponse): void
}>()

const open = ref(false)
const keyword = ref('')
const hoverCreator = ref<CreatorResponse | null>(null)
const hoverStyle = ref<Record<string, string>>({})
let hideTimer: number | null = null

const toggle = () => {
  open.value = !open.value
  if (!open.value) keyword.value = ''
}

const typeMap = computed<Record<string, string>>(() => ({
  1: t('game.creator.curator'),
  2: t('game.creator.blogger'),
  3: t('game.creator.developer'),
  4: t('game.creator.publisher'),
  5: t('game.creator.translator'),
  6: t('game.creator.contentCreator')
}))

// 搜索结果
const results = computed(() => {
  if (!keyword.value) return []
  const k = keyword.value.toLowerCase()
  return props.creators
      .filter(c => c.name?.toLowerCase().includes(k) || c.info?.toLowerCase().includes(k))
      .slice(0, 8)
})

// Hover 卡片显示
const showHover = async (creator: CreatorResponse, e?: MouseEvent) => {
  if (hideTimer) {
    clearTimeout(hideTimer)
    hideTimer = null
  }
  hoverCreator.value = creator

  console.log(111)
  if (e) {
    console.log(222)
    await nextTick()
    const target = e.currentTarget as HTMLElement
    const rect = target.getBoundingClientRect()
    const gap = 8
    const cardWidth = 256

    hoverStyle.value = {
      top: `${rect.top + window.scrollY}px`,
      left: `${rect.left - cardWidth - gap + window.scrollX}px`,
      position: 'fixed'
    }
  }
}

const scheduleHide = () => {
  hideTimer = window.setTimeout(() => {
    hoverCreator.value = null
  }, 150)
}

const cancelHide = () => {
  if (hideTimer) {
    clearTimeout(hideTimer)
    hideTimer = null
  }
}

const select = (creator: CreatorResponse) => {
  emit('select', creator)
  hoverCreator.value = null
  open.value = false
  keyword.value = ''
}

// ESC 关闭
const onKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Escape') {
    open.value = false
    keyword.value = ''
    hoverCreator.value = null
  }
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onUnmounted(() => window.removeEventListener('keydown', onKeydown))
</script>
