<template>
  <header
    class="relative z-[70] w-full text-gray-100 shadow-lg backdrop-blur-xl transition-all duration-300"
    :class="headerClass"
  >
    <div
        class="relative mx-auto flex w-full max-w-[1700px] items-center gap-3 px-4 py-2 transition-all duration-300 sm:px-6"
    >
      <NuxtLink
          to="/"
          class="relative z-10 flex shrink-0 items-center gap-2 px-2 py-1"
          @click.stop="closeMenus"
      >
        <img :src="logo" alt="gofurry logo" class="h-10 w-10" />
        <span class="hidden text-sm font-semibold tracking-wide text-white sm:inline">GoFurry</span>
      </NuxtLink>

      <div class="pointer-events-none absolute left-1/2 top-1/2 z-0 flex max-w-[calc(100vw-10rem)] -translate-x-1/2 -translate-y-1/2 items-center justify-center overflow-hidden transition-all duration-300 sm:max-w-[calc(100vw-14rem)] md:max-w-[760px]">
        <nav class="pointer-events-auto flex min-w-0 items-center justify-center gap-1 transition-all duration-200 opacity-100 md:min-w-max md:max-w-[760px]">
          <template v-for="(link, index) in navLinks" :key="link.label">
            <a
                v-if="link.external"
                :href="link.href"
                target="_blank"
                rel="noopener noreferrer"
                class="rounded-lg px-2 py-2 text-sm font-medium whitespace-nowrap text-gray-100 transition-all duration-200 hover:bg-white/10 hover:text-white sm:px-3 md:px-4"
                :class="index > 1 ? 'hidden md:inline-flex' : 'inline-flex'"
                @click.stop
            >
              {{ link.label }}
            </a>
            <NuxtLink
                v-else
                :to="link.to"
                class="rounded-lg px-2 py-2 text-sm font-medium whitespace-nowrap transition-all duration-200 sm:px-3 md:px-4"
                :class="[
                  index > 1 ? 'hidden md:inline-flex' : 'inline-flex',
                  isActive(link)
                    ? 'bg-white/20 text-white'
                    : 'text-gray-100 hover:bg-white/10 hover:text-white'
                ]"
                @click.stop="closeMenus"
            >
              {{ link.label }}
            </NuxtLink>
          </template>
        </nav>
      </div>

      <div class="relative z-10 ml-auto flex shrink-0 items-center gap-2 transition-all duration-300">
        <a
            href="https://github.com/gofurry/gofurry-nav-site"
            target="_blank"
            rel="noopener noreferrer"
            class="hidden h-8 w-16 items-center justify-center rounded-lg transition hover:bg-white/12 xl:inline-flex"
            aria-label="GitHub"
            title="GitHub"
            @click.stop
        >
          <img :src="githubIconSrc" class="h-auto w-12 object-contain" alt="GitHub" />
        </a>

        <div class="hidden items-center gap-1 xl:flex">
          <button
              v-for="option in languageOptions"
              :key="option.value"
              type="button"
              class="flex h-8 w-8 items-center justify-center rounded-lg transition"
              :class="langStore.lang === option.value
                ? 'bg-white/25 text-white'
                : 'text-gray-200 hover:bg-white/10'"
              @click.stop="switchLang(option.value)"
          >
            <img :src="option.flag" class="h-4 w-4" alt="language" />
          </button>
        </div>

        <button
            type="button"
            class="inline-flex h-10 w-10 items-center justify-center rounded-lg text-sm transition hover:bg-white/12 xl:h-8 xl:w-8"
            :aria-label="themeToggleLabel"
            :title="themeToggleLabel"
            @click.stop="toggleThemeIcon"
        >
          <img :src="themeIconSrc" class="h-4 w-4" alt="" />
        </button>

        <button
            type="button"
            class="hidden h-8 w-8 items-center justify-center rounded-lg bg-white/8 text-sm ring-1 ring-white/10 transition hover:bg-white/20 xl:flex"
            :aria-label="t('navbar.mode')"
            @click.stop="showModeModal = true"
        >
          <img :src="gear" class="h-4 w-4" alt="mode" />
        </button>

        <button
            type="button"
            class="inline-flex h-10 w-10 items-center justify-center rounded-lg border border-white/10 bg-white/8 text-white transition hover:bg-white/14 xl:hidden"
            :aria-expanded="mobileMenuOpen"
            :aria-label="t('navbar.expandNav')"
            @click.stop="mobileMenuOpen = !mobileMenuOpen"
        >
          <svg v-if="!mobileMenuOpen" class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
          </svg>
          <svg v-else class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 6l12 12M6 18L18 6" />
          </svg>
        </button>
      </div>
    </div>

    <transition name="mobile-menu">
      <div
          v-if="mobileMenuOpen"
          class="absolute left-3 right-3 top-full z-[90] mt-2 rounded-xl px-4 pb-4 pt-3 text-gray-100 shadow-2xl shadow-slate-950/30 backdrop-blur-xl transition-colors duration-500 xl:hidden"
          :class="mobileMenuClass"
          @click.stop
      >
        <div class="mx-auto flex w-full max-w-[1700px] flex-col gap-2">
          <template v-for="link in navLinks" :key="link.label">
            <a
                v-if="link.external"
                :href="link.href"
                target="_blank"
                rel="noopener noreferrer"
                class="rounded-lg bg-white/5 px-4 py-3 text-sm font-medium text-gray-300 transition-colors hover:bg-white/10"
                @click="mobileMenuOpen = false"
            >
              {{ link.label }}
            </a>
            <NuxtLink
                v-else
                :to="link.to"
                class="rounded-lg px-4 py-3 text-sm font-medium transition-colors"
                :class="isActive(link)
                  ? 'bg-white/20 text-white'
                  : 'bg-white/5 text-gray-300 hover:bg-white/10'"
                @click="closeMenus"
            >
              {{ link.label }}
            </NuxtLink>
          </template>

          <div class="mt-2 flex flex-col gap-3 border-t border-white/10 pt-4">
            <button
                type="button"
                class="flex items-center justify-between rounded-lg bg-white/5 px-4 py-3 text-sm text-gray-100 transition hover:bg-white/10"
                @click="openModeModalFromMobile"
            >
              <span class="flex items-center gap-2">
                <img :src="gear" class="h-4 w-4" alt="mode" />
                {{ t('navbar.mode') }}
              </span>
              <span class="text-orange-300">{{ mode || '--' }}</span>
            </button>

            <div class="grid grid-cols-2 gap-2">
              <button
                  v-for="option in languageOptions"
                  :key="option.value"
                  type="button"
                  class="flex items-center justify-center gap-2 rounded-lg px-4 py-3 text-sm transition"
                  :class="langStore.lang === option.value
                    ? 'bg-white/20 text-white'
                    : 'bg-white/5 text-gray-300 hover:bg-white/10'"
                  @click="switchLang(option.value)"
              >
                <img :src="option.flag" class="h-4 w-4" alt="language" />
                <span>{{ option.label }}</span>
              </button>
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
  </header>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useLangStore } from '@/store/langStore'
