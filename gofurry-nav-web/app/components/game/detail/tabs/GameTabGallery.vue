<template>
  <div class="game-detail-gallery space-y-6">

    <!-- 主展示区 -->
    <div class="game-detail-media-stage relative aspect-video w-full overflow-hidden">
      <!-- 视频 -->
      <video
          v-if="activeMedia?.type === 'movie'"
          ref="videoRef"
          controls
          :muted="isBlocked"
          :autoplay="false"
          :poster="steamAssetUrl(activeMedia.thumb)"
          preload="metadata"
          playsinline
          class="game-detail-video h-full w-full object-contain"
          @loadeddata="markVideoReady"
          @canplay="markVideoReady"
          @playing="markVideoReady"
          @error="markVideoError"
      />
      <div
          v-if="showVideoPlaceholder"
          class="game-detail-video-loading pointer-events-none absolute inset-0 flex items-center justify-center"
          aria-hidden="true"
      >
        <SteamAssetImage
            v-if="activeMedia?.thumb"
            :src="activeMedia.thumb"
            :alt="mediaAlt(activeMedia)"
            class="game-detail-video-loading__poster absolute inset-0 h-full w-full object-cover"
        />
        <div class="game-detail-video-loading__scrim absolute inset-0"></div>
        <div class="game-detail-video-loading__spinner relative"></div>
      </div>

      <!-- 图片 -->
      <SteamAssetImage
          v-else-if="activeMedia?.type === 'screenshot'"
          :src="activeMedia.src"
          :alt="mediaAlt(activeMedia)"
          class="game-detail-media-image h-full w-full cursor-pointer object-contain"
          @click="openFullscreen = true"
      />

      <!-- 无内容 -->
      <div v-else class="game-detail-empty flex h-full w-full items-center justify-center">
        {{t("game.panel.none")}}
      </div>
    </div>

    <!-- 缩略图轮播 -->
    <div class="game-detail-thumb-grid flex flex-wrap gap-2 py-2">
      <div
          v-for="item in mediaList"
          :key="item.key"
          @click="selectMedia(item)"
          :class="[
            'game-detail-thumb relative h-18 w-32 flex-shrink-0 cursor-pointer overflow-hidden',
            activeKey === item.key ? 'game-detail-thumb--active' : 'game-detail-thumb--idle'
          ]"
      >
        <SteamAssetImage
            :src="item.thumb"
            :alt="mediaAlt(item)"
            class="h-full w-full object-fill transition-transform duration-200"
            loading="lazy"
            decoding="async"
        />

        <div
            v-if="item.type === 'movie'"
            class="game-detail-play-overlay absolute inset-0 flex items-center justify-center"
        >
          <svg viewBox="0 0 24 24" class="game-detail-play-icon h-6 w-6">
            <path d="M8 5v14l11-7z" />
          </svg>
        </div>
      </div>
    </div>

    <!-- 图片全屏弹窗 -->
    <div
        v-if="openFullscreen && activeMedia?.type === 'screenshot'"
        class="game-detail-lightbox fixed inset-0 z-50 flex items-center justify-center p-4"
        @click.self="openFullscreen = false"
    >
      <SteamAssetImage
          :src="activeMedia.src"
          :alt="mediaAlt(activeMedia)"
          class="max-h-full max-w-full object-contain"
      />
      <button
          class="game-detail-lightbox__close absolute right-4 top-4 text-2xl"
          @click="openFullscreen = false"
      >
        ×
      </button>
    </div>

  </div>
</template>

<script setup lang="ts">
import {ref, computed, watch, onMounted, onBeforeUnmount, nextTick} from 'vue'
import { i18n } from '@/main'
import SteamAssetImage from '@/components/common/SteamAssetImage.vue'
import { preferredSteamSharedAssetUrl } from '@/utils/steamAssets'

const { t, locale } = i18n.global

