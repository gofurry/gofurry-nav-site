<template>
  <div class="space-y-6">

    <!-- 主展示区 -->
    <div class="w-full rounded-xl overflow-hidden bg-black shadow-md relative aspect-video">
      <!-- 视频 -->
      <video
          v-if="activeMedia?.type === 'movie'"
          ref="videoRef"
          controls
          :muted="isBlocked"
          :autoplay="false"
          preload="none"
          class="w-full h-full object-contain bg-black"
      />

      <!-- 图片 -->
      <img
          v-else-if="activeMedia?.type === 'screenshot'"
          :src="activeMedia.src"
          alt="screenshot"
          class="w-full h-full object-contain bg-gray-100 cursor-pointer"
          @click="openFullscreen = true"
      />

      <!-- 无内容 -->
      <div v-else class="w-full h-full flex items-center justify-center text-gray-500">
        {{t("game.panel.none")}}
      </div>
    </div>

    <!-- 缩略图轮播 -->
    <div class="flex gap-2 overflow-x-auto py-2">
      <div
          v-for="item in mediaList"
          :key="item.key"
          @click="activeKey = item.key"
          :class="['flex-shrink-0 rounded-lg overflow-hidden cursor-pointer border-2',
                 activeKey === item.key ? 'border-orange-500' : 'border-transparent']"
          class="w-32 h-18 relative"
      >
        <img
            :src="item.thumb"
            alt="thumb"
            class="w-full h-full object-fill transition-transform duration-200 group-hover:scale-105"
        />

        <div
            v-if="item.type === 'movie'"
            class="absolute inset-0 flex items-center justify-center bg-black/20"
        >
          <svg viewBox="0 0 24 24" class="w-6 h-6 fill-white">
            <path d="M8 5v14l11-7z" />
          </svg>
        </div>
      </div>
    </div>

    <!-- 图片全屏弹窗 -->
    <div
        v-if="openFullscreen && activeMedia?.type === 'screenshot'"
        class="fixed inset-0 z-50 bg-black/90 flex items-center justify-center p-4"
        @click.self="openFullscreen = false"
    >
      <img
          :src="activeMedia.src"
          alt="fullscreen"
          class="max-h-full max-w-full object-contain"
      />
      <button
          class="absolute top-4 right-4 text-white text-2xl"
          @click="openFullscreen = false"
      >
        ×
      </button>
    </div>

  </div>
</template>

<script setup lang="ts">
import {ref, computed, watch, onMounted, onBeforeUnmount, nextTick} from 'vue'
import Hls from 'hls.js'
import { i18n } from '@/main.ts'

const { t } = i18n.global

export interface MoviesModel {
  id: number
  name: string
  thumbnail: string
  hls_h264: string
}

export interface ScreenshotsModel {
  id: number
  path_thumbnail: string
  path_full: string
}

const props = defineProps<{
  movies: MoviesModel[] | null
  screenshots: ScreenshotsModel[] | null
  blocked?: boolean
}>()

const isBlocked = computed(() => props.blocked)

// 构建媒体列表
type MediaItem = {
  key: string
  type: 'movie' | 'screenshot'
  src: string
  thumb: string
}

const mediaList = computed<MediaItem[]>(() => {
  const list: MediaItem[] = []

  props.movies?.forEach(m => {
    list.push({
      key: `movie-${m.id}`,
      type: 'movie',
      src: m.hls_h264,
      thumb: m.thumbnail
    })
  })

  props.screenshots?.forEach(s => {
    list.push({
      key: `shot-${s.id}`,
      type: 'screenshot',
      src: s.path_full,
      thumb: s.path_thumbnail
    })
  })

  return list
})

// 当前选中
const activeKey = ref<string | null>(null)
const activeMedia = computed(() =>
    mediaList.value.find(m => m.key === activeKey.value) ?? null
)
const openFullscreen = ref(false)

// HLS 播放控制
const videoRef = ref<HTMLVideoElement | null>(null)
let hls: Hls | null = null

function initVideo(movie: MoviesModel) {
  if (isBlocked.value) return
  if (!videoRef.value) return

  hls?.destroy()
  hls = null
  videoRef.value.src = ''

  if (videoRef.value.canPlayType('application/vnd.apple.mpegurl')) {
    videoRef.value.src = movie.hls_h264
    videoRef.value.load()
    videoRef.value.play().catch(() => {})
  } else if (Hls.isSupported()) {
    hls = new Hls()
    hls.loadSource(movie.hls_h264)
    hls.attachMedia(videoRef.value)
    hls.on(Hls.Events.MANIFEST_PARSED, () => {
      videoRef.value?.play().catch(() => {})
    })
  }
}

// 监听切换
watch(
    [activeMedia, isBlocked],
    ([media, blocked]) => {
      if (blocked) {
        // 强制停止
        hls?.destroy()
        hls = null
        if (videoRef.value) {
          videoRef.value.pause()
          videoRef.value.src = ''
        }
        return
      }

      if (media?.type === 'movie') {
        const movie = props.movies?.find(
            m => `movie-${m.id}` === media.key
        )
        if (movie) initVideo(movie)
      } else {
        hls?.destroy()
        hls = null
        if (videoRef.value) videoRef.value.src = ''
      }
    },
    { immediate: true }
)

watch(isBlocked, blocked => {
  if (!blocked && !activeKey.value && mediaList.value.length > 0) {
    activeKey.value = mediaList.value[0]?.key ?? ''
  }
})

// 初始化
onMounted(async () => {
  if (isBlocked.value) return

  const firstMedia = mediaList.value?.[0]
  if (!firstMedia) return

  activeKey.value = firstMedia.key
  await nextTick()

  if (firstMedia.type === 'movie') {
    const movie = props.movies?.find(
        m => `movie-${m.id}` === firstMedia.key
    )
    if (movie) initVideo(movie)
  }
})

onBeforeUnmount(() => {
  hls?.destroy()
})
</script>

<style scoped>
::-webkit-scrollbar {
  height: 8px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: rgba(249, 115, 22, 0.4);
  border-radius: 4px;
  backdrop-filter: blur(4px);
}

::-webkit-scrollbar-thumb:hover {
  background: rgba(249, 115, 22, 0.7);
}
</style>