import { useThemeStore } from '@/stores/theme'
import { useI18n } from 'vue-i18n'
import cnFlag from '@/assets/flags/cn.svg'
import usFlag from '@/assets/flags/us.svg'
import logo from '@/assets/svgs/logo-mini.svg'
import gear from '@/assets/svgs/gear.svg'
import githubLightIcon from '@/assets/svgs/logo-github-light.svg'
import moonIcon from '@/assets/svgs/moon-light.svg'
import sunIcon from '@/assets/svgs/sun.svg'
import ModeSettingModal from '@/components/common/ModeSettingModal.vue'
import { readMode, subscribeModeChange, writeMode } from '@/utils/modeStorage'

const { t } = useI18n()
const route = useRoute()
const langStore = useLangStore()
const themeStore = useThemeStore()
const props = defineProps<{
  navOverlayDesktop?: boolean
}>()

const showModeModal = ref(false)
const mobileMenuOpen = ref(false)
const mode = ref('')
let stopModeSubscription: (() => void) | null = null

type NavLink = {
  label: string
  to?: string
  href?: string
  external?: boolean
}

const archiveLink = computed<NavLink>(() => (
  { label: t('sidebar.archive'), to: '/archive' }
))

const navLinks = computed<NavLink[]>(() => [
  { label: t('sidebar.nav'), to: '/' },
  { label: t('sidebar.games'), to: '/games' },
  archiveLink.value,
  { label: langStore.lang === 'zh' ? '深度兽研' : 'DeepFurry', href: 'https://www.deepfurry.com', external: true },
])

const languageOptions = [
  { value: 'zh' as const, label: 'CN', flag: cnFlag },
  { value: 'en' as const, label: 'EN', flag: usFlag },
]
const themeIconSrc = computed(() => themeStore.theme === 'light' ? sunIcon : moonIcon)
const githubIconSrc = githubLightIcon
const themeToggleLabel = computed(() => (
  langStore.lang === 'zh' ? '切换明暗主题图标' : 'Toggle theme icon'
))

const headerClass = computed(() => {
  if (themeStore.theme === 'dark') {
    return props.navOverlayDesktop
      ? 'border-b border-white/8 bg-[rgba(5,7,13,0.94)] shadow-black/35 md:border-white/6 md:bg-[rgba(5,7,13,0.76)]'
      : 'border-b border-white/8 bg-[rgba(5,7,13,0.94)] shadow-black/35'
  }

  return props.navOverlayDesktop
    ? 'border-b border-white/15 bg-[rgba(10,16,28,0.88)] md:border-white/10 md:bg-[rgba(18,24,37,0.7)]'
    : 'border-b border-white/15 bg-[rgba(10,16,28,0.88)]'
})

const mobileMenuClass = computed(() => (
  themeStore.theme === 'dark'
    ? 'border border-white/8 bg-[rgba(5,7,13,0.97)]'
    : 'border border-white/10 bg-[rgba(18,24,37,0.96)]'
))

const isActive = (link: NavLink) =>
    Boolean(link.to && (route.path === link.to || route.path.startsWith(`${link.to}/`)))

onMounted(() => {
  mode.value = readMode()
  themeStore.initTheme()

  stopModeSubscription = subscribeModeChange(({ mode: nextMode }) => {
    mode.value = nextMode
  })
})

onUnmounted(() => {
  stopModeSubscription?.()
})

watch(
    () => route.fullPath,
    () => {
      mobileMenuOpen.value = false
    }
)

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

function toggleThemeIcon() {
  themeStore.setTheme(themeStore.theme === 'light' ? 'dark' : 'light')
}

function closeMenus() {
  mobileMenuOpen.value = false
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