export interface MoviesModel {
  id: number
  name: string
  thumbnail: string
  dash_h264?: string
  hls_h264: string
  mp4_url?: string
  webm_url?: string
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
      src: playableMovieSource(m),
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
const videoReady = ref(false)
const videoLoadError = ref(false)
const showVideoPlaceholder = computed(() =>
    activeMedia.value?.type === 'movie' && !videoReady.value && !videoLoadError.value
)

// HLS 播放控制
const videoRef = ref<HTMLVideoElement | null>(null)
let hls: import('hls.js').default | null = null
let hlsModulePromise: Promise<typeof import('hls.js')> | null = null
let videoLoadToken = 0

async function loadHlsModule() {
  if (!hlsModulePromise) {
    hlsModulePromise = import('hls.js')
  }
  return hlsModulePromise
}

function stopVideo() {
  videoLoadToken += 1
  resetVideoState()
  hls?.destroy()
  hls = null
  if (videoRef.value) {
    videoRef.value.pause()
    videoRef.value.removeAttribute('src')
    videoRef.value.src = ''
    videoRef.value.load()
  }
}

async function initVideo(movie: MoviesModel) {
  if (isBlocked.value) return
  if (!videoRef.value) return

  const source = steamAssetUrl(playableMovieSource(movie))
  resetVideoState()
  if (!source) {
    markVideoError()
    return
  }

  const currentToken = videoLoadToken + 1
  videoLoadToken = currentToken
  hls?.destroy()
  hls = null
  videoRef.value.removeAttribute('src')
  videoRef.value.src = ''

  const isHlsSource = /\.m3u8(?:$|\?)/i.test(source)

  if (!isHlsSource) {
    videoRef.value.src = source
    videoRef.value.load()
    if (videoRef.value.readyState >= HTMLMediaElement.HAVE_CURRENT_DATA) {
      markVideoReady()
    }
    return
  }

  const { default: Hls } = await loadHlsModule()
  if (currentToken !== videoLoadToken || !videoRef.value) {
    return
  }

  if (Hls.isSupported()) {
    hls = new Hls()
    hls.attachMedia(videoRef.value)
    hls.on(Hls.Events.MEDIA_ATTACHED, () => {
      if (currentToken === videoLoadToken) {
        hls?.loadSource(source)
      }
    })
    hls.on(Hls.Events.MANIFEST_PARSED, () => {
      videoRef.value?.play().catch(() => {})
    })
    hls.on(Hls.Events.ERROR, (_, data) => {
      if (data.fatal) {
        hls?.destroy()
        hls = null
        markVideoError()
      }
    })
    return
  }

  if (videoRef.value.canPlayType('application/vnd.apple.mpegurl')) {
    videoRef.value.src = source
    videoRef.value.load()
    if (videoRef.value.readyState >= HTMLMediaElement.HAVE_CURRENT_DATA) {
      markVideoReady()
    }
    return
  }

  markVideoError()
}

function playableMovieSource(movie: MoviesModel) {
  return movie.hls_h264 || movie.mp4_url || movie.webm_url || ''
}

function steamAssetUrl(url?: string | null) {
  return preferredSteamSharedAssetUrl(url, locale.value) || url || ''
}

function resetVideoState() {
  videoReady.value = false
  videoLoadError.value = false
}

function markVideoReady() {
  videoReady.value = true
  videoLoadError.value = false
}

function markVideoError() {
  videoReady.value = false
  videoLoadError.value = true
}

function mediaAlt(item: MediaItem) {
  const id = item.key.replace(/^(movie|shot)-/, '')
  if (item.type === 'movie') {
    const movie = props.movies?.find(m => `movie-${m.id}` === item.key)
    return movie?.name || `Game video ${id}`
  }
  return `Game screenshot ${id}`
}

function selectMedia(item: MediaItem) {
  activeKey.value = item.key
}

// 监听切换
watch(
    [activeMedia, isBlocked],
    async ([media, blocked]) => {
      if (blocked) {
        // 强制停止
        stopVideo()
        return
      }

      if (media?.type === 'movie') {
        const movie = props.movies?.find(
            m => `movie-${m.id}` === media.key
        )
        if (movie) {
          await nextTick()
          if (activeMedia.value?.key === media.key) {
            void initVideo(movie)
          }
        }
      } else {
        stopVideo()
      }
    },
    { immediate: true, flush: 'post' }
)

watch(isBlocked, blocked => {
  if (!blocked && !activeKey.value && mediaList.value.length > 0) {
    activeKey.value = preferredInitialMedia.value?.key ?? ''
  }
})

const preferredInitialMedia = computed(() =>
    mediaList.value.find(item => item.type === 'screenshot') ?? mediaList.value[0] ?? null
)

// 初始化
onMounted(async () => {
  if (isBlocked.value) return

  const firstMedia = preferredInitialMedia.value
  if (!firstMedia) return

  activeKey.value = firstMedia.key
  await nextTick()
})

onBeforeUnmount(() => {
  stopVideo()
})
</script>
