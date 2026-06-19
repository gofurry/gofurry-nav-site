<template>
  <!-- 固定悬浮搜索球 -->
  <div class="fixed top-[280px] md:top-[220px] right-4 z-[10000]">
    <!-- 搜索主体 -->
    <div
        class="relative flex items-center
             bg-slate-800 hover:bg-slate-800/80
             text-white shadow-lg
             transition-all duration-300 ease-out
             rounded-full overflow-hidden"
        :class="open ? 'w-72 px-4 h-10' : 'w-10 h-10 justify-center cursor-pointer'"
        @click="toggle"
    >
      <!-- 搜索 icon -->
      <img src="@/assets/svgs/search-white.svg" alt="search" class="w-5 h-5" />

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
             bg-white rounded-xl shadow-xl
             border border-orange-100
             max-h-64 overflow-auto"
    >
      <div
          v-for="item in results"
          :key="item.id"
          class="flex gap-3 px-3 py-2 cursor-pointer
               hover:bg-orange-50 transition"
          @click="go(item)"
      >
        <!-- 图标 -->
        <img
            :src="`${logoPrefix ? logoPrefix + '/' : ''}${item.icon || defaultLogo}`"
            class="w-9 h-9 rounded object-cover shrink-0 bg-orange-100"
            alt=""
        />

        <!-- 名称 + 描述 -->
        <div class="flex-1 min-w-0">
          <p class="text-sm font-medium text-gray-800 truncate">{{ item.name }}</p>
          <p class="text-xs text-gray-500 truncate mt-0.5">{{ item.info }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {ref, computed, onMounted, onUnmounted} from 'vue'
import type { Site } from '@/types/nav'

const logoPrefix = import.meta.env.VITE_SITE_LOGO_PREFIX_URL || ''
const defaultLogo = 'defaultLogo.svg'

const props = defineProps<{
  items: Site[] // 搜索数据
}>()

const open = ref(false)
const keyword = ref('')

const toggle = () => {
  open.value = !open.value
  if (!open.value) keyword.value = ''
}

// 过滤搜索结果
const results = computed(() => {
  if (!keyword.value) return []
  const k = keyword.value.toLowerCase()
  return props.items
      .filter(i => i.name?.toLowerCase().includes(k) || i.info?.toLowerCase().includes(k))
      .slice(0, 8)
})

// 点击跳转
const go = (item: Site) => {
  const domain = Array.isArray(item.domain)
      ? item.domain[0]
      : item.domain
          ? JSON.parse(item.domain)?.domain?.[0] || item.domain
          : ''
  if (domain) window.open(`https://${domain}`, '_blank')
  open.value = false
  keyword.value = ''
}

// ESC 关闭
const onKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Escape') {
    open.value = false
    keyword.value = ''
  }
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onUnmounted(() => window.removeEventListener('keydown', onKeydown))
</script>
