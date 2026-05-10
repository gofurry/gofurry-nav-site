<template>
  <header
    class="relative z-[70] w-full text-gray-100 shadow-lg backdrop-blur-xl transition-all duration-300"
    :class="headerClass"
  >
    <div
        class="mx-auto flex w-full max-w-[1700px] items-center gap-3 px-4 py-2 transition-all duration-300 sm:px-6"
    >
      <NuxtLink
          to="/nav"
          class="flex shrink-0 items-center gap-2 px-2 py-1"
          @click.stop="closeMenus"
      >
        <img :src="logo" alt="GoFurry logo" class="h-10 w-10" />
        <span class="hidden text-sm font-semibold tracking-wide text-white sm:inline">GoFurry</span>
      </NuxtLink>

      <div class="hidden min-w-0 flex-1 items-center justify-center overflow-hidden transition-all duration-300 sm:flex">
        <nav class="flex min-w-max max-w-[760px] items-center justify-center gap-1 transition-all duration-200 opacity-100">
          <template v-for="link in navLinks" :key="link.label">
            <a
                v-if="link.external"
                :href="link.href"
                target="_blank"
                rel="noopener noreferrer"
                class="rounded-lg px-4 py-2 text-sm font-medium whitespace-nowrap text-gray-100 transition-all duration-200 hover:bg-white/10 hover:text-white"
                @click.stop
            >
              {{ link.label }}
            </a>
            <NuxtLink
                v-else
                :to="link.to"
                class="rounded-lg px-4 py-2 text-sm font-medium whitespace-nowrap transition-all duration-200"
                :class="isActive(link)
                  ? 'bg-white/20 text-white'
                  : 'text-gray-100 hover:bg-white/10 hover:text-white'"
                @click.stop="closeMenus"
            >
              {{ link.label }}
            </NuxtLink>
          </template>
        </nav>
      </div>

      <div class="ml-auto hidden items-center gap-2 overflow-hidden transition-all duration-300 xl:flex">
        <div class="flex items-center gap-1">
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
            class="flex h-8 w-8 items-center justify-center rounded-lg bg-white/8 text-sm ring-1 ring-white/10 transition hover:bg-white/20"
            :aria-label="t('navbar.mode')"
            @click.stop="showModeModal = true"
        >
          <img :src="gear" class="h-4 w-4" alt="mode" />
        </button>
      </div>

      <button
          type="button"
          class="ml-auto inline-flex h-10 w-10 items-center justify-center rounded-lg border border-white/10 bg-white/8 text-white transition hover:bg-white/14 xl:hidden"
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

    <transition name="mobile-menu">
      <div
          v-if="mobileMenuOpen"
          class="border-t border-white/10 bg-[rgba(18,24,37,0.96)] px-4 pb-4 pt-3 text-gray-100 shadow-lg xl:hidden"
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
import { useI18n } from 'vue-i18n'
import cnFlag from '@/assets/flags/cn.svg'
import usFlag from '@/assets/flags/us.svg'
import logo from '@/assets/svgs/logo-mini.svg'
import gear from '@/assets/svgs/gear.svg'
import ModeSettingModal from '@/components/common/ModeSettingModal.vue'
import { readMode, subscribeModeChange, writeMode } from '@/utils/modeStorage'

const { t } = useI18n()
const route = useRoute()
const langStore = useLangStore()
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
  { label: t('sidebar.nav'), to: '/nav' },
  { label: t('sidebar.games'), to: '/games' },
  archiveLink.value,
  { label: langStore.lang === 'zh' ? '深度兽研' : 'DeepFurry', href: 'https://www.deepfurry.com', external: true },
])

const languageOptions = [
  { value: 'zh' as const, label: 'CN', flag: cnFlag },
  { value: 'en' as const, label: 'EN', flag: usFlag },
]

const headerClass = computed(() => (
  props.navOverlayDesktop
    ? 'border-b border-white/15 bg-[rgba(10,16,28,0.88)] md:border-white/10 md:bg-[rgba(18,24,37,0.7)]'
    : 'border-b border-white/15 bg-[rgba(10,16,28,0.88)]'
))

const isActive = (link: NavLink) =>
    Boolean(link.to && (route.path === link.to || route.path.startsWith(`${link.to}/`)))

onMounted(() => {
  mode.value = readMode()

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
