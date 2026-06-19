<template>
  <div class="fixed top-40 sm:top-25 right-4 z-[10000]">
    <!-- 搜索主体 -->
    <div
        class="relative flex items-center
             bg-orange-900 hover:bg-orange-800
             text-white shadow-lg
             transition-all duration-300 ease-out
             rounded-full overflow-hidden"
        :class="open ? 'w-80 px-4 h-10' : 'w-10 h-10 justify-center cursor-pointer'"
        @click="toggle"
    >
      <img src="@/assets/svgs/search-white.svg" class="w-5 h-5" alt="" />

      <input
          v-if="open"
          v-model="keyword"
          @click.stop
          type="text"
          placeholder="Search news..."
          class="ml-3 w-full bg-transparent
               placeholder-white/70
               outline-none text-sm"
      />
    </div>

    <!-- 搜索结果 -->
    <div
        v-if="open && keyword && results.length"
        class="mt-2 w-80
             bg-orange-50 rounded-xl shadow-xl
             border border-orange-100
             max-h-72 overflow-auto"
    >
      <div
          v-for="item in results"
          :key="item.id"
          class="flex gap-3 px-3 py-2  cursor-pointer
               hover:bg-orange-100 transition"
          @mouseenter="showHover(item, $event)"
          @mouseleave="scheduleHide"
          @click="select(item)"
      >
        <img
            :src="item.header"
            class="w-9 h-9 rounded object-cover shrink-0
                   bg-orange-100"
            alt=""
        />

        <div class="flex-1 min-w-0">
          <p class="text-sm font-medium text-gray-800 truncate">
            {{ item.headline }}
          </p>

          <p class="text-xs text-orange-800 truncate mt-0.5 flex justify-between">
            <span class="max-w-30 truncate">{{ item.name }}</span>
            <span>{{ item.author }}</span>
          </p>
        </div>
      </div>
    </div>


    <!-- Hover 新闻正文 -->
    <Teleport to="body">
      <div
          v-if="hoverNews"
          class="news-hover-popover fixed z-[10001]
               w-[480px] max-h-72 overflow-auto
               bg-orange-50 rounded-xl shadow-xl
               border border-orange-100
               p-4 text-sm text-gray-700 overflow-x-hidden"
          :style="hoverStyle"
          @mouseenter="cancelHide"
          @mouseleave="scheduleHide"
      >
        <div class="font-semibold text-gray-800 mb-2">
          {{ hoverNews.headline }}
        </div>

        <div v-html="hoverNews.content"></div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onUnmounted } from 'vue'
import type {NewsBaseModel} from "@/types/game.ts";

const props = defineProps<{
  newsList: NewsBaseModel[]
}>()

const emit = defineEmits<{
  (e: 'select', news: NewsBaseModel): void
}>()

const open = ref(false)
const keyword = ref('')
const hoverNews = ref<NewsBaseModel | null>(null)
const hoverStyle = ref<Record<string, string>>({})
let hideTimer: number | null = null

const toggle = () => {
  open.value = !open.value
  if (!open.value) keyword.value = ''
}

// 搜索逻辑
const results = computed(() => {
  if (!keyword.value) return []
  const k = keyword.value.toLowerCase()

  return props.newsList
      .filter(n =>
          n.headline?.toLowerCase().includes(k) ||
          n.content?.toLowerCase().includes(k) ||
          n.author?.toLowerCase().includes(k) ||
          n.name?.toLowerCase().includes(k)
      )
      .slice(0, 10)
})

const showHover = async (news: NewsBaseModel, e?: MouseEvent) => {
  if (hideTimer) {
    clearTimeout(hideTimer)
    hideTimer = null
  }

  hoverNews.value = news

  if (!e) return

  await nextTick()

  const target = e.currentTarget as HTMLElement
  const rect = target.getBoundingClientRect()
  const gap = 12
  const width = 480

  hoverStyle.value = {
    top: `${rect.top + window.scrollY}px`,
    left: `${rect.left - width - gap + window.scrollX}px`,
    position: 'fixed'
  }
}

const scheduleHide = () => {
  hideTimer = window.setTimeout(() => {
    hoverNews.value = null
  }, 150)
}

const cancelHide = () => {
  if (hideTimer) {
    clearTimeout(hideTimer)
    hideTimer = null
  }
}

const select = (news: NewsBaseModel) => {
  emit('select', news)
  hoverNews.value = null
  open.value = false
  keyword.value = ''
}

// ESC 关闭
const onKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Escape') {
    open.value = false
    keyword.value = ''
    hoverNews.value = null
  }
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onUnmounted(() => window.removeEventListener('keydown', onKeydown))
</script>
