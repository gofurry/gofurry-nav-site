<template>
  <div class="pointer-events-none fixed inset-x-0 top-10 z-[70] flex justify-center px-3 md:top-2 sm:px-6">
    <div
        class="pointer-events-auto relative w-full transition-all duration-1000"
        :class="isNavCollapsed ? 'max-w-[132px]' : 'max-w-[400px] sm:max-w-[540px] xl:max-w-[920px]'"
    >
      <nav
          class="mx-auto flex items-center gap-3 border border-white/20 bg-[rgba(18,24,37,0.55)] text-gray-100 shadow-lg ring-1 ring-white/10 backdrop-blur-xl transition-all duration-300"
          :class="isNavCollapsed
            ? 'justify-between rounded-lg px-1 py-1'
            : 'rounded-lg px-1 py-1 md:px-3 md:py-2 sm:px-5'"
      >
        <div class="flex shrink-0 items-center justify-center">
          <img :src="logo" alt="gofurry logo" class="w-10 h-10" />
        </div>

        <div
            class="hidden min-w-0 flex-1 items-center justify-center overflow-hidden transition-[max-width] duration-300 sm:flex"
            :class="isNavCollapsed ? 'max-w-0' : 'max-w-[640px]'"
        >
          <div
              class="flex min-w-max items-center justify-center gap-1 transition-opacity duration-200"
              :class="isNavCollapsed ? 'pointer-events-none opacity-0' : 'opacity-100'"
              :style="{ transitionDelay: isNavCollapsed ? '0ms' : '180ms' }"
          >
            <component
                :is="link.external ? 'a' : RouterLink"
                v-for="link in navLinks"
                :key="link.label"
                v-bind="link.external
                  ? { href: link.href, target: '_blank', rel: 'noopener noreferrer' }
                  : { to: link.to }"
                class="rounded-lg px-4 py-2 text-sm font-medium whitespace-nowrap transition-all duration-200"
                :class="isActive(link)
                  ? 'bg-white/10 text-slate-900'
                  : 'text-gray-100 hover:bg-white/10 hover:text-white'"
            >
              {{ link.label }}
            </component>
          </div>
        </div>

        <div
            class="ml-auto hidden items-center overflow-hidden transition-[max-width] duration-300 xl:flex"
            :class="isNavCollapsed ? 'max-w-0' : 'max-w-[180px]'"
        >
          <div
              class="flex min-w-max items-center gap-2 transition-opacity duration-200"
              :class="isNavCollapsed ? 'pointer-events-none opacity-0' : 'opacity-100'"
              :style="{ transitionDelay: isNavCollapsed ? '0ms' : '180ms' }"
          >
            <div class="flex items-center gap-1">
              <div
                  v-for="option in languageOptions"
                  :key="option.value"
                  class="flex h-6 w-6 items-center justify-center rounded-lg transition cursor-pointer"
                  :class="langStore.lang === option.value
                    ? 'bg-white/30 text-slate-900'
                    : 'text-gray-200 hover:bg-white/10'"
                  @click="switchLang(option.value)"
              >
                <img :src="option.flag" class="h-4 w-4" alt="language" />
              </div>
            </div>

            <div
                class="cursor-pointer flex h-6 w-6 items-center justify-center rounded-lg bg-white/8 text-sm ring-1 ring-white/10 transition hover:bg-white/30"
                @click="showModeModal = true"
            >
              <img :src="gear" class="h-4 w-4" alt="mode" />
            </div>
          </div>
        </div>

        <button
            v-if="isNavCollapsed"
            class="inline-flex h-10 w-10 items-center justify-center rounded-lg border border-white/10 bg-white/8 text-white transition hover:bg-white/14"
            :title="t('navbar.expandNav')"
            :aria-label="t('navbar.expandNav')"
            @click="expandNav"
        >
          <svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>

        <div
            v-else
            class="cursor-pointer ml-auto inline-flex h-10 w-10 items-center justify-center rounded-lg border border-white/10 bg-white/8 text-white transition hover:bg-white/14 xl:hidden"
            :aria-expanded="mobileMenuOpen"
            :aria-label="t('navbar.expandNav')"
            @click="mobileMenuOpen = !mobileMenuOpen"
        >
          <svg v-if="!mobileMenuOpen" class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
          </svg>
          <svg v-else class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 6l12 12M6 18L18 6" />
          </svg>
        </div>
      </nav>

      <transition name="mobile-menu">
        <div
            v-if="mobileMenuOpen && !isNavCollapsed"
            class="mt-3 overflow-hidden rounded-lg border border-white/15 bg-[rgba(18,24,37,0.75)] p-3 text-gray-100 shadow-[0_24px_60px_rgba(15,23,42,0.34)] ring-1 ring-white/10 backdrop-blur-xl xl:hidden"
        >
          <div class="flex flex-col gap-2">
            <component
                :is="link.external ? 'a' : RouterLink"
                v-for="link in navLinks"
                :key="link.label"
                v-bind="link.external
                  ? { href: link.href, target: '_blank', rel: 'noopener noreferrer' }
                  : { to: link.to }"
                class="rounded-lg px-4 py-3 text-sm font-medium transition-colors"
                :class="isActive(link)
                  ? 'bg-white/50'
                  : 'bg-white/5 hover:bg-white/10'"
                @click="mobileMenuOpen = false"
            >
              <div :class="isActive(link)
                  ? 'text-slate-900'
                  : 'text-gray-300'">{{ link.label }}</div>
            </component>
          </div>

          <div class="mt-4 flex flex-col gap-3 border-t border-white/10 pt-4">
            <div
                class="flex items-center justify-between rounded-lg bg-white/5 px-4 py-3 text-sm text-gray-100 transition hover:bg-white/10 cursor-pointer"
                @click="openModeModalFromMobile"
            >
              <span class="flex items-center gap-2">
                <img :src="gear" class="h-4 w-4" alt="mode" />
                {{ t('navbar.mode') }}
              </span>
              <span class="text-orange-300">{{ mode || '--' }}</span>
            </div>

            <div class="grid grid-cols-2 gap-2">
              <div
                  v-for="option in languageOptions"
                  :key="option.value"
                  class="flex items-center justify-center gap-2 rounded-lg px-4 py-3 text-sm transition cursor-pointer"
                  :class="langStore.lang === option.value
                    ? 'bg-white/50'
                    : 'bg-white/5 hover:bg-white/10'"
                  @click="switchLang(option.value)"
              >
                <img :src="option.flag" class="h-4 w-4" alt="language" />
                <span :class="langStore.lang === option.value
                    ? 'text-slate-900'
                    : 'text-gray-300'">{{ option.label }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </transition>

      <ModeSettingModal
          :show="showModeModal"
          :mode="mode"
          @cancel="showModeModal = false"
          @save="saveMode"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { useLangStore } from '@/store/langStore'
