<template>
  <div class="nav-home-page flex w-full flex-col">
    <NavHeader
      :desktop-object-key="navPageData.desktopObjectKey"
      :mobile-object-key="navPageData.mobileObjectKey"
    />
    <NavToolDock v-if="dockActive" />
    <div
        ref="contentRef"
        class="nav-content-shell relative z-10 flex-1 overflow-hidden"
    >
      <NavTransitionBar :initial-saying="navPageData.saying" />
      <div class="relative z-10">
        <div class="mx-auto w-full max-w-[2080px] px-4 sm:px-6 xl:px-8">
          <NavContent
            :initial-groups="navPageData.groups"
            :initial-spotlight="navPageData.spotlight"
            :initial-ping-data="navPageData.pingData"
          />
          <div class="h-10"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { getNavHome } from '~/services/nav'
import type { Delay, Group, NavHomeSpotlight, SayingModel } from '~/types/nav'
import NavHeader from '@/components/nav/NavHeader.vue'
import NavToolDock from '@/components/nav/NavToolDock.vue'
import NavTransitionBar from '@/components/nav/NavTransitionBar.vue'
import NavContent from '@/components/nav/NavContent.vue'
import { heroPreferenceKey } from '~/utils/heroPreferences'

interface NavPageData {
  desktopObjectKey: string | null
  mobileObjectKey: string | null
  saying: SayingModel | null
  groups: Group[]
  spotlight: NavHomeSpotlight
  pingData: Record<string, Delay>
}

const dockActive = ref(false)
const contentRef = ref<HTMLElement | null>(null)
const { locale } = useI18n()

function parsePingData(data: Record<string, string | undefined>) {
  const result: Record<string, Delay> = {}

  for (const key in data) {
    const value = data[key]
    if (typeof value === 'string') {
      try {
        result[key] = JSON.parse(value) as Delay
      } catch {
        result[key] = { status: 'down', delay: '-', loss: '-', time: '-' }
      }
    } else {
      result[key] = { status: 'down', delay: '-', loss: '-', time: '-' }
    }
  }

  return result
}

const lang = computed(() => (locale.value === 'en' ? 'en' : 'zh'))
// Capture the SSR selection. Save applies Hero alone, never refetches Home.
const heroSettings = useHeroPreferences()
const initialHeroPreference = { ...heroSettings.preference.value }
const { data } = await useAsyncData<NavPageData>(
  () => `nav-page:${lang.value}:${heroPreferenceKey(initialHeroPreference)}`,
  async () => {
    const selection = { ...heroSettings.preference.value }
    const home = await getNavHome(lang.value, selection)

    return {
      desktopObjectKey: selection.mode === 'local' ? null : home.hero.desktop?.object_key ?? null,
      mobileObjectKey: selection.mode === 'local' ? null : home.hero.mobile?.object_key ?? null,
      saying: home.saying,
      groups: home.groups.sort((a, b) => Number(a.priority) - Number(b.priority)),
      spotlight: home.spotlight,
      pingData: parsePingData(home.ping),
    }
  },
  {
    watch: [lang],
    default: () => ({
      desktopObjectKey: null,
      mobileObjectKey: null,
      saying: null,
      groups: [],
      spotlight: { page_size: 6, featured: [], popular: [], latest: [], random: [] },
      pingData: {},
    }),
  }
)

const navPageData = computed(() => data.value!)

// Observe natural scrolling only; mounting, Footer and Search need no reveal state.
let scrollFrame: number | null = null
let desktopQuery: MediaQueryList | null = null

function syncDock() {
  scrollFrame = null
  dockActive.value = Boolean(desktopQuery?.matches && contentRef.value
    && contentRef.value.getBoundingClientRect().top <= window.innerHeight / 2)
}

function scheduleDockSync() {
  if (scrollFrame === null) scrollFrame = requestAnimationFrame(syncDock)
}

onMounted(() => {
  // Match the existing NavToolDock CSS breakpoint.
  desktopQuery = window.matchMedia('(min-width: 1024px)')
  window.addEventListener('scroll', scheduleDockSync, { passive: true })
  window.addEventListener('resize', scheduleDockSync)
  syncDock()
})

onUnmounted(() => {
  window.removeEventListener('scroll', scheduleDockSync)
  window.removeEventListener('resize', scheduleDockSync)
  if (scrollFrame !== null) cancelAnimationFrame(scrollFrame)
})
</script>
