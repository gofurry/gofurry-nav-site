<template>
  <header
    class="gf-nav relative z-[70] w-full backdrop-blur-xl transition-all duration-300"
    :class="{ 'gf-nav--overlay': navOverlayDesktop }"
  >
    <div
        class="relative mx-auto flex w-full max-w-[1700px] items-center gap-3 px-4 py-2 transition-all duration-300 sm:px-6"
    >
      <NuxtLink
          :to="localePath('/')"
          class="relative z-10 flex shrink-0 items-center gap-2 px-2 py-1"
          @click.stop="closeMenus"
      >
        <img :src="logo" alt="GoFurry" class="h-10 w-10" />
        <span class="gf-nav__brand-text hidden text-sm font-semibold tracking-wide sm:inline">GoFurry</span>
      </NuxtLink>

      <div class="pointer-events-none absolute left-1/2 top-1/2 z-0 flex max-w-[calc(100vw-10rem)] -translate-x-1/2 -translate-y-1/2 items-center justify-center overflow-hidden transition-all duration-300 sm:max-w-[calc(100vw-14rem)] md:max-w-[760px]">
        <nav class="pointer-events-auto flex min-w-0 items-center justify-center gap-1 transition-all duration-200 opacity-100 md:min-w-max md:max-w-[760px]">
          <template v-for="(link, index) in navLinks" :key="link.label">
            <a
                v-if="link.external"
                :href="link.href"
                target="_blank"
                rel="noopener noreferrer"
                class="gf-nav__link gf-nav__link--idle rounded-lg px-2 py-2 text-sm font-medium whitespace-nowrap sm:px-3 md:px-4"
                :class="index > 1 ? 'hidden md:inline-flex' : 'inline-flex'"
                @click.stop
            >
              {{ link.label }}
            </a>
            <NuxtLink
                v-else
                :to="link.to"
                class="gf-nav__link rounded-lg px-2 py-2 text-sm font-medium whitespace-nowrap sm:px-3 md:px-4"
                :class="[
                  index > 1 ? 'hidden md:inline-flex' : 'inline-flex',
                  isActive(link)
                    ? 'gf-nav__link--active'
                    : 'gf-nav__link--idle'
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
            class="gf-nav__github-link hidden h-8 w-16 items-center justify-center rounded-lg xl:inline-flex"
            aria-label="GitHub"
            title="GitHub"
            @click.stop
        >
          <img :src="githubIconSrc" class="h-auto w-12 object-contain" alt="" />
        </a>

        <div class="hidden items-center gap-1 xl:flex">
          <button
              v-for="option in languageOptions"
              :key="option.value"
              type="button"
              class="gf-nav__icon-button flex h-8 w-8 items-center justify-center rounded-lg"
              :class="langStore.lang === option.value
                ? 'gf-nav__icon-button--active'
                : ''"
              @click.stop="switchLang(option.value)"
          >
            <img :src="option.flag" class="h-4 w-4" :alt="option.label" />
          </button>
        </div>

        <button
            type="button"
            class="gf-nav__icon-button inline-flex h-10 w-10 items-center justify-center rounded-lg text-sm xl:h-8 xl:w-8"
            :aria-label="themeToggleLabel"
            :title="themeToggleLabel"
            @click.stop="toggleThemeIcon"
        >
          <img :src="themeIconSrc" class="h-4 w-4" alt="" />
        </button>

        <button
            type="button"
            class="gf-nav__icon-button gf-nav__mode-button hidden h-8 w-8 items-center justify-center rounded-lg text-sm xl:flex"
            :aria-label="t('navbar.mode')"
            @click.stop="showModeModal = true"
        >
          <img :src="gear" class="h-4 w-4" alt="" />
        </button>

        <button
            type="button"
            class="gf-nav__mobile-toggle inline-flex h-10 w-10 items-center justify-center rounded-lg xl:hidden"
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
          class="gf-nav__mobile-panel absolute left-3 right-3 top-full z-[90] mt-2 rounded-xl px-4 pb-4 pt-3 backdrop-blur-xl transition-colors duration-500 xl:hidden"
          @click.stop
      >
        <div class="mx-auto flex w-full max-w-[1700px] flex-col gap-2">
          <template v-for="link in navLinks" :key="link.label">
            <a
                v-if="link.external"
                :href="link.href"
                target="_blank"
                rel="noopener noreferrer"
                class="gf-nav__mobile-link rounded-lg px-4 py-3 text-sm font-medium"
                @click="mobileMenuOpen = false"
            >
              {{ link.label }}
            </a>
            <NuxtLink
                v-else
                :to="link.to"
                class="gf-nav__mobile-link rounded-lg px-4 py-3 text-sm font-medium"
                :class="isActive(link)
                  ? 'gf-nav__mobile-link--active'
                  : ''"
                @click="closeMenus"
            >
              {{ link.label }}
            </NuxtLink>
          </template>

          <div class="mt-2 flex flex-col gap-3 border-t border-white/10 pt-4">
            <button
                type="button"
                class="gf-nav__mobile-action flex items-center justify-between rounded-lg px-4 py-3 text-sm"
                @click="openModeModalFromMobile"
            >
              <span class="flex items-center gap-2">
                <img :src="gear" class="h-4 w-4" alt="" />
                {{ t('navbar.mode') }}
              </span>
              <span class="gf-nav__mobile-action-value">{{ mode || '--' }}</span>
            </button>

            <div class="grid grid-cols-2 gap-2">
              <button
                  v-for="option in languageOptions"
                  :key="option.value"
                  type="button"
                  class="gf-nav__mobile-language flex items-center justify-center gap-2 rounded-lg px-4 py-3 text-sm"
                  :class="langStore.lang === option.value
                    ? 'gf-nav__mobile-language--active'
                    : ''"
                  @click="switchLang(option.value)"
              >
                <img :src="option.flag" class="h-4 w-4" :alt="option.label" />
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
const router = useRouter()
const localePath = useLocalePath()
const switchLocalePath = useSwitchLocalePath()
const langStore = useLangStore()
const themeStore = useThemeStore()
defineProps<{
  navOverlayDesktop?: boolean
}>()

const showModeModal = ref(false)
const mobileMenuOpen = ref(false)
const mode = ref('')
let stopModeSubscription: (() => void) | null = null

type NavLink = {
  label: string
  to?: string
  activePath?: string
  href?: string
  external?: boolean
}

const archiveLink = computed<NavLink>(() => (
  { label: t('sidebar.archive'), to: localePath('/archive'), activePath: '/archive' }
))

const navLinks = computed<NavLink[]>(() => [
  { label: t('sidebar.nav'), to: localePath('/'), activePath: '/' },
  { label: t('sidebar.games'), to: localePath('/games'), activePath: '/games' },
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

const normalizeRoutePath = (path: string) =>
    path.replace(/^\/(zh|en)(?=\/|$)/, '') || '/'

const isActive = (link: NavLink) => {
  if (!link.to) return false

  const currentPath = normalizeRoutePath(route.path)
  const activePath = link.activePath || normalizeRoutePath(link.to)

  if (activePath === '/') {
    return currentPath === '/'
  }

  return currentPath === activePath || currentPath.startsWith(`${activePath}/`)
}

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
  const nextPath = switchLocalePath(lang)
  if (nextPath && nextPath !== route.fullPath) {
    router.push(nextPath)
  }
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