import { useI18n } from 'vue-i18n'
import cnFlag from '@/assets/flags/cn.svg'
import usFlag from '@/assets/flags/us.svg'
import logo from '@/assets/svgs/logo-mini.svg'
import gear from '@/assets/svgs/gear.svg'
import ModeSettingModal from '@/components/common/ModeSettingModal.vue'
import { debounce, throttle } from '@/utils/util'
import { readMode, subscribeModeChange, writeMode } from '@/utils/modeStorage'

const { t } = useI18n()
const route = useRoute()
const langStore = useLangStore()

const showModeModal = ref(false)
const mobileMenuOpen = ref(false)
const mode = ref('')
const isNavCollapsed = ref(false)
const lastScrollY = ref(0)
let stopModeSubscription: (() => void) | null = null

type NavLink = {
  label: string
  to?: string
  href?: string
  external?: boolean
}

const openPlatformLink = computed<NavLink>(() => (
  import.meta.env.PROD
    ? { label: langStore.lang === 'zh' ? '开放平台' : 'Open Platform', href: 'https://open.go-furry.com', external: true }
    : { label: langStore.lang === 'zh' ? '开放平台' : 'Open Platform', to: '/updates' }
))

const navLinks = computed<NavLink[]>(() => [
  { label: t('sidebar.nav'), to: '/nav' },
  { label: t('sidebar.games'), to: '/games' },
  openPlatformLink.value,
  { label: langStore.lang === 'zh' ? '深度兽研' : 'DeepFurry', href: 'https://www.deepfurry.com', external: true },
])

const languageOptions = [
  { value: 'zh' as const, label: 'CN', flag: cnFlag },
  { value: 'en' as const, label: 'EN', flag: usFlag },
]

const isActive = (link: NavLink) =>
    Boolean(link.to && (route.path === link.to || route.path.startsWith(`${link.to}/`)))

onMounted(() => {
  mode.value = readMode()

  lastScrollY.value = window.scrollY
  window.addEventListener('scroll', handleScroll, { passive: true })
  stopModeSubscription = subscribeModeChange(({ mode: nextMode }) => {
    mode.value = nextMode
  })
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
  stopModeSubscription?.()
})

watch(
    () => route.fullPath,
    () => {
      mobileMenuOpen.value = false
      isNavCollapsed.value = false
    }
)

const handleScrollDirection = throttle(() => {
  const currentScrollY = window.scrollY
  const delta = currentScrollY - lastScrollY.value

  if (Math.abs(delta) < 12) {
    return
  }

  if (currentScrollY <= 40) {
    isNavCollapsed.value = false
    lastScrollY.value = currentScrollY
    return
  }

  if (delta > 0 && currentScrollY > 120) {
    isNavCollapsed.value = true
    mobileMenuOpen.value = false
  } else if (delta < 0) {
    isNavCollapsed.value = false
  }

  lastScrollY.value = currentScrollY
}, 120)

const handleScrollEnd = debounce(() => {
  if (window.scrollY <= 40) {
    isNavCollapsed.value = false
  }
}, 180)

function handleScroll() {
  handleScrollDirection()
  handleScrollEnd()
}

function saveMode(value: string) {
  showModeModal.value = false
  const trimmed = value.trim().slice(0, 32)
  writeMode(trimmed)
}

function switchLang(lang: 'zh' | 'en') {
  langStore.setLang(lang)
  mobileMenuOpen.value = false
}

function openModeModalFromMobile() {
  mobileMenuOpen.value = false
  showModeModal.value = true
}

function expandNav() {
  isNavCollapsed.value = false
}
</script>

<style scoped>
.mobile-menu-enter-active,
.mobile-menu-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.mobile-menu-enter-from,
.mobile-menu-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
