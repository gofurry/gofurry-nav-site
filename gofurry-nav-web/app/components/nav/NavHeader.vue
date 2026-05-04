<template>
  <header
      class="relative flex flex-col h-90 w-full items-center justify-start overflow-hidden px-4 pt-6 shadow-sm md:h-[100vh] md:px-6 md:pt-0"
  >
    <div
        v-if="bgImage"
        class="absolute inset-0 bg-cover bg-center transition-all duration-700"
        :style="{ backgroundImage: `url(${bgImage})` }"
    ></div>
    <div class="relative z-30 flex w-full justify-center md:absolute md:left-1/2 md:top-[8vh] md:w-full md:max-w-[56rem] md:-translate-x-1/2">
      <SearchBox />
    </div>

    <div
        v-if="recentSites.length"
        class="relative z-10 hidden w-full justify-center md:absolute md:left-1/2 md:top-[28vh] md:flex md:w-full md:max-w-[56rem] md:-translate-x-1/2"
    >
      <SiteIconStrip :sites="recentSites" />
    </div>

    <div class="pointer-events-none absolute bottom-6 left-1/2 z-10 hidden -translate-x-1/2 items-center gap-2 text-white/85 md:flex md:flex-col">
      <span class="text-xs font-medium uppercase tracking-[0.28em]">
        {{ t('navHeader.scrollHint') }}
      </span>
      <div class="flex h-6 w-4 items-start justify-center rounded-full border border-white/35 bg-black/10 p-1 backdrop-blur-sm">
        <span class="h-2 w-2 animate-bounce rounded-full bg-white/90"></span>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import SearchBox from './SearchBox.vue'
import SiteIconStrip from './SiteIconStrip.vue'
import { getImageUrl } from '@/utils/api/nav'
import { loadRecentSites, RECENT_SITES_EVENT, type RecentSiteItem } from '@/utils/recentSites'
import {
  CUSTOM_NAV_HEADER_BG_EVENT,
  loadRandomCustomNavHeaderBackground,
} from '@/utils/customNavHeaderBackground'

const { t } = useI18n()
const bgImage = ref<string | null>(null)
const recentSites = ref<RecentSiteItem[]>([])

let fallbackBackgroundUpdater: (() => void) | null = null
let customBgObjectUrl: string | null = null

function revokeCustomBackgroundUrl() {
  if (customBgObjectUrl) {
    URL.revokeObjectURL(customBgObjectUrl)
    customBgObjectUrl = null
  }
}

async function applyBackground() {
  revokeCustomBackgroundUrl()

  try {
    const customBackground = await loadRandomCustomNavHeaderBackground()
    if (customBackground) {
      bgImage.value = customBackground
      customBgObjectUrl = customBackground
      return
    }
  } catch (error) {
    console.error('Load custom nav header background err:', error)
  }

  fallbackBackgroundUpdater?.()
}

function handleCustomBackgroundChange() {
  void applyBackground()
}

function handleResize() {
  if (customBgObjectUrl) {
    return
  }

  fallbackBackgroundUpdater?.()
}

function syncRecentSites() {
  if (!import.meta.client) {
    return
  }

  recentSites.value = loadRecentSites().slice(0, 8)
}

function handleRecentSitesChange() {
  syncRecentSites()
}

onMounted(async () => {
  try {
    const [resizedUrl, normalUrl] = await Promise.all([
      getImageUrl('standard'),
      getImageUrl('mobile'),
    ])

    fallbackBackgroundUpdater = () => {
      bgImage.value = window.innerWidth >= 768 ? resizedUrl : normalUrl
    }

    await applyBackground()
    syncRecentSites()
    window.addEventListener('resize', handleResize)
    window.addEventListener(CUSTOM_NAV_HEADER_BG_EVENT, handleCustomBackgroundChange)
    window.addEventListener(RECENT_SITES_EVENT, handleRecentSitesChange)
    window.addEventListener('storage', handleRecentSitesChange)
  } catch (err) {
    console.error('Get background image URL err:', err)
  }
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  window.removeEventListener(CUSTOM_NAV_HEADER_BG_EVENT, handleCustomBackgroundChange)
  window.removeEventListener(RECENT_SITES_EVENT, handleRecentSitesChange)
  window.removeEventListener('storage', handleRecentSitesChange)
  revokeCustomBackgroundUrl()
})
</script>

<style scoped>
header {
  transition: background 0.4s ease-in-out;
}
</style>
